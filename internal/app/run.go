package app

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"go.uber.org/zap"
	"golang.org/x/tools/go/packages"

	"github.com/skipor/gmg/pkg/gmg"
	"github.com/skipor/gmg/pkg/gogen"
)

type params struct {
	Log *zap.SugaredLogger
	// Source is Go package to search for interfaces. See flag description for details.
	Source string
	// Destination is directory or file relative path or pattern. See flag description for details.
	Destination string
	// Package is package name in generated files. See flag description for details.
	Package string

	Selector interfaceSelector
}

type goGenerateEnv struct {
	// GOLINE set by 'go generate' to line number of the directive in the source file.
	GOLINE int
	// GOFILE set by 'go generate' to the base name of the file.
	GOFILE string
	// GOPACKAGE the name of the package of the file containing the directive.
	GOPACKAGE string
}

func run(env *Environment, params *params) error {
	log := params.Log
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
	files, err := generateAll(env, pkgs, params)
	if err != nil {
		return err
	}
	var fileNames []string
	for _, f := range files {
		fileNames = append(fileNames, f.Path())
	}
	log.Debugf("Generating: %s", strings.Join(fileNames, ", "))
	for _, f := range files {
		err := f.WriteFile(env.Fs)
		if err != nil {
			return fmt.Errorf("file %s: %w", f.Path(), err)
		}
		_, _ = fmt.Fprintf(env.Stderr, "%s\n", f.Path())
	}
	return nil
}

func generateAll(env *Environment, pkgs []*packages.Package, params *params) ([]*gogen.File, error) {
	log := params.Log
	srcPrimaryPkg := pkgs[0]
	dstDir := strings.TrimPrefix(params.Destination, ".")
	fileNamePattern := placeHolder + ".go"
	if path.Ext(dstDir) == ".go" {
		dstDir, fileNamePattern = path.Split(dstDir)
	}
	dstDir = strings.ReplaceAll(dstDir, placeHolder, srcPrimaryPkg.Name)

	packageName, err := getPackageName(log, params.Package, dstDir, srcPrimaryPkg, env)
	if err != nil {
		return nil, fmt.Errorf("get generated file package name: %w", err)
	}
	importPath := path.Join(srcPrimaryPkg.PkgPath, dstDir)

	g := gmg.NewGMG(log)

	ifaces, err := selectInterfaces(log, pkgs, params.Selector)
	if err != nil {
		return nil, err
	}

	isSingleFile := !strings.Contains(fileNamePattern, placeHolder)
	if isSingleFile {
		g.GenerateFile(gmg.GenerateFileParams{
			FilePath:    fileNamePattern,
			ImportPath:  importPath,
			PackageName: packageName,
			Interfaces:  ifaces,
		})
	} else {
		for _, iface := range ifaces {
			baseName := strings.ReplaceAll(fileNamePattern, placeHolder, strcase.ToSnake(iface.Name))
			filePath := filepath.Join(dstDir, baseName)
			g.GenerateFile(gmg.GenerateFileParams{
				FilePath:    filePath,
				ImportPath:  importPath,
				PackageName: packageName,
				Interfaces:  []gmg.Interface{iface},
			})
		}
	}
	return g.Files(), nil
}

func getPackageName(log *zap.SugaredLogger, packageNameTemplate string, dstDir string, srcPrimaryPkg *packages.Package, env *Environment) (string, error) {
	const defaultPackageNameTemplate = "mocks_{}"
	if packageNameTemplate != "" {
		log.Debugf("Package name template explisitly set - using it")
		return executePackageNameTemplate(packageNameTemplate, srcPrimaryPkg), nil
	}
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		log.Debugf("Package name is not set, but destination dir is not exist - using default")
		return executePackageNameTemplate(defaultPackageNameTemplate, srcPrimaryPkg), nil
	}

	absDstDir, err := filepath.Abs(dstDir)
	if err != nil {
		return "", fmt.Errorf("destination dir '%s' abs: %w", dstDir, err)
	}

	// TODO(skipor): optimise - check, maybe it already loaded in pkgs

	log.Debugf("Package name is not set, and destination dir exists - trying to load go package, to get its name, to use it in generated files")
	dstDirPkgs, err := packages.Load(&packages.Config{
		Mode:       packages.NeedName,
		Dir:        env.Dir,
		Env:        env.Env,
		BuildFlags: nil, // TODO(skipor)
	}, absDstDir)

	if err != nil {
		return "", fmt.Errorf("failed to load destination dir '%s' go package to deduce package name: %w\n."+
			"\tSet package name explicitly via --pkg flag.", dstDir, err)
	}
	debugLogPkgs(log, dstDirPkgs)
	dstPackageName := dstDirPkgs[0].Name
	if dstPackageName == "" {
		log.Debugf("Destination dir package has no package name - seems there are no go files. Falling back to default package name template")
		return executePackageNameTemplate(defaultPackageNameTemplate, srcPrimaryPkg), nil
	}

	log.Debugf("Going to use destination dir package package name '%s'", dstPackageName)
	return dstPackageName, nil
}

func executePackageNameTemplate(tmpl string, srcPrimaryPkg *packages.Package) string {
	return strings.ReplaceAll(tmpl, placeHolder, srcPrimaryPkg.Name)
}
