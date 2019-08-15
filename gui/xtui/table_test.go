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

package xtui

import (
	"testing"
)

func TestTable_AppendToFilter(t *testing.T) {
	tests := []struct {
		T     *Table
		input string
		want  string
	}{
		{T: NewTable(false), input: "node", want: "node"},
		{T: NewTable(false), input: "", want: ""},
		{T: NewTable(false), input: "dpdk-65", want: "dpdk-65"},
		{T: NewTable(false), input: "arm-pc", want: "arm-pc"},
	}

	for _, test := range tests {
		test.T.AppendToFilter(test.input)

		got := test.T.filter.String()

		if got != test.want {
			t.Errorf("Error occured got:%v; want:%v", got, test.want)
		}
	}
}

func TestTable_ReduceFilter(t *testing.T) {
	tests := []struct {
		T     *Table
		input string
		n     int
		want  string
	}{
		{T: NewTable(false), input: "node", n: 1, want: "nod"},
		{T: NewTable(false), input: "", n: 1, want: ""},
		{T: NewTable(false), input: "", n: 2, want: ""},
		{T: NewTable(false), input: "", n: -1, want: ""},
		{T: NewTable(false), input: "arm-pc", n: 3, want: "arm"},
		{T: NewTable(false), input: "arm-pc", n: 5, want: "a"},
		{T: NewTable(false), input: "arm-pc", n: 6, want: ""},
		{T: NewTable(false), input: "arm-pc", n: 7, want: "arm-pc"},
	}

	for _, test := range tests {
		test.T.filter.WriteString(test.input)
		test.T.ReduceFilter(test.n)

		got := test.T.filter.String()

		if got != test.want {
			t.Errorf("Error occured got:%v; want:%v", got, test.want)
		}
	}
}

func TestTable_ScrollUp(t *testing.T) {
	tests := []struct {
		T           *Table
		visibleRows int
		// input
		curr   int
		prev   int
		offset int
		// output (want)
		wantCurr   int
		wantPrev   int
		wantOffset int
	}{
		{T: NewTable(false), visibleRows: 2, curr: 0, prev: 0, offset: 0, wantCurr: 0, wantPrev: 0, wantOffset: 0},
		{T: NewTable(false), visibleRows: 2, curr: 1, prev: 0, offset: 0, wantCurr: 0, wantPrev: 1, wantOffset: 0},
		{T: NewTable(false), visibleRows: 5, curr: 0, prev: 0, offset: 3, wantCurr: 0, wantPrev: 0, wantOffset: 2},
		{T: NewTable(false), visibleRows: 10, curr: 2, prev: 1, offset: 5, wantCurr: 1, wantPrev: 2, wantOffset: 5},
		{T: NewTable(false), visibleRows: 10, curr: 0, prev: 1, offset: 5, wantCurr: 0, wantPrev: 1, wantOffset: 4},
		{T: NewTable(false), visibleRows: 10, curr: 0, prev: 1, offset: 0, wantCurr: 0, wantPrev: 1, wantOffset: 0},
	}

	for _, test := range tests {
		test.T.visibleRows = test.visibleRows
		test.T.curr = test.curr
		test.T.prev = test.prev
		test.T.offset = test.offset

		test.T.ScrollUp()

		if test.T.curr != test.wantCurr {
			t.Errorf("Error occured curr do not match got:%v; want:%v\n", test.T.curr, test.wantCurr)
		}

		if test.T.prev != test.wantPrev {
			t.Errorf("Error occured prev do not match got:%v; want:%v\n", test.T.prev, test.wantPrev)
		}

		if test.T.offset != test.wantOffset {
			t.Errorf("Error occured offset do not match got:%v; want:%v\n", test.T.offset, test.wantOffset)
		}
	}
}

func TestTable_ScrollDown(t *testing.T) {
	tests := []struct {
		T           *Table
		visibleRows int
		out         TableRows
		// input
		curr   int
		prev   int
		offset int
		// output (want)
		wantCurr   int
		wantPrev   int
		wantOffset int
	}{
		{T: NewTable(false), visibleRows: 2, out: TableRows{{""}, {""}, {""}, {""}, {""}, {""}}, curr: 0, prev: 0, offset: 0, wantCurr: 1, wantPrev: 0, wantOffset: 0},
		{T: NewTable(false), visibleRows: 2, out: TableRows{{""}, {""}, {""}, {""}, {""}, {""}}, curr: 1, prev: 0, offset: 0, wantCurr: 1, wantPrev: 0, wantOffset: 1},
		{T: NewTable(false), visibleRows: 2, out: TableRows{{""}, {""}, {""}, {""}, {""}, {""}}, curr: 0, prev: 0, offset: 3, wantCurr: 1, wantPrev: 0, wantOffset: 3},
		{T: NewTable(false), visibleRows: 2, out: TableRows{{""}, {""}, {""}, {""}, {""}, {""}}, curr: 1, prev: 0, offset: 3, wantCurr: 1, wantPrev: 0, wantOffset: 4},
		{T: NewTable(false), visibleRows: 1, out: TableRows{{""}, {""}, {""}, {""}, {""}, {""}}, curr: 2, prev: 1, offset: 5, wantCurr: 2, wantPrev: 1, wantOffset: 5},
	}

	for _, test := range tests {
		test.T.visibleRows = test.visibleRows
		test.T.curr = test.curr
		test.T.prev = test.prev
		test.T.offset = test.offset
		test.T.out = test.out

		test.T.ScrollDown()

		if test.T.curr != test.wantCurr {
			t.Errorf("Error occured curr do not match got:%v; want:%v\n", test.T.curr, test.wantCurr)
		}

		if test.T.prev != test.wantPrev {
			t.Errorf("Error occured prev do not match got:%v; want:%v\n", test.T.prev, test.wantPrev)
		}

		if test.T.offset != test.wantOffset {
			t.Errorf("Error occured offset do not match got:%v; want:%v\n", test.T.offset, test.wantOffset)
		}
	}
}
