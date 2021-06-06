package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/token"
	"go/types"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/tools/go/packages"

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
	Fs     afero.Fs
	Env    []string
}

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
	debugLogPkgs(log, pkgs)
	if errNum := errorsNum(pkgs); errNum != 0 {
		b := &bytes.Buffer{}
		p := func(format string, a ...interface{}) { _, _ = fmt.Fprintf(b, format, a...) }
		p("Packages loaded with %v errors. Generation may fail, if type information was not able to load.\n", errNum)
		packages.Visit(pkgs, nil, func(pkg *packages.Package) {
			for _, err := range pkg.Errors {
				p("\t- %s\n", err)
			}
		})
		log.Warnf(strings.TrimSpace(b.String()))
	}
	primaryPkg := pkgs[0]
	log.Infof("Processing package: %s", primaryPkg.ID)

	dstDir := strings.TrimPrefix(params.Destination, ".")
	fileNamePattern := placeHolder + ".go"
	if path.Ext(dstDir) == ".go" {
		dstDir, fileNamePattern = path.Split(dstDir)
	}
	dstDir = strings.ReplaceAll(dstDir, placeHolder, primaryPkg.Name)
	packageName := strings.ReplaceAll(params.Package, placeHolder, primaryPkg.Name)
	importPath := gogen.ImportPath(path.Join(primaryPkg.PkgPath, dstDir))

	var opts []gogen.Option
	if primaryPkg.Module != nil {
		opts = append(opts, gogen.ModulePath(primaryPkg.Module.Path))
	}
	g := gogen.NewGenerator(opts...)

	isSingleFile := !strings.Contains(fileNamePattern, placeHolder)
	var singleFile *gogen.File
	if isSingleFile {
		singleFile = g.NewFile(fileNamePattern, importPath)
		genFileHead(singleFile, packageName, primaryPkg.PkgPath, params.Interfaces)
	}

	for _, interfaceName := range params.Interfaces {
		var obj types.Object
		for _, pkg := range pkgs {
			obj = pkg.Types.Scope().Lookup(interfaceName)
			if obj != nil {
				break
			}
		}
		if obj == nil {
			return fmt.Errorf("type '%s' not found in package '%s'", interfaceName, primaryPkg.PkgPath)
		}
		objType := obj.Type().Underlying()
		log.Debugf("%s is %T which type is %T, and underlying type is %T", interfaceName, obj, obj.Type(), obj.Type().Underlying())
		iface, ok := objType.(*types.Interface)
		if !ok {
			return fmt.Errorf("can mock only interfaces, but '%s' is %s", interfaceName, objType.String())
		}

		file := singleFile
		if !isSingleFile {
			baseName := strings.ReplaceAll(fileNamePattern, placeHolder, strcase.ToSnake(interfaceName))
			path := filepath.Join(dstDir, baseName)
			file = g.NewFile(path, importPath)
			genFileHead(file, packageName, primaryPkg.PkgPath, []string{interfaceName})
		}

		generate(log, file, generateParams{
			InterfaceName: interfaceName,
			Interface:     iface,
			PackagePath:   primaryPkg.PkgPath,
		})
	}
	files := g.Files()
	var fileNames []string
	for _, f := range files {
		fileNames = append(fileNames, f.Path())
	}
	log.Debugf("Generating: %s", strings.Join(fileNames, ", "))

	err = g.WriteFiles(env.Fs)
	if err != nil {
		return fmt.Errorf("write files to '%s': %w", dstDir, err)
	}
	return nil
}

func debugLogPkgs(log *zap.SugaredLogger, pkgs []*packages.Package) {
	w := &bytes.Buffer{}
	p := func(format string, args ...interface{}) { _, _ = fmt.Fprintf(w, format, args...) }

	p("Loaded %v packages:\n", len(pkgs))
	for i, pkg := range pkgs {
		if i != 0 {
			p("\n")
		}
		p("- ID: %s\n", pkg.ID)
		p("  Name: %s\n", pkg.Name)
		p("  PkgPath: %s\n", pkg.PkgPath)

		if m := pkg.Module; m != nil {
			p("  Module:\n")
			p("    Path: %s\n", m.Path)
			p("    Dir: %s\n", m.Dir)
		}
	}
	log.Debugf(w.String())
}

func loadPackages(log *zap.SugaredLogger, env *Environment, src string) ([]*packages.Package, error) {
	log.Debugf("Loading package: %s", src)
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName |
			packages.NeedImports | // Workaround to fix "Unexpected package creation during export data loading". See https://github.com/golang/go/issues/45218
			packages.NeedModule |
			packages.NeedTypes,
		Dir:        env.Dir,
		Env:        env.Env,
		ParseFile:  nil, // TODO(skipor): optimize - remove static functions, methods bodies, to accelerate type checking
		Fset:       token.NewFileSet(),
		Tests:      true,
		BuildFlags: nil, // TODO(skipor)
	}, src)
	return pkgs, err
}

