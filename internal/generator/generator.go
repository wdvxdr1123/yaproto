package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"sort"
	"strings"

	"github.com/emicklei/proto"
)

type Generator struct {
	def *proto.Proto

	buf bytes.Buffer

	version   int
	gopackage string
	messages  []*Message
}

func New(def *proto.Proto) *Generator {
	g := &Generator{
		def:     def,
		version: 2,
	}
	g.parse()

	sort.Slice(g.messages, func(i, j int) bool {
		return g.messages[i].Name < g.messages[j].Name
	})

	return g
}

func (g *Generator) P(args ...interface{}) { _, _ = fmt.Fprint(&g.buf, args...) }

func (g *Generator) Pln(args ...interface{}) { _, _ = fmt.Fprintln(&g.buf, args...) }

func (g *Generator) Pf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(&g.buf, format, args...)
}

func (g *Generator) Generate(out io.Writer) {
	g.generate()

	source, err := format.Source(g.buf.Bytes())
	if err != nil {
		fmt.Printf("%s", g.buf.Bytes())
		panic(err)
	}
	_, err = out.Write(source)
	if err != nil {
		panic(err)
	}
}

func (g *Generator) parse() {
	for _, elem := range g.def.Elements {
		switch elem := elem.(type) {
		case *proto.Syntax:
			switch elem.Value {
			case "proto3":
				g.version = 3
			case "proto2":
				g.version = 2
			default:
				panic("unsupported syntax")
			}
		case *proto.Option:
			if elem.Name == "go_package" {
				p := elem.Constant.Source
				g.gopackage = strings.TrimPrefix(p, "./;")
			}
		case *proto.Message:
			g.message(elem)
		}
	}
}

func (g *Generator) lookup(name string) *Message {
	for _, m := range g.messages {
		if m.Name == name {
			return m
		}
	}
	m := &Message{Name: name}
	g.messages = append(g.messages, m)
	return m
}

func (g *Generator) message(m *proto.Message) {
	msg := g.lookup(m.Name)
	for _, field := range m.Elements {
		switch field := field.(type) {
		case *proto.NormalField:
			f := &Field{
				Name:     field.Name,
				Sequence: field.Sequence,
				Type:     g.typ(field.Type),
				Option:   FNone,
			}

			switch {
			case field.Repeated:
				f.Option |= FRepeated
			case field.Optional:
				f.Option |= FOptional
			case field.Required:
				f.Option |= FRequired
			}
			msg.Fields = append(msg.Fields, f)
		case *proto.Message:
			panic("nested message not implemented")
		}
	}
}
func (g *Generator) generate() {
	g.Pln("// Code generated by yaprotoc. DO NOT EDIT.")
	g.Pln()
	g.Pln("package ", g.gopackage)
	g.Pln()

	// todo(wdvxdr): generate imports
	// g.Pln("import (")
	// g.Pln("\"fmt\"")
	// g.Pln(")")

	g.Pln()

	for _, m := range g.messages {
		g.Pln("type ", GoCamelCase(m.Name), " struct {")
		for _, f := range m.Fields {
			switch f.Option {
			case FNone:
				g.Pf("%s %s `protobuf:\"%d\"`\n", GoCamelCase(f.Name), f.Type.GoType(), f.Sequence)
			case FRepeated:
				g.Pf("%s []%s `protobuf:\"%d\"`\n", GoCamelCase(f.Name), f.Type.GoType(), f.Sequence)
			case FRequired:
				g.Pf("%s %s `protobuf:\"%d,req\"`\n", GoCamelCase(f.Name), f.Type.GoType(), f.Sequence)
			case FOptional:
				g.Pf("%s %s `protobuf:\"%d,opt\"`\n", GoCamelCase(f.Name), f.Type.GoType(), f.Sequence)
			}

		}
		g.Pln("}")
		g.Pln()
	}

}
