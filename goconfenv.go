package main

import "flag"
import "fmt"
import "os"

var env *string
var cemConfPath *string
var cemSvcsHost *string
var projectName *string

func main() {
	env = flag.String("e", "prod", "config env")
	cemConfPath = flag.String("c", "cem.conf", "render config file")
	cemSvcsHost = flag.String("h", "localhost:8002/api/variableList", "cem service host")
	projectName = flag.String("p", "tte", "project name, must set!")

	flag.Parse()

	if *projectName == "" {
		fmt.Printf("Please set project name!\n")
		flag.PrintDefaults()
		return
	}

	cs := NewCemsvcs(*cemSvcsHost, *env)
	cc := NewCemConf(cs)
	cc.ParseConf(*cemConfPath)

	resSlice := cc.Render()

	allSucc := true
	for _, r := range resSlice {
		r.Print()
		if !r.Succ {
			allSucc = false
		}
	}
	if !allSucc {
		os.Exit(-1)
	}
}
