package generator

func (g *Generator) getter(m *Message) {
	for _, f := range m.Fields {
		if g.Options.GenGetter < 2 &&
			(!f.IsPtr() || f.Type.Scope() == SMessage) {
			// option requires not to generate getter for non-pointer fields.
			continue
		}

		g.Pln()
		g.Pf("func (x *%s) Get%s() %s {\n", m.GoType(), f.GoName(), f.rtype())
		if f.Type.Scope() != SMessage && f.IsPtr() {
			g.Pf("if x != nil && %s != nil {\n", f.selector(false))
		} else {
			g.Pln("if x != nil {")
		}
		g.Pf("        return %s\n", f.selector(true))
		g.Pln("    }")
		g.Pln("    return", f.Elem().null())
		g.Pln("}")
		g.Pln()
	}
}
