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
	. "github.com/quarnster/parser"
	"github.com/quarnster/util/text"
)

type Peg struct {
	ParserData  Reader
	IgnoreRange text.Region
	Root        Node
	LastError   int
}

func (p *Peg) RootNode() *Node {
	return &p.Root
}

func (p *Peg) SetData(data string) {
	p.ParserData = NewReader(data)
	p.Root = Node{Name: "Peg", P: p}
	p.IgnoreRange = text.Region{}
	p.LastError = 0
}

func (p *Peg) Parse(data string) bool {
	p.SetData(data)
	ret := p.realParse()
	p.Root.UpdateRange()
	return ret
}

func (p *Peg) Data(start, end int) string {
	return p.ParserData.Substring(start, end)
}

func (p *Peg) Error() Error {
	errstr := ""
	line, column := p.ParserData.LineCol(p.LastError)

	if p.LastError == p.ParserData.Len() {
		errstr = "Unexpected EOF"
	} else {
		p.ParserData.Seek(p.LastError)
		if r := p.ParserData.Read(); r == '\r' || r == '\n' {
			errstr = "Unexpected new line"
		} else {
			errstr = "Unexpected " + string(r)
		}
	}
	return NewError(line, column, errstr)
}

func (p *Peg) realParse() bool {
	return p.Grammar()
}
func (p *Peg) Grammar() bool {
	// Grammar       <- Spacing Definition+ EndOfFile?
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		accept = p.Spacing()
		if accept {
			{
				save := p.ParserData.Pos()
				accept = p.Definition()
				if !accept {
					p.ParserData.Seek(save)
				} else {
					for accept {
						accept = p.Definition()
					}
					accept = true
				}
			}
			if accept {
				accept = p.EndOfFile()
				accept = true
				if accept {
				}
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos() {
				p.LastError = p.ParserData.Pos()
			}
			p.ParserData.Seek(save)
		}
	}
	if accept && start != p.ParserData.Pos() {
		if start < p.IgnoreRange.A || p.IgnoreRange.A == 0 {
			p.IgnoreRange.A = start
		}
		p.IgnoreRange.B = p.ParserData.Pos()
	}
	return accept
}

func (p *Peg) Definition() bool {
	// Definition    <- Identifier LEFTARROW Expression
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		accept = p.Identifier()
		if accept {
			accept = p.LEFTARROW()
			if accept {
				accept = p.Expression()
				if accept {
				}
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos() {
				p.LastError = p.ParserData.Pos()
			}
			p.ParserData.Seek(save)
		}
	}
	end := p.ParserData.Pos()
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "Definition"
		node.P = p
		node.Range = node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.A >= end || p.IgnoreRange.B <= start {
		p.IgnoreRange = text.Region{}
	}
	return accept
}

func (p *Peg) Expression() bool {
	// Expression    <- Sequence (SLASH Sequence)*
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		accept = p.Sequence()
		if accept {
			{
				accept = true
				for accept {
					{
						save := p.ParserData.Pos()
						accept = p.SLASH()
						if accept {
							accept = p.Sequence()
							if accept {
							}
						}
						if !accept {
							if p.LastError < p.ParserData.Pos() {
								p.LastError = p.ParserData.Pos()
							}
							p.ParserData.Seek(save)
						}
					}
				}
				accept = true
			}
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos() {
				p.LastError = p.ParserData.Pos()
			}
			p.ParserData.Seek(save)
		}
	}
	end := p.ParserData.Pos()
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "Expression"
		node.P = p
		node.Range = node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.A >= end || p.IgnoreRange.B <= start {
		p.IgnoreRange = text.Region{}
	}
	return accept
}

func (p *Peg) Sequence() bool {
	// Sequence      <- Prefix+
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		accept = p.Prefix()
		if !accept {
			p.ParserData.Seek(save)
		} else {
			for accept {
				accept = p.Prefix()
			}
			accept = true
		}
	}
	end := p.ParserData.Pos()
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "Sequence"
		node.P = p
		node.Range = node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.A >= end || p.IgnoreRange.B <= start {
		p.IgnoreRange = text.Region{}
	}
	return accept
}

