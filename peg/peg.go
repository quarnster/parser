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
	ParserData struct {
		Pos  int
		Data []rune
	}

	IgnoreRange parser.Range
	Root        parser.Node
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
func (p *Peg) Parse() bool {
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
					return false
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
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	p.Root.P = p
	// Remove any danglers
	p.Root.Cleanup(p.ParserData.Pos, -1)
	node := p.Root.Cleanup(start, p.ParserData.Pos)
	node.Name = "Grammar"
	if accept {
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
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	p.Root.P = p
	// Remove any danglers
	p.Root.Cleanup(p.ParserData.Pos, -1)
	node := p.Root.Cleanup(start, p.ParserData.Pos)
	node.Name = "Definition"
	if accept {
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
				{
					save := p.ParserData.Pos
					accept = p_SLASH(p)
					if accept {
						accept = p_Sequence(p)
						if accept {
						}
					}
					if !accept {
						p.ParserData.Pos = save
					}
				}
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
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	p.Root.P = p
	// Remove any danglers
	p.Root.Cleanup(p.ParserData.Pos, -1)
	node := p.Root.Cleanup(start, p.ParserData.Pos)
	node.Name = "Expression"
	if accept {
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
	return accept
}

func p_Sequence(p *Peg) bool {
	// Sequence      <- Prefix*
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		accept = p_Prefix(p)
		for accept {
			accept = p_Prefix(p)
		}
		accept = true
	}
	end := p.ParserData.Pos
	p.Root.P = p
	// Remove any danglers
	p.Root.Cleanup(p.ParserData.Pos, -1)
	node := p.Root.Cleanup(start, p.ParserData.Pos)
	node.Name = "Sequence"
	if accept {
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
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	p.Root.P = p
	// Remove any danglers
	p.Root.Cleanup(p.ParserData.Pos, -1)
	node := p.Root.Cleanup(start, p.ParserData.Pos)
	node.Name = "Prefix"
	if accept {
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
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	p.Root.P = p
	// Remove any danglers
	p.Root.Cleanup(p.ParserData.Pos, -1)
	node := p.Root.Cleanup(start, p.ParserData.Pos)
	node.Name = "Suffix"
	if accept {
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
				accept = !accept
				if accept {
				}
			}
			if !accept {
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
	p.Root.P = p
	// Remove any danglers
	p.Root.Cleanup(p.ParserData.Pos, -1)
	node := p.Root.Cleanup(start, p.ParserData.Pos)
	node.Name = "Primary"
	if accept {
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
				accept = p_IdentCont(p)
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
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	p.Root.P = p
	// Remove any danglers
	p.Root.Cleanup(p.ParserData.Pos, -1)
	node := p.Root.Cleanup(start, p.ParserData.Pos)
	node.Name = "Identifier"
	if accept {
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
	// Literal       <- ['] (!['] Char)* ['] Spacing
	//                / ["] (!["] Char)* ["] Spacing
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
		save := p.ParserData.Pos
		{
			save := p.ParserData.Pos
			{
				accept = false
				if p.ParserData.Pos < len(p.ParserData.Data) {
					c := p.ParserData.Data[p.ParserData.Pos]
					if c == '\'' {
						p.ParserData.Pos++
						accept = true
					}
				}
			}
			if accept {
				{
					{
						save := p.ParserData.Pos
						s := p.ParserData.Pos
						{
							accept = false
							if p.ParserData.Pos < len(p.ParserData.Data) {
								c := p.ParserData.Data[p.ParserData.Pos]
								if c == '\'' {
									p.ParserData.Pos++
									accept = true
								}
							}
						}
						p.ParserData.Pos = s
						accept = !accept
						if accept {
							accept = p_Char(p)
							if accept {
							}
						}
						if !accept {
							p.ParserData.Pos = save
						}
					}
					for accept {
						{
							save := p.ParserData.Pos
							s := p.ParserData.Pos
							{
								accept = false
								if p.ParserData.Pos < len(p.ParserData.Data) {
									c := p.ParserData.Data[p.ParserData.Pos]
									if c == '\'' {
										p.ParserData.Pos++
										accept = true
									}
								}
							}
							p.ParserData.Pos = s
							accept = !accept
							if accept {
								accept = p_Char(p)
								if accept {
								}
							}
							if !accept {
								p.ParserData.Pos = save
							}
						}
					}
					accept = true
				}
				if accept {
					{
						accept = false
						if p.ParserData.Pos < len(p.ParserData.Data) {
							c := p.ParserData.Data[p.ParserData.Pos]
							if c == '\'' {
								p.ParserData.Pos++
								accept = true
							}
						}
					}
					if accept {
						accept = p_Spacing(p)
						if accept {
						}
					}
				}
			}
			if !accept {
				p.ParserData.Pos = save
			}
		}
		if !accept {
			{
				save := p.ParserData.Pos
				{
					accept = false
					if p.ParserData.Pos < len(p.ParserData.Data) {
						c := p.ParserData.Data[p.ParserData.Pos]
						if c == '"' {
							p.ParserData.Pos++
							accept = true
						}
					}
				}
				if accept {
					{
						{
							save := p.ParserData.Pos
							s := p.ParserData.Pos
							{
								accept = false
								if p.ParserData.Pos < len(p.ParserData.Data) {
									c := p.ParserData.Data[p.ParserData.Pos]
									if c == '"' {
										p.ParserData.Pos++
										accept = true
									}
								}
							}
							p.ParserData.Pos = s
							accept = !accept
							if accept {
								accept = p_Char(p)
								if accept {
								}
							}
							if !accept {
								p.ParserData.Pos = save
							}
						}
						for accept {
							{
								save := p.ParserData.Pos
								s := p.ParserData.Pos
								{
									accept = false
									if p.ParserData.Pos < len(p.ParserData.Data) {
										c := p.ParserData.Data[p.ParserData.Pos]
										if c == '"' {
											p.ParserData.Pos++
											accept = true
										}
									}
								}
								p.ParserData.Pos = s
								accept = !accept
								if accept {
									accept = p_Char(p)
									if accept {
									}
								}
								if !accept {
									p.ParserData.Pos = save
								}
							}
						}
						accept = true
					}
					if accept {
						{
							accept = false
							if p.ParserData.Pos < len(p.ParserData.Data) {
								c := p.ParserData.Data[p.ParserData.Pos]
								if c == '"' {
									p.ParserData.Pos++
									accept = true
								}
							}
						}
						if accept {
							accept = p_Spacing(p)
							if accept {
							}
						}
					}
				}
				if !accept {
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
	p.Root.P = p
	// Remove any danglers
	p.Root.Cleanup(p.ParserData.Pos, -1)
	node := p.Root.Cleanup(start, p.ParserData.Pos)
	node.Name = "Literal"
	if accept {
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
	return accept
}

func p_Class(p *Peg) bool {
	// Class         <- '[' (!']' Range)* ']' Spacing
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
					accept = !accept
					if accept {
						accept = p_Range(p)
						if accept {
						}
					}
					if !accept {
						p.ParserData.Pos = save
					}
				}
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
						accept = !accept
						if accept {
							accept = p_Range(p)
							if accept {
							}
						}
						if !accept {
							p.ParserData.Pos = save
						}
					}
				}
				accept = true
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
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	p.Root.P = p
	// Remove any danglers
	p.Root.Cleanup(p.ParserData.Pos, -1)
	node := p.Root.Cleanup(start, p.ParserData.Pos)
	node.Name = "Class"
	if accept {
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
	p.Root.P = p
	// Remove any danglers
	p.Root.Cleanup(p.ParserData.Pos, -1)
	node := p.Root.Cleanup(start, p.ParserData.Pos)
	node.Name = "Range"
	if accept {
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
						accept = !accept
						if accept {
							if p.ParserData.Pos >= len(p.ParserData.Data) {
								accept = false
							}
							p.ParserData.Pos++
							accept = true
							if accept {
							}
						}
						if !accept {
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
	p.Root.P = p
	// Remove any danglers
	p.Root.Cleanup(p.ParserData.Pos, -1)
	node := p.Root.Cleanup(start, p.ParserData.Pos)
	node.Name = "Char"
	if accept {
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
			p.ParserData.Pos = save
		}
	}
	if accept {
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
			p.ParserData.Pos = save
		}
	}
	if accept {
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
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	p.Root.P = p
	// Remove any danglers
	p.Root.Cleanup(p.ParserData.Pos, -1)
	node := p.Root.Cleanup(start, p.ParserData.Pos)
	node.Name = "AND"
	if accept {
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
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	p.Root.P = p
	// Remove any danglers
	p.Root.Cleanup(p.ParserData.Pos, -1)
	node := p.Root.Cleanup(start, p.ParserData.Pos)
	node.Name = "NOT"
	if accept {
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
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	p.Root.P = p
	// Remove any danglers
	p.Root.Cleanup(p.ParserData.Pos, -1)
	node := p.Root.Cleanup(start, p.ParserData.Pos)
	node.Name = "QUESTION"
	if accept {
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
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	p.Root.P = p
	// Remove any danglers
	p.Root.Cleanup(p.ParserData.Pos, -1)
	node := p.Root.Cleanup(start, p.ParserData.Pos)
	node.Name = "STAR"
	if accept {
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
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	p.Root.P = p
	// Remove any danglers
	p.Root.Cleanup(p.ParserData.Pos, -1)
	node := p.Root.Cleanup(start, p.ParserData.Pos)
	node.Name = "PLUS"
	if accept {
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
			p.ParserData.Pos = save
		}
	}
	if accept {
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
			p.ParserData.Pos = save
		}
	}
	if accept {
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
			p.ParserData.Pos = save
		}
	}
	end := p.ParserData.Pos
	p.Root.P = p
	// Remove any danglers
	p.Root.Cleanup(p.ParserData.Pos, -1)
	node := p.Root.Cleanup(start, p.ParserData.Pos)
	node.Name = "DOT"
	if accept {
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
	return accept
}

func p_Spacing(p *Peg) bool {
	// Spacing       <- (Space / Comment)*
	accept := false
	accept = true
	start := p.ParserData.Pos
	{
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
	if accept {
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
				{
					save := p.ParserData.Pos
					s := p.ParserData.Pos
					accept = p_EndOfLine(p)
					p.ParserData.Pos = s
					accept = !accept
					if accept {
						if p.ParserData.Pos >= len(p.ParserData.Data) {
							accept = false
						}
						p.ParserData.Pos++
						accept = true
						if accept {
						}
					}
					if !accept {
						p.ParserData.Pos = save
					}
				}
				for accept {
					{
						save := p.ParserData.Pos
						s := p.ParserData.Pos
						accept = p_EndOfLine(p)
						p.ParserData.Pos = s
						accept = !accept
						if accept {
							if p.ParserData.Pos >= len(p.ParserData.Data) {
								accept = false
							}
							p.ParserData.Pos++
							accept = true
							if accept {
							}
						}
						if !accept {
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
			p.ParserData.Pos = save
		}
	}
	if accept {
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
	if accept {
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
	if accept {
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
	}
	p.ParserData.Pos++
	accept = true
	p.ParserData.Pos = s
	accept = !accept
	if accept {
		if start < p.IgnoreRange.Start || p.IgnoreRange.Start == 0 {
			p.IgnoreRange.Start = start
		}
		p.IgnoreRange.End = p.ParserData.Pos
	}

	return accept
}
