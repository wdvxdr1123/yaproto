package types

import (
	"github.com/wdvxdr1123/yaproto/internal/utils"
)

type Object struct {
	Name string
	Obj  protoObject
}

func (o *Object) GoType() string {
	switch obj := o.Obj.(type) {
	case *Message:
		return utils.CamelCase(obj.Name)
	case *Enum:
		return utils.CamelCase(obj.Name)
	default:
		return ""
	}
}

type protoObject interface {
	aObj()
}

type object struct{}

func (o *object) aObj() {}

type Message struct {
	object
	scope *Scope

	Name   string
	Fields []*MessageField
}

func (m *Message) GoType() string {
	return m.scope.ResolveName(m.Name)
}

type Enum struct {
	object
	Name   string
	Fields []*EnumField
}

func (e *Enum) GoType() string {
	return utils.CamelCase(e.Name)
}
