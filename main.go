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
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	tui "github.com/gizak/termui/v3"

	"github.com/PantheonTechnologies/vpptop/stats"
	"github.com/PantheonTechnologies/vpptop/xtui"
)

const (
	// Sorting with mouse support is disabled for now
	// 'cause it was problematic to implement it in a way
	// that would fit all tabPanes. For example
	// Nodes and Errors tabPanes contains 1 counter per column which is okay
	// the column can be calculated easily. But the interfaces tabPane
	// contains more counters per column and there is no clear way how
	// to deal with this for now.
	eventMouseLeft      = "<MouseLeft>"
	eventMouseMiddle    = "<MouseMiddle>"
	eventMouseRight     = "<MouseRight>"
	eventMouseRelease   = "<MouseRelease>"
	eventMouseWheelUp   = "<MouseWheelUp>"
	eventMouseWheelDown = "<MouseWheelDown>"

	keyTabLeft    = "<Left>"
	keyTabRight   = "<Right>"
	keyScrollDown = "<Down>"
	keyScrollUp   = "<Up>"
	keyQuit       = "q"
	keyFilter     = "/"
	keyCancel     = "<Escape>"
	keyDeleteChar = "<Backspace>"
	keyF1         = "<F1>"
	keyF2         = "<F2>"
	keyF3         = "<F3>"
	keyF4         = "<F4>"
	keyF5         = "<F5>"
	keyF6         = "<F6>"
	keyF7         = "<F7>"
	keyF8         = "<F8>"
	keyF9         = "<F9>"
	keyF10        = "<F10>"
	keyF11        = "<F11>"
	keyF12        = "<F12>"
	keyInsert     = "<Insert>"
	keyDelete     = "<Delete>"
	keyHome       = "<Home>"
	keyEnd        = "<End>"
	keyPgup       = "<PageUp>"
	keyPgdn       = "<PageDown>"
	keyCtrlSpace  = "<C-<Space>>"
	keyCtrlA      = "<C-a>"
	keyCtrlB      = "<C-b>"
	keyCtrlC      = "<C-c>"
	keyCtrlD      = "<C-d>"
	keyCtrlE      = "<C-e>"
	keyCtrlF      = "<C-f>"
	keyCtrlG      = "<C-g>"
	keyBackspace  = "<C-<Backspace>>"
	keyTab        = "<Tab>"
	keyCtrlJ      = "<C-j>"
	keyCtrlK      = "<C-k>"
	keyCtrlL      = "<C-l>"
	keyEnter      = "<Enter>"
	keyCtrlN      = "<C-n>"
	keyCtrlO      = "<C-o>"
	keyCtrlP      = "<C-p>"
	keyCtrlQ      = "<C-q>"
	keyCtrlR      = "<C-r>"
	keyCtrlS      = "<C-s>"
	keyCtrlT      = "<C-t>"
	keyCtrlU      = "<C-u>"
	keyCtrlV      = "<C-v>"
	keyCtrlW      = "<C-w>"
	keyCtrlX      = "<C-x>"
	keyCtrlY      = "<C-y>"
	keyCtrlZ      = "<C-z>"
	keyCtrl4      = "<C-4>"
	keyCtrl5      = "<C-5>"
	keyCtrl6      = "<C-6>"
	keyCtrl7      = "<C-7>"
)

var (
	// Hold the previous interfaceStats to calculate Bytes/s Packets/s.
	prevInterfaceStats []stats.Interfaces
)

var (
	statSock = flag.String("socket", stats.DefaultSocket, "VPP stats segment socket")
	logFile  = flag.String("log", "vpptop.log", "Log file")
)

