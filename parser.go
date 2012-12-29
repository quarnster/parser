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

type PegParser struct {
	Parser
}

func (p *PegParser) Grammar() bool {
	/* Grammar       <- Spacing Definition+ EndOfFile */
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.Spacing()
			},
			func() bool {
				return p.OneOrMore(func() bool {
					return p.Definition()
				})
			},
			func() bool {
				return p.EndOfFile()
			},
		})
	}, "Grammar")
}

func (p *PegParser) Definition() bool {
	/* Definition    <- Identifier LEFTARROW Expression */
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.Identifier()
			},
			func() bool {
				return p.LEFTARROW()
			},
			func() bool {
				return p.Expression()
			},
		})
	}, "Definition")
}

func (p *PegParser) Expression() bool {
	/* Expression    <- Sequence (SLASH Sequence)* */
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.Sequence()
			},
			func() bool {
				return p.ZeroOrMore(func() bool {
					return p.NeedAll([]func() bool{
						func() bool {
							return p.SLASH()
						},
						func() bool {
							return p.Sequence()
						},
					})
				})
			},
		})
	}, "Expression")
}

func (p *PegParser) Sequence() bool {
	/* Sequence      <- Prefix* */
	return p.addNode(func() bool {
		return p.ZeroOrMore(func() bool {
			return p.Prefix()
		})
	}, "Sequence")
}

func (p *PegParser) Prefix() bool {
	/* Prefix        <- (AND / NOT)? Suffix */
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.Maybe(func() bool {
					return p.NeedOne([]func() bool{
						func() bool {
							return p.AND()
						},
						func() bool {
							return p.NOT()
						},
					})
				})
			},
			func() bool {
				return p.Suffix()
			},
		})
	}, "Prefix")
}

func (p *PegParser) Suffix() bool {
	/* Suffix        <- Primary (QUESTION / STAR / PLUS)? */
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.Primary()
			},
			func() bool {
				return p.Maybe(func() bool {
					return p.NeedOne([]func() bool{
						func() bool {
							return p.QUESTION()
						},
						func() bool {
							return p.STAR()
						},
						func() bool {
							return p.PLUS()
						},
					})
				})
			},
		})
	}, "Suffix")
}

func (p *PegParser) Primary() bool {
	/* Primary       <- Identifier !LEFTARROW
	 *                / OPEN Expression CLOSE
	 *                / Literal / Class / DOT
	 * # Lexical syntax
	 */
	return p.addNode(func() bool {
		return p.NeedOne([]func() bool{
			func() bool {
				return p.NeedAll([]func() bool{
					func() bool {
						return p.Identifier()
					},
					func() bool {
						return p.Not(func() bool {
							return p.LEFTARROW()
						})
					},
				})
			},
			func() bool {
				return p.NeedAll([]func() bool{
					func() bool {
						return p.OPEN()
					},
					func() bool {
						return p.Expression()
					},
					func() bool {
						return p.CLOSE()
					},
				})
			},
			func() bool {
				return p.Literal()
			},
			func() bool {
				return p.Class()
			},
			func() bool {
				return p.DOT()
			},
		})
	}, "Primary")
}

func (p *PegParser) Identifier() bool {
	/* Identifier    <- IdentStart IdentCont* Spacing */
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.IdentStart()
			},
			func() bool {
				return p.ZeroOrMore(func() bool {
					return p.IdentCont()
				})
			},
			func() bool {
				return p.Spacing()
			},
		})
	}, "Identifier")
}

func (p *PegParser) IdentStart() bool {
	/* IdentStart    <- [a-zA-Z_] */
	return p.addNode(func() bool {
		return p.NeedOne([]func() bool{
			func() bool {
				return p.InRange('a', 'z')
			},
			func() bool {
				return p.InRange('A', 'Z')
			},
			func() bool {
				return p.InSet(`_`)
			},
		})
	}, "IdentStart")
}

func (p *PegParser) IdentCont() bool {
	/* IdentCont     <- IdentStart / [0-9] */
	return p.addNode(func() bool {
		return p.NeedOne([]func() bool{
			func() bool {
				return p.IdentStart()
			},
			func() bool {
				return p.NeedOne([]func() bool{
					func() bool {
						return p.InRange('0', '9')
					},
				})
			},
		})
	}, "IdentCont")
}

