package importer

import (
	"sort"
	"strings"

	"github.com/emicklei/proto"

	"github.com/wdvxdr1123/yaproto/internal/types"
	"github.com/wdvxdr1123/yaproto/internal/utils"
)

func (file *File) parseMessage(m *proto.Message, s *types.Scope) {
	msg := s.LookupMessage(m)
	msg.Name = m.Name
	msg.SetPos(m.Position)
	scope := types.NewScope(s, m.Name)
	for _, field := range m.Elements {
		switch field := field.(type) {
		case *proto.NormalField:
			// the field type maybe not defined yet, so we should
			// look up the type later
			file.later(func() {
				f := &types.MessageField{
					Name:     field.Name,
					Sequence: field.Sequence,
				}

				if strings.Contains(field.Type, ".") {
					t := field.Type
					dot := strings.LastIndexByte(t, '.')
					pkg := file.lookup(t[:dot])
					if pkg == nil {
						file.errorf(field.Position, "unknown package: %s", t[:dot])
						return
					}
					gotype := pkg.GoPackage + "." + utils.CamelCase(t[dot+1:])
					f.Type = &types.ImportedType{TypeName: t, Gotype: gotype}
				} else {
					typ, err := scope.Type(field.Type)
					if err != nil && file.Package.Scope != nil {
						// todo(wdvxdr): fix scope find, extra finding type in
						//               file.Package.Scope is weired.
						typ, err = file.Package.Scope.Type(field.Type)
					}
					if err != nil {
						file.errorf(field.Position, "unknown type: %s", field.Type)
						return
					}
					f.Type = typ
				}

				switch {
				case field.Repeated:
					f.Flag.Set(types.FRepeated, true)
				case field.Optional:
					f.Flag.Set(types.FOptional, true)
				case field.Required:
					f.Flag.Set(types.FRequired, true)
				}

				if !f.IsRepeated() {
					if (file.Version == 2 && f.Type.Name() != "bytes") ||
						(file.Version == 3 && f.Type.ScopeClass() == types.CMessage) {
						f.Set(types.FPtr, true)
					}
				}

				msg.Fields = append(msg.Fields, f)
			})

		case *proto.Message:
			file.parseMessage(field, scope)
		}
	}
}

func (file *File) parseEnum(elem *proto.Enum, s *types.Scope) {
	enum := s.LookupEnum(elem)
	enum.Name = elem.Name
	enum.SetPos(elem.Position)
	for _, field := range elem.Elements {
		switch field := field.(type) {
		case *proto.EnumField:
			f := &types.EnumField{
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
