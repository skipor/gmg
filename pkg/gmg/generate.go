package gmg

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
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
	srcPrimaryPkg := pkgs[0]
	dstDir := strings.TrimPrefix(params.Destination, ".")
	fileNamePattern := placeHolder + ".go"
	if path.Ext(dstDir) == ".go" {
		dstDir, fileNamePattern = path.Split(dstDir)
	}
	dstDir = strings.ReplaceAll(dstDir, placeHolder, srcPrimaryPkg.Name)
	packageName := strings.ReplaceAll(params.Package, placeHolder, srcPrimaryPkg.Name)
	importPath := gogen.ImportPath(path.Join(srcPrimaryPkg.PkgPath, dstDir))

	isSingleFile := !strings.Contains(fileNamePattern, placeHolder)
	var singleFile *gogen.File
	if isSingleFile {
		singleFile = g.NewFile(fileNamePattern, importPath)
		genFileHead(singleFile, packageName, srcPrimaryPkg.PkgPath, params.InterfaceNames)
	}

	ifaces, err := findInterfaces(pkgs, params)
	if err != nil {
		return err
	}

	for _, iface := range ifaces {
		file := singleFile
		if !isSingleFile {
			baseName := strings.ReplaceAll(fileNamePattern, placeHolder, strcase.ToSnake(iface.name))
			filePath := filepath.Join(dstDir, baseName)
			file = g.NewFile(filePath, importPath)
			genFileHead(file, packageName, srcPrimaryPkg.PkgPath, []string{iface.name})
		}
		generate(log, file, generateParams{
			InterfaceName: iface.name,
			Interface:     iface.typ,
			PackagePath:   srcPrimaryPkg.PkgPath,
		})

	}
	return nil
}

const goEnvErrMsg = "That may happen because of:\n" +
	"- No 'go generate' run but GOFILE/GOLINE/GOPACKAGE set manually. If so, don't do that.\n" +
	"- File with `//go:generate gmg` ignored on load because of `//+build` annotations. If so, put comment to another file or pass interface name explicitly.\n" +
	"- Bug. If so, fill issue on GitHub with log of '--debug' run."

func findInterfaces(pkgs []*packages.Package, params *params) ([]namedInterface, error) {
	log := params.Log
	if len(params.InterfaceNames) != 0 {
		return findInterfacesByNames(log, pkgs, params.InterfaceNames)
	}

	goGenEnv := params.GoGenerateEnv
	pkg := getPackageByKind(pkgs, goGenEnv.packageKind())
	if pkg == nil {
		return nil, fmt.Errorf(
			"'go generate' env variables indicates that package kind is '%s' but it is not found in loaded packages.\n"+
				goEnvErrMsg,
			goGenEnv.packageKind(),
		)
	}

	// Need to find GOFILE declaration just after GOLINE.
	// GOFILE AST needed for that.
	// It can be loaded by packages.Load with packages.Syntax mode,
	// and that cause packages.Package.TypesInfo computation for all package files.
	// So, seems that it is cheaper to parse only GOFILE by ourselves.

	file, fset, parseErr := parseGOFILE(pkg, goGenEnv.GOFILE)
	if parseErr != nil {
		if file == nil {
			return nil, fmt.Errorf("GOFILE '%s' parse: %w", goGenEnv.GOFILE, parseErr)
		}
		// These errors should be already logged after package load.
		log.Infof("GOFILE '%s' parse recoverable errors: %+v", goGenEnv.GOFILE, parseErr)

		return nil, parseErr
	}
	typeSpec, err := correspondingTypeSpec(fset, file, goGenEnv.GOLINE, parseErr)
	if err != nil {
		return nil, err
	}
	typeSpec.Pos()

	typeName := typeSpec.Name.Name
	obj := pkg.Types.Scope().Lookup(typeName)
	if obj == nil {
		return nil, fmt.Errorf("`//go:generate` comment corresponding to type declaration at %s, which type lookup failed for unxpected reason.\n"+
			"Please fill GitHub issue with your code example.\n"+
			"Pass interface name as argument for workarount.",
			pos(fset, typeSpec))
	}
	typ := obj.Type()
	if !types.IsInterface(typ) {
		return nil, fmt.Errorf("`//go:generate` comment corresponding to type declaration at %s, which is not interface but: %s",
			pos(fset, typeSpec),
			typ.String(),
		)
	}

	return []namedInterface{{
		name: typeName,
		typ:  typ.Underlying().(*types.Interface),
	}}, nil

}

func pos(fset *token.FileSet, node interface{ Pos() token.Pos }) string {
	position := fset.Position(node.Pos())
	return fmt.Sprintf("%s:%v", position.Filename, position.Line)
}