func genFileHead(f *gogen.File, packageName string, src string, interfaceNames []string) {
	f.L(`// Code generated by github.com/skipor/gmg - type-safe, fast and handy alternative GoMock generator. DO NOT EDIT.`)
	f.L(`// Source: `, src, `.`, strings.Join(interfaceNames, ","))
	f.L()
	f.L("package ", packageName)
	f.L()

	f.Import("reflect")
	f.Import("github.com/golang/mock/gomock")
}

type generateParams struct {
	InterfaceName string
	Interface     *types.Interface
	PackagePath   string
}

func generate(log *zap.SugaredLogger, f *gogen.File, p generateParams) {
	mockName := "Mock" + strcase.ToCamel(p.InterfaceName)
	recorderName := mockName + "MockRecorder"
	fg := &fileGenerator{
		File:           f,
		generateParams: p,
		log:            log,
		qualifier: func(pkg *types.Package) string {
			return f.QualifiedImportPath(gogen.ImportPath(pkg.Path()))
		},
		mockName:     mockName,
		recorderName: recorderName,
	}
	fg.generate()
}

type fileGenerator struct {
	*gogen.File
	generateParams
	mockName     string
	qualifier    func(pkg *types.Package) string
	recorderName string
	log          *zap.SugaredLogger
}

func (g *fileGenerator) generate() {
	g.genMock()
	g.genRecorder()
}

func (g *fileGenerator) genMock() {
	g.L(`
	// New`, g.mockName, ` creates a new GoMock for `, g.PackagePath, `.`, g.InterfaceName, `.
	func New`, g.mockName, `(ctrl *gomock.Controller) *`, g.mockName, ` {
		return &`, g.mockName, `{ctrl: ctrl}
	}`)

	g.L(`
	// `, g.mockName, ` is a GoMock of `, g.PackagePath, `.`, g.InterfaceName, `.
	type `, g.mockName, ` struct { ctrl *gomock.Controller }`)

	g.L(`
	// EXPECT returns GoMock recorder.
	func (m *`, g.mockName, `) EXPECT() *`, g.recorderName, ` {
		return (*`, g.recorderName, `)(m)
	}`)
	g.L()

	for i, n := 0, g.Interface.NumMethods(); i < n; i++ {
		g.genMockMethod(g.Interface.Method(i))
	}
}

func (g *fileGenerator) genMockMethod(method *types.Func) {
	scope := g.NewFuncScope()
	sig := method.Type().(*types.Signature)
	results := sig.Results()
	g.L(`// `, method.Name(), ` implements mocked interface.`)
	g.P(`func (m *`, g.mockName, `) `, method.Name(), `(`)
	paramsNames := g.genMockMethodParams(scope, sig)
	g.P(")")

	resultNames := g.genMockMethodFuncResults(scope, results)
	g.L(" {")

	g.L(`m.ctrl.T.Helper()`)
	g.P(`res := m.ctrl.Call(m, "`, method.Name(), `"`)
	for _, paramName := range paramsNames {
		g.P(", ", paramName)
	}
	g.L(")")
	for i := 0; i < results.Len(); i++ {
		result := results.At(i)
		name := resultNames[i]
		g.P(name, ` , _ := res[`, i, `].(`)
		g.writeType(result.Type())
		g.L(`)`)
	}

	g.P("return")
	for i, resultName := range resultNames {
		if i != 0 {
			g.P(",")
		}
		g.P(" ", resultName)
	}
	g.L()
	g.L("}")
	g.L()
}

func (g *fileGenerator) genMockMethodParams(scope *gogen.Scope, sig *types.Signature) []string {
	params := sig.Params()
	var paramsNames []string
	for i, l := 0, params.Len(); i < l; i++ {
		param := params.At(i)
		name := param.Name()
		if name == "" {
			name = "arg"
		}
		name = scope.Declare(name)
		paramsNames = append(paramsNames, name)
		if i != 0 {
			g.P(", ")
		}
		g.P(name, " ")

		typ := param.Type()
		if sig.Variadic() && i == l-1 {
			slice, ok := typ.(*types.Slice)
			if !ok {
				panic(fmt.Sprintf("last arg in variadic signature is not slice, but: %T", typ))
			}
			g.P("...")
			typ = slice.Elem()
		}
		g.writeType(typ)
	}
	return paramsNames
}

func (g *fileGenerator) genMockMethodFuncResults(scope *gogen.Scope, results *types.Tuple) []string {
	if results.Len() == 0 {
		return nil
	}
	g.P(" ")
	if results.Len() > 1 {
		g.P("(")
	}
	resultNames := g.genMockMethodResults(scope, results)
	if results.Len() > 1 {
		g.P(")")
	}
	return resultNames
}

func (g *fileGenerator) genMockMethodResults(scope *gogen.Scope, results *types.Tuple) []string {
	var resultNames []string
	for i := 0; i < results.Len(); i++ {
		if i != 0 {
			g.P(", ")
		}
		result := results.At(i)
		name := result.Name()
		if name != "" {
			g.P(name, " ")
		}
		resultNames = append(resultNames, g.resultName(scope, result, i))
		g.writeType(result.Type())
	}
	return resultNames
}

