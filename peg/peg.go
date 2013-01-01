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
package peg

import (
	"parser"
	"strings"
)

type Peg struct {
	ParserData struct {
	Pos int
	Data []rune
}

	IgnoreRange parser.Range
	Root parser.Node
}

func (p *Peg) RootNode() *parser.Node {
	return p.Root.Children[0]
}

func (p *Peg) SetData(data string) {
	p.ParserData.Data = ([]rune)(data)
	p.Reset()
}

func (p *Peg) Reset() {
	p.ParserData.Pos = 0
	p.Root = parser.Node{}
	p.IgnoreRange = parser.Range{}
}

func (p *Peg) Data(start, end int) string {
	l := len(p.ParserData.Data)
	if l == 0 {
		return ""
	}
	if start < 0 {
		start = 0
	}
	if end > l {
		end = l
	}
	if start > end {
		return ""
	}
	return string(p.ParserData.Data[start:end])

}

func p_Ignore(p *Peg, add func(*Peg) bool) bool {
	start := p.ParserData.Pos
	ret := add(p)
	if ret {
		if start < p.IgnoreRange.Start || p.IgnoreRange.Start == 0 {
			p.IgnoreRange.Start = start
		}
		p.IgnoreRange.End = p.ParserData.Pos
	}
	return ret
}

func p_addNode(p *Peg, add func(*Peg) bool, name string) bool {
	start := p.ParserData.Pos
	shouldAdd := add(p)
	end := p.ParserData.Pos
	p.Root.P = p
	// Remove any danglers
	p.Root.Cleanup(p.ParserData.Pos, -1)

	node := p.Root.Cleanup(start, p.ParserData.Pos)
	node.Name = name
	if shouldAdd {
		node.P = p
		node.Range.Clip(p.IgnoreRange)
		c := make([]*parser.Node, len(node.Children))
		copy(c, node.Children)
		node.Children = c
		p.Root.Append(node)
	}
	if p.IgnoreRange.Start >= end || p.IgnoreRange.End <= start {
		p.IgnoreRange = parser.Range{}
	}
	return shouldAdd
}

func p_Maybe(p *Peg, exp func(*Peg) bool) bool {
	exp(p)
	return true
}

func p_OneOrMore(p *Peg, exp func(*Peg) bool) bool {
	save := p.ParserData.Pos
	if !exp(p) {
		p.ParserData.Pos = save
		return false
	}
	for exp(p) {
	}
	return true
}

func p_NeedAll(p *Peg, exps []func(*Peg) bool) bool {
	save := p.ParserData.Pos
	for _, exp := range exps {
		if !exp(p) {
			p.ParserData.Pos = save
			return false
		}
	}
	return true
}

func p_NeedOne(p *Peg, exps []func(*Peg) bool) bool {
	save := p.ParserData.Pos
	for _, exp := range exps {
		if exp(p) {
			return true
		}
	}
	p.ParserData.Pos = save
	return false
}

func p_ZeroOrMore(p *Peg, exp func(*Peg) bool) bool {
	for exp(p) {
	}
	return true
}

func p_And(p *Peg, exp func(*Peg) bool) bool {
	save := p.ParserData.Pos
	ret := exp(p)
	p.ParserData.Pos = save
	return ret
}

func p_Not(p *Peg, exp func(*Peg) bool) bool {
	return !p_And(p, exp)
}

func p_AnyChar(p *Peg) bool {
	if p.ParserData.Pos >= len(p.ParserData.Data) {
		return false
	}
	p.ParserData.Pos++
	return true
}

func p_InRange(p *Peg, c1, c2 rune) bool {
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

func p_InSet(p *Peg, dataset string) bool {
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

func p_NextRune(p *Peg, n1 rune) bool {
	if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != n1 {
		return false
	}
	p.ParserData.Pos++
	return true
}

func p_Next(p *Peg, n1 []rune) bool {
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
func (p *Peg) Parse() bool {
	return p_Grammar(p)
}
func helper0_Grammar(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		p_Spacing,
		func(p *Peg) bool { return p_OneOrMore(p, p_Definition) },
		func(p *Peg) bool { return p_Maybe(p, p_EndOfFile) },
	})
}
func p_Grammar(p *Peg) bool {
	// Grammar       <- Spacing Definition+ EndOfFile?
	return p_addNode(p, helper0_Grammar, "Grammar")
}

func helper0_Definition(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		p_Identifier,
		p_LEFTARROW,
		p_Expression,
	})
}
func p_Definition(p *Peg) bool {
	// Definition    <- Identifier LEFTARROW Expression
	return p_addNode(p, helper0_Definition, "Definition")
}

