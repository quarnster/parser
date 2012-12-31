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

func p_AnyChar(p *Peg) bool {
	return p.AnyChar()
	/*
	   return p.AnyChar()
	*/
}

func p_InRange(p *Peg, c1, c2 rune) bool {
	return p.InRange(c1, c2)
}

func p_InSet(p *Peg, dataset string) bool {
	return p.InSet(dataset)
}
func p_Next(p *Peg, n1 string) bool {
	return p.Next(n1)
}

func (p *Peg) Parse() bool {
	return p_Grammar(p)
}
func helper0_Grammar(p *Peg) bool {
	return p_Spacing(p)
}
func helper1_Grammar(p *Peg) bool {
	return p_Definition(p)
}
func helper2_Grammar(p *Peg) bool {
	return p_OneOrMore(p, helper1_Grammar)
}
func helper3_Grammar(p *Peg) bool {
	return p_EndOfFile(p)
}
func helper4_Grammar(p *Peg) bool {
	return p_Maybe(p, helper3_Grammar)
}
func helper5_Grammar(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_Grammar,
		helper2_Grammar,
		helper4_Grammar,
	})
}
func p_Grammar(p *Peg) bool {
	// Grammar       <- Spacing Definition+ EndOfFile?
	return p_addNode(p, helper5_Grammar, "Grammar")
}

func helper0_Definition(p *Peg) bool {
	return p_Identifier(p)
}
func helper1_Definition(p *Peg) bool {
	return p_LEFTARROW(p)
}
func helper2_Definition(p *Peg) bool {
	return p_Expression(p)
}
func helper3_Definition(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_Definition,
		helper1_Definition,
		helper2_Definition,
	})
}
func p_Definition(p *Peg) bool {
	// Definition    <- Identifier LEFTARROW Expression
	return p_addNode(p, helper3_Definition, "Definition")
}

func helper0_Expression(p *Peg) bool {
	return p_Sequence(p)
}
func helper1_Expression(p *Peg) bool {
	return p_SLASH(p)
}
func helper2_Expression(p *Peg) bool {
	return p_Sequence(p)
}
func helper3_Expression(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper1_Expression,
		helper2_Expression,
	})
}
func helper4_Expression(p *Peg) bool {
	return p_ZeroOrMore(p, helper3_Expression)
}
func helper5_Expression(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_Expression,
		helper4_Expression,
	})
}
func p_Expression(p *Peg) bool {
	// Expression    <- Sequence (SLASH Sequence)*
	return p_addNode(p, helper5_Expression, "Expression")
}

func helper0_Sequence(p *Peg) bool {
	return p_Prefix(p)
}
func helper1_Sequence(p *Peg) bool {
	return p_ZeroOrMore(p, helper0_Sequence)
}
func p_Sequence(p *Peg) bool {
	// Sequence      <- Prefix*
	return p_addNode(p, helper1_Sequence, "Sequence")
}

func helper0_Prefix(p *Peg) bool {
	return p_AND(p)
}
func helper1_Prefix(p *Peg) bool {
	return p_NOT(p)
}
func helper2_Prefix(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		helper0_Prefix,
		helper1_Prefix,
	})
}
func helper3_Prefix(p *Peg) bool {
	return p_Maybe(p, helper2_Prefix)
}
func helper4_Prefix(p *Peg) bool {
	return p_Suffix(p)
}
func helper5_Prefix(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper3_Prefix,
		helper4_Prefix,
	})
}
func p_Prefix(p *Peg) bool {
	// Prefix        <- (AND / NOT)? Suffix
	return p_addNode(p, helper5_Prefix, "Prefix")
}

func helper0_Suffix(p *Peg) bool {
	return p_Primary(p)
}
func helper1_Suffix(p *Peg) bool {
	return p_QUESTION(p)
}
func helper2_Suffix(p *Peg) bool {
	return p_STAR(p)
}
func helper3_Suffix(p *Peg) bool {
	return p_PLUS(p)
}
func helper4_Suffix(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		helper1_Suffix,
		helper2_Suffix,
		helper3_Suffix,
	})
}
func helper5_Suffix(p *Peg) bool {
	return p_Maybe(p, helper4_Suffix)
}
func helper6_Suffix(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_Suffix,
		helper5_Suffix,
	})
}
func p_Suffix(p *Peg) bool {
	// Suffix        <- Primary (QUESTION / STAR / PLUS)?
	return p_addNode(p, helper6_Suffix, "Suffix")
}

