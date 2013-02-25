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
	"fmt"
	"strings"
)

type PyGenerator struct {
	s                     GeneratorSettings
	output                string
	CustomActions         []CustomAction
	havefunctions         bool
	currentFunctions      string
	currentFunctionsCount int
	currentName           string
	testfile              string
	debug, bench          bool
	inlineCount           int
	calledP               bool
	RootNode              *Node
}

func (g *PyGenerator) SetCustomActions(actions []CustomAction) {
	g.CustomActions = actions
}

func (g *PyGenerator) AddNode(data, defName string) string {
	ret := `accept = True
start = p.ParserData.Pos
` + g.Call(data) + `
end = p.ParserData.Pos
if accept:
	node = p.Root.Cleanup(start, end)
	node.Range.Clip(p.IgnoreRange)
	node.Name = "` + defName + `"
	p.Root.Append(node)
else:
	p.Root.Discard(start)
if p.IgnoreRange.Start >= end or p.IgnoreRange.End <= start:
	p.IgnoreRange = Range()
`
	return ret
}

func (g *PyGenerator) Ignore(data string) string {
	return `accept = True
start = p.ParserData.Pos
` + g.Call(data) + `
if accept and start != p.ParserData.Pos:
	if start < p.IgnoreRange.Start or p.IgnoreRange.Start == 0:
		p.IgnoreRange.Start = start
	p.IgnoreRange.End = p.ParserData.Pos
`
}
func (g *PyGenerator) MakeParserFunction(node *Node) error {
	g.calledP = false
	id := node.Children[0]
	exp := node.Children[len(node.Children)-1]
	defName := helper(g, id)
	g.currentName = defName
	data := helper(g, exp)

	if !g.havefunctions {
		g.havefunctions = true
		g.output += "\tdef realParse(p):\n\t\treturn p.p_" + defName + "()\n\n"
	}

	indenter := CodeFormatter{}
	indenter.Inc()
	indenter.Add("\ndef p_" + defName + "(p):\n")
	indenter.Inc()
	indenter.Add("# " + strings.Replace(strings.TrimSpace(node.Data()), "\n", "\n// ", -1) + "\n")

	defaultAction := true
	for i := range g.CustomActions {
		if defName == g.CustomActions[i].Name {
			defaultAction = false
			data = g.CustomActions[i].Action(g, data)
			break
		}
	}
	if defaultAction {
		data = g.AddNode(data, defName)
	}
	if strings.HasPrefix(data, "accept") || data[0] == '{' {
		end := "return accept\n"
		if data[len(data)-1] != '\n' {
			end = "\n" + end
		}
		indenter.Add("accept = False\n" + data + end)
	} else {
		indenter.Add("return " + data + "\n")
	}
	indenter.Dec()
	indenter.Add("\n")
	g.output += g.currentFunctions
	g.output += indenter.String()
	g.currentFunctions = ""
	g.currentFunctionsCount = 0
	return nil
}

func (g *PyGenerator) MakeParserCall(value string) string {
	g.calledP = true
	return "p.p_" + value
}

func (g *PyGenerator) CheckInRange(a, b string) string {
	return `if p.ParserData.Pos >= len(p.ParserData.Data):
	accept = False
else:
	c = p.ParserData.Data[p.ParserData.Pos]
	if c >= '` + a + `' and c <= '` + b + `':
		p.ParserData.Pos += 1
		accept = True
	else:
		accept = False
`
}

func (g *PyGenerator) CheckInSet(a string) string {
	a = strings.Replace(a, "\\[", "[", -1)
	a = strings.Replace(a, "\\]", "]", -1)
	tests := ""
	for i := 0; i < len(a); i++ {
		if len(tests) > 0 {
			tests += " or "
		}
		c2 := string(a[i])
		if c2[0] == '\\' {
			i++
			c2 += string(a[i])
		}
		if c2 == "'" {
			c2 = "\\'"
		}

		tests += "c == '" + c2 + "'"
	}
	return `
accept = False
if p.ParserData.Pos < len(p.ParserData.Data):
	c = p.ParserData.Data[p.ParserData.Pos]
	if ` + tests + `:
		p.ParserData.Pos += 1
		accept = True
`
}

func (g *PyGenerator) CheckAnyChar() string {
	return `if p.ParserData.Pos >= len(p.ParserData.Data):
	accept = False
else:
	p.ParserData.Pos += 1
	accept = True
`
}

