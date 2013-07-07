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

import (
	"fmt"
)

type (
	Parser interface {
		SetData(data string)
		Reset()
		RootNode() *Node
		Error() Error
	}

	Error interface {
		Line() int
		Column() int
		Description() string
		Error() string
	}

	BasicError struct {
		line        int
		column      int
		description string
	}
	Reader interface {
		Len() int
		Pos() int
		Read() rune
		UnRead()
		LineCol(offset int) (line, col int)
		Substring(start, end int) string
		Seek(offset int)
	}
)

func NewError(line, column int, description string) Error {
	return &BasicError{line, column, description}
}
func (be *BasicError) Error() string {
	return fmt.Sprintf("%d,%d: %s", be.line, be.column, be.description)
}
func (be *BasicError) Line() int           { return be.line }
func (be *BasicError) Column() int         { return be.column }
func (be *BasicError) Description() string { return be.description }
