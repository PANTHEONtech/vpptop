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

package gui

// gui supported keys.
const (
	KeyTabLeft    = "<Left>"
	KeyTabRight   = "<Right>"
	KeyScrollDown = "<Down>"
	KeyScrollUp   = "<Up>"
	KeyQuit       = "q"
	KeyFilter     = "/"
	KeyCancel     = "<Escape>"
	KeyDeleteChar = "<Backspace>"
	KeyF1         = "<F1>"
	KeyF2         = "<F2>"
	KeyF3         = "<F3>"
	KeyF4         = "<F4>"
	KeyF5         = "<F5>"
	KeyF6         = "<F6>"
	KeyF7         = "<F7>"
	KeyF8         = "<F8>"
	KeyF9         = "<F9>"
	KeyF10        = "<F10>"
	KeyF11        = "<F11>"
	KeyF12        = "<F12>"
	KeyInsert     = "<Insert>"
	KeyDelete     = "<Delete>"
	KeyHome       = "<Home>"
	KeyEnd        = "<End>"
	KeyPgup       = "<PageUp>"
	KeyPgdn       = "<PageDown>"
	KeyCtrlSpace  = "<C-<Space>>"
	KeyCtrlA      = "<C-a>"
	KeyCtrlB      = "<C-b>"
	KeyCtrlC      = "<C-c>"
	KeyCtrlD      = "<C-d>"
	KeyCtrlE      = "<C-e>"
	KeyCtrlF      = "<C-f>"
	KeyCtrlG      = "<C-g>"
	KeyBackspace  = "<C-<Backspace>>"
	KeyTab        = "<Tab>"
	KeyCtrlJ      = "<C-j>"
	KeyCtrlK      = "<C-k>"
	KeyCtrlL      = "<C-l>"
	KeyEnter      = "<Enter>"
	KeyCtrlN      = "<C-n>"
	KeyCtrlO      = "<C-o>"
	KeyCtrlP      = "<C-p>"
	KeyCtrlQ      = "<C-q>"
	KeyCtrlR      = "<C-r>"
	KeyCtrlS      = "<C-s>"
	KeyCtrlT      = "<C-t>"
	KeyCtrlU      = "<C-u>"
	KeyCtrlV      = "<C-v>"
	KeyCtrlW      = "<C-w>"
	KeyCtrlX      = "<C-x>"
	KeyCtrlY      = "<C-y>"
	KeyCtrlZ      = "<C-z>"
	KeyCtrl4      = "<C-4>"
	KeyCtrl5      = "<C-5>"
	KeyCtrl6      = "<C-6>"
	KeyCtrl7      = "<C-7>"
	Any           = "<Any>"
)

// Binding encapsulates a keybinding with its given callback function.
type Binding struct {
	key      string
	callback func(Event)
}

// DefaultKeybindings are keybindings for the default view.
func (w *TermWindow) defaultKeybindings() []*Binding {
	return []*Binding{
		{key: KeyQuit, callback: w.handleExit},
		{key: KeyCtrlSpace, callback: w.handleSortMenu},
		{key: KeyScrollDown, callback: w.handleScroll},
		{key: KeyScrollUp, callback: w.handleScroll},
		{key: KeyPgup, callback: w.handleScroll},
		{key: KeyPgdn, callback: w.handleScroll},
		{key: KeyTabLeft, callback: w.handleTabSwitch},
		{key: KeyTabRight, callback: w.handleTabSwitch},
		{key: KeyFilter, callback: w.handleFilterMenu},
		{key: KeyCtrlC, callback: w.handleClear},
	}
}

// FilterKeybindings are keybindings for the filter view.
func (w *TermWindow) filterKeybindings() []*Binding {
	return []*Binding{
		{key: KeyCancel, callback: w.handleFilter},
		{key: KeyScrollUp, callback: w.handleDefaultMenu},
		{key: KeyScrollDown, callback: w.handleDefaultMenu},
		{key: KeyTabLeft, callback: w.handleDefaultMenu},
		{key: KeyTabRight, callback: w.handleDefaultMenu},
		{key: KeyEnter, callback: w.handleFilter},
		{key: KeyTab, callback: w.handleDefaultMenu},
		{key: KeyDeleteChar, callback: w.handleReduceFilter},
		{key: Any, callback: w.handleAppendToFilter},
	}
}

// SortKeybindings are keybindings for the sort view.
func (w *TermWindow) sortKeybindings() []*Binding {
	return []*Binding{
		{key: KeyCancel, callback: w.handleDefaultMenu},
		{key: KeyCtrlSpace, callback: w.handleDefaultMenu},
		{key: KeyEnter, callback: w.handleSort},
		{key: KeyScrollDown, callback: w.handleSortPanelScroll},
		{key: KeyScrollUp, callback: w.handleSortPanelScroll},
		{key: KeyPgup, callback: w.handleSortPanelScroll},
		{key: KeyPgdn, callback: w.handleSortPanelScroll},
	}
}
