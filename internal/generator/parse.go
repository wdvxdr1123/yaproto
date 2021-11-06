package generator

import (
	"sort"
	"strings"

	"github.com/emicklei/proto"
)

func (g *Generator) parse() {
	for _, elem := range g.def.Elements {
		switch elem := elem.(type) {
		case *proto.Syntax:
			switch elem.Value {
			case "proto3":
				g.version = 3
			case "proto2":
				g.version = 2
			default:
				panic("unsupported syntax version")
			}
		case *proto.Option:
			if elem.Name == "go_package" {
				p := elem.Constant.Source
				g.gopackage = strings.TrimPrefix(p, "./;")
			}
		case *proto.Message:
			g.parseMessage(elem, g.universe)
		case *proto.Enum:
			g.parseEnum(elem, g.universe)
		}
	}
}

func (g *Generator) parseMessage(m *proto.Message, s *Scope) {
	msg := s.lookupMessage(m)
	msg.Name = m.Name
	scope := NewScope(s, m.Name)
	for _, field := range m.Elements {
		switch field := field.(type) {
		case *proto.NormalField:
			// the field type maybe not defined yet, so we should
			// look up the type later
			g.later(func() {
				f := &MessageField{
					Name:     field.Name,
					Sequence: field.Sequence,
					Type:     scope.typ(field.Type),
				}

				switch {
				case field.Repeated:
					f.Flag.Set(FRepeated, true)
				case field.Optional:
					f.Flag.Set(FOptional, true)
				case field.Required:
					f.Flag.Set(FRequired, true)
				}

				if !f.IsRepeated() {
					if (g.proto2() && f.Type.Name() != "bytes") ||
						(g.proto3() && f.Type.Scope() == CMessage) {
						f.Set(FPtr, true)
					}
				}

				msg.Fields = append(msg.Fields, f)
			})

		case *proto.Message:
			g.parseMessage(field, scope)
		}
	}
}

func (g *Generator) parseEnum(elem *proto.Enum, s *Scope) {
	enum := s.lookupEnum(elem)
	enum.Name = elem.Name
	for _, field := range elem.Elements {
		switch field := field.(type) {
		case *proto.EnumField:
			f := &EnumField{
				Name:  field.Name,
				Value: field.Integer,
			}
			enum.Fields = append(enum.Fields, f)
		}
	}
	sort.Slice(enum.Fields, func(i, j int) bool {
		return enum.Fields[i].Value < enum.Fields[j].Value
	})
}