func (p *Peg) Prefix() bool {
	// Prefix        <- (AND / NOT)? Suffix
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		{
			save := p.ParserData.Pos()
			accept = p.AND()
			if !accept {
				accept = p.NOT()
				if !accept {
				}
			}
			if !accept {
				p.ParserData.Seek(save)
			}
		}
		accept = true
		if accept {
			accept = p.Suffix()
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos() {
				p.LastError = p.ParserData.Pos()
			}
			p.ParserData.Seek(save)
		}
	}
	end := p.ParserData.Pos()
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "Prefix"
		node.P = p
		node.Range = node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.A >= end || p.IgnoreRange.B <= start {
		p.IgnoreRange = text.Region{}
	}
	return accept
}

func (p *Peg) Suffix() bool {
	// Suffix        <- Primary (QUESTION / STAR / PLUS)?
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		accept = p.Primary()
		if accept {
			{
				save := p.ParserData.Pos()
				accept = p.QUESTION()
				if !accept {
					accept = p.STAR()
					if !accept {
						accept = p.PLUS()
						if !accept {
						}
					}
				}
				if !accept {
					p.ParserData.Seek(save)
				}
			}
			accept = true
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos() {
				p.LastError = p.ParserData.Pos()
			}
			p.ParserData.Seek(save)
		}
	}
	end := p.ParserData.Pos()
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "Suffix"
		node.P = p
		node.Range = node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.A >= end || p.IgnoreRange.B <= start {
		p.IgnoreRange = text.Region{}
	}
	return accept
}

func (p *Peg) Primary() bool {
	// Primary       <- Identifier !LEFTARROW
	//                / OPEN Expression CLOSE
	//                / Literal / Class / DOT
	// # Lexical syntax
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		{
			save := p.ParserData.Pos()
			accept = p.Identifier()
			if accept {
				s := p.ParserData.Pos()
				accept = p.LEFTARROW()
				p.ParserData.Seek(s)
				p.Root.Discard(s)
				accept = !accept
				if accept {
				}
			}
			if !accept {
				if p.LastError < p.ParserData.Pos() {
					p.LastError = p.ParserData.Pos()
				}
				p.ParserData.Seek(save)
			}
		}
		if !accept {
			{
				save := p.ParserData.Pos()
				accept = p.OPEN()
				if accept {
					accept = p.Expression()
					if accept {
						accept = p.CLOSE()
						if accept {
						}
					}
				}
				if !accept {
					if p.LastError < p.ParserData.Pos() {
						p.LastError = p.ParserData.Pos()
					}
					p.ParserData.Seek(save)
				}
			}
			if !accept {
				accept = p.Literal()
				if !accept {
					accept = p.Class()
					if !accept {
						accept = p.DOT()
						if !accept {
						}
					}
				}
			}
		}
		if !accept {
			p.ParserData.Seek(save)
		}
	}
	end := p.ParserData.Pos()
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "Primary"
		node.P = p
		node.Range = node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.A >= end || p.IgnoreRange.B <= start {
		p.IgnoreRange = text.Region{}
	}
	return accept
}

func (p *Peg) Identifier() bool {
	// Identifier    <- IdentStart IdentCont* Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		accept = p.IdentStart()
		if accept {
			{
				accept = true
				for accept {
					accept = p.IdentCont()
				}
				accept = true
			}
			if accept {
				accept = p.Spacing()
				if accept {
				}
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos() {
				p.LastError = p.ParserData.Pos()
			}
			p.ParserData.Seek(save)
		}
	}
	end := p.ParserData.Pos()
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "Identifier"
		node.P = p
		node.Range = node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.A >= end || p.IgnoreRange.B <= start {
		p.IgnoreRange = text.Region{}
	}
	return accept
}

func (p *Peg) IdentStart() bool {
	// IdentStart    <- [a-zA-Z_]
	accept := false
	{
		save := p.ParserData.Pos()
		c := p.ParserData.Read()
		if c >= 'a' && c <= 'z' {
			accept = true
		} else {
			p.ParserData.UnRead()
			accept = false
		}
		if !accept {
			c := p.ParserData.Read()
			if c >= 'A' && c <= 'Z' {
				accept = true
			} else {
				p.ParserData.UnRead()
				accept = false
			}
			if !accept {
				{
					accept = false
					c := p.ParserData.Read()
					if c == '_' {
						accept = true
					} else {
						p.ParserData.UnRead()
					}
				}
				if !accept {
				}
			}
		}
		if !accept {
			p.ParserData.Seek(save)
		}
	}
	return accept
}

