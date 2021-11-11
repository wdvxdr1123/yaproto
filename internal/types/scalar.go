package types

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

func (b ScalarValueType) ScopeClass() Class {
	return CScalar
}

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
	{"fixed32", "uint32", "fixed32"},
	{"fixed64", "uint64", "fixed64"},
	{"bool", "bool", "varint"},
	{"bytes", "[]byte", "bytes"},
	{"string", "string", "bytes"},
}
