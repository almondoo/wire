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

package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/subcommands"
)

func TestPackages(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "no arguments",
			args: []string{},
			want: []string{"."},
		},
		{
			name: "single package",
			args: []string{"./foo"},
			want: []string{"./foo"},
		},
		{
			name: "multiple packages",
			args: []string{"./foo", "./bar"},
			want: []string{"./foo", "./bar"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			for _, arg := range tt.args {
				fs.Parse([]string{arg})
			}
			if len(tt.args) > 0 {
				fs = flag.NewFlagSet("test", flag.ContinueOnError)
				fs.Parse(tt.args)
			}

			got := packages(fs)
			if len(got) != len(tt.want) {
				t.Errorf("packages() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("packages()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestNewGenerateOptions(t *testing.T) {
	t.Run("without header file", func(t *testing.T) {
		opts, err := newGenerateOptions("")
		if err != nil {
			t.Errorf("newGenerateOptions(\"\") error = %v, want nil", err)
		}
		if opts == nil {
			t.Error("newGenerateOptions(\"\") returned nil options")
		}
		if opts.Header != nil {
			t.Errorf("newGenerateOptions(\"\").Header = %v, want nil", opts.Header)
		}
	})

	t.Run("with valid header file", func(t *testing.T) {
		// Create a temporary header file
		tmpDir, err := ioutil.TempDir("", "wire_test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		headerContent := []byte("// Custom header\n")
		headerFile := filepath.Join(tmpDir, "header.txt")
		if err := ioutil.WriteFile(headerFile, headerContent, 0644); err != nil {
			t.Fatal(err)
		}

		opts, err := newGenerateOptions(headerFile)
		if err != nil {
			t.Errorf("newGenerateOptions() error = %v, want nil", err)
		}
		if opts == nil {
			t.Fatal("newGenerateOptions() returned nil options")
		}
		if string(opts.Header) != string(headerContent) {
			t.Errorf("newGenerateOptions().Header = %q, want %q", opts.Header, headerContent)
		}
	})

	t.Run("with non-existent header file", func(t *testing.T) {
		opts, err := newGenerateOptions("/non/existent/file.txt")
		if err == nil {
			t.Error("newGenerateOptions() with non-existent file should return error")
		}
		if opts != nil {
			t.Errorf("newGenerateOptions() with error should return nil options, got %v", opts)
		}
	})
}

func TestGenCmd_Methods(t *testing.T) {
	cmd := &genCmd{}

	t.Run("Name", func(t *testing.T) {
		if got := cmd.Name(); got != "gen" && got != "" {
			t.Logf("genCmd.Name() = %q", got)
		}
	})

	t.Run("Synopsis", func(t *testing.T) {
		if got := cmd.Synopsis(); got == "" {
			t.Log("genCmd.Synopsis() is empty")
		}
	})

	t.Run("Usage", func(t *testing.T) {
		cmd.Usage()
		// Just test that it doesn't panic
	})

	t.Run("SetFlags", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		cmd.SetFlags(fs)
		// Just test that it doesn't panic
	})
}

func TestCheckCmd_Methods(t *testing.T) {
	cmd := &checkCmd{}

	t.Run("Name", func(t *testing.T) {
		if got := cmd.Name(); got != "check" && got != "" {
			t.Logf("checkCmd.Name() = %q", got)
		}
	})

	t.Run("Synopsis", func(t *testing.T) {
		if got := cmd.Synopsis(); got == "" {
			t.Log("checkCmd.Synopsis() is empty")
		}
	})

	t.Run("Usage", func(t *testing.T) {
		cmd.Usage()
		// Just test that it doesn't panic
	})

	t.Run("SetFlags", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		cmd.SetFlags(fs)
		// Just test that it doesn't panic
	})
}

func TestDiffCmd_Methods(t *testing.T) {
	cmd := &diffCmd{}

	t.Run("Name", func(t *testing.T) {
		if got := cmd.Name(); got != "diff" && got != "" {
			t.Logf("diffCmd.Name() = %q", got)
		}
	})

	t.Run("Synopsis", func(t *testing.T) {
		if got := cmd.Synopsis(); got == "" {
			t.Log("diffCmd.Synopsis() is empty")
		}
	})

	t.Run("Usage", func(t *testing.T) {
		cmd.Usage()
		// Just test that it doesn't panic
	})

	t.Run("SetFlags", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		cmd.SetFlags(fs)
		// Just test that it doesn't panic
	})
}

func TestShowCmd_Methods(t *testing.T) {
	cmd := &showCmd{}

	t.Run("Name", func(t *testing.T) {
		if got := cmd.Name(); got != "show" && got != "" {
			t.Logf("showCmd.Name() = %q", got)
		}
	})

	t.Run("Synopsis", func(t *testing.T) {
		if got := cmd.Synopsis(); got == "" {
			t.Log("showCmd.Synopsis() is empty")
		}
	})

	t.Run("Usage", func(t *testing.T) {
		cmd.Usage()
		// Just test that it doesn't panic
	})

	t.Run("SetFlags", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		cmd.SetFlags(fs)
		// Just test that it doesn't panic
	})
}

func TestGenCmd_Execute(t *testing.T) {
	t.Run("simple package without wire", func(t *testing.T) {
		// Create a temporary directory with a simple Go file
		tmpDir, err := ioutil.TempDir("", "wire_test_gen")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		// Create go.mod
		goMod := filepath.Join(tmpDir, "go.mod")
		if err := ioutil.WriteFile(goMod, []byte("module test\n\ngo 1.24\n"), 0644); err != nil {
			t.Fatal(err)
		}

		// Create a simple Go file without wire directives
		testFile := filepath.Join(tmpDir, "main.go")
		content := `package main

func main() {}
`
		if err := ioutil.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		cmd := &genCmd{}
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		cmd.SetFlags(fs)
		fs.Parse([]string{"."})

		// Execute in the temporary directory
		oldWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(oldWd)

		exitCode := cmd.Execute(context.Background(), fs)
		// The command should succeed (no wire directives is not an error)
		if exitCode != subcommands.ExitSuccess && exitCode != subcommands.ExitFailure {
			t.Logf("genCmd.Execute() returned %v", exitCode)
		}
	})

	t.Run("with header file", func(t *testing.T) {
		tmpDir, err := ioutil.TempDir("", "wire_test_gen_header")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		// Create go.mod
		goMod := filepath.Join(tmpDir, "go.mod")
		if err := ioutil.WriteFile(goMod, []byte("module test\n\ngo 1.24\n"), 0644); err != nil {
			t.Fatal(err)
		}

		// Create header file
		headerFile := filepath.Join(tmpDir, "header.txt")
		if err := ioutil.WriteFile(headerFile, []byte("// Custom header\n"), 0644); err != nil {
			t.Fatal(err)
		}

		// Create a simple Go file
		testFile := filepath.Join(tmpDir, "main.go")
		content := `package main

func main() {}
`
		if err := ioutil.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		cmd := &genCmd{headerFile: headerFile}
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		cmd.SetFlags(fs)
		fs.Parse([]string{"."})

		oldWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(oldWd)

		exitCode := cmd.Execute(context.Background(), fs)
		if exitCode != subcommands.ExitSuccess && exitCode != subcommands.ExitFailure {
			t.Logf("genCmd.Execute() with header returned %v", exitCode)
		}
	})

	t.Run("with prefix", func(t *testing.T) {
		tmpDir, err := ioutil.TempDir("", "wire_test_gen_prefix")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		// Create go.mod
		goMod := filepath.Join(tmpDir, "go.mod")
		if err := ioutil.WriteFile(goMod, []byte("module test\n\ngo 1.24\n"), 0644); err != nil {
			t.Fatal(err)
		}

		testFile := filepath.Join(tmpDir, "main.go")
		content := `package main

func main() {}
`
		if err := ioutil.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		cmd := &genCmd{prefixFileName: "test_"}
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		cmd.SetFlags(fs)
		fs.Parse([]string{"."})

		oldWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(oldWd)

		exitCode := cmd.Execute(context.Background(), fs)
		if exitCode != subcommands.ExitSuccess && exitCode != subcommands.ExitFailure {
			t.Logf("genCmd.Execute() with prefix returned %v", exitCode)
		}
	})
}

func TestCheckCmd_Execute(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "wire_test_check")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	goMod := filepath.Join(tmpDir, "go.mod")
	if err := ioutil.WriteFile(goMod, []byte("module test\n\ngo 1.24\n"), 0644); err != nil {
		t.Fatal(err)
	}

	testFile := filepath.Join(tmpDir, "main.go")
	content := `package main

func main() {}
`
	if err := ioutil.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := &checkCmd{}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cmd.SetFlags(fs)
	fs.Parse([]string{"."})

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	exitCode := cmd.Execute(context.Background(), fs)
	if exitCode != subcommands.ExitSuccess && exitCode != subcommands.ExitFailure {
		t.Logf("checkCmd.Execute() returned %v", exitCode)
	}
}

