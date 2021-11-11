package importer

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"text/scanner"

	"github.com/emicklei/proto"
	"golang.org/x/sync/singleflight"

	"github.com/wdvxdr1123/yaproto/internal/types"
)

var (
	packages = sync.Map{}
	single   = singleflight.Group{}
)

var ProtoPath = ""

type File struct {
	Path  string
	Proto *proto.Proto

	Version  int
	Universe *types.Scope
	Package  *types.Package
	Imports  []*types.Package
	Error    func(err error)

	once    sync.Once
	delayed []func()
}

func Import(path string) (*File, error) {
	pkg, ok := packages.Load(path)
	if ok {
		return pkg.(*File), nil
	}

	pkg, err, _ := single.Do(path, func() (interface{}, error) {
		p := &File{
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

		pb, err := proto.NewParser(file).Parse()
		if err != nil {
			return nil, err
		}
		p.Proto = pb
		p.Proto.Filename = path
		p.parse()
		packages.Store(path, p)
		return p, nil
	})
	if err != nil {
		return nil, err
	}
	return pkg.(*File), nil
}

func (file *File) parse() {
	pkg := new(types.Package)
	var gopkgpos scanner.Position
	for _, elem := range file.Proto.Elements {
		switch elem := elem.(type) {
		case *proto.Package:
			pkg.Name = elem.Name
		case *proto.Syntax:
			switch elem.Value {
			case "proto3":
				file.Version = 3
			case "proto2":
				file.Version = 2
			default:
				file.errorf(elem.Position, "unsupported syntax version: %s", elem.Value)
			}
		case *proto.Option:
			if elem.Name == "go_package" {
				gopkgpos = elem.Position
				for i, str := range strings.Split(elem.Constant.Source, ";") {
					switch i {
					case 0:
						pkg.GoOutPath = str
					case 1:
						pkg.GoPackage = str
					}
				}
			}
		case *proto.Import:
			ipkg, err := Import(elem.Filename)
			if err != nil {
				file.error(elem.Position, err)
			}
			if pkg.Name == "" {
				file.errorf(elem.Position, "import an unnamed package: %s", elem.Filename)
			}
			file.Imports = append(file.Imports, ipkg.Package)
		}
	}

	if pkg.Name != "" {
		file.Package = types.LookupPkg(pkg.Name)
		func() {
			file.Package.Lock()
			defer file.Package.Unlock()
			switch file.Package.GoPackage {
			case pkg.GoPackage:
				// ok
			case "":
				file.Package.GoPackage = pkg.GoPackage
			default:
				file.errorf(gopkgpos, "same package has different go package: %s and %s", pkg.GoPackage, file.Package.GoPackage)
			}
			switch file.Package.GoOutPath {
			case pkg.GoOutPath:
				// ok
			case "":
				file.Package.GoOutPath = pkg.GoOutPath
			default:
				file.errorf(gopkgpos, "same package has different go output path: %s and %s", pkg.GoOutPath, file.Package.GoOutPath)
			}
		}()
	} else {
		file.Package = pkg
	}

	for _, elem := range file.Proto.Elements {
		switch elem := elem.(type) {
		case *proto.Message:
			file.parseMessage(elem, file.Universe)
		case *proto.Enum:
			file.parseEnum(elem, file.Universe)
		}
	}

	// merge all scope to package's scope
	if pkg.Name != "" {
		file.Package.Lock()
		defer file.Package.Unlock()
		target := file.Package.Scope
		for name, obj := range file.Universe.Elems {
			if _, ok := target.LookupOK(name); ok {
				file.errorf(obj.Obj.Pos(), "duplicate symbol: %s", name)
				return
			}
			target.Elems[name] = obj
		}
		for _, child := range file.Universe.Children {
			target.Children = append(target.Children, child.Copy(target))
		}
	}

	return
}

func (file *File) later(fn func()) {
	file.delayed = append(file.delayed, fn)
}

func (file *File) Resolve() {
	file.once.Do(func() {
		// delete self import
		var j int
		for i, t := range file.Imports {
			if t == file.Package {
				continue
			}
			if i != j {
				file.Imports[j] = t
			}
			j++
		}
		file.Imports = file.Imports[:j]

		for len(file.delayed) > 0 {
			fn := file.delayed[0]
			file.delayed = file.delayed[1:]
			fn()
		}
	})
}

func (file *File) lookup(s string) *types.Package {
	for _, pkg := range file.Imports {
		if pkg.Name == s {
			return pkg
		}
	}
	return nil
}

func (file *File) error(pos scanner.Position, err error) {
	e := &Error{
		File: file.Path,
		Pos:  pos,
		Err:  err,
	}
	if file.Error != nil {
		file.Error(e)
	}
}

func (file *File) errorf(pos scanner.Position, format string, args ...interface{}) {
	file.error(pos, errors.New(fmt.Sprintf(format, args...)))
}

func RangePackage(f func(pkg *File)) {
	packages.Range(func(_, value interface{}) bool {
		f(value.(*File))
		return true
	})
}
