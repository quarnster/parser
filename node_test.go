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