func (p *Peg) IdentCont() bool {
	// IdentCont     <- IdentStart / [0-9]
	accept := false
	{
		save := p.ParserData.Pos()
		accept = p.IdentStart()
		if !accept {
			c := p.ParserData.Read()
			if c >= '0' && c <= '9' {
				accept = true
			} else {
				p.ParserData.UnRead()
				accept = false
			}
			if !accept {
			}
		}
		if !accept {
			p.ParserData.Seek(save)
		}
	}
	return accept
}

func (p *Peg) Literal() bool {
	// Literal       <- '\'' (!'\'' Char) '\'' Spacing
	//                / '"' (!'"' Char)+ '"' Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		{
			save := p.ParserData.Pos()
			if p.ParserData.Read() != '\'' {
				p.ParserData.UnRead()
				accept = false
			} else {
				accept = true
			}
			if accept {
				{
					save := p.ParserData.Pos()
					s := p.ParserData.Pos()
					if p.ParserData.Read() != '\'' {
						p.ParserData.UnRead()
						accept = false
					} else {
						accept = true
					}
					p.ParserData.Seek(s)
					p.Root.Discard(s)
					accept = !accept
					if accept {
						accept = p.Char()
						if accept {
						}
					}
					if !accept {
						if p.LastError < p.ParserData.Pos() {
							p.LastError = p.ParserData.Pos()
						}
						p.ParserData.Seek(save)
					}
				}
				if accept {
					if p.ParserData.Read() != '\'' {
						p.ParserData.UnRead()
						accept = false
					} else {
						accept = true
					}
					if accept {
						accept = p.Spacing()
						if accept {
						}
					}
				}
			}
			if !accept {
				if p.LastError < p.ParserData.Pos() {
					p.LastError = p.ParserData.Pos()
				}
				p.ParserData.Seek(save)
			}
		}
		if !accept {
			{
				save := p.ParserData.Pos()
				if p.ParserData.Read() != '"' {
					p.ParserData.UnRead()
					accept = false
				} else {
					accept = true
				}
				if accept {
					{
						save := p.ParserData.Pos()
						{
							save := p.ParserData.Pos()
							s := p.ParserData.Pos()
							if p.ParserData.Read() != '"' {
								p.ParserData.UnRead()
								accept = false
							} else {
								accept = true
							}
							p.ParserData.Seek(s)
							p.Root.Discard(s)
							accept = !accept
							if accept {
								accept = p.Char()
								if accept {
								}
							}
							if !accept {
								if p.LastError < p.ParserData.Pos() {
									p.LastError = p.ParserData.Pos()
								}
								p.ParserData.Seek(save)
							}
						}
						if !accept {
							p.ParserData.Seek(save)
						} else {
							for accept {
								{
									save := p.ParserData.Pos()
									s := p.ParserData.Pos()
									if p.ParserData.Read() != '"' {
										p.ParserData.UnRead()
										accept = false
									} else {
										accept = true
									}
									p.ParserData.Seek(s)
									p.Root.Discard(s)
									accept = !accept
									if accept {
										accept = p.Char()
										if accept {
										}
									}
									if !accept {
										if p.LastError < p.ParserData.Pos() {
											p.LastError = p.ParserData.Pos()
										}
										p.ParserData.Seek(save)
									}
								}
							}
							accept = true
						}
					}
					if accept {
						if p.ParserData.Read() != '"' {
							p.ParserData.UnRead()
							accept = false
						} else {
							accept = true
						}
						if accept {
							accept = p.Spacing()
							if accept {
							}
						}
					}
				}
				if !accept {
					if p.LastError < p.ParserData.Pos() {
						p.LastError = p.ParserData.Pos()
					}
					p.ParserData.Seek(save)
				}
			}
			if !accept {
			}
		}
		if !accept {
			p.ParserData.Seek(save)
		}
	}
	end := p.ParserData.Pos()
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "Literal"
		node.P = p
		node.Range = node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.A >= end || p.IgnoreRange.B <= start {
		p.IgnoreRange = text.Region{}
	}
	return accept
}

