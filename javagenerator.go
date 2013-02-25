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

type JavaGenerator struct {
	s               GeneratorSettings
	output          string
	realOutput      string
	AddDebugLogging bool
	CustomActions   []CustomAction
	ParserVariables []string
	Imports         []string
	havefunctions   bool
	currentName     string
	saveCount       int
}

func (g *JavaGenerator) SetCustomActions(actions []CustomAction) {
	g.CustomActions = actions
}

func (g *JavaGenerator) AddNode(data, defName string) string {
	return `accept = true;
int start = parserData.pos;
` + g.Call(data) + `
int end = parserData.pos;
if (accept) {
    Node node = Root.Cleanup(start, parserData.pos);
    node.Name = "` + defName + `";
    node.Range.Clip(ignoreRange);
    Root.Append(node);
} else {
    Root.Discard(start);
}
if (ignoreRange.start >= end || ignoreRange.end <= start) {
    ignoreRange.Reset();
}`
}

func (g *JavaGenerator) Ignore(data string) string {
	return `accept = true;
int start = parserData.pos;
` + g.Call(data) + `
if (accept && start != parserData.pos) {
	if (start < ignoreRange.start || ignoreRange.start == 0) {
		ignoreRange.start = start;
	}
	ignoreRange.end = parserData.pos;
}
`
}
func (g *JavaGenerator) MakeParserFunction(node *Node) error {
	id := node.Children[0]
	exp := node.Children[len(node.Children)-1]
	defName := helper(g, id)
	g.currentName = defName
	g.saveCount = 0
	data := helper(g, exp)

	if !g.havefunctions {
		g.havefunctions = true
		g.output += "private boolean realParse() {\n\treturn " + defName + "();\n}\n"
	}

	indenter := CodeFormatter{}
	indenter.Inc()
	indenter.Add("private boolean " + defName + "() {\n")
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
		data = g.AddNode(data, defName)
	}
	if strings.HasPrefix(data, "accept") || data[0] == '{' {
		end := "return accept;\n"
		if data[len(data)-1] != '\n' {
			end = "\n" + end
		}
		indenter.Add("boolean accept = false;\n" + data + end)
	} else {
		indenter.Add("return " + data + ";\n")
	}
	indenter.Dec()
	indenter.Add("}\n\n")
	g.output += indenter.String()
	return nil
}

func (g *JavaGenerator) MakeParserCall(value string) string {
	return value
}

func (g *JavaGenerator) CheckInRange(a, b string) string {
	return `if (parserData.pos >= parserData.data.length) {
	accept = false;
} else {
	byte c = parserData.data[parserData.pos];
	if (c >= '` + a + `' && c <= '` + b + `') {
		parserData.pos++;
		accept = true;
	} else {
		accept = false;
	}
}`
}

func (g *JavaGenerator) CheckInSet(a string) string {
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
	accept = false;
	if (parserData.pos < parserData.data.length) {
		byte c =  parserData.data[parserData.pos];
		if (` + tests + `) {
			parserData.pos++;
			accept = true;
		}
	}
}`
}

func (g *JavaGenerator) CheckAnyChar() string {
	return `if (parserData.pos >= parserData.data.length) {
	accept = false;
} else {
	parserData.pos++;
	accept = true;
}`
}

func (g *JavaGenerator) CheckNext(a string) string {
	if a[0] == '\'' {
		return `if (parserData.pos >= parserData.data.length || parserData.data[parserData.pos] != ` + a + `) {
	accept = false;
} else {
	parserData.pos++;
	accept = true;
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
		tests += fmt.Sprintf("parserData.data[parserData.pos + %d] != '%s'", pos, c2)
	}
	mysave := fmt.Sprintf("save%d", g.saveCount)
	g.saveCount++
	mysave2 := fmt.Sprintf("save%d", g.saveCount)
	g.saveCount++

	return fmt.Sprintf(`{
	accept = true;
	int `+mysave+` = parserData.pos;
	int `+mysave2+` = `+mysave+` + %d;
	if (`+mysave2+` > parserData.data.length) {
		accept = false;
	} else {
		if (%s) {
			accept = false;
		}
	}
	if (accept) {
		parserData.pos += %d;
	}
}`, pos, tests, pos)
}

func (g *JavaGenerator) AssertNot(a string) string {
	mysave := fmt.Sprintf("save%d", g.saveCount)
	g.saveCount++

	return `int ` + mysave + ` = parserData.pos;
` + g.Call(a) + `
parserData.pos = ` + mysave + `;
accept = !accept;`
}

func (g *JavaGenerator) AssertAnd(a string) string {
	mysave := fmt.Sprintf("save%d", g.saveCount)
	g.saveCount++

	return `int ` + mysave + ` = parserData.pos;
` + g.Call(a) + `
parserData.pos = ` + mysave + `;`
}

