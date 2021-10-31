package main

import (
	"os"

	"github.com/emicklei/proto"

	"yaproto/internal/generator"
)

func main() {
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
	g.Generate(os.Stdout)
}
