package generator

type Scope uint8

const (
	SBuiltin Scope = iota
	SMessage
	SEnum
)

type Type interface {
	GoType() string
	Name() string
	Scope() Scope
}

type BuiltinType struct {
	name   string
	gotype string
}

func (b BuiltinType) GoType() string {
	return b.gotype
}

func (b BuiltinType) Name() string {
	return b.name
}

func (b BuiltinType) Scope() Scope {
	return SBuiltin
}

var BuiltinTypes = [...]BuiltinType{
	{"int8", "int8"},
	{"uint8", "uint8"},
	{"int16", "int16"},
	{"uint16", "uint16"},
	{"int32", "int32"},
	{"uint32", "uint32"},
	{"int64", "int64"},
	{"uint64", "uint64"},
	{"float32", "float32"},
	{"float64", "float64"},
	{"bool", "bool"},
	{"bytes", "[]byte"},
	{"string", "string"},
}

type MessageType struct {
	name string
	def  *Message
}

func (m *MessageType) GoType() string {
	return "*" + GoCamelCase(m.name)
}

func (m *MessageType) Name() string {
	return m.name
}

func (m *MessageType) Scope() Scope {
	return SMessage
}

func (g *Generator) typ(t string) Type {
	for _, bt := range BuiltinTypes {
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

type Field struct {
	Type     Type
	Name     string
	Sequence int
}
