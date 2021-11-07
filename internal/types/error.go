package types

type UnknownTypeError struct {
	Type string
}

func (e *UnknownTypeError) Error() string {
	return "unknown type: " + e.Type
}
