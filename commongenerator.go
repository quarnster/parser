// !darwin
// /*
// Copyright (c) 2012-2013 Fredrik Ehnbom
// All rights reserved.

// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:

// 1. Redistributions of source code must retain the above copyright notice, this
// 		list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright notice,
// 		this list of conditions and the following disclaimer in the documentation
// 		and/or other materials provided with the distribution.

// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
// */

package parser

// import (
// 	"container/list"
// 	"fmt"
// 	"regexp"
// 	"strings"
// )

// type CommonGenerator struct {
// 	CustomActions  []CustomAction
// 	accept         Statement
// 	parserInstance Statement
// }

// func (g *CommonGenerator) SetCustomActions(actions []CustomAction) {
// 	g.CustomActions = actions
// }

// type Statement interface {
// 	Not(Statement) Statement
// 	Set(Statement) Statement
// 	Member(name string) Statement
// }

// type Language interface {
// 	DeclareVar(name string, value Statement) Statement
// 	If(Statement) Statement
// 	Else(Statement) Statement
// 	EndIf(Statement) Statement
// 	While(Statement) Statement
// 	True() Statement
// 	False() Statement
// 	ParserPos() Statement
// 	Or(Statement, Statement) Statement
// 	And(Statement, Statement) Statement
// 	Call(Statement, ...Statement) Statement
// 	String(str string) Statement
// }

// type Compiler interface {
// 	Add(Statement)
// 	Compile() string
// }

// func (g *CommonGenerator) AddNode(data, defName string) string {
// 	var l Language
// 	var c Compiler
// 	var Root Statement
// 	c.Add(g.accept.Set(l.True()))
// 	start := l.DeclareVar("start", l.ParserPos())
// 	c.Add(start)
// 	c.Add(g.Call(data))
// 	end := l.DeclareVar("end", l.ParserPos())
// 	c.Add(end)
// 	c.Add(l.If(g.accept))

// 	node := l.DeclareVar("node", l.Call(Root.Member("Cleanup"), start, end))
// 	c.Add(node)
// 	c.Add(node.Member("Name").Set(l.String(defname)))
// 	c.Add(node.Member("P").Set(g.parserInstance))
// 	c.Add(l.Call(node.Member("Range").Member("Clip"), g.parserInstance.Member("IgnoreRange")))
// 	c.Add(l.Call(Root.Member("Append"), node))
// 	c.Add(l.Else())
// 	c.Add(l.Call(Root.Member("Discard"), start))
// 	ret += `
// }
// if p.IgnoreRange.Start >= end || p.IgnoreRange.End <= start {
// 	p.IgnoreRange = Range{}
// }
// `
// 	return ret
// }

// func (g *CommonGenerator) Ignore(data string) string {
// 	return `accept = true
// start := p.ParserData.Pos
// ` + g.Call(data) + `
// if accept && start != p.ParserData.Pos {
// 	if start < p.IgnoreRange.Start || p.IgnoreRange.Start == 0 {
// 		p.IgnoreRange.Start = start
// 	}
// 	p.IgnoreRange.End = p.ParserData.Pos
// }
// `
// }
// func (g *CommonGenerator) MakeParserFunction(node *Node) error {
// 	g.calledP = false
// 	id := node.Children[0]
// 	exp := node.Children[len(node.Children)-1]
// 	defName := helper(g, id)
// 	g.currentName = defName
// 	data := helper(g, exp)

// 	if !g.havefunctions {
// 		g.havefunctions = true
// 		g.output += "func (p *" + g.s.Name + ") realParse() bool {\n\treturn p.p_" + defName + "()\n}\n"
// 	}

// 	indenter := CodeFormatter{}
// 	indenter.Add("func (p *" + g.s.Name + ") p_" + defName + "() bool {\n")
// 	indenter.Inc()
// 	indenter.Add("// " + strings.Replace(strings.TrimSpace(node.Data()), "\n", "\n// ", -1) + "\n")