func (g *PyGenerator) CheckNext(a string) string {
	if a[0] == '\'' {
		return `if p.ParserData.Pos >= len(p.ParserData.Data) or p.ParserData.Data[p.ParserData.Pos] != ` + a + `:
	accept = False
else:
	p.ParserData.Pos += 1
	accept = True
`
	}
	a = a[1 : len(a)-1]
	tests := ""
	pos := 0
	for i := 0; i < len(a); i, pos = i+1, pos+1 {
		if len(tests) > 0 {
			tests += " or "
		}
		c2 := string(a[i])
		if c2[0] == '\\' {
			i++
			c2 += string(a[i])
		}
		if c2 == "\\\"" {
			c2 = "\""
		}
		if c2 == "'" {
			c2 = "\\'"
		}
		tests += fmt.Sprintf("p.ParserData.Data[s+%d] != '%s'", pos, c2)
	}
	return fmt.Sprintf(`
accept = True
s = p.ParserData.Pos
e = s + %d
if e > len(p.ParserData.Data):
	accept = False
else:
	if %s:
		accept = False
if accept:
	p.ParserData.Pos += %d
`, pos, tests, pos)
}

func (g *PyGenerator) AssertNot(a string) string {
	return `s = p.ParserData.Pos
` + g.Call(a) + `
p.ParserData.Pos = s
accept = not accept`
}

func (g *PyGenerator) AssertAnd(a string) string {
	return `s = p.ParserData.Pos
` + g.Call(a) + `
p.ParserData.Pos = s`
}

func (g *PyGenerator) ZeroOrMore(a string) string {
	var cf CodeFormatter
	cf.Add("accept = True")
	cf.Add("\nwhile accept:\n")
	cf.Inc()
	cf.Add(g.Call(a))
	cf.Dec()
	cf.Add("\n")
	cf.Add("accept = True\n")
	return cf.String()
}

func (g *PyGenerator) OneOrMore(a string) string {
	var cf CodeFormatter
	cf.Add(`save = p.ParserData.Pos
` + g.Call(a) + `
if not accept:
	p.ParserData.Pos = save
else:
	while accept:
`)
	cf.Inc()
	cf.Inc()
	cf.Add(g.Call(a) + "\n")
	cf.Dec()
	cf.Add(`
accept = True
`)
	return cf.String()
}

func (g *PyGenerator) Maybe(a string) string {
	return g.Call(a) + "\naccept = True"
}

type pyNeedAllGroup struct {
	cf    CodeFormatter
	g     Generator
	stack list.List
	label string
}

func (b *pyNeedAllGroup) Add(value, name string) {
	b.cf.Add(b.g.Call(value) + `
if accept:
`)
	b.cf.Inc()
	b.stack.PushBack(name)
}

type pyNeedOneGroup struct {
	cf CodeFormatter
	g  Generator
}

func (b *pyNeedOneGroup) Add(value, name string) {
	b.cf.Add(b.g.Call(value) + "\nif not accept:\n")
	b.cf.Inc()
}

func (g *PyGenerator) BeginGroup(requireAll bool) Group {
	if requireAll {
		r := pyNeedAllGroup{g: g}
		r.cf.Add(`save = p.ParserData.Pos
`)
		return &r
	}
	r := pyNeedOneGroup{g: g}
	r.cf.Add(`save = p.ParserData.Pos
`)
	return &r
}
func (g *PyGenerator) UpdateError(msg string) string {
	return `if p.LastError < p.ParserData.Pos:
	p.LastError = p.ParserData.Pos
`
	// return "{\n\te := fmt.Sprintf(`Expected " + msg + " near %d`, p.ParserData.Pos)\n\tif len(p.LastError) != 0 {\n\t\te = e + \"\\n\" + p.LastError\n\t}\n\tp.LastError = e\n}"
}
func (g *PyGenerator) EndGroup(gr Group) string {
	switch t := gr.(type) {
	case *pyNeedAllGroup:
		t.cf.Add("pass\n")
		for n := t.stack.Back(); len(t.cf.Level()) > 1; n = n.Prev() {
			t.cf.Dec()
		}
		t.cf.Add("if not accept:\n")
		t.cf.Inc()
		t.cf.Add(g.UpdateError("TODO") + "\np.ParserData.Pos = save\n")
		t.cf.Dec()
		t.cf.Add("\n")
		t.cf.Dec()
		t.cf.Add("\n")

		return t.cf.String()
	case *pyNeedOneGroup:
		t.cf.Add("pass\n")
		for len(t.cf.Level()) > 1 {
			t.cf.Dec()
		}
		t.cf.Add("if not accept:\n\tp.ParserData.Pos = save\n\n")
		return t.cf.String()
	}
	panic(gr)
}