func helper0_Primary(p *Peg) bool {
	return p_Identifier(p)
}
func helper1_Primary(p *Peg) bool {
	return p_LEFTARROW(p)
}
func helper2_Primary(p *Peg) bool {
	return p_Not(p, helper1_Primary)
}
func helper3_Primary(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_Primary,
		helper2_Primary,
	})
}
func helper4_Primary(p *Peg) bool {
	return p_OPEN(p)
}
func helper5_Primary(p *Peg) bool {
	return p_Expression(p)
}
func helper6_Primary(p *Peg) bool {
	return p_CLOSE(p)
}
func helper7_Primary(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper4_Primary,
		helper5_Primary,
		helper6_Primary,
	})
}
func helper8_Primary(p *Peg) bool {
	return p_Literal(p)
}
func helper9_Primary(p *Peg) bool {
	return p_Class(p)
}
func helper10_Primary(p *Peg) bool {
	return p_DOT(p)
}
func helper11_Primary(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		helper3_Primary,
		helper7_Primary,
		helper8_Primary,
		helper9_Primary,
		helper10_Primary,
	})
}
func p_Primary(p *Peg) bool {
	// Primary       <- Identifier !LEFTARROW
	//                / OPEN Expression CLOSE
	//                / Literal / Class / DOT
	// # Lexical syntax
	return p_addNode(p, helper11_Primary, "Primary")
}

func helper0_Identifier(p *Peg) bool {
	return p_IdentStart(p)
}
func helper1_Identifier(p *Peg) bool {
	return p_IdentCont(p)
}
func helper2_Identifier(p *Peg) bool {
	return p_ZeroOrMore(p, helper1_Identifier)
}
func helper3_Identifier(p *Peg) bool {
	return p_Spacing(p)
}
func helper4_Identifier(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_Identifier,
		helper2_Identifier,
		helper3_Identifier,
	})
}
func p_Identifier(p *Peg) bool {
	// Identifier    <- IdentStart IdentCont* Spacing
	return p_addNode(p, helper4_Identifier, "Identifier")
}

func helper0_IdentStart(p *Peg) bool {
	return p_InRange(p, 'a', 'z')
}
func helper1_IdentStart(p *Peg) bool {
	return p_InRange(p, 'A', 'Z')
}
func helper2_IdentStart(p *Peg) bool {
	return p_InSet(p, "_")
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
	return p_IdentStart(p)
}
func helper1_IdentCont(p *Peg) bool {
	return p_InRange(p, '0', '9')
}
func helper2_IdentCont(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		helper0_IdentCont,
		helper1_IdentCont,
	})
}
func p_IdentCont(p *Peg) bool {
	// IdentCont     <- IdentStart / [0-9]
	return p_addNode(p, helper2_IdentCont, "IdentCont")
}

func helper0_Literal(p *Peg) bool {
	return p_InSet(p, "'")
}
func helper1_Literal(p *Peg) bool {
	return p_InSet(p, "'")
}
func helper2_Literal(p *Peg) bool {
	return p_Not(p, helper1_Literal)
}
func helper3_Literal(p *Peg) bool {
	return p_Char(p)
}
func helper4_Literal(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper2_Literal,
		helper3_Literal,
	})
}
func helper5_Literal(p *Peg) bool {
	return p_ZeroOrMore(p, helper4_Literal)
}
func helper6_Literal(p *Peg) bool {
	return p_InSet(p, "'")
}
func helper7_Literal(p *Peg) bool {
	return p_Spacing(p)
}
func helper8_Literal(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_Literal,
		helper5_Literal,
		helper6_Literal,
		helper7_Literal,
	})
}
func helper9_Literal(p *Peg) bool {
	return p_InSet(p, "\"")
}
func helper10_Literal(p *Peg) bool {
	return p_InSet(p, "\"")
}
func helper11_Literal(p *Peg) bool {
	return p_Not(p, helper10_Literal)
}
func helper12_Literal(p *Peg) bool {
	return p_Char(p)
}
func helper13_Literal(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper11_Literal,
		helper12_Literal,
	})
}
func helper14_Literal(p *Peg) bool {
	return p_ZeroOrMore(p, helper13_Literal)
}
func helper15_Literal(p *Peg) bool {
	return p_InSet(p, "\"")
}
func helper16_Literal(p *Peg) bool {
	return p_Spacing(p)
}
func helper17_Literal(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper9_Literal,
		helper14_Literal,
		helper15_Literal,
		helper16_Literal,
	})
}
func helper18_Literal(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		helper8_Literal,
		helper17_Literal,
	})
}
func p_Literal(p *Peg) bool {
	// Literal       <- ['] (!['] Char)* ['] Spacing
	//                / ["] (!["] Char)* ["] Spacing
	return p_addNode(p, helper18_Literal, "Literal")
}

