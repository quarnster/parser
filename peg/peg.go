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
)

type Peg struct {
	parser.Parser
}

const (
	nilrune = '\u0000'
)

func p_addNode(p *Peg, add func(*Peg) bool, name string) bool {
	start := p.ParserData.Pos
	shouldAdd := add(p)
	p.Root.P = &p.Parser
	// Remove any danglers
	p.Root.Cleanup(p.ParserData.Pos, -1)

	node := p.Root.Cleanup(start, p.ParserData.Pos)
	node.Name = name
	if shouldAdd {
		node.P = &p.Parser
		c := make([]*parser.Node, len(node.Children))
		copy(c, node.Children)
		node.Children = c
		p.Root.Append(node)
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

func (p *Peg) Parse() bool {
	return p_Grammar(p)
}
func helper0_Grammar(p *Peg) bool {
	return p_OneOrMore(p, p_Definition)
}
func helper1_Grammar(p *Peg) bool {
	return p_Maybe(p, p_EndOfFile)
}
func helper2_Grammar(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		p_Spacing,
		helper0_Grammar,
		helper1_Grammar,
	})
}
func p_Grammar(p *Peg) bool {
	// Grammar       <- Spacing Definition+ EndOfFile?
	return p_addNode(p, helper2_Grammar, "Grammar")
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
	return p_ZeroOrMore(p, helper0_Expression)
}
func helper2_Expression(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		p_Sequence,
		helper1_Expression,
	})
}
func p_Expression(p *Peg) bool {
	// Expression    <- Sequence (SLASH Sequence)*
	return p_addNode(p, helper2_Expression, "Expression")
}

func helper0_Sequence(p *Peg) bool {
	return p_ZeroOrMore(p, p_Prefix)
}
func p_Sequence(p *Peg) bool {
	// Sequence      <- Prefix*
	return p_addNode(p, helper0_Sequence, "Sequence")
}

func helper0_Prefix(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		p_AND,
		p_NOT,
	})
}
func helper1_Prefix(p *Peg) bool {
	return p_Maybe(p, helper0_Prefix)
}
func helper2_Prefix(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper1_Prefix,
		p_Suffix,
	})
}
func p_Prefix(p *Peg) bool {
	// Prefix        <- (AND / NOT)? Suffix
	return p_addNode(p, helper2_Prefix, "Prefix")
}

func helper0_Suffix(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		p_QUESTION,
		p_STAR,
		p_PLUS,
	})
}
func helper1_Suffix(p *Peg) bool {
	return p_Maybe(p, helper0_Suffix)
}
func helper2_Suffix(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		p_Primary,
		helper1_Suffix,
	})
}
func p_Suffix(p *Peg) bool {
	// Suffix        <- Primary (QUESTION / STAR / PLUS)?
	return p_addNode(p, helper2_Suffix, "Suffix")
}

func helper0_Primary(p *Peg) bool {
	return p_Not(p, p_LEFTARROW)
}
func helper1_Primary(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		p_Identifier,
		helper0_Primary,
	})
}
func helper2_Primary(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		p_OPEN,
		p_Expression,
		p_CLOSE,
	})
}
func helper3_Primary(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		helper1_Primary,
		helper2_Primary,
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
	return p_addNode(p, helper3_Primary, "Primary")
}

func helper0_Identifier(p *Peg) bool {
	return p_ZeroOrMore(p, p_IdentCont)
}
func helper1_Identifier(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		p_IdentStart,
		helper0_Identifier,
		p_Spacing,
	})
}
func p_Identifier(p *Peg) bool {
	// Identifier    <- IdentStart IdentCont* Spacing
	return p_addNode(p, helper1_Identifier, "Identifier")
}

func helper0_IdentStart(p *Peg) bool {
	return p.InRange('a', 'z')
}
func helper1_IdentStart(p *Peg) bool {
	return p.InRange('A', 'Z')
}
func helper2_IdentStart(p *Peg) bool {
	return p.InSet("_")
}
func helper3_IdentStart(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		helper0_IdentStart,
		helper1_IdentStart,
		helper2_IdentStart,
	})
}
func p_IdentStart(p *Peg) bool {
	// IdentStart    <- [a-zA-Z_]
	return p_addNode(p, helper3_IdentStart, "IdentStart")
}

func helper0_IdentCont(p *Peg) bool {
	return p.InRange('0', '9')
}
func helper1_IdentCont(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		p_IdentStart,
		helper0_IdentCont,
	})
}
func p_IdentCont(p *Peg) bool {
	// IdentCont     <- IdentStart / [0-9]
	return p_addNode(p, helper1_IdentCont, "IdentCont")
}

