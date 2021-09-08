package test

import (
	"testing"
)

func TestMultiSelect_GoGenerate_All_Primary(t *testing.T) {
	tr := newTester(t, M{
		Name: "repo/pkg",
		Files: map[string]interface{}{
			"primary.go": /* language=go */ `
			package pkg
			//go:generate gmg --all
			type Primary1 interface { P1()  }
			type Primary2 interface { P2()  }
			`,
			"primary_test.go": /* language=go */ `
			package pkg
			type Test1 interface { T1()  }
			type Test2 interface { T2()  }
			`,
			"black_box_test.go": /* language=go */ `
			package pkg_test
			type BlackBox1 interface { BB1()  }
			type BlackBox2 interface { BB2()  }
			`,
		},
	})
	tr.GoGenerate(t).
		Succeed().
		Files("mocks/primary_1.go", "mocks/primary_2.go").
		Golden()
}

func TestMultiSelect_GoGenerate_All_Test(t *testing.T) {
	tr := newTester(t, M{
		Name: "repo/pkg",
		Files: map[string]interface{}{
			"primary.go": /* language=go */ `
			package pkg
			type Primary1 interface { P1()  }
			type Primary2 interface { P2()  }
			`,
			"primary_test.go": /* language=go */ `
			package pkg
			//go:generate gmg --all
			type Test1 interface { T1()  }
			type Test2 interface { T2()  }
			`,
			"black_box_test.go": /* language=go */ `
			package pkg_test
			type BlackBox1 interface { BB1()  }
			type BlackBox2 interface { BB2()  }
			`,
		},
	})
	tr.GoGenerate(t).
		Succeed().
		// TODO(skipor): change default dst: test_1_test.go, test_2_test.go
		Files("mocks/test_1.go", "mocks/test_2.go").
		Golden()
}

func TestMultiSelect_GoGenerate_All_BlackBoxTest(t *testing.T) {
	tr := newTester(t, M{
		Name: "repo/pkg",
		Files: map[string]interface{}{
			"primary.go": /* language=go */ `
			package pkg
			type Primary1 interface { P1()  }
			type Primary2 interface { P2()  }
			`,
			"primary_test.go": /* language=go */ `
			package pkg
			type Test1 interface { T1()  }
			type Test2 interface { T2()  }
			`,
			"black_box_test.go": /* language=go */ `
			//go:generate gmg --all
			package pkg_test
			type BlackBox1 interface { BB1()  }
			type BlackBox2 interface { BB2()  }
			`,
		},
	})
	tr.GoGenerate(t).
		Succeed().
		// TODO(skipor): change default dst: black_box_1_test.go, black_box_2_test.go
		Files("mocks/black_box_1.go", "mocks/black_box_2.go").
		Golden()
}


func TestMultiSelect_GoGenerate_All_BlackBoxTest_DotImport(t *testing.T) {
	t.Log(`
		Given black box test package dot imports primary
		When //go:generate --all in black box test pkg
		Then only black box test package interfaces generated
	`)
	tr := newTester(t, M{
		Name: "repo/pkg",
		Files: map[string]interface{}{
			"primary.go": /* language=go */ `
			package pkg
			type Primary1 interface { P1()  }
			type Primary2 interface { P2()  }
			`,
			"primary_test.go": /* language=go */ `
			package pkg
			type Test1 interface { T1()  }
			type Test2 interface { T2()  }
			`,
			"black_box_test.go": /* language=go */ `
			//go:generate gmg --all
			package pkg_test
            import . "repo/pkg"
			type BlackBox1 interface { BB1()  }
			type BlackBox2 interface { BB2()  }
			`,
		},
	})
	tr.GoGenerate(t).
		Succeed().
		// TODO(skipor): change default dst: black_box_1_test.go, black_box_2_test.go
		Files("mocks/black_box_1.go", "mocks/black_box_2.go")
}

func TestMultiSelect_Console_All_Primary(t *testing.T) {
	tr := newTester(t, M{
		Name: "repo/pkg",
		Files: map[string]interface{}{
			"primary.go": /* language=go */ `
			package pkg
			type Primary1 interface { P1()  }
			type Primary2 interface { P2()  }
			`,
			"primary_test.go": /* language=go */ `
			package pkg
			type Test1 interface { T1()  }
			type Test2 interface { T2()  }
			`,
			"black_box_test.go": /* language=go */ `
			package pkg_test
			type BlackBox1 interface { BB1()  }
			type BlackBox2 interface { BB2()  }
			`,
		},
	})
	tr.Gmg(t, "--all").Succeed().Files("mocks/primary_1.go", "mocks/primary_2.go")
}

func TestMultiSelect_GoGenerate_All_Primary_NoInterfaces(t *testing.T) {
	tr := newTester(t, M{
		Name: "repo/pkg",
		Files: map[string]interface{}{
			"primary.go": /* language=go */ `
			package pkg
			//go:generate gmg --all
			`,
			"primary_test.go": /* language=go */ `
			package pkg
			type Test1 interface { T1()  }
			type Test2 interface { T2()  }
			`,
			"black_box_test.go": /* language=go */ `
			package pkg_test
			type BlackBox1 interface { BB1()  }
			type BlackBox2 interface { BB2()  }
			`,
		},
	})
	tr.GoGenerate(t).
		Fail()
}

func TestMultiSelect_GoGenerate_All_Test_NoInterfaces(t *testing.T) {
	tr := newTester(t, M{
		Name: "repo/pkg",
		Files: map[string]interface{}{
			"primary.go": /* language=go */ `
			package pkg
			type Primary1 interface { P1()  }
			type Primary2 interface { P2()  }
			`,
			"primary_test.go": /* language=go */ `
			package pkg
			//go:generate gmg --all
			`,
			"black_box_test.go": /* language=go */ `
			package pkg_test
			type BlackBox1 interface { BB1()  }
			type BlackBox2 interface { BB2()  }
			`,
		},
	})
	tr.GoGenerate(t).
		Fail()
}

func TestMultiSelect_GoGenerate_All_BlackBoxTest_NoInterfaces(t *testing.T) {
	tr := newTester(t, M{
		Name: "repo/pkg",
		Files: map[string]interface{}{
			"primary.go": /* language=go */ `
			package pkg
			type Primary1 interface { P1()  }
			type Primary2 interface { P2()  }
			`,
			"primary_test.go": /* language=go */ `
			package pkg
			type Test1 interface { T1()  }
			type Test2 interface { T2()  }
			`,
			"black_box_test.go": /* language=go */ `
			//go:generate gmg --all
			package pkg_test
			`,
		},
	})
	tr.GoGenerate(t).
		Fail()
}
