package example

import (
	"go.uber.org/zap/zapcore"
)

// `//go:generate gmg`  without interface names arguments generates mock for next type declaration.
// That is, that comment generates mock for Foo.
//go:generate gmg

// Foo is an example interface.
type Foo interface {
	Bar(s string) error
}

// One or more interface names can be passed as arguments.
//go:generate gmg First Second Third

type (
	First  interface{ One() }
	Second interface{ Two() }
	Third  interface{ Three() }
)

// Absolute or relative source package can be specified via `--src (-s)` flag
//go:generate gmg --src ./sub Baz
//go:generate gmg --src io Reader Writer Closer
//go:generate gmg --src go.uber.org/zap/zapcore Core

// Instead of specifying full source package path, mock can be generated for type alias.
//go:generate gmg

type ZapEncoder = zapcore.Encoder
