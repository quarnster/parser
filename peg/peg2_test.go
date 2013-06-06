package peg

import (
	"bytes"
	"github.com/quarnster/parser"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestPegs(t *testing.T) {
	ignore := []string{"Junk", "EndOfLine"}
	if f, err := os.Open("./testdata"); err != nil {
		t.Fatal(err)
	} else if fi, err := f.Readdir(-1); err != nil {
		t.Fatal(err)
	} else {
		for i := range fi {
			if n := "./testdata/" + fi[i].Name(); strings.HasSuffix(n, ".peg") {
				if data, err := ioutil.ReadFile(n); err != nil {
					t.Error("%s", err)
				} else {
					p := Peg{}
					if !p.Parse(string(data)) {
						t.Error("Didn't parse correctly", p.Error())
						continue
					}
					back := p.RootNode().Children[len(p.RootNode().Children)-1]
					if back.Name != "EndOfFile" {
						t.Log(p.RootNode())
						t.Error("File didn't finish parsing", p.Error())
						continue
					}
					root := fi[i].Name()
					root = root[:len(root)-4]
					gen := &parser.GoGenerator{RootNode: p.RootNode()}

					ignoreFunc := func(g parser.Generator, in string) string {
						return g.Ignore(in)
					}
					var customActions []parser.CustomAction
					for _, action := range ignore {
						customActions = append(customActions, parser.CustomAction{action, ignoreFunc})
					}
					gen.SetCustomActions(customActions)
					s := parser.GeneratorSettings{
						Name:       strings.ToTitle(root),
						Testname:   "../" + n + ".in",
						Debug:      true,
						DebugLevel: 0,
						WriteFile: func(name, data string) error {
							if err := os.Mkdir(root, 0755); err != nil && !os.IsExist(err) {
								return err
							}
							if err := ioutil.WriteFile(root+"/"+name, []byte(data), 0644); err != nil {
								return err
							}
							return nil
						},
					}
					if err := parser.GenerateParser(p.RootNode(), gen, s); err != nil {
						t.Error(err)
					} else {
						cmd := gen.TestCommand()
						c := exec.Command(cmd[0], cmd[1:]...)
						c.Dir = root
						data, err := c.CombinedOutput()
						if i := bytes.LastIndex(data, []byte("\nok ")); i > 0 {
							data = data[:i+1]
						}
						if i := bytes.Index(data, []byte{'='}); i > 0 {
							data = data[i:]
						}
						t.Log(string(data))
						if err != nil {
							t.Error(err)
						} else if exp, err := ioutil.ReadFile(n + ".out"); err != nil {
							t.Log("Unable to read expected output file, it'll be created")
							if err := ioutil.WriteFile(n+".out", data, 0644); err != nil {
								t.Error(err)
							}
						} else if d, err := diff(exp, data); err != nil || len(d) != 0 {
							t.Error(err, string(d))
						}
					}
					os.RemoveAll(root)
				}
			}
		}
	}
}
