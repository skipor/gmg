package main

import (
	"strings"

	"golang.org/x/tools/go/packages"
)

type PackageKind string

const (
	// PrimaryPackageKind is package of usual *.go files.
	// ID is like 'pkg'.
	PrimaryPackageKind = "primary"
	// TestPackageKind is package of *_test.go files with package name like non-test files.
	// ID is like 'pkg [pkg.test]'
	TestPackageKind = "test"
	// BlackBoxTestPackageKind is package of *_test.go files with package name with '_test' suffix.
	// ID is like 'pkg_test [pkg.test]'
	BlackBoxTestPackageKind = "black-box-test"
	// TestExecutablePackageKind is virtual package from test executable files, that are generated during 'go test' rung.
	// ID is like 'pkg.test'
	TestExecutablePackageKind = "test-executable"
)

func GetPackageKind(p *packages.Package) PackageKind {
	if strings.HasSuffix(p.ID, ".test") {
		return TestExecutablePackageKind
	}
	if !strings.HasSuffix(p.ID, ".test]") {
		return PrimaryPackageKind
	}
	if strings.HasSuffix(p.PkgPath, "_test") {
		return BlackBoxTestPackageKind
	}
	return TestPackageKind
}

func (k PackageKind) String() string { return string(k) }

func errorKindStr(k packages.ErrorKind) string {
	switch k {
	case packages.UnknownError:
		return "unknown"
	case packages.ListError:
		return "list"
	case packages.ParseError:
		return "parse"
	case packages.TypeError:
		return "type"
	}
	return "unexpected"
}

func packagesWithoutTestExecutable(pkgs []*packages.Package) []*packages.Package {
	for i, pkg := range pkgs {
		if GetPackageKind(pkg) == TestExecutablePackageKind {
			return append(pkgs[:i], pkgs[i+1:]...)
		}
	}
	return pkgs
}

func errorsNum(pkgs []*packages.Package) int {
	var n int
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		n += len(pkg.Errors)
	})
	return n
}
