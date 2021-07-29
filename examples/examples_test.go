package examples

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var update bool

func init() {
	flag.BoolVar(&update, "update", false, "Update mocks in examples, instead of check that they are not changed.")
}

func TestExamplesGoGenerate(t *testing.T) {
	gmg, err := build()
	if err != nil {
		require.NoError(t, err, "gmg build")
	}
	defer gmg.cleanup()
	err = os.Setenv("PATH", filepath.Dir(gmg.executable)+":"+os.Getenv("PATH"))
	if err != nil {
		require.NoError(t, err, "PATH set")
	}

	entries, err := os.ReadDir(".")
	require.NoError(t, err, "test dir read")
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		dir := ent.Name()
		t.Run(dir, func(t *testing.T) {
			if !update {
				diff := vcsDiff(t, dir)
				if diff != "" {
					t.Fatalf("diff before generate:\n%s", diff)
				}
			}
			cmd := exec.Command("go", "generate")
			out, err := cmd.CombinedOutput()
			t.Logf("%s:\n%s", cmd.String(), out)
			require.NoError(t, err, "go generate failed")
			if !update {
				diff := vcsDiff(t, dir)
				if diff != "" {
					t.Fatalf("diff after generate:\n%s", diff)
				}
			} else {
				vcsAddAll(t, dir)
			}
		})
	}

}

type buildResult struct {
	executable string
	cleanup    func()
}

func build() (res *buildResult, err error) {
	tmp, err := ioutil.TempDir("", "gmg_test_main_*")
	if err != nil {
		return nil, fmt.Errorf("tmp dir: %+v", err)
	}
	executable := filepath.Join(tmp, "gmg")
	if runtime.GOOS == "windows" {
		executable += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", executable, "./..")
	fmt.Printf("Building gmg: %s\n", cmd.String())
	start := time.Now()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s\n%w", out, err)
	}
	fmt.Printf("gmg build succeed in: %s\n", time.Since(start).Truncate(time.Millisecond))
	return &buildResult{
		executable: executable,
		cleanup:    func() { _ = os.RemoveAll(tmp) },
	}, nil
}

func vcsDiff(t *testing.T, dir string) string {
	gitAddNew(t, dir)
	cmd := exec.Command("git", "diff", "--exit-code", dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		_, ok := err.(*exec.ExitError)
		if ok {
			return string(out)
		}
		t.Fatalf("unexpected '%s' fail: %+v", cmd.String(), err)
	}
	return ""
}

// gitAddNew adds untracked files to make them visible in diff.
func gitAddNew(t *testing.T, dir string) {
	cmd := exec.Command("git", "add", "-N", dir)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s\n%s", out)
}

func vcsAddAll(t *testing.T, dir string) {
	cmd := exec.Command("git", "add", "-A", dir)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s\n%s", out)
}

func logf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
}
