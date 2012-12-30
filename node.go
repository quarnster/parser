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
	p        *Parser
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
	return n.p.Data(n.Range.Start, n.Range.End)
}

func (n *Node) String() string {
	cf := CodeFormatter{}
	n.format(&cf)
	return cf.String()
}

func (n *Node) Cleanup(pos, end int) *Node {
	var popped Node
	popped.Range = Range{pos, end}
	for i := len(n.Children) - 1; i >= 0; i-- {
		//			fmt.Println("i:", i, ", l:", len(n.Children))
		node := n.Children[i]
		if node.Range.Start >= pos || node.Range.End > pos {
			popped.Append(node)
			if i > 0 {
				n.Children = n.Children[:i]
			} else {
				n.Children = n.Children[:0]
			}
		} else {
			break
		}
	}
	// Since we pushed from the back, reverse the list
	for i, j := 0, len(popped.Children)-1; i < j; i, j = i+1, j-1 {
		popped.Children[i], popped.Children[j] = popped.Children[j], popped.Children[i]
	}
	return &popped
}

func (n *Node) Append(child *Node) {
	n.Children = append(n.Children, child)
}