func main() {
	flag.Parse()

	if err := stats.Connect(*statSock); err != nil {
		log.Fatalf("Error occured while connecting: %v", err)
	}
	defer stats.Disconnect()

	logs, err := os.Create(*logFile)
	if err != nil {
		log.Fatalf("Error occured while creating file: %v", err)
	}
	defer logs.Close()

	if err := tui.Init(); err != nil {
		log.Fatalf("error occured while initializing termui: %v", err)
	}
	defer tui.Close()

	// set log output to file after tui.Init
	log.SetOutput(logs)

	version.Text, err = stats.Version()
	if err != nil {
		log.Printf("Error stats.Version: %v", err)
	}

	resizeWidgets(tui.TerminalDimensions())

	for i := range views {
		updateTableRows(i)
	}

	renderTicker := time.NewTicker(time.Millisecond * 32).C
	updateTicker := time.NewTicker(time.Second).C
	inputEvents := tui.PollEvents()
	clear := make(chan struct{})
	quit := make(chan struct{})

	go func(clear <-chan struct{}, quit chan struct{}) {
		for {
			select {
			case <-updateTicker:
				updateTableRows(tabPane.ActiveTabIndex)
			case <-clear:
				clearCounters(tabPane.ActiveTabIndex)
			case <-quit:
				close(quit)
				return
			}
		}
	}(clear, quit)

	for {
		select {
		case <-renderTicker:
			tui.Render(tabPane, version)
			tui.Render(views[tabPane.ActiveTabIndex].Header, views[tabPane.ActiveTabIndex].Table)
		case e := <-inputEvents:
			switch e.Type {
			case tui.KeyboardEvent:
				switch e.ID {
				case keyCtrlSpace:
					if tabPane.ActiveTabIndex == Memory || tabPane.ActiveTabIndex == Threads {
						// sorting isn't supported for memory and threads tab.
						break
					}
					renderEventSort(renderTicker, inputEvents)
				case keyQuit:
					quit <- struct{}{}

					tui.Clear()
					// set-up the exit screen
					// to be displayed in the center.
					w, h := tui.TerminalDimensions()
					x1 := w/2 - w/4
					y1 := h/2 - h/4

					x2 := w/2 + w/4
					y2 := h/2 + h/4

					// center the text inside the paragraph
					exitScreen.PaddingLeft = x2/4 - 3
					exitScreen.PaddingTop = y2/4 - 1

					exitScreen.SetRect(x1, y1, x2, y2)

					tui.Render(exitScreen)

					<-quit // wait for goroutine for quit.
					return
				case keyScrollDown:
					views[tabPane.ActiveTabIndex].Table.ScrollDown()
				case keyScrollUp:
					views[tabPane.ActiveTabIndex].Table.ScrollUp()
				case keyPgup:
					views[tabPane.ActiveTabIndex].Table.PageUp()
				case keyPgdn:
					views[tabPane.ActiveTabIndex].Table.PageDown()
				case keyTabLeft:
					tabPane.FocusLeft()
					tui.Clear()
				case keyTabRight:
					tabPane.FocusRight()
					tui.Clear()
				case keyFilter:
					if tabPane.ActiveTabIndex == Threads {
						// filter isn't supported for threads tab.
						break
					}
					renderEventFilter(renderTicker, inputEvents)
				case keyCtrlC:
					clear <- struct{}{}
				}
			case tui.ResizeEvent:
				payload := e.Payload.(tui.Resize)
				resizeWidgets(payload.Width, payload.Height)
				tui.Clear()
			}
		}
	}
}

// renderEventFilter handles the filter event
func renderEventFilter(renderTicker <-chan time.Time, events <-chan tui.Event) {
	for {
		filter.Text = views[tabPane.ActiveTabIndex].Table.Filter()
		select {
		case <-renderTicker:
			tui.Render(tabPane, version, filter, filterExit)
			tui.Render(views[tabPane.ActiveTabIndex].Header, views[tabPane.ActiveTabIndex].Table)
		case e := <-events:
			switch e.Type {
			case tui.ResizeEvent:
				payload := e.Payload.(tui.Resize)
				resizeWidgets(payload.Width, payload.Height)
				tui.Clear()
			case tui.KeyboardEvent:
				switch e.ID {
				case keyCancel, keyScrollUp, keyScrollDown, keyTabLeft, keyTabRight, keyEnter, keyTab:
					tui.Clear()
					return
				case keyDeleteChar:
					views[tabPane.ActiveTabIndex].Table.ReduceFilter(1)
				default:
					if e.ID == "<Space>" {
						e.ID = " "
					}
					views[tabPane.ActiveTabIndex].Table.AppendToFilter(e.ID)
				}
			}
		}
	}
}

// renderEventSort handles the sort event
func renderEventSort(renderTicker <-chan time.Time, events <-chan tui.Event) {
	// Move the table to the right, to clear
	// some space for the sort list to be displayed.
	var (
		ttX = tableTopX
	)
	tableTopX = sortPanelBottomX
	tableHeaderTopX = sortPanelBottomX
	// Resize to the new positions
	resizeWidgets(tui.TerminalDimensions())
	// Clear the old ones..
	tui.Clear()
	for {
		select {
		case <-renderTicker:
			tui.Render(tabPane, version)
			tui.Render(views[tabPane.ActiveTabIndex].Header, views[tabPane.ActiveTabIndex].Table, views[tabPane.ActiveTabIndex].SList)
		case e := <-events:
			switch e.Type {
			case tui.ResizeEvent:
				payload := e.Payload.(tui.Resize)
				resizeWidgets(payload.Width, payload.Height)
				tui.Clear()
			case tui.KeyboardEvent:
				switch e.ID {
				case keyCancel, keyCtrlSpace:
					// Switch back to the original positions
					tableTopX, tableHeaderTopX = ttX, ttX
					// Resize the widgets
					resizeWidgets(tui.TerminalDimensions())
					// Clear the old ones...
					tui.Clear()
					return
				case keyEnter:
					views[tabPane.ActiveTabIndex].ascending = !views[tabPane.ActiveTabIndex].ascending
					views[tabPane.ActiveTabIndex].sortBy = views[tabPane.ActiveTabIndex].SList.SelectedRow
				case keyScrollDown:
					views[tabPane.ActiveTabIndex].SList.ScrollDown()
				case keyScrollUp:
					views[tabPane.ActiveTabIndex].SList.ScrollUp()
				case keyPgdn:
					views[tabPane.ActiveTabIndex].SList.SelectedRow = len(views[tabPane.ActiveTabIndex].SList.Rows) - 1
				case keyPgup:
					views[tabPane.ActiveTabIndex].SList.SelectedRow = 0

				}
			}
		}
	}
}

