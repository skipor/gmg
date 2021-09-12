package example

// `//go:generate gmg` without interface names arguments generates mock for next type declaration.
// That is, that comment generates mock for Foo.
//go:generate gmg

// Foo is an example interface.
type Foo interface {
	Bar(s string) error
}

func Do(foo Foo) error {
	return foo.Bar("string that contains something")
}
