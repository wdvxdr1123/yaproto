package main

import (
	"flag"
	"os"
	"strings"

	"github.com/emicklei/proto"

	"github.com/wdvxdr1123/yaproto/internal/generator"
)

func main() {
	output := flag.String("o", "", "output file")
	getter := flag.Int("getter", 1, "generate getter methods")
	size := flag.Bool("size", false, "generate size methods")
	marshal := flag.Int("marshal", 1, "generate marshal/unmarshal methods")
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
	g.Options.GenSize = *size || *marshal > 1

	var out *os.File
	if *output == "" {
		fname := file.Name()
		*output = strings.TrimSuffix(fname, ".proto") + ".pb.go"
	}
	out, err = os.OpenFile(*output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	g.Generate(out)
}
