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
	"unicode/utf8"
)

const (
	Nilrune = '\u0000'
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
		Size int
		Pos  int
		Data string
	}
}

func (p *Parser) ReadRune() rune {
	if p.ParserData.Pos >= len(p.ParserData.Data) {
		return Nilrune
	}
	if c := p.ParserData.Data[p.ParserData.Pos]; c < utf8.RuneSelf {
		p.ParserData.Pos++
		return rune(c)
	}
	ch, size := utf8.DecodeRuneInString(p.ParserData.Data[p.ParserData.Pos:])
	p.ParserData.Pos += size
	return ch
}

func (p *Parser) SetData(data string) {
	p.ParserData.Size = len(data)
	p.ParserData.Data = data
	p.Reset()
}

func (p *Parser) Reset() {
	p.ParserData.Pos = 0
	p.Root = Node{}
}

func (p *Parser) Data(start, end int) string {
	return p.ParserData.Data[start:end]
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
	return p.ReadRune() != Nilrune
}

func (p *Parser) InRange(c1, c2 rune) bool {
	save := p.ParserData.Pos
	if c := p.ReadRune(); c == Nilrune {
		return false
	} else {
		if c >= c1 && c <= c2 {
			return true
		}
	}
	p.ParserData.Pos = save
	return false
}

func (p *Parser) InSet(dataset string) bool {
	save := p.ParserData.Pos
	if c := p.ReadRune(); c == Nilrune {
		return false
	} else {
		if strings.ContainsRune(dataset, c) {
			return true
		}
	}
	p.ParserData.Pos = save
	return false
}

func (p *Parser) Next(n1 string) bool {
	s := p.ParserData.Pos
	e := s + len(n1)
	if e > p.ParserData.Size {
		return false
	}
	// Todo: should really build a string out of runes to be 100% correct.
	if !strings.EqualFold(n1, p.ParserData.Data[s:e]) {
		return false
	}
	p.ParserData.Pos += len(n1)
	return true
}