// columnFromMousePos calculates the column
// of the table based on the mouse position.
//func columnFromMousePos(tableHeader *xtui.Table, x, y int) (int, error) {
//	if y < tableHeaderTopY || y > tableHeaderBottomY {
//		return 0, errors.New("Y-pos out of bounds")
//	}
//	columnsWidths, err := tableHeader.ColumnWidths()
//	log.Println(columnsWidths)
//	if err != nil {
//		return 0, err
//	}
//
//	c := -1
//	w := 0
//	for _, cw := range columnsWidths {
//		w += cw
//		w += 1 // count the column separator
//		c += 1 // increment column index
//		if x <= w {
//			break
//		}
//	}
//	return c, nil
//}

// clearCounters clears all the counters in the specified tab
func clearCounters(tab int) {
	switch tab {
	case Nodes:
		if err := stats.ClearRuntimeCounters(); err != nil {
			log.Printf("Error occured while clearing runetime counters:%v\n", err)
		}
	case Interfaces:
		for i := range prevInterfaceStats {
			if err := stats.ClearIfaceCounters(prevInterfaceStats[i].InterfaceIndex); err != nil {
				log.Printf("Error occured while clearing interface stats counters:%v\n", err)
			}
		}
		prevInterfaceStats = nil
	case Errors:
		if err := stats.ClearErrorCounters(); err != nil {
			log.Printf("Error occured while clearing errrors counters:%v\n", err)
		}
	}
}

// updateTableRows updates the table rows in the specified table.
func updateTableRows(tab int) {
	switch tab {
	case Nodes:
		rows, err := updateNodes()
		if err != nil {
			log.Printf("Error occured while dumping nodes stats: %v\n", err)
			break
		}
		views[Nodes].Table.Rows = rows
	case Interfaces:
		rows, err := updateInterfaces()
		if err != nil {
			log.Printf("Error occured while dumping interface stats: %v\n", err)
			break
		}
		views[Interfaces].Table.Rows = rows
	case Errors:
		rows, err := updateErrors()
		if err != nil {
			log.Printf("Error occured while dumping error stats: %v\n", err)
			break
		}
		views[Errors].Table.Rows = rows
	case Memory:
		rows, err := updateMemory()
		if err != nil {
			log.Printf("Error occured while dumping memory stats: %v\n", err)
			break
		}
		views[Memory].Table.Rows = rows
	case Threads:
		rows, err := updateThreads()
		if err != nil {
			log.Printf("Error occured while dumping threads stats: %v\n", err)
			break
		}
		views[Threads].Table.Rows = rows
	}
}

// updateMemory fetches updated memory usage per thread from the stats package.
func updateMemory() (xtui.TableRows, error) {
	mem, err := stats.Memory()
	if err != nil {
		return nil, err
	}
	// stats.Memory returns the stats as []string
	// where 8 rows corresponds to one entry.
	const rowsPerEntry = 8
	count := len(mem) / rowsPerEntry              // number of entries.
	rows := make([][]string, RowsPerMemory*count) // our view will have 7 rows per entry.
	for i := 0; i < count; i++ {
		rows[RowsPerMemory*i] = []string{mem[rowsPerEntry*i], mem[rowsPerEntry*i+1]}
		rows[RowsPerMemory*i+1] = []string{xtui.EmptyCell, mem[rowsPerEntry*i+2]}
		rows[RowsPerMemory*i+2] = []string{xtui.EmptyCell, mem[rowsPerEntry*i+3]}
		rows[RowsPerMemory*i+3] = []string{xtui.EmptyCell, mem[rowsPerEntry*i+4]}
		rows[RowsPerMemory*i+4] = []string{xtui.EmptyCell, mem[rowsPerEntry*i+5]}
		rows[RowsPerMemory*i+5] = []string{xtui.EmptyCell, mem[rowsPerEntry*i+6]}
		rows[RowsPerMemory*i+6] = []string{xtui.EmptyCell, mem[rowsPerEntry*i+7]}
		rows[RowsPerMemory*i+7] = []string{xtui.EmptyCell, xtui.EmptyCell}
	}
	return rows, nil
}

