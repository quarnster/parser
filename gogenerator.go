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
	"regexp"
	"strings"
)

type GoGenerator struct {
	s                     GeneratorSettings
	output                string
	AddDebugLogging       bool
	CustomActions         []CustomAction
	ParserVariables       []string
	Imports               []string
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

func (g *GoGenerator) SetCustomActions(actions []CustomAction) {
	g.CustomActions = actions
}

func (g *GoGenerator) AddNode(data, defName string) string {
	ret := `accept = true
start := p.ParserData.Pos
` + g.Call(data) + `
end := p.ParserData.Pos
if accept {
`
	if g.calledP || true {
		ret += `	node := p.Root.Cleanup(start, end)
	node.Name = "` + defName + `"
	node.P = p
	node.Range.Clip(p.IgnoreRange)
	p.Root.Append(node)
} else {
	p.Root.Discard(start)`
	} else {
		ret += `	node := &Node{Range:Range{start,end}}
	node.Name = "` + defName + `"
	node.P = p
	node.Range.Clip(p.IgnoreRange)
	p.Root.Append(node)`
	}
	ret += `
}
if p.IgnoreRange.Start >= end || p.IgnoreRange.End <= start {
	p.IgnoreRange = Range{}
}
`
	return ret
}

func (g *GoGenerator) Ignore(data string) string {
	return `accept = true
start := p.ParserData.Pos
` + g.Call(data) + `
if accept && start != p.ParserData.Pos {
	if start < p.IgnoreRange.Start || p.IgnoreRange.Start == 0 {
		p.IgnoreRange.Start = start
	}
	p.IgnoreRange.End = p.ParserData.Pos
}
`
}
func (g *GoGenerator) MakeParserFunction(node *Node) error {
	g.calledP = false
	id := node.Children[0]
	exp := node.Children[len(node.Children)-1]
	defName := helper(g, id)
	g.currentName = defName
	data := helper(g, exp)

	if !g.havefunctions {
		g.havefunctions = true
		g.output += "func (p *" + g.s.Name + ") realParse() bool {\n\treturn p_" + defName + "(p)\n}\n"
	}

	indenter := CodeFormatter{}
	indenter.Add("func p_" + defName + "(p *" + g.s.Name + ") bool {\n")
	indenter.Inc()
	indenter.Add("// " + strings.Replace(strings.TrimSpace(node.Data()), "\n", "\n// ", -1) + "\n")
	if g.s.Heatmap {
		indenter.Add(`
key := fmt.Sprintf("` + defName + `-%d", p.ParserData.Pos)
hs := time.Now()
ov := p.Heatmap["` + defName + `"]
defer func() {
	dur := time.Since(hs)
	v := p.Heatmap[key]
	v.Calls++
	v.Time += dur
	p.Heatmap[key] = v
	v = p.Heatmap["` + defName + `"]
	v.Calls++
	v.Time = ov.Time + dur
	p.Heatmap["` + defName + `"] = v
}()
`)
	}

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
	if g.AddDebugLogging {
		if strings.HasPrefix(data, "accept") || data[0] == '{' {
			data = "accept := false\n" + data
		} else {
			data = "accept := " + data
		}
		indenter.Add(`var (
	pos = p.ParserData.Pos
	l   = len(p.ParserData.Data)
)

log.Println(fm.Level() + "` + defName + ` entered")
fm.Inc()
` + data + `
fm.Dec()
if !accept && p.ParserData.Pos != pos {
	log.Fatalln("` + defName + `", accept, ", ", pos, ", ", p.ParserData.Pos)
}
p2 := p.ParserData.Pos
data := ""
if p2 < len(p.ParserData.Data) {
	data = string(p.ParserData.Data[pos:p2])
}
log.Println(fm.Level()+"` + defName + ` returned: ", accept, ", ", pos, ", ", p.ParserData.Pos, ", ", l, string(data))
return accept
`)
	} else {
		if strings.HasPrefix(data, "accept") || data[0] == '{' {
			end := "return accept\n"
			if data[len(data)-1] != '\n' {
				end = "\n" + end
			}
			indenter.Add("accept := false\n" + data + end)
		} else {
			indenter.Add("return " + data + "\n")
		}
	}
	indenter.Dec()
	indenter.Add("}\n\n")
	g.output += g.currentFunctions
	g.output += indenter.String()
	g.currentFunctions = ""
	g.currentFunctionsCount = 0
	return nil
}

func (g *GoGenerator) MakeParserCall(value string) string {
	g.calledP = true
	// if g.inlineCount < 0 && g.RootNode != nil {
	// 	g.inlineCount++
	// 	for _, child := range g.RootNode.Children {
	// 		if child.Name == "Definition" && child.Children[0].Data() == value {
	// 			data := helper(g, child.Children[len(child.Children)-1])
	// 			defaultAction := true
	// 			for i := range g.CustomActions {
	// 				if value == g.CustomActions[i].Name {
	// 					defaultAction = false
	// 					data = g.CustomActions[i].Action(g, data)
	// 					break
	// 				}
	// 			}
	// 			if defaultAction {
	// 				data = g.AddNode(data, value)
	// 			}

	// 			ret := "/* inlined " + value + "*/\n" + data
	// 			g.inlineCount--
	// 			return ret
	// 		}
	// 	}
	// 	g.inlineCount--
	// }

	return "p_" + value
}

func (g *GoGenerator) CheckInRange(a, b string) string {
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
}`
}

func (g *GoGenerator) CheckInSet(a string) string {
	a = strings.Replace(a, "\\[", "[", -1)
	a = strings.Replace(a, "\\]", "]", -1)
	tests := ""
	for i := 0; i < len(a); i++ {
		if len(tests) > 0 {
			tests += " || "
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
	return `{
	accept = false
	if p.ParserData.Pos < len(p.ParserData.Data) {
		c := p.ParserData.Data[p.ParserData.Pos]
		if ` + tests + ` {
			p.ParserData.Pos++
			accept = true
		}
	}
}`
}

func (g *GoGenerator) CheckAnyChar() string {
	return `if p.ParserData.Pos >= len(p.ParserData.Data) {
	accept = false
} else {
	p.ParserData.Pos++
	accept = true
}`
}

func (g *GoGenerator) CheckNext(a string) string {
	/*
	 */

	if a[0] == '\'' {
		return `if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != ` + a + ` {
	accept = false
} else {
	p.ParserData.Pos++
	accept = true
}`
	}
	a = a[1 : len(a)-1]
	tests := ""
	pos := 0
	for i := 0; i < len(a); i, pos = i+1, pos+1 {
		if len(tests) > 0 {
			tests += " || "
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
	return fmt.Sprintf(`{
	accept = true
	s := p.ParserData.Pos
	e := s + %d
	if e > len(p.ParserData.Data) {
		accept = false
	} else {
		if %s {
			accept = false
		}
	}
	if accept {
		p.ParserData.Pos += %d
	}
}`, pos, tests, pos)
}

func (g *GoGenerator) AssertNot(a string) string {
	return `s := p.ParserData.Pos
` + g.Call(a) + `
p.ParserData.Pos = s
p.Root.Discard(s)
accept = !accept`
}

func (g *GoGenerator) AssertAnd(a string) string {
	return `s := p.ParserData.Pos
` + g.Call(a) + `
p.ParserData.Pos = s
p.Root.Discard(s)`
}

func (g *GoGenerator) ZeroOrMore(a string) string {
	var cf CodeFormatter
	cf.Add("{\n")
	cf.Inc()
	cf.Add("accept = true")
	cf.Add("\nfor accept {\n")
	cf.Inc()
	cf.Add(g.Call(a))
	cf.Dec()
	cf.Add("\n}\n")
	cf.Add("accept = true\n")
	cf.Dec()
	cf.Add("}")
	return cf.String()
}

func (g *GoGenerator) OneOrMore(a string) string {
	var cf CodeFormatter
	cf.Add("{\n")
	cf.Inc()
	cf.Add(`save := p.ParserData.Pos
` + g.Call(a) + `
if !accept {
	p.ParserData.Pos = save
} else {
	for accept {
`)
	cf.Inc()
	cf.Inc()
	cf.Add(g.Call(a) + "\n")
	cf.Dec()
	cf.Add(`}
accept = true
`)
	cf.Dec()
	cf.Add("}\n")
	cf.Dec()
	cf.Add("}")
	return cf.String()
}

func (g *GoGenerator) Maybe(a string) string {
	return g.Call(a) + "\naccept = true"
}

type needAllGroup struct {
	cf    CodeFormatter
	g     Generator
	stack list.List
	label string
}

func (b *needAllGroup) Add(value, name string) {
	b.cf.Add(b.g.Call(value) + `
if accept {
`)
	b.cf.Inc()
	b.stack.PushBack(name)
}

type needOneGroup struct {
	cf CodeFormatter
	g  Generator
}

func (b *needOneGroup) Add(value, name string) {
	b.cf.Add(b.g.Call(value) + "\nif !accept {\n")
	b.cf.Inc()
}

func (g *GoGenerator) BeginGroup(requireAll bool) Group {
	if requireAll {
		r := needAllGroup{g: g}
		r.cf.Add(`{
	save := p.ParserData.Pos
`)
		r.cf.Inc()
		return &r
	}
	r := needOneGroup{g: g}
	r.cf.Add(`{
	save := p.ParserData.Pos
`)
	r.cf.Inc()
	return &r
}
func (g *GoGenerator) UpdateError(msg string) string {
	return `if p.LastError < p.ParserData.Pos {
	p.LastError = p.ParserData.Pos
}`
	// return "{\n\te := fmt.Sprintf(`Expected " + msg + " near %d`, p.ParserData.Pos)\n\tif len(p.LastError) != 0 {\n\t\te = e + \"\\n\" + p.LastError\n\t}\n\tp.LastError = e\n}"
}
func (g *GoGenerator) EndGroup(gr Group) string {
	switch t := gr.(type) {
	case *needAllGroup:

		for n := t.stack.Back(); len(t.cf.Level()) > 1; n = n.Prev() {
			t.cf.Dec()
			t.cf.Add("}\n")
		}
		t.cf.Add("if !accept {\n")
		t.cf.Inc()
		t.cf.Add(g.UpdateError("TODO") + "\np.ParserData.Pos = save\n")
		t.cf.Dec()
		t.cf.Add("}\n")
		t.cf.Dec()
		t.cf.Add("}")
		return t.cf.String()
	case *needOneGroup:
		for len(t.cf.Level()) > 1 {
			t.cf.Dec()
			t.cf.Add("}\n")
		}
		t.cf.Add("if !accept {\n\tp.ParserData.Pos = save\n}\n")
		t.cf.Dec()
		t.cf.Add("}")
		return t.cf.String()
	}
	panic(gr)
}

var (
	callre1 = regexp.MustCompile(`^\s*accept\s`)
	callre2 = regexp.MustCompile(`^\s*func\(`)
)

func (g *GoGenerator) Call(value string) string {
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

func (g *GoGenerator) Begin(s GeneratorSettings) error {
	g.s = s
	imports := "\n\nimport (\n\t\"bytes\"\n\t. \"github.com/quarnster/parser\"\n"
	impList := g.Imports
	members := g.ParserVariables
	if g.s.Heatmap {
		members = append(members, "Heatmap map[string]Heat")
		impList = append(impList, "fmt", "time", "sort")
	}
	if g.AddDebugLogging {
		impList = append(impList, "log")
	}
	if len(impList) > 0 {
		imports += "\t\"" + strings.Join(impList, "\"\n\t\"") + "\"\n"
	}
	imports += ")\n"

	g.output = g.s.Header + "\n"
	members = append(members, `ParserData struct {
		Pos  int
		Data []byte
	}
`, "IgnoreRange Range",
		"Root        Node",
		"LastError   int")
	g.output += fmt.Sprintln("package " + strings.ToLower(g.s.Name) + imports + "\ntype " + g.s.Name + " struct {\n\t" + strings.Join(members, "\n\t") + "\n}\n")

	if g.AddDebugLogging {
		g.output += "var fm CodeFormatter\n\n"
	}
	if g.s.Heatmap {
		g.output += `type Heat struct {
	Name string
	Calls int
	Time time.Duration
}

type TotHeat struct {
	Heat []Heat
}

func (t *TotHeat) Len() int {
	return len(t.Heat)
}

func (t *TotHeat) Less(i, j int) bool {
	return t.Heat[i].Time >= t.Heat[j].Time
}

func (t *TotHeat) Swap(i, j int) {
	t.Heat[i], t.Heat[j] = t.Heat[j], t.Heat[i]
}

func (t *TotHeat) Add(h Heat) {
	t.Heat = append(t.Heat, h)
}

func (t *TotHeat) String() (ret string) {
	sort.Sort(t)
	for _, h := range t.Heat {
		ret += fmt.Sprintf("%v\n", h)
	}
	return ret
}
`
	}
	g.output += `func (p *` + g.s.Name + `) RootNode() *Node {
	return &p.Root
}

func (p *` + g.s.Name + `) Parse(data string) bool {
	p.ParserData.Data = ([]byte)(data)
`
	if g.s.Heatmap {
		g.output += "	p.Heatmap = make(map[string]Heat)\n"
	}
	g.output += `	p.ParserData.Pos = 0
	p.Root = Node{Name: "` + g.s.Name + `", P: p}
	p.IgnoreRange = Range{}
	p.LastError = 0
	ret := p.realParse()
	p.Root.UpdateRange()
	return ret
}

func (p *` + g.s.Name + `) Data(start, end int) string {
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

func (p *` + g.s.Name + `) Error() Error {
	errstr := ""

	line := 1
	column := 1
	for _, r := range p.ParserData.Data[:p.LastError] {
		column++
		if r == '\n' {
			line++
			column = 1
		}
	}

	if p.LastError == len(p.ParserData.Data) {
		errstr = "Unexpected EOF"
	} else {
		e := p.LastError + 4
		if e > len(p.ParserData.Data) {
			e = len(p.ParserData.Data)
		}

		reader := bytes.NewReader(p.ParserData.Data[p.LastError:e])
		r, _, _ := reader.ReadRune()
		if r == '\r' || r == '\n' {
			errstr = "Unexpected new line"
		} else {
			errstr = "Unexpected " + string(r)
		}
	}
	return NewError(line, column, errstr)
}

`
	return nil
}

func (g *GoGenerator) Finish() error {
	ret := g.output
	if ret[len(ret)-2:] == "\n\n" {
		ret = ret[:len(ret)-1]
	}
	g.output = ""
	ln := strings.ToLower(g.s.Name)
	if err := g.s.WriteFile(ln+".go", ret); err != nil {
		return err
	}

	dumptree_s := ""
	heatmap_s := ""
	if g.s.Debug {
		dumptree_s = "t.Log(\"\\n\"+root.String())"
	}
	if g.s.Heatmap {
		heatmap_s = `var wasted time.Duration
			var th TotHeat
			for k, v := range p.Heatmap {
				if strings.Contains(k, "-") {
					wasted += time.Duration(int64(v.Calls-1)*int64(v.Time) / int64(v.Calls))
				} else {
					v.Name = k
					th.Add(v)
				}
			}
			t.Logf("Wasted %s", wasted)
			t.Log(&th)
			`
	}
	test := `package ` + strings.ToLower(g.s.Name) + `
import (
	"archive/zip"
	"io/ioutil"
	"net/http"
	"testing"
	"strings"
	"time"
)

var _ = time.Time{}
const testname = "` + g.s.Testname + `"

func loadData(path string) (retdata string, err error) {
	var data []byte
	if strings.HasPrefix(path, "http://") {
		p2 := strings.Replace(strings.Replace(path, "http://", "http_", -1), "/", "_", -1)
		if retdata, err = loadData(p2); err != nil {
			if res, err := http.Get(path); err != nil {
				return "", err
			} else {
				defer res.Body.Close()
				if data, err = ioutil.ReadAll(res.Body); err != nil {
					return "", err
				} else if err = ioutil.WriteFile(p2, data, 0644); err != nil {
					return "", err
				} else {
					return loadData(p2)
				}
			}
		} else {
			return
		}
	} else if strings.HasSuffix(path, ".zip") {
		if zf, err := zip.OpenReader(path); err != nil {
			return "", err
		} else {
			defer zf.Close()
			f, err := zf.File[0].Open()
			if err != nil {
				return "", err
			}
			defer f.Close()
			if data, err = ioutil.ReadAll(f); err != nil {
				return "", err
			}
		}
	} else {
		data, err = ioutil.ReadFile(path)
	}
	return string(data), err
}

func TestParser(t *testing.T) {
	var p ` + g.s.Name + `
	if data, err := loadData(testname); err != nil {
		t.Fatal(err)
	} else {
		if !p.Parse(data) {
			t.Fatalf("Didn't parse correctly: %s\n", p.Error())
		} else {
			root := p.RootNode()
			` + dumptree_s + `
			` + heatmap_s + `
			if root.Range.End != len(p.ParserData.Data) {
				t.Fatalf("Parsing didn't finish: %v\n%s", root, p.Error())
			}
		}
	}
}

func BenchmarkParser(b *testing.B) {
	var p ` + g.s.Name + `
	if data, err := loadData(testname); err != nil {
		b.Fatal(err)
	} else {
		for i := 0; i < b.N; i++ {
			p.Parse(data)
		}
	}
}

`
	if err := g.s.WriteFile(ln+"_test.go", test); err != nil {
		return err
	}
	return nil
}

func (g *GoGenerator) TestCommand() []string {
	cmd := []string{"go", "test", "-v", "-gcflags", "-B"}
	if g.s.Bench {
		cmd = append(cmd, "-bench", ".")
	}
	return cmd
}
