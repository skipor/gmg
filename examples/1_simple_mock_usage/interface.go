package simple_mock_usage

//go:generate gmg Foo

type Foo interface {
	Bar(s string) error
}
