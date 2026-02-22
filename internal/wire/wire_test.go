// Copyright 2018 The Wire Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package wire

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/google/go-cmp/cmp"
)

var record = flag.Bool("record", false, "whether to run tests against cloud resources and record the interactions")

// getGoVersion returns the major.minor version of the current Go runtime (e.g., "1.22")
func getGoVersion() string {
	version := runtime.Version()
	// version format: "go1.22.4" or "go1.23.0"
	if !strings.HasPrefix(version, "go") {
		return ""
	}
	version = strings.TrimPrefix(version, "go")
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return ""
	}
	return parts[0] + "." + parts[1]
}

// getVersionedErrorFile returns the path to the version-specific wire_errs.txt file.
// It first looks for a file matching the current Go version (e.g., wire_errs_go1.26.txt).
// If not found, it falls back to the latest available version-specific error file.
// Returns an error if no version-specific files exist at all.
func getVersionedErrorFile(root string) (string, error) {
	goVersion := getGoVersion()
	if goVersion == "" {
		return "", fmt.Errorf("could not determine Go version")
	}
	// Try exact match first.
	versionedPath := filepath.Join(root, "want", "wire_errs_go"+goVersion+".txt")
	if _, err := os.Stat(versionedPath); err == nil {
		return versionedPath, nil
	}
	// Fallback: find the latest available version-specific error file.
	entries, err := os.ReadDir(filepath.Join(root, "want"))
	if err != nil {
		return "", fmt.Errorf("could not read want directory: %v", err)
	}
	var candidates []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "wire_errs_go") && strings.HasSuffix(e.Name(), ".txt") {
			candidates = append(candidates, e.Name())
		}
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("no version-specific error files found in %s", filepath.Join(root, "want"))
	}
	sort.Strings(candidates)
	return filepath.Join(root, "want", candidates[len(candidates)-1]), nil
}

func goBuildCheck(goToolPath, gopath string, test *testCase) error {
	// Run `go build`.
	testExePath := filepath.Join(gopath, "bin", "testprog")
	buildCmd := []string{"build", "-o", testExePath}
	buildCmd = append(buildCmd, test.pkg)
	cmd := exec.Command(goToolPath, buildCmd...)
	cmd.Dir = filepath.Join(gopath, "src", "example.com")
	cmd.Env = append(os.Environ(), "GOPATH="+gopath)
	if buildOut, err := cmd.CombinedOutput(); err != nil {
		if len(buildOut) > 0 {
			return fmt.Errorf("build: %v; output:\n%s", err, buildOut)
		}
		return fmt.Errorf("build: %v", err)
	}

	// Run the resulting program and compare its output to the expected
	// output.
	out, err := exec.Command(testExePath).Output()
	if err != nil {
		return fmt.Errorf("run compiled program: %v", err)
	}
	if !bytes.Equal(out, test.wantProgramOutput) {
		gotS, wantS := string(out), string(test.wantProgramOutput)
		diff := cmp.Diff(strings.Split(gotS, "\n"), strings.Split(wantS, "\n"))
		return fmt.Errorf("compiled program output doesn't match:\n*** got:\n%s\n\n*** want:\n%s\n\n*** diff:\n%s", gotS, wantS, diff)
	}
	return nil
}

// scrubError rewrites the given string to remove occurrences of GOPATH/src,
// rewrites OS-specific path separators to slashes, and any line/column
// information to a fixed ":x:y". For example, if the gopath parameter is
// "C:\GOPATH" and running on Windows, the string
// "C:\GOPATH\src\foo\bar.go:15:4" would be rewritten to "foo/bar.go:x:y".
func scrubError(gopath string, s string) string {
	sb := new(strings.Builder)
	query := gopath + string(os.PathSeparator) + "src" + string(os.PathSeparator)
	for {
		// Find next occurrence of source root. This indicates the next path to
		// scrub.
		start := strings.Index(s, query)
		if start == -1 {
			sb.WriteString(s)
			break
		}

		// Find end of file name (extension ".go").
		fileStart := start + len(query)
		fileEnd := strings.Index(s[fileStart:], ".go")
		if fileEnd == -1 {
			// If no ".go" occurs to end of string, further searches will fail too.
			// Break the loop.
			sb.WriteString(s)
			break
		}
		fileEnd += fileStart + 3 // Advance to end of extension.

		// Write out file name and advance scrub position.
		file := s[fileStart:fileEnd]
		if os.PathSeparator != '/' {
			file = strings.Replace(file, string(os.PathSeparator), "/", -1)
		}
		sb.WriteString(s[:start])
		sb.WriteString(file)
		s = s[fileEnd:]

		// Peek past to see if there is line/column info.
		linecol, linecolLen := scrubLineColumn(s)
		sb.WriteString(linecol)
		s = s[linecolLen:]
	}
	return sb.String()
}

func scrubLineColumn(s string) (replacement string, n int) {
	if !strings.HasPrefix(s, ":") {
		return "", 0
	}
	// Skip first colon and run of digits.
	for n++; len(s) > n && '0' <= s[n] && s[n] <= '9'; {
		n++
	}
	if n == 1 {
		// No digits followed colon.
		return "", 0
	}

	// Start on column part.
	if !strings.HasPrefix(s[n:], ":") {
		return ":x", n
	}
	lineEnd := n
	// Skip second colon and run of digits.
	for n++; len(s) > n && '0' <= s[n] && s[n] <= '9'; {
		n++
	}
	if n == lineEnd+1 {
		// No digits followed second colon.
		return ":x", lineEnd
	}
	return ":x:y", n
}

