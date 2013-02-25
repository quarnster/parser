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

type CGenerator struct {
	s               GeneratorSettings
	output          string
	realOutput      string
	AddDebugLogging bool
	CustomActions   []CustomAction
	ParserVariables []string
	Imports         []string
	havefunctions   bool
	currentName     string
}

func (g *CGenerator) SetCustomActions(actions []CustomAction) {
	g.CustomActions = actions
}

func (g *CGenerator) AddNode(data, defName string) string {
	return `accept = TRUE;
const char* __restrict__ start = p->parserData.pos;
` + g.Call(data) + `
const char* __restrict__ end = p->parserData.pos;
if (accept) {
	Node* node = Node_cleanup(&p->_root, start, p->parserData.pos);
	node->name = "` + defName + `";
	Range_clip(&node->range, &p->ignoreRange);
	Node_pushback(&p->_root, node);
} else {
	Node_discard(&p->_root, start);
}
if (p->ignoreRange.start >= end || p->ignoreRange.end <= start) {
	Range_reset(&p->ignoreRange);
}`
}

func (g *CGenerator) Ignore(data string) string {
	return `accept = TRUE;
const char* __restrict__ start = p->parserData.pos;
` + g.Call(data) + `
if (accept && start != p->parserData.pos) {
	if (start < p->ignoreRange.start || p->ignoreRange.start == 0) {
		p->ignoreRange.start = start;
	}
	p->ignoreRange.end = p->parserData.pos;
}
`
}
func (g *CGenerator) MakeParserFunction(node *Node) error {
	id := node.Children[0]
	exp := node.Children[len(node.Children)-1]
	defName := helper(g, id)
	g.currentName = defName
	data := helper(g, exp)

	g.realOutput += "static int p_" + defName + "(" + g.s.Name + "*);\n"

	if !g.havefunctions {
		g.havefunctions = true
		g.output += "static int " + g.s.Name + "_parse2(" + g.s.Name + "* __restrict__ p) {\n\treturn p_" + defName + "(p);\n}\n"
	}

	indenter := CodeFormatter{}
	indenter.Add("static int p_" + defName + "(" + g.s.Name + "  * __restrict__ p) {\n")
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
		// 			data = "bool accept = FALSE;\n" + data
		// 		} else {
		// 			data = "bool accept = " + data + ";"
		// 		}
		// 		indenter.Add(`int pos = p->parserData.pos;
		// 	int l   = len(p->parserData.data);

		// log.Println(fm.Level() + "` + defName + ` entered")
		// fm.Inc()
		// ` + data + `
		// fm.Dec()
		// if !accept && p->parserData.pos != pos {
		// 	log.Fatalln("` + defName + `", accept, ", ", pos, ", ", p->parserData.pos)
		// }
		// p2 := p->parserData.pos
		// data := ""
		// if p2 < len(p->parserData.data) {
		// 	data = string(p->parserData.data[pos:p2])
		// }
		// log.Println(fm.Level()+"` + defName + ` returned: ", accept, ", ", pos, ", ", p->parserData.pos, ", ", l, string(data))
		// return accept
		// `)
	} else {
		if strings.HasPrefix(data, "accept") || data[0] == '{' {
			end := "return accept;\n"
			if data[len(data)-1] != '\n' {
				end = "\n" + end
			}
			indenter.Add("int accept = FALSE;\n" + data + end)
		} else {
			indenter.Add("return " + data + ";\n")
		}
	}
	indenter.Dec()
	indenter.Add("}\n\n")
	g.output += indenter.String()
	return nil
}

func (g *CGenerator) MakeParserCall(value string) string {
	return "p_" + value
}

func (g *CGenerator) CheckInRange(a, b string) string {
	return `if (p->parserData.pos >= p->parserData.end) {
	accept = FALSE;
} else {
	const char c = *p->parserData.pos;
	if (c >= '` + a + `' && c <= '` + b + `') {
		p->parserData.pos++;
		accept = TRUE;
	} else {
		accept = FALSE;
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
	accept = FALSE;
	if (p->parserData.pos < p->parserData.end) {
		const char c =  *p->parserData.pos;
		if (` + tests + `) {
			p->parserData.pos++;
			accept = TRUE;
		}
	}
}`
}

