/*
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
package parser

// http://en.wikipedia.org/wiki/Parsing_expression_grammar
// http://pdos.csail.mit.edu/papers/parsing:popl04.pdf

import (
	"bytes"
	"container/list"
	"fmt"
	"io/ioutil"
	"strings"
)

const (
	nilrune = '\u0000'
)

type IntStack struct {
	data []int
}

func (i *IntStack) Push(value int) {
	i.data = append(i.data, value)
}

func (i *IntStack) Pop() (ret int) {
	end := len(i.data) - 1
	ret = i.data[end]
	i.data = i.data[:end]
	return
}

type Range struct {
	Start, End int
}

type Node struct {
	Range    Range
	Name     string
	Children list.List
	Data     string
}

func (i *IntStack) Clear() {
	i.data = i.data[:]
}

func (n *Node) String() string {
	s := fmt.Sprintf("%s%d-%d: \"%s\" - Data: \"%s\"\n", indent, n.Range.Start, n.Range.End, n.Name, n.Data)
	indent += "\t"
	for i := n.Children.Front(); i != nil; i = i.Next() {
		s += i.Value.(*Node).String()
	}
	indent = indent[:len(indent)-1]
	return s
}

func (n *Node) Cleanup(pos, end int) *Node {
	var popped Node
	popped.Range = Range{pos, end}
	for i := n.Children.Front(); i != nil; {
		next := i.Next()
		if i.Value.(*Node).Range.Start >= pos {
			popped.Append(n.Children.Remove(i).(*Node))
		}
		i = next
	}
	return &popped
}

func (n *Node) Append(child *Node) {
	n.Children.PushBack(child)
}

func (p *Parser) addNode(add func() bool, name string) bool {
	start := p.Pos()
	shouldAdd := add()
	node := p.currentNode.Cleanup(start, p.Pos())
	node.Name = name
	if shouldAdd {
		end := p.Pos()
		data := make([]byte, end-start)
		p.data.Seek(int64(start), 0)
		if _, err := p.data.Read(data); err == nil {
			node.Data = string(data)
		}
		p.data.Seek(int64(end), 0)
		p.currentNode.Append(node)
	}
	return shouldAdd
}

type Parser struct {
	stack       IntStack
	data        *strings.Reader
	currentNode Node
}

func (p *Parser) Init() {
}

func (p *Parser) RootNode() *Node {
	return p.currentNode.Children.Front().Value.(*Node)
}

func (p *Parser) Pos() int {
	pos := p.data.Len()
	p.data.Seek(0, 0)
	pos = p.data.Len() - pos
	p.data.Seek(int64(pos), 0)
	return pos
}

func (p *Parser) push() {
	p.stack.Push(p.Pos())
}
func (p *Parser) pop() int {
	return p.stack.Pop()
}

func (p *Parser) Accept() bool {
	p.pop()
	return true
}

func (p *Parser) Reject() bool {
	p.data.Seek(int64(p.pop()), 0)
	return false
}

func (p *Parser) Maybe(exp func() bool) bool {
	exp()
	return true
}

func (p *Parser) OneOrMore(exp func() bool) bool {
	p.push()
	if !exp() {
		return p.Reject()
	}
	for exp() {
	}
	return p.Accept()
}

func (p *Parser) NeedAll(exps []func() bool) bool {
	p.push()
	for _, exp := range exps {
		if !exp() {
			return p.Reject()
		}
	}
	return p.Accept()
}

func (p *Parser) NeedOne(exps []func() bool) bool {
	p.push()
	for _, exp := range exps {
		if exp() {
			return p.Accept()
		}
	}
	return p.Reject()
}

func (p *Parser) ZeroOrMore(exp func() bool) bool {
	for exp() {
	}
	return true
}

func (p *Parser) And(exp func() bool) bool {
	p.push()
	ret := exp()
	p.Reject()
	return ret
}

func (p *Parser) Not(exp func() bool) bool {
	return !p.And(exp)
}

func (p *Parser) AnyChar() bool {
	if _, _, err := p.data.ReadRune(); err == nil {
		return true
	}
	return false
}

func (p *Parser) pushNext() (rune, bool) {
	p.push()
	c, _, err := p.data.ReadRune()
	if err != nil {
		return nilrune, false
	}
	return c, true
}

func (p *Parser) InRange(c1, c2 rune) bool {
	if c, ok := p.pushNext(); !ok {
		return p.Reject()
	} else {
		if c >= c1 && c <= c2 {
			return p.Accept()
		}
	}
	return p.Reject()
}

func (p *Parser) InSet(dataset string) bool {
	if c, ok := p.pushNext(); !ok {
		return false
	} else {
		if strings.ContainsRune(dataset, c) {
			return p.Accept()
		}
	}
	return p.Reject()
}
func (p *Parser) Next(n1 string) bool {
	p.push()
	n2 := make([]byte, len(n1))
	if n, err := p.data.Read(n2); err != nil || n != len(n2) || bytes.Compare([]byte(n1), n2) != 0 {
		return p.Reject()
	}
	return p.Accept()
}

var (
	indent = ""
)

type PegParser struct {
	Parser
}

func (p *PegParser) Grammar() bool {
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool { return p.Spacing() },
			func() bool { return p.OneOrMore(func() bool { return p.Definition() }) },
			func() bool { return p.EndOfFile() }})
	}, "Grammar")
}

func (p *PegParser) Definition() bool {
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool { return p.Identifier() },
			func() bool { return p.LEFTARROW() },
			func() bool { return p.Expression() }})
	}, "Definition")
}

func (p *PegParser) Expression() bool {
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool { return p.Sequence() },
			func() bool {
				return p.ZeroOrMore(
					func() bool {
						return p.NeedAll([]func() bool{
							func() bool { return p.SLASH() },
							func() bool { return p.Sequence() }})
					})
			},
		})
	}, "Expression")
}

func (p *PegParser) Sequence() bool {
	return p.addNode(func() bool { return p.ZeroOrMore(func() bool { return p.Prefix() }) }, "Sequence")
}

func (p *PegParser) Prefix() bool {
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.Maybe(
					func() bool {
						return p.NeedOne([]func() bool{
							func() bool { return p.AND() },
							func() bool { return p.NOT() },
						})
					},
				)
			},
			func() bool { return p.Suffix() }})
	}, "Prefix")
}

func (p *PegParser) Suffix() bool {
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool { return p.Primary() },
			func() bool {
				return p.Maybe(func() bool {
					return p.NeedOne([]func() bool{
						func() bool { return p.QUESTION() },
						func() bool { return p.STAR() },
						func() bool { return p.PLUS() }})
				})
			}})
	}, "Suffix")
}

func (p *PegParser) Primary() bool {
	return p.addNode(func() bool {
		return p.NeedOne([]func() bool{
			func() bool {
				return p.NeedAll([]func() bool{
					func() bool { return p.Identifier() },
					func() bool {
						return p.Not(func() bool { return p.LEFTARROW() })
					}})
			},
			func() bool {
				return p.NeedAll([]func() bool{
					func() bool { return p.OPEN() },
					func() bool { return p.Expression() },
					func() bool { return p.CLOSE() }})
			},
			func() bool { return p.Literal() },
			func() bool { return p.Class() },
			func() bool { return p.DOT() }})
	}, "Primary")
}

func (p *PegParser) Identifier() bool {
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool { return p.IdentStart() },
			func() bool { return p.ZeroOrMore(func() bool { return p.IdentCont() }) },
			func() bool { return p.Spacing() }})
	}, "Identifier")
}

func (p *PegParser) IdentStart() bool {
	return p.addNode(func() bool {
		return p.NeedOne([]func() bool{
			func() bool { return p.InRange('a', 'z') },
			func() bool { return p.InRange('A', 'Z') },
			func() bool { return p.InSet("_") }})
	}, "IdentStart")
}

func (p *PegParser) IdentCont() bool {
	return p.addNode(func() bool {
		return p.NeedOne([]func() bool{
			func() bool { return p.IdentStart() },
			func() bool { return p.InRange('0', '9') }})
	}, "IdentCont")
}

func (p *PegParser) Literal() bool {
	return p.addNode(func() bool {
		return p.NeedOne([]func() bool{
			func() bool {
				return p.NeedAll([]func() bool{
					func() bool { return p.InSet("'") },
					func() bool {
						return p.ZeroOrMore(
							func() bool {
								return p.NeedAll([]func() bool{
									func() bool { return p.Not(func() bool { return p.InSet("'") }) },
									func() bool { return p.Char() }})
							})
					},
					func() bool { return p.InSet("'") },
					func() bool { return p.Spacing() }})
			},
			func() bool {
				return p.NeedAll([]func() bool{
					func() bool { return p.InSet("\"") },
					func() bool {
						return p.ZeroOrMore(
							func() bool {
								return p.NeedAll([]func() bool{
									func() bool { return p.Not(func() bool { return p.InSet("\"") }) },
									func() bool { return p.Char() }})
							})
					},
					func() bool { return p.InSet("\"") },
					func() bool { return p.Spacing() }})
			}})
	}, "Literal")
}

func (p *PegParser) Class() bool {
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool { return p.Next("[") },
			func() bool {
				return p.ZeroOrMore(func() bool {
					return p.NeedAll([]func() bool{
						func() bool { return p.Not(func() bool { return p.Next("]") }) },
						func() bool { return p.Range() }})
				})
			},
			func() bool { return p.Next("]") },
			func() bool { return p.Spacing() }})
	}, "Class")
}

func (p *PegParser) Range() bool {
	return p.addNode(func() bool {
		return p.NeedOne([]func() bool{
			func() bool {
				return p.NeedAll([]func() bool{
					func() bool { return p.Char() },
					func() bool { return p.Next("-") },
					func() bool { return p.Char() }})
			},
			func() bool { return p.Char() }})
	}, "Range")
}

func (p *PegParser) Char() bool {
	return p.addNode(
		func() bool {
			return p.NeedOne([]func() bool{
				func() bool {
					return p.NeedAll([]func() bool{
						func() bool { return p.Next("\\") },
						func() bool { return p.InSet("nrt'\"\\[\\]\\\\") }})
				},
				func() bool {
					return p.NeedAll([]func() bool{
						func() bool { return p.Next("\\") },
						func() bool { return p.InRange('0', '2') },
						func() bool { return p.InRange('0', '7') },
						func() bool { return p.InRange('0', '7') }})
				},
				func() bool {
					return p.NeedAll([]func() bool{
						func() bool { return p.Next("\\") },
						func() bool { return p.InRange('0', '7') },
						func() bool { return p.Maybe(func() bool { return p.InRange('0', '7') }) }})
				},
				func() bool {
					return p.NeedAll([]func() bool{
						func() bool { return p.Not(func() bool { return p.Next("\\") }) },
						func() bool { return p.AnyChar() }})
				}})
		}, "Char")
}

func (p *PegParser) LEFTARROW() bool {
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool { return p.Next("<-") },
			func() bool { return p.Spacing() }})
	}, "LEFTARROW")
}

func (p *PegParser) SLASH() bool {
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool { return p.Next("/") },
			func() bool { return p.Spacing() }})
	}, "SLASH")
}

func (p *PegParser) AND() bool {
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool { return p.Next("&") },
			func() bool { return p.Spacing() }})
	}, "AND")
}

func (p *PegParser) NOT() bool {
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool { return p.Next("!") },
			func() bool { return p.Spacing() }})
	}, "NOT")
}

func (p *PegParser) QUESTION() bool {
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool { return p.Next("?") },
			func() bool { return p.Spacing() }})
	}, "QUESTION")
}

func (p *PegParser) STAR() bool {
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool { return p.Next("*") },
			func() bool { return p.Spacing() }})
	}, "STAR")
}

func (p *PegParser) PLUS() bool {
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool { return p.Next("+") },
			func() bool { return p.Spacing() }})
	}, "PLUS")
}

func (p *PegParser) OPEN() bool {
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool { return p.Next("(") },
			func() bool { return p.Spacing() }})
	}, "OPEN")
}

func (p *PegParser) CLOSE() bool {
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool { return p.Next(")") },
			func() bool { return p.Spacing() }})
	}, "CLOSE")
}

func (p *PegParser) DOT() bool {
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool { return p.Next(".") },
			func() bool { return p.Spacing() }})
	}, "DOT")
}

func (p *PegParser) Spacing() bool {
	return p.addNode(func() bool {
		return p.ZeroOrMore(
			func() bool {
				return p.NeedOne([]func() bool{
					func() bool { return p.Space() },
					func() bool { return p.Comment() }})
			})
	}, "Spacing")
}

func (p *PegParser) Comment() bool {
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool { return p.Next("#") },
			func() bool {
				return p.ZeroOrMore(func() bool {
					return p.NeedAll([]func() bool{
						func() bool { return p.Not(func() bool { return p.EndOfLine() }) },
						func() bool { return p.AnyChar() }})
				})
			},
			func() bool { return p.EndOfLine() }})
	}, "Comment")
}

func (p *PegParser) Space() bool {
	return p.addNode(func() bool {
		return p.NeedOne([]func() bool{
			func() bool { return p.Next(" ") },
			func() bool { return p.Next("\t") },
			func() bool { return p.EndOfLine() }})
	}, "Space")
}

func (p *PegParser) EndOfLine() bool {
	return p.addNode(func() bool {
		return p.NeedOne([]func() bool{
			func() bool { return p.Next("\r\n") },
			func() bool { return p.Next("\n") },
			func() bool { return p.Next("\r") }})
	}, "EndOfLine")
}

func (p *PegParser) EndOfFile() bool {
	return p.Not(func() bool { return p.AnyChar() })
}

var (
	convMap    = make(map[string]*Node)
	visitedMap = make(map[string]bool)
	indenter   CodeFormatter
	//	labeler    Labeler
	blocks list.List
)

const (
	labelData       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	enableInlining  = false
	inlineSuffix    = false
	inlineClass     = false
	inlinePrefix    = false
	addDebugLogging = true
	addNodes        = false
)

type Labeler struct {
	currLabel string
}

func (l *Labeler) NewLabel() string {
	ll := len(l.currLabel) - 1
	if len(l.currLabel) == 0 {
		l.currLabel += string(labelData[0])
	} else {
		idx := strings.Index(labelData, string(l.currLabel[ll]))
		if idx == len(labelData)-1 {
			l.currLabel += string(labelData[0])
		} else {
			l.currLabel = l.currLabel[:ll] + string(labelData[idx+1])
		}
	}
	return l.currLabel
}

type CodeFormatter struct {
	level string
	data  string
}

func (i *CodeFormatter) Inc() {
	i.level += "\t"
	l := len(i.data)
	if l > 0 && (i.data[l-1] == '\t' || i.data[l-1] == '\n') {
		i.data += "\t"
	}
}
func (i *CodeFormatter) Dec() {
	i.level = i.level[:len(i.level)-1]
	l := len(i.data)
	if l > 0 && i.data[l-1] == '\t' {
		i.data = i.data[:l-1]
	}
}
func (i *CodeFormatter) Add(add string) {
	i.data += strings.Replace(add, "\n", "\n"+i.level, -1)
}

func (i *CodeFormatter) String() string {
	return i.data
}

const (
	ACTION_REJECT_IF_TRUE = (1 << (iota + 1))
	ACTION_ACCEPT_IF_TRUE
	ACTION_REJECT_IF_FALSE
	ACTION_ACCEPT_IF_FALSE
	ACTION_REJECT_ALWAYS
	ACTION_ACCEPT_ALWAYS
	ACTION_RETURN_VALUE
	ACTION_REJECT_ADD
	ACTION_ACCEPT_ADD
	ACTION_DEFAULT = ACTION_REJECT_IF_FALSE | ACTION_ACCEPT_IF_TRUE
)

type Block struct {
	pcount      int
	formatter   CodeFormatter
	haveComplex bool
	needPush    bool
}

func currentBlock() *Block {
	return blocks.Back().Value.(*Block)
}
func addLine(line string) {
	currentBlock().formatter.Add(line)
}

func addPush() {
	currentBlock().pcount++
	addLine("p.push()\n")
}
func addPop() {
	currentBlock().pcount--
}
func addAccept() {
	addPop()
	currentBlock().needPush = true
	//addLine("true\n")
	addLine("p.Accept()\n")
}
func addReject() {
	addPop()
	currentBlock().needPush = true
	//addLine("false\n")
	addLine("p.Reject()\n")
}

func startBlock(name string) {
	blocks.PushBack(&Block{})
	//	addLine("/* " + name + " */\n{\n")
	//	currentBlock().formatter.Inc()
	//addPush()
}

