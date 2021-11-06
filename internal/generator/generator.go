package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"sort"
	"strconv"

	"github.com/emicklei/proto"
)

type GoPackage struct {
	Path  string
	Alias string
}

type Generator struct {
	def *proto.Proto

	buf *bytes.Buffer

	version   int
	gopackage string
	objects   map[string]*Object
	goImport  map[GoPackage]bool

	delayed []func()

	Options struct {
		GenGetter  bool
		GenSize    bool
		GenMarshal bool
	}
}

func New(def *proto.Proto) *Generator {
	g := &Generator{
		def:      def,
		version:  2,
		objects:  make(map[string]*Object),
		goImport: make(map[GoPackage]bool),
	}
	g.parse()

	for len(g.delayed) > 0 {
		g.delayed[0]()
		g.delayed = g.delayed[1:]
	}

	return g
}

func (g *Generator) later(f func()) {
	g.delayed = append(g.delayed, f)
}

func (g *Generator) Pln(args ...interface{}) { _, _ = fmt.Fprintln(g.buf, args...) }
func (g *Generator) Pf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(g.buf, format, args...)
}

func (g *Generator) Generate(out io.Writer) {
	body := new(bytes.Buffer)
	g.generate(body)
	buf := new(bytes.Buffer)
	g.header(buf)
	_, _ = io.Copy(buf, body)

	source, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Printf("%s", buf.Bytes())
		panic(err)
	}
	_, err = out.Write(source)
	if err != nil {
		panic(err)
	}
}

func (g *Generator) header(buffer *bytes.Buffer) {
	g.buf = buffer
	g.Pf("// Code generated by yaprotoc. DO NOT EDIT.\n")
	g.Pf("// source: %s\n", g.def.Filename)
	g.Pf("\n")
	g.Pf("package %s\n", g.gopackage)
	g.Pf("\n")

	if len(g.goImport) > 0 {
		g.Pf("import (\n")
		var imports []GoPackage
		for p, ok := range g.goImport {
			if ok {
				imports = append(imports, p)
			}
		}
		sort.Slice(imports, func(i, j int) bool {
			return imports[i].Path < imports[j].Path
		})
		for _, p := range imports {
			g.Pf("%s \"%s\"", p.Alias, strconv.Quote(p.Path))
		}
		g.Pln("}\n")
	}
}

func (g *Generator) generate(buffer *bytes.Buffer) {
	g.buf = buffer

	// todo(wdvxdr): sort by name
	for _, obj := range g.objects {
		switch obj := obj.Obj.(type) {
		case *Message:
			sort.Slice(obj.Fields, func(i, j int) bool {
				return obj.Fields[i].Sequence < obj.Fields[j].Sequence
			})
			g.generateMessage(obj)
		case *Enum:
			g.generateEnum(obj)
		}
	}
}

func (g *Generator) generateMessage(m *Message) {
	g.Pln("type ", m.GoType(), " struct {")
	for _, f := range m.Fields {
		switch {
		case f.Is(FRepeated):
			g.Pf("%s %s `protobuf:\"%s,%d,rep\"`\n", f.GoName(), f.ftype(), wireString(f.Type), f.Sequence)
		default:
			g.Pf("%s %s `protobuf:\"%s,%d,opt\"`\n", f.GoName(), f.ftype(), wireString(f.Type), f.Sequence)
		}
	}
	g.Pln("}")
	g.Pln()

	if g.Options.GenGetter {
		g.getter(m)
	}
	if g.Options.GenSize {
		g.size(m)
	}
	if g.Options.GenMarshal {
		g.marshal(m)
	}
}

func (g *Generator) generateEnum(enum *Enum) {
	g.Pf("type %s int32\n", enum.GoType())
	g.Pln()
	if len(enum.Fields) > 0 {
		g.Pln("const (\n")
		for _, field := range enum.Fields {
			g.Pf("%s_%s %s = %d\n", enum.GoType(), field.Name, enum.GoType(), field.Value)
		}
		g.Pln(")\n")
	}

	g.Pf(`func (x %s) Enum() *%s {
	p := new(%s)
	*p = x
	return p
}`, enum.GoType(), enum.GoType(), enum.GoType())
	g.Pln()
}
