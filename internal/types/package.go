package types

import (
	"sync"
)

var allPackage sync.Map

type Package struct {
	Name      string
	GoPackage string
	GoOutPath string
	Scope     *Scope

	sync.Mutex
}

func LookupPkg(name string) *Package {
	val, _ := allPackage.LoadOrStore(name, &Package{
		Name:  name,
		Scope: NewScope(nil, ""),
	})
	return val.(*Package)
}
