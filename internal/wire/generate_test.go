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
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnexport(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"", ""},
		{"a", "a"},
		{"ab", "ab"},
		{"A", "a"},
		{"AB", "ab"},
		{"A_", "a_"},
		{"ABc", "aBc"},
		{"ABC", "abc"},
		{"AB_", "ab_"},
		{"foo", "foo"},
		{"Foo", "foo"},
		{"HTTPClient", "httpClient"},
		{"IFace", "iFace"},
		{"SNAKE_CASE", "snake_CASE"},
		{"HTTP", "http"},
	}
	for _, test := range tests {
		if got := unexport(test.name); got != test.want {
			t.Errorf("unexport(%q) = %q; want %q", test.name, got, test.want)
		}
	}
}

func TestExport(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"", ""},
		{"a", "A"},
		{"ab", "Ab"},
		{"A", "A"},
		{"AB", "AB"},
		{"A_", "A_"},
		{"ABc", "ABc"},
		{"ABC", "ABC"},
		{"AB_", "AB_"},
		{"foo", "Foo"},
		{"Foo", "Foo"},
		{"HTTPClient", "HTTPClient"},
		{"httpClient", "HttpClient"},
		{"IFace", "IFace"},
		{"iFace", "IFace"},
		{"SNAKE_CASE", "SNAKE_CASE"},
		{"HTTP", "HTTP"},
	}
	for _, test := range tests {
		if got := export(test.name); got != test.want {
			t.Errorf("export(%q) = %q; want %q", test.name, got, test.want)
		}
	}
}

func TestDisambiguate(t *testing.T) {
	tests := []struct {
		name     string
		want     string
		collides map[string]bool
	}{
		{"foo", "foo", nil},
		{"foo", "foo2", map[string]bool{"foo": true}},
		{"foo", "foo3", map[string]bool{"foo": true, "foo1": true, "foo2": true}},
		{"foo1", "foo1_2", map[string]bool{"foo": true, "foo1": true, "foo2": true}},
		{"foo\u0661", "foo\u0661", map[string]bool{"foo": true, "foo1": true, "foo2": true}},
		{"foo\u0661", "foo\u06612", map[string]bool{"foo": true, "foo1": true, "foo2": true, "foo\u0661": true}},
		{"select", "select2", nil},
		{"var", "var2", nil},
		// Additional edge cases
		{"_", "_", nil},
		{"_", "_2", map[string]bool{"_": true}},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("disambiguate(%q, %v)", test.name, test.collides), func(t *testing.T) {
			got := disambiguate(test.name, func(name string) bool { return test.collides[name] })
			if !isIdent(got) {
				t.Errorf("%q is not an identifier", got)
			}
			if got != test.want {
				t.Errorf("got %q want %q", got, test.want)
			}
			if test.collides[got] {
				t.Errorf("%q collides", got)
			}
		})
	}
}

func TestTypeVariableName(t *testing.T) {
	var (
		boolT           = types.Typ[types.Bool]
		stringT         = types.Typ[types.String]
		fooVarT         = types.NewNamed(types.NewTypeName(0, nil, "foo", stringT), stringT, nil)
		nonameVarT      = types.NewNamed(types.NewTypeName(0, nil, "", stringT), stringT, nil)
		barVarInFooPkgT = types.NewNamed(types.NewTypeName(0, types.NewPackage("my.example/foo", "foo"), "bar", stringT), stringT, nil)
		ptrToFooT       = types.NewPointer(fooVarT)
	)
	tests := []struct {
		description     string
		typ             types.Type
		defaultName     string
		transformAppend string
		collides        map[string]bool
		want            string
	}{
		{"basic type", boolT, "", "", map[string]bool{}, "bool"},
		{"basic type with transform", boolT, "", "suffix", map[string]bool{}, "boolsuffix"},
		{"basic type with collision", boolT, "", "", map[string]bool{"bool": true}, "bool2"},
		{"basic type with transform and collision", boolT, "", "suffix", map[string]bool{"boolsuffix": true}, "boolsuffix2"},
		{"a different basic type", stringT, "", "", map[string]bool{}, "string"},
		{"named type", fooVarT, "", "", map[string]bool{}, "foo"},
		{"named type with transform", fooVarT, "", "suffix", map[string]bool{}, "foosuffix"},
		{"named type with collision", fooVarT, "", "", map[string]bool{"foo": true}, "foo2"},
		{"named type with transform and collision", fooVarT, "", "suffix", map[string]bool{"foosuffix": true}, "foosuffix2"},
		{"noname type", nonameVarT, "bar", "", map[string]bool{}, "bar"},
		{"noname type with transform", nonameVarT, "bar", "s", map[string]bool{}, "bars"},
		{"noname type with transform and collision", nonameVarT, "bar", "s", map[string]bool{"bars": true}, "bars2"},
		{"var in pkg type", barVarInFooPkgT, "", "", map[string]bool{}, "bar"},
		{"var in pkg type with collision", barVarInFooPkgT, "", "", map[string]bool{"bar": true}, "fooBar"},
		{"var in pkg type with double collision", barVarInFooPkgT, "", "", map[string]bool{"bar": true, "fooBar": true}, "bar2"},
		{"pointer type unwrap", ptrToFooT, "", "", map[string]bool{}, "foo"},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			got := typeVariableName(test.typ, test.defaultName, func(name string) string { return name + test.transformAppend }, func(name string) bool { return test.collides[name] })
			if !isIdent(got) {
				t.Errorf("%q is not an identifier", got)
			}
			if got != test.want {
				t.Errorf("got %q want %q", got, test.want)
			}
			if test.collides[got] {
				t.Errorf("%q collides", got)
			}
		})
	}
}

