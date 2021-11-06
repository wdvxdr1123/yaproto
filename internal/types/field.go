package types

import (
	"github.com/wdvxdr1123/yaproto/internal/utils"
)

type Flag uint

const (
	FOptional Flag = 1 << iota
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

type MessageField struct {
	Flag

	Type     Type
	Name     string
	Sequence int
}

func (f *MessageField) IsRepeated() bool {
	return f.Is(FRepeated)
}

func (f *MessageField) IsPtr() bool {
	return f.Is(FPtr)
}

func (f *MessageField) GoName() string {
	return utils.CamelCase(f.Name)
}

func (f *MessageField) GoType() string {
	return f.Type.GoType()
}

// Ftype is type of the field in struct field definition
func (f *MessageField) Ftype() (s string) {
	switch f.Type.Scope() {
	case CMessage:
		s = "*" + f.GoType()
	case CScalar, CEnum:
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

// Rtype is return type of the field
func (f *MessageField) Rtype() string {
	if f.Flag == FRepeated {
		return f.Ftype()
	}
	switch f.Type.Scope() {
	case CMessage:
		return "*" + f.GoType()
	case CScalar, CEnum:
		return f.GoType()
	}
	panic("unreachable")
}

func (f *MessageField) Elem() *MessageField {
	if f.IsPtr() {
		nf := *f
		nf.Set(FPtr, false)
		return &nf
	}
	return f
}

// null returns the null value of the field.
func (f *MessageField) Null() string {
	if f.IsRepeated() || f.IsPtr() {
		return "nil"
	}
	switch f.Type.Scope() {
	case CScalar:
		switch f.Type.Name() {
		case "int32", "uint32", "int64", "uint64", "sint32", "sint64":
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
	case CMessage:
		return "nil"
	case CEnum:
		return "0"
	}
	panic("unreachable")
}

func (f *MessageField) Selector(deref bool) string {
	x := "x." + f.GoName()
	if deref && f.Type.Scope() != CMessage && f.IsPtr() {
		x = "*" + x
	}
	return x
}

// Conv converts the filed to dst Type.
func (f *MessageField) Conv(dst Type) string {
	return Convert(f.Selector(true), f.Type, dst)
}

type EnumField struct {
	Name  string
	Value int
}

func (e *EnumField) GoName() string {
	return utils.CamelCase(e.Name)
}
