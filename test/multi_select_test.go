package test

import (
	"testing"
)

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

func TestMultiSelect_GoGenerate_All_InTest(t *testing.T) {
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
		Files("mocks/primary_1.go", "mocks/primary_2.go").
		Golden()
}

func TestMultiSelect_GoGenerate_All_InBlackBoxTest(t *testing.T) {
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
		Files("mocks/primary_1.go", "mocks/primary_2.go").
		Golden()
}

func TestMultiSelect_Console_All_Primary_NoInterfaces(t *testing.T) {
	tr := newTester(t, M{
		Name: "repo/pkg",
		Files: map[string]interface{}{
			"primary.go": /* language=go */ `
			package pkg
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
	tr.Gmg(t, "--all").Succeed().Files()
}

