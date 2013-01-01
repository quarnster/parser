package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"parser"
	"parser/peg"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <peg file> <parser input file>\n", os.Args[0])
	}
	p := peg.Peg{}
	if data, err := ioutil.ReadFile(os.Args[1]); err != nil {
		log.Fatalf("%s", err)
	} else {
		p.SetData(string(data))
		if !p.Parse() {
			log.Fatalf("Didn't parse correctly\n")
		} else {
			log.Println(p.RootNode())
			back := p.RootNode().Children[len(p.RootNode().Children)-1]
			if back.Name != "EndOfFile" {
				log.Println("File didn't finish parsing")
			}
			name := filepath.Base(os.Args[1])
			name = name[:len(name)-len(filepath.Ext(name))]
			if err := os.Mkdir(name, 0755); err != nil && !os.IsExist(err) {
				log.Fatalln(err)
			}
			var abs string
			if abs, err = filepath.Abs(os.Args[2]); err != nil {
				log.Fatalln(err)
			}

			gen := parser.GoGenerator2{}
			ignore := func(g parser.Generator, in string) string {
				return gen.Ignore(in)
			}

			gen = parser.GoGenerator2{
				Name:            strings.ToTitle(name),
				AddDebugLogging: false,
				CustomActions: []parser.CustomAction{
					{"Spacing", ignore},
					{"Quote", ignore},
					{"QuotedText", ignore},
					{"Value", ignore},
					{"Values", ignore},
					{"KeyValuePairs", ignore},
				},
			}
			if err := ioutil.WriteFile(filepath.Join("./"+name, name+".go"), []byte(parser.GenerateParser(p.RootNode(), &gen)), 0644); err != nil {
				log.Fatalln(err)
			}

			if err := ioutil.WriteFile(filepath.Join("./"+name, name+"_test.go"), []byte(`package `+name+`
import (
	"io/ioutil"
	"log"
	"testing"
)

func TestParser(t *testing.T) {
	var p `+gen.Name+`
	if data, err := ioutil.ReadFile("`+abs+`"); err != nil {
		log.Fatalf("%s", err)
	} else {
		p.SetData(string(data))
		if !p.Parse() {
			t.Fatalf("Didn't parse correctly\n")
		} else {
			root := p.RootNode()
//			log.Println(root)
			if root.Range.End != len(p.ParserData.Data) {
				t.Fatalf("Parsing didn't finish: %v", root)
			}
		}
	}
}

func BenchmarkParser(b *testing.B) {
	var p `+gen.Name+`
	if data, err := ioutil.ReadFile("`+abs+`"); err != nil {
		b.Fatalf("%s", err)
	} else {
		p.SetData(string(data))
		for i := 0; i < b.N; i++ { //use b.N for looping
			p.Reset()
			p.Parse()
		}
	}
}


`), 0644); err != nil {
				log.Fatalln(err)
			}
			//			log.Println(filepath.Join(dir, "testmain.go"))
			c := exec.Command("go", "test", "-bench", ".")
			c.Dir = name
			data, _ := c.CombinedOutput()
			log.Println(string(data))
			// r1, err := c.StderrPipe()
			// log.Println(err)
			// log.Println(c.Start())
			// go func() {
			// 	r := bufio.NewReader(r1)
			// 	for {
			// 		if s, err := r.ReadString('\n'); err != nil {
			// 			break
			// 		} else {
			// 			log.Print(s)
			// 		}
			// 	}
			// }()
			// log.Println(c.Wait())
		}
	}
}
