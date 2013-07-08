/*
Copyright (c) 2012-2013 Fredrik Ehnbom
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this
   list of conditions and the following disclaimer.
2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/
package parser

import (
	"reflect"
	"testing"
)

func TestRangeClip(t *testing.T) {
	tests := [][]Range{
		[]Range{{10, 20}, {0, 5}, {10, 20}},
		[]Range{{10, 20}, {0, 11}, {11, 20}},
		[]Range{{10, 20}, {0, 15}, {15, 20}},
		[]Range{{10, 20}, {15, 30}, {10, 15}},
		[]Range{{10, 20}, {20, 30}, {10, 20}},
		[]Range{{10, 20}, {0, 30}, {10, 20}},
		[]Range{{10, 20}, {10, 20}, {10, 20}},
	}
	for i := range tests {
		a := tests[i][0]
		ignoreRange := tests[i][1]
		a.Clip(ignoreRange)
		if a != tests[i][2] {
			t.Errorf("Expected %v, got: %v", a, tests[i][2])
		}
	}
}

func TestRangeInside(t *testing.T) {
	r := Range{10, 20}
	tests := []struct {
		In  int
		Exp bool
	}{
		{8, false},
		{9, false},
		{10, true},
		{18, true},
		{19, true},
		{20, false},
	}

	for i, test := range tests {
		if res := r.Inside(test.In); res != test.Exp {
			t.Errorf("Test %d; Expected %v, got: %v", i, test.Exp, res)
		}
	}
}

func TestRangeJoin(t *testing.T) {
	tests := []struct {
		A, B, Out Range
		Res       bool
	}{
		{Range{10, 20}, Range{0, 5}, Range{10, 20}, false},
		{Range{10, 20}, Range{0, 11}, Range{0, 20}, true},
		{Range{10, 20}, Range{5, 15}, Range{5, 20}, true},
		{Range{10, 20}, Range{15, 30}, Range{10, 30}, true},
		{Range{10, 20}, Range{20, 30}, Range{10, 30}, true},
		{Range{10, 20}, Range{21, 30}, Range{10, 20}, false},
		{Range{10, 20}, Range{10, 30}, Range{10, 30}, true},
		{Range{10, 30}, Range{10, 20}, Range{10, 30}, true},
	}
	for i, test := range tests {
		if res := test.A.Join(test.B); !reflect.DeepEqual(test.A, test.Out) {
			t.Errorf("Test %d; Expected %v, got: %v", i, test.Out, test.A)
		} else if res != test.Res {
			t.Errorf("Test %d; Expected %v, got: %v", i, test.Res, res)
		}
	}
}

func TestRangeXor(t *testing.T) {
	tests := []struct {
		A, B Range
		Out  []Range
	}{
		{Range{10, 20}, Range{0, 5}, []Range{{10, 20}}},
		{Range{10, 20}, Range{12, 15}, []Range{{10, 12}, {15, 20}}},
		{Range{10, 20}, Range{5, 15}, []Range{{15, 20}}},
		{Range{10, 20}, Range{15, 20}, []Range{{10, 15}}},
	}
	for i, test := range tests {
		if res := test.A.Xor(test.B); !reflect.DeepEqual(res, test.Out) {
			t.Errorf("Test %d; Expected %v, got: %v", i, test.Out, res)
		}
	}
}

func TestRangeSetXor(t *testing.T) {
	tests := []struct {
		A, B Range
		Out  RangeSet
	}{
		{Range{10, 20}, Range{0, 5}, []Range{{10, 20}}},
		{Range{10, 20}, Range{12, 15}, []Range{{10, 12}, {15, 20}}},
		{Range{10, 20}, Range{5, 15}, []Range{{15, 20}}},
		{Range{10, 20}, Range{15, 20}, []Range{{10, 15}}},
	}
	for i, test := range tests {
		var rs RangeSet
		rs.Add(test.A)
		t.Log(rs)
		if res := rs.Xor(test.B); !reflect.DeepEqual(res, test.Out) {
			t.Errorf("Test %d; Expected %v, got: %v", i, test.Out, res)
		}
	}
}

func TestRangeSetAdd(t *testing.T) {
	tests := []struct {
		A   RangeSet
		B   Range
		Out RangeSet
	}{
		{[]Range{{10, 20}}, Range{0, 5}, []Range{{0, 5}, {10, 20}}},
		{[]Range{{10, 20}}, Range{12, 15}, []Range{{10, 20}}},
		{[]Range{{10, 20}}, Range{5, 15}, []Range{{5, 20}}},
		{[]Range{{10, 20}}, Range{15, 25}, []Range{{10, 25}}},
		{[]Range{{10, 15}, {20, 25}}, Range{12, 23}, []Range{{10, 25}}},
	}
	for i, test := range tests {
		if test.A.Add(test.B); !reflect.DeepEqual(test.A, test.Out) {
			t.Errorf("Test %d; Expected %v, got: %v", i, test.Out, test.A)
		}
	}
}

type ds int

func (s ds) Data(start, end int) string {
	return ""
}

func TestNodeClone(t *testing.T) {
	var s ds
	n := Node{Name: "Test", P: s}
	n.Children = append(n.Children, &Node{Name: "1", P: s}, &Node{Name: "2", P: s}, &Node{Name: "3", P: s})
	n2 := n.Clone()
	n2.Children[1].Children = append(n2.Children[1].Children, &Node{Name: "a", P: s})
	if a, b := n.String(), n2.String(); a == b {
		t.Error("Shouldn't be equal", a, b)
	} else if a, b := n.String(), n.Clone().String(); a != b {
		t.Error("Should be equal", a, b)
	}
}