func (g *CGenerator) CheckAnyChar() string {
	return `if (p->parserData.pos >= p->parserData.end) {
	accept = FALSE;
} else {
	p->parserData.pos++;
	accept = TRUE;
}`
}

func (g *CGenerator) CheckNext(a string) string {
	/*
	 */

	if a[0] == '\'' {
		return `if (p->parserData.pos >= p->parserData.end || *p->parserData.pos != ` + a + `) {
	accept = FALSE;
} else {
	p->parserData.pos++;
	accept = TRUE;
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
		tests += fmt.Sprintf("*(p->parserData.pos + %d) != '%s'", pos, c2)
	}
	return fmt.Sprintf(`{
	accept = TRUE;
	const char* __restrict__ s = p->parserData.pos;
	const char* __restrict__ e = s + %d;
	if (e > p->parserData.end) {
		accept = FALSE;
	} else {
		if (%s) {
			accept = FALSE;
		}
	}
	if (accept) {
		p->parserData.pos += %d;
	}
}`, pos, tests, pos)
}

func (g *CGenerator) AssertNot(a string) string {
	return `const char* __restrict__ s = p->parserData.pos;
` + g.Call(a) + `
p->parserData.pos = s;
accept = !accept;`
}

func (g *CGenerator) AssertAnd(a string) string {
	return `const char* __restrict__ s = p->parserData.pos;
` + g.Call(a) + `
p->parserData.pos = s;`
}

func (g *CGenerator) ZeroOrMore(a string) string {
	var cf CodeFormatter
	cf.Add("{\n")
	cf.Inc()
	cf.Add("accept = TRUE;")
	cf.Add("\nwhile (accept) {\n")
	cf.Inc()
	cf.Add(g.Call(a))
	cf.Dec()
	cf.Add("\n}\n")
	cf.Add("accept = TRUE;\n")
	cf.Dec()
	cf.Add("}")
	return cf.String()
}

func (g *CGenerator) OneOrMore(a string) string {
	var cf CodeFormatter
	cf.Add("{\n")
	cf.Inc()
	cf.Add(`const char* __restrict__ save = p->parserData.pos;
` + g.Call(a) + `
if (!accept) {
	p->parserData.pos = save;
} else {
	while (accept) {
`)
	cf.Inc()
	cf.Inc()
	cf.Add(g.Call(a) + "\n")
	cf.Dec()
	cf.Add(`}
accept = TRUE;
`)
	cf.Dec()
	cf.Add("}\n")
	cf.Dec()
	cf.Add("}")
	return cf.String()
}

func (g *CGenerator) Maybe(a string) string {
	return g.Call(a) + "\naccept = TRUE;"
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
	const char* __restrict__ save = p->parserData.pos;
`)
		r.cf.Inc()
		return &r
	}
	r := cNeedOneGroup{g: g}
	r.cf.Add(`{
	const char* __restrict__ save = p->parserData.pos;
`)
	r.cf.Inc()
	return &r
}
func (g *CGenerator) UpdateError(msg string) string {
	return "{ /* if (p->LastError < p->parserData.pos) { p->LastError = p->parserData.pos; */ }"
	// return "{\n\te := fmt.Sprintf(`Expected " + msg + " near %d`, p->parserData.pos)\n\tif len(p.LastError) != 0 {\n\t\te = e + \"\\n\" + p.LastError\n\t}\n\tp.LastError = e\n}"
}
func (g *CGenerator) EndGroup(gr Group) string {
	switch t := gr.(type) {
	case *cNeedAllGroup:

		for n := t.stack.Back(); len(t.cf.Level()) > 1; n = n.Prev() {
			t.cf.Dec()
			t.cf.Add("}\n")
		}
		t.cf.Add("if (!accept) {\n")
		t.cf.Inc()
		t.cf.Add(g.UpdateError("TODO") + "\np->parserData.pos = save;\n")
		t.cf.Dec()
		t.cf.Add("}\n")
		t.cf.Dec()
		t.cf.Add("}")
		return t.cf.String()
	case *cNeedOneGroup:
		for len(t.cf.Level()) > 1 {
			t.cf.Dec()
			t.cf.Add("}\n")
		}
		t.cf.Add("if (!accept) {\n\tp->parserData.pos = save;\n}\n")
		t.cf.Dec()
		t.cf.Add("}")
		return t.cf.String()
	}
	panic(gr)
}

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

