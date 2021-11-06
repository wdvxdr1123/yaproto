package types

import "reflect"

type Class uint8

const (
	CScalar Class = iota
	CMessage
	CEnum
)

var _, _, _ = CScalar, CMessage, CEnum

type Type interface {
	GoType() string
	Name() string
	Scope() Class
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

func (b ScalarValueType) Scope() Class {
	return CScalar
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
	name   string
	gotype string
}

func (m *MessageType) GoType() string {
	return m.gotype
}

func (m *MessageType) Name() string {
	return m.name
}

func (m *MessageType) Scope() Class {
	return CMessage
}

type EnumType struct {
	name   string
	gotype string
	def    *Enum
}

func (e *EnumType) GoType() string {
	return e.gotype
}

func (e *EnumType) Name() string {
	return e.name
}

func (e *EnumType) Scope() Class {
	return CEnum
}

type ImportedType struct {
	TypeName string
	Gotype   string
}

func (i *ImportedType) Scope() Class {
	return CMessage
}

func (i *ImportedType) GoType() string {
	return i.Gotype
}

func (i *ImportedType) Name() string {
	return i.TypeName
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

func WireType(t Type) Wire {
	switch t := t.(type) {
	case ScalarValueType:
		return t.WireType()
	case *MessageType:
		return WireBytes
	default:
		panic("unreachable")
	}
}

func WireString(t Type) string {
	switch t := t.(type) {
	case ScalarValueType:
		return t.wire
	case *MessageType, *ImportedType:
		return "bytes"
	case *EnumType:
		return "varint"
	}
	panic("unknown wire" + reflect.TypeOf(t).String())
}

func Convert(x string, src, dst Type) string {
	if dst != src {
		x = dst.GoType() + "(" + x + ")"
	}
	return x
}

func KeySize(fieldNumber int, wire Wire) int {
	x := KeyValue(fieldNumber, wire)
	size := 0
	for size = 0; x > 127; size++ {
		x >>= 7
	}
	size++
	return size
}

func KeyValue(fieldNumber int, wire Wire) uint32 {
	return uint32(fieldNumber)<<3 | uint32(wire)
}
