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

type Range struct {
	Start, End int
}
type Node struct {
	Range    Range
	Name     string
	Children list.List
	Data     string
}

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

func (i *IntStack) Clear() {
	i.data = i.data[:]
}

type Parser struct {
	stack IntStack
	data  *strings.Reader
}

func (p *Parser) Init() {
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
		return false
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

type PegParser struct {
	Parser
	currentNode Node
}

func (p *PegParser) addNode(add func() bool, name string) bool {
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
	labeler    Labeler
)

const (
	labelData      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	enableInlining = true
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
	//	for _, line := range strings.Split(add, "\n") {
	i.data += strings.Replace(add, "\n", "\n"+i.level, -1)
	//	}
}

func (i *CodeFormatter) String() string {
	return strings.Replace(i.data, "\t", "    ", -1)
}

func (p *PegParser) helper(node *Node) {
	switch node.Name {
	case "Class":
		indenter.Add("/* Class: " + strings.TrimSpace(node.Data) + " */\n{\n")
		indenter.Inc()
		indenter.Add("pos := p.Pos()\n")

		label := labeler.NewLabel()
		indenter.Add("c, _, err := p.data.ReadRune()\naccept = err == nil\nif !accept { goto " + label + " }\n")
		others := ""
		for n := node.Children.Front(); n != nil; n = n.Next() {
			child := n.Value.(*Node)
			if child.Name == "Range" {
				if child.Children.Len() == 2 {
					indenter.Add("accept = c >= '" + child.Children.Front().Value.(*Node).Data + "' && c <= '" + child.Children.Back().Value.(*Node).Data + "'\n")
					indenter.Add("if accept { goto " + label + " }\n")
				} else {
					others += child.Data
				}
			}
		}
		if others != "" {
			indenter.Add("accept = strings.ContainsRune(`" + others + "`, c)\n")
		}
		indenter.Add(label + ":\n")
		indenter.Add("if !accept { p.data.Seek(int64(pos), 0) }\n")
		indenter.Dec()
		indenter.Add("}\n")
		return
	case "DOT":
		indenter.Add("/* DOT */ { _, _, err := p.data.ReadRune(); accept = err == nil }\n")
		return
	case "Identifier":
		back := node.Children.Back().Value.(*Node)
		data := node.Data
		if back.Name == "Spacing" {
			data = strings.Replace(data, back.Data, "", -1)
		}
		indenter.Add(data)
		return
	case "Literal":
		data := strings.TrimSpace(node.Data)
		if data[0] == '\'' && data[len(data)-1] == '\'' {
			data = "\"" + data[1:len(data)-1] + "\""
		}
		indenter.Add("accept = p.Next(" + data + ")\n")
		return
	case "Expression":
		if node.Children.Len() == 1 {
			p.helper(node.Children.Front().Value.(*Node))
		} else {
			label := labeler.NewLabel()
			indenter.Add("/*need one: " + strings.TrimSpace(node.Data) + " */\n{\n")
			indenter.Inc()
			indenter.Add("pos := p.Pos()\n")
			for n := node.Children.Front(); n != nil; n = n.Next() {
				child := n.Value.(*Node)
				if child.Name == "SLASH" {
					continue
				}
				p.helper(child)
				indenter.Add("if (accept) { goto " + label + " } else { p.data.Seek(int64(pos), 0) }\n")
			}
			indenter.Add(label + ":\n")
			indenter.Dec()
			indenter.Add("}\n")
		}
		return
	case "Sequence":
		if node.Children.Len() == 1 {
			p.helper(node.Children.Front().Value.(*Node))
		} else {
			label := labeler.NewLabel()
			indenter.Add("/* need all: " + strings.TrimSpace(node.Data) + " */\n{\n")
			indenter.Inc()
			indenter.Add("pos := p.Pos()\n")
			for n := node.Children.Front(); n != nil; n = n.Next() {
				p.helper(n.Value.(*Node))
				indenter.Add("if (!accept) { p.data.Seek(int64(pos), 0); goto " + label + " }\n")
			}
			indenter.Add(label + ":\n")
			indenter.Dec()
			indenter.Add("}\n")
		}
		return
	case "Prefix":
		front := node.Children.Front().Value.(*Node)
		if node.Children.Len() == 1 {
			p.helper(front)
		} else {
			indenter.Add("/* prefix: " + strings.TrimSpace(node.Data) + " */\n{\n")
			indenter.Inc()
			indenter.Add("pos := p.Pos()\n")
			p.helper(node.Children.Back().Value.(*Node))
			indenter.Add("p.data.Seek(int64(pos), 0)\n")
			switch front.Name {
			case "NOT":
				indenter.Add("/* NOT */\naccept = !accept\n")
			case "AND":
				// Don't need to do anything for and
			default:
				indenter.Add("NOT_IMPLEMENTED_" + front.Name)
			}
			indenter.Dec()
			indenter.Add("}\n")
		}
		return
	case "Suffix":
		if node.Children.Len() == 1 {
			p.helper(node.Children.Front().Value.(*Node))
		} else {
			back := node.Children.Back().Value.(*Node)
			indenter.Add("/* Suffix: " + strings.TrimSpace(node.Data) + " */\n{\n")
			indenter.Inc()
			label := labeler.NewLabel()
			switch back.Name {
			case "PLUS":
				indenter.Add("pos := p.Pos()\n")
				indenter.Add("/* + */\n")
				p.helper(node.Children.Front().Value.(*Node))
				indenter.Add("if (!accept) { p.data.Seek(int64(pos), 0); goto " + label + " }\n\n")
				fallthrough
			case "STAR":
				indenter.Add("/* + / * */\n")
				indenter.Add("for accept {\n")
				indenter.Inc()
			}
			p.helper(node.Children.Front().Value.(*Node))
			switch back.Name {
			case "STAR", "PLUS":
				indenter.Dec()
				indenter.Add("}\n")
			}
			indenter.Add("/* ? / * / + */\naccept = true\n")
			if back.Name == "PLUS" {
				indenter.Add(label + ":\n")
			}
			indenter.Dec()
			indenter.Add("}\n")
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
				indenter.Add("/* inlined " + fd + " */\n{\n")
				indenter.Inc()

				// just to prevent an inline from inlining itself
				visitedMap[fd] = true
				p.helper(r)
				visitedMap[fd] = false
				/*
					indenter.Add(r + "\n")
				*/
				indenter.Dec()
				indenter.Add("}\n")
			} else {
				indenter.Add("accept = p.")
				p.helper(front)
				indenter.Add("()\n")
			}
			return
		} else if front.Name == "OPEN" {
			p.helper(node.Children.Front().Next().Value.(*Node))
			return
		}
	case "Spacing", "Space":
		// ignore
	default:
		indenter.Add("\n\n-----------------------------------------------------\n" + node.Name + ", " + node.Data + "-----------------------------------------------------\n")
	}
	//	indenter.Inc()
	for n := node.Children.Front(); n != nil; n = n.Next() {
		p.helper(n.Value.(*Node))
	}
	//	indenter.Dec()
}

