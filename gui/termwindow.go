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

import (
	"fmt"
	"time"

	tui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

// viewType represents the current state of the gui.
// As of now it supports only 3 views.
// 1 - default (where only the tabPane Version, and tabViews are rendered).
// 2 - sort (where on top of the default widgets a sort panel is rendered).
// 3 - filter (where on top of the default widgets a filter is rendered).
type viewType uint

const (
	sort viewType = iota
	filter
	def
)

// TermWindow represents terminal gui that can handle up to multiple tabs
// with different content.
type TermWindow struct {
	// current state.
	view viewType

	// indexes for tabs which can be cleared.
	// (for these tabs a notification will be displayed).
	clearTabs []int

	// gui components.
	mainView TabView
	views    []TabView

	exitView     TabView
	sortPanel    *widgets.List
	tabPane      *widgets.TabPane
	filter       *widgets.Paragraph
	filterExit   *widgets.Paragraph
	version      *widgets.Paragraph
	notification *widgets.Paragraph

	// keybidings
	keybindings []*Binding

	timerDuration     time.Duration
	notificationTimer *time.Timer

	// channels & callbacks.
	stop chan struct{}

	windowEvents <-chan tui.Event
	refresh      <-chan time.Time

	onExit      func(Event)
	onSort      func(Event)
	onClear     func(Event)
	onTabswitch func(Event)
}

// NewTermWindow returns an instance of <*TermWindow>
// you can also set the theme of gui (however the gui cannot change the supplied views
// so it's up to the user to set the color of each view).
func NewTermWindow(refreshInterval time.Duration, views []TabView, viewNames []string, clearTabs []int, exitView TabView) *TermWindow {
	window := new(TermWindow)

	window.refresh = time.NewTicker(refreshInterval).C
	window.windowEvents = tui.PollEvents()
	window.stop = make(chan struct{})

	window.timerDuration = 1 * time.Second
	window.notificationTimer = time.NewTimer(window.timerDuration)

	window.keybindings = window.defaultKeybindings()
	window.view = def

	window.clearTabs = clearTabs

	window.views = views
	if len(window.views) != 0 {
		window.mainView = window.views[0]
	}

	window.exitView = exitView

	window.sortPanel = widgets.NewList()
	window.sortPanel.Border = true
	window.sortPanel.TextStyle = tui.NewStyle(textStyle, tui.ColorBlue, tui.ModifierBold)
	window.sortPanel.SelectedRowStyle = tui.NewStyle(tui.ColorYellow, tui.ColorBlue, tui.ModifierBold)
	window.sortPanel.Title = "Sort by"

	window.tabPane = widgets.NewTabPane(viewNames...)
	window.tabPane.SetRect(TabPaneTopX, TabPaneTopY, TabPaneBottomX, TabPaneBottomY)
	window.tabPane.Border = false

	window.filter = widgets.NewParagraph()
	window.filter.SetRect(FilterTopX, FilterTopY, FilterBottomX, FilterBottomY)
	window.filter.Border = false
	window.filter.WrapText = false
	window.filter.TextStyle = tui.NewStyle(textStyle, filterBackground, tui.ModifierBold)

	window.filterExit = widgets.NewParagraph()
	window.filterExit.SetRect(FilterExitTopX, FilterExitTopY, FilterExitBottomX, FilterExitBottomY)
	window.filterExit.Border = false
	window.filterExit.WrapText = false
	window.filterExit.Text = fmt.Sprintf("Exit:%v filter:", KeyCancel)
	window.filterExit.TextStyle = tui.NewStyle(textStyle, filterBackground, tui.ModifierBold)

	window.version = widgets.NewParagraph()
	window.version.SetRect(VersionTopX, VersionTopY, VersionBottomX, VersionBottomY)
	window.version.Border = false
	window.version.WrapText = true

	window.notification = widgets.NewParagraph()
	window.notification.Border = false
	window.notification.WrapText = false
	window.notification.TextStyle = tui.NewStyle(textStyle, tui.ColorBlue, tui.ModifierBold)

	widgets.NewTabPane()
	return window
}

// AddOnExitCallback registers a single function that will be called
// on gui exit.
func (w *TermWindow) AddOnExitCallback(f func(Event)) {
	w.onExit = f
}

// AddOnClearCallback registers a single function that will be called
// on clear event. The Event payload is the tab at which the event occurred.
func (w *TermWindow) AddOnClearCallback(f func(Event)) {
	w.onClear = f
}

// AddOnClearCallback registers a single function that will be called
// on sort event. The Event payload is of type SortMetadata.
func (w *TermWindow) AddOnSortCallback(f func(Event)) {
	w.onSort = f
}

// AddOnTabSwitchCallback registers a single function that will be called
// on each tab switch (left or right). The payload of the Event is the new
// tab index.
func (w *TermWindow) AddOnTabSwitchCallback(f func(Event)) {
	w.onTabswitch = f
}

// SetVersion sets the text to the version paragraph.
func (w *TermWindow) SetVersion(s string) {
	w.version.Text = s
}

// handleExit changes the main view to the exit screen, and notifies
// all listeners for the onExit event.
func (w *TermWindow) handleExit(event Event) {
	close(w.stop)
	w.mainView = w.exitView
	if w.onExit != nil {
		w.onExit(event)
	}
}

// pushNotification resets the timer for the displayed
// notification and updates the text.
func (w *TermWindow) pushNotification(text string) {
	isPresent := func(tabs []int, currTab int) bool {
		for _, tab := range tabs {
			if tab == currTab {
				return true
			}
		}
		return false
	}

	currTab := w.currentTab()
	if isPresent(w.clearTabs, currTab) {
		w.notificationTimer.Reset(w.timerDuration)
		w.notification.Text = text
	}
}

// handleSortMenu changes the main view to the sort menu.
func (w *TermWindow) handleSortMenu(_ Event) {
	w.view = sort
	w.sortPanel.Rows = w.mainView.ItemsList()
	w.keybindings = w.sortKeybindings()
}

// handleFilterMenu change the main view to the filter menu.
func (w *TermWindow) handleFilterMenu(_ Event) {
	w.view = filter
	w.keybindings = w.filterKeybindings()
}

// handleDefaultMenu changes the gui state back to the default.
func (w *TermWindow) handleDefaultMenu(event Event) {
	switch w.view {
	case sort:
		w.sortPanel.Rows = []string{""}
	case filter:
		w.filter.Text = ""
	}
	w.handleFilter(event)
}

// handleFilter changes the gui state to the default state.
func (w *TermWindow) handleFilter(_ Event) {
	w.view = def
	w.keybindings = w.defaultKeybindings()
}

// handleScroll is called when a scroll event occurs.
func (w *TermWindow) handleScroll(event Event) {
	w.mainView.OnScrollEvent(event)
}

// handlePreviousTab is called when a tab switch event occurs.
func (w *TermWindow) handleTabSwitch(event Event) {
	switch event.Payload.(string) {
	case KeyTabLeft:
		w.tabPane.FocusLeft()
	case KeyTabRight:
		w.tabPane.FocusRight()
	}
	if w.filter.Text != "" {
		w.filter.Text = ""
	}
	w.mainView = w.views[w.tabPane.ActiveTabIndex]
	w.onTabswitch(Event{
		Payload: w.tabPane.ActiveTabIndex,
	})
}

// handleClear is called when an on clear event occurs.
func (w *TermWindow) handleClear(_ Event) {
	currTab := w.currentTab()
	w.pushNotification(fmt.Sprintf("clearing tab: %s", w.tabPane.TabNames[currTab]))
	if w.onClear != nil {
		w.onClear(Event{
			Payload: currTab,
		})
	}
}

// handleReduceFilter is called when the users shortens the filter.
func (w *TermWindow) handleReduceFilter(_ Event) {
	if len(w.filter.Text) != 0 {
		w.filter.Text = w.filter.Text[:len(w.filter.Text)-1]
	}
}

// handleAppendToFilter is called when the users appends to the filter.
func (w *TermWindow) handleAppendToFilter(event Event) {
	payload := event.Payload.(string)
	if payload == "<Space>" {
		payload = " "
	}
	w.filter.Text = w.filter.Text + payload
}

// handleSort is called when an sort event occurs.
func (w *TermWindow) handleSort(_ Event) {
	if w.onSort != nil {
		w.onSort(Event{
			Payload: SortMetadata{
				CurrRow: w.sortPanel.SelectedRow,
				CurrTab: w.currentTab(),
			},
		})
	}
}

// handleSortPanelScrollDown is called in sort state of the gui
// to scroll the sort panel
func (w *TermWindow) handleSortPanelScroll(event Event) {
	switch event.Payload.(string) {
	case KeyScrollUp:
		w.sortPanel.ScrollUp()
	case KeyScrollDown:
		if len(w.sortPanel.Rows) != 0 {
			w.sortPanel.ScrollDown()
		}
	case KeyPgdn:
		w.sortPanel.SelectedRow = 0
	case KeyPgup:
		w.sortPanel.SelectedRow = len(w.sortPanel.Rows) - 1
	}
}

// processInput is called when a keyboard event occurs.
func (w *TermWindow) processInput(key string) {
	isPresent := func(bindings []*Binding, key string) bool {
		for i := range bindings {
			if bindings[i].key == key {
				return true
			}
		}
		return false
	}

	if w.view == filter && !isPresent(w.keybindings, key) {
		w.keybindings[len(w.keybindings)-1].callback(Event{
			Payload: key,
		})
		return
	}

	for _, keybinding := range w.keybindings {
		if keybinding.key == key {
			keybinding.callback(Event{
				Payload: key,
			})
		}
	}
}

// render is called on gui refresh.
func (w *TermWindow) render() {
	widgts := []tui.Drawable{
		w.tabPane,
		w.version,
		w.notification,
	}

	if w.mainView != nil {
		w.mainView.Filter(Event{
			Payload: w.filter.Text,
		})
		widgts = append(widgts, w.mainView.Widgets()...)

		switch w.view {
		case sort:
			widgts = append(widgts, w.sortPanel)
		case filter:
			widgts = append(widgts, w.filter, w.filterExit)
		}
	}
	tui.Clear()
	tui.Render(widgts...)
}

// Init initializes the gui.
func (w *TermWindow) Init() error {
	if err := tui.Init(); err != nil {
		return fmt.Errorf("error occured while initializing tui: %v", err)
	}
	return nil
}

// CurrentTab returns the index of the current tab.
func (w *TermWindow) currentTab() int {
	return w.tabPane.ActiveTabIndex
}

// ViewAtTab returns the tableView at index.
// if out of bounds panics.
func (w *TermWindow) ViewAtTab(i int) TabView {
	return w.views[i]
}

// Start starts the gui main loop to listen for event.
// The gui starts rendering the view at index 0.
// if no view is present panics.
func (w *TermWindow) Start() {
	w.resize(tui.TerminalDimensions())
	for {
		select {
		case <-w.refresh:
			w.render()
		case e := <-w.windowEvents:
			switch e.Type {
			case tui.KeyboardEvent:
				w.processInput(e.ID)
			case tui.ResizeEvent:
				payload := e.Payload.(tui.Resize)
				w.resize(payload.Width, payload.Height)
			}
		case <-w.notificationTimer.C:
			w.notification.Text = ""
		case <-w.stop:
			return
		}
	}
}

// Destroy de-initializes gui.
func (w *TermWindow) Destroy() {
	tui.Close()
}

// resize resizes all widgets.
func (w *TermWindow) resize(width, height int) {
	for i := range w.views {
		w.views[i].Resize(width, height)
	}
	w.exitView.Resize(width, height)
	w.sortPanel.SetRect(SortPanelTopX, SortPanelTopY, SortPanelBottomX, height)
	w.notification.SetRect(SortPanelTopX, height-2, NotificationBottomX, NotificationBottomY)
}
