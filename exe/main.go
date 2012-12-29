package main

import (
	"io/ioutil"
	"log"
	"os"
	"parser"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <peg file>\n", os.Args[0])
	}
	p := parser.PegParser{}
	if data, err := ioutil.ReadFile(os.Args[1]); err != nil {
		log.Fatalf("%s", err)
	} else {
		p.SetData(string(data))
		if !p.Grammar() {
			log.Fatalf("Didn't parse correctly\n")
		} else {
			log.Println(parser.GenerateParser(p.RootNode(), &parser.GoGenerator{Name: "Test"}))
		}
	}
}
