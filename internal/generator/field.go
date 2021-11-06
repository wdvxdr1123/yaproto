package generator

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
	return CamelCase(f.Name)
}

func (f *MessageField) GoType() string {
	return f.Type.GoType()
}

// ftype is type of the field in struct field definition
func (f *MessageField) ftype() (s string) {
	switch f.Type.Scope() {
	case SMessage:
		s = "*" + f.GoType()
	case SBuiltin, SEnum:
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
func (f *MessageField) rtype() string {
	if f.Flag == FRepeated {
		return f.ftype()
	}
	switch f.Type.Scope() {
	case SMessage:
		return "*" + f.GoType()
	case SBuiltin, SEnum:
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
func (f *MessageField) null() string {
	if f.IsRepeated() || f.IsPtr() {
		return "nil"
	}
	switch f.Type.Scope() {
	case SBuiltin:
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
	case SMessage:
		return "nil"
	case SEnum:
		return "0"
	}
	panic("unreachable")
}

func (f *MessageField) selector(deref bool) string {
	x := "x." + f.GoName()
	if deref && f.Type.Scope() != SMessage && f.IsPtr() {
		x = "*" + x
	}
	return x
}

// conv converts the filed to dst Type.
func (f *MessageField) conv(dst Type) string {
	return conv(f.selector(true), f.Type, dst)
}

type EnumField struct {
	Name  string
	Value int
}

func (e *EnumField) GoName() string {
	return CamelCase(e.Name)
}