func endBlock(endtest string, actions int) string {
	// for actions == 0 && currentBlock().pcount > 0 {
	// 	addReject()
	// }
	addReturn(endtest, actions)
	//	currentBlock().formatter.Dec()
	//	addLine("}\n")
	back := blocks.Back()
	blocks.Remove(back)
	bb := back.Value.(*Block)
	ret := bb.formatter.String()
	if bb.needPush {
		ret = makeComplex("p.push()\n"+ret) + "()"
	}
	return ret
}

func addReturn(value string, actions int) {
	if actions&(ACTION_REJECT_IF_TRUE|ACTION_ACCEPT_IF_TRUE) != 0 {
		addLine("if " + value + " {\n\treturn ")
		if actions&ACTION_REJECT_IF_TRUE != 0 {
			addReject()
		} else {
			addAccept()
		}
		if actions&(ACTION_REJECT_IF_FALSE|ACTION_ACCEPT_IF_FALSE) != 0 {
			addLine("}\n\treturn ")
			if actions&ACTION_REJECT_IF_FALSE != 0 {
				addReject()
			} else {
				addAccept()
			}
			addLine("\n")
		} else {
			addLine("}\n")
		}
	} else if actions&(ACTION_REJECT_IF_FALSE|ACTION_ACCEPT_IF_FALSE) != 0 {
		addLine("if !(" + value + ") {\n\treturn ")
		if actions&ACTION_REJECT_IF_FALSE != 0 {
			addReject()
		} else {
			addAccept()
		}
		addLine("}\n")
	} else {
		if actions&ACTION_REJECT_ADD != 0 {
			addReject()
		} else if actions&ACTION_ACCEPT_ADD != 0 {
			addAccept()
		}
		if actions&ACTION_REJECT_ALWAYS != 0 {
			addLine(value + "\n")
			addLine("return ")
			addReject()
		} else if actions&ACTION_ACCEPT_ALWAYS != 0 {
			addLine(value + "\n")
			addLine("return ")
			addAccept()
		} else if len(value) > 0 {
			addLine("return " + value + "\n")
		}
	}
}