func (p *PegParser) RootNode() *Node {
	return p.currentNode.Children.Front().Value.(*Node)
}

func (p *PegParser) Dump() {
	rootNode := p.RootNode()
	if enableInlining {
		for n := rootNode.Children.Back(); n != nil; n = n.Prev() {
			node := n.Value.(*Node)
			if node.Name == "Definition" {
				id := node.Children.Front().Value.(*Node)
				exp := node.Children.Back().Value.(*Node)

				indenter = CodeFormatter{}
				p.helper(id)
				defName := indenter.String()
				convMap[defName] = exp
			}
		}
	}
	output := fmt.Sprintln(`package parser
import (
	"strings"
)

type PegParser2 struct {
	Parser
}`)

	for n := rootNode.Children.Front(); n != nil; n = n.Next() {
		node := n.Value.(*Node)
		if node.Name == "Definition" {
			id := node.Children.Front().Value.(*Node)
			exp := node.Children.Back().Value.(*Node)
			indenter = CodeFormatter{}
			p.helper(exp)
			data := indenter.String()
			indenter = CodeFormatter{}
			p.helper(id)
			defName := indenter.String()
			//			fmt.Println(node)
			indenter = CodeFormatter{}
			indenter.Inc()
			indenter.Add("func (p *PegParser2) " + defName + "() bool {\n")

			indenter.Add("accept := true\n")
			indenter.Add(data)
			indenter.Add("return accept\n")
			indenter.Dec()
			indenter.Add("}\n")
			output += indenter.String()
			// if enableInlining {
			// 	break
			// }
			//			break
		}
	}
	ioutil.WriteFile("./parser2.go", []byte(output), 0644)
}

