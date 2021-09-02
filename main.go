package main

import (
	"os"

	"github.com/skipor/gmg/internal/app"
)

func main() {
	env := app.RealEnvironment()
	exitCode := app.Main(env)
	os.Exit(exitCode)
}
