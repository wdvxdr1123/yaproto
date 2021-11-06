package generator

import (
	"sync"

	"github.com/emicklei/proto"
)

type Object struct {
	Name string
	Obj  obj
	Type Type

	once sync.Once
}

func (o *Object) GoType() string {
	switch obj := o.Obj.(type) {
	case *Message:
		return CamelCase(obj.Name)
	case *Enum:
		return CamelCase(obj.Name)
	default:
		return ""
	}
}

type obj interface {
	aObj()
}

type object struct{}

func (o *object) aObj() {}

type Message struct {
	object
	Name   string
	Fields []*MessageField
}

func (m *Message) GoType() string {
	return CamelCase(m.Name)
}

type Enum struct {
	object
	Name   string
	Fields []*EnumField
}

func (e *Enum) GoType() string {
	return CamelCase(e.Name)
}

func (g *Generator) lookup(name string) *Object {
	if m, ok := g.objects[name]; ok {
		return m
	}
	obj := new(Object)
	obj.Name = name
	g.objects[name] = obj
	return obj
}

func (g *Generator) lookupMessage(m *proto.Message) (msg *Message) {
	obj := g.lookup(m.Name)
	if obj.Obj != nil {
		msg = obj.Obj.(*Message)
	} else {
		msg = new(Message)
		obj.Obj = msg
	}
	return msg
}

func (g *Generator) lookupEnum(m *proto.Enum) (msg *Enum) {
	obj := g.lookup(m.Name)
	if obj.Obj != nil {
		msg = obj.Obj.(*Enum)
	} else {
		msg = new(Enum)
		obj.Obj = msg
	}
	return msg
}
