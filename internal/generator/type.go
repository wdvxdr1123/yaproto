package generator

type Scope uint8

const (
	SBuiltin Scope = iota
	SMessage
	SEnum
)

var _, _, _ = SBuiltin, SMessage, SEnum

type Type interface {
	GoType() string
	Name() string
	Scope() Scope
}

type ScalarValueType struct {
	name   string
	gotype string
	wire   string
}

func (b ScalarValueType) GoType() string {
	return b.gotype
}

func (b ScalarValueType) Name() string {
	return b.name
}

func (b ScalarValueType) Scope() Scope {
	return SBuiltin
}

var _ Type = ScalarValueType{}
var _ Type = &ScalarValueType{}

func (b ScalarValueType) WireType() Wire {
	switch b.wire {
	case "fixed32":
		return WireFixed32
	case "fixed64":
		return WireFixed64
	case "varint", "zigzag32", "zigzag64":
		return WireVarint
	case "bytes":
		return WireBytes
	}
	panic("unknown wire type")
}

const (
	TUINT64 = 3
)

var ScalarValueTypes = [...]ScalarValueType{
	{"int32", "int32", "varint"},
	{"uint32", "uint32", "varint"},
	{"int64", "int64", "varint"},
	{"uint64", "uint64", "varint"},
	{"sint32", "int32", "zigzag32"},
	{"sint64", "int64", "zigzag64"},
	{"float32", "float32", "fixed32"},
	{"float64", "float64", "fixed64"},
	{"bool", "bool", "varint"},
	{"bytes", "[]byte", "bytes"},
	{"string", "string", "bytes"},
}

type MessageType struct {
	name string
	def  *Message
}

func (m *MessageType) GoType() string {
	return CamelCase(m.name)
}

func (m *MessageType) Name() string {
	return m.name
}

func (m *MessageType) Scope() Scope {
	return SMessage
}

type EnumType struct {
	name string
	def  *Enum
}

func (e *EnumType) GoType() string {
	return CamelCase(e.name)
}

func (e *EnumType) Name() string {
	return e.name
}

func (e *EnumType) Scope() Scope {
	return SEnum
}

func (g *Generator) typ(t string) Type {
	for _, bt := range ScalarValueTypes {
		if bt.Name() == t {
			return bt
		}
	}
	obj := g.lookup(t)
	switch obj := obj.Obj.(type) {
	case *Message:
		return &MessageType{name: obj.Name, def: obj}
	case *Enum:
		return &EnumType{name: obj.Name, def: obj}
	case nil:
		panic("nil type")
	default:
		panic("unknown type")
	}
}

type Wire uint8

const (
	WireVarint Wire = iota
	WireFixed64
	WireBytes
	WireStartGroup
	WireEndGroup
	WireFixed32
)

func (w Wire) String() string {
	switch w {
	case WireVarint:
		return "varint"
	case WireFixed64:
		return "fixed64"
	case WireBytes:
		return "bytes"
	case WireFixed32:
		return "fixed32"
	}
	panic("unreachable")
}

func wire(t Type) Wire {
	switch t := t.(type) {
	case ScalarValueType:
		return t.WireType()
	case *MessageType:
		return WireBytes
	default:
		panic("unreachable")
	}
}
