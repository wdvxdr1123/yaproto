package generator

import (
	"fmt"
)

func (g *Generator) marshal(m *Message) {
	if g.Options.GenMarshal == 1 {
		g.Pf(`func (x *%s) Marshal() ([]byte, error) {
	if x == nil {
		return nil, errors.New("nil message")
	}
	return proto.Marshal(x)
}`, m.GoType())
		g.Pln()
		return
	}

	g.Pf(`func (x *%s) Marshal() ([]byte, error) {
	size := x.Size()
	buf := make([]byte, size)
	n := x.MarshalTo(buf[:size])
	return buf[:n], nil
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

func (g *Generator) marshalField(f *MessageField) {
	wt := wire(f.Type)

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
			if f.Type.Name() == "bool" {
				g.Pf("proto.PutBool(buf, &i, %s)\n", name)
			} else {
				g.Pf("proto.PutVarint(buf, &i, %s)\n", conv(name, t, ScalarValueTypes[TUINT64]))
			}

		case WireBytes:
			if f.Type.Scope() == SMessage {
				g.Pf("l := %s.Size()\n", name)
				g.Pf("proto.PutVarint(buf, &i, uint64(l))\n")
				g.Pf("i += %s.MarshalTo(buf[i:])\n", name)
			} else {
				g.Pf("proto.PutVarint(buf, &i, uint64(len(%s)))\n", name)
				g.Pf("i += copy(buf[i:], %s)\n", name)
			}
		}
	}

	if f.IsRepeated() {
		g.Pf("    for _, e := range %s {\n", f.selector(true))
		key(keyValue(f.Sequence, WireBytes))
		body("e", f.Type)
		g.Pf("    }\n")
	} else {
		g.Pf("if %s != %s {\n", f.selector(false), f.null())
		key(keyValue(f.Sequence, wt))
		body(f.selector(true), f.Type)
		g.Pln("}")
	}
}
