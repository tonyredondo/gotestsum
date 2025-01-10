package junitxml

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"gotest.tools/gotestsum/internal/util"
	"gotest.tools/gotestsum/testjson"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/golden"
)

func TestWrite(t *testing.T) {
	out := new(bytes.Buffer)
	exec := createExecution(t)

	t.Setenv("GOVERSION", "go7.7.7")
	err := Write(out, exec, Config{
		ProjectName:     "test",
		customTimestamp: new(time.Time).Format(time.RFC3339),
		customElapsed:   "2.1",
	})
	assert.NilError(t, err)
	golden.Assert(t, out.String(), "junitxml-report.golden")
}

func TestWrite_HideEmptyPackages(t *testing.T) {
	out := new(bytes.Buffer)
	exec := createExecution(t)

	t.Setenv("GOVERSION", "go7.7.7")
	err := Write(out, exec, Config{
		ProjectName:       "test",
		HideEmptyPackages: true,
		customTimestamp:   new(time.Time).Format(time.RFC3339),
		customElapsed:     "2.1",
	})
	assert.NilError(t, err)
	golden.Assert(t, out.String(), "junitxml-report-skip-empty.golden")
}

func createExecution(t *testing.T) *testjson.Execution {
	modulePath, moduleDir, rootDir := util.GetModuleInfo("")
	exec, err := testjson.ScanTestOutput(testjson.ScanConfig{
		Stdout:     readTestData(t, "out"),
		Stderr:     readTestData(t, "err"),
		ModulePath: modulePath,
		ModuleDir:  moduleDir,
		RootDir:    rootDir,
	})
	assert.NilError(t, err)
	return exec
}

func readTestData(t *testing.T, stream string) io.Reader {
	raw, err := os.ReadFile("../../testjson/testdata/input/go-test-json." + stream)
	assert.NilError(t, err)
	return bytes.NewReader(raw)
}

func TestGoVersion(t *testing.T) {
	t.Run("unknown", func(t *testing.T) {
		t.Setenv("PATH", "/bogus")
		assert.Equal(t, goVersion(), "unknown")
	})

	t.Run("current version", func(t *testing.T) {
		expected := `go1.23.4 darwin/arm64`
		assert.Equal(t, goVersion(), expected)
	})
}
