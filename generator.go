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
	"fmt"
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

type GoGenerator struct {
	output          string
	addDebugLogging bool
	Name            string
}

func (g *GoGenerator) Begin() {
	imports := ""
	if g.addDebugLogging {
		imports += "\nimports (\n\t\"log\"\n)\n"
	}
	g.output = `/*
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
`
	g.output += fmt.Sprintln("package parser\n" + imports + "\ntype " + g.Name + " struct {\n\tParser\n}\n")
	if g.addDebugLogging {
		g.output += "var fm CodeFormatter\n\n"
	}
}

func (g *GoGenerator) Finish() string {
	ret := g.output
	if ret[len(ret)-2:] == "\n\n" {
		ret = ret[:len(ret)-1]
	}
	g.output = ""
	return ret
}

func (g *GoGenerator) MakeParserFunction(node *Node) {
	id := node.Children.Front().Value.(*Node)
	exp := node.Children.Back().Value.(*Node)
	data := helper(g, exp)
	defName := helper(g, id)

	indenter := CodeFormatter{}
	indenter.Add("func (p *" + g.Name + ") " + defName + "() bool {\n")
	indenter.Inc()
	comment := "/* " + strings.Replace(strings.TrimSpace(node.Data), "\n", "\n * ", -1)
	indenter.Add(comment)
	if strings.ContainsRune(comment, '\n') {
		indenter.Add("\n")
	}
	indenter.Add(" */\n")

	if g.addDebugLogging {
		indenter.Add(`var (
    pos = p.Pos()
    l   = p.data.Len()
)
`)
		indenter.Add(`log.Println(fm.level + "` + defName + " entered\")\n")
		indenter.Add("fm.Inc()\n")
	}
	data = "p.addNode(" + g.MakeFunction("return "+data) + ", \"" + defName + "\")"
	if g.addDebugLogging {
		indenter.Add("res := " + data)
		indenter.Add("\nfm.Dec()\n")
		indenter.Add(`if !res && p.Pos() != pos {
    log.Fatalln("` + defName + `", res, ", ", pos, ", ", p.Pos())
}
p2 := p.Pos()
data := make([]byte, p2-pos)
p.data.Seek(int64(pos), 0)
p.data.Read(data)
p.data.Seek(int64(p2), 0)
`)
		indenter.Add("log.Println(fm.level+\"" + defName + ` returned: ", res, ", ", pos, ", ", p.Pos(), ", ", l, string(data))` + "\n")
		indenter.Add("return res\n")
	} else {
		indenter.Add("return " + data + "\n")
	}
	indenter.Dec()
	indenter.Add("}\n\n")
	g.output += indenter.String()
}

func (g *GoGenerator) MakeParserCall(value string) string {
	return "p." + value
}

func (g *GoGenerator) Return(value string) string {
	return "return " + value
}

func (g *GoGenerator) Call(value string) string {
	return value + "()"
}

func (g *GoGenerator) MakeFunction(value string) string {
	f := CodeFormatter{}
	f.Add("func() bool {\n")
	f.Inc()
	f.Add(value)
	f.Dec()
	f.Add("\n}")
	return f.String()
}

func (g *GoGenerator) CheckInRange(a, b string) string {
	return "p.InRange('" + a + "', '" + b + "')"
}

func (g *GoGenerator) CheckInSet(a string) string {
	return "p.InSet(`" + a + "`)"
}

func (g *GoGenerator) CheckAnyChar() string {
	return "p.AnyChar()"
}

func (g *GoGenerator) CheckNext(a string) string {
	if a[0] == '\'' && a[len(a)-1] == '\'' {
		a = "\"" + a[1:len(a)-1] + "\""
	}
	return "p.Next(" + a + ")"
}

func (g *GoGenerator) AssertNot(a string) string {
	return "p.Not(" + a + ")"
}

func (g *GoGenerator) AssertAnd(a string) string {
	return "p.And(" + a + ")"
}

func (g *GoGenerator) ZeroOrMore(a string) string {
	return "p.ZeroOrMore(" + a + ")"
}

func (g *GoGenerator) OneOrMore(a string) string {
	return "p.OneOrMore(" + a + ")"
}

