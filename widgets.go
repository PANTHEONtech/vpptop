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
	"fmt"
	"sync"
	"time"

	"github.com/PantheonTechnologies/vpptop/xtui"
	tui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type view struct {
	Table     *xtui.Table   // stats/data table.
	Header    *xtui.Table   // header of the table.
	SList     *widgets.List // list containing the column names the user can select
	sortBy    int           // column
	ascending bool          // if ascending or descending
}

// To change the order of the view
// Its trivial just swap the indexes below
// and then swap the tabPane names.

// Index for each view.
const (
	Interfaces = iota
	Nodes
	Errors
	Memory
	Threads
)

const (
	// RowsPerIface represents number of rows in the table per interface
	RowsPerIface = 11
	// RowsPerNode represents number of rows in the table per node
	RowsPerNode = 1
	// RowsPerError represents number of rows in the table per error
	RowsPerError = 1
	// RowsPerMemory represents number of rows in the table per memory.
	RowsPerMemory = 8
)

// Widgets positions that will not change.
const (
	tabPaneTopX    = 0
	tabPaneTopY    = 0
	tabPaneBottomX = 50
	tabPaneBottomY = 4

	versionTopX    = 50
	versionTopY    = 0
	versionBottomX = 110
	versionBottomY = 4

	filterTopX    = 24
	filterTopY    = 3
	filterBottomX = 200
	filterBottomY = 6

	filterExitTopX    = 0
	filterExitTopY    = 3
	filterExitBottomX = 24
	filterExitBottomY = 6
)

// Widgets positions that might change.
var (
	tableTopX = 0
	tableTopY = 7
	//tableBottomX - resized with terminal window
	//tableBottomY - resized with terminal window

	tableHeaderTopX    = tableTopX
	tableHeaderTopY    = tabPaneBottomY + 1
	tableHeaderBottomY = tabPaneBottomY + 4

	sortPanelTopX    = tableHeaderTopX
	sortPanelTopY    = tableHeaderTopY
	sortPanelBottomX = 23
	//sortPanelBottomY - resized with terminal window
)

// Widgets.
var (
	views         []*view
	tabPane       *widgets.TabPane
	version       *widgets.Paragraph
	filter        *widgets.Paragraph
	filterExit    *widgets.Paragraph
	exitScreen    *widgets.Paragraph
	lastOperation *widgets.Paragraph
)

// mu is used for synchronizing write operation for lastOperation variable (lastOperation.Text)
// this mutex is only used in the pushLastOperation func.
var (
	opMu = new(sync.Mutex)
)

