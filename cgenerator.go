package parser

import (
	"container/list"
	"fmt"
	//	"regexp"
	"strings"
)

type CGenerator struct {
	output          string
	realOutput      string
	AddDebugLogging bool
	Name            string
	CustomActions   []CustomAction
	ParserVariables []string
	Imports         []string
	havefunctions   bool
	currentName     string
}

func (g *CGenerator) AddNode(data, defName string) string {
	return `accept = true;
const char* __restrict__ start = p->ParserData.Pos;
` + g.Call(data) + `
const char* __restrict__ end = p->ParserData.Pos;
p->Root.P = p;
if (accept) {
	NodeMember node = p->Root.Cleanup(start, p->ParserData.Pos);
	node->Name = "` + defName + `";
	node->P = p;
	node->Range.Clip(p->IgnoreRange);
	p->Root.Append(node);
} else {
	p->Root.Discard(start);
}
if (p->IgnoreRange.Start >= end || p->IgnoreRange.End <= start) {
	p->IgnoreRange.Reset();
}`
}

func (g *CGenerator) Ignore(data string) string {
	return `accept = true;
const char* __restrict__ start = p->ParserData.Pos;
` + g.Call(data) + `
if (accept && start != p->ParserData.Pos) {
	if (start < p->IgnoreRange.Start || p->IgnoreRange.Start == 0) {
		p->IgnoreRange.Start = start;
	}
	p->IgnoreRange.End = p->ParserData.Pos;
}
`
}
func (g *CGenerator) MakeParserFunction(node *Node) {
	id := node.Children[0]
	exp := node.Children[len(node.Children)-1]
	defName := helper(g, id)
	g.currentName = defName
	data := helper(g, exp)

	g.realOutput += "static bool p_" + defName + "(" + g.Name + "*);\n"

	if !g.havefunctions {
		g.havefunctions = true
		g.output += "bool " + g.Name + "::Parse() {\n\treturn p_" + defName + "(this);\n}\n"
	}

	indenter := CodeFormatter{}
	indenter.Add("static bool p_" + defName + "(" + g.Name + "  * __restrict__ p) {\n")
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
	if g.AddDebugLogging {
		// 		if strings.HasPrefix(data, "accept") || data[0] == '{' {
		// 			data = "bool accept = false;\n" + data
		// 		} else {
		// 			data = "bool accept = " + data + ";"
		// 		}
		// 		indenter.Add(`int pos = p->ParserData.Pos;
		// 	int l   = len(p->ParserData.Data);

		// log.Println(fm.Level() + "` + defName + ` entered")
		// fm.Inc()
		// ` + data + `
		// fm.Dec()
		// if !accept && p->ParserData.Pos != pos {
		// 	log.Fatalln("` + defName + `", accept, ", ", pos, ", ", p->ParserData.Pos)
		// }
		// p2 := p->ParserData.Pos
		// data := ""
		// if p2 < len(p->ParserData.Data) {
		// 	data = string(p->ParserData.Data[pos:p2])
		// }
		// log.Println(fm.Level()+"` + defName + ` returned: ", accept, ", ", pos, ", ", p->ParserData.Pos, ", ", l, string(data))
		// return accept
		// `)
	} else {
		if strings.HasPrefix(data, "accept") || data[0] == '{' {
			end := "return accept;\n"
			if data[len(data)-1] != '\n' {
				end = "\n" + end
			}
			indenter.Add("bool accept = false;\n" + data + end)
		} else {
			indenter.Add(g.Return(data) + "\n")
		}
	}
	indenter.Dec()
	indenter.Add("}\n\n")
	g.output += indenter.String()
}

func (g *CGenerator) MakeParserCall(value string) string {
	return "p_" + value
}

func (g *CGenerator) CheckInRange(a, b string) string {
	return `if (p->ParserData.Pos >= p->ParserData.Len) {
	accept = false;
} else {
	const char c = *p->ParserData.Pos;
	if (c >= '` + a + `' && c <= '` + b + `') {
		p->ParserData.Pos++;
		accept = true;
	} else {
		accept = false;
	}
}`
}

