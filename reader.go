package parser

import (
	"unicode/utf8"
)

type BasicReader struct {
	pos  int
	data string
}

const nilrune = '\u0000'

func (p *BasicReader) Substring(start, end int) string {
	l := len(p.data)
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
	return string(p.data[start:end])
}

func (p *BasicReader) LineCol(offset int) (line, column int) {
	line = 1
	column = 1
	for _, r := range p.data[:offset] {
		column++
		if r == '\n' {
			line++
			column = 1
		}
	}
	return
}

func (p *BasicReader) Len() int {
	return len(p.data)
}

func (p *BasicReader) Pos() int {
	return p.pos
}
func (p *BasicReader) eof() bool {
	return p.pos >= len(p.data)
}
func (p *BasicReader) Read() rune {
	if p.eof() {
		p.pos++
		return nilrune
	}
	r, s := utf8.DecodeRuneInString(p.data[p.pos:])
	p.pos += s

	return r
}

func (p *BasicReader) UnRead() {
	p.pos--
	for !p.eof() && p.pos > 0 && !utf8.RuneStart(p.data[p.pos]) {
		p.pos--
	}
}

func (p *BasicReader) Seek(n int) {
	p.pos = n
}

func NewReader(data string) Reader {
	return &BasicReader{0, data}
}
