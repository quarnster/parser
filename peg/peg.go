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
func (p *Peg) Parse() bool {
	return p_Grammar(p)
}
func p_Grammar(p *Peg) bool {
	// Grammar       <- Spacing Definition+ EndOfFile?
	return p_addNode(p, func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		accept = p_Spacing(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		accept = func(p *Peg) bool {
			save := p.ParserData.Pos
			accept := true
			accept = p_Definition(p)
			if !accept {
				p.ParserData.Pos = save
			return false
			}
			for accept {
				accept = p_Definition(p)
			}
			return true
		}(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		accept = p_EndOfFile(p)
		accept = true
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		return true
	}, "Grammar")
}

func p_Definition(p *Peg) bool {
	// Definition    <- Identifier LEFTARROW Expression
	return p_addNode(p, func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		accept = p_Identifier(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		accept = p_LEFTARROW(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		accept = p_Expression(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		return true
	}, "Definition")
}

func p_Expression(p *Peg) bool {
	// Expression    <- Sequence (SLASH Sequence)*
	return p_addNode(p, func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		accept = p_Sequence(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		accept = func(p *Peg) bool {
			accept := true
			accept = func(p *Peg) bool {
			accept := false
			save := p.ParserData.Pos
			accept = p_SLASH(p)
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			accept = p_Sequence(p)
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			return true
		}(p)
			for accept {
				accept = func(p *Peg) bool {
			accept := false
			save := p.ParserData.Pos
			accept = p_SLASH(p)
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			accept = p_Sequence(p)
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			return true
		}(p)
			}
			return true
		}(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		return true
	}, "Expression")
}

func p_Sequence(p *Peg) bool {
	// Sequence      <- Prefix*
	return p_addNode(p, func(p *Peg) bool {
		accept := true
		accept = p_Prefix(p)
		for accept {
			accept = p_Prefix(p)
		}
		return true
	}, "Sequence")
}

func p_Prefix(p *Peg) bool {
	// Prefix        <- (AND / NOT)? Suffix
	return p_addNode(p, func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		accept = func(p *Peg) bool {
			accept := false
			save := p.ParserData.Pos
			accept = p_AND(p)
			if(accept) { return true }
			accept = p_NOT(p)
			if(accept) { return true }
			p.ParserData.Pos = save
			return false
		}(p)
		accept = true
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		accept = p_Suffix(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		return true
	}, "Prefix")
}

func p_Suffix(p *Peg) bool {
	// Suffix        <- Primary (QUESTION / STAR / PLUS)?
	return p_addNode(p, func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		accept = p_Primary(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		accept = func(p *Peg) bool {
			accept := false
			save := p.ParserData.Pos
			accept = p_QUESTION(p)
			if(accept) { return true }
			accept = p_STAR(p)
			if(accept) { return true }
			accept = p_PLUS(p)
			if(accept) { return true }
			p.ParserData.Pos = save
			return false
		}(p)
		accept = true
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		return true
	}, "Suffix")
}

func p_Primary(p *Peg) bool {
	// Primary       <- Identifier !LEFTARROW
	//                / OPEN Expression CLOSE
	//                / Literal / Class / DOT
	// # Lexical syntax
	return p_addNode(p, func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		accept = func(p *Peg) bool {
			accept := false
			save := p.ParserData.Pos
			accept = p_Identifier(p)
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			s := p.ParserData.Pos
			accept = p_LEFTARROW(p)
			p.ParserData.Pos = s
			accept = !accept
			
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			return true
		}(p)
		if(accept) { return true }
		accept = func(p *Peg) bool {
			accept := false
			save := p.ParserData.Pos
			accept = p_OPEN(p)
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			accept = p_Expression(p)
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			accept = p_CLOSE(p)
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			return true
		}(p)
		if(accept) { return true }
		accept = p_Literal(p)
		if(accept) { return true }
		accept = p_Class(p)
		if(accept) { return true }
		accept = p_DOT(p)
		if(accept) { return true }
		p.ParserData.Pos = save
		return false
	}, "Primary")
}

func p_Identifier(p *Peg) bool {
	// Identifier    <- IdentStart IdentCont* Spacing
	return p_addNode(p, func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		accept = p_IdentStart(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		accept = func(p *Peg) bool {
			accept := true
			accept = p_IdentCont(p)
			for accept {
				accept = p_IdentCont(p)
			}
			return true
		}(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		accept = p_Spacing(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		return true
	}, "Identifier")
}

func p_IdentStart(p *Peg) bool {
	// IdentStart    <- [a-zA-Z_]
	accept := false
	accept = func(p *Peg) bool {
		accept := false
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
		
		if(accept) { return true }
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
		
		if(accept) { return true }
		accept = func(p *Peg) bool {
			dataset := []rune("_")
			if p.ParserData.Pos >= len(p.ParserData.Data) {
				return false
			}
			c := p.ParserData.Data[p.ParserData.Pos]
			for _, r := range dataset {
				if r == c {
					p.ParserData.Pos++
					return true
				}
			}
			return false
		}(p)
		if(accept) { return true }
		p.ParserData.Pos = save
		return false
	}(p)
	return accept
}

func p_IdentCont(p *Peg) bool {
	// IdentCont     <- IdentStart / [0-9]
	accept := false
	accept = func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		accept = p_IdentStart(p)
		if(accept) { return true }
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
		
		if(accept) { return true }
		p.ParserData.Pos = save
		return false
	}(p)
	return accept
}

func p_Literal(p *Peg) bool {
	// Literal       <- ['] (!['] Char)* ['] Spacing
	//                / ["] (!["] Char)* ["] Spacing
	return p_addNode(p, func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		accept = func(p *Peg) bool {
			accept := false
			save := p.ParserData.Pos
			accept = func(p *Peg) bool {
				dataset := []rune("'")
				if p.ParserData.Pos >= len(p.ParserData.Data) {
					return false
				}
				c := p.ParserData.Data[p.ParserData.Pos]
				for _, r := range dataset {
					if r == c {
						p.ParserData.Pos++
						return true
					}
				}
				return false
			}(p)
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			accept = func(p *Peg) bool {
				accept := true
				accept = func(p *Peg) bool {
				accept := false
				save := p.ParserData.Pos
				s := p.ParserData.Pos
				accept = func(p *Peg) bool {
					dataset := []rune("'")
					if p.ParserData.Pos >= len(p.ParserData.Data) {
						return false
					}
					c := p.ParserData.Data[p.ParserData.Pos]
					for _, r := range dataset {
						if r == c {
							p.ParserData.Pos++
							return true
						}
					}
					return false
				}(p)
				p.ParserData.Pos = s
				accept = !accept
				
				if(!accept) {
					p.ParserData.Pos = save
					return false
				}
				accept = p_Char(p)
				if(!accept) {
					p.ParserData.Pos = save
					return false
				}
				return true
			}(p)
				for accept {
					accept = func(p *Peg) bool {
				accept := false
				save := p.ParserData.Pos
				s := p.ParserData.Pos
				accept = func(p *Peg) bool {
					dataset := []rune("'")
					if p.ParserData.Pos >= len(p.ParserData.Data) {
						return false
					}
					c := p.ParserData.Data[p.ParserData.Pos]
					for _, r := range dataset {
						if r == c {
							p.ParserData.Pos++
							return true
						}
					}
					return false
				}(p)
				p.ParserData.Pos = s
				accept = !accept
				
				if(!accept) {
					p.ParserData.Pos = save
					return false
				}
				accept = p_Char(p)
				if(!accept) {
					p.ParserData.Pos = save
					return false
				}
				return true
			}(p)
				}
				return true
			}(p)
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			accept = func(p *Peg) bool {
				dataset := []rune("'")
				if p.ParserData.Pos >= len(p.ParserData.Data) {
					return false
				}
				c := p.ParserData.Data[p.ParserData.Pos]
				for _, r := range dataset {
					if r == c {
						p.ParserData.Pos++
						return true
					}
				}
				return false
			}(p)
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			accept = p_Spacing(p)
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			return true
		}(p)
		if(accept) { return true }
		accept = func(p *Peg) bool {
			accept := false
			save := p.ParserData.Pos
			accept = func(p *Peg) bool {
				dataset := []rune("\"")
				if p.ParserData.Pos >= len(p.ParserData.Data) {
					return false
				}
				c := p.ParserData.Data[p.ParserData.Pos]
				for _, r := range dataset {
					if r == c {
						p.ParserData.Pos++
						return true
					}
				}
				return false
			}(p)
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			accept = func(p *Peg) bool {
				accept := true
				accept = func(p *Peg) bool {
				accept := false
				save := p.ParserData.Pos
				s := p.ParserData.Pos
				accept = func(p *Peg) bool {
					dataset := []rune("\"")
					if p.ParserData.Pos >= len(p.ParserData.Data) {
						return false
					}
					c := p.ParserData.Data[p.ParserData.Pos]
					for _, r := range dataset {
						if r == c {
							p.ParserData.Pos++
							return true
						}
					}
					return false
				}(p)
				p.ParserData.Pos = s
				accept = !accept
				
				if(!accept) {
					p.ParserData.Pos = save
					return false
				}
				accept = p_Char(p)
				if(!accept) {
					p.ParserData.Pos = save
					return false
				}
				return true
			}(p)
				for accept {
					accept = func(p *Peg) bool {
				accept := false
				save := p.ParserData.Pos
				s := p.ParserData.Pos
				accept = func(p *Peg) bool {
					dataset := []rune("\"")
					if p.ParserData.Pos >= len(p.ParserData.Data) {
						return false
					}
					c := p.ParserData.Data[p.ParserData.Pos]
					for _, r := range dataset {
						if r == c {
							p.ParserData.Pos++
							return true
						}
					}
					return false
				}(p)
				p.ParserData.Pos = s
				accept = !accept
				
				if(!accept) {
					p.ParserData.Pos = save
					return false
				}
				accept = p_Char(p)
				if(!accept) {
					p.ParserData.Pos = save
					return false
				}
				return true
			}(p)
				}
				return true
			}(p)
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			accept = func(p *Peg) bool {
				dataset := []rune("\"")
				if p.ParserData.Pos >= len(p.ParserData.Data) {
					return false
				}
				c := p.ParserData.Data[p.ParserData.Pos]
				for _, r := range dataset {
					if r == c {
						p.ParserData.Pos++
						return true
					}
				}
				return false
			}(p)
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			accept = p_Spacing(p)
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			return true
		}(p)
		if(accept) { return true }
		p.ParserData.Pos = save
		return false
	}, "Literal")
}

func p_Class(p *Peg) bool {
	// Class         <- '[' (!']' Range)* ']' Spacing
	return p_addNode(p, func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '[' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		accept = func(p *Peg) bool {
			accept := true
			accept = func(p *Peg) bool {
			accept := false
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
			
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			accept = p_Range(p)
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			return true
		}(p)
			for accept {
				accept = func(p *Peg) bool {
			accept := false
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
			
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			accept = p_Range(p)
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			return true
		}(p)
			}
			return true
		}(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != ']' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		accept = p_Spacing(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		return true
	}, "Class")
}

func p_Range(p *Peg) bool {
	// Range         <- Char '-' Char / Char
	return p_addNode(p, func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		accept = func(p *Peg) bool {
			accept := false
			save := p.ParserData.Pos
			accept = p_Char(p)
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '-' {
				accept = false
			} else {
				p.ParserData.Pos++
				accept = true
			}
			
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			accept = p_Char(p)
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			return true
		}(p)
		if(accept) { return true }
		accept = p_Char(p)
		if(accept) { return true }
		p.ParserData.Pos = save
		return false
	}, "Range")
}

func p_Char(p *Peg) bool {
	// Char          <- '\\' [nrt'"\[\]\\]
	//                / '\\' [0-2][0-7][0-7]
	//                / '\\' [0-7][0-7]?
	//                / !'\\' .
	return p_addNode(p, func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		accept = func(p *Peg) bool {
			accept := false
			save := p.ParserData.Pos
			if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '\\' {
				accept = false
			} else {
				p.ParserData.Pos++
				accept = true
			}
			
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			accept = func(p *Peg) bool {
				dataset := []rune("nrt'\"[]\\")
				if p.ParserData.Pos >= len(p.ParserData.Data) {
					return false
				}
				c := p.ParserData.Data[p.ParserData.Pos]
				for _, r := range dataset {
					if r == c {
						p.ParserData.Pos++
						return true
					}
				}
				return false
			}(p)
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			return true
		}(p)
		if(accept) { return true }
		accept = func(p *Peg) bool {
			accept := false
			save := p.ParserData.Pos
			if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '\\' {
				accept = false
			} else {
				p.ParserData.Pos++
				accept = true
			}
			
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
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
			
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
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
			
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
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
			
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			return true
		}(p)
		if(accept) { return true }
		accept = func(p *Peg) bool {
			accept := false
			save := p.ParserData.Pos
			if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '\\' {
				accept = false
			} else {
				p.ParserData.Pos++
				accept = true
			}
			
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
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
			
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
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
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			return true
		}(p)
		if(accept) { return true }
		accept = func(p *Peg) bool {
			accept := false
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
			
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			accept = func(p *Peg) bool {
				if p.ParserData.Pos >= len(p.ParserData.Data) {
					return false
				}
				p.ParserData.Pos++
				return true
			}(p)
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			return true
		}(p)
		if(accept) { return true }
		p.ParserData.Pos = save
		return false
	}, "Char")
}

func p_LEFTARROW(p *Peg) bool {
	// LEFTARROW     <- "<-" Spacing
	return p_Ignore(p, func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		{
			accept = true
			n1 := []rune{'<', '-'}
			s := p.ParserData.Pos
			e := s + len(n1)
			if e > len(p.ParserData.Data) {
				accept = false
			} else {
				for i := 0; i < len(n1); i++ {
					if n1[i] != p.ParserData.Data[s+i] {
						accept = false
						break
					}
				}
			}
			if (accept) {
				p.ParserData.Pos += len(n1)
			}
		}
		
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		accept = p_Spacing(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		return true
	})
}

func p_SLASH(p *Peg) bool {
	// SLASH         <- '/' Spacing
	return p_Ignore(p, func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '/' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		accept = p_Spacing(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		return true
	})
}

func p_AND(p *Peg) bool {
	// AND           <- '&' Spacing
	return p_addNode(p, func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '&' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		accept = p_Spacing(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		return true
	}, "AND")
}

func p_NOT(p *Peg) bool {
	// NOT           <- '!' Spacing
	return p_addNode(p, func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '!' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		accept = p_Spacing(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		return true
	}, "NOT")
}

func p_QUESTION(p *Peg) bool {
	// QUESTION      <- '?' Spacing
	return p_addNode(p, func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '?' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		accept = p_Spacing(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		return true
	}, "QUESTION")
}

func p_STAR(p *Peg) bool {
	// STAR          <- '*' Spacing
	return p_addNode(p, func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '*' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		accept = p_Spacing(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		return true
	}, "STAR")
}

func p_PLUS(p *Peg) bool {
	// PLUS          <- '+' Spacing
	return p_addNode(p, func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '+' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		accept = p_Spacing(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		return true
	}, "PLUS")
}

func p_OPEN(p *Peg) bool {
	// OPEN          <- '(' Spacing
	return p_Ignore(p, func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '(' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		accept = p_Spacing(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		return true
	})
}

func p_CLOSE(p *Peg) bool {
	// CLOSE         <- ')' Spacing
	return p_Ignore(p, func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != ')' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		accept = p_Spacing(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		return true
	})
}

func p_DOT(p *Peg) bool {
	// DOT           <- '.' Spacing
	return p_addNode(p, func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '.' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		accept = p_Spacing(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		return true
	}, "DOT")
}

func p_Spacing(p *Peg) bool {
	// Spacing       <- (Space / Comment)*
	return p_Ignore(p, func(p *Peg) bool {
		accept := true
		accept = func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		accept = p_Space(p)
		if(accept) { return true }
		accept = p_Comment(p)
		if(accept) { return true }
		p.ParserData.Pos = save
		return false
	}(p)
		for accept {
			accept = func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		accept = p_Space(p)
		if(accept) { return true }
		accept = p_Comment(p)
		if(accept) { return true }
		p.ParserData.Pos = save
		return false
	}(p)
		}
		return true
	})
}