func (p *Peg) Class() bool {
	// Class         <- '[' (!']' Range)+ ']' Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		if p.ParserData.Read() != '[' {
			p.ParserData.UnRead()
			accept = false
		} else {
			accept = true
		}
		if accept {
			{
				save := p.ParserData.Pos()
				{
					save := p.ParserData.Pos()
					s := p.ParserData.Pos()
					if p.ParserData.Read() != ']' {
						p.ParserData.UnRead()
						accept = false
					} else {
						accept = true
					}
					p.ParserData.Seek(s)
					p.Root.Discard(s)
					accept = !accept
					if accept {
						accept = p.Range()
						if accept {
						}
					}
					if !accept {
						if p.LastError < p.ParserData.Pos() {
							p.LastError = p.ParserData.Pos()
						}
						p.ParserData.Seek(save)
					}
				}
				if !accept {
					p.ParserData.Seek(save)
				} else {
					for accept {
						{
							save := p.ParserData.Pos()
							s := p.ParserData.Pos()
							if p.ParserData.Read() != ']' {
								p.ParserData.UnRead()
								accept = false
							} else {
								accept = true
							}
							p.ParserData.Seek(s)
							p.Root.Discard(s)
							accept = !accept
							if accept {
								accept = p.Range()
								if accept {
								}
							}
							if !accept {
								if p.LastError < p.ParserData.Pos() {
									p.LastError = p.ParserData.Pos()
								}
								p.ParserData.Seek(save)
							}
						}
					}
					accept = true
				}
			}
			if accept {
				if p.ParserData.Read() != ']' {
					p.ParserData.UnRead()
					accept = false
				} else {
					accept = true
				}
				if accept {
					accept = p.Spacing()
					if accept {
					}
				}
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos() {
				p.LastError = p.ParserData.Pos()
			}
			p.ParserData.Seek(save)
		}
	}
	end := p.ParserData.Pos()
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "Class"
		node.P = p
		node.Range = node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.A >= end || p.IgnoreRange.B <= start {
		p.IgnoreRange = text.Region{}
	}
	return accept
}

func (p *Peg) Range() bool {
	// Range         <- Char '-' Char / Char
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		{
			save := p.ParserData.Pos()
			accept = p.Char()
			if accept {
				if p.ParserData.Read() != '-' {
					p.ParserData.UnRead()
					accept = false
				} else {
					accept = true
				}
				if accept {
					accept = p.Char()
					if accept {
					}
				}
			}
			if !accept {
				if p.LastError < p.ParserData.Pos() {
					p.LastError = p.ParserData.Pos()
				}
				p.ParserData.Seek(save)
			}
		}
		if !accept {
			accept = p.Char()
			if !accept {
			}
		}
		if !accept {
			p.ParserData.Seek(save)
		}
	}
	end := p.ParserData.Pos()
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "Range"
		node.P = p
		node.Range = node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.A >= end || p.IgnoreRange.B <= start {
		p.IgnoreRange = text.Region{}
	}
	return accept
}