func helper0_Class(p *Peg) bool {
	return p_Next(p, "[")
}
func helper1_Class(p *Peg) bool {
	return p_Next(p, "]")
}
func helper2_Class(p *Peg) bool {
	return p_Not(p, helper1_Class)
}
func helper3_Class(p *Peg) bool {
	return p_Range(p)
}
func helper4_Class(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper2_Class,
		helper3_Class,
	})
}
func helper5_Class(p *Peg) bool {
	return p_ZeroOrMore(p, helper4_Class)
}
func helper6_Class(p *Peg) bool {
	return p_Next(p, "]")
}
func helper7_Class(p *Peg) bool {
	return p_Spacing(p)
}
func helper8_Class(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_Class,
		helper5_Class,
		helper6_Class,
		helper7_Class,
	})
}
func p_Class(p *Peg) bool {
	// Class         <- '[' (!']' Range)* ']' Spacing
	return p_addNode(p, helper8_Class, "Class")
}

func helper0_Range(p *Peg) bool {
	return p_Char(p)
}
func helper1_Range(p *Peg) bool {
	return p_Next(p, "-")
}
func helper2_Range(p *Peg) bool {
	return p_Char(p)
}
func helper3_Range(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_Range,
		helper1_Range,
		helper2_Range,
	})
}
func helper4_Range(p *Peg) bool {
	return p_Char(p)
}
func helper5_Range(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		helper3_Range,
		helper4_Range,
	})
}
func p_Range(p *Peg) bool {
	// Range         <- Char '-' Char / Char
	return p_addNode(p, helper5_Range, "Range")
}

func helper0_Char(p *Peg) bool {
	return p_Next(p, "\\")
}
func helper1_Char(p *Peg) bool {
	return p_InSet(p, "nrt'\"[]\\")
}
func helper2_Char(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_Char,
		helper1_Char,
	})
}
func helper3_Char(p *Peg) bool {
	return p_Next(p, "\\")
}
func helper4_Char(p *Peg) bool {
	return p_InRange(p, '0', '2')
}
func helper5_Char(p *Peg) bool {
	return p_InRange(p, '0', '7')
}
func helper6_Char(p *Peg) bool {
	return p_InRange(p, '0', '7')
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
	return p_Next(p, "\\")
}
func helper9_Char(p *Peg) bool {
	return p_InRange(p, '0', '7')
}
func helper10_Char(p *Peg) bool {
	return p_InRange(p, '0', '7')
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
	return p_Next(p, "\\")
}
func helper14_Char(p *Peg) bool {
	return p_Not(p, helper13_Char)
}
func helper15_Char(p *Peg) bool {
	return p_AnyChar(p)
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
	return p_Next(p, "<-")
}
func helper1_LEFTARROW(p *Peg) bool {
	return p_Spacing(p)
}
func helper2_LEFTARROW(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_LEFTARROW,
		helper1_LEFTARROW,
	})
}
func p_LEFTARROW(p *Peg) bool {
	// LEFTARROW     <- '<-' Spacing
	return p_addNode(p, helper2_LEFTARROW, "LEFTARROW")
}

func helper0_SLASH(p *Peg) bool {
	return p_Next(p, "/")
}
func helper1_SLASH(p *Peg) bool {
	return p_Spacing(p)
}
func helper2_SLASH(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_SLASH,
		helper1_SLASH,
	})
}
func p_SLASH(p *Peg) bool {
	// SLASH         <- '/' Spacing
	return p_addNode(p, helper2_SLASH, "SLASH")
}

func helper0_AND(p *Peg) bool {
	return p_Next(p, "&")
}
func helper1_AND(p *Peg) bool {
	return p_Spacing(p)
}
func helper2_AND(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_AND,
		helper1_AND,
	})
}
func p_AND(p *Peg) bool {
	// AND           <- '&' Spacing
	return p_addNode(p, helper2_AND, "AND")
}

func helper0_NOT(p *Peg) bool {
	return p_Next(p, "!")
}
func helper1_NOT(p *Peg) bool {
	return p_Spacing(p)
}
func helper2_NOT(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_NOT,
		helper1_NOT,
	})
}
func p_NOT(p *Peg) bool {
	// NOT           <- '!' Spacing
	return p_addNode(p, helper2_NOT, "NOT")
}

func helper0_QUESTION(p *Peg) bool {
	return p_Next(p, "?")
}
func helper1_QUESTION(p *Peg) bool {
	return p_Spacing(p)
}
func helper2_QUESTION(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_QUESTION,
		helper1_QUESTION,
	})
}
func p_QUESTION(p *Peg) bool {
	// QUESTION      <- '?' Spacing
	return p_addNode(p, helper2_QUESTION, "QUESTION")
}

