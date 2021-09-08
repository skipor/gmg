package test

import (
	"testing"
)

func TestGoGenerate_ExplicitName(t *testing.T) {
	tr := newTester(t, M{
		Name: "pkg",
		Files: map[string]interface{}{
			"file.go": /* language=go */ `
			package pkg
			//go:generate gmg Foo
			type Baz interface { Qux() }
			type Foo interface { Bar() }
			`,
		},
	})
	tr.GoGenerate(t).Succeed().Files("mocks/foo.go")
}

func TestGoGenerate_ImplicitName(t *testing.T) {
	tr := newTester(t, M{
		Name: "pkg",
		Files: map[string]interface{}{
			"file.go": /* language=go */ `
			package pkg
			//go:generate gmg
			type Foo interface { Bar() string }
			`,
		},
	})
	tr.GoGenerate(t).Succeed().Files("mocks/foo.go")
}

func TestGoGenerate_ImplicitName_BeforeTypeDecl(t *testing.T) {
	tr := newTester(t, M{
		Name: "pkg",
		Files: map[string]interface{}{
			"file.go": /* language=go */ `
			package pkg
			//go:generate gmg
			type (
				Foo interface { Bar() string }
			)
			`,
		},
	})
	tr.GoGenerate(t).Succeed().Files("mocks/foo.go")
}

func TestGoGenerate_ImplicitName_Fail_AtEndOfTypeDecl(t *testing.T) {
	tr := newTester(t, M{
		Name: "pkg",
		Files: map[string]interface{}{
			"file.go": /* language=go */ `
			package pkg
			type (
				Foo interface { Bar() string }
			//go:generate gmg
			)
			`,
		},
	})
	tr.GoGenerate(t).Fail()
}

func TestGoGenerate_ImplicitName_Fail_FuncDecl(t *testing.T) {
	tr := newTester(t, M{
		Name: "pkg",
		Files: map[string]interface{}{
			"file.go": /* language=go */ `
			package pkg
			//go:generate gmg
			func Baz() {}
			type Foo interface { Bar() string }
			`,
		},
	})
	tr.GoGenerate(t).Fail()
}

func TestGoGenerate_ImplicitName_Fail_StructDecl(t *testing.T) {
	tr := newTester(t, M{
		Name: "pkg",
		Files: map[string]interface{}{
			"file.go": /* language=go */ `
			package pkg
			//go:generate gmg
			type Baz struct {}
			type Foo interface { Bar() string }
			`,
		},
	})
	tr.GoGenerate(t).Fail()
}

func TestGoGenerate_ImplicitName_Fail_IntDecl(t *testing.T) {
	tr := newTester(t, M{
		Name: "pkg",
		Files: map[string]interface{}{
			"file.go": /* language=go */ `
			package pkg
			//go:generate gmg
			type Baz int
			type Foo interface { Bar() string }
			`,
		},
	})
	tr.GoGenerate(t).Fail()
}

func TestGoGenerate_ImplicitName_Fail_NoDecl(t *testing.T) {
	tr := newTester(t, M{
		Name: "pkg",
		Files: map[string]interface{}{
			"file.go": /* language=go */ `
			package pkg
			type Foo interface { Bar() string }
			//go:generate gmg
			`,
		},
	})
	tr.GoGenerate(t).Fail()
}

func TestGoGenerate_ImplicitPackageName_Deduce(t *testing.T) {
	tr := newTester(t, M{
		Name: "pkg",
		Files: map[string]interface{}{
			"file.go": /* language=go */ `
			package pkg
			//go:generate gmg
			
			type Foo interface { Bar() string }
			`,
			"mocks/doc.go": /* language=go */ `
			package custom_mocks_dir_package
			`,
		},
	})
	tr.GoGenerate(t).Succeed().Golden()
}

func TestGoGenerate_ImplicitPackageName_NoGoFiles(t *testing.T) {
	tr := newTester(t, M{
		Name: "pkg",
		Files: map[string]interface{}{
			"file.go": /* language=go */ `
			package pkg
			//go:generate gmg
			
			type Foo interface { Bar() string }
			`,
			"mocks/some.txt": `
            non go file
			`,
		},
	})
	tr.GoGenerate(t).Succeed().Golden()
}