func (g *fileGenerator) resultName(scope *gogen.Scope, res *types.Var, i int) string {
	name := res.Name()
	switch name {
	case "", "_":
		name = fmt.Sprintf("res%v", i) // TODO(skipor): better names
	}
	return scope.Declare(name)
}

func (g *fileGenerator) genRecorder() {
	g.L(`
	// `, g.recorderName, ` is the mock recorder for `, g.mockName, `.
	type `, g.recorderName, ` `, g.mockName, `
	`)

	for i, n := 0, g.Interface.NumMethods(); i < n; i++ {
		g.genRecorderMethod(g.Interface.Method(i))
	}

	g.L(`
	func (r *`, g.recorderName, `) mock() *`, g.mockName, ` {
		return (*`, g.mockName, `)(r)
	}`)
}

func (g *fileGenerator) genRecorderMethod(method *types.Func) {
	callWrapperName := g.mockName + strcase.ToCamel(method.Name()) + "Call"
	scope := g.NewFuncScope()
	scope.Redeclare("r")
	sig := method.Type().(*types.Signature)
	g.L(`// `, method.Name(), ` makes call expectation.`)
	g.P(`func (r *`, g.recorderName, `) `, method.Name(), `(`)
	paramsNames := g.genRecorderMethodParams(sig.Params(), scope)
	g.L(`) `, callWrapperName, ` {`)
	g.L(`r.ctrl.T.Helper()`)

	callVarName := scope.Declare("call")
	g.P(callVarName, ` := r.ctrl.RecordCallWithMethodType(r.mock(), "`, method.Name(), `", reflect.TypeOf((*`, g.mockName, `)(nil).`, method.Name(), `)`)
	for _, paramName := range paramsNames {
		g.P(", ", paramName)
	}
	g.L(")")

	g.L("return ", callWrapperName, `{`, callVarName, `}`)
	g.L(`}`)
	g.L()
	g.genGomockCallWrapper(callWrapperName, method.Type().(*types.Signature))
}

func (g *fileGenerator) genRecorderMethodParams(params *types.Tuple, scope *gogen.Scope) []string {
	var paramsNames []string
	for i, l := 0, params.Len(); i < l; i++ {
		param := params.At(i)
		name := paramName(param, scope)
		paramsNames = append(paramsNames, name)
		if i != 0 {
			g.P(", ")
		}
		g.P(name, " interface{}")
	}
	return paramsNames
}

func (g *fileGenerator) genGomockCallWrapper(callWrapperName string, sig *types.Signature) {
	g.L(`
	// `, callWrapperName, ` is type safe wrapper of *gomock.Call.
	type `, callWrapperName, ` struct{ *gomock.Call }
	`)

	results := sig.Results()
	{
		g.P(`
		// DoAndReturn is type safe wrapper of *gomock.Call DoAndReturn.
		func (c `, callWrapperName, `) DoAndReturn(f func(`)
		lambdaScope := g.NewFuncScope()
		g.genMockMethodParams(lambdaScope, sig)
		g.P(`) `)
		g.genMockMethodFuncResults(lambdaScope, results)
		g.L(`) `, callWrapperName, ` {
			c.Call.DoAndReturn(f)
			return c
		}
		`)
		g.L()
	}
	{
		g.P(`
		// Do is type safe wrapper of *gomock.Call Do.
		func (c `, callWrapperName, `) Do(f func(`)
		lambdaScope := g.NewFuncScope()
		g.genMockMethodParams(lambdaScope, sig)
		g.L(`)) `, callWrapperName, ` {
			c.Call.Do(f)
		    return c
		}
		`)
		g.L()
	}
	if results.Len() > 0 {
		g.P(`
		// Return is type safe wrapper of *gomock.Call Return.
		func (c `, callWrapperName, `) Return(`)
		scope := g.NewFuncScope()
		scope.Redeclare("c")
		var resultNames []string
		for i := 0; i < results.Len(); i++ {
			if i != 0 {
				g.P(", ")
			}
			result := results.At(i)
			name := g.resultName(scope, result, i)
			g.P(name, " ")
			resultNames = append(resultNames, name)
			g.writeType(result.Type())
		}
		g.P(`) `, callWrapperName, ` {
			c.Call.Return(`)
		for i, name := range resultNames {
			if i != 0 {
				g.P(` ,`)
			}
			g.P(name)
		}
		g.L(`)
			return c
		}
		`)
		g.L()
	}
}

func (g *fileGenerator) writeType(t types.Type) {
	types.WriteType(g.Buffer(), t, g.qualifier)
}

func paramName(param *types.Var, scope *gogen.Scope) string {
	name := param.Name()
	if name == "" {
		name = "arg"
	}
	name = scope.Declare(name)
	return name
}

func errorsNum(pkgs []*packages.Package) int {
	var n int
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		n += len(pkg.Errors)
	})
	return n
}
