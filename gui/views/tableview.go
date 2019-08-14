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

package views

import (
	"github.com/PantheonTechnologies/vpptop/gui"
	"github.com/PantheonTechnologies/vpptop/gui/xtui"
	tui "github.com/gizak/termui/v3"
)

const (
	// TableColResizedWithWindow represent that the column
	// of the tableView should be resized with the terminal window.
	TableColResizedWithWindow = -1
)

// Table positions to match sort panel
// provided by the gui.
const (
	tableTopX = gui.SortPanelTopX
	tableTopY = gui.SortPanelTopY

	tableHeaderTopX    = tableTopX
	tableHeaderTopY    = gui.TabPaneBottomY + 1
	tableHeaderBottomY = gui.TabPaneBottomY + 4
)

// TableView implements the view interface. It is a table build on xtui.Table.
type TableView struct {
	table  *xtui.Table
	header *xtui.Table

	itemsList []string
	colWidth  []int

	tw      int
	resized []int
}

// NewTableView returns a new instance of <*TableView>
func NewTableView(itemsList []string, headerRows xtui.TableRows, filterCol, rowsPerEntry int, colWidths []int, light bool) *TableView {
	v := &TableView{
		table:     xtui.NewTable(light),
		header:    xtui.NewTable(light),
		itemsList: itemsList,
	}
	v.table.TextAlignment = tui.AlignLeft
	v.table.Border = false
	v.table.RowSeparator = false
	v.table.FillRow = true

	v.header.TextAlignment = tui.AlignLeft
	v.header.Border = false
	v.header.RowSeparator = false
	v.header.FillRow = true
	v.header.Colors.SelectedRowFg = tui.ColorWhite
	v.header.Colors.SelectedRowBg = tui.ColorRed

	v.header.Rows = headerRows

	v.table.InitFilter(filterCol, rowsPerEntry)

	v.colWidth = colWidths

	for i, val := range v.colWidth {
		if val == TableColResizedWithWindow {
			v.resized = append(v.resized, i)
		} else {
			v.tw += v.colWidth[i]
		}
	}
	return v
}

// Resize resizes the tableView.
func (v *TableView) Resize(w, h int) {
	v.table.SetRect(tableTopX, tableTopY, w, h-1)
	v.header.SetRect(tableHeaderTopX, tableHeaderTopY, w, tableHeaderBottomY)

	if v.colWidth != nil {
		cw := (w - v.tw) / len(v.resized)

		for _, i := range v.resized {
			v.colWidth[i] = cw
		}

		v.table.Table.ColumnWidths = v.colWidth
		v.header.Table.ColumnWidths = v.colWidth
	}
}

// Filter applies the filter from the gui.Event to the xtui.Table.
func (v *TableView) Filter(event gui.Event) {
	filter := event.Payload.(string)
	newLen := len(filter)
	oldLen := len(v.table.Filter())

	if newLen < oldLen {
		v.table.ReduceFilter(oldLen - newLen)
	} else if newLen > oldLen {
		v.table.AppendToFilter(filter[oldLen:])
	}
}

// OnScrollEvent handles the scroll event based on the key pressed.
func (v *TableView) OnScrollEvent(event gui.Event) {
	switch event.Payload.(string) {
	case gui.KeyScrollUp:
		v.table.ScrollUp()
	case gui.KeyScrollDown:
		v.table.ScrollDown()
	case gui.KeyPgdn:
		v.table.PageDown()
	case gui.KeyPgup:
		v.table.PageUp()
	}
}

// Update updates the table rows.
// The lock from the table is used.
func (v *TableView) Update(payload interface{}) {
	rows := payload.(xtui.TableRows)

	v.table.Lock()
	v.table.Rows = rows
	v.table.Unlock()

}

// Widgets returns all widgets to be drawed by this view.
func (v *TableView) Widgets() []tui.Drawable { return []tui.Drawable{v.table, v.header} }
// ItemsList returns a list with names based on which the table can be sorted.
func (v *TableView) ItemsList() []string     { return v.itemsList }
