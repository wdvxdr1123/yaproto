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

type BuiltinType struct {
	name   string
	gotype string
	wire   string
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

func (b BuiltinType) WireType() Wire {
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

var BuiltinTypes = [...]BuiltinType{
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

func (m *Message) GoType() string {
	return CamelCase(m.Name)
}

type Flag uint

const (
	FOptional Flag = 1<<iota - 1
	FRepeated
	FRequired

	// FPtr mark a pointer field.
	FPtr
)

func (f Flag) Is(mask Flag) bool {
	return f&mask != 0
}

func (f *Flag) Set(mask Flag, value bool) {
	if value {
		*f |= mask
	} else {
		*f &^= mask
	}
}

type Field struct {
	Flag

	Type     Type
	Name     string
	Sequence int
}

func (f *Field) IsRepeated() bool {
	return f.Is(FRepeated)
}

func (f *Field) IsPtr() bool {
	return f.Is(FPtr)
}

func (f *Field) GoName() string {
	return CamelCase(f.Name)
}

func (f *Field) GoType() string {
	return f.Type.GoType()
}

// ftype is type of the field in struct field definition
func (f *Field) ftype() (s string) {
	switch f.Type.Scope() {
	case SMessage:
		s = "*" + f.GoType()
	case SBuiltin:
		if f.IsPtr() {
			s = "*" + f.GoType()
		} else {
			s = f.GoType()
		}
	default:
		panic("unreachable")
	}
	if f.Flag == FRepeated {
		s = "[]" + s
	}
	return
}

// rtype is return type of the field
func (f *Field) rtype() string {
	if f.Flag == FRepeated {
		return f.ftype()
	}
	switch f.Type.Scope() {
	case SMessage:
		return "*" + f.GoType()
	case SBuiltin:
		return f.GoType()
	}
	panic("unreachable")
}

func (f *Field) DefaultValue() string {
	if f.Flag&FRepeated != 0 {
		return "nil"
	}
	switch f.Type.Scope() {
	case SBuiltin:
		switch f.Type.Name() {
		case "int32", "uint32", "int64", "uint64":
			return "0"
		case "float32", "float64":
			return "0.0"
		case "bool":
			return "false"
		case "bytes":
			return "nil"
		case "string":
			return `""`
		}
	case SMessage:
		return "nil"
	}
	panic("unreachable")
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
	case BuiltinType:
		return t.WireType()
	case *MessageType:
		return WireBytes
	default:
		panic("unreachable")
	}
}
