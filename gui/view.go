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
	tui "github.com/gizak/termui/v3"
)

type (
	// TabView is the interface that wraps all methods needed by gui.
	TabView interface {
		// Filter filters the view. The event passed to the function.
		// contains the full text of the new filter.
		Filter(Event)

		// OnScrollEvent is called when an scroll event occurs. The key on which
		// the scroll event occurred is passed to the function.
		OnScrollEvent(Event)

		// Update is never called by the termWindow. It should only by used
		// if the user wants to update the contents of the view.
		// If this function is called from a different go-routine than the termWindow,
		// then this function should use a lock when updating the contents.
		// The lock should be the same as the widget that is to be rendered.
		// Every widget in termui comes with a lock (which is used when the widget
		// is rendered). The same lock should be used in the update function.
		Update(interface{})

		// Resize should resize the view. The new terminal width and height
		// are passed as arguments to the function.
		Resize(int, int)

		// Widgets returns the items to be rendered by the gui.
		// The widgets will be rendered in the order as they were passed,
		// i.e. widget at index 0 is rendered first, widget at last index
		// is rendered as last.
		Widgets() []tui.Drawable

		// ItemsList returns the list of items to be sorted.
		ItemsList() []string
	}
)