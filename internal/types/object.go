package types

import (
	"text/scanner"

	"github.com/wdvxdr1123/yaproto/internal/utils"
)

type Object struct {
	Obj protoObject
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
	Pos() scanner.Position
}

type object struct {
	pos scanner.Position
}

func (o *object) aObj() {}

func (o *object) SetPos(pos scanner.Position) {
	o.pos = pos
}

func (o *object) Pos() scanner.Position {
	return o.pos
}

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