func (p *Peg) Char() bool {
	// Char          <- '\\' [nrt'"\[\]\\]
	//                / '\\' [0-2][0-7][0-7]
	//                / '\\' [0-7][0-7]?
	//                / "\\u" Hex Hex Hex Hex
	//                / "\\U" Hex Hex Hex Hex Hex Hex Hex Hex
	//                / !'\\' .
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		{
			save := p.ParserData.Pos()
			if p.ParserData.Read() != '\\' {
				p.ParserData.UnRead()
				accept = false
			} else {
				accept = true
			}
			if accept {
				{
					accept = false
					c := p.ParserData.Read()
					if c == 'n' || c == 'r' || c == 't' || c == '\'' || c == '"' || c == '[' || c == ']' || c == '\\' {
						accept = true
					} else {
						p.ParserData.UnRead()
					}
				}
				if accept {
				}
			}
			if !accept {
				if p.LastError < p.ParserData.Pos() {
					p.LastError = p.ParserData.Pos()
				}
				p.ParserData.Seek(save)
			}
		}
		if !accept {
			{
				save := p.ParserData.Pos()
				if p.ParserData.Read() != '\\' {
					p.ParserData.UnRead()
					accept = false
				} else {
					accept = true
				}
				if accept {
					c := p.ParserData.Read()
					if c >= '0' && c <= '2' {
						accept = true
					} else {
						p.ParserData.UnRead()
						accept = false
					}
					if accept {
						c := p.ParserData.Read()
						if c >= '0' && c <= '7' {
							accept = true
						} else {
							p.ParserData.UnRead()
							accept = false
						}
						if accept {
							c := p.ParserData.Read()
							if c >= '0' && c <= '7' {
								accept = true
							} else {
								p.ParserData.UnRead()
								accept = false
							}
							if accept {
							}
						}
					}
				}
				if !accept {
					if p.LastError < p.ParserData.Pos() {
						p.LastError = p.ParserData.Pos()
					}
					p.ParserData.Seek(save)
				}
			}
			if !accept {
				{
					save := p.ParserData.Pos()
					if p.ParserData.Read() != '\\' {
						p.ParserData.UnRead()
						accept = false
					} else {
						accept = true
					}
					if accept {
						c := p.ParserData.Read()
						if c >= '0' && c <= '7' {
							accept = true
						} else {
							p.ParserData.UnRead()
							accept = false
						}
						if accept {
							c := p.ParserData.Read()
							if c >= '0' && c <= '7' {
								accept = true
							} else {
								p.ParserData.UnRead()
								accept = false
							}
							accept = true
							if accept {
							}
						}
					}
					if !accept {
						if p.LastError < p.ParserData.Pos() {
							p.LastError = p.ParserData.Pos()
						}
						p.ParserData.Seek(save)
					}
				}
				if !accept {
					{
						save := p.ParserData.Pos()
						{
							accept = true
							s := p.ParserData.Pos()
							if p.ParserData.Read() != '\\' || p.ParserData.Read() != 'u' {
								p.ParserData.Seek(s)
								accept = false
							}
						}
						if accept {
							accept = p.Hex()
							if accept {
								accept = p.Hex()
								if accept {
									accept = p.Hex()
									if accept {
										accept = p.Hex()
										if accept {
										}
									}
								}
							}
						}
						if !accept {
							if p.LastError < p.ParserData.Pos() {
								p.LastError = p.ParserData.Pos()
							}
							p.ParserData.Seek(save)
						}
					}
					if !accept {
						{
							save := p.ParserData.Pos()
							{
								accept = true
								s := p.ParserData.Pos()
								if p.ParserData.Read() != '\\' || p.ParserData.Read() != 'U' {
									p.ParserData.Seek(s)
									accept = false
								}
							}
							if accept {
								accept = p.Hex()
								if accept {
									accept = p.Hex()
									if accept {
										accept = p.Hex()
										if accept {
											accept = p.Hex()
											if accept {
												accept = p.Hex()
												if accept {
													accept = p.Hex()
													if accept {
														accept = p.Hex()
														if accept {
															accept = p.Hex()
															if accept {
															}
														}
													}
												}
											}
										}
									}
								}
							}
							if !accept {
								if p.LastError < p.ParserData.Pos() {
									p.LastError = p.ParserData.Pos()
								}
								p.ParserData.Seek(save)
							}
						}
						if !accept {
							{
								save := p.ParserData.Pos()
								s := p.ParserData.Pos()
								if p.ParserData.Read() != '\\' {
									p.ParserData.UnRead()
									accept = false
								} else {
									accept = true
								}
								p.ParserData.Seek(s)
								p.Root.Discard(s)
								accept = !accept
								if accept {
									if p.ParserData.Pos() >= p.ParserData.Len() {
										accept = false
									} else {
										p.ParserData.Read()
										accept = true
									}
									if accept {
									}
								}
								if !accept {
									if p.LastError < p.ParserData.Pos() {
										p.LastError = p.ParserData.Pos()
									}
									p.ParserData.Seek(save)
								}
							}
							if !accept {
							}
						}
					}
				}
			}
		}
		if !accept {
			p.ParserData.Seek(save)
		}
	}
	end := p.ParserData.Pos()
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "Char"
		node.P = p
		node.Range = node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.A >= end || p.IgnoreRange.B <= start {
		p.IgnoreRange = text.Region{}
	}
	return accept
}

