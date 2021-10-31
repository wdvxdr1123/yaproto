package generator

func (g *Generator) marshal(m *Message) {
	g.Pf("func (x *%s) Marshal() ([]byte, error) {", m.Name)
	g.Pln("    if x == nil {")
	g.Pln("        return []byte{},nil")
	g.Pln("    }")
	g.Pln("    buf := make([]byte, x.Size())")

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

}

func (g *Generator) marshalMessage(field *Field, repeated bool) {

}