func (g *JavaGenerator) ZeroOrMore(a string) string {
	var cf CodeFormatter
	cf.Add("{\n")
	cf.Inc()
	cf.Add("accept = true;")
	cf.Add("\nwhile (accept) {\n")
	cf.Inc()
	cf.Add(g.Call(a))
	cf.Dec()
	cf.Add("\n}\n")
	cf.Add("accept = true;\n")
	cf.Dec()
	cf.Add("}")
	return cf.String()
}

func (g *JavaGenerator) OneOrMore(a string) string {
	var cf CodeFormatter
	mysave := fmt.Sprintf("save_%d", g.saveCount)
	g.saveCount++
	cf.Add("{\n")
	cf.Inc()
	cf.Add(`int ` + mysave + ` = parserData.pos;
` + g.Call(a) + `
if (!accept) {
	parserData.pos = ` + mysave + `;
} else {
	while (accept) {
`)
	cf.Inc()
	cf.Inc()
	cf.Add(g.Call(a) + "\n")
	cf.Dec()
	cf.Add(`}
accept = true;
`)
	cf.Dec()
	cf.Add("}\n")
	cf.Dec()
	cf.Add("}")
	return cf.String()
}

func (g *JavaGenerator) Maybe(a string) string {
	return g.Call(a) + "\naccept = true;"
}

type jNeedAllGroup struct {
	cf     CodeFormatter
	g      Generator
	stack  list.List
	label  string
	mysave string
}

func (b *jNeedAllGroup) Add(value, name string) {
	b.cf.Add(b.g.Call(value) + `
if (accept) {
`)
	b.cf.Inc()
	b.stack.PushBack(name)
}

type jNeedOneGroup struct {
	cf     CodeFormatter
	g      Generator
	mysave string
}

func (b *jNeedOneGroup) Add(value, name string) {
	b.cf.Add(b.g.Call(value) + "\nif (!accept) {\n")
	b.cf.Inc()
}

func (g *JavaGenerator) BeginGroup(requireAll bool) Group {
	mysave := fmt.Sprintf("save_%d", g.saveCount)
	g.saveCount++

	if requireAll {
		r := jNeedAllGroup{g: g, mysave: mysave}
		r.cf.Add(`{
	int ` + mysave + ` = parserData.pos;
`)
		r.cf.Inc()
		return &r
	}
	r := jNeedOneGroup{g: g, mysave: mysave}
	r.cf.Add(`{
	int ` + mysave + ` = parserData.pos;
`)
	r.cf.Inc()
	return &r
}
func (g *JavaGenerator) UpdateError(msg string) string {
	return "{ /* if (LastError < parserData.pos) { LastError = parserData.pos; */ }"
	// return "{\n\te := fmt.Sprintf(`Expected " + msg + " near %d`, parserData.pos)\n\tif len(LastError) != 0 {\n\t\te = e + \"\\n\" + LastError\n\t}\n\tLastError = e\n}"
}
func (g *JavaGenerator) EndGroup(gr Group) string {
	switch t := gr.(type) {
	case *jNeedAllGroup:

		for n := t.stack.Back(); len(t.cf.Level()) > 1; n = n.Prev() {
			t.cf.Dec()
			t.cf.Add("}\n")
		}
		t.cf.Add("if (!accept) {\n")
		t.cf.Inc()
		t.cf.Add(g.UpdateError("TODO") + "\nparserData.pos = " + t.mysave + ";\n")
		t.cf.Dec()
		t.cf.Add("}\n")
		t.cf.Dec()
		t.cf.Add("}")
		return t.cf.String()
	case *jNeedOneGroup:
		for len(t.cf.Level()) > 1 {
			t.cf.Dec()
			t.cf.Add("}\n")
		}
		t.cf.Add("if (!accept) {\n\tparserData.pos = " + t.mysave + ";\n}\n")
		t.cf.Dec()
		t.cf.Add("}")
		return t.cf.String()
	}
	panic(gr)
}

var javare = regexp.MustCompile("(^if |(}|;)$)")

func (g *JavaGenerator) Call(value string) string {
	pref := "accept = "
	if callre1.MatchString(value) {
		pref = ""
	}
	if strings.HasSuffix(value, "();") {
		return pref + value
	}
	if javare.MatchString(value) {
		return value
	}
	return pref + value + "();"
}