func (p *PegParser) Literal() bool {
	/* Literal       <- ['] (!['] Char)* ['] Spacing
	 *                / ["] (!["] Char)* ["] Spacing
	 */
	return p.addNode(func() bool {
		return p.NeedOne([]func() bool{
			func() bool {
				return p.NeedAll([]func() bool{
					func() bool {
						return p.NeedOne([]func() bool{
							func() bool {
								return p.InSet(`'`)
							},
						})
					},
					func() bool {
						return p.ZeroOrMore(func() bool {
							return p.NeedAll([]func() bool{
								func() bool {
									return p.Not(func() bool {
										return p.NeedOne([]func() bool{
											func() bool {
												return p.InSet(`'`)
											},
										})
									})
								},
								func() bool {
									return p.Char()
								},
							})
						})
					},
					func() bool {
						return p.NeedOne([]func() bool{
							func() bool {
								return p.InSet(`'`)
							},
						})
					},
					func() bool {
						return p.Spacing()
					},
				})
			},
			func() bool {
				return p.NeedAll([]func() bool{
					func() bool {
						return p.NeedOne([]func() bool{
							func() bool {
								return p.InSet(`"`)
							},
						})
					},
					func() bool {
						return p.ZeroOrMore(func() bool {
							return p.NeedAll([]func() bool{
								func() bool {
									return p.Not(func() bool {
										return p.NeedOne([]func() bool{
											func() bool {
												return p.InSet(`"`)
											},
										})
									})
								},
								func() bool {
									return p.Char()
								},
							})
						})
					},
					func() bool {
						return p.NeedOne([]func() bool{
							func() bool {
								return p.InSet(`"`)
							},
						})
					},
					func() bool {
						return p.Spacing()
					},
				})
			},
		})
	}, "Literal")
}

func (p *PegParser) Class() bool {
	/* Class         <- '[' (!']' Range)* ']' Spacing */
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.Next("[")
			},
			func() bool {
				return p.ZeroOrMore(func() bool {
					return p.NeedAll([]func() bool{
						func() bool {
							return p.Not(func() bool {
								return p.Next("]")
							})
						},
						func() bool {
							return p.Range()
						},
					})
				})
			},
			func() bool {
				return p.Next("]")
			},
			func() bool {
				return p.Spacing()
			},
		})
	}, "Class")
}

func (p *PegParser) Range() bool {
	/* Range         <- Char '-' Char / Char */
	return p.addNode(func() bool {
		return p.NeedOne([]func() bool{
			func() bool {
				return p.NeedAll([]func() bool{
					func() bool {
						return p.Char()
					},
					func() bool {
						return p.Next("-")
					},
					func() bool {
						return p.Char()
					},
				})
			},
			func() bool {
				return p.Char()
			},
		})
	}, "Range")
}

func (p *PegParser) Char() bool {
	/* Char          <- '\\' [nrt'"\[\]\\]
	 *                / '\\' [0-2][0-7][0-7]
	 *                / '\\' [0-7][0-7]?
	 *                / !'\\' .
	 */
	return p.addNode(func() bool {
		return p.NeedOne([]func() bool{
			func() bool {
				return p.NeedAll([]func() bool{
					func() bool {
						return p.Next("\\")
					},
					func() bool {
						return p.NeedOne([]func() bool{
							func() bool {
								return p.InSet(`nrt'"\[\]\\`)
							},
						})
					},
				})
			},
			func() bool {
				return p.NeedAll([]func() bool{
					func() bool {
						return p.Next("\\")
					},
					func() bool {
						return p.NeedOne([]func() bool{
							func() bool {
								return p.InRange('0', '2')
							},
						})
					},
					func() bool {
						return p.NeedOne([]func() bool{
							func() bool {
								return p.InRange('0', '7')
							},
						})
					},
					func() bool {
						return p.NeedOne([]func() bool{
							func() bool {
								return p.InRange('0', '7')
							},
						})
					},
				})
			},
			func() bool {
				return p.NeedAll([]func() bool{
					func() bool {
						return p.Next("\\")
					},
					func() bool {
						return p.NeedOne([]func() bool{
							func() bool {
								return p.InRange('0', '7')
							},
						})
					},
					func() bool {
						return p.Maybe(func() bool {
							return p.NeedOne([]func() bool{
								func() bool {
									return p.InRange('0', '7')
								},
							})
						})
					},
				})
			},
			func() bool {
				return p.NeedAll([]func() bool{
					func() bool {
						return p.Not(func() bool {
							return p.Next("\\")
						})
					},
					func() bool {
						return p.AnyChar()
					},
				})
			},
		})
	}, "Char")
}