func (g *CGenerator) Begin(s GeneratorSettings) error {
	g.s = s
	// imports := ""
	// impList := g.Imports
	// if g.AddDebugLogging {
	// 	impList = append(impList, "log")
	// }
	dumptree_s := ""
	if g.s.Debug {
		dumptree_s = "Node_print(&p, &p._root, \"\");"
	}
	members := g.ParserVariables
	members = append(members, `struct {
		const char* __restrict__ Pos;
		const char* __restrict__ Data;
		const char* __restrict__ Len;
	} parserData;
`, "RangeClass ignoreRange;",
		"Node _root;",
		"const char* __restrict__ LastError;")
	g.realOutput += g.s.Header + `
#include <stdio.h>
#include <stdlib.h>
#include <memory.h>
#include <assert.h>

typedef struct {
    const char* __restrict__ start;
    const char* __restrict__ end;
} Range;

#define FALSE 0
#define TRUE 1

static void Range_clip(Range* self, const Range* other)
{
    if (self->start >= other->start && self->start < other->end) {
        self->start = other->end;
    }
    if (self->end >= other->start && self->end <= other->end) {
        self->end = other->start;
    }
    if (self->end < self->start) {
        self->end = self->start;
    }
}
static void Range_reset(Range* r) { r->start = r->end = 0; }

typedef struct _Node {
    struct _Node** children;
    int size;
    int capacity;

    const char* __restrict__ name;
    Range range;
} Node;

static void Node_init(Node *n);
static void Node_free(Node *n);

#define CreateNode(a) a = malloc(sizeof(Node)); Node_init(a)
#define DeleteNode(a) Node_free(a); free(a)

static Node* Node_cleanup(Node *n, const char* __restrict__ pos, const char* __restrict__ end)
{
    Node* popped;
    CreateNode(popped);
    popped->range.start = pos;
    popped->range.end = end;
    int popIdx = 0;
    int i;

    for (i = n->size - 1; i >= 0; i--) {
        if (n->children[i]->range.end <= pos) {
            popIdx = i+1;
            break;
        }
    }
    int num = n->size-popIdx;
    if (num > 0) {
        popped->children = malloc(num * sizeof(Node*));
        popped->size = popped->capacity = num;
        int j = 0;
        int i = popIdx;
        int b;
        for (b = n->size-1; b >= popIdx; b--, i++, j++) {
            popped->children[j] = n->children[i];
        }
        n->size = popIdx;
    }

    return popped;
}
static void Node_discard(Node *n, const char* __restrict__ pos)
{
    if (n->size == 0)
        return;
    int popIdx = 0;
    int i;
    for (i = n->size-1; i >= 0; i--) {
        if (n->children[i]->range.end <= pos) {
            popIdx = i+1;
            break;
        }
        DeleteNode(n->children[i]);
    }
    n->size = popIdx;
}
static void Node_reset(Node* n)
{
    Node_discard(n, 0);
}

static void Node_init(Node *n) { memset(n, 0, sizeof(Node)); }
static void Node_free(Node *n) {
    Node_reset(n);
    free(n->children);
}

static void Node_popback(Node* n) { n->size--; }
static void Node_pushback(Node *n, Node* child) {
    if (n->size+1 > n->capacity) {
        Node **c2 = malloc((n->size+1)*sizeof(Node*));
        n->capacity = n->size+1;
        memcpy(c2, n->children, n->size*sizeof(Node*));
        free(n->children);
        n->children = c2;
    }
    n->children[n->size] = child;
    n->size++;
}

typedef struct {
    struct {
        const char * __restrict__ data;
        const char * __restrict__ end;
        const char * __restrict__ pos;
    } parserData;
    Node _root;
    Range ignoreRange;
} {{ParserName}};

Node* {{ParserName}}_rootNode({{ParserName}}* p) {
    return &p->_root;
}
`

	g.output += `
static int {{ParserName}}_parse2({{ParserName}}* p);
int {{ParserName}}_parse({{ParserName}}* p, const char* __restrict__ data, unsigned int len)
{
    Node_reset(&p->_root);
    p->_root.name = "` + g.s.Name + `";
    Range_reset(&p->ignoreRange);
    p->parserData.data = data;
    p->parserData.end = data + len;
    p->parserData.pos = data;
    int ret = {{ParserName}}_parse2(p);
    p->_root.range.start = p->_root.range.end = data;
    if (p->_root.size) {
        p->_root.range.end = p->_root.children[p->_root.size-1]->range.end;
    }
    return ret;
}
static char databuf[1024];
static const char * {{ParserName}}_data({{ParserName}} *p, const char *start, const char *end) {
    if (start < p->parserData.data) {
        start = p->parserData.data;
    }
    if (end > p->parserData.end) {
        end = p->parserData.end;
    }
    if (start > end) {
        return "";
    }

    int len = end-start;
    strncpy(databuf, start, len);
    databuf[len] = '\0';
    return databuf;
}

static void Node_print({{ParserName}} *p, Node *n, const char* indent) {
    if (n->size) {
        int i;
        printf("%s%d-%d: \"%s\"\n", indent, (int)(n->range.start-p->parserData.data), (int)(n->range.end-p->parserData.data), n->name);
        char myindent[20] = "\t";
        strcat(myindent, indent);
        for (i = 0; i < n->size; i++) {
            Node_print(p, n->children[i], myindent);
        }
    } else {
         printf("%s%d-%d: \"%s\" - Data: \"%s\"\n", indent, (int)(n->range.start-p->parserData.data), (int)(n->range.end-p->parserData.data), n->name, {{ParserName}}_data(p, n->range.start, n->range.end));
    }
}

#include <sys/time.h>

int main(int argc, char **argv)
{
    FILE *fp = fopen("` + g.s.Testname + `", "rb");
    if (!fp)
        return -1;
    fseek(fp, 0, SEEK_END);
    int size = ftell(fp);
    char *data = malloc(size);
    fseek(fp, 0, SEEK_SET);
    fread(data, 1, size, fp);
    fclose(fp);

    {{ParserName}} p;
    Node_init(&p._root);
    if ({{ParserName}}_parse(&p, data, size)) {
        ` + dumptree_s + `
    } else {
        printf("Failed to parse\n");
        exit(1);
    }`

	if g.s.Bench {
		g.output += `
    int N = 1000;

    struct timeval t,t2;
    double perf;
    gettimeofday(&t, NULL);
    int i;
    for (i = 0; i < N; i++) {
	    {{ParserName}}_parse(&p, data, size);
    }

    gettimeofday(&t2, NULL);
    perf = (t2.tv_sec-t.tv_sec) * 10e8 + (t2.tv_usec - t.tv_usec)*10e2;
    printf("Finished in %f seconds, %f ns/op\n", perf/10e8, perf/N);

`
	}
	g.output += `
    free(data);
    Node_free(&p._root);
    return 0;
}
`
	return nil
}

func (g *CGenerator) Finish() error {
	ret := strings.Replace(g.realOutput+g.output, "{{ParserName}}", g.s.Name, -1)
	if ret[len(ret)-2:] == "\n\n" {
		ret = ret[:len(ret)-1]
	}
	g.realOutput = ""
	g.output = ""
	ln := strings.ToLower(g.s.Name)

	if err := g.s.WriteFile(ln+".c", ret); err != nil {
		return err
	}
	return nil
}

func (g *CGenerator) TestCommand() []string {
	return []string{"bash", "-c", "cc -I. -O3 ./" + strings.ToLower(g.s.Name) + ".c -ltcmalloc -o/tmp/a.out && /tmp/a.out"}
}
