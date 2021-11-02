package generator

func (g *Generator) size(m *Message) {
	g.Pf("func (x *%s) Size() (n int) {\n", m.GoType())
	g.Pln("    if x == nil {")
	g.Pln("        return 0")
	g.Pln("    }")
	g.Pln("    var l int")
	g.Pln("    _ = l")

	for _, field := range m.Fields {
		repeated := field.Option == FRepeated
		switch field.Type.Scope() {
		case SBuiltin:
			g.sizeBuiltin(field, repeated)
		case SMessage:
			g.sizeMessage(field, repeated)
		}
	}

	g.Pln("return")
	g.Pln("}")
	g.Pln()
}

func (g *Generator) sizeBuiltin(field *Field, repeated bool) {
	typ := field.Type.(BuiltinType)
	ks := keySize(field.Sequence, typ.WireType())
	fixed := func(size int, field *Field) {
		if repeated {
			g.Pf("n += %d*len(%s)\n", size, g.sel(field))
		} else {
			g.Pf("if x.%s != nil {\n", field.GoName())
			g.Pln("n +=", size)
			g.Pln("}")
		}
	}
	switch typ.WireType() {
	case WireVarint:
		switch typ.Name() {
		case "bool":
			fixed(ks+1, field)

		default:
			if repeated {
				g.Pf("for _,e := range x.%s {\n", field.GoName())
				g.Pf("    n += %d + proto.VarintSize(%s)\n", ks, conv("e", field.Type, BuiltinTypes[TUINT64]))
				g.Pln("}")
			} else {
				g.Pf("if x.%s != nil {\n", field.GoName())
				g.Pf("n += %d + proto.VarintSize(%s)\n", ks, g.selConv(field, BuiltinTypes[TUINT64]))
				g.Pln("}")
			}
		}

	case WireFixed32:
		fixed(ks+4, field)

	case WireFixed64:
		fixed(ks+8, field)

	case WireBytes:
		if repeated {
			g.Pf("for _, b := range x.%s {\n", field.GoName())
			g.Pln("    l = len(b)")
			g.Pf("     n += %d + proto.VarintSize(uint64(l)) + l\n", ks)
			g.Pln("}")
		} else if g.proto3() {
			g.Pln("l = len(", g.sel(field), ")")
			g.Pln("if l>0 {\n")
			g.Pf("    n += %d + proto.VarintSize(uint64(l)) + l\n", ks)
			g.Pln("}")
		} else {
			g.Pf("if x.%s != nil {\n", field.GoName())
			g.Pln("    l = len(", g.sel(field), ")")
			g.Pf("    n += %d + proto.VarintSize(uint64(l)) + l\n", ks)
			g.Pln("}")
		}
	}
}

func (g *Generator) sizeMessage(field *Field, repeated bool) {
	ks := keySize(field.Sequence, WireBytes)
	if repeated {
		g.Pf("for _, e := range x.%s {\n", field.GoName())
		g.Pln("    l = e.Size()")
		g.Pf("     n += %d + proto.VarintSize(uint64(l)) + l\n", ks)
		g.Pln("}")
	} else {
		g.Pf("if e := x.%s;e != nil {\n", field.GoName())
		g.Pln("    l = e.Size()")
		g.Pf("    n += %d + proto.VarintSize(uint64(l)) + l\n", ks)
		g.Pln("}")
	}
}
