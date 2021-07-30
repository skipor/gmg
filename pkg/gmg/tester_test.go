package gmg

import (
	"fmt"
	"go/format"
	"os/exec"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages/packagestest"
)

var x = packagestest.Modules

type M = packagestest.Module

func newTester(t *testing.T, modules ...M) *Tester {
	for _, m := range modules {
		formatModuleFiles(t, m)
	}
	e := export(t, modules...)
	return &Tester{
		exported: e,
	}
}

func formatModuleFiles(t *testing.T, m M) {
	for path, data := range m.Files {
		str, ok := data.(string)
		if !ok {
			continue
		}
		bytes, err := format.Source([]byte(str))
		if err != nil {
			if bytes == nil {
				t.Fatalf("Module '%s' file '%s' format failed: %s", m.Name, path, err)
			}
			t.Logf("WARN: module '%s' file '%s' format errors: %s", m.Name, path, err)
		}
		fmt.Printf("After format:\n%s\n", bytes)
		m.Files[path] = string(bytes)
	}
}

func (tr *Tester) Gmg(t *testing.T, args ...string) *RunResult {
	args = append(args, "--debug")
	t.Logf("Run: gmg %s", strings.Join(args, " "))
	FS := &afero.MemMapFs{}
	env := &Environment{
		Args:   args,
		Stderr: testWriter{t},
		Dir:    tr.exported.Config.Dir,
		Env:    tr.exported.Config.Env,
		Fs:     FS,
	}
	exitCode := Main(env)

	return &RunResult{
		t:        t,
		ExitCode: exitCode,
		FS:       FS,
	}
}

func (tr *Tester) GoGenerate(t *testing.T) *RunResult {
	t.Helper()
	dir := tr.exported.Config.Dir
	cmd := exec.Command("go", "generate", "./...")
	w := testWriter{t}
	cmd.Stdout = w
	cmd.Stderr = w
	cmd.Dir = dir
	cmd.Env = append(tr.exported.Config.Env, "GMG_DEBUG=true")

	beforeFsMap := fsToMap(t, afero.NewBasePathFs(afero.NewOsFs(), dir))

	t.Logf("Run: %s", cmd.String())
	err := cmd.Run()
	var exitCode int
	if err != nil {
		err, ok := err.(*exec.ExitError)
		if ok {
			t.Fatalf("Unexpected run fail: %+v", err)
		}
		exitCode = err.ExitCode()
	}

	t.Logf("Tree of '%s' after '%s': %s", dir, cmd.String(), dirTree(t, dir))

	afterFsMap := fsToMap(t, afero.NewBasePathFs(afero.NewOsFs(), dir))

	for path := range beforeFsMap {
		_, ok := afterFsMap[path]
		if !ok {
			t.Fatalf("file '%s' removed after '%s' run", path, cmd.String())
		}
	}
	changed := afero.NewMemMapFs().(*afero.MemMapFs)
	for path, before := range afterFsMap {
		after, ok := beforeFsMap[path]
		if !ok || before != after {
			err := afero.WriteFile(changed, path, []byte(after), 0644)
			require.NoError(t, err)
		}
	}

	return &RunResult{
		t:        t,
		ExitCode: exitCode,
		FS:       changed,
	}
}

type RunResult struct {
	t        *testing.T
	ExitCode int
	FS       *afero.MemMapFs
}

func (r *RunResult) Succeed() *RunResult {
	require.Zero(r.t, r.ExitCode)
	return r
}

func (r *RunResult) Fail() *RunResult {
	require.NotZero(r.t, r.ExitCode)
	return r
}

func (r *RunResult) Files(expectedFiles ...string) *RunResult {
	var actualFiles []string
	for file := range fsToMap(r.t, r.FS) {
		actualFiles = append(actualFiles, file)
	}
	sort.Strings(expectedFiles)
	sort.Strings(actualFiles)

	diff := cmp.Diff(expectedFiles, actualFiles)
	if len(diff) > 0 {
		r.t.Fatalf("Expecated and actual files diff:\n%s", diff)
	}
	return r
}

func (r *RunResult) Golden() *RunResult {
	golden.Dir(r.t, r.FS)
	return r
}

type Tester struct {
	exported *packagestest.Exported
}

func export(t *testing.T, modules ...packagestest.Module) *packagestest.Exported {
	e := packagestest.Export(t, x, modules)
	t.Cleanup(e.Cleanup)
	exportedInfo(t, e)
	return e
}

func exportedInfo(t *testing.T, e *packagestest.Exported) {
	t.Logf("Work dir: %s", e.Config.Dir)
	t.Logf("Temp dir: %s", e.Temp())
	t.Logf("File located at: %s", e.File("pkg", "file.go"))
	t.Logf("Tree of temp dir:\n%s", dirTree(t, e.Temp()))
}

func dirTree(t *testing.T, dir string) string {
	tree := exec.Command("tree", dir)
	treeOut, err := tree.Output()
	require.NoError(t, err)
	return string(treeOut)
}

type testWriter struct{ t *testing.T }

func (w testWriter) Write(p []byte) (int, error) {
	w.t.Helper()
	w.t.Logf("%s", p)
	return len(p), nil
}

func formatGoFile(data []byte) ([]byte, error) {
	return format.Source(data)
	//fset := token.NewFileSet()
	//ast, err := parser.ParseFile(fset, "", data, parser.ParseComments)
	//if err != nil {
	//	return nil, fmt.Errorf("parse: %w", err)
	//}
	//conf := printer.Config{Mode: printer.TabIndent | printer.UseSpaces, Tabwidth: 4}
	//buf := &bytes.Buffer{}
	//err = conf.Fprint(buf, fset, ast)
	//if err != nil {
	//	return nil, fmt.Errorf("format: %w", err)
	//}
	//return buf.Bytes(), nil
}
