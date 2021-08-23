package examples

import (
	"flag"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/skipor/gmg/internal/testutil"
)

var update bool

func init() {
	flag.BoolVar(&update, "update", false, "Update mocks in examples, instead of check that they are not changed.")
}

func TestExamplesGoGenerate(t *testing.T) {
	PATH := testutil.TestInstallGmgOnce(t)
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
			cmd := exec.Command("go", "generate", "./"+dir+"/...")
			cmd.Env = append(os.Environ(), "GMG_DEBUG=true", PATH)
			out, err := cmd.CombinedOutput()
			t.Logf("%s\n%s", cmd.String(), out)
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
