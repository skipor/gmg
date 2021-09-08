package test

import (
	"testing"
)

func TestTrivial(t *testing.T) {
	tr := newTester(t, M{
		Name: "pkg",
		Files: map[string]interface{}{
			"file.go": /* language=go */ `
			package pkg
			type Foo interface { Bar() string }
			`,
		},
	})
	tr.
		Gmg(t, "Foo").Succeed().
		Golden()
}
func TestTrivial_TestOnly(t *testing.T) {
	tr := newTester(t, M{
		Name: "pkg",
		Files: map[string]interface{}{
			"file_test.go": /* language=go */ `
			package mypkg
			type Foo interface { Bar() string }
			`,
		},
	})
	tr.
		Gmg(t, "Foo").Succeed().
		Golden()
}

func TestTrivial_BlackBoxTestOnly(t *testing.T) {
	tr := newTester(t, M{
		Name: "pkg",
		Files: map[string]interface{}{
			"doc.go": /* language=go */ `
			package mypkg
			`,
			"file_test.go": /* language=go */ `
			package mypkg_test
			type Foo interface { Bar() string }
			`,
		},
	})
	tr.
		Gmg(t, "Foo").Succeed().
		Golden()
}

func TestSignatureCornerCases(t *testing.T) {
	tr := newTester(t, M{
		Name: "pkg",
		Files: map[string]interface{}{
			"file.go": /* language=go */ `
			package pkg
			import (
				"context"
				"testing"
			)
			type Foo interface { 
				NoArgsAndResults() 
				NamedArgsAndResults(a int) (b int)
				VariadicArgs(f string, as ...int) 
				UnderscoreArgsAndResults(_ int) (_ int)
				ReservedArgNames(c, r, m, res, call, reflect, gomock int) 
				ReservedResultNames() (c, r, m, res, call, reflect, gomock int) 
				BeforeOtherPackagesNamesArgs(testing int)
				BeforeOtherPackagesNamesResults() (testing int)
				WellKnownNamesArgs(context.Context, *testing.T, error)
				WellKnownNamesResults() (context.Context, *testing.T, error)
				AfterOtherPackagesNamesArgs(context int)
				AfterOtherPackagesNamesResults() (context int)
			}
			`,
		},
	})
	tr.
		Gmg(t, "Foo").Succeed().
		Golden()
}

func TestIoWriter(t *testing.T) {
	newTester(t, M{
		Name: "pkg",
		Files: map[string]interface{}{
			"file.go": /* language=go */ `
			package pkg
			`,
		},
	}).
		Gmg(t, "--src", "io", "Writer").Succeed().
		Golden()
}

func Test_RelativeSrc(t *testing.T) {
	tr := newTester(t, M{
		Name: "pkg",
		Files: map[string]interface{}{
			"sub/file.go": /* language=go */ `
			package sub
			type Foo interface { Bar() string }
			`,
		},
	})
	tr.Gmg(t, "--src", "./sub", "Foo").Files("mocks/foo.go").Succeed()
}

func TestPackageNotFound(t *testing.T) {
	tr := newTester(t, M{
		Name: "pkg",
		Files: map[string]interface{}{
			"sub/file.go": /* language=go */ `
			package sub
			type Foo interface { Bar() string }
			`,
		},
	})
	tr.Gmg(t, "--src", "./not_found", "Foo").Fail()
}

func TestPackageNameConflictSucceed(t *testing.T) {
	tr := newTester(t, M{
		Name: "pkg",
		Files: map[string]interface{}{
			"a.go": /* language=go */ `
			package a
			type Foo interface { Bar() string }
			`,
			"b.go": /* language=go */ `
			package b
			`,
		},
	})
	tr.Gmg(t, "Foo").Files("mocks/foo.go").Succeed()
}

func TestPackageNameConflictFail(t *testing.T) {
	tr := newTester(t, M{
		Name: "pkg",
		Files: map[string]interface{}{
			"a.go": /* language=go */ `
			package a
			`,
			"b.go": /* language=go */ `
			package b
			type Foo interface { Bar() string }
			`,
		},
	})
	tr.Gmg(t, "Foo").Fail()
}

func TestConsoleNamesSelect_NoArgs(t *testing.T) {
	tr := newTester(t, M{
		Name: "pkg",
		Files: map[string]interface{}{
			"file.go": /* language=go */ `
			package pkg
			type Foo interface { Bar() string }
			`,
		},
	})
	tr.Gmg(t).Fail()
}