func helper0_Literal(p *Peg) bool {
	return p.InSet("'")
}
func helper1_Literal(p *Peg) bool {
	return p.InSet("'")
}
func helper2_Literal(p *Peg) bool {
	return p_Not(p, helper1_Literal)
}
func helper3_Literal(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper2_Literal,
		p_Char,
	})
}
func helper4_Literal(p *Peg) bool {
	return p_ZeroOrMore(p, helper3_Literal)
}
func helper5_Literal(p *Peg) bool {
	return p.InSet("'")
}
func helper6_Literal(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_Literal,
		helper4_Literal,
		helper5_Literal,
		p_Spacing,
	})
}
func helper7_Literal(p *Peg) bool {
	return p.InSet("\"")
}
func helper8_Literal(p *Peg) bool {
	return p.InSet("\"")
}
func helper9_Literal(p *Peg) bool {
	return p_Not(p, helper8_Literal)
}
func helper10_Literal(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper9_Literal,
		p_Char,
	})
}
func helper11_Literal(p *Peg) bool {
	return p_ZeroOrMore(p, helper10_Literal)
}
func helper12_Literal(p *Peg) bool {
	return p.InSet("\"")
}
func helper13_Literal(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper7_Literal,
		helper11_Literal,
		helper12_Literal,
		p_Spacing,
	})
}
func helper14_Literal(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		helper6_Literal,
		helper13_Literal,
	})
}
func p_Literal(p *Peg) bool {
	// Literal       <- ['] (!['] Char)* ['] Spacing
	//                / ["] (!["] Char)* ["] Spacing
	return p_addNode(p, helper14_Literal, "Literal")
}

func helper0_Class(p *Peg) bool {
	return p.Next("[")
}
func helper1_Class(p *Peg) bool {
	return p.Next("]")
}
func helper2_Class(p *Peg) bool {
	return p_Not(p, helper1_Class)
}
func helper3_Class(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper2_Class,
		p_Range,
	})
}
func helper4_Class(p *Peg) bool {
	return p_ZeroOrMore(p, helper3_Class)
}
func helper5_Class(p *Peg) bool {
	return p.Next("]")
}
func helper6_Class(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_Class,
		helper4_Class,
		helper5_Class,
		p_Spacing,
	})
}
func p_Class(p *Peg) bool {
	// Class         <- '[' (!']' Range)* ']' Spacing
	return p_addNode(p, helper6_Class, "Class")
}

func helper0_Range(p *Peg) bool {
	return p.Next("-")
}
func helper1_Range(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		p_Char,
		helper0_Range,
		p_Char,
	})
}
func helper2_Range(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		helper1_Range,
		p_Char,
	})
}
func p_Range(p *Peg) bool {
	// Range         <- Char '-' Char / Char
	return p_addNode(p, helper2_Range, "Range")
}

func helper0_Char(p *Peg) bool {
	return p.Next("\\")
}
func helper1_Char(p *Peg) bool {
	return p.InSet("nrt'\"[]\\")
}
func helper2_Char(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_Char,
		helper1_Char,
	})
}
func helper3_Char(p *Peg) bool {
	return p.Next("\\")
}
func helper4_Char(p *Peg) bool {
	return p.InRange('0', '2')
}
func helper5_Char(p *Peg) bool {
	return p.InRange('0', '7')
}
func helper6_Char(p *Peg) bool {
	return p.InRange('0', '7')
}
func helper7_Char(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper3_Char,
		helper4_Char,
		helper5_Char,
		helper6_Char,
	})
}
func helper8_Char(p *Peg) bool {
	return p.Next("\\")
}
func helper9_Char(p *Peg) bool {
	return p.InRange('0', '7')
}
func helper10_Char(p *Peg) bool {
	return p.InRange('0', '7')
}
func helper11_Char(p *Peg) bool {
	return p_Maybe(p, helper10_Char)
}
func helper12_Char(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper8_Char,
		helper9_Char,
		helper11_Char,
	})
}
func helper13_Char(p *Peg) bool {
	return p.Next("\\")
}
func helper14_Char(p *Peg) bool {
	return p_Not(p, helper13_Char)
}
func helper15_Char(p *Peg) bool {
	return p.AnyChar()
}
func helper16_Char(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper14_Char,
		helper15_Char,
	})
}
func helper17_Char(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		helper2_Char,
		helper7_Char,
		helper12_Char,
		helper16_Char,
	})
}
func p_Char(p *Peg) bool {
	// Char          <- '\\' [nrt'"\[\]\\]
	//                / '\\' [0-2][0-7][0-7]
	//                / '\\' [0-7][0-7]?
	//                / !'\\' .
	return p_addNode(p, helper17_Char, "Char")
}

func helper0_LEFTARROW(p *Peg) bool {
	return p.Next("<-")
}
func helper1_LEFTARROW(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_LEFTARROW,
		p_Spacing,
	})
}
func p_LEFTARROW(p *Peg) bool {
	// LEFTARROW     <- '<-' Spacing
	return p_addNode(p, helper1_LEFTARROW, "LEFTARROW")
}

func helper0_SLASH(p *Peg) bool {
	return p.Next("/")
}
func helper1_SLASH(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_SLASH,
		p_Spacing,
	})
}
func p_SLASH(p *Peg) bool {
	// SLASH         <- '/' Spacing
	return p_addNode(p, helper1_SLASH, "SLASH")
}