func (p *Peg) Hex() bool {
	// Hex           <- [A-Fa-f0-9]
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		c := p.ParserData.Read()
		if c >= 'A' && c <= 'F' {
			accept = true
		} else {
			p.ParserData.UnRead()
			accept = false
		}
		if !accept {
			c := p.ParserData.Read()
			if c >= 'a' && c <= 'f' {
				accept = true
			} else {
				p.ParserData.UnRead()
				accept = false
			}
			if !accept {
				c := p.ParserData.Read()
				if c >= '0' && c <= '9' {
					accept = true
				} else {
					p.ParserData.UnRead()
					accept = false
				}
				if !accept {
				}
			}
		}
		if !accept {
			p.ParserData.Seek(save)
		}
	}
	end := p.ParserData.Pos()
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "Hex"
		node.P = p
		node.Range = node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.A >= end || p.IgnoreRange.B <= start {
		p.IgnoreRange = text.Region{}
	}
	return accept
}

func (p *Peg) LEFTARROW() bool {
	// LEFTARROW     <- "<-" Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		{
			accept = true
			s := p.ParserData.Pos()
			if p.ParserData.Read() != '<' || p.ParserData.Read() != '-' {
				p.ParserData.Seek(s)
				accept = false
			}
		}
		if accept {
			accept = p.Spacing()
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos() {
				p.LastError = p.ParserData.Pos()
			}
			p.ParserData.Seek(save)
		}
	}
	if accept && start != p.ParserData.Pos() {
		if start < p.IgnoreRange.A || p.IgnoreRange.A == 0 {
			p.IgnoreRange.A = start
		}
		p.IgnoreRange.B = p.ParserData.Pos()
	}
	return accept
}

func (p *Peg) SLASH() bool {
	// SLASH         <- '/' Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		if p.ParserData.Read() != '/' {
			p.ParserData.UnRead()
			accept = false
		} else {
			accept = true
		}
		if accept {
			accept = p.Spacing()
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos() {
				p.LastError = p.ParserData.Pos()
			}
			p.ParserData.Seek(save)
		}
	}
	if accept && start != p.ParserData.Pos() {
		if start < p.IgnoreRange.A || p.IgnoreRange.A == 0 {
			p.IgnoreRange.A = start
		}
		p.IgnoreRange.B = p.ParserData.Pos()
	}
	return accept
}

func (p *Peg) AND() bool {
	// AND           <- '&' Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		if p.ParserData.Read() != '&' {
			p.ParserData.UnRead()
			accept = false
		} else {
			accept = true
		}
		if accept {
			accept = p.Spacing()
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos() {
				p.LastError = p.ParserData.Pos()
			}
			p.ParserData.Seek(save)
		}
	}
	end := p.ParserData.Pos()
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "AND"
		node.P = p
		node.Range = node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.A >= end || p.IgnoreRange.B <= start {
		p.IgnoreRange = text.Region{}
	}
	return accept
}

func (p *Peg) NOT() bool {
	// NOT           <- '!' Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		if p.ParserData.Read() != '!' {
			p.ParserData.UnRead()
			accept = false
		} else {
			accept = true
		}
		if accept {
			accept = p.Spacing()
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos() {
				p.LastError = p.ParserData.Pos()
			}
			p.ParserData.Seek(save)
		}
	}
	end := p.ParserData.Pos()
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "NOT"
		node.P = p
		node.Range = node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.A >= end || p.IgnoreRange.B <= start {
		p.IgnoreRange = text.Region{}
	}
	return accept
}