func makeReturn(value string) string {
	return "return " + value
}

func makeCall(value string) string {
	return value + "()"
}

func makeReturnCall(value string) string {
	return makeCall(makeReturn(value))
}

func makeComplex(value string) string {
	f := CodeFormatter{}
	f.Add("func() bool {\n")
	f.Inc()
	f.Add(value)
	f.Dec()
	f.Add("\n}")
	return f.String()
}

func makeComplexReturn(value string) string {
	return makeComplex(makeReturn(value))
}

func makeComplexReturnCall(value string) string {
	return makeCall(makeComplexReturn(value))
}

func addComplexReturn(value string, actions int) {
	// if !currentBlock().haveComplex {
	// 	currentBlock().haveComplex = true
	// 	addLine("var ret bool\n")
	// }
	addReturn(makeComplex(value)+"()", actions)
}

func (p *PegParser) helper(node *Node) (retstring string) {
	f := CodeFormatter{}
	// startBlock("tmp")
	// defer endBlock("", 0)
	switch node.Name {
	case "Class":
		if inlineClass {
			startBlock("Class: " + strings.TrimSpace(node.Data))

			addLine("c, _, err := p.data.ReadRune()\n")
			addReturn("err == nil", ACTION_REJECT_IF_FALSE)
			others := ""
			for n := node.Children.Front(); n != nil; n = n.Next() {
				child := n.Value.(*Node)
				if child.Name == "Range" {
					if child.Children.Len() == 2 {
						addReturn("c >= '"+child.Children.Front().Value.(*Node).Data+"' && c <= '"+child.Children.Back().Value.(*Node).Data+"'", ACTION_ACCEPT_IF_TRUE)
					} else {
						others += child.Data
					}
				}
			}
			if others != "" {
				addReturn("strings.ContainsRune(`"+others+"`, c)", ACTION_ACCEPT_IF_TRUE)
			}
			return endBlock("", ACTION_REJECT_ALWAYS)
		} else {
			f.Add("p.NeedOne([]func() bool{\n")
			f.Inc()
			others := ""
			for n := node.Children.Front(); n != nil; n = n.Next() {
				child := n.Value.(*Node)
				if child.Name == "Range" {
					if child.Children.Len() == 2 {
						f.Add(makeComplexReturn("p.InRange('"+child.Children.Front().Value.(*Node).Data+"', '"+child.Children.Back().Value.(*Node).Data+"')") + ",\n")
					} else {
						others += child.Data
					}
				}
			}
			if others != "" {
				f.Add(makeComplexReturn("p.InSet(`"+others+"`)") + ",\n")
			}
			f.Dec()
			f.Add("})")
			return f.String()

		}
	case "DOT":
		return "p.AnyChar()"
	case "Identifier":
		back := node.Children.Back().Value.(*Node)
		data := node.Data
		if back.Name == "Spacing" {
			data = strings.Replace(data, back.Data, "", -1)
		}
		return data
	case "Literal":
		data := strings.TrimSpace(node.Data)
		if data[0] == '\'' && data[len(data)-1] == '\'' {
			data = "\"" + data[1:len(data)-1] + "\""
		}
		return "p.Next(" + data + ")"
	case "Expression":
		if node.Children.Len() == 1 {
			return p.helper(node.Children.Front().Value.(*Node))
		} else {
			//startBlock("need one: " + strings.TrimSpace(node.Data))
			f.Add("p.NeedOne([]func() bool{\n")
			f.Inc()
			i := 0
			for n := node.Children.Front(); n != nil; n = n.Next() {
				i++
				child := n.Value.(*Node)
				if i&1 == 0 {
					continue
				}
				f.Add(makeComplexReturn(p.helper(child)) + ",\n")
			}
			f.Dec()
			f.Add("})")

			return f.String()
		}
		return
	case "Sequence":
		if node.Children.Len() == 1 {
			return p.helper(node.Children.Front().Value.(*Node))
		} else {
			f.Add("p.NeedAll([]func() bool{\n")
			f.Inc()
			for n := node.Children.Front(); n != nil; n = n.Next() {
				f.Add(makeComplexReturn(p.helper(n.Value.(*Node))) + ",\n")
			}
			f.Dec()
			f.Add("})")

			return f.String() //endBlock(data, 0)
		}
		return
	case "Prefix":
		front := node.Children.Front().Value.(*Node)
		if node.Children.Len() == 1 {
			return p.helper(front)
		} else {
			exp := p.helper(node.Children.Back().Value.(*Node))
			if inlinePrefix {
				startBlock("prefix: " + strings.TrimSpace(node.Data))
				addLine("ret := " + makeComplexReturnCall(exp) + "\n")
				switch front.Name {
				case "NOT":
					return endBlock("!ret", ACTION_REJECT_ADD)
				case "AND":
					return endBlock("ret", ACTION_REJECT_ADD)
				}
			} else {
				switch front.Name {
				case "NOT":
					return "p.Not(" + makeComplexReturn(exp) + ")"
				case "AND":
					return "p.And(" + makeComplexReturn(exp) + ")"
				}
			}
		}
		panic("Shouldn't reach this: " + front.Name)
	case "Suffix":
		if node.Children.Len() == 1 {
			return p.helper(node.Children.Front().Value.(*Node))
		} else {
			back := node.Children.Back().Value.(*Node)

			if !inlineSuffix {
				exp := makeComplexReturn(p.helper(node.Children.Front().Value.(*Node)))

				switch back.Name {
				case "PLUS":
					f.Add("p.OneOrMore(")
					//					f.Inc()
					f.Add(exp + ")")
				case "STAR":
					f.Add("p.ZeroOrMore(")
					//					f.Inc()
					f.Add(exp + ")")
				case "QUESTION":
					f.Add("p.Maybe(")
					//					f.Inc()
					f.Add(exp + ")")
				}
				return f.String()
			} else {
				startBlock("Suffix: " + strings.TrimSpace(node.Data))
				exp := makeComplexReturnCall(p.helper(node.Children.Front().Value.(*Node)))

				switch back.Name {
				case "PLUS":
					addLine("/* + */\n")
					addReturn(exp, ACTION_REJECT_IF_FALSE)
					fallthrough
				case "STAR":
					addLine("/* + / * */\n")
					addLine("for " + exp + " {\n")
					addLine("}\n")
				case "QUESTION":
					addLine(exp + "\n")
				}
				addLine("/* ? / * / + */\n")
				return endBlock("", ACTION_ACCEPT_ALWAYS)
			}
		}
		return
	case "Primary":
		front := node.Children.Front().Value.(*Node)

		if front.Name == "Identifier" {
			// Inline opportunity
			fd := strings.TrimSpace(front.Data)
			r, ok := convMap[fd]
			alreadyInlined := visitedMap[fd]
			if ok && enableInlining && !alreadyInlined {
				// just to prevent an inline from inlining itself
				visitedMap[fd] = true
				data := p.helper(r)
				visitedMap[fd] = false
				return data
			} else {
				//startBlock("Literal: " + front.Data)
				return "p." + p.helper(front) + "()"
			}
		} else if front.Name == "OPEN" {
			return p.helper(node.Children.Front().Next().Value.(*Node))
		} else if front.Name == "Literal" {
			//startBlock("Literal: " + front.Data)
			return p.helper(front)
		}
	case "Spacing", "Space":
		// ignore
	default:
		return "\n\n-----------------------------------------------------\n" + node.Name + ", " + node.Data + "-----------------------------------------------------\n"
	}
	//	indenter.Inc()
	for n := node.Children.Front(); n != nil; n = n.Next() {
		retstring += p.helper(n.Value.(*Node))
	}
	return
	//	indenter.Dec()
}

