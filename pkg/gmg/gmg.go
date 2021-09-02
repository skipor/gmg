package gmg

import (
	"bytes"
	"fmt"
	"go/types"

	"github.com/iancoleman/strcase"
	"go.uber.org/zap"

	"github.com/skipor/gmg/pkg/gogen"
)

func NewGMG(log *zap.SugaredLogger) *GMG {
	return &GMG{
		log: log,
		gen: gogen.NewGenerator(),
	}
}

type GMG struct {
	log *zap.SugaredLogger
	gen *gogen.Generator
}

type Interface struct {
	Name string
	Type *types.Interface
	// ImportPath is to add interface source to generated godoc.
	ImportPath string
}

type GenerateFileParams struct {
	FilePath    string
	ImportPath  string
	PackageName string
	Interfaces  []Interface
	Options     GenerateOptions
}

type GenerateOptions struct {
	// TODO:
}

func (g *GMG) GenerateFile(p GenerateFileParams) {
	file := g.gen.NewFile(p.FilePath, gogen.ImportPath(p.ImportPath))
	genFileHeadV2(file, p.PackageName, p.Interfaces)
	for _, iface := range p.Interfaces {
		generate(g.log, file, generateParams{
			InterfaceName: iface.Name,
			Interface:     iface.Type,
			PackagePath:   iface.ImportPath,
		})
	}
}

func genFileHeadV2(f *gogen.File, packageName string, interfaces []Interface) {
	f.L(`// Code generated by github.com/skipor/gmg - type-safe, fast and handy alternative GoMock generator. DO NOT EDIT.`)
	f.P(`// Source: `)
	var prevImportPath string
	for i, iface := range interfaces {
		if iface.ImportPath == prevImportPath {
			f.P(",", iface.Name)
			continue
		}
		prevImportPath = iface.ImportPath
		if i != 0 {
			f.P(" ;")
		}
		f.P(iface.ImportPath, ".", iface.Name)
	}
	f.L()
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
	// Indent with spaces, to make comment go doc code block.
	g.P(`//   `, method.Name())
	writeSignature(g.Buffer(), method.Type().(*types.Signature), func(p *types.Package) string {
		return p.Name()
	})
	g.L()
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

func writeSignature(buf *bytes.Buffer, sig *types.Signature, qf types.Qualifier) {
	writeTuple(buf, sig.Params(), sig.Variadic(), qf)
	res := sig.Results()
	if res.Len() == 0 {
		return
	}

	buf.WriteByte(' ')
	firstRes := res.At(0)
	if res.Len() == 1 && firstRes.Name() == "" {
		types.WriteType(buf, firstRes.Type(), qf)
		return
	}
	writeTuple(buf, res, false, qf)
}

func writeTuple(buf *bytes.Buffer, tup *types.Tuple, variadic bool, qf types.Qualifier) {
	buf.WriteByte('(')
	if tup != nil {
		for i, l := 0, tup.Len(); i < l; i++ {
			v := tup.At(i)
			if i > 0 {
				buf.WriteString(", ")
			}
			if v.Name() != "" {
				buf.WriteString(v.Name())
				buf.WriteByte(' ')
			}
			typ := v.Type()
			if variadic && i == l-1 {
				if slice, ok := typ.(*types.Slice); ok {
					buf.WriteString("...")
					typ = slice.Elem()
				}
			}
			types.WriteType(buf, typ, qf)
		}
	}
	buf.WriteByte(')')
}
