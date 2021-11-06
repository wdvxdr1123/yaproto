package main

import (
	"flag"
	"os"

	"github.com/emicklei/proto"

	"github.com/wdvxdr1123/yaproto/internal/generator"
)

func main() {
	getter := flag.Bool("getter", false, "generate getter methods")
	size := flag.Bool("size", false, "generate size methods")
	marshal := flag.Bool("marshal", false, "generate marshal/unmarshal methods")
	flag.Parse()
	file, err := os.Open(flag.Args()[0])
	if err != nil {
		panic(err)
	}
	defer file.Close()

	defination, err := proto.NewParser(file).Parse()
	if err != nil {
		panic(err)
	}
	defination.Filename = flag.Args()[0]

	g := generator.New(defination)
	g.Options.GenGetter = *getter
	g.Options.GenMarshal = *marshal
	g.Options.GenSize = *size || *marshal
	g.Generate(os.Stdout)
}
