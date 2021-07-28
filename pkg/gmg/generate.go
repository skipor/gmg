package gmg

import (
	"fmt"
	"go/types"
	"path"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"go.uber.org/zap"
	"golang.org/x/tools/go/packages"

	"github.com/skipor/gmg/pkg/gogen"
)

func generateAll(g *gogen.Generator, pkgs []*packages.Package, params *params) error {
	log := params.Log
	primaryPkg := pkgs[0]
	dstDir := strings.TrimPrefix(params.Destination, ".")
	fileNamePattern := placeHolder + ".go"
	if path.Ext(dstDir) == ".go" {
		dstDir, fileNamePattern = path.Split(dstDir)
	}
	dstDir = strings.ReplaceAll(dstDir, placeHolder, primaryPkg.Name)
	packageName := strings.ReplaceAll(params.Package, placeHolder, primaryPkg.Name)
	importPath := gogen.ImportPath(path.Join(primaryPkg.PkgPath, dstDir))

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
	return nil
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

const (
	mockReceiver     = "m_"
	recorderReceiver = "r_"
	callReceiver     = "c_"
)

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
	func (`, mockReceiver, ` *`, g.mockName, `) EXPECT() *`, g.recorderName, ` {
		return (*`, g.recorderName, `)(`, mockReceiver, `)
	}`)
	g.L()

	for i, n := 0, g.Interface.NumMethods(); i < n; i++ {
		g.genMockMethod(g.Interface.Method(i))
	}
}

func (g *fileGenerator) genMockMethod(method *types.Func) {
	scope := g.NewFuncScope()
	receiver := scope.Declare(mockReceiver)
	sig := method.Type().(*types.Signature)
	results := sig.Results()
	g.L(`// `, method.Name(), ` implements mocked interface.`)
	g.P(`func (`, receiver, ` *`, g.mockName, `) `, method.Name(), `(`)
	paramsNames := g.genMockMethodParams(scope, sig)
	g.P(")")

	resultNames := g.genMockMethodFuncResults(scope, results)
	g.L(" {")

	res := scope.Declare("res_")
	g.L(receiver, `.ctrl.T.Helper()`)
	if results.Len() > 0 {
		g.P(res, ` := `)
	}
	g.P(receiver, `.ctrl.Call(`, receiver, `, "`, method.Name(), `"`)
	for _, paramName := range paramsNames {
		g.P(", ", paramName)
	}
	g.L(")")
	for i := 0; i < results.Len(); i++ {
		result := results.At(i)
		name := resultNames[i]
		g.P(name, ` , _ `)
		if noName(result) {
			g.P(":")
		}
		g.P(`= `, res, `[`, i, `].(`)
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

func noName(v *types.Var) bool {
	return emptyOrUnderscore(v.Name())
}

func emptyOrUnderscore(name string) bool {
	return name == "" || name == "_"
}

func (g *fileGenerator) genMockMethodParams(scope *gogen.Scope, sig *types.Signature) []string {
	params := sig.Params()
	var paramsNames []string
	for i, l := 0, params.Len(); i < l; i++ {
		param := params.At(i)
		name := paramName(param, scope)
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
	g.P("(")
	resultNames := g.genMockMethodResults(scope, results)
	g.P(")")
	return resultNames
}

func (g *fileGenerator) genMockMethodResults(scope *gogen.Scope, results *types.Tuple) []string {
	var resultNames []string
	for i := 0; i < results.Len(); i++ {
		if i != 0 {
			g.P(", ")
		}
		result := results.At(i)
		name := g.resultName(scope, result, i)
		if result.Name() != "" {
			sigName := name
			if result.Name() == "_" {
				sigName = "_"
			}
			g.P(sigName, " ")
		}
		resultNames = append(resultNames, name)
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
	func (`, recorderReceiver, `*`, g.recorderName, `) mock() *`, g.mockName, ` {
		return (*`, g.mockName, `)(`, recorderReceiver, `)
	}`)
}

func (g *fileGenerator) genRecorderMethod(method *types.Func) {
	callWrapperName := g.mockName + strcase.ToCamel(method.Name()) + "Call"
	scope := g.NewFuncScope()
	receiver := scope.Declare(recorderReceiver)
	sig := method.Type().(*types.Signature)
	g.L(`// `, method.Name(), ` makes call expectation.`)
	g.P(`func (`, receiver, ` *`, g.recorderName, `) `, method.Name(), `(`)
	paramsNames := g.genRecorderMethodParams(sig.Params(), scope)
	g.L(`) `, callWrapperName, ` {`)
	g.L(receiver, `.ctrl.T.Helper()`)

	callVarName := scope.Declare("call")
	g.P(callVarName, ` := `, receiver, `.ctrl.RecordCallWithMethodType(`, receiver, `.mock(), "`, method.Name(), `", reflect.TypeOf((*`, g.mockName, `)(nil).`, method.Name(), `)`)
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
	var paramNames []string
	for i, l := 0, params.Len(); i < l; i++ {
		param := params.At(i)
		name := paramName(param, scope)
		paramNames = append(paramNames, name)
		if i != 0 {
			g.P(", ")
		}
		g.P(name, " interface{}")
	}
	return paramNames
}

func (g *fileGenerator) genGomockCallWrapper(callWrapperName string, sig *types.Signature) {
	g.L(`
	// `, callWrapperName, ` is type safe wrapper of *gomock.Call.
	type `, callWrapperName, ` struct{ *gomock.Call }
	`)

	results := sig.Results()
	{
		scope := g.NewFuncScope()
		receiver := scope.Declare(callReceiver)
		g.P(`
		// DoAndReturn is type safe wrapper of *gomock.Call DoAndReturn.
		func (`, receiver, ` `, callWrapperName, `) DoAndReturn(f func(`)
		g.genMockMethodParams(scope, sig)
		g.P(`) `)
		g.genMockMethodFuncResults(scope, results)
		g.L(`) `, callWrapperName, ` {
			`, receiver, `.Call.DoAndReturn(f)
			return `, receiver, `
		}
		`)
		g.L()
	}
	{
		scope := g.NewFuncScope()
		receiver := scope.Declare(callReceiver)
		g.P(`
		// Do is type safe wrapper of *gomock.Call Do.
		func (`, receiver, ` `, callWrapperName, `) Do(f func(`)
		g.genMockMethodParams(scope, sig)
		g.L(`)) `, callWrapperName, ` {
			`, receiver, `.Call.Do(f)
		    return `, receiver, `
		}
		`)
		g.L()
	}
	if results.Len() > 0 {
		scope := g.NewFuncScope()
		receiver := scope.Declare(callReceiver)
		g.P(`
		// Return is type safe wrapper of *gomock.Call Return.
		func (`, receiver, ` `, callWrapperName, `) Return(`)
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
			`, receiver, `.Call.Return(`)
		for i, name := range resultNames {
			if i != 0 {
				g.P(` ,`)
			}
			g.P(name)
		}
		g.L(`)
			return `, receiver, `
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
	if emptyOrUnderscore(name) {
		// TODO(skipor): well known names: ctx, err
		// TODO(skipor): deduce better name from type
		name = "arg"
	}
	return scope.Declare(name)
}