type testCase struct {
	name                 string
	pkg                  string
	header               []byte
	goFiles              map[string][]byte
	wantProgramOutput    []byte
	wantWireOutput       []byte
	wantWireError        bool
	wantWireErrorStrings []string
}

// loadTestCase reads a test case from a directory.
//
// The directory structure is:
//
//	root/
//
//		pkg
//			file containing the package name containing the inject function
//			(must also be package main)
//
//		...
//			any Go files found recursively placed under GOPATH/src/...
//
//		want/
//
//			wire_errs_go1.XX.txt
//					Expected errors from the Wire Generate function for Go version 1.XX,
//					missing if no errors expected.
//					Distinct errors are separated by a blank line,
//					and line numbers and line positions are scrubbed
//					(e.g. "$GOPATH/src/foo.go:52:8" --> "foo.go:x:y").
//					Version-specific error files are required (e.g., wire_errs_go1.22.txt).
//
//			wire_gen.go
//					verified output of wire from a test run with
//					-record, missing if wire_errs_go1.XX.txt is present
//
//			program_out.txt
//					expected output from the final compiled program,
//					missing if wire_errs_go1.XX.txt is present
func loadTestCase(root string, wireGoSrc []byte) (*testCase, error) {
	name := filepath.Base(root)
	pkg, err := os.ReadFile(filepath.Join(root, "pkg"))
	if err != nil {
		return nil, fmt.Errorf("load test case %s: %v", name, err)
	}
	header, _ := os.ReadFile(filepath.Join(root, "header"))
	var wantProgramOutput []byte
	var wantWireOutput []byte
	// Try to load version-specific error file
	wireErrsPath, err := getVersionedErrorFile(root)
	var wireErrb []byte
	var wantWireError bool
	if err == nil {
		wireErrb, err = os.ReadFile(wireErrsPath)
		wantWireError = err == nil
	}
	var wantWireErrorStrings []string
	if wantWireError {
		for _, errs := range strings.Split(string(wireErrb), "\n\n") {
			// Allow for trailing newlines, which can be hard to remove in some editors.
			wantWireErrorStrings = append(wantWireErrorStrings, strings.TrimRight(errs, "\n\r"))
		}
	} else {
		if !*record {
			wantWireOutput, err = os.ReadFile(filepath.Join(root, "want", "wire_gen.go"))
			if err != nil {
				return nil, fmt.Errorf("load test case %s: %v, if this is a new testcase, run with -record to generate the wire_gen.go file", name, err)
			}
		}
		wantProgramOutput, err = os.ReadFile(filepath.Join(root, "want", "program_out.txt"))
		if err != nil {
			return nil, fmt.Errorf("load test case %s: %v", name, err)
		}
	}
	goFiles := map[string][]byte{
		"github.com/almondoo/wire/wire.go": wireGoSrc,
	}
	err = filepath.Walk(root, func(src string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, src)
		if err != nil {
			return err // unlikely
		}
		if info.Mode().IsDir() && rel == "want" {
			// The "want" directory should not be included in goFiles.
			return filepath.SkipDir
		}
		if !info.Mode().IsRegular() || filepath.Ext(src) != ".go" {
			return nil
		}
		data, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		goFiles["example.com/"+filepath.ToSlash(rel)] = data
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("load test case %s: %v", name, err)
	}
	return &testCase{
		name:                 name,
		pkg:                  string(bytes.TrimSpace(pkg)),
		header:               header,
		goFiles:              goFiles,
		wantWireOutput:       wantWireOutput,
		wantProgramOutput:    wantProgramOutput,
		wantWireError:        wantWireError,
		wantWireErrorStrings: wantWireErrorStrings,
	}, nil
}

// materialize creates a new GOPATH at the given directory, which may or
// may not exist.
func (test *testCase) materialize(gopath string) error {
	for name, content := range test.goFiles {
		dst := filepath.Join(gopath, "src", filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(dst), 0777); err != nil {
			return fmt.Errorf("materialize GOPATH: %v", err)
		}
		if err := os.WriteFile(dst, content, 0666); err != nil {
			return fmt.Errorf("materialize GOPATH: %v", err)
		}
	}

	// Add go.mod files to example.com and github.com/almondoo/wire.
	const importPath = "example.com"
	const depPath = "github.com/almondoo/wire"
	depLoc := filepath.Join(gopath, "src", filepath.FromSlash(depPath))
	example := fmt.Sprintf("module %s\n\nrequire %s v0.1.0\nreplace %s => %s\n", importPath, depPath, depPath, depLoc)
	gomod := filepath.Join(gopath, "src", filepath.FromSlash(importPath), "go.mod")
	if err := os.WriteFile(gomod, []byte(example), 0666); err != nil {
		return fmt.Errorf("generate go.mod for %s: %v", gomod, err)
	}
	if err := os.WriteFile(filepath.Join(depLoc, "go.mod"), []byte("module "+depPath+"\n"), 0666); err != nil {
		return fmt.Errorf("generate go.mod for %s: %v", depPath, err)
	}
	return nil
}
