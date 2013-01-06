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
	"fmt"
)

type Range struct {
	Start, End int
}

type DataSource interface {
	Data(start, end int) string
}

func (r *Range) Clip(r2 Range) (clipped bool) {
	if r.Start >= r2.Start && r.Start < r2.End {
		clipped = true
		r.Start = r2.End
	}
	if r.End >= r2.Start && r.End <= r2.End {
		clipped = true
		r.End = r2.Start
	}
	if r.End < r.Start {
		r.End = r.Start
	}
	return clipped
}

type Node struct {
	Range    Range
	Name     string
	Children []*Node
	P        DataSource
}

func (n *Node) format(indent string) string {
	if len(n.Children) == 0 {
		return indent + fmt.Sprintf("%d-%d: \"%s\" - Data: \"%s\"\n", n.Range.Start, n.Range.End, n.Name, n.Data())
	}
	ret := indent + fmt.Sprintf("%d-%d: \"%s\"\n", n.Range.Start, n.Range.End, n.Name)
	indent += "\t"
	for _, child := range n.Children {
		ret += child.format(indent)
	}
	return ret
}

func (n *Node) Data() string {
	return n.P.Data(n.Range.Start, n.Range.End)
}

func (n *Node) String() string {
	return n.format("")
}

func (n *Node) Discard(pos int) {
	if pos == 0 {
		return
	}
	back := len(n.Children)
	popIdx := 0
	for i := back - 1; i >= 0; i-- {
		node := n.Children[i]
		if node.Range.End <= pos {
			popIdx = i + 1
			break
		}
	}
	if popIdx != back {
		n.Children = n.Children[:popIdx]
	}
}
func (n *Node) Cleanup(pos, end int) *Node {
	var popped Node
	popped.Range = Range{pos, end}
	back := len(n.Children)
	popIdx := 0
	popEnd := back
	if end == 0 {
		end = -1
	}
	if pos == 0 {
		pos = -1
	}

	for i := back - 1; i >= 0; i-- {
		node := n.Children[i]
		if node.Range.End <= pos {
			popIdx = i + 1
			break
		}
		if node.Range.Start > end {
			popEnd = i + 1
		}
	}

	if popEnd != 0 {
		popped.Children = n.Children[popIdx:popEnd]
	}
	if popIdx != back {
		n.Children = n.Children[:popIdx]
	}
	return &popped
}

func (n *Node) Append(child *Node) {
	n.Children = append(n.Children, child)
}
