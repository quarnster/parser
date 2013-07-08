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
	"sort"
)

type (
	Range struct {
		Start, End int
	}
	RangeSet []Range

	DataSource interface {
		Data(start, end int) string
	}

	Node struct {
		Range    Range
		Name     string
		Children []*Node
		P        DataSource
	}
)

func (r *Range) Intersect(other Range) (ret Range, intersects bool) {
	if r.Inside(other.Start) {
		ret.Start = other.Start
		if r.End < other.End {
			ret.End = r.End
		} else {
			ret.End = other.End
		}
		intersects = true
	} else if r.Inside(other.End) {
		ret.End = other.End
		if r.Start > other.Start {
			ret.Start = r.Start
		} else {
			ret.Start = other.Start
		}
		intersects = true
	}
	return
}

func (r *Range) Contains(other Range) bool {
	return r.Start <= other.Start && r.End >= other.End
}

func (r *Range) Inside(x int) bool {
	return x >= r.Start && x < r.End
}

func (r *Range) Join(other Range) (could_join bool) {
	mi, ma := r, &other
	if mi.Start > ma.Start {
		mi, ma = ma, mi
	}
	if could_join = mi.End == ma.Start || mi.Inside(ma.Start); could_join {
		r.Start = mi.Start
		if mi.End < ma.End {
			r.End = ma.End
		} else {
			r.End = mi.End
		}
	}
	return
}

func (r *Range) Clip(ignore Range) (clipped bool) {
	if ignore.Contains(*r) {
		// this range is a subrange within the ignore range
		return false
	}
	if r.Start >= ignore.Start && r.Start < ignore.End {
		clipped = true
		r.Start = ignore.End
	}
	if r.End >= ignore.Start && r.End <= ignore.End {
		clipped = true
		r.End = ignore.Start
	}
	if r.End < r.Start {
		r.End = r.Start
	}
	return clipped
}

func (r *Range) Xor(other Range) (ret []Range) {
	if r.Inside(other.Start) {
		ret = append(ret, Range{r.Start, other.Start})
	}
	if r.Inside(other.End) {
		ret = append(ret, Range{other.End, r.End})
	}
	if len(ret) == 0 && !other.Contains(*r) {
		ret = append(ret, *r)
	}
	return
}

func (r *RangeSet) Len() int {
	return len(*r)
}
func (r *RangeSet) Swap(i, j int) {
	(*r)[i], (*r)[j] = (*r)[j], (*r)[i]
}
func (r *RangeSet) Less(i, j int) bool {
	return (*r)[i].Start < (*r)[j].Start
}

func (r *RangeSet) Add(r2 Range) {
	added := false
	for i := 0; i < len(*r); i++ {
		if (*r)[i].Join(r2) {
			added = true
		}
		for ; added && i+1 < r.Len() && (*r)[i].Join((*r)[i+1]); i++ {
			*r = append((*r)[:i+1], (*r)[i+2:]...)
		}
	}
	if !added {
		*r = append(*r, r2)
		sort.Sort(r)
	}
}

func (r RangeSet) Xor(r2 Range) (ret RangeSet) {
	for i := 0; i < len(r); i++ {
		for _, xor := range r[i].Xor(r2) {
			ret.Add(xor)
		}
	}
	return
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
		c := make([]*Node, len(popped.Children))
		copy(c, popped.Children)
		popped.Children = c
	}
	if popIdx != back {
		n.Children = n.Children[:popIdx]
	}
	return &popped
}

func (n *Node) Clone() *Node {
	ret := *n
	ret.Children = make([]*Node, len(n.Children))
	for i := range n.Children {
		ret.Children[i] = n.Children[i].Clone()
	}
	return &ret
}

func (n *Node) UpdateRange() Range {
	for _, child := range n.Children {
		curr := child.UpdateRange()
		if curr.Start < n.Range.Start {
			n.Range.Start = curr.Start
		}
		if curr.End > n.Range.End {
			n.Range.End = curr.End
		}
	}
	return n.Range
}

func (n *Node) Append(child *Node) {
	n.Children = append(n.Children, child)
}

func (n *Node) Simplify() {
	for _, child := range n.Children {
		child.Simplify()
	}
	if len(n.Children) == 1 && n.Children[0].Range == n.Range {
		*n = *n.Children[0]
	}
}