func (p *PegParser) Dump() {
	rootNode := p.RootNode()
	if enableInlining {
		for n := rootNode.Children.Back(); n != nil; n = n.Prev() {
			node := n.Value.(*Node)
			if node.Name == "Definition" {
				id := node.Children.Front().Value.(*Node)
				exp := node.Children.Back().Value.(*Node)

				defName := p.helper(id)
				convMap[defName] = exp
			}
		}
	}
	imports := ""
	if inlineClass {
		imports += "\t\"strings\"\n"
	}
	if addDebugLogging {
		imports += "\t\"log\"\n"
	}
	output := fmt.Sprintln(`package parser

import (
` + imports + `)

type PegParser2 struct {
	Parser
}
`)
	if addDebugLogging {
		output += "var fm CodeFormatter\n\n"
	}

	for n := rootNode.Children.Front(); n != nil; n = n.Next() {
		node := n.Value.(*Node)
		if node.Name == "Definition" {
			id := node.Children.Front().Value.(*Node)
			exp := node.Children.Back().Value.(*Node)
			data := p.helper(exp)
			defName := p.helper(id)
			//			fmt.Println(node)

			indenter = CodeFormatter{}
			indenter.Add("func (p *PegParser2) " + defName + "() bool {\n")
			indenter.Inc()
			comment := "/* " + strings.Replace(strings.TrimSpace(node.Data), "\n", "\n * ", -1)
			indenter.Add(comment)
			if strings.ContainsRune(comment, '\n') {
				indenter.Add("\n")
			}
			indenter.Add(" */\n")

			if addDebugLogging {
				indenter.Add(`var (
	pos = p.Pos()
	l   = p.data.Len()
)
`)
				indenter.Add(`log.Println(fm.level + "` + defName + " entered\")\n")
				indenter.Add("fm.Inc()\n")
			}
			if addNodes {
				data = "p.addNode(" + makeComplexReturn(data) + ", \"" + defName + "\")"
			}
			if addDebugLogging {
				indenter.Add("res := " + data)
				indenter.Add("\nfm.Dec()\n")
				indenter.Add(`if !res && p.Pos() != pos {
	log.Fatalln("` + defName + `", res, ", ", pos, ", ", p.Pos())
}
p2 := p.Pos()
data := make([]byte, p2-pos)
p.data.Seek(int64(pos), 0)
p.data.Read(data)
p.data.Seek(int64(p2), 0)
`)
				indenter.Add("log.Println(fm.level+\"" + defName + ` returned: ", res, ", ", pos, ", ", p.Pos(), ", ", l, string(data))` + "\n")
				indenter.Add(makeReturn("res\n"))
			} else {
				indenter.Add(makeReturn(data + "\n"))
			}
			indenter.Dec()
			indenter.Add("}\n")
			output += indenter.String()
		}
	}
	ioutil.WriteFile("./parser2.go", []byte(output), 0644)
}
