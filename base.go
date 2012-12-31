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

import (
	"bytes"
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

func (i *IntStack) Clear() {
	i.data = i.data[:]
}

func (p *Parser) AddNode(add func() bool, name string) bool {
	start := p.Pos()
	shouldAdd := add()
	p.Root.P = p
	// Remove any danglers
	p.Root.Cleanup(p.Pos(), -1)

	node := p.Root.Cleanup(start, p.Pos())
	node.Name = name
	if shouldAdd {
		node.P = p
		p.Root.Append(node)
	}
	return shouldAdd
}

type Parser struct {
	stack      IntStack
	ParserData *strings.Reader
	Root       Node
}

func (p *Parser) Init() {
}

func (p *Parser) SetData(data string) {
	p.ParserData = strings.NewReader(data)
}

func (p *Parser) Reset() {
	p.Root = Node{}
	p.stack = IntStack{}
	p.ParserData.Seek(0, 0)
}

func (p *Parser) Data(start, end int) string {
	p.Push()
	defer p.Reject()
	p.ParserData.Seek(int64(start), 0)
	b := make([]byte, end-start)
	p.ParserData.Read(b)
	return string(b)
}

func (p *Parser) RootNode() *Node {
	return p.Root.Children[0]
}

func (p *Parser) Pos() int {
	pos := p.ParserData.Len()
	p.ParserData.Seek(0, 0)
	pos = p.ParserData.Len() - pos
	p.ParserData.Seek(int64(pos), 0)
	return pos
}

func (p *Parser) Push() {
	p.stack.Push(p.Pos())
}
func (p *Parser) Pop() int {
	return p.stack.Pop()
}

func (p *Parser) Accept() bool {
	p.Pop()
	return true
}

func (p *Parser) Reject() bool {
	p.ParserData.Seek(int64(p.Pop()), 0)
	return false
}

func (p *Parser) Maybe(exp func() bool) bool {
	exp()
	return true
}

func (p *Parser) OneOrMore(exp func() bool) bool {
	p.Push()
	if !exp() {
		return p.Reject()
	}
	for exp() {
	}
	return p.Accept()
}

func (p *Parser) NeedAll(exps []func() bool) bool {
	p.Push()
	for _, exp := range exps {
		if !exp() {
			return p.Reject()
		}
	}
	return p.Accept()
}

func (p *Parser) NeedOne(exps []func() bool) bool {
	p.Push()
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
	p.Push()
	ret := exp()
	p.Reject()
	return ret
}

func (p *Parser) Not(exp func() bool) bool {
	return !p.And(exp)
}

func (p *Parser) AnyChar() bool {
	if _, _, err := p.ParserData.ReadRune(); err == nil {
		return true
	}
	return false
}

func (p *Parser) PushNext() (rune, bool) {
	p.Push()
	c, _, err := p.ParserData.ReadRune()
	if err != nil {
		return nilrune, false
	}
	return c, true
}

func (p *Parser) InRange(c1, c2 rune) bool {
	if c, ok := p.PushNext(); !ok {
		return p.Reject()
	} else {
		if c >= c1 && c <= c2 {
			return p.Accept()
		}
	}
	return p.Reject()
}

func (p *Parser) InSet(dataset string) bool {
	if c, ok := p.PushNext(); !ok {
		return false
	} else {
		if strings.ContainsRune(dataset, c) {
			return p.Accept()
		}
	}
	return p.Reject()
}
func (p *Parser) Next(n1 string) bool {
	p.Push()
	n2 := make([]byte, len(n1))
	if n, err := p.ParserData.Read(n2); err != nil || n != len(n2) || bytes.Compare([]byte(n1), n2) != 0 {
		return p.Reject()
	}
	return p.Accept()
}
