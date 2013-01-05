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

type GeneratorSettings struct {
	Debug     bool
	Bench     bool
	Testname  string
	WriteFile func(name, data string) error
}

type Group interface {
	Add(value, name string)
}

type Generator interface {
	TestCommand() []string
	SetName(name string)
	Name() string
	SetCustomActions([]CustomAction)
	AddNode(data, defName string) string
	Ignore(value string) string
	Return(value string) string
	Call(value string) string
	MakeFunction(value string) string
	MakeParserFunction(definitionNode *Node) error
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
	Begin(GeneratorSettings) error
	Finish() error
}
type CustomAction struct {
	Name   string
	Action func(Generator, string) string
}

func helper(gen Generator, node *Node) (retstring string) {
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
				g.Add(e, "")
			}
			return gen.EndGroup(g)
		} else {
			return exps[0]
		}
	case "DOT":
		return gen.CheckAnyChar()
	case "Identifier":
		return node.Data()
	case "Literal":
		return gen.CheckNext(node.Data())
	case "Expression":
		if len(node.Children) == 1 {
			return helper(gen, node.Children[0])
		} else {
			g := gen.BeginGroup(false)
			for _, child := range node.Children {
				g.Add(helper(gen, child), child.Name)
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
				g.Add(helper(gen, child), child.Name)
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
				return gen.AssertNot(exp)
			case "AND":
				return gen.AssertAnd(exp)
			}
		}
		panic("Shouldn't reach this: " + front.Name)
	case "Suffix":
		if len(node.Children) == 1 {
			return helper(gen, node.Children[0])
		} else {
			back := node.Children[len(node.Children)-1]
			exp := helper(gen, node.Children[0])
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
			return gen.Call(gen.MakeParserCall(helper(gen, front)))
		} else {
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

func GenerateParser(rootNode *Node, gen Generator, s GeneratorSettings) error {
	if err := gen.Begin(s); err != nil {
		return err
	}
	for _, node := range rootNode.Children {
		if node.Name == "Definition" {
			if err := gen.MakeParserFunction(node); err != nil {
				return err
			}
		}
	}
	return gen.Finish()
}