func helper0_AND(p *Peg) bool {
	return p.Next("&")
}
func helper1_AND(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_AND,
		p_Spacing,
	})
}
func p_AND(p *Peg) bool {
	// AND           <- '&' Spacing
	return p_addNode(p, helper1_AND, "AND")
}

func helper0_NOT(p *Peg) bool {
	return p.Next("!")
}
func helper1_NOT(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_NOT,
		p_Spacing,
	})
}
func p_NOT(p *Peg) bool {
	// NOT           <- '!' Spacing
	return p_addNode(p, helper1_NOT, "NOT")
}

func helper0_QUESTION(p *Peg) bool {
	return p.Next("?")
}
func helper1_QUESTION(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_QUESTION,
		p_Spacing,
	})
}
func p_QUESTION(p *Peg) bool {
	// QUESTION      <- '?' Spacing
	return p_addNode(p, helper1_QUESTION, "QUESTION")
}

func helper0_STAR(p *Peg) bool {
	return p.Next("*")
}
func helper1_STAR(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_STAR,
		p_Spacing,
	})
}
func p_STAR(p *Peg) bool {
	// STAR          <- '*' Spacing
	return p_addNode(p, helper1_STAR, "STAR")
}

func helper0_PLUS(p *Peg) bool {
	return p.Next("+")
}
func helper1_PLUS(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_PLUS,
		p_Spacing,
	})
}
func p_PLUS(p *Peg) bool {
	// PLUS          <- '+' Spacing
	return p_addNode(p, helper1_PLUS, "PLUS")
}

func helper0_OPEN(p *Peg) bool {
	return p.Next("(")
}
func helper1_OPEN(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_OPEN,
		p_Spacing,
	})
}
func p_OPEN(p *Peg) bool {
	// OPEN          <- '(' Spacing
	return p_addNode(p, helper1_OPEN, "OPEN")
}

func helper0_CLOSE(p *Peg) bool {
	return p.Next(")")
}
func helper1_CLOSE(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_CLOSE,
		p_Spacing,
	})
}
func p_CLOSE(p *Peg) bool {
	// CLOSE         <- ')' Spacing
	return p_addNode(p, helper1_CLOSE, "CLOSE")
}

func helper0_DOT(p *Peg) bool {
	return p.Next(".")
}
func helper1_DOT(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_DOT,
		p_Spacing,
	})
}
func p_DOT(p *Peg) bool {
	// DOT           <- '.' Spacing
	return p_addNode(p, helper1_DOT, "DOT")
}

func helper0_Spacing(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		p_Space,
		p_Comment,
	})
}
func helper1_Spacing(p *Peg) bool {
	return p_ZeroOrMore(p, helper0_Spacing)
}
func p_Spacing(p *Peg) bool {
	// Spacing       <- (Space / Comment)*
	return p_addNode(p, helper1_Spacing, "Spacing")
}

func helper0_Comment(p *Peg) bool {
	return p.Next("#")
}
func helper1_Comment(p *Peg) bool {
	return p_Not(p, p_EndOfLine)
}
func helper2_Comment(p *Peg) bool {
	return p.AnyChar()
}
func helper3_Comment(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper1_Comment,
		helper2_Comment,
	})
}
func helper4_Comment(p *Peg) bool {
	return p_ZeroOrMore(p, helper3_Comment)
}
func helper5_Comment(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_Comment,
		helper4_Comment,
		p_EndOfLine,
	})
}
func p_Comment(p *Peg) bool {
	// Comment       <- '#' (!EndOfLine .)* EndOfLine
	return p_addNode(p, helper5_Comment, "Comment")
}

func helper0_Space(p *Peg) bool {
	return p.Next(" ")
}
func helper1_Space(p *Peg) bool {
	return p.Next("\t")
}
func helper2_Space(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		helper0_Space,
		helper1_Space,
		p_EndOfLine,
	})
}
func p_Space(p *Peg) bool {
	// Space         <- ' ' / '\t' / EndOfLine
	return p_addNode(p, helper2_Space, "Space")
}

func helper0_EndOfLine(p *Peg) bool {
	return p.Next("\r\n")
}
func helper1_EndOfLine(p *Peg) bool {
	return p.Next("\n")
}
func helper2_EndOfLine(p *Peg) bool {
	return p.Next("\r")
}
func helper3_EndOfLine(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		helper0_EndOfLine,
		helper1_EndOfLine,
		helper2_EndOfLine,
	})
}
func p_EndOfLine(p *Peg) bool {
	// EndOfLine     <- '\r\n' / '\n' / '\r'
	return p_addNode(p, helper3_EndOfLine, "EndOfLine")
}

func helper0_EndOfFile(p *Peg) bool {
	return p.AnyChar()
}
func helper1_EndOfFile(p *Peg) bool {
	return p_Not(p, helper0_EndOfFile)
}
func p_EndOfFile(p *Peg) bool {
	// EndOfFile     <- !.
	return p_addNode(p, helper1_EndOfFile, "EndOfFile")
}