func (g *CGenerator) CheckInSet(a string) string {
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
	if (p->ParserData.Pos < p->ParserData.Len) {
		const char c =  *p->ParserData.Pos;
		if (` + tests + `) {
			p->ParserData.Pos++;
			accept = true;
		}
	}
}`
}

func (g *CGenerator) CheckAnyChar() string {
	return `if (p->ParserData.Pos >= p->ParserData.Len) {
	accept = false;
} else {
	p->ParserData.Pos++;
	accept = true;
}`
}

func (g *CGenerator) CheckNext(a string) string {
	/*
	 */

	if a[0] == '\'' {
		return `if (p->ParserData.Pos >= p->ParserData.Len || *p->ParserData.Pos != ` + a + `) {
	accept = false;
} else {
	p->ParserData.Pos++;
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
		tests += fmt.Sprintf("*(p->ParserData.Pos + %d) != '%s'", pos, c2)
	}
	return fmt.Sprintf(`{
	accept = true;
	const char* __restrict__ s = p->ParserData.Pos;
	const char* __restrict__ e = s + %d;
	if (e > p->ParserData.Len) {
		accept = false;
	} else {
		if (%s) {
			accept = false;
		}
	}
	if (accept) {
		p->ParserData.Pos += %d;
	}
}`, pos, tests, pos)
}

func (g *CGenerator) AssertNot(a string) string {
	return `const char* __restrict__ s = p->ParserData.Pos;
` + g.Call(a) + `
p->ParserData.Pos = s;
accept = !accept;`
}

func (g *CGenerator) AssertAnd(a string) string {
	return `const char* __restrict__ s = p->ParserData.Pos;
` + g.Call(a) + `
p->ParserData.Pos = s;`
}