func correspondingTypeSpec(fset *token.FileSet, file *ast.File, goline int, parseErr error) (*ast.TypeSpec, error) {
	decl := lineDecl(fset, file, goline)
	if decl == nil {
		return nil, fmt.Errorf("there is no type declarations after `//go:generate` comment.\n" +
			"Put it just above interface type declaration or pass interface name(s) as argument(s).")
	}
	switch decl := decl.(type) {
	case *ast.GenDecl:
		if decl.Tok != token.TYPE {
			tokenStr := strings.ToLower(decl.Tok.String())
			return nil, fmt.Errorf("`//go:generate` comment corresponding to declaration at %s which is not `type`, but `%s`.\n"+
				"Put it just above interface type declaration or pass interface name(s) as argument(s).",
				pos(fset, decl), tokenStr)
		}

		genSpec := lineDeclSpec(fset, decl, goline)
		if genSpec == nil {
			// Example:
			//   type (
			//       Foo interface{ Bar() }
			//   //go:generate gmg
			//   )
			return nil, fmt.Errorf("`//go:generate` comment corresponding to declaration at %s, but there is no declaration spec inside it after comment line.\n"+
				"Put it just above interface type declaration or pass interface name(s) as argument(s).",
				pos(fset, decl))
		}
		return genSpec.(*ast.TypeSpec), nil
	case *ast.FuncDecl:
		return nil, fmt.Errorf("`//go:generate` comment corresponding to declaration at %s which is not `type`, but `func`.\n"+
			"Put it just above interface type declaration or pass interface name(s) as argument(s).",
			pos(fset, decl))
	case *ast.BadDecl:
		return nil, fmt.Errorf("failed to parse declaration next to `//go:generate` at %s.\n"+
			"Parse errors: %w",
			pos(fset, decl), parseErr)
	default:
		return nil, fmt.Errorf("unexpected declaration %T after `//go:generate` comment at %s.\n"+
			"Please fill GitHub issue with your code example.\n"+
			"Pass interface name as argument for workarount.",
			decl, pos(fset, decl))
	}
}

func lineDecl(fset *token.FileSet, file *ast.File, goline int) ast.Decl {
	for _, decl := range file.Decls {
		end := fset.Position(decl.End())
		// Checking end, because line can be inside `type ()` declaration.
		if end.Line > goline {
			return decl
		}
	}
	return nil
}

func lineDeclSpec(fset *token.FileSet, decl *ast.GenDecl, goline int) ast.Spec {
	for _, spec := range decl.Specs {
		if fset.Position(spec.Pos()).Line > goline {
			return spec
		}
	}
	return nil
}

func parseGOFILE(pkg *packages.Package, gofile string) (*ast.File, *token.FileSet, error) {
	filePath := gofilePath(pkg, gofile)
	if filePath == "" {
		return nil, nil, fmt.Errorf(
			"'go generate' env variables indicates that `//go:generate gmg` comment located in file '%s' but it is not found in loaded package.\n"+
				goEnvErrMsg,
			gofile,
		)
	}

	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("GOFILE '%s' read: %w", filePath, err)
	}

	fset := token.NewFileSet()
	ast, err := parser.ParseFile(fset, gofile, fileData, parser.ParseComments|parser.DeclarationErrors)
	unrecoverable := err != nil && ast == nil
	if unrecoverable {
		return nil, nil, err
	}
	return ast, fset, err
}

func gofilePath(pkg *packages.Package, gofile string) string {
	for _, absPath := range pkg.CompiledGoFiles {
		baseName := filepath.Base(absPath)
		if baseName == gofile {
			return absPath
		}
	}
	return ""
}

func findInterfacesByNames(log *zap.SugaredLogger, pkgs []*packages.Package, interfaceNames []string) ([]namedInterface, error) {
	srcPrimaryPkg := pkgs[0]
	var ifaces []namedInterface
	for _, interfaceName := range interfaceNames {
		var obj types.Object
		for _, pkg := range pkgs {
			obj = pkg.Types.Scope().Lookup(interfaceName)
			if obj != nil {
				break
			}
		}
		if obj == nil {
			msg := fmt.Sprintf("type '%s' was not found in package '%s'", interfaceName, srcPrimaryPkg.PkgPath)
			if packagesErrorsNum(pkgs) > 0 {
				msg += ".\nPay attention to the package loading errors that were warned about above, they may be the cause of this."
			}
			return nil, fmt.Errorf(msg)
		}
		objType := obj.Type().Underlying()
		log.Debugf("%s is %T which type is %T, and underlying type is %T", interfaceName, obj, obj.Type(), obj.Type().Underlying())
		iface, ok := objType.(*types.Interface)
		if !ok {
			return nil, fmt.Errorf("can mock only interfaces, but '%s' is %s", interfaceName, objType.String())
		}

		ifaces = append(ifaces, namedInterface{
			name: interfaceName,
			typ:  iface,
		})
	}
	return ifaces, nil
}

type namedInterface struct {
	name string
	typ  *types.Interface
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
