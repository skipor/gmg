package main

import (
	"bytes"
	"fmt"
	"go/token"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/tools/go/packages"
)

func loadPackages(log *zap.SugaredLogger, env *Environment, src string) ([]*packages.Package, error) {
	log.Debugf("Loading package: %s", src)
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName |
			packages.NeedImports | // Workaround to fix "Unexpected package creation during export data loading". See https://github.com/golang/go/issues/45218
			packages.NeedModule |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedTypes,
		Dir:        env.Dir,
		Env:        env.Env,
		ParseFile:  nil, // TODO(skipor): optimize - remove static functions, methods bodies, to accelerate type checking
		Fset:       token.NewFileSet(),
		Tests:      true,
		BuildFlags: nil, // TODO(skipor)
	}, src)
	if err != nil {
		return nil, err
	}
	pkgs = packagesWithoutTestExecutable(pkgs)
	debugLogPkgs(log, pkgs)
	if errNum := errorsNum(pkgs); errNum != 0 {
		str := printPackagesErrors(pkgs)
		log.Warnf("Packages loaded with %v errors. Generation may fail, if type information was not able to load.\n%s", errNum, str)
	}
	return pkgs, nil
}

func debugLogPkgs(log *zap.SugaredLogger, pkgs []*packages.Package) {
	w := &bytes.Buffer{}
	p := func(format string, args ...interface{}) { _, _ = fmt.Fprintf(w, format, args...) }

	p("Loaded %v packages:\n", len(pkgs))
	for i, pkg := range pkgs {
		if i != 0 {
			p("\n")
		}
		p("- ID: %s\n", pkg.ID)
		p("  Name: %s\n", pkg.Name)
		p("  PkgPath: %s\n", pkg.PkgPath)
		p("  Files: %s\n", pkg.GoFiles)
		p("  Compiled files: %s\n", pkg.CompiledGoFiles)
		p("  Types information: %v\n", pkg.Types != nil)

		if m := pkg.Module; m != nil {
			p("  Module:\n")
			p("    Path: %s\n", m.Path)
			p("    Dir: %s\n", m.Dir)
		}
	}
	log.Debugf(w.String())
}

func printPackagesErrors(pkgs []*packages.Package) string {
	b := &bytes.Buffer{}
	p := func(format string, a ...interface{}) { _, _ = fmt.Fprintf(b, format, a...) }
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, err := range pkg.Errors {
			p("\t- ")
			if err.Pos != "" {
				p("%s: ", err.Pos)
			}
			p(err.Msg)
			p("\n")
		}
	})
	return strings.TrimSpace(b.String())
}
