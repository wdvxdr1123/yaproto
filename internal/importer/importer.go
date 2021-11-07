package importer

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"text/scanner"

	"github.com/emicklei/proto"

	"github.com/wdvxdr1123/yaproto/internal/types"
)

var Packages = make(map[string]*Package)
var ProtoPath = ""

type Package struct {
	Path    string
	Package string
	Error   func(err error)

	Proto      *proto.Proto
	Imported   []string
	GoPackage  string
	OutputPath string
	Version    int
	Universe   *types.Scope

	once    sync.Once
	delayed []func()
	pos     scanner.Position
}

func Import(path string) (*Package, error) {
	if pkg, ok := Packages[path]; ok {
		return pkg, nil
	}

	p := &Package{
		Path: path,
		Error: func(err error) {
			panic(err)
		},
		Universe: types.NewScope(nil, ""),
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if runtime.GOOS == "windows" {
		p.Path = strings.Replace(p.Path, "\\", "/", -1)
	}

	pb, err := proto.NewParser(file).Parse()
	if err != nil {
		return nil, err
	}
	p.Proto = pb
	p.Proto.Filename = path
	if err := p.parse(); err != nil {
		return nil, err
	}
	Packages[path] = p
	return p, nil
}

func (pkg *Package) parse() error {
	for _, elem := range pkg.Proto.Elements {
		switch elem := elem.(type) {
		case *proto.Package:
			pkg.Package = elem.Name
		case *proto.Syntax:
			pkg.setPos(elem.Position)
			switch elem.Value {
			case "proto3":
				pkg.Version = 3
			case "proto2":
				pkg.Version = 2
			default:
				pkg.errorf("unsupported syntax version: %s", elem.Value)
			}
		case *proto.Option:
			if elem.Name == "go_package" {
				gopkg := elem.Constant.Source
				for i, str := range strings.Split(gopkg, ";") {
					switch i {
					case 0:
						pkg.OutputPath = str
					case 1:
						pkg.GoPackage = str
					}
				}
			}
		case *proto.Import:
			_, err := Import(elem.Filename)
			if err != nil {
				pkg.error(err)
			}
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
	return nil
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

func (pkg *Package) setPos(pos scanner.Position) {
	pkg.pos = pos
}

func (pkg *Package) error(err error) {
	e := &Error{
		Pos: pkg.pos,
		Err: err,
	}
	if pkg.Error != nil {
		pkg.Error(e)
	}
}

func (pkg *Package) errorf(format string, args ...interface{}) {
	pkg.error(errors.New(fmt.Sprintf(format, args...)))
}