func helper0_Expression(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		p_SLASH,
		p_Sequence,
	})
}
func helper1_Expression(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		p_Sequence,
		func(p *Peg) bool { return p_ZeroOrMore(p, helper0_Expression) },
	})
}
func p_Expression(p *Peg) bool {
	// Expression    <- Sequence (SLASH Sequence)*
	return p_addNode(p, helper1_Expression, "Expression")
}

func p_Sequence(p *Peg) bool {
	// Sequence      <- Prefix*
	return p_addNode(p, func(p *Peg) bool { return p_ZeroOrMore(p, p_Prefix) }, "Sequence")
}

func helper0_Prefix(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		p_AND,
		p_NOT,
	})
}
func helper1_Prefix(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		func(p *Peg) bool { return p_Maybe(p, helper0_Prefix) },
		p_Suffix,
	})
}
func p_Prefix(p *Peg) bool {
	// Prefix        <- (AND / NOT)? Suffix
	return p_addNode(p, helper1_Prefix, "Prefix")
}

func helper0_Suffix(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		p_QUESTION,
		p_STAR,
		p_PLUS,
	})
}
func helper1_Suffix(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		p_Primary,
		func(p *Peg) bool { return p_Maybe(p, helper0_Suffix) },
	})
}
func p_Suffix(p *Peg) bool {
	// Suffix        <- Primary (QUESTION / STAR / PLUS)?
	return p_addNode(p, helper1_Suffix, "Suffix")
}

func helper0_Primary(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		p_Identifier,
		func(p *Peg) bool { return p_Not(p, p_LEFTARROW) },
	})
}
func helper1_Primary(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		p_OPEN,
		p_Expression,
		p_CLOSE,
	})
}
func helper2_Primary(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		helper0_Primary,
		helper1_Primary,
		p_Literal,
		p_Class,
		p_DOT,
	})
}
func p_Primary(p *Peg) bool {
	// Primary       <- Identifier !LEFTARROW
	//                / OPEN Expression CLOSE
	//                / Literal / Class / DOT
	// # Lexical syntax
	return p_addNode(p, helper2_Primary, "Primary")
}

func helper0_Identifier(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		p_IdentStart,
		func(p *Peg) bool { return p_ZeroOrMore(p, p_IdentCont) },
		p_Spacing,
	})
}
func p_Identifier(p *Peg) bool {
	// Identifier    <- IdentStart IdentCont* Spacing
	return p_addNode(p, helper0_Identifier, "Identifier")
}

func helper0_IdentStart(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		func(p *Peg) bool { return p_InRange(p, 'a', 'z') },
		func(p *Peg) bool { return p_InRange(p, 'A', 'Z') },
		func(p *Peg) bool { return p_InSet(p, "_") },
	})
}
func p_IdentStart(p *Peg) bool {
	// IdentStart    <- [a-zA-Z_]
	return helper0_IdentStart(p)
}

func helper0_IdentCont(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		p_IdentStart,
		func(p *Peg) bool { return p_InRange(p, '0', '9') },
	})
}
func p_IdentCont(p *Peg) bool {
	// IdentCont     <- IdentStart / [0-9]
	return helper0_IdentCont(p)
}

