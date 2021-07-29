package main

import (
	"os"

	"github.com/skipor/gmg/pkg/gmg"
)

func main() {
	env := gmg.RealEnvironment()
	exitCode := gmg.Main(env)
	os.Exit(exitCode)
}
