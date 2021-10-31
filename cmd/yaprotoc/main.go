package main

import (
	"flag"
	"os"

	"github.com/emicklei/proto"

	"yaproto/internal/generator"
)

func main() {
	getter := flag.Bool("getter", false, "generate getter for message")
	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer file.Close()

	defination, err := proto.NewParser(file).Parse()
	if err != nil {
		panic(err)
	}

	g := generator.New(defination)
	g.Options.GenGetter = *getter
	g.Generate(os.Stdout)
}