func (g *CGenerator) ZeroOrMore(a string) string {
	var cf CodeFormatter
	cf.Add("{\n")
	cf.Inc()
	cf.Add(g.Call(a))
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

func (g *CGenerator) OneOrMore(a string) string {
	var cf CodeFormatter
	cf.Add("{\n")
	cf.Inc()
	cf.Add(`const char* __restrict__ save = p->ParserData.Pos;
` + g.Call(a) + `
if (!accept) {
	p->ParserData.Pos = save;
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

func (g *CGenerator) Maybe(a string) string {
	return g.Call(a) + "\naccept = true;"
}

type cNeedAllGroup struct {
	cf    CodeFormatter
	g     Generator
	stack list.List
	label string
}

func (b *cNeedAllGroup) Add(value, name string) {
	b.cf.Add(b.g.Call(value) + `
if (accept) {
`)
	b.cf.Inc()
	b.stack.PushBack(name)
}

type cNeedOneGroup struct {
	cf CodeFormatter
	g  Generator
}

func (b *cNeedOneGroup) Add(value, name string) {
	b.cf.Add(b.g.Call(value) + "\nif (!accept) {\n")
	b.cf.Inc()
}

func (g *CGenerator) BeginGroup(requireAll bool) Group {
	if requireAll {
		r := cNeedAllGroup{g: g}
		r.cf.Add(`{
	const char* __restrict__ save = p->ParserData.Pos;
`)
		r.cf.Inc()
		return &r
	}
	r := cNeedOneGroup{g: g}
	r.cf.Add(`{
	const char* __restrict__ save = p->ParserData.Pos;
`)
	r.cf.Inc()
	return &r
}
func (g *CGenerator) UpdateError(msg string) string {
	return "if (p->LastError < p->ParserData.Pos) { p->LastError = p->ParserData.Pos; } "
	// return "{\n\te := fmt.Sprintf(`Expected " + msg + " near %d`, p->ParserData.Pos)\n\tif len(p.LastError) != 0 {\n\t\te = e + \"\\n\" + p.LastError\n\t}\n\tp.LastError = e\n}"
}
func (g *CGenerator) EndGroup(gr Group) string {
	switch t := gr.(type) {
	case *cNeedAllGroup:

		for n := t.stack.Back(); len(t.cf.Level()) > 1; n = n.Prev() {
			t.cf.Dec()
			t.cf.Add("} else " + g.UpdateError(n.Value.(string)) + "\n")
		}
		t.cf.Add("if (!accept) {\n\tp->ParserData.Pos = save;\n}\n")
		t.cf.Dec()
		t.cf.Add("}")
		return t.cf.String()
	case *cNeedOneGroup:
		for len(t.cf.Level()) > 1 {
			t.cf.Dec()
			t.cf.Add("}\n")
		}
		t.cf.Add("if (!accept) {\n\tp->ParserData.Pos = save;\n}\n")
		t.cf.Dec()
		t.cf.Add("}")
		return t.cf.String()
	}
	panic(gr)
}

// var inlinere = regexp.MustCompile(`^return ([\s\S]*?)\(p\)$`)
// var inlinere2 = regexp.MustCompile(`^return (\w+\(p, [\s\S]*?\))($|\}\(p\)$)`)

func (g *CGenerator) MakeFunction(value string) string {
	if strings.HasSuffix(value, ")") || strings.HasSuffix(value, "}") {
		return value
	}

	if inlinere.MatchString(value) {
		return inlinere.FindStringSubmatch(value)[1]
	}
	panic("here")
}
func (g *CGenerator) Return(value string) string {
	return "return " + value
}

// var (
// 	callre1 = regexp.MustCompile(`^\s*accept\s`)
// 	callre2 = regexp.MustCompile(`^\s*func\(`)
// )

func (g *CGenerator) Call(value string) string {
	pref := "accept = "
	if callre1.MatchString(value) {
		pref = ""
	}
	if strings.HasSuffix(value, "(p);") {
		return pref + value
	}
	if strings.HasPrefix(value, "p_") || callre2.MatchString(value) {
		return pref + value + "(p);"
	}
	return value
}

func (g *CGenerator) Begin() {
	// imports := ""
	// impList := g.Imports
	// if g.AddDebugLogging {
	// 	impList = append(impList, "log")
	// }
	g.realOutput = `/*
Copyright (c) 2012 Fredrik Ehnbom

This software is provided 'as-is', without any express or implied
warranty. In no event will the authors be held liable for any damages
arising from the use of this software.

Permission is granted to anyone to use this software for any purpose,
including commercial applications, and to alter it and redistribute it
freely, subject to the following __restrict__ions:

   1. The origin of this software must not be misrepresented; you must not
   claim that you wrote the original software. If you use this software
   in a product, an acknowledgment in the product documentation would be
   appreciated but is not required.

   2. Altered source versions must be plainly marked as such, and must not be
   misrepresented as being the original software.

   3. This notice may not be removed or altered from any source
   distribution.
*/
#include <deque>
#include <memory>
#include <string>
#include <vector>
#include <algorithm>
#include <tr1/memory>

using namespace std;
using namespace tr1;
`
	members := g.ParserVariables
	members = append(members, `struct {
		const char* __restrict__ Pos;
		const char* __restrict__ Data;
		const char* __restrict__ Len;
	} ParserData;
`, "RangeClass IgnoreRange;",
		"Node Root;",
		"const char* __restrict__ LastError;")
	g.realOutput += `class ` + g.Name + `;

class Node;
class MyNodeMemberContainer {
public:
	MyNodeMemberContainer() : _size(0), _capacity(0), _content(0) { }
	~MyNodeMemberContainer() { delete[] _content; }
	Node *&operator[](int idx) const { return _content[idx]; }
	void resize(int newsize);

	void pop_back() { _size--; }
	void push_back(Node* n) {
		resize(_size+1);
		_size--;
		 _content[_size] = n;
		 _size++;
	}
	Node *back() { return _content[_size-1]; }
	int size() { return _size; }
	int _size;
	int _capacity;
	Node ** _content;
};

#define NodeMember Node*
//#define NodeMemberContainer vector<NodeMember >
#define NodeMemberContainer MyNodeMemberContainer


#define CreateNode new Node
#define DeleteNode(a) delete a


class RangeClass {
public:
	void Clip(const RangeClass& other)
	{
		if (Start >= other.Start && Start < other.End) {
			Start = other.End;
		}
		if (End >= other.Start && End <= other.End) {
			End = other.Start;
		}
		if (End < Start) {
			End = Start;
		}
	}
	void Reset() { Start = End = 0; }
	const char* __restrict__ Start;
	const char* __restrict__ End;
};
class Node {
public:
	NodeMemberContainer Children;
	const char* __restrict__ Name;
	RangeClass Range;

	void print(string indent="");
	~Node() {
		Reset();
	}
	NodeMember Cleanup(const char* __restrict__ pos, const char* __restrict__ end)
	{
		NodeMember popped = CreateNode;
		popped->Range.Start = pos;
		popped->Range.End = end;
		int popIdx = 0;

		for (int i = Children.size() - 1; i >= 0; i--) {
			if (Children[i]->Range.End <= pos) {
				popIdx = i+1;
				break;
			}
		}
		int num = Children.size()-popIdx;
		if (num > 0) {
			popped->Children.resize(num);
			int j = 0;
			int i = popIdx;
			for (int b = Children.size()-1; b >= popIdx; b--, i++, j++) {
				popped->Children[j] = Children[i];
			}
			Children.resize(popIdx);
		}

		return popped;
	}
	void Discard(const char* __restrict__ pos)
	{
		if (Children.size() == 0)
			return;
		int popIdx = 0;
		for (int i = Children.size()-1; i >= 0; i--) {
			if (Children[i]->Range.End <= pos) {
				popIdx = i+1;
				break;
			}
			DeleteNode(Children[i]);
		}
		Children.resize(popIdx);
	}
	void Reset()
	{
		Discard(0);
	}
	void Append(NodeMember n)
	{
		Children.push_back(n);
	}
	` + g.Name + ` *P;
};

class ` + g.Name + ` {
public:
	bool Parse();
	void SetData(const char* __restrict__ data, unsigned int len);
	NodeMember RootNode();
	void Reset();
	string Data(const char* __restrict__ start, const char* __restrict__ end);
	` + strings.Join(members, "\n\t") + "\n};\n"
	// if g.AddDebugLogging {
	// 	g.output += "var fm parser.CodeFormatter\n\n"
	// }
	g.realOutput += `NodeMember ` + g.Name + `::RootNode() {
	return Root.Children[0];
}

void MyNodeMemberContainer::resize(int newsize) {
	if (newsize > _capacity) {
		if (_capacity != 0) {
			printf("Growing from %d -> %d\n", _capacity, newsize);
		}
		Node** newcontent = new Node*[newsize];
		memcpy(newcontent, _content, _size*sizeof(_content));
		delete[] _content;
		_capacity = _size = newsize;
		_content = newcontent;
	} else {
		_size = newsize;
	}
}

void Node::print(string indent)
{
	if (Children.size()) {
		printf("%s%s - %d %d\n", indent.c_str(), Name, int(Range.Start-P->ParserData.Data), int(Range.End-P->ParserData.Data));
		indent += "  ";
		for (int i = 0; i < Children.size(); i++) {
			Children[i]->print(indent);
		}
	} else {
		printf("%s%s - %d %d - %s\n", indent.c_str(), Name, int(Range.Start-P->ParserData.Data), int(Range.End-P->ParserData.Data), P->Data(Range.Start, Range.End).c_str());
	}
}

void ` + g.Name + `::SetData(const char* __restrict__ data, unsigned int len) {
	ParserData.Data = data;
	ParserData.Len = data + len;
	Reset();
}

void ` + g.Name + `::Reset() {
	ParserData.Pos = ParserData.Data;
	Root.Reset();
	IgnoreRange.Reset();
	LastError = 0;
}

string ` + g.Name + `::Data(const char* __restrict__ start, const char* __restrict__ end) {
	if (this->ParserData.Len == 0) {
		return "";
	}
	if (start < 0) {
		start = 0;
	}
	if (end > this->ParserData.Len) {
		end = this->ParserData.Len;
	}
	if (start > end) {
		return "";
	}
	char *data = new char[end-start+1];
	memcpy(data, start, end-start);
	data[end-start] = '\0';
	string ret(data);
	delete[] data;

	return ret;
}

#include <fstream>
#include <iostream>
#include <sys/time.h>

int main(int argc, char **argv)
{
	fstream f("/Users/quarnster/Library/Application Support/Sublime Text 2/Packages/User/Preferences.sublime-settings");
	char *data = new char[1024*1024];
	f.read(data, 1024*1024);
	JSON p;
	p.SetData(data, f.tellg());
	p.Root.Children.resize(50);
	p.Root.Children.resize(0);
	if (p.Parse()) {
		p.RootNode()->print();
	}

	int N = 10000;

    struct timeval t,t2;
    double perf;
    gettimeofday(&t, NULL);
    for (int i = 0; i < N; i++) {
    	p.Reset();
    	p.Parse();
    }
	p.RootNode()->print();

    gettimeofday(&t2, NULL);
    perf = (t2.tv_sec-t.tv_sec) * 10e9 + (t2.tv_usec - t.tv_usec)*10e3;
    printf("%f, %f\n", perf, perf/N);

	return 0;
}

// func (p *` + g.Name + `) Error() parser.Error {
// 	errstr := ""

// 	line := 1
// 	column := 1
// 	for _, r := range p->ParserData.Data[:p.LastError] {
// 		column ++
// 		if r == '\n' {
// 			line++
// 			column = 1
// 		}
// 	}

// 	if p.LastError == len(p->ParserData.Data) {
// 		errstr = "Unexpected EOF"
// 	} else {
// 		r := p->ParserData.Data[p.LastError]
// 		if r == '\r' || r == '\n' {
// 			errstr = "Unexpected new line"
// 		} else {
// 			errstr = "Unexpected " + string(r)
// 		}
// 	}
// 	return parser.NewError(line, column, errstr)
// }

`
}

func (g *CGenerator) Finish() string {
	ret := g.realOutput + g.output
	if ret[len(ret)-2:] == "\n\n" {
		ret = ret[:len(ret)-1]
	}
	g.realOutput = ""
	g.output = ""
	return ret
}
