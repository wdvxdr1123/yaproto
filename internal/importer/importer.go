package importer

import (
	"os"
	"strings"
	"sync"

	"github.com/emicklei/proto"

	"github.com/wdvxdr1123/yaproto/internal/types"
)

var Packages = make(map[string]*Package)
var ProtoPath = ""

type Package struct {
	Path    string
	Package string

	Proto      *proto.Proto
	Imported   []string
	GoPackage  string
	OutputPath string
	Version    int
	Universe   *types.Scope

	once    sync.Once
	delayed []func()
}

func Import(path string) (*Package, error) {
	if pkg, ok := Packages[path]; ok {
		return pkg, nil
	}

	p := &Package{
		Path:     path,
		Universe: types.NewScope(nil, ""),
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	pb, err := proto.NewParser(file).Parse()
	if err != nil {
		return nil, err
	}
	p.Proto = pb
	p.Proto.Filename = path
	p.parse()
	Packages[path] = p
	return p, nil
}

func (pkg *Package) parse() {
	for _, elem := range pkg.Proto.Elements {
		switch elem := elem.(type) {
		case *proto.Package:
			pkg.Package = elem.Name
		case *proto.Syntax:
			switch elem.Value {
			case "proto3":
				pkg.Version = 3
			case "proto2":
				pkg.Version = 2
			default:
				panic("unsupported syntax version")
			}
		case *proto.Option:
			if elem.Name == "go_package" {
				gopkg := elem.Constant.Source
				for i, str := range strings.Split(gopkg, ";") {
					switch i {
					case 0:
						pkg.OutputPath = ProtoPath + "/" + str
					case 1:
						pkg.GoPackage = str
					}
				}
			}
		case *proto.Import:
			Import(elem.Filename)
			pkg.Imported = append(pkg.Imported, elem.Filename)
		}
	}

	for _, elem := range pkg.Proto.Elements {
		switch elem := elem.(type) {
		case *proto.Message:
			pkg.parseMessage(elem, pkg.Universe)
		case *proto.Enum:
			pkg.parseEnum(elem, pkg.Universe)
		}

	}

}

func (pkg *Package) later(f func()) {
	pkg.delayed = append(pkg.delayed, f)
}

func (pkg *Package) Resolve() {
	pkg.once.Do(func() {
		for len(pkg.delayed) > 0 {
			f := pkg.delayed[0]
			pkg.delayed = pkg.delayed[1:]
			f()
		}
	})
}

func (pkg *Package) lookup(s string) *Package {
	for _, imp := range pkg.Imported {
		if Packages[imp].Package == s {
			return Packages[imp]
		}
	}
	return nil
}