func TestZeroValue(t *testing.T) {
	pkg := types.NewPackage("example.com/test", "test")
	noQualify := func(p *types.Package) string { return p.Name() }

	tests := []struct {
		name string
		typ  types.Type
		want string
	}{
		{"bool", types.Typ[types.Bool], "false"},
		{"int", types.Typ[types.Int], "0"},
		{"int64", types.Typ[types.Int64], "0"},
		{"float64", types.Typ[types.Float64], "0"},
		{"complex128", types.Typ[types.Complex128], "0"},
		{"string", types.Typ[types.String], `""`},
		{"pointer", types.NewPointer(types.Typ[types.Int]), "nil"},
		{"slice", types.NewSlice(types.Typ[types.Int]), "nil"},
		{"map", types.NewMap(types.Typ[types.String], types.Typ[types.Int]), "nil"},
		{"chan", types.NewChan(types.SendRecv, types.Typ[types.Int]), "nil"},
		{"interface", types.NewInterfaceType(nil, nil), "nil"},
		{"signature", types.NewSignature(nil, nil, nil, false), "nil"},
		{"array", types.NewArray(types.Typ[types.Int], 3), "[3]int{}"},
		{"named struct", types.NewNamed(
			types.NewTypeName(0, pkg, "MyStruct", nil),
			types.NewStruct(nil, nil),
			nil,
		), "test.MyStruct{}"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := zeroValue(test.typ, noQualify)
			if got != test.want {
				t.Errorf("zeroValue(%s) = %q; want %q", test.name, got, test.want)
			}
		})
	}
}

func TestDetectOutputDir(t *testing.T) {
	tests := []struct {
		name    string
		paths   []string
		wantDir string
		wantErr bool
	}{
		{
			name:    "single file",
			paths:   []string{"/a/b/c.go"},
			wantDir: "/a/b",
		},
		{
			name:    "same directory",
			paths:   []string{"/a/b/c.go", "/a/b/d.go"},
			wantDir: "/a/b",
		},
		{
			name:    "no files",
			paths:   nil,
			wantErr: true,
		},
		{
			name:    "different directories",
			paths:   []string{"/a/b/c.go", "/x/y/z.go"},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir, err := detectOutputDir(test.paths)
			if test.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if dir != test.wantDir {
				t.Errorf("got %q; want %q", dir, test.wantDir)
			}
		})
	}
}

func TestAccessibleFrom(t *testing.T) {
	// Build a minimal types environment for testing.
	fset := token.NewFileSet()
	pkg1 := types.NewPackage("example.com/pkg1", "pkg1")
	pkg2 := types.NewPackage("example.com/pkg2", "pkg2")

	// Create exported and unexported objects in pkg1.
	exportedObj := types.NewVar(token.NoPos, pkg1, "Exported", types.Typ[types.Int])
	unexportedObj := types.NewVar(token.NoPos, pkg1, "unexported", types.Typ[types.Int])

	// Create idents referencing these objects.
	exportedIdent := ast.NewIdent("Exported")
	unexportedIdent := ast.NewIdent("unexported")

	// Scope setup: objects are in package scope.
	scope := pkg1.Scope()
	scope.Insert(exportedObj)
	scope.Insert(unexportedObj)

	info := &types.Info{
		Uses: map[*ast.Ident]types.Object{
			exportedIdent:   exportedObj,
			unexportedIdent: unexportedObj,
		},
	}
	_ = fset
	_ = pkg2

	t.Run("exported from same package", func(t *testing.T) {
		err := accessibleFrom(info, exportedIdent, pkg1.Path())
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("exported from other package", func(t *testing.T) {
		err := accessibleFrom(info, exportedIdent, pkg2.Path())
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("unexported from same package", func(t *testing.T) {
		err := accessibleFrom(info, unexportedIdent, pkg1.Path())
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("unexported from other package", func(t *testing.T) {
		err := accessibleFrom(info, unexportedIdent, pkg2.Path())
		if err == nil {
			t.Error("expected error for unexported identifier from different package")
		}
	})
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

func TestScrubError(t *testing.T) {
	// Use a platform-appropriate path for testing
	gopath := filepath.Join(string(os.PathSeparator), "test", "gopath")
	src := filepath.Join(gopath, "src")

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no gopath reference",
			input: "plain error message",
			want:  "plain error message",
		},
		{
			name:  "single path with line:col",
			input: src + string(os.PathSeparator) + "foo" + string(os.PathSeparator) + "bar.go:15:4: something",
			want:  "foo/bar.go:x:y: something",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := scrubError(gopath, test.input)
			if got != test.want {
				t.Errorf("scrubError(%q, %q) =\n%q\nwant:\n%q", gopath, test.input, got, test.want)
			}
		})
	}
}
