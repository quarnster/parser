package parser

import (
	"fmt"
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

	if g.AddDebugLogging {
		indenter.Add(`var (
    pos = p.Pos()
    l   = p.ParserData.Len()
)
`)
		indenter.Add(`log.Println(fm.level + "` + defName + " entered\")\n")
		indenter.Add("fm.Inc()\n")
	}
	defaultAction := defName[:1] == strings.ToUpper(defName[:1])
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
		indenter.Add("res := " + data)
		indenter.Add("\nfm.Dec()\n")
		indenter.Add(`if !res && p.Pos() != pos {
    log.Fatalln("` + defName + `", res, ", ", pos, ", ", p.Pos())
}
p2 := p.Pos()
data := make([]byte, p2-pos)
p.ParserData.Seek(int64(pos), 0)
p.ParserData.Read(data)
p.ParserData.Seek(int64(p2), 0)
`)
		indenter.Add("log.Println(fm.level+\"" + defName + ` returned: ", res, ", ", pos, ", ", p.Pos(), ", ", l, string(data))` + "\n")
		indenter.Add("return res\n")
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
	a = "\"" + strings.Replace(a[1:len(a)-1], "\"", "\\\"", -1) + "\""
	return "p_Next(p, " + a + ")"
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

func (g *GoGenerator2) MakeFunction(value string) string {
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
	members = append(members, "parser.Parser")
	g.output += fmt.Sprintln("package " + strings.ToLower(g.Name) + "\n" + imports + "\ntype " + g.Name + " struct {\n\t" + strings.Join(members, "\n\t") + "\n}\n")
	if g.AddDebugLogging {
		g.output += "var fm CodeFormatter\n\n"
	}
	g.output += `
const (
    nilrune = '\u0000'
)

func p_addNode(p *` + g.Name + `, add func(*` + g.Name + `) bool, name string) bool {
    start := p.ParserData.Pos
    shouldAdd := add(p)
    p.Root.P = &p.Parser
    // Remove any danglers
    p.Root.Cleanup(p.ParserData.Pos, -1)

    node := p.Root.Cleanup(start, p.ParserData.Pos)
    node.Name = name
    if shouldAdd {
        node.P = &p.Parser
        p.Root.Append(node)
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
    return p.AnyChar()
}

func p_InRange(p *` + g.Name + `, c1, c2 rune) bool {
    return p.InRange(c1, c2)
}

func p_InSet(p *` + g.Name + `, dataset string) bool {
    return p.InSet(dataset)
}

func p_Next(p *` + g.Name + `, n1 string) bool {
    return p.Next(n1)
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
