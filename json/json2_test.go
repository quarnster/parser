package json

import (
	"testing"
)

/*
The json test input data in the map below comes from the Jansson library
with the following license:

Copyright (c) 2009-2012 Petri Lehtinen <petri@digip.org>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

var tests = map[string]string{
	`[1E+2]
`: `0-7: "JSON"
	0-6: "Array"
		1-5: "Float" - Data: "1E+2"
	7-7: "EndOfFile" - Data: ""
`,
	`["\uD834\uDD1E surrogate, four-byte UTF-8"]
`: `0-44: "JSON"
	0-43: "Array"
		2-41: "Text" - Data: "\uD834\uDD1E surrogate, four-byte UTF-8"
	44-44: "EndOfFile" - Data: ""
`,
	`[true]
`: `0-7: "JSON"
	0-6: "Array"
		1-5: "Boolean" - Data: "true"
	7-7: "EndOfFile" - Data: ""
`,
	`[{}]
`: `0-5: "JSON"
	0-4: "Array"
		1-3: "Dictionary" - Data: "{}"
	5-5: "EndOfFile" - Data: ""
`,
	`{}
`: `0-3: "JSON"
	0-2: "Dictionary" - Data: "{}"
	3-3: "EndOfFile" - Data: ""
`,
	`[1E-2]
`: `0-7: "JSON"
	0-6: "Array"
		1-5: "Float" - Data: "1E-2"
	7-7: "EndOfFile" - Data: ""
`,
	`[]
`: `0-3: "JSON"
	0-2: "Array" - Data: "[]"
	3-3: "EndOfFile" - Data: ""
`,
	`[-123]
`: `0-7: "JSON"
	0-6: "Array"
		1-5: "Integer" - Data: "-123"
	7-7: "EndOfFile" - Data: ""
`,
	`[1,2,3,4,
"a", "b", "c",
{"foo": "bar", "core": "dump"},
true, false, true, true, null, false
]
`: `0-96: "JSON"
	0-95: "Array"
		1-2: "Integer" - Data: "1"
		3-4: "Integer" - Data: "2"
		5-6: "Integer" - Data: "3"
		7-8: "Integer" - Data: "4"
		11-12: "Text" - Data: "a"
		16-17: "Text" - Data: "b"
		21-22: "Text" - Data: "c"
		25-55: "Dictionary"
			26-37: "KeyValuePair"
				27-30: "Text" - Data: "foo"
				34-37: "Text" - Data: "bar"
			40-53: "KeyValuePair"
				41-45: "Text" - Data: "core"
				49-53: "Text" - Data: "dump"
		57-61: "Boolean" - Data: "true"
		63-68: "Boolean" - Data: "false"
		70-74: "Boolean" - Data: "true"
		76-80: "Boolean" - Data: "true"
		82-86: "Null" - Data: "null"
		88-93: "Boolean" - Data: "false"
	96-96: "EndOfFile" - Data: ""
`,
	`[false]
`: `0-8: "JSON"
	0-7: "Array"
		1-6: "Boolean" - Data: "false"
	8-8: "EndOfFile" - Data: ""
`,
	`[1e+2]
`: `0-7: "JSON"
	0-6: "Array"
		1-5: "Float" - Data: "1e+2"
	7-7: "EndOfFile" - Data: ""
`,
	`["\u0012 escaped control character"]
`: `0-37: "JSON"
	0-36: "Array"
		2-34: "Text" - Data: "\u0012 escaped control character"
	37-37: "EndOfFile" - Data: ""
`,
	`[0]
`: `0-4: "JSON"
	0-3: "Array"
		1-2: "Integer" - Data: "0"
	4-4: "EndOfFile" - Data: ""
`,
	`{"a":[]}
`: `0-9: "JSON"
	0-8: "Dictionary"
		1-7: "KeyValuePair"
			2-3: "Text" - Data: "a"
			5-7: "Array" - Data: "[]"
	9-9: "EndOfFile" - Data: ""
`,
	`[-1]
`: `0-5: "JSON"
	0-4: "Array"
		1-3: "Integer" - Data: "-1"
	5-5: "EndOfFile" - Data: ""
`,
	`[1E22]
`: `0-7: "JSON"
	0-6: "Array"
		1-5: "Float" - Data: "1E22"
	7-7: "EndOfFile" - Data: ""
`,
	`["\u0821 three-byte UTF-8"]
`: `0-28: "JSON"
	0-27: "Array"
		2-25: "Text" - Data: "\u0821 three-byte UTF-8"
	28-28: "EndOfFile" - Data: ""
`,
	`["€þıœəßð some utf-8 ĸʒ×ŋµåäö𝄞"]
`: `0-52: "JSON"
	0-51: "Array"
		2-49: "Text" - Data: "€þıœəßð some utf-8 ĸʒ×ŋµåäö𝄞"
	52-52: "EndOfFile" - Data: ""
`,
	`[1]
`: `0-4: "JSON"
	0-3: "Array"
		1-2: "Integer" - Data: "1"
	4-4: "EndOfFile" - Data: ""
`,
	`[123]
`: `0-6: "JSON"
	0-5: "Array"
		1-4: "Integer" - Data: "123"
	6-6: "EndOfFile" - Data: ""
`,
	`["\"\\\/\b\f\n\r\t"]
`: `0-21: "JSON"
	0-20: "Array"
		2-18: "Text" - Data: "\"\\\/\b\f\n\r\t"
	21-21: "EndOfFile" - Data: ""
`,
	`[123.456e78]
`: `0-13: "JSON"
	0-12: "Array"
		1-11: "Float" - Data: "123.456e78"
	13-13: "EndOfFile" - Data: ""
`,
	`[-0]
`: `0-5: "JSON"
	0-4: "Array"
		1-3: "Integer" - Data: "-0"
	5-5: "EndOfFile" - Data: ""
`,
	`["\u0123 two-byte UTF-8"]
`: `0-26: "JSON"
	0-25: "Array"
		2-23: "Text" - Data: "\u0123 two-byte UTF-8"
	26-26: "EndOfFile" - Data: ""
`,
	`[123.456789]
`: `0-13: "JSON"
	0-12: "Array"
		1-11: "Float" - Data: "123.456789"
	13-13: "EndOfFile" - Data: ""
`,
	`[""]
`: `0-5: "JSON"
	0-4: "Array"
		2-2: "Text" - Data: ""
	5-5: "EndOfFile" - Data: ""
`,
	`[1e-2]
`: `0-7: "JSON"
	0-6: "Array"
		1-5: "Float" - Data: "1e-2"
	7-7: "EndOfFile" - Data: ""
`,
	`["a"]
`: `0-6: "JSON"
	0-5: "Array"
		2-3: "Text" - Data: "a"
	6-6: "EndOfFile" - Data: ""
`,
	`["\u002c one-byte UTF-8"]
`: `0-26: "JSON"
	0-25: "Array"
		2-23: "Text" - Data: "\u002c one-byte UTF-8"
	26-26: "EndOfFile" - Data: ""
`,
	`["abcdefghijklmnopqrstuvwxyz1234567890 "]
`: `0-42: "JSON"
	0-41: "Array"
		2-39: "Text" - Data: "abcdefghijklmnopqrstuvwxyz1234567890 "
	42-42: "EndOfFile" - Data: ""
`,
	`[123e45]
`: `0-9: "JSON"
	0-8: "Array"
		1-7: "Float" - Data: "123e45"
	9-9: "EndOfFile" - Data: ""
`,
	`[null]
`: `0-7: "JSON"
	0-6: "Array"
		1-5: "Null" - Data: "null"
	7-7: "EndOfFile" - Data: ""
`,
	`[123e-10000000]
`: `0-16: "JSON"
	0-15: "Array"
		1-14: "Float" - Data: "123e-10000000"
	16-16: "EndOfFile" - Data: ""
`}

var invalid = map[string]string{`['
`: `1,2: Unexpected '`,
	`aå
`: `1,1: Unexpected a`,
	`{,
`: `1,2: Unexpected ,`,
	`[,
`: `1,2: Unexpected ,`,
	`[1,
`: `2,1: Unexpected EOF`,
	``: `1,1: Unexpected EOF`,
	`[1,]
`: `1,4: Unexpected ]`,
	`[1,
2,
3,
4,
5,
]
`: `6,1: Unexpected ]`,
	`[1,2,3]
foo
`: `2,1: Unexpected f`,
	`[1,2,3]foo
`: `1,8: Unexpected f`,
	`[012]
`: `1,5: Unexpected ]`,
	`[troo
`: `1,2: Unexpected t`,
	`[-123foo]
`: `1,6: Unexpected f`,
	`[-123.123foo]
`: `1,10: Unexpected f`,
	`{
`: `2,1: Unexpected EOF`,
	`[
`: `2,1: Unexpected EOF`,
	`[-foo]
`: `1,3: Unexpected f`,
	`[-012]
`: `1,6: Unexpected ]`,
	`{'a'
`: `1,2: Unexpected '`,
	`{"a":"a" 123}
`: `1,10: Unexpected 1`,
	`[{}
`: `2,1: Unexpected EOF`,
	`{"a"
`: `1,5: Unexpected new line`,
	`{"a":
`: `2,1: Unexpected EOF`,
	`{"a":"a
`: `2,1: Unexpected EOF`,
	`[1ea]
`: `1,4: Unexpected a`,
	`[1e]
`: `1,4: Unexpected ]`,
	`[1.]
`: `1,4: Unexpected ]`,
	`å
`: `1,1: Unexpected å`,
	`["a"
`: `2,1: Unexpected EOF`,
	`[{
`: `2,1: Unexpected EOF`,
	`{"
`: `2,1: Unexpected EOF`,
	`{"a
`: `2,1: Unexpected EOF`,
	`{[
`: `1,2: Unexpected [`,
	`["a
`: `2,1: Unexpected EOF`}

func TestParserComprehensive(t *testing.T) {
	for k, v := range tests {
		var p JSON
		if p.Parse(k) != true {
			t.Fatalf("Didn't parse correctly: %s", k)
		} else if p.RootNode().String() != v {
			t.Logf("Test %s failed\nExpected: %s\nReceived, %s", k, v, p.RootNode())
			t.Fail()
		}
	}
	for k, v := range invalid {
		var p JSON
		if p.Parse(k) {
			root := p.RootNode()
			if root.Children[len(root.Children)-1].Name == "EndOfFile" {
				t.Logf("Succeeded, but shouldn't have: %s", k)
				t.Fail()
			}
		}
		if p.Error().Error() != v {
			t.Logf("Test %s failed\nExpected: %s\nReceived, %s", k, v, p.Error().Error())
			t.Fail()
		}
	}
}