func p_Comment(p *Peg) bool {
	// Comment       <- '#' (!EndOfLine .)* EndOfLine
	return p_Ignore(p, func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '#' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		accept = func(p *Peg) bool {
			accept := true
			accept = func(p *Peg) bool {
			accept := false
			save := p.ParserData.Pos
			s := p.ParserData.Pos
			accept = p_EndOfLine(p)
			p.ParserData.Pos = s
			accept = !accept
			
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			accept = func(p *Peg) bool {
				if p.ParserData.Pos >= len(p.ParserData.Data) {
					return false
				}
				p.ParserData.Pos++
				return true
			}(p)
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			return true
		}(p)
			for accept {
				accept = func(p *Peg) bool {
			accept := false
			save := p.ParserData.Pos
			s := p.ParserData.Pos
			accept = p_EndOfLine(p)
			p.ParserData.Pos = s
			accept = !accept
			
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			accept = func(p *Peg) bool {
				if p.ParserData.Pos >= len(p.ParserData.Data) {
					return false
				}
				p.ParserData.Pos++
				return true
			}(p)
			if(!accept) {
				p.ParserData.Pos = save
				return false
			}
			return true
		}(p)
			}
			return true
		}(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		accept = p_EndOfLine(p)
		if(!accept) {
			p.ParserData.Pos = save
			return false
		}
		return true
	})
}

