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
	"strings"
)

type CPPGenerator struct {
	CGenerator
}

func (g *CPPGenerator) AddNode(data, defName string) string {
	return `accept = true;
const char* __restrict__ start = _pos;
` + g.Call(data) + `
const char* __restrict__ end = _pos;
p->Root.P = p;
if (accept) {
	NodeMember node = p->Root.Cleanup(start, _pos);
	node->Name = "` + defName + `";
	node->P = p;
	node->Range.Clip(p->ignoreRange);
	p->Root.Append(node);
} else {
	p->Root.Discard(start);
}
if (p->ignoreRange.start >= end || p->ignoreRange.end <= start) {
	p->ignoreRange.Reset();
}`
}

func (g *CPPGenerator) MakeParserFunction(node *Node) error {
	id := node.Children[0]
	exp := node.Children[len(node.Children)-1]
	defName := helper(g, id)
	g.currentName = defName
	data := helper(g, exp)

	g.realOutput += "static bool p_" + defName + "(" + g.s.Name + "*);\n"

	if !g.havefunctions {
		g.havefunctions = true
		g.output += "bool " + g.s.Name + "::realParse() {\n\treturn p_" + defName + "(this);\n}\n"
	}

	indenter := CodeFormatter{}
	indenter.Add("static bool p_" + defName + "(" + g.s.Name + "  * __restrict__ p) {\n")
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
		// 		indenter.Add(`int pos = _pos;
		// 	int l   = len(p->parserData.data);

		// log.Println(fm.Level() + "` + defName + ` entered")
		// fm.Inc()
		// ` + data + `
		// fm.Dec()
		// if !accept && _pos != pos {
		// 	log.Fatalln("` + defName + `", accept, ", ", pos, ", ", _pos)
		// }
		// p2 := _pos
		// data := ""
		// if p2 < len(p->parserData.data) {
		// 	data = string(p->parserData.data[pos:p2])
		// }
		// log.Println(fm.Level()+"` + defName + ` returned: ", accept, ", ", pos, ", ", _pos, ", ", l, string(data))
		// return accept
		// `)
	} else {
		if strings.HasPrefix(data, "accept") || data[0] == '{' {
			end := "return accept;\n"
			if data[len(data)-1] != '\n' {
				end = "\n" + end
			}
			indenter.Add("bool accept = false;\nconst register char * __restrict__ &_pos = p->parserData.pos;\n" + data + end)
		} else {
			indenter.Add("return " + data + ";\n")
		}
	}
	indenter.Dec()
	indenter.Add("}\n\n")
	g.output += indenter.String()
	return nil
}

