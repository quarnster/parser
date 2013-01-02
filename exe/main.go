package main

import (
	"flag"
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
	var (
		pegfile  = ""
		testfile = ""
		bench    = false
		debug    = false
		dumptree = false
		notest   = false
		ignore   = ""
	)
	flag.StringVar(&ignore, "ignore", ignore, "List of definitions to ignore (not generate nodes for)")
	flag.StringVar(&pegfile, "peg", pegfile, "Pegfile for which to generate a parser for")
	flag.StringVar(&testfile, "testfile", testfile, "Test file to be used in testing")
	flag.BoolVar(&bench, "bench", bench, "Whether to run a benchmark test or not")
	flag.BoolVar(&debug, "debug", debug, "Whether to make the generated parser spit out debug info")
	flag.BoolVar(&dumptree, "dumptree", dumptree, "Whether to make the generated parser spit out the generated tree")
	flag.BoolVar(&notest, "notest", notest, "Whether to run the tests after generating the parser")
	flag.Parse()
	if pegfile == "" || testfile == "" {
		flag.Usage()
		os.Exit(1)
	}
	p := peg.Peg{}
	if data, err := ioutil.ReadFile(pegfile); err != nil {
		log.Fatalf("%s", err)
	} else {
		p.SetData(string(data))
		if !p.Parse() {
			log.Fatalf("Didn't parse correctly\n")
		} else {
			log.Println(p.RootNode())
			dumptree_s := ""
			if dumptree {
				dumptree_s = "log.Println(root)"
			}
			back := p.RootNode().Children[len(p.RootNode().Children)-1]
			if back.Name != "EndOfFile" {
				log.Println("File didn't finish parsing")
			}
			name := filepath.Base(pegfile)
			name = name[:len(name)-len(filepath.Ext(name))]
			if err := os.Mkdir(name, 0755); err != nil && !os.IsExist(err) {
				log.Fatalln(err)
			}
			var abs string
			if abs, err = filepath.Abs(testfile); err != nil {
				log.Fatalln(err)
			}

			gen := parser.GoGenerator2{}
			ignoreFunc := func(g parser.Generator, in string) string {
				return gen.Ignore(in)
			}
			var customActions []parser.CustomAction
			for _, action := range strings.Split(ignore, ",") {
				action = strings.TrimSpace(action)
				customActions = append(customActions, parser.CustomAction{action, ignoreFunc})
			}

			gen = parser.GoGenerator2{
				Name:            strings.ToTitle(name),
				AddDebugLogging: debug,
				CustomActions:   customActions,
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
const testname = "`+abs+`"
func TestParser(t *testing.T) {
	var p `+gen.Name+`
	if data, err := ioutil.ReadFile(testname); err != nil {
		log.Fatalf("%s", err)
	} else {
		p.SetData(string(data))
		if !p.Parse() {
			t.Fatalf("Didn't parse correctly: %s\n", p.Error())
		} else {
			root := p.RootNode()
			`+dumptree_s+`
			if root.Range.End != len(p.ParserData.Data) {
				t.Fatalf("Parsing didn't finish: %v", root)
			}
		}
	}
}

func BenchmarkParser(b *testing.B) {
	var p `+gen.Name+`
	if data, err := ioutil.ReadFile(testname); err != nil {
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
			if !notest {
				cmd := []string{"go", "test"}
				if bench {
					cmd = append(cmd, "-bench", "Parser")
				}
				c := exec.Command(cmd[0], cmd[1:]...)
				c.Dir = name
				data, _ := c.CombinedOutput()
				log.Println(string(data))
			}
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
