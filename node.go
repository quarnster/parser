/*
Copyright (c) 2012 Fredrik Ehnbom

This software is provided 'as-is', without any express or implied
warranty. In no event will the authors be held liable for any damages
arising from the use of this software.

Permission is granted to anyone to use this software for any purpose,
including commercial applications, and to alter it and redistribute it
freely, subject to the following restrictions:

   1. The origin of this software must not be misrepresented; you must not
   claim that you wrote the original software. If you use this software
   in a product, an acknowledgment in the product documentation would be
   appreciated but is not required.

   2. Altered source versions must be plainly marked as such, and must not be
   misrepresented as being the original software.

   3. This notice may not be removed or altered from any source
   distribution.
*/
package parser

import (
	"fmt"
)

type Range struct {
	Start, End int
}

type Node struct {
	Range    Range
	Name     string
	Children []*Node
	P        *Parser
}

func (n *Node) format(cf *CodeFormatter) {
	cf.Add(fmt.Sprintf("%d-%d: \"%s\" - Data: \"%s\"\n", n.Range.Start, n.Range.End, n.Name, n.Data()))
	cf.Inc()
	for _, child := range n.Children {
		child.format(cf)
	}
	cf.Dec()
}

func (n *Node) Data() string {
	return n.P.Data(n.Range.Start, n.Range.End)
}

func (n *Node) String() string {
	cf := CodeFormatter{}
	n.format(&cf)
	return cf.String()
}

func (n *Node) Cleanup(pos, end int) *Node {
	var popped Node
	popped.Range = Range{pos, end}
	back := len(n.Children)
	popIdx := 0

	for i := back - 1; i >= 0; i-- {
		node := n.Children[i]
		if node.Range.End <= pos && pos != 0 {
			popIdx = i + 1
			break
		}
	}

	popped.Children = n.Children[popIdx:back]
	n.Children = n.Children[:popIdx]
	return &popped
}

func (n *Node) Append(child *Node) {
	n.Children = append(n.Children, child)
}
