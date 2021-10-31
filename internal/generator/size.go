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
			g.Pf("n += %d*len(x.%s)\n", size, field.GoName())
		} else {
			g.Pln("n +=", size)
		}
	}
	switch typ.WireType() {
	case WireVarint:
		switch typ.Name() {
		case "uint64":
			if repeated {
				g.Pf("for _,e := range x.%s {\n", field.GoName())
				g.Pf("    n += %d + proto.VarintSize(e)\n", ks)
				g.Pln("}")
			} else if g.proto2() {
				g.Pf("if x.%s != nil {\n", field.GoName())
				g.Pf("n += %d + proto.VarintSize(*x.%s)\n", ks, field.GoName())
				g.Pln("}")
			}

		case "bool":
			fixed(ks+1, field)

		default:
			if repeated {
				g.Pf("for _,e := range x.%s {\n", field.GoName())
				g.Pf("    n += %d + proto.VarintSize(uint64(e))\n", ks)
				g.Pln("}")
			} else {
				g.Pf("if x.%s != nil {\n", field.GoName())
				g.Pf("n += %d + proto.VarintSize(uint64(*x.%s))\n", ks, field.GoName())
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
			g.Pf("l = len(x.%s)\n", field.GoName())
			g.Pln("if l>0 {\n")
			g.Pf("    n += %d + proto.VarintSize(uint64(l)) + l\n", ks)
			g.Pln("}")
		} else {
			g.Pf("if x.%s != nil {\n", field.GoName())
			if g.isptr(field) {
				g.Pf("    l = len(*x.%s)\n", field.GoName())
			} else {
				g.Pf("    l = len(x.%s)\n", field.GoName())
			}
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