func (g *CPPGenerator) Begin(s GeneratorSettings) error {
	g.s = s
	// imports := ""
	// impList := g.Imports
	// if g.AddDebugLogging {
	// 	impList = append(impList, "log")
	// }
	g.realOutput = g.s.Header + `
#include <deque>
#include <memory>
#include <string>
#include <vector>
#include <algorithm>
#include <memory>

using namespace std;
#define TRUE true
#define FALSE false
`
	members := g.ParserVariables
	members = append(members, `struct {
		const char* __restrict__ pos;
		const char* __restrict__ data;
		const char* __restrict__ end;
	} parserData;
`, "Range ignoreRange;",
		"Node Root;",
		"const char* __restrict__ LastError;")
	g.realOutput += `class ` + g.s.Name + `;

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


class Range {
public:
	void Clip(const Range& other)
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
	void Reset() { start = end = 0; }
	const char* __restrict__ start;
	const char* __restrict__ end;
};
class Node {
public:
	NodeMemberContainer Children;
	const char* __restrict__ Name;
	Range Range;

	void print(string indent="");
	~Node() {
		Reset();
	}
	NodeMember Cleanup(const char* __restrict__ pos, const char* __restrict__ end)
	{
		NodeMember popped = CreateNode;
		popped->Range.start = pos;
		popped->Range.end = end;
		int popIdx = 0;

		for (int i = Children.size() - 1; i >= 0; i--) {
			if (Children[i]->Range.end <= pos) {
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
			if (Children[i]->Range.end <= pos) {
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
	` + g.s.Name + ` *P;
};

class ` + g.s.Name + ` {
public:
	bool Parse(const char* __restrict__ data, unsigned int len);
	NodeMember RootNode();
	string Data(const char* __restrict__ start, const char* __restrict__ end);
	` + strings.Join(members, "\n\t") + `
private:
	bool realParse();
};
`
	// if g.AddDebugLogging {
	// 	g.output += "var fm parser.CodeFormatter\n\n"
	// }
	g.realOutput += `NodeMember ` + g.s.Name + `::RootNode() {
	return &Root;
}

void MyNodeMemberContainer::resize(int newsize) {
	if (newsize > _capacity) {
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
        printf("%s%d-%d: \"%s\"\n", indent.c_str(), (int)(Range.start-P->parserData.data), (int)(Range.end-P->parserData.data), Name);
		indent += "\t";
		for (int i = 0; i < Children.size(); i++) {
			Children[i]->print(indent);
		}
	} else {
         printf("%s%d-%d: \"%s\" - Data: \"%s\"\n", indent.c_str(), (int)(Range.start-P->parserData.data), (int)(Range.end-P->parserData.data), Name, P->Data(Range.start, Range.end).c_str());
	}
}

bool ` + g.s.Name + `::Parse(const char* __restrict__ data, unsigned int len) {
	parserData.data = data;
	parserData.end = data + len;
	parserData.pos = parserData.data;
	Root.Reset();
	ignoreRange.Reset();
	LastError = 0;
	bool ret = realParse();
	Root.Name = "` + g.s.Name + `";
	Root.Range.start = data;
	Root.Range.end = parserData.end;
	return ret;
}

string ` + g.s.Name + `::Data(const char* __restrict__ start, const char* __restrict__ end) {
	if (this->parserData.end == 0) {
		return "";
	}
	if (start < 0) {
		start = 0;
	}
	if (end > this->parserData.end) {
		end = this->parserData.end;
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
}`

	dumptree_s := ""
	if g.s.Debug {
		dumptree_s = "p.RootNode()->print();"
	}

	g.realOutput += strings.Replace(`
#include <fstream>
#include <iostream>
#include <sys/time.h>

int main(int argc, char **argv)
{
    FILE *fp = fopen("`+g.s.Testname+`", "rb");
    if (!fp)
        return -1;
    fseek(fp, 0, SEEK_END);
    int size = ftell(fp);
    char *data = new char[size];
    fseek(fp, 0, SEEK_SET);
    fread(data, 1, size, fp);
    fclose(fp);

	{{ParserName}} p;
	if (p.Parse(data, size)) {
		`+dumptree_s+`
	}`, "{{ParserName}}", g.s.Name, -1)
	if g.s.Bench {
		g.realOutput += `
	int N = 1000;

    struct timeval t,t2;
    double perf;
    gettimeofday(&t, NULL);
    for (int i = 0; i < N; i++) {
    	p.Parse(data, size);
    }

    gettimeofday(&t2, NULL);
    perf = (t2.tv_sec-t.tv_sec) * 10e8 + (t2.tv_usec - t.tv_usec)*10e2;
    printf("Finished in %f seconds, %f ns/op\n", perf/10e8, perf/N);
`
	}
	g.realOutput += `
    delete[] data;
	return 0;
}
`
	return nil
}

func (g *CPPGenerator) Finish() error {
	ret := strings.Replace(g.realOutput+g.output, "{{ParserName}}", g.s.Name, -1)
	if ret[len(ret)-2:] == "\n\n" {
		ret = ret[:len(ret)-1]
	}
	g.realOutput = ""
	g.output = ""

	ln := strings.ToLower(g.s.Name)
	if err := g.s.WriteFile(ln+".cpp", ret); err != nil {
		return err
	}
	return nil
}

func (g *CPPGenerator) TestCommand() []string {
	return []string{"bash", "-c", "c++ -I. -O3 " + strings.ToLower(g.s.Name) + ".cpp -ltcmalloc -o/tmp/a.out && /tmp/a.out"}
}
