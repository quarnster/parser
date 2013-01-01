package parser

import (
	"fmt"
	"regexp"
	"strings"
)

type GoGenerator2 struct {
	output                string
	AddDebugLogging       bool
	Name                  string
	CustomActions         []CustomAction
	ParserVariables       []string
	Imports               []string
	havefunctions         bool
	currentFunctions      string
	currentFunctionsCount int
	currentName           string
}

func (g *GoGenerator2) MakeParserFunction(node *Node) {
	id := node.Children[0]
	exp := node.Children[len(node.Children)-1]
	defName := helper(g, id)
	g.currentName = defName
	data := helper(g, exp)

	if !g.havefunctions {
		g.havefunctions = true
		g.output += "func (p *" + g.Name + ") Parse() bool {\n\treturn p_" + defName + "(p)\n}\n"
	}

	indenter := CodeFormatter{}
	indenter.Add("func p_" + defName + "(p *" + g.Name + ") bool {\n")
	indenter.Inc()
	indenter.Add("// " + strings.Replace(strings.TrimSpace(node.Data()), "\n", "\n// ", -1) + "\n")

	defaultAction := true
	for i := range g.CustomActions {
		if defName == g.CustomActions[i].Name {
			defaultAction = false
			data = g.CustomActions[i].Action(g, data)
			break
		}
	}
	if defaultAction {
		data = g.MakeFunction(data)
		data = "p_addNode(p, " + data + ", \"" + defName + "\")"
	}
	if g.AddDebugLogging {
		indenter.Add(`var (
	pos = p.ParserData.Pos
	l   = len(p.ParserData.Data)
)

log.Println(fm.Level() + "` + defName + ` entered")
fm.Inc()
res := ` + data + `
fm.Dec()
if !res && p.ParserData.Pos != pos {
	log.Fatalln("` + defName + `", res, ", ", pos, ", ", p.ParserData.Pos)
}
p2 := p.ParserData.Pos
data := string(p.ParserData.Data[pos:p2])
log.Println(fm.Level()+"` + defName + ` returned: ", res, ", ", pos, ", ", p.ParserData.Pos, ", ", l, string(data))
return res
`)
	} else {
		if strings.HasPrefix(data, "accept") {
			indenter.Add(`accept := false
` + data + `
return accept
`)
		} else {
			indenter.Add(g.Return(data) + "\n")
		}
	}
	indenter.Dec()
	indenter.Add("}\n\n")
	g.output += g.currentFunctions
	g.output += indenter.String()
	g.currentFunctions = ""
	g.currentFunctionsCount = 0
}

func (g *GoGenerator2) accept() string {
	return g.Return("true")
	//	return "accept = true"
}
func (g *GoGenerator2) reject() string {
	return g.Return("false")
	//	return "accept = false"
}

func (g *GoGenerator2) MakeParserCall(value string) string {
	return "p_" + value
}

func (g *GoGenerator2) CheckInRange(a, b string) string {
	return `if p.ParserData.Pos >= len(p.ParserData.Data) {
	accept = false
} else {
	c := p.ParserData.Data[p.ParserData.Pos]
	if c >= '` + a + `' && c <= '` + b + `' {
		p.ParserData.Pos++
		accept = true
	} else {
		accept = false
	}
}
`
}

func (g *GoGenerator2) CheckInSet(a string) string {
	a = strings.Replace(a, "\\[", "[", -1)
	a = strings.Replace(a, "\\]", "]", -1)
	a = strings.Replace(a, "\n", "\\n", -1)
	a = strings.Replace(a, "\r", "\\r", -1)
	a = strings.Replace(a, "\"", "\\\"", -1)
	return `func(p *` + g.Name + `) bool {
	dataset := []rune("` + a + `")
	if p.ParserData.Pos >= len(p.ParserData.Data) {
		` + g.Return("false") + `
	}
	c := p.ParserData.Data[p.ParserData.Pos]
	for _, r := range dataset {
		if r == c {
			p.ParserData.Pos++
			` + g.Return("true") + `
		}
	}
	` + g.Return("false") + `
}`
}