func (p *Peg) QUESTION() bool {
	// QUESTION      <- '?' Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		if p.ParserData.Read() != '?' {
			p.ParserData.UnRead()
			accept = false
		} else {
			accept = true
		}
		if accept {
			accept = p.Spacing()
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos() {
				p.LastError = p.ParserData.Pos()
			}
			p.ParserData.Seek(save)
		}
	}
	end := p.ParserData.Pos()
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "QUESTION"
		node.P = p
		node.Range = node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.A >= end || p.IgnoreRange.B <= start {
		p.IgnoreRange = text.Region{}
	}
	return accept
}

func (p *Peg) STAR() bool {
	// STAR          <- '*' Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		if p.ParserData.Read() != '*' {
			p.ParserData.UnRead()
			accept = false
		} else {
			accept = true
		}
		if accept {
			accept = p.Spacing()
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos() {
				p.LastError = p.ParserData.Pos()
			}
			p.ParserData.Seek(save)
		}
	}
	end := p.ParserData.Pos()
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "STAR"
		node.P = p
		node.Range = node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.A >= end || p.IgnoreRange.B <= start {
		p.IgnoreRange = text.Region{}
	}
	return accept
}

func (p *Peg) PLUS() bool {
	// PLUS          <- '+' Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		if p.ParserData.Read() != '+' {
			p.ParserData.UnRead()
			accept = false
		} else {
			accept = true
		}
		if accept {
			accept = p.Spacing()
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos() {
				p.LastError = p.ParserData.Pos()
			}
			p.ParserData.Seek(save)
		}
	}
	end := p.ParserData.Pos()
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "PLUS"
		node.P = p
		node.Range = node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.A >= end || p.IgnoreRange.B <= start {
		p.IgnoreRange = text.Region{}
	}
	return accept
}

func (p *Peg) OPEN() bool {
	// OPEN          <- '(' Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		if p.ParserData.Read() != '(' {
			p.ParserData.UnRead()
			accept = false
		} else {
			accept = true
		}
		if accept {
			accept = p.Spacing()
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos() {
				p.LastError = p.ParserData.Pos()
			}
			p.ParserData.Seek(save)
		}
	}
	if accept && start != p.ParserData.Pos() {
		if start < p.IgnoreRange.A || p.IgnoreRange.A == 0 {
			p.IgnoreRange.A = start
		}
		p.IgnoreRange.B = p.ParserData.Pos()
	}
	return accept
}

func (p *Peg) CLOSE() bool {
	// CLOSE         <- ')' Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		if p.ParserData.Read() != ')' {
			p.ParserData.UnRead()
			accept = false
		} else {
			accept = true
		}
		if accept {
			accept = p.Spacing()
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos() {
				p.LastError = p.ParserData.Pos()
			}
			p.ParserData.Seek(save)
		}
	}
	if accept && start != p.ParserData.Pos() {
		if start < p.IgnoreRange.A || p.IgnoreRange.A == 0 {
			p.IgnoreRange.A = start
		}
		p.IgnoreRange.B = p.ParserData.Pos()
	}
	return accept
}

func (p *Peg) DOT() bool {
	// DOT           <- '.' Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		if p.ParserData.Read() != '.' {
			p.ParserData.UnRead()
			accept = false
		} else {
			accept = true
		}
		if accept {
			accept = p.Spacing()
			if accept {
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos() {
				p.LastError = p.ParserData.Pos()
			}
			p.ParserData.Seek(save)
		}
	}
	end := p.ParserData.Pos()
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "DOT"
		node.P = p
		node.Range = node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.A >= end || p.IgnoreRange.B <= start {
		p.IgnoreRange = text.Region{}
	}
	return accept
}

func (p *Peg) Spacing() bool {
	// Spacing       <- (Space / Comment)*
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		accept = true
		for accept {
			{
				save := p.ParserData.Pos()
				accept = p.Space()
				if !accept {
					accept = p.Comment()
					if !accept {
					}
				}
				if !accept {
					p.ParserData.Seek(save)
				}
			}
		}
		accept = true
	}
	if accept && start != p.ParserData.Pos() {
		if start < p.IgnoreRange.A || p.IgnoreRange.A == 0 {
			p.IgnoreRange.A = start
		}
		p.IgnoreRange.B = p.ParserData.Pos()
	}
	return accept
}