func helper0_Literal(p *Peg) bool {
	return p_Not(p, func(p *Peg) bool { return p_InSet(p, "'") })
}
func helper1_Literal(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_Literal,
		p_Char,
	})
}
func helper2_Literal(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		func(p *Peg) bool { return p_InSet(p, "'") },
		func(p *Peg) bool { return p_ZeroOrMore(p, helper1_Literal) },
		func(p *Peg) bool { return p_InSet(p, "'") },
		p_Spacing,
	})
}
func helper3_Literal(p *Peg) bool {
	return p_Not(p, func(p *Peg) bool { return p_InSet(p, "\"") })
}
func helper4_Literal(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper3_Literal,
		p_Char,
	})
}
func helper5_Literal(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		func(p *Peg) bool { return p_InSet(p, "\"") },
		func(p *Peg) bool { return p_ZeroOrMore(p, helper4_Literal) },
		func(p *Peg) bool { return p_InSet(p, "\"") },
		p_Spacing,
	})
}
func helper6_Literal(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		helper2_Literal,
		helper5_Literal,
	})
}
func p_Literal(p *Peg) bool {
	// Literal       <- ['] (!['] Char)* ['] Spacing
	//                / ["] (!["] Char)* ["] Spacing
	return p_addNode(p, helper6_Literal, "Literal")
}

func helper0_Class(p *Peg) bool {
	return p_Not(p, func(p *Peg) bool { return p_NextRune(p, ']') })
}
func helper1_Class(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_Class,
		p_Range,
	})
}
func helper2_Class(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		func(p *Peg) bool { return p_NextRune(p, '[') },
		func(p *Peg) bool { return p_ZeroOrMore(p, helper1_Class) },
		func(p *Peg) bool { return p_NextRune(p, ']') },
		p_Spacing,
	})
}
func p_Class(p *Peg) bool {
	// Class         <- '[' (!']' Range)* ']' Spacing
	return p_addNode(p, helper2_Class, "Class")
}

func helper0_Range(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		p_Char,
		func(p *Peg) bool { return p_NextRune(p, '-') },
		p_Char,
	})
}
func helper1_Range(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		helper0_Range,
		p_Char,
	})
}
func p_Range(p *Peg) bool {
	// Range         <- Char '-' Char / Char
	return p_addNode(p, helper1_Range, "Range")
}

func helper0_Char(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		func(p *Peg) bool { return p_NextRune(p, '\\') },
		func(p *Peg) bool { return p_InSet(p, "nrt'\"[]\\") },
	})
}
func helper1_Char(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		func(p *Peg) bool { return p_NextRune(p, '\\') },
		func(p *Peg) bool { return p_InRange(p, '0', '2') },
		func(p *Peg) bool { return p_InRange(p, '0', '7') },
		func(p *Peg) bool { return p_InRange(p, '0', '7') },
	})
}
func helper2_Char(p *Peg) bool {
	return p_Maybe(p, func(p *Peg) bool { return p_InRange(p, '0', '7') })
}
func helper3_Char(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		func(p *Peg) bool { return p_NextRune(p, '\\') },
		func(p *Peg) bool { return p_InRange(p, '0', '7') },
		helper2_Char,
	})
}
func helper4_Char(p *Peg) bool {
	return p_Not(p, func(p *Peg) bool { return p_NextRune(p, '\\') })
}
func helper5_Char(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper4_Char,
		p_AnyChar,
	})
}
func helper6_Char(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		helper0_Char,
		helper1_Char,
		helper3_Char,
		helper5_Char,
	})
}
func p_Char(p *Peg) bool {
	// Char          <- '\\' [nrt'"\[\]\\]
	//                / '\\' [0-2][0-7][0-7]
	//                / '\\' [0-7][0-7]?
	//                / !'\\' .
	return p_addNode(p, helper6_Char, "Char")
}

func helper0_LEFTARROW(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		func(p *Peg) bool { return p_Next(p, []rune{'<', '-'}) },
		p_Spacing,
	})
}
func p_LEFTARROW(p *Peg) bool {
	// LEFTARROW     <- "<-" Spacing
	return p_Ignore(p, helper0_LEFTARROW)
}

func helper0_SLASH(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		func(p *Peg) bool { return p_NextRune(p, '/') },
		p_Spacing,
	})
}
func p_SLASH(p *Peg) bool {
	// SLASH         <- '/' Spacing
	return p_Ignore(p, helper0_SLASH)
}

