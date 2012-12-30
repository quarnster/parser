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

type JSONParser struct {
	Parser
}

func (p *JSONParser) JsonFile() bool {
	// JsonFile       <-    values EndOfFile?
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.values()
			},
			func() bool {
				return p.Maybe(func() bool {
					return p.EndOfFile()
				})
			},
		})
	}, "JsonFile")
}

func (p *JSONParser) values() bool {
	// values         <-    spacing* value spacing* (',' spacing* value spacing*)*
	return p.NeedAll([]func() bool{
		func() bool {
			return p.ZeroOrMore(func() bool {
				return p.spacing()
			})
		},
		func() bool {
			return p.value()
		},
		func() bool {
			return p.ZeroOrMore(func() bool {
				return p.spacing()
			})
		},
		func() bool {
			return p.ZeroOrMore(func() bool {
				return p.NeedAll([]func() bool{
					func() bool {
						return p.Next(",")
					},
					func() bool {
						return p.ZeroOrMore(func() bool {
							return p.spacing()
						})
					},
					func() bool {
						return p.value()
					},
					func() bool {
						return p.ZeroOrMore(func() bool {
							return p.spacing()
						})
					},
				})
			})
		},
	})
}

func (p *JSONParser) value() bool {
	// value          <-    (Dictionary / Array / text / Float / Integer / Boolean / 'null')
	return p.NeedOne([]func() bool{
		func() bool {
			return p.Dictionary()
		},
		func() bool {
			return p.Array()
		},
		func() bool {
			return p.text()
		},
		func() bool {
			return p.Float()
		},
		func() bool {
			return p.Integer()
		},
		func() bool {
			return p.Boolean()
		},
		func() bool {
			return p.Next("null")
		},
	})
}

func (p *JSONParser) Dictionary() bool {
	// Dictionary     <-    '{' keyValuePairs '}'
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.Next("{")
			},
			func() bool {
				return p.keyValuePairs()
			},
			func() bool {
				return p.Next("}")
			},
		})
	}, "Dictionary")
}

func (p *JSONParser) Array() bool {
	// Array          <-    '[' values ']'
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.Next("[")
			},
			func() bool {
				return p.values()
			},
			func() bool {
				return p.Next("]")
			},
		})
	}, "Array")
}

func (p *JSONParser) keyValuePairs() bool {
	// keyValuePairs  <-    spacing* KeyValuePair spacing* (',' spacing* KeyValuePair spacing*)*
	return p.NeedAll([]func() bool{
		func() bool {
			return p.ZeroOrMore(func() bool {
				return p.spacing()
			})
		},
		func() bool {
			return p.KeyValuePair()
		},
		func() bool {
			return p.ZeroOrMore(func() bool {
				return p.spacing()
			})
		},
		func() bool {
			return p.ZeroOrMore(func() bool {
				return p.NeedAll([]func() bool{
					func() bool {
						return p.Next(",")
					},
					func() bool {
						return p.ZeroOrMore(func() bool {
							return p.spacing()
						})
					},
					func() bool {
						return p.KeyValuePair()
					},
					func() bool {
						return p.ZeroOrMore(func() bool {
							return p.spacing()
						})
					},
				})
			})
		},
	})
}

func (p *JSONParser) KeyValuePair() bool {
	// KeyValuePair   <-    text ':' spacing* value
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.text()
			},
			func() bool {
				return p.Next(":")
			},
			func() bool {
				return p.ZeroOrMore(func() bool {
					return p.spacing()
				})
			},
			func() bool {
				return p.value()
			},
		})
	}, "KeyValuePair")
}

func (p *JSONParser) text() bool {
	// text           <-    quote Text quote
	return p.NeedAll([]func() bool{
		func() bool {
			return p.quote()
		},
		func() bool {
			return p.Text()
		},
		func() bool {
			return p.quote()
		},
	})
}

func (p *JSONParser) quote() bool {
	// quote          <-    !'\\' '"'
	return p.NeedAll([]func() bool{
		func() bool {
			return p.Not(func() bool {
				return p.Next("\\")
			})
		},
		func() bool {
			return p.Next("\"")
		},
	})
}

func (p *JSONParser) Text() bool {
	// Text           <-    ('\\' . / (!quote .))*
	return p.addNode(func() bool {
		return p.ZeroOrMore(func() bool {
			return p.NeedOne([]func() bool{
				func() bool {
					return p.NeedAll([]func() bool{
						func() bool {
							return p.Next("\\")
						},
						func() bool {
							return p.AnyChar()
						},
					})
				},
				func() bool {
					return p.NeedAll([]func() bool{
						func() bool {
							return p.Not(func() bool {
								return p.quote()
							})
						},
						func() bool {
							return p.AnyChar()
						},
					})
				},
			})
		})
	}, "Text")
}

func (p *JSONParser) Integer() bool {
	// Integer        <-    '-'? [0-9]+
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.Maybe(func() bool {
					return p.Next("-")
				})
			},
			func() bool {
				return p.OneOrMore(func() bool {
					return p.InRange('0', '9')
				})
			},
		})
	}, "Integer")
}

func (p *JSONParser) Float() bool {
	// Float          <-    '-'? [0-9]* '.' [0-9]+ 'E'? '-'? [0-9]*
	return p.addNode(func() bool {
		return p.NeedAll([]func() bool{
			func() bool {
				return p.Maybe(func() bool {
					return p.Next("-")
				})
			},
			func() bool {
				return p.ZeroOrMore(func() bool {
					return p.InRange('0', '9')
				})
			},
			func() bool {
				return p.Next(".")
			},
			func() bool {
				return p.OneOrMore(func() bool {
					return p.InRange('0', '9')
				})
			},
			func() bool {
				return p.Maybe(func() bool {
					return p.Next("E")
				})
			},
			func() bool {
				return p.Maybe(func() bool {
					return p.Next("-")
				})
			},
			func() bool {
				return p.ZeroOrMore(func() bool {
					return p.InRange('0', '9')
				})
			},
		})
	}, "Float")
}

func (p *JSONParser) Boolean() bool {
	// Boolean        <-    'true' / 'false'
	return p.addNode(func() bool {
		return p.NeedOne([]func() bool{
			func() bool {
				return p.Next("true")
			},
			func() bool {
				return p.Next("false")
			},
		})
	}, "Boolean")
}

func (p *JSONParser) spacing() bool {
	// spacing        <-    [ \t\n\r]+
	return p.OneOrMore(func() bool {
		return p.InSet(" \t\n\r")
	})
}

func (p *JSONParser) EndOfFile() bool {
	// EndOfFile      <-    !.
	return p.addNode(func() bool {
		return p.Not(func() bool {
			return p.AnyChar()
		})
	}, "EndOfFile")
}
