package generator

import "github.com/emicklei/proto"

type Scope struct {
	name     string
	parent   *Scope
	children []*Scope

	elems map[string]*Object
}

func NewScope(parent *Scope, name string) *Scope {
	s := &Scope{
		parent: parent,
		name:   CamelCase(name),
		elems:  make(map[string]*Object),
	}
	if parent != nil {
		parent.children = append(parent.children, s)
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
	s.elems[name] = obj
	return obj
}

func (s *Scope) LookupOK(name string) (*Object, bool) {
	m, ok := s.elems[name]
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

func (s *Scope) typ(t string) Type {
	for _, bt := range ScalarValueTypes {
		if bt.Name() == t {
			return bt
		}
	}

	scope, obj := s.LookupParent(t)

	switch obj := obj.Obj.(type) {
	case *Message:
		return &MessageType{name: obj.Name, gotype: scope.resolveName(obj.Name), def: obj}
	case *Enum:
		return &EnumType{name: obj.Name, gotype: scope.resolveName(obj.Name), def: obj}
	case nil:
		panic("nil type: " + t)
	default:
		panic("unknown type")
	}
}

func (s *Scope) lookupMessage(m *proto.Message) (msg *Message) {
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

func (s *Scope) lookupEnum(m *proto.Enum) (msg *Enum) {
	obj := s.Lookup(m.Name)
	if obj.Obj != nil {
		msg = obj.Obj.(*Enum)
	} else {
		msg = new(Enum)
		obj.Obj = msg
	}
	return msg
}

func (s *Scope) resolveName(name string) string {
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
		return prefix + CamelCase(name)
	}
	return CamelCase(name)
}
