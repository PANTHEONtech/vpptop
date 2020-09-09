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

package client

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/PantheonTechnologies/vpptop/gui"
	"github.com/PantheonTechnologies/vpptop/gui/views"
	"github.com/PantheonTechnologies/vpptop/gui/xtui"
	"github.com/PantheonTechnologies/vpptop/stats"
	"github.com/PantheonTechnologies/vpptop/stats/api"
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
	// RowsPerMemory represents number of rows in the xtui table per memory.
	RowsPerMemory = 8
)

// VPP API handler definition list determines supported versions
// - VPPs supported by Ligato VPP-Agent
// - VPPs supported by the local implementation
var Defs []api.HandlerDef

// App groups VPP provider, GUI and caches
type App struct {
	gui         *gui.TermWindow
	vppProvider api.VppProviderAPI

	// Cache for interface stats to
	// be able to calculate bytes/s packets/s.
	ifCache []api.Interface

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

func NewApp(lightTheme bool) (*App, error) {
	app := new(App)

	app.sortLock = new(sync.Mutex)
	app.tabLock = new(sync.Mutex)
	app.vppLock = new(sync.Mutex)

	if len(Defs) == 0 {
		return nil, fmt.Errorf("no VPP handler definition was provided")
	}
	app.vppProvider = stats.NewVppProvider(Defs)
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
				[]int{24, 5, 5, 20, 10, 16, 11, 16, 11, 11, 11, views.Resize},
				lightTheme,
			),
			// node tab.
			views.NewTableView(
				[]string{
					"Name",
					"State",
					"Calls",
					"Vectors",
					"Suspends",
					"Clocks",
					"Vectors/Calls",
				},
				xtui.TableRows{{"Name", "State", "Calls", "Vectors", "Suspends", "Clocks", "Vectors/Calls"}},
				NodeStatNodeName,
				1,
				[]int{50, views.Resize, views.Resize, views.Resize, views.Resize, views.Resize, 22},
				lightTheme,
			),
			// errors tab.
			views.NewTableView(
				[]string{"Counter", "Node", "Reason"},
				xtui.TableRows{{"Counter", "Node", "Reason"}},
				ErrorStatErrorNodeName,
				1,
				nil,
				lightTheme,
			),
			// memory tab.
			views.NewTableView(
				[]string{},
				xtui.TableRows{{"Thread/ID/Name", "Current memory usage per Thread"}},
				MemoryStatName,
				RowsPerMemory,
				[]int{30, views.Resize},
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

	return app, nil
}

// Init initializes app.
func (app *App) Init(soc, rAddr string) error {
	switch rAddr {
	case "":
		if err := app.vppProvider.Connect(soc); err != nil {
			return err
		}
	default:
		if err := app.vppProvider.ConnectRemote(rAddr); err != nil {
			return err
		}
	}

	if err := app.gui.Init(); err != nil {
		return err
	}

	app.gui.SetVersion(app.vppProvider.GetVersion())

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
					ifaces, err := app.vppProvider.GetInterfaces(ctx)
					if err != nil {
						log.Printf("error occured while polling interface stats: %v\n", err)
					}

					app.sortLock.Lock()
					s := app.sortBy[Interfaces]
					app.sortLock.Unlock()

					app.sortInterfaceStats(ifaces, s.field, s.asc)
					app.gui.ViewAtTab(Interfaces).Update(app.formatInterfaces(ifaces))

				case Nodes:
					nodes, err := app.vppProvider.GetNodes(ctx)
					if err != nil {
						log.Printf("error occured while polling nodes stats: %v\n", err)
					}

					app.sortLock.Lock()
					s := app.sortBy[Nodes]
					app.sortLock.Unlock()

					app.sortNodeStats(nodes, s.field, s.asc)
					app.gui.ViewAtTab(Nodes).Update(app.formatNodes(nodes))

				case Errors:
					errors, err := app.vppProvider.GetErrors(ctx)
					if err != nil {
						log.Printf("error occured while polling errors stats: %v\n", err)
					}

					app.sortLock.Lock()
					s := app.sortBy[Errors]
					app.sortLock.Unlock()

					app.sortErrorStats(errors, s.field, s.asc)
					app.gui.ViewAtTab(Errors).Update(app.formatErrors(errors))

				case Memory:
					memstats, err := app.vppProvider.GetMemory(ctx)
					if err != nil {
						log.Printf("error occured while polling memory stats: %v\n", err)
					}

					app.gui.ViewAtTab(Memory).Update(app.formatMemstats(memstats))

				case Threads:
					threads, err := app.vppProvider.GetThreads(ctx)
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
				if err := app.vppProvider.ClearInterfaceCounters(ctx); err != nil {
					log.Printf("error occured while clearing interface stats: %v\n", err)
				}
				app.ifCache = nil
			case Nodes:
				if err := app.vppProvider.ClearRuntimeCounters(ctx); err != nil {
					log.Printf("error occured while clearing node stats: %v\n", err)
				}
			case Errors:
				if err := app.vppProvider.ClearErrorCounters(ctx); err != nil {
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
		app.vppProvider.Disconnect()
	})

	app.gui.AddOnTabSwitchCallback(func(event gui.Event) {
		app.tabLock.Lock()
		defer app.tabLock.Unlock()
		app.currTab = event.Payload.(int)
	})

	app.gui.Start()
}

