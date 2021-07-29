package simple_mock_usage

//go:generate gmg Foo

type Foo interface {
	Bar(s string) error
}

func Do(foo Foo) error {
	return foo.Bar("string that contains something")
}
