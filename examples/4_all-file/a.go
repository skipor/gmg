package example

// go:generate gmg --all-file generates mocks for all interfaces in current file.
//go:generate gmg --all-file

type A1 interface{ A1() }
type A2 interface{ A2() }
