package types

import (
	"github.com/emicklei/proto"

	"github.com/wdvxdr1123/yaproto/internal/utils"
)

type Scope struct {
	name     string
	parent   *Scope
	Children []*Scope

	Elems map[string]*Object
}

func NewScope(parent *Scope, name string) *Scope {
	s := &Scope{
		parent: parent,
		name:   utils.CamelCase(name),
		Elems:  make(map[string]*Object),
	}
	if parent != nil {
		parent.Children = append(parent.Children, s)
	}
	return s
}

func (s *Scope) Lookup(name string) *Object {
	m, ok := s.LookupOK(name)
	if ok {
		return m
	}
	obj := new(Object)
	obj.Name = name
	s.Elems[name] = obj
	return obj
}

func (s *Scope) LookupOK(name string) (*Object, bool) {
	m, ok := s.Elems[name]
	return m, ok
}

func (s *Scope) LookupParent(name string) (*Scope, *Object) {
	for ; s != nil; s = s.parent {
		if obj, ok := s.LookupOK(name); ok {
			return s, obj
		}
	}
	return nil, nil
}

func (s *Scope) Type(t string) (Type, error) {
	for _, bt := range ScalarValueTypes {
		if bt.Name() == t {
			return bt, nil
		}
	}

	scope, obj := s.LookupParent(t)
	if obj == nil {
		return nil, &UnknownTypeError{Type: t}
	}

	switch obj := obj.Obj.(type) {
	case *Message:
		return &MessageType{name: obj.Name, gotype: scope.ResolveName(obj.Name)}, nil
	case *Enum:
		return &EnumType{name: obj.Name, gotype: scope.ResolveName(obj.Name)}, nil
	default:
		return nil, &UnknownTypeError{Type: t}
	}
}

func (s *Scope) LookupMessage(m *proto.Message) (msg *Message) {
	obj := s.Lookup(m.Name)
	if obj.Obj != nil {
		msg = obj.Obj.(*Message)
	} else {
		msg = new(Message)
		msg.scope = s
		obj.Obj = msg
	}
	return msg
}

func (s *Scope) LookupEnum(m *proto.Enum) (msg *Enum) {
	obj := s.Lookup(m.Name)
	if obj.Obj != nil {
		msg = obj.Obj.(*Enum)
	} else {
		msg = new(Enum)
		obj.Obj = msg
	}
	return msg
}

func (s *Scope) ResolveName(name string) string {
	scope := s
	var prefixes []string
	for scope.parent != nil {
		if scope.name != "" {
			prefixes = append(prefixes, scope.name)
		}
		scope = scope.parent
	}

	var prefix string
	for i := len(prefixes) - 1; i >= 0; i-- {
		prefix += prefixes[i] + "_"
	}

	if prefix != "" {
		return prefix + utils.CamelCase(name)
	}
	return utils.CamelCase(name)
}
