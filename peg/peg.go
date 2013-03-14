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

package peg

import (
	"bytes"
	. "github.com/quarnster/parser"
)

type Peg struct {
	ParserData struct {
		Pos  int
		Data []byte
	}

	IgnoreRange Range
	Root        Node
	LastError   int
}

func (p *Peg) RootNode() *Node {
	return &p.Root
}

func (p *Peg) Parse(data string) bool {
	p.ParserData.Data = ([]byte)(data)
	p.ParserData.Pos = 0
	p.Root = Node{Name: "Peg", P: p}
	p.IgnoreRange = Range{}
	p.LastError = 0
	ret := p.realParse()
	if len(p.Root.Children) > 0 {
		p.Root.Range = Range{p.Root.Children[0].Range.Start, p.Root.Children[len(p.Root.Children)-1].Range.End}
	}
	return ret
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

func (p *Peg) Error() Error {
	errstr := ""

	line := 1
	column := 1
	for _, r := range p.ParserData.Data[:p.LastError] {
		column++
		if r == '\n' {
			line++
			column = 1
		}
	}

	if p.LastError == len(p.ParserData.Data) {
		errstr = "Unexpected EOF"
	} else {
		e := p.LastError + 4
		if e > len(p.ParserData.Data) {
			e = len(p.ParserData.Data)
		}

		reader := bytes.NewReader(p.ParserData.Data[p.LastError:e])
		r, _, _ := reader.ReadRune()
		if r == '\r' || r == '\n' {
			errstr = "Unexpected new line"
		} else {
			errstr = "Unexpected " + string(r)
		}
	}
	return NewError(line, column, errstr)
}

func (p *Peg) realParse() bool {
	return p_Grammar(p)
}
func p_Grammar(p *Peg) bool {
	// Grammar       <- Spacing Definition+ EndOfFile?
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		accept = p_Spacing(p)
		if accept {
			{
				save := p.ParserData.Pos
				accept = p_Definition(p)
				if !accept {
					p.ParserData.Pos = save
				} else {
					for accept {
						accept = p_Definition(p)
					}
					accept = true
				}
			}
			if accept {
				accept = p_EndOfFile(p)
				accept = true
				if accept {
				}
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos {
				p.LastError = p.ParserData.Pos
			}
			p.ParserData.Pos = save
		}
	}
	if accept && start != p.ParserData.Pos {
		if start < p.IgnoreRange.Start || p.IgnoreRange.Start == 0 {
			p.IgnoreRange.Start = start
		}
		p.IgnoreRange.End = p.ParserData.Pos
	}
	return accept
}

func p_Definition(p *Peg) bool {
	// Definition    <- Identifier LEFTARROW Expression
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		accept = p_Identifier(p)
		if accept {
			accept = p_LEFTARROW(p)
			if accept {
				accept = p_Expression(p)
				if accept {
				}
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos {
				p.LastError = p.ParserData.Pos
			}
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "Definition"
		node.P = p
		node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.Start >= end || p.IgnoreRange.End <= start {
		p.IgnoreRange = Range{}
	}
	return accept
}

func p_Expression(p *Peg) bool {
	// Expression    <- Sequence (SLASH Sequence)*
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		accept = p_Sequence(p)
		if accept {
			{
				accept = true
				for accept {
					{
						save := p.ParserData.Pos
						accept = p_SLASH(p)
						if accept {
							accept = p_Sequence(p)
							if accept {
							}
						}
						if !accept {
							if p.LastError < p.ParserData.Pos {
								p.LastError = p.ParserData.Pos
							}
							p.ParserData.Pos = save
						}
					}
				}
				accept = true
			}
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos {
				p.LastError = p.ParserData.Pos
			}
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "Expression"
		node.P = p
		node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.Start >= end || p.IgnoreRange.End <= start {
		p.IgnoreRange = Range{}
	}
	return accept
}

func p_Sequence(p *Peg) bool {
	// Sequence      <- Prefix+
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		accept = p_Prefix(p)
		if !accept {
			p.ParserData.Pos = save
		} else {
			for accept {
				accept = p_Prefix(p)
			}
			accept = true
		}
	}
	end := p.ParserData.Pos
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "Sequence"
		node.P = p
		node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.Start >= end || p.IgnoreRange.End <= start {
		p.IgnoreRange = Range{}
	}
	return accept
}

func p_Prefix(p *Peg) bool {
	// Prefix        <- (AND / NOT)? Suffix
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		{
			save := p.ParserData.Pos
			accept = p_AND(p)
			if !accept {
				accept = p_NOT(p)
				if !accept {
				}
			}
			if !accept {
				p.ParserData.Pos = save
			}
		}
		accept = true
		if accept {
			accept = p_Suffix(p)
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos {
				p.LastError = p.ParserData.Pos
			}
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "Prefix"
		node.P = p
		node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.Start >= end || p.IgnoreRange.End <= start {
		p.IgnoreRange = Range{}
	}
	return accept
}

func p_Suffix(p *Peg) bool {
	// Suffix        <- Primary (QUESTION / STAR / PLUS)?
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		accept = p_Primary(p)
		if accept {
			{
				save := p.ParserData.Pos
				accept = p_QUESTION(p)
				if !accept {
					accept = p_STAR(p)
					if !accept {
						accept = p_PLUS(p)
						if !accept {
						}
					}
				}
				if !accept {
					p.ParserData.Pos = save
				}
			}
			accept = true
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos {
				p.LastError = p.ParserData.Pos
			}
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "Suffix"
		node.P = p
		node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.Start >= end || p.IgnoreRange.End <= start {
		p.IgnoreRange = Range{}
	}
	return accept
}

func p_Primary(p *Peg) bool {
	// Primary       <- Identifier !LEFTARROW
	//                / OPEN Expression CLOSE
	//                / Literal / Class / DOT
	// # Lexical syntax
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		{
			save := p.ParserData.Pos
			accept = p_Identifier(p)
			if accept {
				s := p.ParserData.Pos
				accept = p_LEFTARROW(p)
				p.ParserData.Pos = s
				p.Root.Discard(s)
				accept = !accept
				if accept {
				}
			}
			if !accept {
				if p.LastError < p.ParserData.Pos {
					p.LastError = p.ParserData.Pos
				}
				p.ParserData.Pos = save
			}
		}
		if !accept {
			{
				save := p.ParserData.Pos
				accept = p_OPEN(p)
				if accept {
					accept = p_Expression(p)
					if accept {
						accept = p_CLOSE(p)
						if accept {
						}
					}
				}
				if !accept {
					if p.LastError < p.ParserData.Pos {
						p.LastError = p.ParserData.Pos
					}
					p.ParserData.Pos = save
				}
			}
			if !accept {
				accept = p_Literal(p)
				if !accept {
					accept = p_Class(p)
					if !accept {
						accept = p_DOT(p)
						if !accept {
						}
					}
				}
			}
		}
		if !accept {
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "Primary"
		node.P = p
		node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.Start >= end || p.IgnoreRange.End <= start {
		p.IgnoreRange = Range{}
	}
	return accept
}

func p_Identifier(p *Peg) bool {
	// Identifier    <- IdentStart IdentCont* Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		accept = p_IdentStart(p)
		if accept {
			{
				accept = true
				for accept {
					accept = p_IdentCont(p)
				}
				accept = true
			}
			if accept {
				accept = p_Spacing(p)
				if accept {
				}
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos {
				p.LastError = p.ParserData.Pos
			}
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "Identifier"
		node.P = p
		node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.Start >= end || p.IgnoreRange.End <= start {
		p.IgnoreRange = Range{}
	}
	return accept
}

func p_IdentStart(p *Peg) bool {
	// IdentStart    <- [a-zA-Z_]
	accept := false
	{
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) {
			accept = false
		} else {
			c := p.ParserData.Data[p.ParserData.Pos]
			if c >= 'a' && c <= 'z' {
				p.ParserData.Pos++
				accept = true
			} else {
				accept = false
			}
		}
		if !accept {
			if p.ParserData.Pos >= len(p.ParserData.Data) {
				accept = false
			} else {
				c := p.ParserData.Data[p.ParserData.Pos]
				if c >= 'A' && c <= 'Z' {
					p.ParserData.Pos++
					accept = true
				} else {
					accept = false
				}
			}
			if !accept {
				{
					accept = false
					if p.ParserData.Pos < len(p.ParserData.Data) {
						c := p.ParserData.Data[p.ParserData.Pos]
						if c == '_' {
							p.ParserData.Pos++
							accept = true
						}
					}
				}
				if !accept {
				}
			}
		}
		if !accept {
			p.ParserData.Pos = save
		}
	}
	return accept
}

func p_IdentCont(p *Peg) bool {
	// IdentCont     <- IdentStart / [0-9]
	accept := false
	{
		save := p.ParserData.Pos
		accept = p_IdentStart(p)
		if !accept {
			if p.ParserData.Pos >= len(p.ParserData.Data) {
				accept = false
			} else {
				c := p.ParserData.Data[p.ParserData.Pos]
				if c >= '0' && c <= '9' {
					p.ParserData.Pos++
					accept = true
				} else {
					accept = false
				}
			}
			if !accept {
			}
		}
		if !accept {
			p.ParserData.Pos = save
		}
	}
	return accept
}

func p_Literal(p *Peg) bool {
	// Literal       <- '\'' (!'\'' Char) '\'' Spacing
	//                / '"' (!'"' Char)+ '"' Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		{
			save := p.ParserData.Pos
			if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '\'' {
				accept = false
			} else {
				p.ParserData.Pos++
				accept = true
			}
			if accept {
				{
					save := p.ParserData.Pos
					s := p.ParserData.Pos
					if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '\'' {
						accept = false
					} else {
						p.ParserData.Pos++
						accept = true
					}
					p.ParserData.Pos = s
					p.Root.Discard(s)
					accept = !accept
					if accept {
						accept = p_Char(p)
						if accept {
						}
					}
					if !accept {
						if p.LastError < p.ParserData.Pos {
							p.LastError = p.ParserData.Pos
						}
						p.ParserData.Pos = save
					}
				}
				if accept {
					if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '\'' {
						accept = false
					} else {
						p.ParserData.Pos++
						accept = true
					}
					if accept {
						accept = p_Spacing(p)
						if accept {
						}
					}
				}
			}
			if !accept {
				if p.LastError < p.ParserData.Pos {
					p.LastError = p.ParserData.Pos
				}
				p.ParserData.Pos = save
			}
		}
		if !accept {
			{
				save := p.ParserData.Pos
				if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '"' {
					accept = false
				} else {
					p.ParserData.Pos++
					accept = true
				}
				if accept {
					{
						save := p.ParserData.Pos
						{
							save := p.ParserData.Pos
							s := p.ParserData.Pos
							if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '"' {
								accept = false
							} else {
								p.ParserData.Pos++
								accept = true
							}
							p.ParserData.Pos = s
							p.Root.Discard(s)
							accept = !accept
							if accept {
								accept = p_Char(p)
								if accept {
								}
							}
							if !accept {
								if p.LastError < p.ParserData.Pos {
									p.LastError = p.ParserData.Pos
								}
								p.ParserData.Pos = save
							}
						}
						if !accept {
							p.ParserData.Pos = save
						} else {
							for accept {
								{
									save := p.ParserData.Pos
									s := p.ParserData.Pos
									if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '"' {
										accept = false
									} else {
										p.ParserData.Pos++
										accept = true
									}
									p.ParserData.Pos = s
									p.Root.Discard(s)
									accept = !accept
									if accept {
										accept = p_Char(p)
										if accept {
										}
									}
									if !accept {
										if p.LastError < p.ParserData.Pos {
											p.LastError = p.ParserData.Pos
										}
										p.ParserData.Pos = save
									}
								}
							}
							accept = true
						}
					}
					if accept {
						if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '"' {
							accept = false
						} else {
							p.ParserData.Pos++
							accept = true
						}
						if accept {
							accept = p_Spacing(p)
							if accept {
							}
						}
					}
				}
				if !accept {
					if p.LastError < p.ParserData.Pos {
						p.LastError = p.ParserData.Pos
					}
					p.ParserData.Pos = save
				}
			}
			if !accept {
			}
		}
		if !accept {
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "Literal"
		node.P = p
		node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.Start >= end || p.IgnoreRange.End <= start {
		p.IgnoreRange = Range{}
	}
	return accept
}

func p_Class(p *Peg) bool {
	// Class         <- '[' (!']' Range)+ ']' Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '[' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		if accept {
			{
				save := p.ParserData.Pos
				{
					save := p.ParserData.Pos
					s := p.ParserData.Pos
					if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != ']' {
						accept = false
					} else {
						p.ParserData.Pos++
						accept = true
					}
					p.ParserData.Pos = s
					p.Root.Discard(s)
					accept = !accept
					if accept {
						accept = p_Range(p)
						if accept {
						}
					}
					if !accept {
						if p.LastError < p.ParserData.Pos {
							p.LastError = p.ParserData.Pos
						}
						p.ParserData.Pos = save
					}
				}
				if !accept {
					p.ParserData.Pos = save
				} else {
					for accept {
						{
							save := p.ParserData.Pos
							s := p.ParserData.Pos
							if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != ']' {
								accept = false
							} else {
								p.ParserData.Pos++
								accept = true
							}
							p.ParserData.Pos = s
							p.Root.Discard(s)
							accept = !accept
							if accept {
								accept = p_Range(p)
								if accept {
								}
							}
							if !accept {
								if p.LastError < p.ParserData.Pos {
									p.LastError = p.ParserData.Pos
								}
								p.ParserData.Pos = save
							}
						}
					}
					accept = true
				}
			}
			if accept {
				if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != ']' {
					accept = false
				} else {
					p.ParserData.Pos++
					accept = true
				}
				if accept {
					accept = p_Spacing(p)
					if accept {
					}
				}
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos {
				p.LastError = p.ParserData.Pos
			}
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "Class"
		node.P = p
		node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.Start >= end || p.IgnoreRange.End <= start {
		p.IgnoreRange = Range{}
	}
	return accept
}

func p_Range(p *Peg) bool {
	// Range         <- Char '-' Char / Char
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		{
			save := p.ParserData.Pos
			accept = p_Char(p)
			if accept {
				if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '-' {
					accept = false
				} else {
					p.ParserData.Pos++
					accept = true
				}
				if accept {
					accept = p_Char(p)
					if accept {
					}
				}
			}
			if !accept {
				if p.LastError < p.ParserData.Pos {
					p.LastError = p.ParserData.Pos
				}
				p.ParserData.Pos = save
			}
		}
		if !accept {
			accept = p_Char(p)
			if !accept {
			}
		}
		if !accept {
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "Range"
		node.P = p
		node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.Start >= end || p.IgnoreRange.End <= start {
		p.IgnoreRange = Range{}
	}
	return accept
}

func p_Char(p *Peg) bool {
	// Char          <- '\\' [nrt'"\[\]\\]
	//                / '\\' [0-2][0-7][0-7]
	//                / '\\' [0-7][0-7]?
	//                / !'\\' .
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		{
			save := p.ParserData.Pos
			if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '\\' {
				accept = false
			} else {
				p.ParserData.Pos++
				accept = true
			}
			if accept {
				{
					accept = false
					if p.ParserData.Pos < len(p.ParserData.Data) {
						c := p.ParserData.Data[p.ParserData.Pos]
						if c == 'n' || c == 'r' || c == 't' || c == '\'' || c == '"' || c == '[' || c == ']' || c == '\\' {
							p.ParserData.Pos++
							accept = true
						}
					}
				}
				if accept {
				}
			}
			if !accept {
				if p.LastError < p.ParserData.Pos {
					p.LastError = p.ParserData.Pos
				}
				p.ParserData.Pos = save
			}
		}
		if !accept {
			{
				save := p.ParserData.Pos
				if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '\\' {
					accept = false
				} else {
					p.ParserData.Pos++
					accept = true
				}
				if accept {
					if p.ParserData.Pos >= len(p.ParserData.Data) {
						accept = false
					} else {
						c := p.ParserData.Data[p.ParserData.Pos]
						if c >= '0' && c <= '2' {
							p.ParserData.Pos++
							accept = true
						} else {
							accept = false
						}
					}
					if accept {
						if p.ParserData.Pos >= len(p.ParserData.Data) {
							accept = false
						} else {
							c := p.ParserData.Data[p.ParserData.Pos]
							if c >= '0' && c <= '7' {
								p.ParserData.Pos++
								accept = true
							} else {
								accept = false
							}
						}
						if accept {
							if p.ParserData.Pos >= len(p.ParserData.Data) {
								accept = false
							} else {
								c := p.ParserData.Data[p.ParserData.Pos]
								if c >= '0' && c <= '7' {
									p.ParserData.Pos++
									accept = true
								} else {
									accept = false
								}
							}
							if accept {
							}
						}
					}
				}
				if !accept {
					if p.LastError < p.ParserData.Pos {
						p.LastError = p.ParserData.Pos
					}
					p.ParserData.Pos = save
				}
			}
			if !accept {
				{
					save := p.ParserData.Pos
					if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '\\' {
						accept = false
					} else {
						p.ParserData.Pos++
						accept = true
					}
					if accept {
						if p.ParserData.Pos >= len(p.ParserData.Data) {
							accept = false
						} else {
							c := p.ParserData.Data[p.ParserData.Pos]
							if c >= '0' && c <= '7' {
								p.ParserData.Pos++
								accept = true
							} else {
								accept = false
							}
						}
						if accept {
							if p.ParserData.Pos >= len(p.ParserData.Data) {
								accept = false
							} else {
								c := p.ParserData.Data[p.ParserData.Pos]
								if c >= '0' && c <= '7' {
									p.ParserData.Pos++
									accept = true
								} else {
									accept = false
								}
							}
							accept = true
							if accept {
							}
						}
					}
					if !accept {
						if p.LastError < p.ParserData.Pos {
							p.LastError = p.ParserData.Pos
						}
						p.ParserData.Pos = save
					}
				}
				if !accept {
					{
						save := p.ParserData.Pos
						s := p.ParserData.Pos
						if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '\\' {
							accept = false
						} else {
							p.ParserData.Pos++
							accept = true
						}
						p.ParserData.Pos = s
						p.Root.Discard(s)
						accept = !accept
						if accept {
							if p.ParserData.Pos >= len(p.ParserData.Data) {
								accept = false
							} else {
								p.ParserData.Pos++
								accept = true
							}
							if accept {
							}
						}
						if !accept {
							if p.LastError < p.ParserData.Pos {
								p.LastError = p.ParserData.Pos
							}
							p.ParserData.Pos = save
						}
					}
					if !accept {
					}
				}
			}
		}
		if !accept {
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "Char"
		node.P = p
		node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.Start >= end || p.IgnoreRange.End <= start {
		p.IgnoreRange = Range{}
	}
	return accept
}

func p_LEFTARROW(p *Peg) bool {
	// LEFTARROW     <- "<-" Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		{
			accept = true
			s := p.ParserData.Pos
			e := s + 2
			if e > len(p.ParserData.Data) {
				accept = false
			} else {
				if p.ParserData.Data[s+0] != '<' || p.ParserData.Data[s+1] != '-' {
					accept = false
				}
			}
			if accept {
				p.ParserData.Pos += 2
			}
		}
		if accept {
			accept = p_Spacing(p)
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos {
				p.LastError = p.ParserData.Pos
			}
			p.ParserData.Pos = save
		}
	}
	if accept && start != p.ParserData.Pos {
		if start < p.IgnoreRange.Start || p.IgnoreRange.Start == 0 {
			p.IgnoreRange.Start = start
		}
		p.IgnoreRange.End = p.ParserData.Pos
	}
	return accept
}

func p_SLASH(p *Peg) bool {
	// SLASH         <- '/' Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '/' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		if accept {
			accept = p_Spacing(p)
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos {
				p.LastError = p.ParserData.Pos
			}
			p.ParserData.Pos = save
		}
	}
	if accept && start != p.ParserData.Pos {
		if start < p.IgnoreRange.Start || p.IgnoreRange.Start == 0 {
			p.IgnoreRange.Start = start
		}
		p.IgnoreRange.End = p.ParserData.Pos
	}
	return accept
}

func p_AND(p *Peg) bool {
	// AND           <- '&' Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '&' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		if accept {
			accept = p_Spacing(p)
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos {
				p.LastError = p.ParserData.Pos
			}
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "AND"
		node.P = p
		node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.Start >= end || p.IgnoreRange.End <= start {
		p.IgnoreRange = Range{}
	}
	return accept
}

func p_NOT(p *Peg) bool {
	// NOT           <- '!' Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '!' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		if accept {
			accept = p_Spacing(p)
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos {
				p.LastError = p.ParserData.Pos
			}
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "NOT"
		node.P = p
		node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.Start >= end || p.IgnoreRange.End <= start {
		p.IgnoreRange = Range{}
	}
	return accept
}

func p_QUESTION(p *Peg) bool {
	// QUESTION      <- '?' Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '?' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		if accept {
			accept = p_Spacing(p)
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos {
				p.LastError = p.ParserData.Pos
			}
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "QUESTION"
		node.P = p
		node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.Start >= end || p.IgnoreRange.End <= start {
		p.IgnoreRange = Range{}
	}
	return accept
}

func p_STAR(p *Peg) bool {
	// STAR          <- '*' Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '*' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		if accept {
			accept = p_Spacing(p)
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos {
				p.LastError = p.ParserData.Pos
			}
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "STAR"
		node.P = p
		node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.Start >= end || p.IgnoreRange.End <= start {
		p.IgnoreRange = Range{}
	}
	return accept
}

func p_PLUS(p *Peg) bool {
	// PLUS          <- '+' Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '+' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		if accept {
			accept = p_Spacing(p)
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos {
				p.LastError = p.ParserData.Pos
			}
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "PLUS"
		node.P = p
		node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.Start >= end || p.IgnoreRange.End <= start {
		p.IgnoreRange = Range{}
	}
	return accept
}

func p_OPEN(p *Peg) bool {
	// OPEN          <- '(' Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '(' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		if accept {
			accept = p_Spacing(p)
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos {
				p.LastError = p.ParserData.Pos
			}
			p.ParserData.Pos = save
		}
	}
	if accept && start != p.ParserData.Pos {
		if start < p.IgnoreRange.Start || p.IgnoreRange.Start == 0 {
			p.IgnoreRange.Start = start
		}
		p.IgnoreRange.End = p.ParserData.Pos
	}
	return accept
}

func p_CLOSE(p *Peg) bool {
	// CLOSE         <- ')' Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != ')' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		if accept {
			accept = p_Spacing(p)
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos {
				p.LastError = p.ParserData.Pos
			}
			p.ParserData.Pos = save
		}
	}
	if accept && start != p.ParserData.Pos {
		if start < p.IgnoreRange.Start || p.IgnoreRange.Start == 0 {
			p.IgnoreRange.Start = start
		}
		p.IgnoreRange.End = p.ParserData.Pos
	}
	return accept
}

func p_DOT(p *Peg) bool {
	// DOT           <- '.' Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '.' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		if accept {
			accept = p_Spacing(p)
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos {
				p.LastError = p.ParserData.Pos
			}
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "DOT"
		node.P = p
		node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.Start >= end || p.IgnoreRange.End <= start {
		p.IgnoreRange = Range{}
	}
	return accept
}

func p_Spacing(p *Peg) bool {
	// Spacing       <- (Space / Comment)*
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		accept = true
		for accept {
			{
				save := p.ParserData.Pos
				accept = p_Space(p)
				if !accept {
					accept = p_Comment(p)
					if !accept {
					}
				}
				if !accept {
					p.ParserData.Pos = save
				}
			}
		}
		accept = true
	}
	if accept && start != p.ParserData.Pos {
		if start < p.IgnoreRange.Start || p.IgnoreRange.Start == 0 {
			p.IgnoreRange.Start = start
		}
		p.IgnoreRange.End = p.ParserData.Pos
	}
	return accept
}

func p_Comment(p *Peg) bool {
	// Comment       <- '#' (!EndOfLine .)* EndOfLine
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '#' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		if accept {
			{
				accept = true
				for accept {
					{
						save := p.ParserData.Pos
						s := p.ParserData.Pos
						accept = p_EndOfLine(p)
						p.ParserData.Pos = s
						p.Root.Discard(s)
						accept = !accept
						if accept {
							if p.ParserData.Pos >= len(p.ParserData.Data) {
								accept = false
							} else {
								p.ParserData.Pos++
								accept = true
							}
							if accept {
							}
						}
						if !accept {
							if p.LastError < p.ParserData.Pos {
								p.LastError = p.ParserData.Pos
							}
							p.ParserData.Pos = save
						}
					}
				}
				accept = true
			}
			if accept {
				accept = p_EndOfLine(p)
				if accept {
				}
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos {
				p.LastError = p.ParserData.Pos
			}
			p.ParserData.Pos = save
		}
	}
	if accept && start != p.ParserData.Pos {
		if start < p.IgnoreRange.Start || p.IgnoreRange.Start == 0 {
			p.IgnoreRange.Start = start
		}
		p.IgnoreRange.End = p.ParserData.Pos
	}
	return accept
}

func p_Space(p *Peg) bool {
	// Space         <- ' ' / '\t' / EndOfLine
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != ' ' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		if !accept {
			if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '\t' {
				accept = false
			} else {
				p.ParserData.Pos++
				accept = true
			}
			if !accept {
				accept = p_EndOfLine(p)
				if !accept {
				}
			}
		}
		if !accept {
			p.ParserData.Pos = save
		}
	}
	if accept && start != p.ParserData.Pos {
		if start < p.IgnoreRange.Start || p.IgnoreRange.Start == 0 {
			p.IgnoreRange.Start = start
		}
		p.IgnoreRange.End = p.ParserData.Pos
	}
	return accept
}

