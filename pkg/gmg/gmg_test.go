package gmg

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
		Succeed(t, "Foo").
		Golden()
}

func TestTrivial_TestOnly(t *testing.T) {
	tr := newTester(t, M{
		Name: "pkg",
		Files: map[string]interface{}{
			"file_test.go": /* language=go */ `
			package pkg
			type Foo interface { Bar() string }
			`,
		},
	})
	tr.
		Succeed(t, "Foo").
		Golden()
}

func TestTrivial_BlackBoxTestOnly(t *testing.T) {
	tr := newTester(t, M{
		Name: "pkg",
		Files: map[string]interface{}{
			"file_test.go": /* language=go */ `
			package pkg_test
			type Foo interface { Bar() string }
			`,
		},
	})
	tr.
		Succeed(t, "Foo").
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
		Succeed(t, "Foo").
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
		Succeed(t, "--src", "io", "Writer").
		Golden()
}

func TestTrivial_RelativeSrc(t *testing.T) {
	tr := newTester(t, M{
		Name: "pkg",
		Files: map[string]interface{}{
			"sub/file.go": /* language=go */ `
			package sub
			type Foo interface { Bar() string }
			`,
		},
	})
	tr.Succeed(t, "--src", "./sub", "Foo").Files("mocks/foo.go")
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
	tr.Fail(t, "--src", "./not_found", "Foo")
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
	tr.Succeed(t, "Foo")
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
	tr.Fail(t, "Foo")
}