func init() {
	views = make([]*view, 5)
	for i := range views {
		views[i] = &view{
			Table:     xtui.NewTable(),
			Header:    xtui.NewTable(),
			SList:     widgets.NewList(),
			sortBy:    NoColumn,
			ascending: true,
		}
		views[i].SList.Border = true
		views[i].SList.TextStyle = tui.NewStyle(tui.ColorWhite, tui.ColorBlue, tui.ModifierBold)
		views[i].SList.SelectedRowStyle = tui.NewStyle(tui.ColorYellow, tui.ColorBlue, tui.ModifierBold)
		views[i].SList.Title = "Sort by"

		views[i].Table.TextAlignment = tui.AlignLeft
		views[i].Table.Border = false
		views[i].Table.RowSeparator = false
		views[i].Table.FillRow = true

		views[i].Header.TextAlignment = tui.AlignLeft
		views[i].Header.Border = false
		views[i].Header.RowSeparator = false
		views[i].Header.FillRow = true
		views[i].Header.Colors.SelectedRowFg = tui.ColorWhite
		views[i].Header.Colors.SelectedRowBg = tui.ColorRed
	}
	views[Interfaces].Header.Rows = [][]string{{"Name", "Idx", "State", "MTU(L3/IP4/IP6/MPLS)", "RxCounters", "RxCount", "TxCounters", "TxCount", "Drops", "Punts", "IP4", "IP6"}}
	views[Errors].Header.Rows = [][]string{{"Counter", "Node", "Reason"}}
	views[Nodes].Header.Rows = [][]string{{"NodeName", "NodeIndex", "Clocks", "Vectors", "Calls", "Suspends", "Vectors/Calls"}}
	views[Memory].Header.Rows = [][]string{{"Thread/ID/Name", "Current memory usage per Thread"}}
	views[Threads].Header.Rows = [][]string{{"ID", "Name", "Type", "PID", "CPUID", "Core", "CPUSocket"}}

	views[Interfaces].Table.InitFilter(IfaceStatIfaceName, RowsPerIface) // names column for interfaces.
	views[Errors].Table.InitFilter(ErrorStatErrorNodeName, RowsPerError) // names column for errors.
	views[Nodes].Table.InitFilter(NodeStatNodeName, RowsPerNode)         // names column for nodes.
	views[Memory].Table.InitFilter(0, RowsPerMemory)                     // names column for memory.

	views[Memory].SList.Rows = []string{""}  // avoid panic
	views[Threads].SList.Rows = []string{""} // avoid panic
	// The order must match with values from sortstats.go !
	views[Errors].SList.Rows = []string{"Counter", "Node", "Reason"}
	// The order must match with values from sortstats.go !
	views[Interfaces].SList.Rows = []string{
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
	}
	// The order must match with values from sortstats.go !
	views[Nodes].SList.Rows = []string{
		"NodeName",
		"NodeIndex",
		"Clocks",
		"Vectors",
		"Calls",
		"Suspends",
		"Vectors/Calls",
	}

	tabPane = widgets.NewTabPane("Interfaces", "Nodes", "Errors", "Memory", "Threads")
	tabPane.SetRect(tabPaneTopX, tabPaneTopY, tabPaneBottomX, tabPaneBottomY)
	tabPane.Border = false

	version = widgets.NewParagraph()
	version.SetRect(versionTopX, versionTopY, versionBottomX, versionBottomY)
	version.Border = false
	version.WrapText = true

	filter = widgets.NewParagraph()
	filter.SetRect(filterTopX, filterTopY, filterBottomX, filterBottomY)
	filter.Border = false
	filter.WrapText = false
	filter.TextStyle = tui.NewStyle(tui.ColorWhite, tui.ColorBlue, tui.ModifierBold)

	filterExit = widgets.NewParagraph()
	filterExit.SetRect(filterExitTopX, filterExitTopY, filterExitBottomX, filterExitBottomY)
	filterExit.Border = false
	filterExit.WrapText = false
	filterExit.Text = fmt.Sprintf("Exit:%v filter:", keyCancel)
	filterExit.TextStyle = tui.NewStyle(tui.ColorWhite, tui.ColorBlue, tui.ModifierBold)

	// resized with window
	lastOperation = widgets.NewParagraph()
	lastOperation.Border = false
	lastOperation.WrapText = false
	lastOperation.TextStyle = tui.NewStyle(tui.ColorWhite, tui.ColorBlue, tui.ModifierBold)

	// Need to set size when rendered !
	exitScreen = widgets.NewParagraph()
	exitScreen.Border = false
	exitScreen.WrapText = true
	exitScreen.Text = "Closing connections.."
	exitScreen.TextStyle = tui.NewStyle(tui.ColorWhite, tui.ColorBlack, tui.ModifierBold)
}

// resizeWidgets resizes all the widgets (tables, headers, lists...).
// based on the width and height.
func resizeWidgets(w, h int) {
	// If an panic occurs this might be the main source.
	// Check if the number of columns of the table corresponds with the number
	// of columns provided below.
	var (
		nCol      = 4                      // number of columns to be resized with the window
		usedWidth = 82                     // total width used by the non resizing column
		cw        = (w - usedWidth) / nCol // new width for each column that is resized with the window
	)
	for i := range views {
		views[i].Table.SetRect(tableTopX, tableTopY, w, h-1)
		views[i].Header.SetRect(tableHeaderTopX, tableHeaderTopY, w, tableHeaderBottomY)
		views[i].SList.SetRect(sortPanelTopX, sortPanelTopY, sortPanelBottomX, h)
	}
	views[Interfaces].Table.Table.ColumnWidths = []int{24, 5, 5, 20, 10, 16, 11, 16, 11, 11, 11, (w - 140)}
	views[Interfaces].Header.Table.ColumnWidths = []int{24, 5, 5, 20, 10, 16, 11, 16, 11, 11, 11, (w - 140)}

	views[Nodes].Header.Table.ColumnWidths = []int{50, 10, cw, cw, cw, cw, 22}
	views[Nodes].Table.Table.ColumnWidths = []int{50, 10, cw, cw, cw, cw, 22}

	views[Memory].Header.Table.ColumnWidths = []int{30, w - 15}
	views[Memory].Table.Table.ColumnWidths = []int{30, w - 15}

	lastOperation.SetRect(tableTopX, h-2, 75, 75)
}

// pushLastOperation synchronizes writes to the lastOperation paragraph.
func pushLastOperation(opText string) {
	go func() {
		opMu.Lock()
		defer opMu.Unlock()

		lastOperation.Text = opText
		time.Sleep(time.Second * 1)
		lastOperation.Text = ""
	}()
}