// 	defaultAction := true
// 	for i := range g.CustomActions {
// 		if defName == g.CustomActions[i].Name {
// 			defaultAction = false
// 			data = g.CustomActions[i].Action(g, data)
// 			break
// 		}
// 	}
// 	if defaultAction {
// 		data = g.AddNode(data, defName)
// 	}
// 	if g.AddDebugLogging {
// 		if strings.HasPrefix(data, "accept") || data[0] == '{' {
// 			data = "accept := false\n" + data
// 		} else {
// 			data = "accept := " + data
// 		}
// 		indenter.Add(`var (
// 	pos = p.ParserData.Pos
// 	l   = len(p.ParserData.Data)
// )

// log.Println(fm.Level() + "` + defName + ` entered")
// fm.Inc()
// ` + data + `
// fm.Dec()
// if !accept && p.ParserData.Pos != pos {
// 	log.Fatalln("` + defName + `", accept, ", ", pos, ", ", p.ParserData.Pos)
// }
// p2 := p.ParserData.Pos
// data := ""
// if p2 < len(p.ParserData.Data) {
// 	data = string(p.ParserData.Data[pos:p2])
// }
// log.Println(fm.Level()+"` + defName + ` returned: ", accept, ", ", pos, ", ", p.ParserData.Pos, ", ", l, string(data))
// return accept
// `)
// 	} else {
// 		if strings.HasPrefix(data, "accept") || data[0] == '{' {
// 			end := "return accept\n"
// 			if data[len(data)-1] != '\n' {
// 				end = "\n" + end
// 			}
// 			indenter.Add("accept := false\n" + data + end)
// 		} else {
// 			indenter.Add("return " + data + "\n")
// 		}
// 	}
// 	indenter.Dec()
// 	indenter.Add("}\n\n")
// 	g.output += g.currentFunctions
// 	g.output += indenter.String()
// 	g.currentFunctions = ""
// 	g.currentFunctionsCount = 0
// 	return nil
// }

// func (g *CommonGenerator) MakeParserCall(value string) string {
// 	g.calledP = true
// 	// if g.inlineCount < 0 && g.RootNode != nil {
// 	// 	g.inlineCount++
// 	// 	for _, child := range g.RootNode.Children {
// 	// 		if child.Name == "Definition" && child.Children[0].Data() == value {
// 	// 			data := helper(g, child.Children[len(child.Children)-1])
// 	// 			defaultAction := true
// 	// 			for i := range g.CustomActions {
// 	// 				if value == g.CustomActions[i].Name {
// 	// 					defaultAction = false
// 	// 					data = g.CustomActions[i].Action(g, data)
// 	// 					break
// 	// 				}
// 	// 			}
// 	// 			if defaultAction {
// 	// 				data = g.AddNode(data, value)
// 	// 			}

// 	// 			ret := "/* inlined " + value + "*/\n" + data
// 	// 			g.inlineCount--
// 	// 			return ret
// 	// 		}
// 	// 	}
// 	// 	g.inlineCount--
// 	// }

// 	return "p.p_" + value
// }

// func (g *CommonGenerator) CheckInRange(a, b string) string {
// 	return `if p.ParserData.Pos >= len(p.ParserData.Data) {
// 	accept = false
// } else {
// 	c := p.ParserData.Data[p.ParserData.Pos]
// 	if c >= '` + a + `' && c <= '` + b + `' {
// 		p.ParserData.Pos++
// 		accept = true
// 	} else {
// 		accept = false
// 	}
// }`
// }

// func (g *CommonGenerator) CheckInSet(a string) string {
// 	a = strings.Replace(a, "\\[", "[", -1)
// 	a = strings.Replace(a, "\\]", "]", -1)
// 	tests := ""
// 	for i := 0; i < len(a); i++ {
// 		if len(tests) > 0 {
// 			tests += " || "
// 		}
// 		c2 := string(a[i])
// 		if c2[0] == '\\' {
// 			i++
// 			c2 += string(a[i])
// 		}
// 		if c2 == "'" {
// 			c2 = "\\'"
// 		}

