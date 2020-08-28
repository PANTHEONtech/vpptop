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
	tui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

// exitView represents a view that is rendered on gui exit.
type exitView struct {
	exitScreen *widgets.Paragraph
}

// Resize resizes the exitView.
func (v *exitView) Resize(w, h int) {
	x1 := w/2 - w/4
	y1 := h/2 - h/4

	x2 := w/2 + w/4
	y2 := h/2 + h/4

	// center the text inside the paragraph
	v.exitScreen.PaddingLeft = x2/4 - 3
	v.exitScreen.PaddingTop = y2/4 - 1

	v.exitScreen.SetRect(x1, y1, x2, y2)
}

// NewExitView returns an instance of <*exitView>
func NewExitView() *exitView {
	s := &exitView{
		exitScreen: widgets.NewParagraph(),
	}
	s.exitScreen.Border = false
	s.exitScreen.WrapText = true
	s.exitScreen.Text = "Closing.."
	s.exitScreen.TextStyle = tui.NewStyle(tui.ColorWhite, tui.ColorBlack, tui.ModifierBold)

	return s
}

// Drawables returns the widget to be drawn.
func (v *exitView) Widgets() []tui.Drawable { return []tui.Drawable{v.exitScreen} }

// These functions do nothing.
func (v *exitView) Update(interface{})            {}
func (v *exitView) ItemsList() []string           { return nil }
func (v *exitView) Filter(gui.Event)              {}
func (v *exitView) OnScrollEvent(_ gui.Event) {}
