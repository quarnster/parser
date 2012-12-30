package main

import (
	"io/ioutil"
	"log"
	"os"
	"parser"
	"runtime/pprof"
)

func main() {
	if data, err := ioutil.ReadFile(os.Args[1]); err != nil {
		log.Fatalf("%s", err)
	} else {
		if f, err := os.Create("./cpu.txt"); err != nil {
			log.Fatalln("%s", err)
		} else {
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
			for i := 0; i < 100; i++ {
				p := parser.JSONParser{}
				p.SetData(string(data))
				if !p.JsonFile() {
					log.Fatalf("Didn't parse correctly\n")
				}
			}
		}
	}
}
