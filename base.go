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
	"strings"
)

func (p *Parser) AddNode(add func() bool, name string) bool {
	start := p.ParserData.Pos
	shouldAdd := add()
	p.Root.P = p
	// Remove any danglers
	p.Root.Cleanup(p.ParserData.Pos, -1)

	node := p.Root.Cleanup(start, p.ParserData.Pos)
	node.Name = name
	if shouldAdd {
		node.P = p
		p.Root.Append(node)
	}
	return shouldAdd
}

type Parser struct {
	Root       Node
	ParserData struct {
		Pos  int
		Data []rune
	}
}

func (p *Parser) SetData(data string) {
	p.ParserData.Data = ([]rune)(data)
	p.Reset()
}

func (p *Parser) Reset() {
	p.ParserData.Pos = 0
	p.Root = Node{}
}

func (p *Parser) Data(start, end int) string {
	return string(p.ParserData.Data[start:end])
}

func (p *Parser) RootNode() *Node {
	return p.Root.Children[0]
}

func (p *Parser) Maybe(exp func() bool) bool {
	exp()
	return true
}

func (p *Parser) OneOrMore(exp func() bool) bool {
	save := p.ParserData.Pos
	if !exp() {
		p.ParserData.Pos = save
		return false
	}
	for exp() {
	}
	return true
}

func (p *Parser) NeedAll(exps []func() bool) bool {
	save := p.ParserData.Pos
	for _, exp := range exps {
		if !exp() {
			p.ParserData.Pos = save
			return false
		}
	}
	return true
}

func (p *Parser) NeedOne(exps []func() bool) bool {
	save := p.ParserData.Pos
	for _, exp := range exps {
		if exp() {
			return true
		}
	}
	p.ParserData.Pos = save
	return false
}

func (p *Parser) ZeroOrMore(exp func() bool) bool {
	for exp() {
	}
	return true
}

func (p *Parser) And(exp func() bool) bool {
	save := p.ParserData.Pos
	ret := exp()
	p.ParserData.Pos = save
	return ret
}

func (p *Parser) Not(exp func() bool) bool {
	return !p.And(exp)
}

func (p *Parser) AnyChar() bool {
	if p.ParserData.Pos >= len(p.ParserData.Data) {
		return false
	}
	p.ParserData.Pos++
	return true
}

func (p *Parser) InRange(c1, c2 rune) bool {
	if p.ParserData.Pos >= len(p.ParserData.Data) {
		return false
	}
	c := p.ParserData.Data[p.ParserData.Pos]
	if c >= c1 && c <= c2 {
		p.ParserData.Pos++
		return true
	}
	return false
}

func (p *Parser) InSet(dataset string) bool {
	if p.ParserData.Pos >= len(p.ParserData.Data) {
		return false
	}
	c := p.ParserData.Data[p.ParserData.Pos]
	if strings.ContainsRune(dataset, c) {
		p.ParserData.Pos++
		return true
	}
	return false
}
func (p *Parser) NextRune(n1 rune) bool {
	if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != n1 {
		return false
	}
	p.ParserData.Pos++
	return true
}
func (p *Parser) Next(n1 []rune) bool {
	s := p.ParserData.Pos
	e := s + len(n1)
	if e > len(p.ParserData.Data) {
		return false
	}
	for i := 0; i < len(n1); i++ {
		if n1[i] != p.ParserData.Data[s+i] {
			return false
		}
	}
	p.ParserData.Pos += len(n1)
	return true
}
