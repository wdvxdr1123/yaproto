package generator

import (
	"fmt"
)

func (g *Generator) marshal(m *Message) {
	g.Pf(`func (x *%s) Marshal() []byte {
	size := x.Size()
	buf := make([]byte, size)
	n := x.MarshalTo(buf[:size])
	return buf[:n]
}

func (x *%s) MarshalTo(buf []byte) int {
	var i int
	_ = i`, CamelCase(m.Name), CamelCase(m.Name))
	g.Pln()

	for _, field := range m.Fields {
		g.marshalField(field)
	}

	g.Pln("    return i")
	g.Pln("}")
	g.Pln()
}

func (g *Generator) marshalField(field *Field) {
	wt := wire(field.Type)

	key := func(kv uint32) {
		for kv >= 0x80 {
			x := byte(kv) | 0x80
			g.Pf("buf[i] = 0x%x\n", x)
			g.Pf("i++\n")
			kv >>= 7
		}
		g.Pf("buf[i] = 0x%x\n", byte(kv))
		g.Pf("i++\n")
	}

	body := func(name string, t Type) {
		// value
		switch wt {
		default:
			panic(fmt.Errorf("unhandled wire type: %d", wt))

		case WireVarint:
			if field.Type.Name() == "bool" {
				g.Pf("proto.PutBool(buf, &i, %s)\n", name)
			} else {
				g.Pf("proto.PutVarint(buf, &i, %s)\n", conv(name, t, BuiltinTypes[TUINT64]))
			}

		case WireBytes:
			if field.Type.Scope() == SMessage {
				g.Pf("l := %s.Size()\n", name)
				g.Pf("proto.PutVarint(buf, &i, uint64(l))\n")
				g.Pf("i += %s.MarshalTo(buf[i:])\n", name)
			} else {
				g.Pf("proto.PutVarint(buf, &i, uint64(len(%s)))\n", name)
				g.Pf("i += copy(buf[i:], %s)\n", name)
			}
		}
	}

	if g.proto2() {
		if field.Option == FRepeated {
			g.Pf("    for _, e := range %s {\n", g.sel(field))
			key(keyValue(field.Sequence, WireBytes))
			body("e", field.Type)
			g.Pf("    }\n")
		} else {
			g.Pf("if x.%s != nil {\n", field.GoName())
			key(keyValue(field.Sequence, wt))
			body(g.sel(field), field.Type)
			g.Pln("}")
		}
	} else {

	}
}
