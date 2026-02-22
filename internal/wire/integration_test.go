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
	"context"
	"go/build"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// integratedTests lists the testdata cases that must remain as integration tests.
// These verify end-to-end code generation correctness that cannot be covered
// by unit tests alone.
//
// Success cases: verify generated code patterns.
// Error cases: verify errors that require full Generate pipeline.
var integratedTests = []string{
	// --- Success cases: code generation patterns ---
	"Chain",
	"Cleanup",
	"PartialCleanup",
	"ReturnError",
	"InjectInput",
	"InjectWithPanic",
	"InterfaceBinding",
	"ImportedInterfaceBinding",
	"InterfaceBindingReuse",
	"Struct",
	"StructPointer",
	"FieldsOfStruct",
	"FieldsOfStructPointer",
	"FieldsOfImportedStruct",
	"ValueChain",
	"ValueIsStruct",
	"InterfaceValue",
	"NamingWorstCase",
	"PkgImport",
	"Varargs",
	"ExportedValue",
	"DocComment",
	"Header",
	"CopyOtherDecls",
	"BuildTagsAllPackages",
	"ReservedKeywords",
	"NoopBuild",
	"ExampleWithMocks",
	"NiladicIdentity",
	"TwoDeps",
	"MultipleSimilarPackages",
	"NoInjectParamNames",
	"VarValue",
	"ReturnArgumentAsInterface",

	// --- Error cases: require full pipeline ---
	"InjectorCleanupMismatch",
	"InjectorErrorMismatch",
	"NoImplicitInterface",
	"FuncArgProvider",
}

// TestGenerate runs integration tests against the full Generate pipeline.
// Each test case materializes a temporary GOPATH, runs Generate, and
// compares the output against golden files.
func TestGenerate(t *testing.T) {
	const testRoot = "testdata"

	wireGo, err := os.ReadFile(filepath.Join("..", "..", "wire.go"))
	if err != nil {
		t.Fatal(err)
	}

	// Build a set for O(1) lookup.
	testSet := make(map[string]bool, len(integratedTests))
	for _, name := range integratedTests {
		testSet[name] = true
	}

	// Verify all listed tests actually exist.
	for _, name := range integratedTests {
		if _, err := os.Stat(filepath.Join(testRoot, name)); os.IsNotExist(err) {
			t.Errorf("listed integration test %q does not exist in testdata", name)
		}
	}

	var goToolPath string
	if *record {
		goToolPath = filepath.Join(build.Default.GOROOT, "bin", "go")
		if _, err := os.Stat(goToolPath); err != nil {
			t.Fatal("go toolchain not available:", err)
		}
	}

	ctx := context.Background()

	for _, name := range integratedTests {
		test, err := loadTestCase(filepath.Join(testRoot, name), wireGo)
		if err != nil {
			t.Errorf("load %s: %v", name, err)
			continue
		}

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			gopath, err := os.MkdirTemp("", "wire_test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(gopath)
			gopath, err = filepath.EvalSymlinks(gopath)
			if err != nil {
				t.Fatal(err)
			}
			if err := test.materialize(gopath); err != nil {
				t.Fatal(err)
			}
			wd := filepath.Join(gopath, "src", "example.com")
			gens, errs := Generate(ctx, wd, append(os.Environ(), "GOPATH="+gopath), []string{test.pkg}, &GenerateOptions{Header: test.header})
			var gen GenerateResult
			if len(gens) > 1 {
				t.Fatalf("got %d generated files, want 0 or 1", len(gens))
			}
			if len(gens) == 1 {
				gen = gens[0]
				if len(gen.Errs) > 0 {
					errs = append(errs, gen.Errs...)
				}
				if len(gen.Content) > 0 {
					defer t.Logf("wire_gen.go:\n%s", gen.Content)
				}
			}
			if len(errs) > 0 {
				gotErrStrings := make([]string, len(errs))
				for i, e := range errs {
					t.Log(e.Error())
					gotErrStrings[i] = scrubError(gopath, e.Error())
				}
				if !test.wantWireError {
					t.Fatal("Did not expect errors. To -record an error, run tests with -record flag.")
				}
				if *record {
					goVersion := getGoVersion()
					if goVersion == "" {
						t.Fatal("could not determine Go version for generating error file")
					}
					wireErrsFile := filepath.Join(testRoot, test.name, "want", "wire_errs_go"+goVersion+".txt")
					if err := os.WriteFile(wireErrsFile, []byte(strings.Join(gotErrStrings, "\n\n")), 0666); err != nil {
						t.Fatalf("failed to write version-specific wire_errs file: %v", err)
					}
				} else {
					if diff := cmp.Diff(gotErrStrings, test.wantWireErrorStrings); diff != "" {
						t.Errorf("Errors didn't match expected errors from wire_errors.txt:\n%s", diff)
					}
				}
				return
			}
			if test.wantWireError {
				t.Fatal("wire succeeded; want error")
			}
			outPathSane := true
			if prefix := gopath + string(os.PathSeparator) + "src" + string(os.PathSeparator); !strings.HasPrefix(gen.OutputPath, prefix) {
				outPathSane = false
				t.Errorf("suggested output path = %q; want to start with %q", gen.OutputPath, prefix)
			}

			if *record {
				if !outPathSane {
					return
				}
				if err := gen.Commit(); err != nil {
					t.Fatalf("failed to write wire_gen.go to test GOPATH: %v", err)
				}
				if err := goBuildCheck(goToolPath, gopath, test); err != nil {
					t.Fatalf("go build check failed: %v", err)
				}
				testdataWireGenPath := filepath.Join(testRoot, test.name, "want", "wire_gen.go")
				if err := os.WriteFile(testdataWireGenPath, gen.Content, 0666); err != nil {
					t.Fatalf("failed to record wire_gen.go to testdata: %v", err)
				}
			} else {
				if !bytes.Equal(gen.Content, test.wantWireOutput) {
					gotS, wantS := string(gen.Content), string(test.wantWireOutput)
					diff := cmp.Diff(strings.Split(gotS, "\n"), strings.Split(wantS, "\n"))
					t.Fatalf("wire output differs from golden file. If this change is expected, run with -record to update the wire_gen.go file.\n*** got:\n%s\n\n*** want:\n%s\n\n*** diff:\n%s", gotS, wantS, diff)
				}
			}
		})
	}
}