func (g *GoGenerator2) CheckAnyChar() string {
	return `func(p *` + g.Name + `) bool {
	if p.ParserData.Pos >= len(p.ParserData.Data) {
		` + g.Return("false") + `
	}
	p.ParserData.Pos++
	` + g.Return("true") + `
}`
}

func (g *GoGenerator2) CheckNext(a string) string {
	/*
	 */

	if a[0] == '\'' {
		return `if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != ` + a + ` {
	accept = false
} else {
	p.ParserData.Pos++
	accept = true
}
`
	}
	a = a[1 : len(a)-1]
	ret := ""
	for i := 0; i < len(a); i++ {
		if len(ret) != 0 {
			ret += ", "
		}
		ch := string(a[i])
		if a[i] == '\\' {
			i++
			ch += string(a[i])
		}
		ret += fmt.Sprintf("'%s'", ch)
	}
	return `{
	accept = true
	n1 := []rune{` + ret + `}
	s := p.ParserData.Pos
	e := s + len(n1)
	if e > len(p.ParserData.Data) {
		accept = false
	} else {
		for i := 0; i < len(n1); i++ {
			if n1[i] != p.ParserData.Data[s+i] {
				accept = false
				break
			}
		}
	}
	if (accept) {
		p.ParserData.Pos += len(n1)
	}
}
`
}

func (g *GoGenerator2) AssertNot(a string) string {
	return `s := p.ParserData.Pos
` + g.Call(a) + `
p.ParserData.Pos = s
accept = !accept
`
}

func (g *GoGenerator2) AssertAnd(a string) string {
	return "p_And(p, " + g.Call(a) + ")"
}

func (g *GoGenerator2) ZeroOrMore(a string) string {
	return `func(p *` + g.Name + `) bool {
	accept := true
	` + g.Call(a) + `
	for accept {
		` + g.Call(a) + `
	}
	` + g.Return("true") + `
}`
}

func (g *GoGenerator2) OneOrMore(a string) string {
	return `func(p *` + g.Name + `) bool {
	save := p.ParserData.Pos
	accept := true
	` + g.Call(a) + `
	if !accept {
		p.ParserData.Pos = save
	` + g.Return("false") + `
	}
	for accept {
		` + g.Call(a) + `
	}
	` + g.Return("true") + `
}`
}

func (g *GoGenerator2) Maybe(a string) string {
	return g.Call(a) + "\naccept = true"
}

type needAllGroup struct {
	cf CodeFormatter
	g  Generator
}

func (b *needAllGroup) Add(value string) {
	b.cf.Add(b.g.Call(value) + `
if(!accept) {
	p.ParserData.Pos = save
	` + b.g.Return("false") + `
}
`)
}

type needOneGroup struct {
	cf CodeFormatter
	g  Generator
}

func (b *needOneGroup) Add(value string) {
	b.cf.Add(b.g.Call(value) + "\nif(accept) { " + b.g.Return("true") + " }\n")
}

func (g *GoGenerator2) BeginGroup(requireAll bool) Group {
	if requireAll {
		r := needAllGroup{g: g}
		r.cf.Add("func(p *" + g.Name + `) bool {
	accept := false
	save := p.ParserData.Pos
`)
		r.cf.Inc()
		return &r
	}
	r := needOneGroup{g: g}
	r.cf.Add("func(p *" + g.Name + `) bool {
	accept := false
	save := p.ParserData.Pos
`)
	r.cf.Inc()
	return &r
}

func (g *GoGenerator2) EndGroup(gr Group) string {
	switch t := gr.(type) {
	case *needAllGroup:
		t.cf.Add(g.Return("true") + "\n")
		t.cf.Dec()
		t.cf.Add("}")
		return t.cf.String()
	case *needOneGroup:
		t.cf.Add(`p.ParserData.Pos = save
` + g.Return("false") + "\n")
		t.cf.Dec()
		t.cf.Add("}")
		return t.cf.String()
	}
	panic(gr)
}

var inlinere = regexp.MustCompile(`^return ([\s\S]*?)\(p\)$`)
var inlinere2 = regexp.MustCompile(`^return (\w+\(p, [\s\S]*?\))($|\}\(p\)$)`)