func (g *GoGenerator) Maybe(a string) string {
	return "p.Maybe(" + a + ")"
}

func (g *GoGenerator) BeginGroup(requireAll bool) Group {
	r := baseGroup{}
	if requireAll {
		r.cf.Add("p.NeedAll([]func() bool{\n")
	} else {
		r.cf.Add("p.NeedOne([]func() bool{\n")
	}
	r.cf.Inc()
	return &r
}

func (g *GoGenerator) EndGroup(gr Group) string {
	bg := gr.(*baseGroup)
	bg.cf.Dec()
	bg.cf.Add("})")
	return bg.cf.String()
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
		g := gen.BeginGroup(false)
		others := ""
		for n := node.Children.Front(); n != nil; n = n.Next() {
			child := n.Value.(*Node)
			if child.Name == "Range" {
				if child.Children.Len() == 2 {
					g.Add(makeComplexReturn(gen.CheckInRange(child.Children.Front().Value.(*Node).Data, child.Children.Back().Value.(*Node).Data)))
				} else {
					others += child.Data
				}
			}
		}
		if others != "" {
			g.Add(makeComplexReturn(gen.CheckInSet(others)))
		}
		return gen.EndGroup(g)
	case "DOT":
		return gen.CheckAnyChar()
	case "Identifier":
		back := node.Children.Back().Value.(*Node)
		data := node.Data
		if back.Name == "Spacing" {
			data = strings.Replace(data, back.Data, "", -1)
		}
		return data
	case "Literal":
		data := strings.TrimSpace(node.Data)
		return gen.CheckNext(data)
	case "Expression":
		if node.Children.Len() == 1 {
			return helper(gen, node.Children.Front().Value.(*Node))
		} else {
			g := gen.BeginGroup(false)
			i := 0
			for n := node.Children.Front(); n != nil; n = n.Next() {
				i++
				child := n.Value.(*Node)
				if i&1 == 0 {
					continue
				}
				g.Add(makeComplexReturn(helper(gen, child)))
			}
			return gen.EndGroup(g)
		}
		return
	case "Sequence":
		if node.Children.Len() == 1 {
			return helper(gen, node.Children.Front().Value.(*Node))
		} else {
			g := gen.BeginGroup(true)
			for n := node.Children.Front(); n != nil; n = n.Next() {
				g.Add(makeComplexReturn(helper(gen, n.Value.(*Node))))
			}
			return gen.EndGroup(g)
		}
		return
	case "Prefix":
		front := node.Children.Front().Value.(*Node)
		if node.Children.Len() == 1 {
			return helper(gen, front)
		} else {
			exp := helper(gen, node.Children.Back().Value.(*Node))
			switch front.Name {
			case "NOT":
				return gen.AssertNot(makeComplexReturn(exp))
			case "AND":
				return gen.AssertAnd(makeComplexReturn(exp))
			}
		}
		panic("Shouldn't reach this: " + front.Name)
	case "Suffix":
		if node.Children.Len() == 1 {
			return helper(gen, node.Children.Front().Value.(*Node))
		} else {
			back := node.Children.Back().Value.(*Node)
			exp := makeComplexReturn(helper(gen, node.Children.Front().Value.(*Node)))
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
		front := node.Children.Front().Value.(*Node)

		if front.Name == "Identifier" {
			return makeCall(gen.MakeParserCall(helper(gen, front)))
		} else if front.Name == "OPEN" {
			return helper(gen, node.Children.Front().Next().Value.(*Node))
		} else if front.Name == "Literal" {
			return helper(gen, front)
		}
	case "Spacing", "Space":
		// ignore
	default:
		return "\n\n-----------------------------------------------------\n" + node.Name + ", " + node.Data + "-----------------------------------------------------\n"
	}
	for n := node.Children.Front(); n != nil; n = n.Next() {
		retstring += helper(gen, n.Value.(*Node))
	}
	return
}

func GenerateParser(rootNode *Node, gen Generator) string {
	gen.Begin()
	for n := rootNode.Children.Front(); n != nil; n = n.Next() {
		node := n.Value.(*Node)
		if node.Name == "Definition" {
			gen.MakeParserFunction(node)
		}
	}
	return gen.Finish()
}