func helper0_AND(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		func(p *Peg) bool { return p_NextRune(p, '&') },
		p_Spacing,
	})
}
func p_AND(p *Peg) bool {
	// AND           <- '&' Spacing
	return p_addNode(p, helper0_AND, "AND")
}

func helper0_NOT(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		func(p *Peg) bool { return p_NextRune(p, '!') },
		p_Spacing,
	})
}
func p_NOT(p *Peg) bool {
	// NOT           <- '!' Spacing
	return p_addNode(p, helper0_NOT, "NOT")
}

func helper0_QUESTION(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		func(p *Peg) bool { return p_NextRune(p, '?') },
		p_Spacing,
	})
}
func p_QUESTION(p *Peg) bool {
	// QUESTION      <- '?' Spacing
	return p_addNode(p, helper0_QUESTION, "QUESTION")
}

func helper0_STAR(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		func(p *Peg) bool { return p_NextRune(p, '*') },
		p_Spacing,
	})
}
func p_STAR(p *Peg) bool {
	// STAR          <- '*' Spacing
	return p_addNode(p, helper0_STAR, "STAR")
}

func helper0_PLUS(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		func(p *Peg) bool { return p_NextRune(p, '+') },
		p_Spacing,
	})
}
func p_PLUS(p *Peg) bool {
	// PLUS          <- '+' Spacing
	return p_addNode(p, helper0_PLUS, "PLUS")
}

func helper0_OPEN(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		func(p *Peg) bool { return p_NextRune(p, '(') },
		p_Spacing,
	})
}
func p_OPEN(p *Peg) bool {
	// OPEN          <- '(' Spacing
	return p_Ignore(p, helper0_OPEN)
}

func helper0_CLOSE(p *Peg) bool {
	return p_NextRune(p, ')')
}
func helper1_CLOSE(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_CLOSE,
		p_Spacing,
	})
}
func p_CLOSE(p *Peg) bool {
	// CLOSE         <- ')' Spacing
	return p_Ignore(p, helper1_CLOSE)
}

func helper0_DOT(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		func(p *Peg) bool { return p_NextRune(p, '.') },
		p_Spacing,
	})
}
func p_DOT(p *Peg) bool {
	// DOT           <- '.' Spacing
	return p_addNode(p, helper0_DOT, "DOT")
}

func helper0_Spacing(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		p_Space,
		p_Comment,
	})
}
func p_Spacing(p *Peg) bool {
	// Spacing       <- (Space / Comment)*
	return p_Ignore(p, func(p *Peg) bool { return p_ZeroOrMore(p, helper0_Spacing) })
}

func helper0_Comment(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		func(p *Peg) bool { return p_Not(p, p_EndOfLine) },
		p_AnyChar,
	})
}
func helper1_Comment(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		func(p *Peg) bool { return p_NextRune(p, '#') },
		func(p *Peg) bool { return p_ZeroOrMore(p, helper0_Comment) },
		p_EndOfLine,
	})
}
func p_Comment(p *Peg) bool {
	// Comment       <- '#' (!EndOfLine .)* EndOfLine
	return p_Ignore(p, helper1_Comment)
}

func helper0_Space(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		func(p *Peg) bool { return p_NextRune(p, ' ') },
		func(p *Peg) bool { return p_NextRune(p, '\t') },
		p_EndOfLine,
	})
}
func p_Space(p *Peg) bool {
	// Space         <- ' ' / '\t' / EndOfLine
	return p_Ignore(p, helper0_Space)
}

func helper0_EndOfLine(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		func(p *Peg) bool { return p_Next(p, []rune{'\r', '\n'}) },
		func(p *Peg) bool { return p_NextRune(p, '\n') },
		func(p *Peg) bool { return p_NextRune(p, '\r') },
	})
}
func p_EndOfLine(p *Peg) bool {
	// EndOfLine     <- "\r\n" / '\n' / '\r'
	return p_Ignore(p, helper0_EndOfLine)
}

func p_EndOfFile(p *Peg) bool {
	// EndOfFile     <- !.
	return p_Ignore(p, func(p *Peg) bool { return p_Not(p, p_AnyChar) })
}
