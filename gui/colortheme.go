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

// gui lightTheme settings.
var (
	lightTheme = tui.RootTheme{
		Default: tui.NewStyle(tui.ColorBlack),

		Block: tui.BlockTheme{
			Title:  tui.NewStyle(tui.ColorBlack),
			Border: tui.NewStyle(tui.ColorBlack),
		},

		BarChart: tui.BarChartTheme{
			Bars:   tui.StandardColors,
			Nums:   tui.StandardStyles,
			Labels: tui.StandardStyles,
		},

		Paragraph: tui.ParagraphTheme{
			Text: tui.NewStyle(tui.ColorBlack),
		},

		PieChart: tui.PieChartTheme{
			Slices: tui.StandardColors,
		},

		List: tui.ListTheme{
			Text: tui.NewStyle(tui.ColorBlack),
		},

		StackedBarChart: tui.StackedBarChartTheme{
			Bars:   tui.StandardColors,
			Nums:   tui.StandardStyles,
			Labels: tui.StandardStyles,
		},

		Gauge: tui.GaugeTheme{
			Bar:   tui.ColorBlack,
			Label: tui.NewStyle(tui.ColorBlack),
		},

		Sparkline: tui.SparklineTheme{
			Title: tui.NewStyle(tui.ColorBlack),
			Line:  tui.ColorBlack,
		},

		Plot: tui.PlotTheme{
			Lines: tui.StandardColors,
			Axes:  tui.ColorBlack,
		},

		Table: tui.TableTheme{
			Text: tui.NewStyle(tui.ColorBlack),
		},

		Tab: tui.TabTheme{
			Active:   tui.NewStyle(tui.ColorRed),
			Inactive: tui.NewStyle(tui.ColorBlack),
		},
	}

	textStyle        = tui.ColorWhite
	filterBackground = tui.ColorBlue
)

// SetLightTheme changes the basic colors of the tui lib to
// darker colors which are better visible on lighter background.
// This should be called before any tui widget created.
func SetLightTheme() {
	textStyle = tui.ColorBlack
	filterBackground = tui.ColorCyan
	tui.Theme = lightTheme
}
