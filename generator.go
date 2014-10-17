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

const (
	DebugLevelNone DebugLevel = iota
	DebugLevelNodeCreation
	DebugLevelEnterExit
	DebugLevelAccept
)

type (
	CodeFormatter struct {
		level string
		data  string
	}

	DebugLevel int

	GeneratorSettings struct {
		// Whether to generate a parser which outputs debug information
		DebugLevel DebugLevel
		Header     string
		Debug      bool
		Bench      bool
		Testname   string
		Name       string
		FileName   string
		WriteFile  func(name, data string) error
		Heatmap    bool
	}

	Group interface {
		Add(value, name string)
	}

	Value string

	Context interface {
	}

	Generator interface {
		TestCommand() []string
		SetCustomActions([]CustomAction)
		AddNode(data, defName string) string
		Ignore(value string) string

		// Make a call to the given function
		Call(value string) string

		MakeParserFunction(definitionNode *Node) error

		// Get the name of a parser function for the given
		// definition
		MakeParserCall(value string) string

		// Accept and consume input if a character between a and b follows.
		// Backtracks if it doesn't.
		CheckInRange(a, b string) string

		// Accept and consume input if any character in the set follows.
		// Backtracks if it doesn't.
		CheckInSet(a string) string

		// Accept and consume input if any character follows.
		// Backtracks if it doesn't.
		CheckAnyChar() string

		// Accept and consume input if "a" follows.
		// Backtracks if it doesn't.
		CheckNext(a string) string

		// Make sure that "a" does not follow, without consuming input
		AssertNot(a string) string

		// Make sure that "a" follows, without consuming input
		AssertAnd(a string) string

		// Zero or More occurances of "a" follows
		ZeroOrMore(a string) string

		// At least one of "a" follows, but more than
		// one is ok too.
		OneOrMore(a string) string

		// Might or might not follow once, but pass either way
		Maybe(a string) string

		// Begin a grouping either requiring all values
		// to be true, or just one of the values to be true.
		BeginGroup(requireAll bool) Group

		// End the previously started group
		EndGroup(g Group) string

		// Called when generation starts
		Begin(GeneratorSettings) error
		// Called when generation is done
		Finish() error
	}
	CustomAction struct {
		Name   string
		Action func(Generator, string) string
	}
)

func (i *CodeFormatter) Level() string {
	return i.level
}
func (i *CodeFormatter) Inc(data ...string) {
	i.level += "\t"
	l := len(i.data)
	if l > 0 && (i.data[l-1] == '\t' || i.data[l-1] == '\n') {
		i.data += "\t"
	}
	for idx := range data {
		i.Add(data[idx])
	}
}
func (i *CodeFormatter) Dec(data ...string) {
	i.level = i.level[:len(i.level)-1]
	l := len(i.data)
	if l > 0 && i.data[l-1] == '\t' {
		i.data = i.data[:l-1]
	}
	for idx := range data {
		i.Add(data[idx])
	}
}
func (i *CodeFormatter) Add(add string) {
	i.data += strings.Replace(add, "\n", "\n"+i.level, -1)
}

func (i *CodeFormatter) String() string {
	return i.data
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
