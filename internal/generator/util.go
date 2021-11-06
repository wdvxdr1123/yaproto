package generator

func (g *Generator) proto2() bool { return g.Pkg.Version == 2 }
func (g *Generator) proto3() bool { return g.Pkg.Version == 3 }

func (g *Generator) importGoPackage(path, alias string) {
	g.goImport[GoPackage{path, alias}] = true
}
