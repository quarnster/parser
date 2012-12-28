package parser

import (
	"container/list"
	"io/ioutil"
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	var p PegParser
	if data, err := ioutil.ReadFile("./peg.peg"); err != nil {
		t.Fatalf("%s", err)
	} else {
		p.data = strings.NewReader(string(data))
		if !p.Grammar() {
			t.Fatalf("Didn't parse correctly")
		} else {
			t.Log(p.currentNode)
		}
	}
}

func BenchmarkParser(b *testing.B) {
	var p PegParser
	if data, err := ioutil.ReadFile("./peg.peg"); err != nil {
		b.Fatalf("%s", err)
	} else {
		p.data = strings.NewReader(string(data))
		for i := 0; i < b.N; i++ { //use b.N for looping
			p.data.Seek(0, 0)
			p.stack.Clear()
			p.currentNode.Children = list.List{}
			p.Grammar()
		}
	}
}
