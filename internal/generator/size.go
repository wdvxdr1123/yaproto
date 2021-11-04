package generator

func (g *Generator) size(m *Message) {
	g.Pf("func (x *%s) Size() (n int) {\n", m.GoType())
	g.Pln("    if x == nil {")
	g.Pln("        return 0")
	g.Pln("    }")
	g.Pln("    var l int")
	g.Pln("    _ = l")

	for _, field := range m.Fields {
		switch field.Type.Scope() {
		case SBuiltin:
			g.sizeBuiltin(field)
		case SMessage:
			g.sizeMessage(field)
		}
	}

	g.Pln("return")
	g.Pln("}")
	g.Pln()
}

func (g *Generator) sizeBuiltin(f *Field) {
	typ := f.Type.(ScalarValueType)
	ks := keySize(f.Sequence, typ.WireType())
	fixed := func(size int, field *Field) {
		if field.IsRepeated() {
			g.Pf("n += %d*len(%s)\n", size, field.selector(true))
		} else if g.proto2() {
			g.Pf("if x.%s != nil {\n", field.GoName())
			g.Pln("n +=", size)
			g.Pln("}")
		} else {
			if field.Type.GoType() == "bool" {
				g.Pf("if %s {\n", field.selector(true))
			} else {
				g.Pf("if %s != %s {\n", field.selector(true), field.null())
			}
			g.Pln("n +=", size)
			g.Pln("}")
		}
	}
	switch typ.WireType() {
	case WireVarint:
		switch typ.Name() {
		case "bool":
			fixed(ks+1, f)

		default:
			if f.IsRepeated() {
				g.Pf("for _,e := range %s {\n", f.selector(false))
				g.Pf("    n += %d + proto.VarintSize(%s)\n", ks, conv("e", f.Type, ScalarValueTypes[TUINT64]))
				g.Pln("}")
			} else {
				g.Pf("if %s != %s {\n", f.selector(false), f.null())
				g.Pf("n += %d + proto.VarintSize(%s)\n", ks, f.conv(ScalarValueTypes[TUINT64]))
				g.Pln("}")
			}
		}

	case WireFixed32:
		fixed(ks+4, f)

	case WireFixed64:
		fixed(ks+8, f)

	case WireBytes:
		if f.IsRepeated() {
			g.Pf("for _, b := range %s {\n", f.selector(false))
			g.Pln("    l = len(b)")
			g.Pf("     n += %d + proto.VarintSize(uint64(l)) + l\n", ks)
			g.Pln("}")
		} else if g.proto3() {
			g.Pln("l = len(", f.selector(true), ")")
			g.Pln("if l>0 {")
			g.Pf("    n += %d + proto.VarintSize(uint64(l)) + l\n", ks)
			g.Pln("}")
		} else {
			g.Pf("if %s != nil {\n", f.selector(false))
			g.Pln("    l = len(", f.selector(true), ")")
			g.Pf("    n += %d + proto.VarintSize(uint64(l)) + l\n", ks)
			g.Pln("}")
		}
	}
}

func (g *Generator) sizeMessage(f *Field) {
	ks := keySize(f.Sequence, WireBytes)
	if f.IsRepeated() {
		g.Pf("for _, e := range x.%s {\n", f.GoName())
		g.Pln("    l = e.Size()")
		g.Pf("     n += %d + proto.VarintSize(uint64(l)) + l\n", ks)
		g.Pln("}")
	} else {
		g.Pf("if e := %s;e != nil {\n", f.selector(false))
		g.Pln("    l = e.Size()")
		g.Pf("    n += %d + proto.VarintSize(uint64(l)) + l\n", ks)
		g.Pln("}")
	}
}
