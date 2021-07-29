package gmg

import (
	"strings"

	"golang.org/x/tools/go/packages"
)

type packageKind string

const (
	// primaryPackageKind is package of usual *.go files.
	// ID is like 'pkg'.
	primaryPackageKind = "primary"
	// testPackageKind is package of *_test.go files with package name like non-test files.
	// ID is like 'pkg [pkg.test]'
	testPackageKind = "test"
	// blackBoxTestPackageKind is package of *_test.go files with package name with '_test' suffix.
	// ID is like 'pkg_test [pkg.test]'
	blackBoxTestPackageKind = "black-box-test"
	// testExecutablePackageKind is virtual package from test executable files, that are generated during 'go test' rung.
	// ID is like 'pkg.test'
	testExecutablePackageKind = "test-executable"
)

func getPackageKind(p *packages.Package) packageKind {
	if strings.HasSuffix(p.ID, ".test") {
		return testExecutablePackageKind
	}
	if !strings.HasSuffix(p.ID, ".test]") {
		return primaryPackageKind
	}
	if strings.HasSuffix(p.PkgPath, "_test") {
		return blackBoxTestPackageKind
	}
	return testPackageKind
}

func (k packageKind) String() string { return string(k) }

func packagesWithoutTestExecutable(pkgs []*packages.Package) []*packages.Package {
	for i, pkg := range pkgs {
		if getPackageKind(pkg) == testExecutablePackageKind {
			return append(pkgs[:i], pkgs[i+1:]...)
		}
	}
	return pkgs
}

func packagesErrorsNum(pkgs []*packages.Package) int {
	var n int
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		n += len(pkg.Errors)
	})
	return n
}
