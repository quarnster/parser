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
