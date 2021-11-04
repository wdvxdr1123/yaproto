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

func (f *Field) Elem() *Field {
	if f.IsPtr() {
		nf := *f
		nf.Set(FPtr, false)
		return &nf
	}
	return f
}

// null returns the null value of the field.
func (f *Field) null() string {
	if f.IsRepeated() || f.IsPtr() {
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

func (f *Field) selector(deref bool) string {
	x := "x." + f.GoName()
	if deref && f.Type.Scope() != SMessage && f.IsPtr() {
		x = "*" + x
	}
	return x
}

// conv converts the filed to dst Type.
func (f *Field) conv(dst Type) string {
	return conv(f.selector(true), f.Type, dst)
}