// 		tests += "c == '" + c2 + "'"
// 	}
// 	return `{
// 	accept = false
// 	if p.ParserData.Pos < len(p.ParserData.Data) {
// 		c := p.ParserData.Data[p.ParserData.Pos]
// 		if ` + tests + ` {
// 			p.ParserData.Pos++
// 			accept = true
// 		}
// 	}
// }`
// }

// func (g *CommonGenerator) CheckAnyChar() string {
// 	return `if p.ParserData.Pos >= len(p.ParserData.Data) {
// 	accept = false
// } else {
// 	p.ParserData.Pos++
// 	accept = true
// }`
// }

// func (g *CommonGenerator) CheckNext(a string) string {
// 	/*
// 	 */

// 	if a[0] == '\'' {
// 		return `if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != ` + a + ` {
// 	accept = false
// } else {
// 	p.ParserData.Pos++
// 	accept = true
// }`
// 	}
// 	a = a[1 : len(a)-1]
// 	tests := ""
// 	pos := 0
// 	for i := 0; i < len(a); i, pos = i+1, pos+1 {
// 		if len(tests) > 0 {
// 			tests += " || "
// 		}
// 		c2 := string(a[i])
// 		if c2[0] == '\\' {
// 			i++
// 			c2 += string(a[i])
// 		}
// 		if c2 == "\\\"" {
// 			c2 = "\""
// 		}
// 		if c2 == "'" {
// 			c2 = "\\'"
// 		}
// 		tests += fmt.Sprintf("p.ParserData.Data[s+%d] != '%s'", pos, c2)
// 	}
// 	return fmt.Sprintf(`{
// 	accept = true
// 	s := p.ParserData.Pos
// 	e := s + %d
// 	if e > len(p.ParserData.Data) {
// 		accept = false
// 	} else {
// 		if %s {
// 			accept = false
// 		}
// 	}
// 	if accept {
// 		p.ParserData.Pos += %d
// 	}
// }`, pos, tests, pos)
// }

// func (g *CommonGenerator) AssertNot(a string) string {
// 	return `s := p.ParserData.Pos
// ` + g.Call(a) + `
// p.ParserData.Pos = s
// accept = !accept`
// }

// func (g *CommonGenerator) AssertAnd(a string) string {
// 	return `s := p.ParserData.Pos
// ` + g.Call(a) + `
// p.ParserData.Pos = s`
// }

// func (g *CommonGenerator) ZeroOrMore(a string) string {
// 	var cf CodeFormatter
// 	cf.Add("{\n")
// 	cf.Inc()
// 	cf.Add("accept = true")
// 	cf.Add("\nfor accept {\n")
// 	cf.Inc()
// 	cf.Add(g.Call(a))
// 	cf.Dec()
// 	cf.Add("\n}\n")
// 	cf.Add("accept = true\n")
// 	cf.Dec()
// 	cf.Add("}")
// 	return cf.String()
// }

// func (g *CommonGenerator) OneOrMore(a string) string {
// 	var cf CodeFormatter
// 	cf.Add("{\n")
// 	cf.Inc()
// 	cf.Add(`save := p.ParserData.Pos
// ` + g.Call(a) + `
// if !accept {
// 	p.ParserData.Pos = save
// } else {
// 	for accept {
// `)
// 	cf.Inc()
// 	cf.Inc()
// 	cf.Add(g.Call(a) + "\n")
// 	cf.Dec()
// 	cf.Add(`}
// accept = true
// `)
// 	cf.Dec()
// 	cf.Add("}\n")
// 	cf.Dec()
// 	cf.Add("}")
// 	return cf.String()
// }

// func (g *CommonGenerator) Maybe(a string) string {
// 	return g.Call(a) + "\naccept = true"
// }

// type needAllGroup struct {
// 	cf    CodeFormatter
// 	g     Generator
// 	stack list.List
// 	label string
// }

// func (b *needAllGroup) Add(value, name string) {
// 	b.cf.Add(b.g.Call(value) + `
// if accept {
// `)
// 	b.cf.Inc()
// 	b.stack.PushBack(name)
// }

// type needOneGroup struct {
// 	cf CodeFormatter
// 	g  Generator
// }