// updateThreads fetches updated thread stats from the stats package.
func updateThreads() (xtui.TableRows, error) {
	threadStats, err := stats.Threads()
	if err != nil {
		return nil, err
	}
	rows := make(xtui.TableRows, len(threadStats))
	for i, thread := range threadStats {
		rows[i] = strings.Split(fmt.Sprintf("%d %s %s %d %d %d %d", thread.ID, thread.Name, thread.Type, thread.PID, thread.CPUID, thread.Core, thread.CPUSocket), " ")
	}
	return rows, nil
}

// updateErrors fetches updated errors stats from the stats package
// sorts them based on the defined column.
func updateErrors() (xtui.TableRows, error) {
	errorStats, err := stats.GetErrors()
	if err != nil {
		return nil, err
	}
	sortErrorStats(errorStats, views[Errors].sortBy, views[Errors].ascending)

	rows := make(xtui.TableRows, len(errorStats))
	for i, errorC := range errorStats {
		rows[i] = strings.Split(fmt.Sprintf("%d/%s/%s", errorC.Value, errorC.NodeName, errorC.Reason), "/")
	}
	return rows, nil
}

// updateNodes fetches updated nodes stats from the stats package
// sorts them based on the defined column.
func updateNodes() (xtui.TableRows, error) {
	nodeStats, err := stats.GetNodes()
	if err != nil {
		return nil, err
	}
	sortNodeStats(nodeStats, views[Nodes].sortBy, views[Nodes].ascending)

	rows := make(xtui.TableRows, len(nodeStats))
	for i, node := range nodeStats {
		rows[i] = strings.Split(fmt.Sprintf("%s %d %d %d %d %d %.2f", node.NodeName, node.NodeIndex, node.Clocks, node.Vectors, node.Calls, node.Suspends, node.VC), " ")
	}
	return rows, nil
}

// updateInterfaces fetches updated interfaces stats from the stats package
// sorts them based on the defined column.
func updateInterfaces() (xtui.TableRows, error) {
	interfaceStats, err := stats.GetInterfaces()
	if err != nil {
		return nil, err
	}

	// Since the updated interfaces could have changed
	// (new interfaces could be created, or some interfaces could be deleted)
	// build a lookup table to be able to check for common interfaces
	// to be able to calculate bytes/s, packets/s
	nameToIdx := make(map[string]int)
	for i, iface := range prevInterfaceStats {
		nameToIdx[iface.InterfaceName] = i
	}
	sortInterfaceStats(interfaceStats, views[Interfaces].sortBy, views[Interfaces].ascending)

	rows := make(xtui.TableRows, RowsPerIface*len(interfaceStats))
	for i, iface := range interfaceStats {
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
			rxbbs = iface.RxBytes - prevInterfaceStats[idx].RxBytes
			txbbs = iface.TxBytes - prevInterfaceStats[idx].TxBytes

			rxpps = iface.RxPackets - prevInterfaceStats[idx].RxPackets
			txpps = iface.TxPackets - prevInterfaceStats[idx].TxPackets
		}
		rows[RowsPerIface*i+1] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "Packets/s", fmt.Sprint(rxpps), "Packets/s", fmt.Sprint(txpps), xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+2] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "Bytes", fmt.Sprint(iface.RxBytes), "Bytes", fmt.Sprint(iface.TxBytes), xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+3] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "Bytes/s", fmt.Sprint(rxbbs), "Bytes/s", fmt.Sprint(txbbs), xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+4] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "Errors", fmt.Sprint(iface.RxErrors), "Errors", fmt.Sprint(iface.TxErrors), xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+5] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "Unicast", fmt.Sprintf("%d/%d", iface.RxUnicast[0], iface.RxUnicast[1]), "UnicastMiss", fmt.Sprintf("%d/%d", iface.TxUnicastMiss[0], iface.TxUnicastMiss[1]), xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+6] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "Multicast", fmt.Sprintf("%d/%d", iface.RxMulticast[0], iface.RxMulticast[1]), "Multicast", fmt.Sprintf("%d/%d", iface.TxMulticast[0], iface.TxMulticast[1]), xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+7] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "Broadcast", fmt.Sprintf("%d/%d", iface.RxBroadcast[0], iface.RxBroadcast[1]), "Broadcast", fmt.Sprintf("%d/%d", iface.TxBroadcast[0], iface.TxBroadcast[1]), xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+8] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "NoBuf", fmt.Sprint(iface.RxNoBuf), xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+9] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, "Miss", fmt.Sprint(iface.RxMiss), xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell}
		rows[RowsPerIface*i+10] = []string{xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell, xtui.EmptyCell}

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
	prevInterfaceStats = interfaceStats

	return rows, nil
}
