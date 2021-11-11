package generator

import (
	"fmt"

	"github.com/wdvxdr1123/yaproto/internal/types"
	"github.com/wdvxdr1123/yaproto/internal/utils"
)

func (g *Generator) marshal(m *types.Message) {
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
	_ = i`, utils.CamelCase(m.Name), utils.CamelCase(m.Name))
	g.Pln()

	for _, field := range m.Fields {
		g.marshalField(field)
	}

	g.Pln("    return i")
	g.Pln("}")
	g.Pln()
}

func (g *Generator) marshalField(f *types.MessageField) {
	wt := types.WireType(f.Type)

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

	body := func(name string, t types.Type) {
		// value
		switch wt {
		default:
			panic(fmt.Errorf("unhandled wire type: %d", wt))

		case types.WireVarint:
			if f.Type.Name() == "bool" {
				g.Pf("proto.PutBool(buf, &i, %s)\n", name)
			} else {
				g.Pf("proto.PutVarint(buf, &i, %s)\n", types.Convert(name, t, types.ScalarValueTypes[types.TUINT64]))
			}

		case types.WireBytes:
			if f.Type.ScopeClass() == types.CMessage {
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
		g.Pf("    for _, e := range %s {\n", f.Selector(true))
		key(types.KeyValue(f.Sequence, types.WireBytes))
		body("e", f.Type)
		g.Pf("    }\n")
	} else {
		g.Pf("if %s != %s {\n", f.Selector(false), f.Null())
		key(types.KeyValue(f.Sequence, wt))
		body(f.Selector(true), f.Type)
		g.Pln("}")
	}
}