// func (b *needOneGroup) Add(value, name string) {
// 	b.cf.Add(b.g.Call(value) + "\nif !accept {\n")
// 	b.cf.Inc()
// }

// func (g *CommonGenerator) BeginGroup(requireAll bool) Group {
// 	if requireAll {
// 		r := needAllGroup{g: g}
// 		r.cf.Add(`{
// 	save := p.ParserData.Pos
// `)
// 		r.cf.Inc()
// 		return &r
// 	}
// 	r := needOneGroup{g: g}
// 	r.cf.Add(`{
// 	save := p.ParserData.Pos
// `)
// 	r.cf.Inc()
// 	return &r
// }
// func (g *CommonGenerator) UpdateError(msg string) string {
// 	return `if p.LastError < p.ParserData.Pos {
// 	p.LastError = p.ParserData.Pos
// }`
// 	// return "{\n\te := fmt.Sprintf(`Expected " + msg + " near %d`, p.ParserData.Pos)\n\tif len(p.LastError) != 0 {\n\t\te = e + \"\\n\" + p.LastError\n\t}\n\tp.LastError = e\n}"
// }
// func (g *CommonGenerator) EndGroup(gr Group) string {
// 	switch t := gr.(type) {
// 	case *needAllGroup:

// 		for n := t.stack.Back(); len(t.cf.Level()) > 1; n = n.Prev() {
// 			t.cf.Dec()
// 			t.cf.Add("}\n")
// 		}
// 		t.cf.Add("if !accept {\n")
// 		t.cf.Inc()
// 		t.cf.Add(g.UpdateError("TODO") + "\np.ParserData.Pos = save\n")
// 		t.cf.Dec()
// 		t.cf.Add("}\n")
// 		t.cf.Dec()
// 		t.cf.Add("}")
// 		return t.cf.String()
// 	case *needOneGroup:
// 		for len(t.cf.Level()) > 1 {
// 			t.cf.Dec()
// 			t.cf.Add("}\n")
// 		}
// 		t.cf.Add("if !accept {\n\tp.ParserData.Pos = save\n}\n")
// 		t.cf.Dec()
// 		t.cf.Add("}")
// 		return t.cf.String()
// 	}
// 	panic(gr)
// }

// var (
// 	callre1 = regexp.MustCompile(`^\s*accept\s`)
// 	callre2 = regexp.MustCompile(`^\s*func\(`)
// )

// func (g *CommonGenerator) Call(value string) string {
// 	pref := "accept = "
// 	if callre1.MatchString(value) {
// 		pref = ""
// 	}
// 	if strings.HasSuffix(value, "(p)") {
// 		return pref + value
// 	}
// 	if strings.HasPrefix(value, "p.p_") || callre2.MatchString(value) {
// 		return pref + value + "()"
// 	}
// 	return value
// }

// func (g *CommonGenerator) Begin(s GeneratorSettings) error {
// 	g.s = s
// 	imports := "\n\nimport (\n\t\"bytes\"\n\t. \"parser\"\n"
// 	impList := g.Imports
// 	if g.AddDebugLogging {
// 		impList = append(impList, "log")
// 	}
// 	if len(impList) > 0 {
// 		imports += "\t\"" + strings.Join(impList, "\"\n\t\"") + "\"\n"
// 	}
// 	imports += ")\n"

// 	g.output = g.s.Header + "\n"
// 	members := g.ParserVariables
// 	members = append(members, `ParserData struct {
// 		Pos  int
// 		Data []byte
// 	}
// `, "IgnoreRange Range",
// 		"Root        Node",
// 		"LastError   int")
// 	g.output += fmt.Sprintln("package " + strings.ToLower(g.s.Name) + imports + "\ntype " + g.s.Name + " struct {\n\t" + strings.Join(members, "\n\t") + "\n}\n")

// 	if g.AddDebugLogging {
// 		g.output += "var fm CodeFormatter\n\n"
// 	}
// 	g.output += `func (p *` + g.s.Name + `) RootNode() *Node {
// 	return &p.Root
// }