func (g *JavaGenerator) Begin(s GeneratorSettings) error {
	g.s = s
	dumptree_s := ""
	if g.s.Debug {
		dumptree_s = "p.Root.print(p, \"\");"
	}
	members := g.ParserVariables
	members = append(members, `class Something {
        int pos;
        byte data[];
        int end;
    };
    Something parserData = new Something();

`, "Range ignoreRange = new Range();",
		"Node Root = new Node();",
		"int LastError;")
	g.realOutput += g.s.Header + `
import java.io.*;

class NodeContainer {
    public NodeContainer()
    {
        _size = 0;
        _capacity = 0;
    }
    public void resize(int newsize)
    {
        if (newsize > _capacity) {
            Node newcontent[] = new Node[newsize];
            System.arraycopy(_content, 0, newcontent, 0, _size);
            _capacity = _size = newsize;
            _content = newcontent;
        } else {
            _size = newsize;
        }
    }
    public void set(int i, Node data) {
        _content[i] = data;
    }
    public Node get(int i)
    {
        return _content[i];
    }

    public void pop_back() { _size--; }
    public void push_back(Node n) {
        resize(_size+1);
        _size--;
         _content[_size] = n;
         _size++;
    }
    public Node back() { return _content[_size-1]; }
    public int size() { return _size; }
    private int _size;
    private int _capacity;
    private Node _content[] = new Node[0];
}

class Range {
    public void Clip(Range other)
    {
        if (start >= other.start && start < other.end) {
            start = other.end;
        }
        if (end >= other.start && end <= other.end) {
            end = other.start;
        }
        if (end < start) {
            end = start;
        }
    }
    public void Reset() { start = end = 0; }
    public int start;
    public int end;
}

class Node {
    public NodeContainer Children = new NodeContainer();
    public String Name = new String("");
    public Range Range = new Range();

    public void print(` + g.s.Name + ` p, String indent) {
        if (Children.size() != 0) {
            System.out.println(String.format("%s%d-%d: \"%s\"", indent, Range.start, Range.end, Name));
            indent += "\t";
            for (int i = 0; i < Children.size(); i++) {
                Children.get(i).print(p, indent);
            }
        } else {
             System.out.println(String.format("%s%d-%d: \"%s\" - Data: \"%s\"", indent, Range.start, Range.end, Name, p.Data(Range.start, Range.end)));
        }

    }
    public Node Cleanup(int pos, int end)
    {
        Node popped = new Node();
        popped.Range.start = pos;
        popped.Range.end = end;
        int popIdx = 0;

        for (int i = Children.size() - 1; i >= 0; i--) {
            if (Children.get(i).Range.end <= pos) {
                popIdx = i+1;
                break;
            }
        }
        int num = Children.size()-popIdx;
        if (num > 0) {
            popped.Children.resize(num);
            int j = 0;
            int i = popIdx;
            for (int b = Children.size()-1; b >= popIdx; b--, i++, j++) {
                popped.Children.set(j, Children.get(i));
            }
            Children.resize(popIdx);
        }

        return popped;
    }
    public void Discard(int pos)
    {
        if (Children.size() == 0)
            return;
        int popIdx = 0;
        for (int i = Children.size()-1; i >= 0; i--) {
            if (Children.get(i).Range.end <= pos) {
                popIdx = i+1;
                break;
            }
        }
        Children.resize(popIdx);
    }
    public void Append(Node n)
    {
        Children.push_back(n);
    }
}
`

	g.output += `
class ` + g.s.Name + ` {
    ` + strings.Join(members, "\n\t") + `
    boolean Parse(byte data[]) {
        parserData.data = data;
        parserData.pos = 0;
        Root = new Node();
        ignoreRange.Reset();
        LastError = 0;
        boolean ret = realParse();
        Root.Name = "` + g.s.Name + `";
        Root.Range.start = 0;
        Root.Range.end = parserData.data.length;
        return ret;
    }

    String Data(int start, int end) {
        if (start < 0) {
            start = 0;
        }
        if (end > parserData.data.length) {
            end = parserData.data.length;
        }
        if (start > end) {
            return "";
        }
        return new String(parserData.data, start, end-start);
    }

    public static void main(String argv[]) {
        try {
            FileInputStream fs = new FileInputStream("` + g.s.Testname + `");
            byte data[] = new byte[fs.available()];
            fs.read(data);

            {{ParserName}} p = new {{ParserName}}();
            if (p.Parse(data)) {
                ` + dumptree_s + `
            } else {
                System.out.println("Failed to parse");
                System.exit(1);
            }`

	if g.s.Bench {
		g.output += `
            int N = 1000;

            long t = System.currentTimeMillis();
            for (int i = 0; i < N; i++) {
        	    p.Parse(data);
            }
            long t2 = System.currentTimeMillis();
            double perf = (t2-t) * 10e5;
            System.out.println("Finished in " + (perf/10e8) + " seconds, " + (perf/N) +  " ns/op");`
	}
	g.output += `
        } catch (Exception e) {
            System.err.println(e);
            e.printStackTrace();
        }
    }
`
	return nil
}

func (g *JavaGenerator) Finish() error {
	ret := strings.Replace(g.realOutput+g.output+"\n}", "{{ParserName}}", g.s.Name, -1)
	if ret[len(ret)-2:] == "\n\n" {
		ret = ret[:len(ret)-1]
	}
	g.realOutput = ""
	g.output = ""

	if err := g.s.WriteFile(g.s.Name+".java", ret); err != nil {
		return err
	}
	return nil
}

func (g *JavaGenerator) TestCommand() []string {
	return []string{"bash", "-c", "javac ./" + g.s.Name + ".java && java " + g.s.Name}
}