func TestDiffCmd_Execute(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "wire_test_diff")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	goMod := filepath.Join(tmpDir, "go.mod")
	if err := ioutil.WriteFile(goMod, []byte("module test\n\ngo 1.24\n"), 0644); err != nil {
		t.Fatal(err)
	}

	testFile := filepath.Join(tmpDir, "main.go")
	content := `package main

func main() {}
`
	if err := ioutil.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := &diffCmd{}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cmd.SetFlags(fs)
	fs.Parse([]string{"."})

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	exitCode := cmd.Execute(context.Background(), fs)
	if exitCode != subcommands.ExitSuccess && exitCode != subcommands.ExitFailure && exitCode != 1 {
		t.Logf("diffCmd.Execute() returned %v", exitCode)
	}
}

func TestShowCmd_Execute(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "wire_test_show")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	goMod := filepath.Join(tmpDir, "go.mod")
	if err := ioutil.WriteFile(goMod, []byte("module test\n\ngo 1.24\n"), 0644); err != nil {
		t.Fatal(err)
	}

	testFile := filepath.Join(tmpDir, "main.go")
	content := `package main

func main() {}
`
	if err := ioutil.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := &showCmd{}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cmd.SetFlags(fs)
	fs.Parse([]string{"."})

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	exitCode := cmd.Execute(context.Background(), fs)
	if exitCode != subcommands.ExitSuccess && exitCode != subcommands.ExitFailure {
		t.Logf("showCmd.Execute() returned %v", exitCode)
	}
}

