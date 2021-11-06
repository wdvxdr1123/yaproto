package importer

import (
	"sort"
	"strings"

	"github.com/emicklei/proto"

	"github.com/wdvxdr1123/yaproto/internal/types"
	"github.com/wdvxdr1123/yaproto/internal/utils"
)

func (pkg *Package) parseMessage(m *proto.Message, s *types.Scope) {
	msg := s.LookupMessage(m)
	msg.Name = m.Name
	scope := types.NewScope(s, m.Name)
	for _, field := range m.Elements {
		switch field := field.(type) {
		case *proto.NormalField:
			// the field type maybe not defined yet, so we should
			// look up the type later
			pkg.later(func() {
				f := &types.MessageField{
					Name:     field.Name,
					Sequence: field.Sequence,
				}

				if strings.Contains(field.Type, ".") {
					t := field.Type
					dot := strings.LastIndexByte(t, '.')
					ipkg := pkg.lookup(t[:dot])
					gotype := ipkg.GoPackage + "." + utils.CamelCase(t[dot+1:])
					f.Type = &types.ImportedType{TypeName: t, Gotype: gotype}
				} else {
					f.Type = scope.Type(field.Type)
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
					if (pkg.Version == 2 && f.Type.Name() != "bytes") ||
						(pkg.Version == 3 && f.Type.Scope() == types.CMessage) {
						f.Set(types.FPtr, true)
					}
				}

				msg.Fields = append(msg.Fields, f)
			})

		case *proto.Message:
			pkg.parseMessage(field, scope)
		}
	}
}

func (pkg *Package) parseEnum(elem *proto.Enum, s *types.Scope) {
	enum := s.LookupEnum(elem)
	enum.Name = elem.Name
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
