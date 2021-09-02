package test

import (
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

var update bool

func init() {
	flag.BoolVar(&update, "update", false, "Update golden files in testdata.")
}

func TestMain(m *testing.M) {
	os.Exit(testMainRun(m))
}

func testMainRun(m *testing.M) int {
	code := m.Run()
	if code != 0 {
		// Don't check on run fail.
		return code
	}
	golden.MustPostRun()
	return 0
}

var golden = NewGoldenTestData("testdata")

func NewGoldenTestData(dir string) *GoldenTestData {
	return &GoldenTestData{
		path:        dir,
		checkedDirs: map[string]struct{}{},
	}
}

type GoldenTestData struct {
	path        string
	checkedDirs map[string]struct{}
}

func (td *GoldenTestData) Dir(t *testing.T, actualFS afero.Fs) bool {
	dirName := t.Name()
	dirName = strings.ReplaceAll(dirName, "/", "-")
	td.checkedDirs[dirName] = struct{}{}
	require.True(t, fs.ValidPath(dirName), "generated dir path is invalid")
	dir := filepath.Join(td.path, dirName)

	actual := fsToMap(t, actualFS)
	if update {
		td.update(t, dir, actual)
		return true
	}
	dirExists, err := afero.Exists(afero.NewOsFs(), dir)
	require.NoError(t, err)
	if !dirExists {
		t.Fatalf("Golden dir '%s' is not exist. Seems test is new or it's name changed. Run tests with '--update' to generate it.", dir)
	}

	expected := fsToMap(t, afero.NewBasePathFs(afero.NewOsFs(), dir))

	if !cmp.Equal(expected, actual) {
		t.Errorf("Diff in golden files:\n%s",
			cmp.Diff(expected, actual),
		)
		return false
	}
	return true
}

func (td *GoldenTestData) update(t *testing.T, dir string, actual pathToFileContentMap) {
	t.Logf("Updating: %s", dir)
	_ = os.RemoveAll(dir)
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err)

	for file, data := range actual {
		file = filepath.Join(dir, file)

		fileDir := filepath.Dir(file)
		err := os.MkdirAll(fileDir, 0755)
		require.NoError(t, err, "File '%s' mkdir failed", fileDir)

		err = ioutil.WriteFile(file, []byte(data), 0644)
		require.NoError(t, err, "File '%s' write failed", file)
	}
}
func (td *GoldenTestData) MustPostRun() {
	err := golden.PostRun()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Golden post check failed: %+v", err)
		os.Exit(1)
	}
}

// PostRun searches for unchecked
func (td *GoldenTestData) PostRun() error {
	if isPartialRun() {
		// Some dirs were not checked, so can't check for extra.
		return nil
	}
	infos, err := os.ReadDir(td.path)
	if err != nil {
		return fmt.Errorf("dir '%s' read failed: %w", td.path, err)
	}

	var extra []string
	for _, info := range infos {
		if !info.IsDir() {
			continue
		}
		if _, ok := td.checkedDirs[info.Name()]; ok {
			continue
		}
		path := filepath.Join(td.path, info.Name())
		if update {
			err := os.RemoveAll(path)
			if err != nil {
				return fmt.Errorf("failed to remove all in extra dir '%s'", path)
			}
			fmt.Printf("Extra golden dir '%s'removed\n", path)
			continue
		} else {
			extra = append(extra, info.Name())
		}
	}
	if update {
		return nil
	}
	if len(extra) > 0 {
		return fmt.Errorf("%v extra dirs found in '%s': %s",
			len(extra), td.path, strings.Join(extra, ", "),
		)
	}
	fmt.Println("Golden PostRun succeed")
	return nil
}

type pathToFileContentMap map[string]string

func fsToMap(t *testing.T, FS afero.Fs) pathToFileContentMap {
	m := pathToFileContentMap{}
	err := afero.Walk(FS, "", func(path string, info fs.FileInfo, err error) error {
		require.NoError(t, err, "'%s' walk failed", path)
		if info.IsDir() {
			return nil
		}
		data, err := afero.ReadFile(FS, path)
		require.NoError(t, err, "File '%s' read failed", path)
		m[path] = string(data)
		return nil
	})
	require.NoError(t, err)
	return m
}

func isPartialRun() bool {
	for _, arg := range os.Args {
		switch arg {
		case "-test.run", "--test.run":
			return true
		}
	}
	return false
}
