package generator

type Object struct {
	Name string
	Obj  protoObject
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
	return m.scope.resolveName(m.Name)
}

type Enum struct {
	object
	Name   string
	Fields []*EnumField
}

func (e *Enum) GoType() string {
	return CamelCase(e.Name)
}

func (g *Generator) Lookup(name string) *Object {
	return g.universe.Lookup(name)
}
