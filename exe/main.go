package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"parser"
	"path/filepath"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <peg file> <parser input file>\n", os.Args[0])
	}
	p := parser.PegParser{}
	if data, err := ioutil.ReadFile(os.Args[1]); err != nil {
		log.Fatalf("%s", err)
	} else {
		p.SetData(string(data))
		if !p.Grammar() {
			log.Fatalf("Didn't parse correctly\n")
		} else {
			back := p.RootNode().Children.Back().Value.(*parser.Node)
			if back.Name != "EndOfFile" {
				log.Println("File didn't finish parsing")
				log.Println(p.RootNode())
			}
			parserEntry := "DoesntExist"
			for n := p.RootNode().Children.Front(); n != nil; n = n.Next() {
				node := n.Value.(*parser.Node)
				if node.Name == "Definition" {
					parserEntry = node.Children.Front().Value.(*parser.Node).Data()
					break
				}
			}
			dir, err := ioutil.TempDir("", "parser")
			if err != nil {
				log.Fatalln(err)
			}
			defer os.RemoveAll(dir)
			if err := ioutil.WriteFile(filepath.Join(".", "testparser.go"), []byte(parser.GenerateParser(p.RootNode(), &parser.GoGenerator{
				Name:            "Test",
				AddDebugLogging: true,
			})), 0644); err != nil {
				log.Fatalln(err)
			}

			if err := ioutil.WriteFile(filepath.Join(dir, "testmain.go"), []byte(`package main

import (
	"io/ioutil"
	"log"
	"parser"
)
func main() {
	p := parser.Test{}
	if data, err := ioutil.ReadFile("`+os.Args[2]+`"); err != nil {
		log.Fatalf("%s", err)
	} else {
		p.SetData(string(data))
		if !p.`+parserEntry+`() {
			log.Fatalf("Didn't parse correctly\n")
		} else {
			log.Println(p.RootNode())
		}
	}
}

`), 0644); err != nil {
				log.Fatalln(err)
			}
			log.Println(filepath.Join(dir, "testmain.go"))
			c := exec.Command("go", "run", filepath.Join(dir, "testmain.go"))
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