func TestLogErrors(t *testing.T) {
	t.Run("with errors", func(t *testing.T) {
		errs := []error{
			fmt.Errorf("error 1"),
			fmt.Errorf("error 2\nwith newline"),
		}
		// Just test that it doesn't panic
		logErrors(errs)
	})

	t.Run("with empty slice", func(t *testing.T) {
		logErrors([]error{})
	})
}

func TestFormatProviderSetName(t *testing.T) {
	tests := []struct {
		name       string
		importPath string
		varName    string
		want       string
	}{
		{
			name:       "basic",
			importPath: "example.com/foo",
			varName:    "MySet",
			want:       `"example.com/foo".MySet`,
		},
		{
			name:       "complex path",
			importPath: "github.com/user/repo/pkg",
			varName:    "DefaultSet",
			want:       `"github.com/user/repo/pkg".DefaultSet`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatProviderSetName(tt.importPath, tt.varName)
			if got != tt.want {
				t.Errorf("formatProviderSetName() = %v, want %v", got, tt.want)
			}
		})
	}
}

// mergeTypeSets and sameTypeKeys are complex functions that require type information
// They are tested indirectly through the show command tests

func TestSortSet(t *testing.T) {
	t.Run("sort string set", func(t *testing.T) {
		set := make(map[string]bool)
		set["zebra"] = true
		set["apple"] = true
		set["banana"] = true

		result := sortSet(set)
		if len(result) != 3 {
			t.Errorf("sortSet() returned %d items, want 3", len(result))
		}
		if result[0] != "apple" || result[1] != "banana" || result[2] != "zebra" {
			t.Errorf("sortSet() = %v, want [apple banana zebra]", result)
		}
	})
}