func (g *PyGenerator) Call(value string) string {
	pref := "accept = "
	if callre1.MatchString(value) {
		pref = ""
	}
	if strings.HasSuffix(value, "(p)") {
		return pref + value
	}
	if strings.HasPrefix(value, "p.p_") || callre2.MatchString(value) {
		return pref + value + "()"
	}
	return value
}

func (g *PyGenerator) Begin(s GeneratorSettings) error {
	g.s = s

	g.output = g.s.Header + `
class Range:
	def __init__(self, s=0, e=0):
		self.Start = s
		self.End = e

	def Clip(self, other):
		pass

class Node:
	def __init__(self, Name="", Range=Range()):
		self.Name = Name
		self.Range = Range
		self.Children = []
		self.P = None

	def format(self, indent, p):
		if len(self.Children) == 0:
			return indent + "%d-%d: \"%s\" - Data: \"%s\"\n" % (self.Range.Start, self.Range.End, self.Name, p.Data(self.Range.Start, self.Range.End))
		ret = indent + "%d-%d: \"%s\"\n" % (self.Range.Start, self.Range.End, self.Name)
		indent += "\t"
		for child in self.Children:
			ret += child.format(indent, p)
		return ret

	def __repr__(self):
		return self.format("", self.P)

	def Append(self, other):
		self.Children.append(other)

	def Cleanup(self, pos, end):
		popped = Node()
		popped.Range = Range(pos, end)
		back = len(self.Children)
		popIdx = 0
		popEnd = back
		if end == 0:
			end = -1
		if pos == 0:
			pos = -1

		i = back-1
		while i >= 0:
			node = self.Children[i]
			if node.Range.End <= pos:
				popIdx = i + 1
				break
			if node.Range.Start > end:
				popEnd = i + 1
			i -= 1

		if popEnd != 0:
			popped.Children = self.Children[popIdx:popEnd]
		if popIdx != back:
			self.Children = self.Children[:popIdx]
		return popped

	def Discard(self, pos):
		i = len(self.Children)-1
		while i >= 0:
			if self.Children[i].Range.End <= pos:
				break
			self.Children.pop()

class Pd:
	def __init__(self):
		self.Pos = 0
		self.Data = []

class ` + g.s.Name + `:
	def __init__(self):
		self.IgnoreRange = Range()
		self.LastError = 0
		self.ParserData = Pd()

	def Parse(p, data):
		p.ParserData.Data = data
		p.ParserData.Pos = 0
		p.Root = Node("` + g.s.Name + `")
		p.Root.P = p
		p.IgnoreRange = Range()
		p.LastError = 0
		ret = p.realParse()
		if len(p.Root.Children) > 0:
			p.Root.Range = Range(p.Root.Children[0].Range.Start, p.Root.Children[len(p.Root.Children)-1].Range.End)
		return ret

	def Data(self, start, end):
		l = len(p.ParserData.Data)
		if l == 0:
			return ""
		if start < 0:
			start = 0
		if end > l:
			end = l
		if start > end:
			return ""
		return p.ParserData.Data[start:end]

`
	return nil
}

func (g *PyGenerator) Finish() error {
	ret := g.output + `
import time

f = open("` + g.s.Testname + `")
data = f.read()
f.close()
p = ` + g.s.Name + `()
assert p.Parse(data)
`
	if g.s.Debug {
		ret += "print p.Root\n"
	}
	if g.s.Bench {
		ret += `
t = time.time()
N = 1000
for i in range(1000):
	p.Parse(data)
t2 = time.time()
print "%f ns/op" % ((t2-t)*10e8/N)
`
	}
	if ret[len(ret)-2:] == "\n\n" {
		ret = ret[:len(ret)-1]
	}
	g.output = ""
	ln := strings.ToLower(g.s.Name)
	if err := g.s.WriteFile(ln+".py", ret); err != nil {
		return err
	}

	return nil
}

func (g *PyGenerator) TestCommand() []string {
	return []string{"python", strings.ToLower(g.s.Name) + ".py"}
}