func (p *PegParser) LEFTARROW() bool {
	/* LEFTARROW     <- '<-' Spacing */
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.Next("<-")
			},
			func() bool {
				return p.Spacing()
			},
		})
	}, "LEFTARROW")
}

func (p *PegParser) SLASH() bool {
	/* SLASH         <- '/' Spacing */
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.Next("/")
			},
			func() bool {
				return p.Spacing()
			},
		})
	}, "SLASH")
}

func (p *PegParser) AND() bool {
	/* AND           <- '&' Spacing */
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.Next("&")
			},
			func() bool {
				return p.Spacing()
			},
		})
	}, "AND")
}

func (p *PegParser) NOT() bool {
	/* NOT           <- '!' Spacing */
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.Next("!")
			},
			func() bool {
				return p.Spacing()
			},
		})
	}, "NOT")
}

func (p *PegParser) QUESTION() bool {
	/* QUESTION      <- '?' Spacing */
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.Next("?")
			},
			func() bool {
				return p.Spacing()
			},
		})
	}, "QUESTION")
}

func (p *PegParser) STAR() bool {
	/* STAR          <- '*' Spacing */
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.Next("*")
			},
			func() bool {
				return p.Spacing()
			},
		})
	}, "STAR")
}

func (p *PegParser) PLUS() bool {
	/* PLUS          <- '+' Spacing */
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.Next("+")
			},
			func() bool {
				return p.Spacing()
			},
		})
	}, "PLUS")
}

func (p *PegParser) OPEN() bool {
	/* OPEN          <- '(' Spacing */
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.Next("(")
			},
			func() bool {
				return p.Spacing()
			},
		})
	}, "OPEN")
}

func (p *PegParser) CLOSE() bool {
	/* CLOSE         <- ')' Spacing */
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.Next(")")
			},
			func() bool {
				return p.Spacing()
			},
		})
	}, "CLOSE")
}

func (p *PegParser) DOT() bool {
	/* DOT           <- '.' Spacing */
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.Next(".")
			},
			func() bool {
				return p.Spacing()
			},
		})
	}, "DOT")
}

func (p *PegParser) Spacing() bool {
	/* Spacing       <- (Space / Comment)* */
	return p.addNode(func() bool {
		return p.ZeroOrMore(func() bool {
			return p.NeedOne([]func() bool{
				func() bool {
					return p.Space()
				},
				func() bool {
					return p.Comment()
				},
			})
		})
	}, "Spacing")
}

func (p *PegParser) Comment() bool {
	/* Comment       <- '#' (!EndOfLine .)* EndOfLine */
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.Next("#")
			},
			func() bool {
				return p.ZeroOrMore(func() bool {
					return p.NeedAll([]func() bool{
						func() bool {
							return p.Not(func() bool {
								return p.EndOfLine()
							})
						},
						func() bool {
							return p.AnyChar()
						},
					})
				})
			},
			func() bool {
				return p.EndOfLine()
			},
		})
	}, "Comment")
}

func (p *PegParser) Space() bool {
	/* Space         <- ' ' / '\t' / EndOfLine */
	return p.addNode(func() bool {
		return p.NeedOne([]func() bool{
			func() bool {
				return p.Next(" ")
			},
			func() bool {
				return p.Next("\t")
			},
			func() bool {
				return p.EndOfLine()
			},
		})
	}, "Space")
}

func (p *PegParser) EndOfLine() bool {
	/* EndOfLine     <- '\r\n' / '\n' / '\r' */
	return p.addNode(func() bool {
		return p.NeedOne([]func() bool{
			func() bool {
				return p.Next("\r\n")
			},
			func() bool {
				return p.Next("\n")
			},
			func() bool {
				return p.Next("\r")
			},
		})
	}, "EndOfLine")
}

func (p *PegParser) EndOfFile() bool {
	/* EndOfFile     <- !. */
	return p.addNode(func() bool {
		return p.Not(func() bool {
			return p.AnyChar()
		})
	}, "EndOfFile")
}