func p_Space(p *Peg) bool {
	// Space         <- ' ' / '\t' / EndOfLine
	return p_Ignore(p, func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != ' ' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		
		if(accept) { return true }
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '\t' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		
		if(accept) { return true }
		accept = p_EndOfLine(p)
		if(accept) { return true }
		p.ParserData.Pos = save
		return false
	})
}

func p_EndOfLine(p *Peg) bool {
	// EndOfLine     <- "\r\n" / '\n' / '\r'
	return p_Ignore(p, func(p *Peg) bool {
		accept := false
		save := p.ParserData.Pos
		{
			accept = true
			n1 := []rune{'\r', '\n'}
			s := p.ParserData.Pos
			e := s + len(n1)
			if e > len(p.ParserData.Data) {
				accept = false
			} else {
				for i := 0; i < len(n1); i++ {
					if n1[i] != p.ParserData.Data[s+i] {
						accept = false
						break
					}
				}
			}
			if (accept) {
				p.ParserData.Pos += len(n1)
			}
		}
		
		if(accept) { return true }
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '\n' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		
		if(accept) { return true }
		if p.ParserData.Pos >= len(p.ParserData.Data) || p.ParserData.Data[p.ParserData.Pos] != '\r' {
			accept = false
		} else {
			p.ParserData.Pos++
			accept = true
		}
		
		if(accept) { return true }
		p.ParserData.Pos = save
		return false
	})
}

func p_EndOfFile(p *Peg) bool {
	// EndOfFile     <- !.
	return p_Ignore(p, func(p *Peg) bool {
		accept := true
		s := p.ParserData.Pos
		accept = func(p *Peg) bool {
			if p.ParserData.Pos >= len(p.ParserData.Data) {
				return false
			}
			p.ParserData.Pos++
			return true
		}(p)
		p.ParserData.Pos = s
		accept = !accept
		
		return accept
	
	})
}