func (p *Peg) Comment() bool {
	// Comment       <- '#' (!EndOfLine .)* EndOfLine
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		if p.ParserData.Read() != '#' {
			p.ParserData.UnRead()
			accept = false
		} else {
			accept = true
		}
		if accept {
			{
				accept = true
				for accept {
					{
						save := p.ParserData.Pos()
						s := p.ParserData.Pos()
						accept = p.EndOfLine()
						p.ParserData.Seek(s)
						p.Root.Discard(s)
						accept = !accept
						if accept {
							if p.ParserData.Pos() >= p.ParserData.Len() {
								accept = false
							} else {
								p.ParserData.Read()
								accept = true
							}
							if accept {
							}
						}
						if !accept {
							if p.LastError < p.ParserData.Pos() {
								p.LastError = p.ParserData.Pos()
							}
							p.ParserData.Seek(save)
						}
					}
				}
				accept = true
			}
			if accept {
				accept = p.EndOfLine()
				if accept {
				}
			}
		}
		if !accept {
			if p.LastError < p.ParserData.Pos() {
				p.LastError = p.ParserData.Pos()
			}
			p.ParserData.Seek(save)
		}
	}
	if accept && start != p.ParserData.Pos() {
		if start < p.IgnoreRange.A || p.IgnoreRange.A == 0 {
			p.IgnoreRange.A = start
		}
		p.IgnoreRange.B = p.ParserData.Pos()
	}
	return accept
}

func (p *Peg) Space() bool {
	// Space         <- ' ' / '\t' / EndOfLine
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		if p.ParserData.Read() != ' ' {
			p.ParserData.UnRead()
			accept = false
		} else {
			accept = true
		}
		if !accept {
			if p.ParserData.Read() != '\t' {
				p.ParserData.UnRead()
				accept = false
			} else {
				accept = true
			}
			if !accept {
				accept = p.EndOfLine()
				if !accept {
				}
			}
		}
		if !accept {
			p.ParserData.Seek(save)
		}
	}
	if accept && start != p.ParserData.Pos() {
		if start < p.IgnoreRange.A || p.IgnoreRange.A == 0 {
			p.IgnoreRange.A = start
		}
		p.IgnoreRange.B = p.ParserData.Pos()
	}
	return accept
}

func (p *Peg) EndOfLine() bool {
	// EndOfLine     <- "\r\n" / '\n' / '\r'
	accept := false
	accept = true
	start := p.ParserData.Pos()
	{
		save := p.ParserData.Pos()
		{
			accept = true
			s := p.ParserData.Pos()
			if p.ParserData.Read() != '\r' || p.ParserData.Read() != '\n' {
				p.ParserData.Seek(s)
				accept = false
			}
		}
		if !accept {
			if p.ParserData.Read() != '\n' {
				p.ParserData.UnRead()
				accept = false
			} else {
				accept = true
			}
			if !accept {
				if p.ParserData.Read() != '\r' {
					p.ParserData.UnRead()
					accept = false
				} else {
					accept = true
				}
				if !accept {
				}
			}
		}
		if !accept {
			p.ParserData.Seek(save)
		}
	}
	if accept && start != p.ParserData.Pos() {
		if start < p.IgnoreRange.A || p.IgnoreRange.A == 0 {
			p.IgnoreRange.A = start
		}
		p.IgnoreRange.B = p.ParserData.Pos()
	}
	return accept
}

func (p *Peg) EndOfFile() bool {
	// EndOfFile     <- !.
	accept := false
	accept = true
	start := p.ParserData.Pos()
	s := p.ParserData.Pos()
	if p.ParserData.Pos() >= p.ParserData.Len() {
		accept = false
	} else {
		p.ParserData.Read()
		accept = true
	}
	p.ParserData.Seek(s)
	p.Root.Discard(s)
	accept = !accept
	end := p.ParserData.Pos()
	if accept {
		node := p.Root.Cleanup(start, end)
		node.Name = "EndOfFile"
		node.P = p
		node.Range = node.Range.Clip(p.IgnoreRange)
		p.Root.Append(node)
	} else {
		p.Root.Discard(start)
	}
	if p.IgnoreRange.A >= end || p.IgnoreRange.B <= start {
		p.IgnoreRange = text.Region{}
	}
	return accept
}
