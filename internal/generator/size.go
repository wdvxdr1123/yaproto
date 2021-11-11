package generator

import (
	"github.com/wdvxdr1123/yaproto/internal/types"
)

func (g *Generator) size(m *types.Message) {
	g.Pf("func (x *%s) Size() (n int) {\n", m.GoType())
	g.Pln("    if x == nil {")
	g.Pln("        return 0")
	g.Pln("    }")
	g.Pln("    var l int")
	g.Pln("    _ = l")

	for _, field := range m.Fields {
		switch field.Type.ScopeClass() {
		case types.CScalar:
			g.sizeBuiltin(field)
		case types.CMessage:
			g.sizeMessage(field)
		}
	}

	g.Pln("return")
	g.Pln("}")
	g.Pln()
}

func (g *Generator) sizeBuiltin(f *types.MessageField) {
	typ := f.Type.(types.ScalarValueType)
	ks := types.KeySize(f.Sequence, typ.WireType())
	fixed := func(size int, field *types.MessageField) {
		if field.IsRepeated() {
			g.Pf("n += %d*len(%s)\n", size, field.Selector(true))
		} else if g.proto2() {
			g.Pf("if x.%s != nil {\n", field.GoName())
			g.Pln("n +=", size)
			g.Pln("}")
		} else {
			if field.Type.GoType() == "bool" {
				g.Pf("if %s {\n", field.Selector(true))
			} else {
				g.Pf("if %s != %s {\n", field.Selector(true), field.Null())
			}
			g.Pln("n +=", size)
			g.Pln("}")
		}
	}
	switch typ.WireType() {
	case types.WireVarint:
		switch typ.Name() {
		case "bool":
			fixed(ks+1, f)

		default:
			if f.IsRepeated() {
				g.Pf("for _,e := range %s {\n", f.Selector(false))
				g.Pf("    n += %d + proto.VarintSize(%s)\n", ks, types.Convert("e", f.Type, types.ScalarValueTypes[types.TUINT64]))
				g.Pln("}")
			} else {
				g.Pf("if %s != %s {\n", f.Selector(false), f.Null())
				g.Pf("n += %d + proto.VarintSize(%s)\n", ks, f.Conv(types.ScalarValueTypes[types.TUINT64]))
				g.Pln("}")
			}
		}

	case types.WireFixed32:
		fixed(ks+4, f)

	case types.WireFixed64:
		fixed(ks+8, f)

	case types.WireBytes:
		if f.IsRepeated() {
			g.Pf("for _, b := range %s {\n", f.Selector(false))
			g.Pln("    l = len(b)")
			g.Pf("     n += %d + proto.VarintSize(uint64(l)) + l\n", ks)
			g.Pln("}")
		} else if g.proto3() {
			g.Pln("l = len(", f.Selector(true), ")")
			g.Pln("if l>0 {")
			g.Pf("    n += %d + proto.VarintSize(uint64(l)) + l\n", ks)
			g.Pln("}")
		} else {
			g.Pf("if %s != nil {\n", f.Selector(false))
			g.Pln("    l = len(", f.Selector(true), ")")
			g.Pf("    n += %d + proto.VarintSize(uint64(l)) + l\n", ks)
			g.Pln("}")
		}
	}
}

func (g *Generator) sizeMessage(f *types.MessageField) {
	ks := types.KeySize(f.Sequence, types.WireBytes)
	if f.IsRepeated() {
		g.Pf("for _, e := range x.%s {\n", f.GoName())
		g.Pln("    l = e.Size()")
		g.Pf("     n += %d + proto.VarintSize(uint64(l)) + l\n", ks)
		g.Pln("}")
	} else {
		g.Pf("if e := %s;e != nil {\n", f.Selector(false))
		g.Pln("    l = e.Size()")
		g.Pf("    n += %d + proto.VarintSize(uint64(l)) + l\n", ks)
		g.Pln("}")
	}
}
