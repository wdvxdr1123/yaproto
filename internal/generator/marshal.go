package generator

func (g *Generator) marshal(m *Message) {
	g.Pf("func (x *%s) Marshal() ([]byte, error) {\n", m.Name)
	g.Pln("    if x == nil {")
	g.Pln("        return []byte{},nil")
	g.Pln("    }")
	g.Pln("    buf := make([]byte, x.Size())")
	g.Pln("    var i int")
	g.Pln("    _ = i")

	for _, field := range m.Fields {
		repeated := field.Option == FRepeated
		switch field.Type.Scope() {
		case SBuiltin:
			g.marshalBuiltin(field, repeated)
		case SMessage:
			g.marshalMessage(field, repeated)
		}
	}

	g.Pln("    return buf, nil")
	g.Pln("}")
	g.Pln()
}

func (g *Generator) marshalBuiltin(field *Field, repeated bool) {
	wt := wire(field.Type)
	ks := keySize(field.Sequence, wt)
	kv := keyValue(field.Sequence, wt)

	key := func() {
		switch ks {
		case 1:
			g.Pf("buf[i] = %d\n", kv)
			g.Pf("i++\n")
		default:
			// todo: expand uvarint
			g.Pf("proto.PutVarint(buf, &i, %d)\n", kv)
		}
	}

	body := func() {
		key()
		switch wt {
		case WireVarint:
			if field.Type.Name() == "bool" {
				g.Pf("proto.PutBool(buf, &i, %s)\n", g.selConv(field, BuiltinTypes[TBOOL]))
			} else {
				g.Pf("proto.PutVarint(buf, &i, %s)\n", g.selConv(field, BuiltinTypes[TUINT64]))
			}

		case WireBytes:
			if field.Type.Scope() == SMessage {
				// g.Pf("proto.PutVarint(buf, &i, uint64(len(%s)))\n", g.selConv(field, BuiltinTypes[TSTRING]))
				// g.Pf("i += copy(buf[i:], %s)\n", g.sel(field))
			} else {
				g.Pf("i += copy(buf[i:], %s)\n", g.sel(field))
			}
		}
	}

	if g.proto2() {
		if repeated {
			field.Option = FNone
			// todo(wdvxdr): implement
			field.Option = FRepeated
		} else {
			g.Pf("if x.%s != nil {\n", field.GoName())
			body()
			g.Pln("}")
		}
	} else {

	}
}

func (g *Generator) marshalMessage(field *Field, repeated bool) {

}
