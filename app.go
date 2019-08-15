/*
 * Copyright (c) 2019 PANTHEON.tech.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/PantheonTechnologies/vpptop/bin_api/vpe"
	"github.com/PantheonTechnologies/vpptop/gui"
	"github.com/PantheonTechnologies/vpptop/gui/views"
	"github.com/PantheonTechnologies/vpptop/gui/xtui"
	"github.com/PantheonTechnologies/vpptop/stats"
)

// Index for each TableView. (total of 5 tabs)
const (
	Interfaces = iota
	Nodes
	Errors
	Memory
	Threads
)

const (
	// RowsPerIface represents number of rows in the xtui table per interface
	RowsPerIface = 11
	// RowsPerNode represents number of rows in the xtui table per node
	RowsPerNode = 1
	// RowsPerError represents number of rows in the xtui table per error
	RowsPerError = 1
	// RowsPerMemory represents number of rows in the xtui table per memory.
	RowsPerMemory = 8
)

type App struct {
	gui *gui.TermWindow
	vpp *stats.VPP

	// Cache for interface stats to
	// be able to calculate bytes/s packates/s.
	IfCache []stats.Interface

	// sortBy carries information used at sorting stats
	// for each tab.
	sortBy []struct {
		asc   bool
		field int
	}

	// current gui tab.
	currTab int

	// go routine management.
	wg       *sync.WaitGroup
	sortLock *sync.Mutex
	tabLock  *sync.Mutex
	vppLock  *sync.Mutex
	cancel   context.CancelFunc
}

func NewApp() *App {
	app := new(App)

	app.sortLock = new(sync.Mutex)
	app.tabLock = new(sync.Mutex)
	app.vppLock = new(sync.Mutex)

	app.vpp = new(stats.VPP)
	app.wg = new(sync.WaitGroup)
	app.sortBy = make([]struct {
		asc   bool
		field int
	}, 5)

	for i := range app.sortBy {
		app.sortBy[i].field = NoColumn
		app.sortBy[i].asc = !app.sortBy[i].asc
	}

	app.gui = gui.NewTermWindow(
		16*time.Millisecond,
		[]gui.TabView{
			// interface tab.
			views.NewTableView(
				[]string{
					"Name",
					"Index",
					"State",
					"MTU-L3",
					"MTU-IP4",
					"MTU-IP6",
					"MTU-MPLS",
					"RxPackets",
					"RxBytes",
					"RxErrors",
					"RxUnicast-packets",
					"RxUnicast-bytes",
					"RxMulticast-packets",
					"RxMulticast-bytes",
					"RxBroadcast-packets",
					"RxBroadcast-bytes",
					"TxPackets",
					"TxBytes",
					"TxErrors",
					"TxUnicastMiss-packets",
					"TxUnicastMiss-bytes",
					"TxMulticast-packets",
					"TxMulticast-bytes",
					"TxBroadcast-packets",
					"TxBroadcast-bytes",
					"Drops",
					"Punts",
					"IP4",
					"IP6",
				},
				xtui.TableRows{{"Name", "Idx", "State", "MTU(L3/IP4/IP6/MPLS)", "RxCounters", "RxCount", "TxCounters", "TxCount", "Drops", "Punts", "IP4", "IP6"}},
				IfaceStatIfaceName,
				RowsPerIface,
				[]int{24, 5, 5, 20, 10, 16, 11, 16, 11, 11, 11, views.TableColResizedWithWindow},
				lightTheme,
			),
			// node tab.
			views.NewTableView(
				[]string{
					"NodeName",
					"NodeIndex",
					"Clocks",
					"Vectors",
					"Calls",
					"Suspends",
					"Vectors/Calls",
				},
				xtui.TableRows{{"NodeName", "NodeIndex", "Clocks", "Vectors", "Calls", "Suspends", "Vectors/Calls"}},
				NodeStatNodeName,
				RowsPerNode,
				[]int{50, 10, views.TableColResizedWithWindow, views.TableColResizedWithWindow, views.TableColResizedWithWindow, views.TableColResizedWithWindow, 22},
				lightTheme,
			),
			// errors tab.
			views.NewTableView(
				[]string{"Counter", "Node", "Reason"},
				xtui.TableRows{{"Counter", "Node", "Reason"}},
				ErrorStatErrorNodeName,
				RowsPerError,
				nil,
				lightTheme,
			),
			// memory tab.
			views.NewTableView(
				[]string{},
				xtui.TableRows{{"Thread/ID/Name", "Current memory usage per Thread"}},
				0,
				RowsPerMemory,
				[]int{30, views.TableColResizedWithWindow},
				lightTheme,
			),
			// threads tab.
			views.NewTableView(
				[]string{},
				xtui.TableRows{{"ID", "Name", "Type", "PID", "CPUID", "Core", "CPUSocket"}},
				NoColumn,
				1,
				nil,
				lightTheme,
			),
		},
		[]string{"Interfaces", "Nodes", "Errors", "Memory", "Threads"},
		[]int{Interfaces, Nodes, Errors},
		views.NewExitView(),
	)

	return app
}

// Init initializes app.
func (app *App) Init(soc string) error {
	// connect to vpp.
	if err := app.vpp.Connect(soc); err != nil {
		return err
	}

	// init gui.
	if err := app.gui.Init(); err != nil {
		return err
	}

	// set the vpp version.
	v, err := app.vpp.Version()
	if err != nil {
		return err
	}
	app.gui.SetVersion(v)

	return nil
}

// Start starts the application.
func (app *App) Run() {
	var ctx context.Context
	ctx, app.cancel = context.WithCancel(context.Background())

	currTab := func() int {
		app.tabLock.Lock()
		defer app.tabLock.Unlock()
		return app.currTab
	}

	app.wg.Add(1)

	go func() {
		updateTicker := time.NewTicker(1 * time.Second).C
		for {
			select {
			case <-updateTicker:
				app.vppLock.Lock()

				switch currTab() {
				case Interfaces:
					ifaces, err := app.vpp.GetInterfaces()
					if err != nil {
						log.Printf("error occured while polling interface stats: %v\n", err)
					}
					app.sortLock.Lock()
					s := app.sortBy[Interfaces]
					app.sortLock.Unlock()

					app.sortInterfaceStats(ifaces, s.field, s.asc)
					app.gui.ViewAtTab(Interfaces).Update(app.formatInterfaces(ifaces))
				case Nodes:
					nodes, err := app.vpp.GetNodes()
					if err != nil {
						log.Printf("error occured while polling nodes stats: %v\n", err)
					}
					app.sortLock.Lock()
					s := app.sortBy[Nodes]
					app.sortLock.Unlock()

					app.sortNodeStats(nodes, s.field, s.asc)
					app.gui.ViewAtTab(Nodes).Update(app.formatNodes(nodes))
				case Errors:
					errors, err := app.vpp.GetErrors()
					if err != nil {
						log.Printf("error occured while polling errors stats: %v\n", err)
					}
					app.sortLock.Lock()
					s := app.sortBy[Errors]
					app.sortLock.Unlock()

					app.sortErrorStats(errors, s.field, s.asc)
					app.gui.ViewAtTab(Errors).Update(app.formatErrors(errors))
				case Memory:
					memstats, err := app.vpp.Memory()
					if err != nil {
						log.Printf("error occured while polling memory stats: %v\n", err)
					}
					app.gui.ViewAtTab(Memory).Update(app.formatMemstats(memstats))
				case Threads:
					threads, err := app.vpp.Threads()
					if err != nil {
						log.Printf("error occured while polling threads stats: %v\n", err)
					}
					app.gui.ViewAtTab(Threads).Update(app.formatThreads(threads))
				}
				app.vppLock.Unlock()
			case <-ctx.Done():
				app.wg.Done()
				return
			}
		}
	}()

	app.gui.AddOnClearCallback(func(event gui.Event) {
		tab := event.Payload.(int)
		// launch in background
		app.wg.Add(1)
		go func() {
			app.vppLock.Lock()
			defer app.vppLock.Unlock()
			defer app.wg.Done()

			switch tab {
			case Interfaces:
				if err := app.vpp.ClearIfaceCounters(); err != nil {
					log.Printf("error occured while clearing interface stats: %v\n", err)
				}
				app.IfCache = nil
			case Nodes:
				if err := app.vpp.ClearRuntimeCounters(); err != nil {
					log.Printf("error occured while clearing node stats: %v\n", err)
				}
			case Errors:
				if err := app.vpp.ClearErrorCounters(); err != nil {
					log.Printf("error occured while clearing error stats: %v\n", err)
				}
			}

		}()
	})

	app.gui.AddOnSortCallback(func(event gui.Event) {
		payload := event.Payload.(gui.SortMetadata)

		app.wg.Add(1)
		go func() {
			defer app.wg.Done()

			app.sortLock.Lock()
			defer app.sortLock.Unlock()

			switch payload.CurrTab {
			case Interfaces:
				app.sortBy[Interfaces].field = payload.CurrRow
				app.sortBy[Interfaces].asc = !app.sortBy[Interfaces].asc
			case Nodes:
				app.sortBy[Nodes].field = payload.CurrRow
				app.sortBy[Nodes].asc = !app.sortBy[Nodes].asc
			case Errors:
				app.sortBy[Errors].field = payload.CurrRow
				app.sortBy[Errors].asc = !app.sortBy[Errors].asc
			}
		}()
	})

	app.gui.AddOnExitCallback(func(_ gui.Event) {
		app.cancel()
		app.wg.Wait()
		app.gui.Destroy()
		app.vpp.Disconnect()
	})

	app.gui.AddOnTabSwitchCallback(func(event gui.Event) {
		app.tabLock.Lock()
		defer app.tabLock.Unlock()
		app.currTab = event.Payload.(int)
	})

	app.gui.Start()
}

// formatInterfaces formats interface stats to xtui.TableRows
func (app *App) formatInterfaces(ifaces []stats.Interface) xtui.TableRows {
	// Since the updated interfaces could have changed
	// (new interfaces could be created, or some interfaces could be deleted)
	// build a lookup table to be able to check for common interfaces
	// to be able to calculate bytes/s, packets/s
	nameToIdx := make(map[string]int)
	for i, iface := range app.IfCache {
		nameToIdx[iface.InterfaceName] = i
	}

	rows := make(xtui.TableRows, RowsPerIface*len(ifaces))
	for i, iface := range ifaces {
		rows[RowsPerIface*i] = append(rows[RowsPerIface*i], iface.InterfaceName)
		rows[RowsPerIface*i] = append(rows[RowsPerIface*i], fmt.Sprint(iface.InterfaceIndex))
		rows[RowsPerIface*i] = append(rows[RowsPerIface*i], iface.State)
		rows[RowsPerIface*i] = append(rows[RowsPerIface*i], fmt.Sprintf("%d/%d/%d/%d", iface.Mtu[0], iface.Mtu[1], iface.Mtu[2], iface.Mtu[3]))
		rows[RowsPerIface*i] = append(rows[RowsPerIface*i], "Packets")
		rows[RowsPerIface*i] = append(rows[RowsPerIface*i], fmt.Sprint(iface.RxPackets))
		rows[RowsPerIface*i] = append(rows[RowsPerIface*i], "Packets")
		rows[RowsPerIface*i] = append(rows[RowsPerIface*i], fmt.Sprint(iface.TxPackets))
		rows[RowsPerIface*i] = append(rows[RowsPerIface*i], fmt.Sprint(iface.Drops))
		rows[RowsPerIface*i] = append(rows[RowsPerIface*i], fmt.Sprint(iface.Punts))
		rows[RowsPerIface*i] = append(rows[RowsPerIface*i], fmt.Sprint(iface.IP4))
		rows[RowsPerIface*i] = append(rows[RowsPerIface*i], fmt.Sprint(iface.IP6))

		rxbbs := uint64(0) //rx bytes/s
		txbbs := uint64(0) //tx bytes/s
		rxpps := uint64(0) //rx packets/s
		txpps := uint64(0) //tx packets/s

		if idx, ok := nameToIdx[iface.InterfaceName]; ok {
			// Calculate bytes/s, packets/s
			rxbbs = iface.RxBytes - app.IfCache[idx].RxBytes
			txbbs = iface.TxBytes - app.IfCache[idx].TxBytes

			rxpps = iface.RxPackets - app.IfCache[idx].RxPackets
			txpps = iface.TxPackets - app.IfCache[idx].TxPackets
		}
		rows[RowsPerIface*i+1] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "Packets/s", fmt.Sprint(rxpps), "Packets/s", fmt.Sprint(txpps), xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+2] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "Bytes", fmt.Sprint(iface.RxBytes), "Bytes", fmt.Sprint(iface.TxBytes), xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+3] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "Bytes/s", fmt.Sprint(rxbbs), "Bytes/s", fmt.Sprint(txbbs), xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+4] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "Errors", fmt.Sprint(iface.RxErrors), "Errors", fmt.Sprint(iface.TxErrors), xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+5] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "Unicast", fmt.Sprintf("%d/%d", iface.RxUnicast[0], iface.RxUnicast[1]), "UnicastMiss", fmt.Sprintf("%d/%d", iface.TxUnicastMiss[0], iface.TxUnicastMiss[1]), xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+6] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "Multicast", fmt.Sprintf("%d/%d", iface.RxMulticast[0], iface.RxMulticast[1]), "Multicast", fmt.Sprintf("%d/%d", iface.TxMulticast[0], iface.TxMulticast[1]), xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+7] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "Broadcast", fmt.Sprintf("%d/%d", iface.RxBroadcast[0], iface.RxBroadcast[1]), "Broadcast", fmt.Sprintf("%d/%d", iface.TxBroadcast[0], iface.TxBroadcast[1]), xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+8] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "NoBuf", fmt.Sprint(iface.RxNoBuf), xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+9] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "Miss", fmt.Sprint(iface.RxMiss), xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+10] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell}

		// start from the second row, the first is taken up
		// by the interface name.
		row := RowsPerIface*i + 1

		ip4Len := len(iface.IPv4)
		ip6Len := len(iface.IPv6)

		maxRow := RowsPerIface*i + RowsPerIface // last row of each entry.
		// fill Ipv4 addresses
		for ip4Len > 0 && row < maxRow {
			rows[row][0] = iface.IPv4[ip4Len-1]
			ip4Len--
			row++
		}
		// fill Ipv6 addresses
		for ip6Len > 0 && row < maxRow {
			rows[row][0] = iface.IPv6[ip6Len-1]
			ip6Len--
			row++
		}
	}
	app.IfCache = ifaces

	return rows
}

// formatNodes formats nodes stats to xtui.TableRows
func (app *App) formatNodes(nodes []stats.Node) xtui.TableRows {
	rows := make(xtui.TableRows, len(nodes))
	for i, node := range nodes {
		rows[i] = strings.Split(fmt.Sprintf("%s %d %d %d %d %d %.2f", node.NodeName, node.NodeIndex, node.Clocks, node.Vectors, node.Calls, node.Suspends, node.VC), " ")
	}
	return rows
}

// formatErrors formats error stats to xtui.TableRows
func (app *App) formatErrors(errors []stats.Error) xtui.TableRows {
	rows := make(xtui.TableRows, len(errors))
	for i, errorC := range errors {
		rows[i] = strings.Split(fmt.Sprintf("%d/%s/%s", errorC.Value, errorC.NodeName, errorC.Reason), "/")
	}
	return rows
}

// formatMemstats formats memory stats to xtui.TableRows
func (app *App) formatMemstats(memstats []string) xtui.TableRows {
	// stats.Memory returns the stats as []string
	// where 7 rows corresponds to one entry.
	const rowsPerEntry = 7
	count := len(memstats) / rowsPerEntry         // number of entries.
	rows := make([][]string, RowsPerMemory*count) // our view will have 6 rows per entry.
	for i := 0; i < count; i++ {
		rows[RowsPerMemory*i] = []string{memstats[rowsPerEntry*i], memstats[rowsPerEntry*i+1]}
		rows[RowsPerMemory*i+1] = []string{xtui.EmptyCell, memstats[rowsPerEntry*i+2]}
		rows[RowsPerMemory*i+2] = []string{xtui.EmptyCell, memstats[rowsPerEntry*i+3]}
		rows[RowsPerMemory*i+3] = []string{xtui.EmptyCell, memstats[rowsPerEntry*i+4]}
		rows[RowsPerMemory*i+4] = []string{xtui.EmptyCell, memstats[rowsPerEntry*i+5]}
		rows[RowsPerMemory*i+5] = []string{xtui.EmptyCell, memstats[rowsPerEntry*i+6]}
		rows[RowsPerMemory*i+6] = []string{xtui.EmptyCell, xtui.EmptyCell}
	}
	return rows
}

// formatThreads formats memory stats to xtui.TableRows
func (app *App) formatThreads(threads []vpe.ThreadData) xtui.TableRows {
	rows := make(xtui.TableRows, len(threads))
	for i, thread := range threads {
		rows[i] = strings.Split(fmt.Sprintf("%d %s %s %d %d %d %d", thread.ID, thread.Name, thread.Type, thread.PID, thread.CPUID, thread.Core, thread.CPUSocket), " ")
	}
	return rows
}
