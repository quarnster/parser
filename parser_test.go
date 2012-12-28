package parser

import (
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
