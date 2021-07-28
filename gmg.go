package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/skipor/gmg/pkg/gogen"
)

const gmgVersion = "v0.3.0"

func main() {
	env := RealEnvironment()
	exitCode := Main(env)
	os.Exit(exitCode)
}

func Main(env *Environment) int {
	params, err := LoadParams(env)
	if errors.Is(err, ErrExitZero) {
		return 0
	}
	if err != nil {
		return handleError(env, err)
	}
	err = run(env, params)
	return handleError(env, err)
}

func handleError(env *Environment, err error) int {
	if err == nil {
		return 0
	}
	_, _ = fmt.Fprintf(env.Stderr, "ERROR: %+v\n", err)
	return 1
}

type Environment struct {
	Args   []string
	Stderr io.Writer
	Dir    string
	Env    []string
	// Fs is for output only. Go tooling invoked under hood, that read real files.
	Fs afero.Fs
}

func RealEnvironment() *Environment {
	dir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("get workdir: %+v", err))
	}
	return &Environment{
		Args:   os.Args[1:],
		Stderr: os.Stderr,
		Dir:    dir,
		Fs:     afero.NewOsFs(),
		Env:    os.Environ(),
	}
}

var ErrExitZero = errors.New("should exit with zero code")

type Params struct {
	Log        *zap.SugaredLogger
	Interfaces []string
	// Source is Go package to search for interfaces. See flag description for details.
	Source string
	// Destination is directory or file relative path or pattern. See flag description for details.
	Destination string
	// Package is package name in generated files. See flag description for details.
	Package string
}

func LoadParams(env *Environment) (*Params, error) {
	fs := pflag.NewFlagSet("gmg", pflag.ContinueOnError)
	fs.PrintDefaults()
	fs.Usage = func() {
		b := &bytes.Buffer{}
		p := func(format string, a ...interface{}) { _, _ = fmt.Fprintf(b, format, a...) }
		p("gmg is type-safe, fast and handy alternative GoMock generator. See details at: https://github.com/skipor/gmg\n")
		p("\n")
		p("Usage: gmg [--src <package path>] [--dst <file path>] [--pkg <package name>] <interface name> [<interface name> ...]\n\n")
		p("Flags:\n%s", fs.FlagUsages())
		_, _ = b.WriteTo(env.Stderr)
	}
	var (
		pkg     string
		src     string
		dst     string
		debug   bool
		version bool
	)
	fs.StringVarP(&src, "src", "s", ".",
		"Source Go package to search for interfaces. Absolute or relative.\n"+
			"Maybe third-party or standard library package.\n"+
			"Examples:\n"+
			"	.\n"+
			"	./relative/pkg\n"+
			"	github.com/third-party/pkg\n"+
			"	io\n",
	)
	fs.StringVarP(&dst, "dst", "d", "./mocks",
		"Destination directory or file relative path or pattern.\n"+
			"'{}' in directory path will be replaced with the source package name.\n"+
			"'{}' in file name will be replaced with snake case interface name.\n"+
			"If no file name pattern specified, then '{}.go' used by default.\n"+
			"Examples:\n"+
			"	./mocks\n"+
			"	./{}mocks\n"+
			"	./mocks/{}_gomock.go\n"+
			"	./mocks_test.go # All mocks will be put to single file.\n",
	)
	fs.StringVarP(&pkg, "pkg", "p", "mocks_{}",
		"Package name in generated files.\n"+
			"'{}' will be replaced with source package name.\n"+
			"Examples:\n"+
			"	mocks_{} # mockgen style\n"+
			"	{}mocks # mockery style\n")
	fs.BoolVar(&debug, "debug", false, "Verbose debug logging.")
	fs.BoolVar(&version, "version", false, "Show version and exit.")
	err := fs.Parse(env.Args)
	if err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return nil, ErrExitZero
		}
		return nil, fmt.Errorf("flags parse: %w", err)
	}
	if version {
		_, _ = fmt.Fprintf(env.Stderr, "gmg %s\n", gmgVersion)
		return nil, ErrExitZero
	}

	if strings.HasSuffix(src, "/...") {
		return nil, fmt.Errorf("--src: can't use recursive pattern as a destination")
	}

	encConf := zap.NewDevelopmentEncoderConfig()
	encConf.TimeKey = ""
	level := zapcore.WarnLevel
	if debug {
		level = zapcore.DebugLevel
	}
	log := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(encConf),
		zapcore.AddSync(env.Stderr),
		level,
	))

	interfaces := fs.Args()
	if len(interfaces) == 0 {
		return nil, fmt.Errorf("need one or more interface names passed as arguments")
	}

	return &Params{
		Log:         log.Sugar(),
		Source:      src,
		Destination: path.Clean(dst),
		Package:     pkg,
		Interfaces:  interfaces,
	}, nil

}

const placeHolder = "{}"

func run(env *Environment, params *Params) error {
	log := params.Log
	pkgs, err := loadPackages(log, env, params.Source)
	if err != nil {
		return fmt.Errorf("package '%s' load failed: %w", params.Source, err)
	}
	primaryPkg := pkgs[0]
	log.Infof("Processing package: %s", primaryPkg.ID)
	var opts []gogen.Option
	if primaryPkg.Module != nil {
		opts = append(opts, gogen.ModulePath(primaryPkg.Module.Path))
	}
	g := gogen.NewGenerator(opts...)
	err = generateAll(g, pkgs, params)
	if err != nil {
		return err
	}
	var fileNames []string
	for _, f := range g.Files() {
		fileNames = append(fileNames, f.Path())
	}
	log.Debugf("Generating: %s", strings.Join(fileNames, ", "))
	for _, f := range g.Files() {
		if f.Skipped() {
			continue
		}
		err := f.WriteFile(env.Fs)
		if err != nil {
			return fmt.Errorf("file %s: %w", f.Path(), err)
		}
		_, _ = fmt.Fprintf(env.Stderr, "%s\n", f.Path())
	}
	return nil
}
