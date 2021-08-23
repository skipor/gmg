package testutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
)

var installErr error
var installOnce sync.Once

// InstallGmgOnce installs gmg from source to temporary dir and adds it to PATH for current process.
func InstallGmgOnce() (string, error) {
	installOnce.Do(func() {
		installErr = goInstallGmg()
	})
	if installErr != nil {
		return "", installErr
	}
	return "PATH=" + tmpGobin + ":" + os.Getenv("PATH"), nil
}

func TestInstallGmgOnce(t *testing.T) string {
	t.Helper()
	path, err := InstallGmgOnce()
	if err != nil {
		t.Fatalf("gmg install failed:\n%s", err)
	}
	return path
}

var tmpGobin string

func goInstallGmg() error {
	tmpGobin = filepath.Join(os.TempDir(), "gmg_test")
	// Use stable tmp dir, to cache binary between different test packages and runs.
	err := os.MkdirAll(tmpGobin, 0755)
	if err != nil {
		return err
	}
	executable := filepath.Join(tmpGobin, "gmg")
	if runtime.GOOS == "windows" {
		executable += ".exe"
	}

	cmd := exec.Command("go", "install", "github.com/skipor/gmg")
	cmd.Env = append(os.Environ(), "GOBIN="+tmpGobin)
	fmt.Printf("Installing gmg: GOBIN=%s %s\n", tmpGobin, cmd.String())
	start := time.Now()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s:\n%s\n%w", cmd.String(), out, err)
	}
	fmt.Printf("gmg install succeed in: %s\n", time.Since(start).Truncate(time.Millisecond))
	return nil
}