// if __name__ == "__main__":
//     f = open("peg.peg")
//     peg = f.read()
//     f.close()

//     parser = PegPegParser(peg)
//     parser.Grammar()

//     _indent = "\t\t"
//     def get_data(node):
//         global _indent
//         name = node.name
//         names = {"PLUS": "func() bool { return p.OneOrMore",
//                  "STAR": "func() bool { return p.ZeroOrMore",
//                  "QUESTION": "func() bool { return p.Maybe",
//                  "NOT": "func() bool { return p.Not",
//                  "AND": "func() bool { return p.And"
//                  }
//         nodes = node.nodes
//         if len(nodes) > 0 and nodes[-1].name == "Spacing":
//             nodes = nodes[:-1]

//         if name == "OPEN" or name == "CLOSE":
//             return ""
//         elif name == "Identifier":
//             if len(node.nodes) > 0 and node.nodes[-1].name == "Spacing":
//                 return "p." + parser._PegParserdata[node.pos:node.nodes[-1].pos]
//             return "p." + node.data
//         elif name == "Literal":
//             return "func() bool { return p.Next(" + node.data.replace("\"", "\\\"").strip()  + ")"
//         elif name == "Range":
//             if len(node.nodes) == 2:
//                 return "func() bool { return p.InRange('%s', '%s')" % (node.nodes[0].data, node.nodes[1].data)
//             else:
//                 return node.data
//         elif name == "Prefix":
//             if len(node.nodes) == 1:
//                 return get_data(node.nodes[0})
//             else:
//                 return names[node.nodes[0].name] + "(" + get_data(node.nodes[1}) + ")"
//         elif name == "Suffix":
//             data = get_data(node.nodes[0})
//             if len(node.nodes) > 1:
//                 data = names[node.nodes[1].name] + "(" + data + ")"
//             return data
//         elif name == "Sequence":
//             if len(node.nodes) == 1:
//                 return get_data(node.nodes[0})
//             data = ""
//             _indent += "\t"
//             for n in node.nodes:
//                 if len(data) > 0:
//                     data += ","
//                 data += "\n%s%s" % (_indent.replace("\t", "    "), get_data(n))
//             _indent = _indent[:-1]
//             return "func() bool { return p.NeedAll([]func() bool {" + data + "})"
//         elif name == "DOT":
//             return "func() bool { return p.AnyChar()"
//         elif name == "Expression":
//             if len(node.nodes) == 1:
//                 return get_data(node.nodes[0})
//             else:
//                 data = ""
//                 _indent += "\t"
//                 for i in range(len(node.nodes)):
//                     if (i & 1) == 0:
//                         if len(data) > 0:
//                             data += ","
//                         data += "\n%s%s" % (_indent.replace("\t", "    "), get_data(node.nodes[i}))
//                 _indent = _indent[:-1]
//                 data = "func() bool { return p.NeedOne([]func() bool {" + data + "})"
//                 return data
//         elif name == "Class":
//             data = []
//             others = ""
//             l = 0

//             for n in nodes:
//                 sub = get_data(n).replace("\\", "\\\\").replace("\"", "\\\"")
//                 if not sub.startswith("lambda"):
//                     others += sub
//                 else:
//                     data.append(sub)
//                     l++

//             if len(others) > 0:
//                 l++
//                 data.append("%sfunc() bool { return p.InSet(\"%s\")" % ("\n%s" % _indent.replace("\t", "    ") if len(nodes) != 1 else "", others))

//             if len(data) != 1:
//                 _indent += "\t"

//             s = ",\n%s" % (_indent.replace("\t", "    "))
//             data = s.join(data)

//             if l == 1:
//                 return data
//             else:
//                 _indent = _indent[:-1]
//                 return "func() bool { return p.NeedOne([]func() bool {" + data + "})"
//         r = ""
//         for n in nodes:
//             r += get_data(n)
//         return r

//     f = open(__file__)
//     data = f.read()
//     f.close()

//     print data[:data.find("class PegPegParser")].strip()
//     print "\nclass PegPegParser(PegParser):"
//     for node in parser.currentNode.nodes[0].nodes:
//         if node.name == "Definition":
//             print "    @addNode\n    def %s(self):\n        %s\n" % (get_data(node.nodes[0})[5:], get_data(node.nodes[2}).replace("func() bool { return ", "return ", 1))
//     print data[data.find("""if __name__ == "__main__":"""):],