// func (p *` + g.s.Name + `) Parse(data string) bool {
// 	p.ParserData.Data = ([]byte)(data)

// 	p.ParserData.Pos = 0
// 	p.Root = Node{Name: "` + g.s.Name + `", P:p}
// 	p.IgnoreRange = Range{}
// 	p.LastError = 0
// 	ret := p.realParse()
// 	if len(p.Root.Children) > 0 {
// 		p.Root.Range = Range{p.Root.Children[0].Range.Start, p.Root.Children[len(p.Root.Children)-1].Range.End}
// 	}
// 	return ret
// }

// func (p *` + g.s.Name + `) Data(start, end int) string {
// 	l := len(p.ParserData.Data)
// 	if l == 0 {
// 		return ""
// 	}
// 	if start < 0 {
// 		start = 0
// 	}
// 	if end > l {
// 		end = l
// 	}
// 	if start > end {
// 		return ""
// 	}
// 	return string(p.ParserData.Data[start:end])
// }

// func (p *` + g.s.Name + `) Error() Error {
// 	errstr := ""

// 	line := 1
// 	column := 1
// 	for _, r := range p.ParserData.Data[:p.LastError] {
// 		column++
// 		if r == '\n' {
// 			line++
// 			column = 1
// 		}
// 	}

// 	if p.LastError == len(p.ParserData.Data) {
// 		errstr = "Unexpected EOF"
// 	} else {
// 		e := p.LastError + 4
// 		if e > len(p.ParserData.Data) {
// 			e = len(p.ParserData.Data)
// 		}

// 		reader := bytes.NewReader(p.ParserData.Data[p.LastError:e])
// 		r, _, _ := reader.ReadRune()
// 		if r == '\r' || r == '\n' {
// 			errstr = "Unexpected new line"
// 		} else {
// 			errstr = "Unexpected " + string(r)
// 		}
// 	}
// 	return NewError(line, column, errstr)
// }

// `
// 	return nil
// }

// func (g *CommonGenerator) Finish() error {
// 	ret := g.output
// 	if ret[len(ret)-2:] == "\n\n" {
// 		ret = ret[:len(ret)-1]
// 	}
// 	g.output = ""
// 	ln := strings.ToLower(g.s.Name)
// 	if err := g.s.WriteFile(ln+".go", ret); err != nil {
// 		return err
// 	}

// 	dumptree_s := ""

// 	if g.s.Debug {
// 		dumptree_s = "t.Log(\"\\n\"+root.String())"
// 	}
// 	test := `package ` + strings.ToLower(g.s.Name) + `
// 		import (
// 			"io/ioutil"
// 			"log"
// 			"testing"
// 		)
// const testname = "` + g.s.Testname + `"
// func TestParser(t *testing.T) {
// 	var p ` + g.s.Name + `
// 	if data, err := ioutil.ReadFile(testname); err != nil {
// 		log.Fatalf("%s", err)
// 	} else {
// 		if !p.Parse(string(data)) {
// 			t.Fatalf("Didn't parse correctly: %s\n", p.Error())
// 		} else {
// 			root := p.RootNode()
// 			` + dumptree_s + `
// 			if root.Range.End != len(p.ParserData.Data) {
// 				t.Fatalf("Parsing didn't finish: %v\n%s", root, p.Error())
// 			}
// 		}
// 	}
// }

// func BenchmarkParser(b *testing.B) {
// 	var p ` + g.s.Name + `
// 	if data, err := ioutil.ReadFile(testname); err != nil {
// 		b.Fatalf("%s", err)
// 	} else {
// 		d2 := string(data)
// 		for i := 0; i < b.N; i++ {
// 			p.Parse(d2)
// 		}
// 	}
// }

// 		`
// 	if err := g.s.WriteFile(ln+"_test.go", test); err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (g *CommonGenerator) TestCommand() []string {
// 	cmd := []string{"go", "test", "-v", "-gcflags", "-B"}
// 	if g.s.Bench {
// 		cmd = append(cmd, "-bench", ".")
// 	}
// 	return cmd
// }
