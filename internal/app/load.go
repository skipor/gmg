package app

import (
	"bytes"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/tools/go/packages"
)

func loadPackages(log *zap.SugaredLogger, env *Environment, src string) ([]*packages.Package, error) {
	log.Debugf("Loading package: %s", src)
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedModule |
			// Workaround to fix "Unexpected package creation during export data loading".
			// See https://github.com/golang/go/issues/45218.
			packages.NeedImports |
			// Print files in debug, to see what files where loaded.
			packages.NeedFiles | packages.NeedCompiledGoFiles,
		Dir:        env.Dir,
		Env:        env.Env,
		Tests:      true,
		BuildFlags: nil, // TODO(skipor)
	}, src)
	if err != nil {
		return nil, err
	}
	pkgs = packagesWithoutTestExecutable(pkgs)
	debugLogPkgs(log, pkgs)

	loadFailed := len(pkgs) == 1 && pkgs[0].Name == ""
	if loadFailed {
		return nil, &loadErrs{pkgs[0].Errors}
	}

	if errNum := packagesErrorsNum(pkgs); errNum != 0 {
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
		p("  Compiled files: %s\n", pkg.CompiledGoFiles)
		p("  Ignored files: %s\n", pkg.IgnoredFiles)
		if m := pkg.Module; m != nil {
			p("  Module:\n")
			p("    Path: %s\n", m.Path)
			p("    Dir: %s\n", m.Dir)
		}
		errsPrint := printPackagesErrors(pkgs[i : i+1])
		if errsPrint != "" {
			p("  Errors:\n")
			for _, line := range strings.Split(errsPrint, "\n") {
				p("  %s\n", line)
			}
		}
	}
	log.Debugf(w.String())
}

func printPackagesErrors(pkgs []*packages.Package) string {
	b := &bytes.Buffer{}
	p := func(format string, a ...interface{}) { _, _ = fmt.Fprintf(b, format, a...) }
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, err := range pkg.Errors {
			p("\t- %s\n", packagesErrorString(err))
		}
	})
	return strings.TrimRight(b.String(), "\n")
}

func packagesErrorString(err packages.Error) string {
	if err.Pos == "" {
		return err.Msg
	}
	return fmt.Sprintf("%s: %s", err.Pos, err.Msg)
}

type loadErrs struct {
	errs []packages.Error
}

func (e *loadErrs) Error() string {
	if len(e.errs) == 1 {
		return packagesErrorString(e.errs[0])
	}
	buf := &bytes.Buffer{}
	for _, err := range e.errs {
		_, _ = fmt.Fprintf(buf, "- %s\n", packagesErrorString(err))
	}
	return buf.String()
}
