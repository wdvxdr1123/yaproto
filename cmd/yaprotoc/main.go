package main

import (
	"flag"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/wdvxdr1123/yaproto/internal/generator"
	"github.com/wdvxdr1123/yaproto/internal/importer"
)

var (
	getter  = flag.Int("getter", 1, "generate getter methods")
	size    = flag.Bool("size", false, "generate size methods")
	marshal = flag.Int("marshal", 1, "generate marshal/unmarshal methods")
	output  = flag.String("output", ".", "output directory")
)

func main() {
	flag.StringVar(&importer.ProtoPath, "pkg", "", "package prefix")
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	flag.Parse()

	group := errgroup.Group{}
	_ = filepath.Walk(flag.Args()[0], func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		// only parse proto files
		if filepath.Ext(path) != ".proto" {
			return nil
		}
		group.Go(func() error {
			p, err := importer.Import(path)
			if err != nil {
				return err
			}
			p.Error = func(err error) {
				println(err)
			}
			return nil
		})
		return nil
	})

	if err := group.Wait(); err != nil {
		println(err.Error())
		return
	}

	all := make([]*importer.Package, 0)
	for _, p := range importer.Packages {
		all = append(all, p)
	}

	for _, p := range all {
		pkg := p
		group.Go(func() error {
			pkg.Resolve()
			g := generator.New(pkg)
			g.Options.GenGetter = *getter
			g.Options.GenMarshal = *marshal
			g.Options.GenSize = *size || *marshal > 1

			if runtime.GOOS == "windows" {
				wd = strings.Replace(wd, "\\", "/", -1)
			}
			outputPath := path.Clean(path.Join(wd, *output, pkg.OutputPath))

			filename := strings.TrimSuffix(pkg.Path, ".proto")
			dot := strings.LastIndexByte(filename, '/')
			if dot > 0 {
				filename = filename[dot+1:]
			}

			outfile := path.Join(outputPath, filename+".pb.go")
			err := os.MkdirAll(path.Dir(outfile), 0644)
			if err != nil {
				return err
			}

			out, err := os.OpenFile(outfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}

			g.Generate(out)
			return out.Close()
		})
	}
	err = group.Wait()
	if err != nil {
		println(err.Error())
		return
	}
}
