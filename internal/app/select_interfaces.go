package app

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/tools/go/packages"

	"github.com/skipor/gmg/pkg/gmg"
)

type interfaceSelector struct {
	names    []string
	all      bool
	allFile  bool
	goGenEnv goGenerateEnv
}

func selectInterfaces(log *zap.SugaredLogger, pkgs []*packages.Package, sel interfaceSelector) ([]gmg.Interface, error) {
	log.Debugf("Selecting interfaces: %+v", sel)
	if len(sel.names) != 0 {
		return selectInterfacesByNames(log, pkgs, sel.names)
	}
	if sel.all {
		return selectAllPrimaryPackageInterfaces(log, pkgs)
	}
	if sel.allFile {
		if !sel.goGenEnv.isSet() {
			log.Panic("Validation failed: 'all-file' selector passed but no 'go generate' env set")
		}
		return selectAllFileInterfaces(log, pkgs, sel.goGenEnv)
	}
	if !sel.goGenEnv.isSet() {
		log.Panic("Validation failed: neither selector passed nor 'go generate' env set")
	}
	return selectInterfaceCorrespondingToGoGenerateComment(log, pkgs, sel.goGenEnv)
}

func selectInterfacesByNames(log *zap.SugaredLogger, pkgs []*packages.Package, interfaceNames []string) ([]gmg.Interface, error) {
	srcPrimaryPkg := pkgs[0]
	log.Infof("Selecting package '%s' interface names: %s", srcPrimaryPkg.PkgPath, interfaceNames)
	var ifaces []gmg.Interface
	for _, interfaceName := range interfaceNames {
		var obj types.Object
		var objPkg *packages.Package
		for _, pkg := range pkgs {
			obj = pkg.Types.Scope().Lookup(interfaceName)
			if obj != nil {
				objPkg = pkg
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

		ifaces = append(ifaces, gmg.Interface{
			Name:       interfaceName,
			ImportPath: objPkg.PkgPath,
			Type:       iface,
		})
	}
	return ifaces, nil
}

func selectInterfaceCorrespondingToGoGenerateComment(log *zap.SugaredLogger, pkgs []*packages.Package, goGenEnv goGenerateEnv) ([]gmg.Interface, error) {
	pkg, file, fset, parseErr, err := parseGoGenerateCommentFile(log, pkgs, goGenEnv)
	if err != nil {
		return nil, err
	}

	typeSpec, err := correspondingTypeSpec(file, fset, goGenEnv.GOLINE, parseErr)
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
		return nil, fmt.Errorf("`//go:generate` comment corresponding to type declaration at %s, which is not interface but: %s %s",
			pos(fset, typeSpec),
			typ.String(),
			typ.Underlying().String(),
		)
	}

	return []gmg.Interface{{
		Name:       typeName,
		ImportPath: pkg.PkgPath,
		Type:       typ.Underlying().(*types.Interface),
	}}, nil
}

func parseGoGenerateCommentFile(log *zap.SugaredLogger, pkgs []*packages.Package, goGenEnv goGenerateEnv) (*packages.Package, *ast.File, *token.FileSet, error, error) {
	pkg := getPackageByKind(pkgs, goGenEnv.packageKind())
	if pkg == nil {
		return nil, nil, nil, nil, fmt.Errorf(
			"'go generate' env variables indicates that package kind is '%s' but it is not found in loaded packages.\n"+
				goEnvErrMsg,
			goGenEnv.packageKind(),
		)
	}

	// AST can be loaded by packages.Load with packages.Syntax mode,
	// and that cause packages.Package.TypesInfo computation for all package files.
	// So, seems that it is cheaper to parse only GOFILE by ourselves.

	file, fset, parseErr := parseGOFILE(pkg, goGenEnv.GOFILE)
	if parseErr != nil {
		if file == nil {
			return nil, nil, nil, nil, fmt.Errorf("GOFILE '%s' parse: %w", goGenEnv.GOFILE, parseErr)
		}
		// These errors should be already logged after package load.
		log.Infof("GOFILE '%s' parse recoverable errors: %+v", goGenEnv.GOFILE, parseErr)
		return nil, nil, nil, nil, parseErr
	}
	return pkg, file, fset, parseErr, nil
}

func correspondingTypeSpec(file *ast.File, fset *token.FileSet, goline int, parseErr error) (*ast.TypeSpec, error) {
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
			pos(fset, decl),
		)
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

func selectAllPrimaryPackageInterfaces(log *zap.SugaredLogger, pkgs []*packages.Package) ([]gmg.Interface, error) {
	// TODO(skipor): support both test packages from flags (--primary-pkg --test-pkg, --black-box-test-pkg ?)
	//   or --all-test for both test packages?
	pkg := getPackageByKind(pkgs, primaryPackageKind)
	if pkg == nil {
		log.Infof("There is no non *_test.go files, so --all selector doesn't select anythig")
		return nil, nil
	}
	ifaces := selectAllPkgInterfaces(log, pkg)
	if len(ifaces) == 0 {
		log.Infof("No interfaces found in non *_test.go files")
		return nil, nil
	}
	return ifaces, nil
}

func selectAllPkgInterfaces(log *zap.SugaredLogger, pkg *packages.Package) []gmg.Interface {
	// TODO(skipor): handle unexported methods and interfaces
	log.Debugf("Selecting all interfaces from package %s of kind '%s'", pkg.ID, getPackageKind(pkg))
	var ifaces []gmg.Interface
	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := pkg.Types.Scope().Lookup(name)
		objType := obj.Type().Underlying()
		log.Debugf("Name %s is %T which type is %T, and underlying type is %T", name, obj, obj.Type(), obj.Type().Underlying())
		iface, ok := objType.(*types.Interface)
		if !ok {
			continue
		}
		ifaces = append(ifaces, gmg.Interface{
			Name:       name,
			ImportPath: pkg.PkgPath,
			Type:       iface,
		})
	}
	return ifaces
}

