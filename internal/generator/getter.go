package generator

import "github.com/wdvxdr1123/yaproto/internal/types"

func (g *Generator) getter(m *types.Message) {
	for _, f := range m.Fields {
		if g.Options.GenGetter < 2 &&
			(!f.IsPtr() || f.Type.Scope() == types.CMessage) {
			// option requires not to generate getter for non-pointer fields.
			continue
		}

		g.Pln()
		g.Pf("func (x *%s) Get%s() %s {\n", m.GoType(), f.GoName(), f.Rtype())
		if f.Type.Scope() != types.CMessage && f.IsPtr() {
			g.Pf("if x != nil && %s != nil {\n", f.Selector(false))
		} else {
			g.Pln("if x != nil {")
		}
		g.Pf("        return %s\n", f.Selector(true))
		g.Pln("    }")
		g.Pln("    return", f.Elem().Null())
		g.Pln("}")
		g.Pln()
	}
}
