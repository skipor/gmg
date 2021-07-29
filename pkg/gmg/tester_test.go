package gmg

import (
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
	e := export(t, modules...)
	return &Tester{
		exported: e,
	}
}

func (tr *Tester) run(t *testing.T, args ...string) *RunResult {
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

func (tr *Tester) Succeed(t *testing.T, args ...string) *RunResult {
	res := tr.run(t, args...)
	require.Zero(t, res.ExitCode)
	return res
}

func (tr *Tester) Fail(t *testing.T, args ...string) *RunResult {
	res := tr.run(t, args...)
	require.NotZero(t, res.ExitCode)
	return res
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

type RunResult struct {
	t        *testing.T
	ExitCode int
	FS       *afero.MemMapFs
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
	cmd := exec.Command("tree", e.Temp())
	out, err := cmd.Output()
	require.NoError(t, err)
	t.Logf("Tree of temp dir:\n%s", out)
	//logGoEnv(t, e.Config.Env)
}

type testWriter struct{ t *testing.T }

func (w testWriter) Write(p []byte) (int, error) {
	w.t.Helper()
	w.t.Logf("%s", p)
	return len(p), nil
}