func p_EndOfLine(p *Peg) bool {
	// EndOfLine     <- "\r\n" / '\n' / '\r'
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		{
			accept = true
			s := p.ParserData.Pos
			e := s + 2
			if e > len(p.ParserData.Data) {
				accept = false
			} else {
				if p.ParserData.Data[s+0] != '\r' || p.ParserData.Data[s+1] != '\n' {
					accept = false
				}
			}
			if accept {
				p.ParserData.Pos += 2
			}
		}
		if !accept {
			if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '\n' {
				accept = false
			} else {
				p.ParserData.Pos++
				accept = true
			}
			if !accept {
				if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '\r' {
					accept = false
				} else {
					p.ParserData.Pos++
					accept = true
				}
				if !accept {
				}
			}
		}
		if !accept {
			p.ParserData.Pos = save
		}
	}
	if accept && start != p.ParserData.Pos {
		if start < p.IgnoreRange.Start || p.IgnoreRange.Start == 0 {
			p.IgnoreRange.Start = start
		}
		p.IgnoreRange.End = p.ParserData.Pos
	}
	return accept
}

func p_EndOfFile(p *Peg) bool {
	// EndOfFile     <- !.
	accept := false
	accept = true
	start := p.ParserData.Pos
	s := p.ParserData.Pos
	if p.ParserData.Pos >= len(p.ParserData.Data) {
		accept = false
	} else {
		p.ParserData.Pos++
		accept = true
	}
	p.ParserData.Pos = s
	p.Root.Discard(s)
	accept = !accept
	end := p.ParserData.Pos
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "EndOfFile"
		node.P = p
		node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.Start >= end || p.IgnoreRange.End <= start {
		p.IgnoreRange = Range{}
	}
	return accept
}