// formatInterfaces formats interface stats to xtui.TableRows
func (app *App) formatInterfaces(ifaces []api.Interface) xtui.TableRows {
	nameToIdx := make(map[string]int)

	for i, iface := range app.ifCache {
		nameToIdx[iface.InterfaceName] = i
	}

	rows := make(xtui.TableRows, RowsPerIface*len(ifaces))
	for i, iface := range ifaces {
		rows[RowsPerIface*i] = append(rows[RowsPerIface*i], iface.InterfaceName)
		rows[RowsPerIface*i] = append(rows[RowsPerIface*i], fmt.Sprint(iface.InterfaceIndex))
		rows[RowsPerIface*i] = append(rows[RowsPerIface*i], iface.State)
		rows[RowsPerIface*i] = append(rows[RowsPerIface*i], fmt.Sprintf("%d/%d/%d/%d", iface.MTU[0], iface.MTU[1], iface.MTU[2], iface.MTU[3]))
		rows[RowsPerIface*i] = append(rows[RowsPerIface*i], "Packets")
		rows[RowsPerIface*i] = append(rows[RowsPerIface*i], fmt.Sprint(iface.Rx.Packets))
		rows[RowsPerIface*i] = append(rows[RowsPerIface*i], "Packets")
		rows[RowsPerIface*i] = append(rows[RowsPerIface*i], fmt.Sprint(iface.Tx.Packets))
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
			rxbbs = iface.Rx.Bytes - app.ifCache[idx].Rx.Bytes
			txbbs = iface.Tx.Bytes - app.ifCache[idx].Tx.Bytes

			rxpps = iface.Rx.Packets - app.ifCache[idx].Rx.Packets
			txpps = iface.Tx.Packets - app.ifCache[idx].Tx.Packets
		}

		rows[RowsPerIface*i+1] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "Packets/s", fmt.Sprint(rxpps), "Packets/s", fmt.Sprint(txpps), xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+2] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "Bytes", fmt.Sprint(iface.Rx.Bytes), "Bytes", fmt.Sprint(iface.Tx.Bytes), xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+3] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "Bytes/s", fmt.Sprint(rxbbs), "Bytes/s", fmt.Sprint(txbbs), xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+4] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "Errors", fmt.Sprint(iface.RxErrors), "Errors", fmt.Sprint(iface.TxErrors), xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+5] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "Unicast", fmt.Sprintf("%d/%d", iface.RxUnicast.Packets, iface.RxUnicast.Bytes), "UnicastMiss", fmt.Sprintf("%d/%d", iface.TxUnicast.Packets, iface.TxUnicast.Bytes), xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+6] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "Multicast", fmt.Sprintf("%d/%d", iface.RxMulticast.Packets, iface.RxMulticast.Bytes), "Multicast", fmt.Sprintf("%d/%d", iface.TxMulticast.Packets, iface.TxMulticast.Bytes), xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+7] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "Broadcast", fmt.Sprintf("%d/%d", iface.RxBroadcast.Packets, iface.RxBroadcast.Bytes), "Broadcast", fmt.Sprintf("%d/%d", iface.TxBroadcast.Packets, iface.TxBroadcast.Bytes), xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+8] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "NoBuf", fmt.Sprint(iface.RxNoBuf), xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+9] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "Miss", fmt.Sprint(iface.RxMiss), xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+10] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell}

		// start from the second row, the first is taken up
		// by the interface name.
		row := RowsPerIface*i + 1
		ip := len(iface.IPAddresses)
		maxRow := RowsPerIface*i + RowsPerIface // last row of each entry.

		for ip > 0 && row < maxRow {
			rows[row][0] = strings.Split(iface.IPAddresses[ip-1], "/")[0]
			ip--
			row++
		}
	}

	app.ifCache = ifaces

	return rows
}

// formatNodes formats nodes stats to xtui.TableRows
func (app *App) formatNodes(nodes []api.Node) xtui.TableRows {
	rows := make(xtui.TableRows, len(nodes))

	for i, node := range nodes {
		rows[i] = strings.Split(fmt.Sprintf("%s %s %d %d %d %d %.2f", node.Name, node.State, node.Calls, node.Vectors, node.Suspends, uint64(node.Clocks), node.VectorsPerCall), " ")
	}

	return rows
}

// formatErrors formats error stats to xtui.TableRows
func (app *App) formatErrors(errors []api.Error) xtui.TableRows {
	rows := make(xtui.TableRows, len(errors))

	for i, errorC := range errors {
		rows[i] = strings.Split(fmt.Sprintf("%d;%s;%s", errorC.Value, errorC.Node, errorC.Name), ";")
	}

	if len(rows) == 0 {
		rows = append(rows, []string{"", "", ""})
	}

	return rows

}

// formatMemstats formats memory stats to xtui.TableRows
func (app *App) formatMemstats(memstats []string) xtui.TableRows {
	// vppProvider.GetMemory returns the stats as []string
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
func (app *App) formatThreads(threads []api.ThreadData) xtui.TableRows {
	rows := make(xtui.TableRows, len(threads))

	for i, thread := range threads {
		rows[i] = strings.Split(fmt.Sprintf("%d %s %s %d %d %d %d", thread.ID, thread.Name, thread.Type, thread.PID, thread.CPUID, thread.Core, thread.CPUSocket), " ")
	}

	return rows
}
