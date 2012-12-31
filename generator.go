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
	"container/list"
	"strings"
)

var (
	convMap    = make(map[string]*Node)
	visitedMap = make(map[string]bool)
	blocks     list.List
)

type CodeFormatter struct {
	level string
	data  string
}

func (i *CodeFormatter) Level() string {
	return i.level
}
func (i *CodeFormatter) Inc() {
	i.level += "\t"
	l := len(i.data)
	if l > 0 && (i.data[l-1] == '\t' || i.data[l-1] == '\n') {
		i.data += "\t"
	}
}
func (i *CodeFormatter) Dec() {
	i.level = i.level[:len(i.level)-1]
	l := len(i.data)
	if l > 0 && i.data[l-1] == '\t' {
		i.data = i.data[:l-1]
	}
}
func (i *CodeFormatter) Add(add string) {
	i.data += strings.Replace(add, "\n", "\n"+i.level, -1)
}

func (i *CodeFormatter) String() string {
	return i.data
}

type Group interface {
	Add(value string)
}

type baseGroup struct {
	cf CodeFormatter
}

func (b *baseGroup) Add(value string) {
	b.cf.Add(value + ",\n")
}

type Generator interface {
	Return(value string) string
	Call(value string) string
	MakeFunction(value string) string
	MakeParserFunction(definitionNode *Node)
	MakeParserCall(value string) string
	CheckInRange(a, b string) string
	CheckInSet(a string) string
	CheckAnyChar() string
	CheckNext(a string) string
	AssertNot(a string) string
	AssertAnd(a string) string
	ZeroOrMore(a string) string
	OneOrMore(a string) string
	Maybe(a string) string
	BeginGroup(requireAll bool) Group
	EndGroup(g Group) string
	Begin()
	Finish() string
}
type CustomAction struct {
	Name   string
	Action string
}

func helper(gen Generator, node *Node) (retstring string) {
	makeReturn := func(value string) string {
		return gen.Return(value)
	}
	makeCall := func(value string) string {
		return gen.Call(value)
	}
	makeComplex := func(value string) string {
		return gen.MakeFunction(value)
	}
	makeComplexReturn := func(value string) string {
		return makeComplex(makeReturn(value))
	}
	switch node.Name {
	case "Class":
		others := ""
		var exps []string
		for _, child := range node.Children {
			if child.Name == "Range" {
				if len(child.Children) == 2 {
					exps = append(exps, gen.CheckInRange(child.Children[0].Data(), child.Children[1].Data()))
				} else {
					others += child.Data()
				}
			}
		}
		if others != "" {
			exps = append(exps, gen.CheckInSet(others))
		}
		if len(exps) > 1 {
			g := gen.BeginGroup(false)
			for _, e := range exps {
				g.Add(makeComplexReturn(e))
			}
			return gen.EndGroup(g)
		} else {
			return exps[0]
		}
	case "DOT":
		return gen.CheckAnyChar()
	case "Identifier":
		back := node.Children[len(node.Children)-1]
		data := node.Data()
		if back.Name == "Spacing" {
			data = data[:strings.LastIndex(data, back.Data())]
		}
		return data
	case "Literal":
		back := node.Children[len(node.Children)-1]
		data := node.Data()
		if back.Name == "Spacing" {
			data = data[:strings.LastIndex(data, back.Data())]
		}
		return gen.CheckNext(data)
	case "Expression":
		if len(node.Children) == 1 {
			return helper(gen, node.Children[0])
		} else {
			g := gen.BeginGroup(false)
			for i, child := range node.Children {
				if i&1 != 0 {
					continue
				}
				g.Add(makeComplexReturn(helper(gen, child)))
			}
			return gen.EndGroup(g)
		}
		return
	case "Sequence":
		if len(node.Children) == 1 {
			return helper(gen, node.Children[0])
		} else {
			g := gen.BeginGroup(true)
			for _, child := range node.Children {
				g.Add(makeComplexReturn(helper(gen, child)))
			}
			return gen.EndGroup(g)
		}
		return
	case "Prefix":
		front := node.Children[0]
		if len(node.Children) == 1 {
			return helper(gen, front)
		} else {
			exp := helper(gen, node.Children[len(node.Children)-1])
			switch front.Name {
			case "NOT":
				return gen.AssertNot(makeComplexReturn(exp))
			case "AND":
				return gen.AssertAnd(makeComplexReturn(exp))
			}
		}
		panic("Shouldn't reach this: " + front.Name)
	case "Suffix":
		if len(node.Children) == 1 {
			return helper(gen, node.Children[0])
		} else {
			back := node.Children[len(node.Children)-1]
			exp := makeComplexReturn(helper(gen, node.Children[0]))
			switch back.Name {
			case "PLUS":
				return gen.OneOrMore(exp)
			case "STAR":
				return gen.ZeroOrMore(exp)
			case "QUESTION":
				return gen.Maybe(exp)
			}
		}
		panic("Shouldn't reach this")
	case "Primary":
		front := node.Children[0]

		if front.Name == "Identifier" {
			return makeCall(gen.MakeParserCall(helper(gen, front)))
		} else if front.Name == "OPEN" {
			return helper(gen, node.Children[1])
		} else /*if front.Name == "Literal" */ {
			return helper(gen, front)
		}
	case "Spacing", "Space":
		// ignore
	default:
		return "\n\n-----------------------------------------------------\n" + node.Name + ", " + node.Data() + "-----------------------------------------------------\n"
	}
	for _, n := range node.Children {
		retstring += helper(gen, n)
	}
	return
}

func GenerateParser(rootNode *Node, gen Generator) string {
	gen.Begin()
	for _, node := range rootNode.Children {
		if node.Name == "Definition" {
			gen.MakeParserFunction(node)
		}
	}
	return gen.Finish()
}
