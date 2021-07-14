package main

import (
	"testing"
)

func TestTrivial_NoPackageArgument(t *testing.T) {
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