func selectAllFileInterfaces(log *zap.SugaredLogger, pkgs []*packages.Package, goGenEnv goGenerateEnv) ([]gmg.Interface, error) {
	pkg, file, fset, _, err := parseGoGenerateCommentFile(log, pkgs, goGenEnv)
	if err != nil {
		return nil, err
	}
	var ifaces []gmg.Interface
	for _, decl := range file.Decls {
		decl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if decl.Tok != token.TYPE {
			continue
		}
		for _, spec := range decl.Specs {
			spec := spec.(*ast.TypeSpec)
			name := spec.Name.Name
			obj := pkg.Types.Scope().Lookup(name)
			if obj == nil {
				log.Warnf("Type '%s' declaration at %s type is not deduced. "+
					"See package load errors warning for details.",
					name,
					pos(fset, spec),
				)
				continue
			}
			log.Debugf("Name %s is %T which type is %T, and underlying type is %T", name, obj, obj.Type(), obj.Type().Underlying())
			typ := obj.Type()
			if !types.IsInterface(typ) {
				continue
			}
			ifaces = append(ifaces, gmg.Interface{
				Name:       name,
				ImportPath: pkg.PkgPath,
				Type:       typ.Underlying().(*types.Interface),
			})
		}
	}
	return ifaces, nil
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
	file, err := parser.ParseFile(fset, gofile, fileData, parser.ParseComments|parser.DeclarationErrors)
	unrecoverable := err != nil && file == nil
	if unrecoverable {
		return nil, nil, err
	}
	return file, fset, err
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

func pos(fset *token.FileSet, node interface{ Pos() token.Pos }) string {
	position := fset.Position(node.Pos())
	return fmt.Sprintf("%s:%v", position.Filename, position.Line)
}

const goEnvErrMsg = "That may happen because of:\n" +
	"- No 'go generate' run but GOFILE/GOLINE/GOPACKAGE set manually. If so, don't do that.\n" +
	"- File with `//go:generate gmg` ignored on load because of `//+build` annotations. If so, put comment to another file or pass interface name explicitly.\n" +
	"- Bug. If so, fill issue on GitHub with log of '--debug' run."

func ifaceNames(ifaces []gmg.Interface) []string {
	var names []string
	for _, iface := range ifaces {
		names = append(names, iface.Name)
	}
	return names
}