func (g *GoGenerator2) MakeFunction(value string) string {
	if strings.HasSuffix(value, ")") || strings.HasSuffix(value, "}") {
		return value
	}

	if inlinere.MatchString(value) {
		return inlinere.FindStringSubmatch(value)[1]
	} else if inlinere2.MatchString(value) {
		return "func(p *" + g.Name + ") bool { " + value + " }"
	}
	f := CodeFormatter{}
	f.Add("func(p *" + g.Name + ") bool {\n")
	f.Inc()
	f.Add("accept := true\n" + value + "\n" + g.Return("accept") + "\n")
	f.Dec()
	f.Add("\n}")
	return f.String()
}
func (g *GoGenerator2) Return(value string) string {
	return "return " + value
}

var (
	callre1 = regexp.MustCompile(`^\s*accept\s`)
	callre2 = regexp.MustCompile(`^\s*func\(`)
)

func (g *GoGenerator2) Call(value string) string {
	pref := "accept = "
	if callre1.MatchString(value) {
		pref = ""
	}
	if strings.HasSuffix(value, "(p)") {
		return pref + value
	}
	if strings.HasPrefix(value, "p_") || callre2.MatchString(value) {
		return pref + value + "(p)"
	}
	return value
}

func (g *GoGenerator2) Begin() {
	imports := ""
	impList := g.Imports
	if g.AddDebugLogging {
		impList = append(impList, "log")
	}
	impList = append(impList, "parser")
	if len(impList) > 0 {
		imports += "\nimport (\n\t\"" + strings.Join(impList, "\"\n\t\"") + "\"\n)\n"
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
	members := g.ParserVariables
	members = append(members, `ParserData struct {
	Pos int
	Data []rune
}
`, "IgnoreRange parser.Range", "Root parser.Node")
	g.output += fmt.Sprintln("package " + strings.ToLower(g.Name) + "\n" + imports + "\ntype " + g.Name + " struct {\n\t" + strings.Join(members, "\n\t") + "\n}\n")
	if g.AddDebugLogging {
		g.output += "var fm parser.CodeFormatter\n\n"
	}
	g.output += `func (p *` + g.Name + `) RootNode() *parser.Node {
	return p.Root.Children[0]
}

func (p *` + g.Name + `) SetData(data string) {
	p.ParserData.Data = ([]rune)(data)
	p.Reset()
}

func (p *` + g.Name + `) Reset() {
	p.ParserData.Pos = 0
	p.Root = parser.Node{}
	p.IgnoreRange = parser.Range{}
}

func (p *` + g.Name + `) Data(start, end int) string {
	l := len(p.ParserData.Data)
	if l == 0 {
		return ""
	}
	if start < 0 {
		start = 0
	}
	if end > l {
		end = l
	}
	if start > end {
		return ""
	}
	return string(p.ParserData.Data[start:end])

}

func p_Ignore(p *` + g.Name + `, add func(*` + g.Name + `) bool) bool {
	start := p.ParserData.Pos
	ret := add(p)
	if ret {
		if start < p.IgnoreRange.Start || p.IgnoreRange.Start == 0 {
			p.IgnoreRange.Start = start
		}
		p.IgnoreRange.End = p.ParserData.Pos
	}
	return ret
}

func p_addNode(p *` + g.Name + `, add func(*` + g.Name + `) bool, name string) bool {
	start := p.ParserData.Pos
	shouldAdd := add(p)
	end := p.ParserData.Pos
	p.Root.P = p
	// Remove any danglers
	p.Root.Cleanup(p.ParserData.Pos, -1)

	node := p.Root.Cleanup(start, p.ParserData.Pos)
	node.Name = name
	if shouldAdd {
		node.P = p
		node.Range.Clip(p.IgnoreRange)
		c := make([]*parser.Node, len(node.Children))
		copy(c, node.Children)
		node.Children = c
		p.Root.Append(node)
	}
	if p.IgnoreRange.Start >= end || p.IgnoreRange.End <= start {
		p.IgnoreRange = parser.Range{}
	}
	return shouldAdd
}
`
}

func (g *GoGenerator2) Finish() string {
	ret := g.output
	if ret[len(ret)-2:] == "\n\n" {
		ret = ret[:len(ret)-1]
	}
	g.output = ""
	return ret
}