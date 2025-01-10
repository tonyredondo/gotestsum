package util

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os/exec"
	"path/filepath"
	"strings"
)

func GetModuleInfo(dir string) (modulePath string, moduleDir string, rootDir string) {
	// Let's try to get the module path and dir from the go list command and root dir from git
	goListCmd := exec.Command("go", "list", "-f", "{{.Module.Path}};{{.Module.Dir}}")
	goListCmd.Dir = dir
	goListStdOut, goListErr := goListCmd.CombinedOutput()
	if goListErr == nil {
		parts := strings.Split(string(goListStdOut), ";")
		if len(parts) == 2 {
			modulePath = strings.TrimSpace(parts[0])
			moduleDir = strings.TrimSpace(parts[1])

			// If we have a module Dir, let's try then calculate the git path
			if moduleDir != "" {
				gitCmd := exec.Command("git", "rev-parse", "--show-toplevel")
				gitCmd.Dir = moduleDir
				gitAbsoluteDir, gitErr := gitCmd.CombinedOutput()
				if gitErr == nil {
					rootDir = strings.TrimSpace(strings.Trim(string(gitAbsoluteDir), "\n"))
				}
			}
		}
	}
	return modulePath, moduleDir, rootDir
}

// GetTestFilePath uses the Go parser to accurately detect any function named testName.
// If found, returns the file path; otherwise returns packagePath. Any errors also result in packagePath.
func GetTestFilePath(packagePath, testName string) string {
	var foundFilePath string

	if strings.Contains(testName, "/") {
		testName = testName[:strings.Index(testName, "/")]
	}

	err := filepath.WalkDir(packagePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// If we can't read a subdirectory, skip it
			return nil
		}

		// Skip non-Go files
		if d.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}

		// Parse the file into an AST
		fset := token.NewFileSet()
		node, parseErr := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if parseErr != nil {
			// If parse fails, skip it
			return nil
		}

		// Inspect every declaration in the file
		for _, decl := range node.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				// Compare the function name with testName
				if fn.Name.Name == testName {
					foundFilePath = path
					// Using filepath.SkipDir to stop the entire walk
					return filepath.SkipDir
				}
			}
		}

		return nil
	})

	// If there's a real error (besides SkipDir), fallback to packagePath
	if err != nil && !errors.Is(err, filepath.SkipDir) {
		return packagePath
	}
	if foundFilePath == "" {
		return packagePath
	}
	return foundFilePath
}

func GetRelativePathFrom(basePath, path string) string {
	if basePath == "" {
		return path
	}
	relPath, err := filepath.Rel(basePath, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(relPath)
}
