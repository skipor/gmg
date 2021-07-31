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
func InstallGmgOnce() error {
	installOnce.Do(func() {
		installErr = goInstallGmg()
	})
	return installErr
}

func TestInstallGmgOnce(t *testing.T) {
	t.Helper()
	err := InstallGmgOnce()
	if err != nil {
		t.Fatalf("gmg install failed:\n%s", err)
	}
}

func goInstallGmg() error {
	// Use stable tmp dir, to cache binary between different test packages and runs.
	tmpGobin := filepath.Join(os.TempDir(), "gmg_test")
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
	err = os.Setenv("PATH", tmpGobin+":"+os.Getenv("PATH"))
	if err != nil {
		return fmt.Errorf("PATH set: %s", err)
	}
	return nil
}
