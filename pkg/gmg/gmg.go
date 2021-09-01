package gmg

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/skipor/gmg/pkg/gogen"
)

const gmgVersion = "v0.3.0"

func Main(env *Environment) int {
	params, err := loadParams(env)
	if errors.Is(err, errExitZero) {
		return 0
	}
	if err != nil {
		return handleError(env, err)
	}
	err = run(env, params)
	return handleError(env, err)
}

type Environment struct {
	Args   []string
	Stderr io.Writer
	Dir    string
	Env    []string
	// Fs is for output only. Go tooling invoked under hood, that read real files.
	Fs afero.Fs
}

func (e *Environment) Getenv(key string) string {
	for i := len(e.Env) - 1; i >= 0; i-- {
		kv := e.Env[i]
		split := strings.SplitN(kv, "=", 2)
		k := split[0]
		if k == key {
			if len(split) != 2 {
				panic(fmt.Sprintf("Environment.Env[%v]: expect 'key=value' format, but got: '%s'", k, kv))
			}
			return split[1]
		}
	}
	return ""
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

var errExitZero = errors.New("should exit with zero code")

type params struct {
	Log            *zap.SugaredLogger
	InterfaceNames []string
	// Source is Go package to search for interfaces. See flag description for details.
	Source string
	// Destination is directory or file relative path or pattern. See flag description for details.
	Destination string
	// Package is package name in generated files. See flag description for details.
	Package       string
	GoGenerateEnv goGenerateEnv
}

type goGenerateEnv struct {
	// GOLINE set by 'go generate' to line number of the directive in the source file.
	GOLINE int
	// GOFILE set by 'go generate' to the base name of the file.
	GOFILE string
	// GOPACKAGE the name of the package of the file containing the directive.
	GOPACKAGE string
}

func (e goGenerateEnv) isSet() bool {
	return e.GOLINE != 0 && e.GOFILE != "" && e.GOPACKAGE != ""
}

func (e goGenerateEnv) packageKind() packageKind {
	if strings.HasSuffix(e.GOPACKAGE, "_test") {
		return blackBoxTestPackageKind
	}
	if strings.HasSuffix(e.GOFILE, "_test.go") {
		return testPackageKind
	}
	return primaryPackageKind
}

func loadParams(env *Environment) (*params, error) {
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
	fs.StringVarP(&pkg, "pkg", "p", "",
		"Package name in generated files.\n"+
			"'{}' will be replaced with source package name.\n"+
			"By default, --dst package name used, or 'mocks_{}' if --dst package is not exist.\n"+
			"Examples:\n"+
			"	mocks_{} # mockgen style\n"+
			"	{}mocks # mockery style\n")
	fs.BoolVar(&debug, "debug", os.Getenv("GMG_DEBUG") != "", "Verbose debug logging.")
	fs.BoolVar(&version, "version", false, "Show version and exit.")
	err := fs.Parse(env.Args)
	if err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return nil, errExitZero
		}
		return nil, fmt.Errorf("flags parse: %w", err)
	}
	if version {
		_, _ = fmt.Fprintf(env.Stderr, "gmg %s\n", gmgVersion)
		return nil, errExitZero
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

	var goLine int
	if goLineStr := env.Getenv("GOLINE"); goLineStr != "" {
		goLineInt64, err := strconv.ParseInt(goLineStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("GOLINE='%s'is not an integer: %w", goLineStr, err)
		}
		goLine = int(goLineInt64)
	}

	goGenerateEnv := goGenerateEnv{
		GOLINE:    goLine,
		GOFILE:    env.Getenv("GOFILE"),
		GOPACKAGE: env.Getenv("GOPACKAGE"),
	}
	log.Sugar().Debugf("Go env: %+v", goGenerateEnv)
	if !goGenerateEnv.isSet() && len(interfaces) == 0 {
		return nil, fmt.Errorf("pass interface names as arguments.\n" +
			"Or put `//go:generate gmg` comment before interface declaration and run `go generate`.\n" +
			"Or run `gmg --help` to get more information.")
	}

	return &params{
		Log:            log.Sugar(),
		Source:         src,
		Destination:    path.Clean(dst),
		Package:        pkg,
		InterfaceNames: interfaces,
		GoGenerateEnv:  goGenerateEnv,
	}, nil

}

func handleError(env *Environment, err error) int {
	if err == nil {
		return 0
	}
	_, _ = fmt.Fprintf(env.Stderr, "ERROR: %+v\n", err)
	return 1
}

const placeHolder = "{}"

func run(env *Environment, params *params) error {
	log := params.Log
	log.Debugf("gmg version %s %s/%s", gmgVersion, runtime.GOOS, runtime.GOARCH)
	pkgs, err := loadPackages(log, env, params.Source)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "\n") {
			errStr = "\n" + errStr
		}
		return fmt.Errorf("package '%s' load failed: %s", params.Source, errStr)
	}
	primaryPkg := pkgs[0]
	log.Infof("Processing package: %s", primaryPkg.ID)
	var opts []gogen.Option
	if primaryPkg.Module != nil {
		opts = append(opts, gogen.ModulePath(primaryPkg.Module.Path))
	}
	g := gogen.NewGenerator(opts...)
	err = generateAll(g, env, pkgs, params)
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
