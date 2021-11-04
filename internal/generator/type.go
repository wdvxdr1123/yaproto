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

func (b ScalarValueType) WireType() Wire {
	switch b.wire {
	case "fixed32":
		return WireFixed32
	case "fixed64":
		return WireFixed64
	case "varint":
		return WireVarint
	case "bytes":
		return WireBytes
	}
	panic("unknown wire type")
}

const (
	TINT32 = iota
	TUINT32
	TINT64
	TUINT64
	TFLOAT32
	TFLOAT64
	TBOOL
	TBYTES
	TSTRING
)

var ScalarValueTypes = [...]ScalarValueType{
	{"int32", "int32", "varint"},
	{"uint32", "uint32", "varint"},
	{"int64", "int64", "varint"},
	{"uint64", "uint64", "varint"},
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

func (g *Generator) typ(t string) Type {
	for _, bt := range ScalarValueTypes {
		if bt.Name() == t {
			return bt
		}
	}
	m := g.lookup(t)
	return &MessageType{name: t, def: m}
}

type Message struct {
	Name   string
	Fields []*Field
}

func (m *Message) GoType() string {
	return CamelCase(m.Name)
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