func helper0_STAR(p *Peg) bool {
	return p_Next(p, "*")
}
func helper1_STAR(p *Peg) bool {
	return p_Spacing(p)
}
func helper2_STAR(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_STAR,
		helper1_STAR,
	})
}
func p_STAR(p *Peg) bool {
	// STAR          <- '*' Spacing
	return p_addNode(p, helper2_STAR, "STAR")
}

func helper0_PLUS(p *Peg) bool {
	return p_Next(p, "+")
}
func helper1_PLUS(p *Peg) bool {
	return p_Spacing(p)
}
func helper2_PLUS(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_PLUS,
		helper1_PLUS,
	})
}
func p_PLUS(p *Peg) bool {
	// PLUS          <- '+' Spacing
	return p_addNode(p, helper2_PLUS, "PLUS")
}

func helper0_OPEN(p *Peg) bool {
	return p_Next(p, "(")
}
func helper1_OPEN(p *Peg) bool {
	return p_Spacing(p)
}
func helper2_OPEN(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_OPEN,
		helper1_OPEN,
	})
}
func p_OPEN(p *Peg) bool {
	// OPEN          <- '(' Spacing
	return p_addNode(p, helper2_OPEN, "OPEN")
}

func helper0_CLOSE(p *Peg) bool {
	return p_Next(p, ")")
}
func helper1_CLOSE(p *Peg) bool {
	return p_Spacing(p)
}
func helper2_CLOSE(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_CLOSE,
		helper1_CLOSE,
	})
}
func p_CLOSE(p *Peg) bool {
	// CLOSE         <- ')' Spacing
	return p_addNode(p, helper2_CLOSE, "CLOSE")
}

func helper0_DOT(p *Peg) bool {
	return p_Next(p, ".")
}
func helper1_DOT(p *Peg) bool {
	return p_Spacing(p)
}
func helper2_DOT(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_DOT,
		helper1_DOT,
	})
}
func p_DOT(p *Peg) bool {
	// DOT           <- '.' Spacing
	return p_addNode(p, helper2_DOT, "DOT")
}

func helper0_Spacing(p *Peg) bool {
	return p_Space(p)
}
func helper1_Spacing(p *Peg) bool {
	return p_Comment(p)
}
func helper2_Spacing(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		helper0_Spacing,
		helper1_Spacing,
	})
}
func helper3_Spacing(p *Peg) bool {
	return p_ZeroOrMore(p, helper2_Spacing)
}
func p_Spacing(p *Peg) bool {
	// Spacing       <- (Space / Comment)*
	return p_addNode(p, helper3_Spacing, "Spacing")
}

func helper0_Comment(p *Peg) bool {
	return p_Next(p, "#")
}
func helper1_Comment(p *Peg) bool {
	return p_EndOfLine(p)
}
func helper2_Comment(p *Peg) bool {
	return p_Not(p, helper1_Comment)
}
func helper3_Comment(p *Peg) bool {
	return p_AnyChar(p)
}
func helper4_Comment(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper2_Comment,
		helper3_Comment,
	})
}
func helper5_Comment(p *Peg) bool {
	return p_ZeroOrMore(p, helper4_Comment)
}
func helper6_Comment(p *Peg) bool {
	return p_EndOfLine(p)
}
func helper7_Comment(p *Peg) bool {
	return p_NeedAll(p, []func(*Peg) bool{
		helper0_Comment,
		helper5_Comment,
		helper6_Comment,
	})
}
func p_Comment(p *Peg) bool {
	// Comment       <- '#' (!EndOfLine .)* EndOfLine
	return p_addNode(p, helper7_Comment, "Comment")
}

func helper0_Space(p *Peg) bool {
	return p_Next(p, " ")
}
func helper1_Space(p *Peg) bool {
	return p_Next(p, "\t")
}
func helper2_Space(p *Peg) bool {
	return p_EndOfLine(p)
}
func helper3_Space(p *Peg) bool {
	return p_NeedOne(p, []func(*Peg) bool{
		helper0_Space,
		helper1_Space,
		helper2_Space,
	})
}
func p_Space(p *Peg) bool {
	// Space         <- ' ' / '\t' / EndOfLine
	return p_addNode(p, helper3_Space, "Space")
}

func helper0_EndOfLine(p *Peg) bool {
	return p_Next(p, "\r\n")
}
func helper1_EndOfLine(p *Peg) bool {
	return p_Next(p, "\n")
}
func helper2_EndOfLine(p *Peg) bool {
	return p_Next(p, "\r")
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
	return p_AnyChar(p)
}
func helper1_EndOfFile(p *Peg) bool {
	return p_Not(p, helper0_EndOfFile)
}
func p_EndOfFile(p *Peg) bool {
	// EndOfFile     <- !.
	return p_addNode(p, helper1_EndOfFile, "EndOfFile")
}
