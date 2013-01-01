package parser

import (
	"fmt"
	"log"
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
			data = fmt.Sprintf(g.CustomActions[i].Action, g.MakeFunction(g.Return(data)))
			break
		}
	}
	if defaultAction {
		data = g.MakeFunction(g.Return(data))
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
		indenter.Add("return " + data + "\n")
	}
	indenter.Dec()
	indenter.Add("}\n\n")
	g.output += g.currentFunctions
	g.output += indenter.String()
	g.currentFunctions = ""
	g.currentFunctionsCount = 0
}
func (g *GoGenerator2) MakeParserCall(value string) string {
	return "p_" + value
}

func (g *GoGenerator2) CheckInRange(a, b string) string {
	return "p_InRange(p, '" + a + "', '" + b + "')"
}

func (g *GoGenerator2) CheckInSet(a string) string {
	a = strings.Replace(a, "\\[", "[", -1)
	a = strings.Replace(a, "\\]", "]", -1)
	a = strings.Replace(a, "\n", "\\n", -1)
	a = strings.Replace(a, "\r", "\\r", -1)
	a = strings.Replace(a, "\"", "\\\"", -1)
	return "p_InSet(p, \"" + a + "\")"
}

func (g *GoGenerator2) CheckAnyChar() string {
	return "p_AnyChar(p)"
}

func (g *GoGenerator2) CheckNext(a string) string {
	if a[0] == '\'' {
		return "p_NextRune(p, " + a + ")"
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
	return "p_Next(p, []rune{" + ret + "})"
}

func (g *GoGenerator2) AssertNot(a string) string {
	return "p_Not(p, " + a + ")"
}

func (g *GoGenerator2) AssertAnd(a string) string {
	return "p_And(p, " + a + ")"
}

func (g *GoGenerator2) ZeroOrMore(a string) string {
	return "p_ZeroOrMore(p, " + a + ")"
}

func (g *GoGenerator2) OneOrMore(a string) string {
	return "p_OneOrMore(p, " + a + ")"
}

func (g *GoGenerator2) Maybe(a string) string {
	return "p_Maybe(p, " + a + ")"
}

func (g *GoGenerator2) BeginGroup(requireAll bool) Group {
	r := baseGroup{}
	if requireAll {
		r.cf.Add("p_NeedAll(p, []func(*" + g.Name + ") bool{\n")
	} else {
		r.cf.Add("p_NeedOne(p, []func(*" + g.Name + ") bool{\n")
	}
	r.cf.Inc()
	return &r
}

func (g *GoGenerator2) EndGroup(gr Group) string {
	bg := gr.(*baseGroup)
	bg.cf.Dec()
	bg.cf.Add("})")
	return bg.cf.String()
}

var inlinere = regexp.MustCompile(`^return ([\s\S]*?)\(p\)$`)
var inlinere2 = regexp.MustCompile(`^return (\w+\(p, [^)]*\)|p.\w+\([^)]*\))($|\}\(p\)$)`)

func (g *GoGenerator2) MakeFunction(value string) string {
	if strings.HasSuffix(value, "}(p)") {
		m1 := inlinere.MatchString(value)
		mm1 := ""
		if m1 {
			mm1 = strings.Join(inlinere.FindStringSubmatch(value), "---")
		}
		log.Println(value, m1, mm1, inlinere2.MatchString(value))
	}
	if inlinere.MatchString(value) {
		return inlinere.FindStringSubmatch(value)[1]
	} else if inlinere2.MatchString(value) {
		return "func(p *" + g.Name + ") bool { " + value + " }"
	}
	fname := fmt.Sprintf("helper%d_%s", g.currentFunctionsCount, g.currentName)
	g.currentFunctionsCount++
	f := CodeFormatter{}
	f.Add("func " + fname + "(p *" + g.Name + ") bool {\n")
	f.Inc()
	f.Add(value)
	f.Dec()
	f.Add("\n}\n")
	g.currentFunctions += f.String()
	return fname
}
func (g *GoGenerator2) Return(value string) string {
	return "return " + value
}

func (g *GoGenerator2) Call(value string) string {
	return value + "(p)"
}

func (g *GoGenerator2) Begin() {
	imports := ""
	impList := g.Imports
	if g.AddDebugLogging {
		impList = append(impList, "log")
	}
	impList = append(impList, "parser", "strings")
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

func p_Maybe(p *` + g.Name + `, exp func(*` + g.Name + `) bool) bool {
	exp(p)
	return true
}

func p_OneOrMore(p *` + g.Name + `, exp func(*` + g.Name + `) bool) bool {
	save := p.ParserData.Pos
	if !exp(p) {
		p.ParserData.Pos = save
		return false
	}
	for exp(p) {
	}
	return true
}

func p_NeedAll(p *` + g.Name + `, exps []func(*` + g.Name + `) bool) bool {
	save := p.ParserData.Pos
	for _, exp := range exps {
		if !exp(p) {
			p.ParserData.Pos = save
			return false
		}
	}
	return true
}

func p_NeedOne(p *` + g.Name + `, exps []func(*` + g.Name + `) bool) bool {
	save := p.ParserData.Pos
	for _, exp := range exps {
		if exp(p) {
			return true
		}
	}
	p.ParserData.Pos = save
	return false
}

func p_ZeroOrMore(p *` + g.Name + `, exp func(*` + g.Name + `) bool) bool {
	for exp(p) {
	}
	return true
}

func p_And(p *` + g.Name + `, exp func(*` + g.Name + `) bool) bool {
	save := p.ParserData.Pos
	ret := exp(p)
	p.ParserData.Pos = save
	return ret
}

func p_Not(p *` + g.Name + `, exp func(*` + g.Name + `) bool) bool {
	return !p_And(p, exp)
}

func p_AnyChar(p *` + g.Name + `) bool {
	if p.ParserData.Pos >= len(p.ParserData.Data) {
		return false
	}
	p.ParserData.Pos++
	return true
}

func p_InRange(p *` + g.Name + `, c1, c2 rune) bool {
	if p.ParserData.Pos >= len(p.ParserData.Data) {
		return false
	}
	c := p.ParserData.Data[p.ParserData.Pos]
	if c >= c1 && c <= c2 {
		p.ParserData.Pos++
		return true
	}
	return false
}

func p_InSet(p *` + g.Name + `, dataset string) bool {
	if p.ParserData.Pos >= len(p.ParserData.Data) {
		return false
	}
	c := p.ParserData.Data[p.ParserData.Pos]
	if strings.ContainsRune(dataset, c) {
		p.ParserData.Pos++
		return true
	}
	return false
}

func p_NextRune(p *` + g.Name + `, n1 rune) bool {
	if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != n1 {
		return false
	}
	p.ParserData.Pos++
	return true
}

func p_Next(p *` + g.Name + `, n1 []rune) bool {
	s := p.ParserData.Pos
	e := s + len(n1)
	if e > len(p.ParserData.Data) {
		return false
	}
	for i := 0; i < len(n1); i++ {
		if n1[i] != p.ParserData.Data[s+i] {
			return false
		}
	}
	p.ParserData.Pos += len(n1)
	return true
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
