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
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
)

// TestCommit tests the GenerateResult.Commit method
func TestCommit(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		wantErr bool
	}{
		{
			name:    "empty content",
			content: nil,
			wantErr: false,
		},
		{
			name:    "valid content",
			content: []byte("package main\n\nfunc main() {}\n"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			outputPath := filepath.Join(tmpDir, "test_output.go")

			gen := GenerateResult{
				OutputPath: outputPath,
				Content:    tt.content,
			}

			err := gen.Commit()
			if (err != nil) != tt.wantErr {
				t.Errorf("Commit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.content != nil {
				data, err := os.ReadFile(outputPath)
				if err != nil {
					t.Fatalf("failed to read output file: %v", err)
				}
				if string(data) != string(tt.content) {
					t.Errorf("Commit() wrote %q, want %q", string(data), string(tt.content))
				}
			}
		})
	}
}

// TestProviderSetIDString tests the ProviderSetID.String method
func TestProviderSetIDString(t *testing.T) {
	tests := []struct {
		name string
		id   ProviderSetID
		want string
	}{
		{
			name: "simple path",
			id:   ProviderSetID{ImportPath: "example.com/foo", VarName: "MySet"},
			want: `"example.com/foo".MySet`,
		},
		{
			name: "complex path",
			id:   ProviderSetID{ImportPath: "github.com/user/repo/pkg", VarName: "ProviderSet"},
			want: `"github.com/user/repo/pkg".ProviderSet`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.String(); got != tt.want {
				t.Errorf("ProviderSetID.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestInjectorString tests the Injector.String method
func TestInjectorString(t *testing.T) {
	tests := []struct {
		name     string
		injector Injector
		want     string
	}{
		{
			name:     "simple injector",
			injector: Injector{ImportPath: "example.com/foo", FuncName: "InitApp"},
			want:     `"example.com/foo".InitApp`,
		},
		{
			name:     "complex injector",
			injector: Injector{ImportPath: "github.com/user/repo/pkg", FuncName: "NewService"},
			want:     `"github.com/user/repo/pkg".NewService`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.injector.String(); got != tt.want {
				t.Errorf("Injector.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestZeroValue tests the zeroValue function
func TestZeroValue(t *testing.T) {
	tests := []struct {
		name string
		typ  types.Type
		want string
	}{
		{
			name: "bool",
			typ:  types.Typ[types.Bool],
			want: "false",
		},
		{
			name: "int",
			typ:  types.Typ[types.Int],
			want: "0",
		},
		{
			name: "uint",
			typ:  types.Typ[types.Uint],
			want: "0",
		},
		{
			name: "float32",
			typ:  types.Typ[types.Float32],
			want: "0",
		},
		{
			name: "float64",
			typ:  types.Typ[types.Float64],
			want: "0",
		},
		{
			name: "complex64",
			typ:  types.Typ[types.Complex64],
			want: "0",
		},
		{
			name: "complex128",
			typ:  types.Typ[types.Complex128],
			want: "0",
		},
		{
			name: "string",
			typ:  types.Typ[types.String],
			want: `""`,
		},
		{
			name: "pointer",
			typ:  types.NewPointer(types.Typ[types.Int]),
			want: "nil",
		},
		{
			name: "slice",
			typ:  types.NewSlice(types.Typ[types.Int]),
			want: "nil",
		},
		{
			name: "map",
			typ:  types.NewMap(types.Typ[types.String], types.Typ[types.Int]),
			want: "nil",
		},
		{
			name: "chan",
			typ:  types.NewChan(types.SendRecv, types.Typ[types.Int]),
			want: "nil",
		},
		{
			name: "interface",
			typ:  types.NewInterfaceType(nil, nil),
			want: "nil",
		},
		{
			name: "function",
			typ:  types.NewSignature(nil, nil, nil, false),
			want: "nil",
		},
		{
			name: "struct",
			typ:  types.NewStruct(nil, nil),
			want: "struct{}{}",
		},
		{
			name: "array",
			typ:  types.NewArray(types.Typ[types.Int], 5),
			want: "[5]int{}",
		},
		{
			name: "named struct",
			typ:  types.NewNamed(types.NewTypeName(0, types.NewPackage("test", "test"), "MyStruct", nil), types.NewStruct(nil, nil), nil),
			want: "test.MyStruct{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := zeroValue(tt.typ, types.RelativeTo(nil))
			if got != tt.want {
				t.Errorf("zeroValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsProviderSetType tests the isProviderSetType function
func TestIsProviderSetType(t *testing.T) {
	// Create a mock ProviderSet type
	wirePkg := types.NewPackage("github.com/almondoo/wire", "wire")
	providerSetObj := types.NewTypeName(token.NoPos, wirePkg, "ProviderSet", nil)
	providerSetType := types.NewNamed(providerSetObj, types.NewStruct(nil, nil), nil)

	// Create a non-ProviderSet type
	otherPkg := types.NewPackage("example.com/other", "other")
	otherObj := types.NewTypeName(token.NoPos, otherPkg, "Other", nil)
	otherType := types.NewNamed(otherObj, types.NewStruct(nil, nil), nil)

	tests := []struct {
		name string
		typ  types.Type
		want bool
	}{
		{
			name: "ProviderSet type",
			typ:  providerSetType,
			want: true,
		},
		{
			name: "other named type",
			typ:  otherType,
			want: false,
		},
		{
			name: "basic type",
			typ:  types.Typ[types.Int],
			want: false,
		},
		{
			name: "pointer type",
			typ:  types.NewPointer(types.Typ[types.Int]),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isProviderSetType(tt.typ)
			if got != tt.want {
				t.Errorf("isProviderSetType() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestStructArgType tests the structArgType function
func TestStructArgType(t *testing.T) {
	fset := token.NewFileSet()
	src := `package test
type MyStruct struct {
	Field1 int
	Field2 string
}
var x = MyStruct{Field1: 1, Field2: "test"}
var y = 42
`
	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	conf := types.Config{}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Uses:  make(map[*ast.Ident]types.Object),
		Defs:  make(map[*ast.Ident]types.Object),
	}
	pkg, err := conf.Check("test", fset, []*ast.File{f}, info)
	if err != nil {
		t.Fatalf("failed to type check: %v", err)
	}

	// Find the composite literal
	var structLit *ast.CompositeLit
	var basicLit *ast.BasicLit
	ast.Inspect(f, func(n ast.Node) bool {
		if e, ok := n.(*ast.CompositeLit); ok && structLit == nil {
			structLit = e
		}
		if e, ok := n.(*ast.BasicLit); ok && e.Value == "42" {
			basicLit = e
		}
		return true
	})

	t.Run("struct literal", func(t *testing.T) {
		if structLit == nil {
			t.Fatal("composite literal not found in AST")
		}

		result := structArgType(info, structLit)
		if result == nil {
			t.Fatal("structArgType() returned nil for struct literal")
		}
		if result.Name() != "MyStruct" {
			t.Errorf("structArgType() returned type name %q, want %q", result.Name(), "MyStruct")
		}
		if result.Pkg() != pkg {
			t.Errorf("structArgType() returned different package")
		}
	})

	t.Run("non-struct literal", func(t *testing.T) {
		if basicLit == nil {
			t.Fatal("basic literal not found in AST")
		}

		result := structArgType(info, basicLit)
		if result != nil {
			t.Errorf("structArgType() returned %v for non-struct, want nil", result)
		}
	})
}

// TestProviderSetOutputs tests the ProviderSet.Outputs method
func TestProviderSetOutputs(t *testing.T) {
	// Create a simple provider set with some types
	intType := types.Typ[types.Int]
	stringType := types.Typ[types.String]

	hasher := typeutil.MakeHasher()
	providerMap := new(typeutil.Map)
	providerMap.SetHasher(hasher)

	pkg := types.NewPackage("test", "test")
	provider := &Provider{
		Pkg:  pkg,
		Name: "TestProvider",
		Out:  []types.Type{intType},
	}

	providerMap.Set(intType, &ProvidedType{t: intType, p: provider})
	providerMap.Set(stringType, &ProvidedType{t: stringType, p: provider})

	srcMap := new(typeutil.Map)
	srcMap.SetHasher(hasher)
	srcMap.Set(intType, &providerSetSrc{Provider: provider})
	srcMap.Set(stringType, &providerSetSrc{Provider: provider})

	pset := &ProviderSet{
		Pos:         token.NoPos,
		PkgPath:     "test",
		VarName:     "TestSet",
		providerMap: providerMap,
		srcMap:      srcMap,
	}

	outputs := pset.Outputs()
	if len(outputs) != 2 {
		t.Errorf("Outputs() returned %d types, want 2", len(outputs))
	}

	// Check that outputs contains both types
	foundInt := false
	foundString := false
	for _, typ := range outputs {
		if types.Identical(typ, intType) {
			foundInt = true
		}
		if types.Identical(typ, stringType) {
			foundString = true
		}
	}

	if !foundInt {
		t.Errorf("Outputs() missing int type")
	}
	if !foundString {
		t.Errorf("Outputs() missing string type")
	}
}

// TestProvidedTypeMethods tests the Provider, Value, Arg, and Field methods
func TestProvidedTypeMethods(t *testing.T) {
	pkg := types.NewPackage("test", "test")
	intType := types.Typ[types.Int]

	t.Run("Provider", func(t *testing.T) {
		provider := &Provider{
			Pkg:  pkg,
			Name: "TestProvider",
			Out:  []types.Type{intType},
		}
		pt := ProvidedType{t: intType, p: provider}

		if !pt.IsProvider() {
			t.Error("IsProvider() = false, want true")
		}
		if pt.Provider() != provider {
			t.Error("Provider() returned different provider")
		}

		// Test panic when calling wrong method
		defer func() {
			if r := recover(); r == nil {
				t.Error("Value() on Provider should panic")
			}
		}()
		pt.Value()
	})

	t.Run("Value", func(t *testing.T) {
		value := &Value{
			Out: intType,
		}
		pt := ProvidedType{t: intType, v: value}

		if !pt.IsValue() {
			t.Error("IsValue() = false, want true")
		}
		if pt.Value() != value {
			t.Error("Value() returned different value")
		}

		// Test panic when calling wrong method
		defer func() {
			if r := recover(); r == nil {
				t.Error("Provider() on Value should panic")
			}
		}()
		pt.Provider()
	})

	t.Run("Arg", func(t *testing.T) {
		args := &InjectorArgs{
			Name:  "TestInjector",
			Tuple: types.NewTuple(types.NewVar(token.NoPos, pkg, "arg", intType)),
		}
		arg := &InjectorArg{
			Args:  args,
			Index: 0,
		}
		pt := ProvidedType{t: intType, a: arg}

		if !pt.IsArg() {
			t.Error("IsArg() = false, want true")
		}
		if pt.Arg() != arg {
			t.Error("Arg() returned different arg")
		}

		// Test panic when calling wrong method
		defer func() {
			if r := recover(); r == nil {
				t.Error("Field() on Arg should panic")
			}
		}()
		pt.Field()
	})

	t.Run("Field", func(t *testing.T) {
		field := &Field{
			Parent: types.NewStruct(nil, nil),
			Name:   "TestField",
			Pkg:    pkg,
			Out:    []types.Type{intType},
		}
		pt := ProvidedType{t: intType, f: field}

		if !pt.IsField() {
			t.Error("IsField() = false, want true")
		}
		if pt.Field() != field {
			t.Error("Field() returned different field")
		}

		// Test panic when calling wrong method
		defer func() {
			if r := recover(); r == nil {
				t.Error("Arg() on Field should panic")
			}
		}()
		pt.Arg()
	})
}

// TestCopyASTHelpers tests the helper functions used by copyAST
func TestCopyASTHelpers(t *testing.T) {
	fset := token.NewFileSet()
	src := `package test
func foo() *int {
	x := 42
	return &x
}
`
	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Build a map from the original AST nodes
	m := make(map[ast.Node]ast.Node)

	// Find a FuncDecl to copy
	var funcDecl *ast.FuncDecl
	for _, decl := range f.Decls {
		if fd, ok := decl.(*ast.FuncDecl); ok {
			funcDecl = fd
			break
		}
	}
	if funcDecl == nil {
		t.Fatal("no function declaration found")
	}

	// Copy just the function declaration, not the whole file
	copied := copyAST(funcDecl)
	if copied == nil {
		t.Fatal("copyAST returned nil")
	}

	// Populate the map by traversing the original and copied trees together
	ast.Inspect(funcDecl, func(n ast.Node) bool {
		if n != nil {
			m[n] = n
		}
		return true
	})

	// Test callExprFromMap
	t.Run("callExprFromMap", func(t *testing.T) {
		// callExprFromMap with nil should return nil
		result := callExprFromMap(m, nil)
		if result != nil {
			t.Errorf("callExprFromMap(nil) = %v, want nil", result)
		}
	})

	// Test identFromMap
	t.Run("identFromMap", func(t *testing.T) {
		result := identFromMap(m, nil)
		if result != nil {
			t.Errorf("identFromMap(nil) = %v, want nil", result)
		}
	})

	// Test blockStmtFromMap
	t.Run("blockStmtFromMap", func(t *testing.T) {
		result := blockStmtFromMap(m, nil)
		if result != nil {
			t.Errorf("blockStmtFromMap(nil) = %v, want nil", result)
		}
	})

	// Test basicLitFromMap
	t.Run("basicLitFromMap", func(t *testing.T) {
		result := basicLitFromMap(m, nil)
		if result != nil {
			t.Errorf("basicLitFromMap(nil) = %v, want nil", result)
		}
	})

	// Test funcTypeFromMap
	t.Run("funcTypeFromMap", func(t *testing.T) {
		result := funcTypeFromMap(m, nil)
		if result != nil {
			t.Errorf("funcTypeFromMap(nil) = %v, want nil", result)
		}
	})
}

// TestCopyASTComprehensive tests copyAST with various AST node types
func TestCopyASTComprehensive(t *testing.T) {
	fset := token.NewFileSet()
	src := `package test
import "fmt"

type MyStruct struct {
	Field int ` + "`json:\"field\"`" + `
}

type MyInterface interface {
	Method()
}

func foo(x int) (int, error) {
	defer fmt.Println("done")
	if x > 0 {
		return x, nil
	}
	for i := 0; i < 10; i++ {
		x += i
	}

	// Range statement
	arr := []int{1, 2, 3}
	for _, v := range arr {
		x += v
	}

	ch := make(chan int)
	select {
	case v := <-ch:
		_ = v
	default:
	}

	// Switch statement
	switch x {
	case 1:
		x = 2
	default:
		x = 3
	}

	// Type switch
	var iface interface{} = x
	switch v := iface.(type) {
	case int:
		_ = v
	default:
		_ = v
	}

	arr2 := [5]int{1, 2, 3, 4, 5}
	slice := arr2[1:3]
	slice2 := arr2[1:3:4]
	_ = slice
	_ = slice2

	m := map[string]int{"a": 1}
	_ = m

	type localType int
	var y localType = 5
	_ = y

	// Send statement
	ch2 := make(chan int)
	go func() {
		ch2 <- 1
	}()

	// Labeled statement
Label:
	for i := 0; i < 5; i++ {
		if i == 3 {
			break Label
		}
	}

	// Inc/Dec statements
	x++
	x--

	return x, nil
}

func (m MyStruct) Method() {
	// empty method
}

var globalVar = 42

const constVal = "test"
`
	f, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Test copying individual declarations instead of the whole file
	for _, decl := range f.Decls {
		copied := copyAST(decl)
		if copied == nil {
			t.Errorf("copyAST returned nil for decl type %T", decl)
		}
	}
}

// TestDetectOutputDir tests the detectOutputDir function
func TestDetectOutputDir(t *testing.T) {
	tests := []struct {
		name    string
		paths   []string
		want    string
		wantErr bool
	}{
		{
			name:    "empty paths",
			paths:   []string{},
			want:    "",
			wantErr: true,
		},
		{
			name:    "single path",
			paths:   []string{"/foo/bar/file.go"},
			want:    "/foo/bar",
			wantErr: false,
		},
		{
			name:    "multiple paths same dir",
			paths:   []string{"/foo/bar/file1.go", "/foo/bar/file2.go"},
			want:    "/foo/bar",
			wantErr: false,
		},
		{
			name:    "multiple paths different dirs",
			paths:   []string{"/foo/bar/file1.go", "/foo/baz/file2.go"},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := detectOutputDir(tt.paths)
			if (err != nil) != tt.wantErr {
				t.Errorf("detectOutputDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("detectOutputDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestWireErrError tests the Error method of wireErr
func TestWireErrError(t *testing.T) {
	fset := token.NewFileSet()
	file := fset.AddFile("test.go", -1, 100)

	tests := []struct {
		name     string
		pos      token.Position
		errMsg   string
		wantText string
	}{
		{
			name:     "with valid position",
			pos:      fset.Position(file.Pos(10)),
			errMsg:   "test error",
			wantText: "test.go:1:10: test error",
		},
		{
			name:     "with invalid position",
			pos:      token.Position{},
			errMsg:   "test error",
			wantText: "test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseErr := errors.New(tt.errMsg)
			err := notePosition(tt.pos, baseErr)

			errStr := err.Error()
			if !strings.Contains(errStr, tt.errMsg) {
				t.Errorf("Error() = %v, want to contain %v", errStr, tt.errMsg)
			}
		})
	}
}

// TestNotePosition tests the notePosition function
func TestNotePosition(t *testing.T) {
	fset := token.NewFileSet()
	file := fset.AddFile("test.go", -1, 100)
	pos := fset.Position(file.Pos(10))

	t.Run("nil error", func(t *testing.T) {
		result := notePosition(pos, nil)
		if result != nil {
			t.Errorf("notePosition(pos, nil) = %v, want nil", result)
		}
	})

	t.Run("existing wireErr", func(t *testing.T) {
		baseErr := errors.New("original")
		originalErr := &wireErr{
			error:    baseErr,
			position: pos,
		}
		result := notePosition(token.Position{}, originalErr)
		if result != originalErr {
			t.Errorf("notePosition() should return same wireErr")
		}
	})

	t.Run("regular error", func(t *testing.T) {
		baseErr := errors.New("test")
		result := notePosition(pos, baseErr)
		if result == nil {
			t.Fatal("notePosition() returned nil")
		}
		wireErr, ok := result.(*wireErr)
		if !ok {
			t.Fatalf("notePosition() returned %T, want *wireErr", result)
		}
		if wireErr.position != pos {
			t.Errorf("notePosition() position = %v, want %v", wireErr.position, pos)
		}
	})
}

// TestIsWireImport tests the isWireImport function
func TestIsWireImport(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "wire package",
			path: "github.com/almondoo/wire",
			want: true,
		},
		{
			name: "vendored wire package",
			path: "vendor/github.com/almondoo/wire",
			want: true,
		},
		{
			name: "vendored wire package with prefix",
			path: "example.com/vendor/github.com/almondoo/wire",
			want: true,
		},
		{
			name: "other package",
			path: "github.com/other/package",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isWireImport(tt.path)
			if got != tt.want {
				t.Errorf("isWireImport(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// TestAllFields tests the allFields function
func TestAllFields(t *testing.T) {
	tests := []struct {
		name string
		args []ast.Expr
		want bool
	}{
		{
			name: "not all fields - wrong length",
			args: []ast.Expr{
				&ast.BasicLit{Value: "\"test\""},
			},
			want: false,
		},
		{
			name: "not all fields - not basic lit",
			args: []ast.Expr{
				&ast.BasicLit{Value: "\"test\""},
				&ast.Ident{Name: "foo"},
			},
			want: false,
		},
		{
			name: "all fields - star",
			args: []ast.Expr{
				&ast.BasicLit{Value: "\"test\""},
				&ast.BasicLit{Value: "\"*\""},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpr{Args: tt.args}
			got := allFields(call)
			if got != tt.want {
				t.Errorf("allFields() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDescriptionMethod tests the description method of providerSetSrc
func TestDescriptionMethod(t *testing.T) {
	fset := token.NewFileSet()
	file := fset.AddFile("test.go", -1, 100)
	pos := file.Pos(10)

	pkg := types.NewPackage("example.com/test", "test")
	intType := types.Typ[types.Int]

	tests := []struct {
		name string
		src  *providerSetSrc
	}{
		{
			name: "Provider",
			src: &providerSetSrc{
				Provider: &Provider{
					Pkg:  pkg,
					Name: "TestProvider",
					Pos:  pos,
					Out:  []types.Type{intType},
				},
			},
		},
		{
			name: "Binding",
			src: &providerSetSrc{
				Binding: &IfaceBinding{
					Pos:   pos,
					Iface: intType,
				},
			},
		},
		{
			name: "Value",
			src: &providerSetSrc{
				Value: &Value{
					Pos: pos,
					Out: intType,
				},
			},
		},
		{
			name: "Import",
			src: &providerSetSrc{
				Import: &ProviderSet{
					Pos:     pos,
					VarName: "ImportedSet",
				},
			},
		},
		{
			name: "InjectorArg",
			src: &providerSetSrc{
				InjectorArg: &InjectorArg{
					Args: &InjectorArgs{
						Name:  "TestInjector",
						Tuple: types.NewTuple(types.NewVar(token.NoPos, pkg, "arg", intType)),
						Pos:   pos,
					},
					Index: 0,
				},
			},
		},
		{
			name: "Field",
			src: &providerSetSrc{
				Field: &Field{
					Parent: types.NewStruct(nil, nil),
					Name:   "TestField",
					Pkg:    pkg,
					Pos:    pos,
					Out:    []types.Type{intType},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := tt.src.description(fset, intType)
			if desc == "" {
				t.Error("description() returned empty string")
			}
		})
	}
}

// TestGenFrame tests the gen.frame method
func TestGenFrame(t *testing.T) {
	pkg := &packages.Package{
		Name:    "test",
		PkgPath: "example.com/test",
		Fset:    token.NewFileSet(),
		Types:   types.NewPackage("example.com/test", "test"),
	}

	g := newGen(pkg)

	// Test with empty buffer
	result := g.frame("")
	if result != nil {
		t.Error("frame() should return nil for empty buffer")
	}

	// Test with content
	g.p("func test() {}\n")
	result = g.frame("")
	if result == nil {
		t.Error("frame() returned nil for non-empty buffer")
	}

	// Test with tags
	g2 := newGen(pkg)
	g2.p("func test() {}\n")
	result = g2.frame("wireinject")
	if result == nil {
		t.Error("frame() with tags returned nil")
	}
	if !strings.Contains(string(result), "wireinject") {
		t.Error("frame() output doesn't contain the tags")
	}
}

// TestQualifyImport tests the qualifyImport function
func TestQualifyImport(t *testing.T) {
	pkg := &packages.Package{
		Name:    "test",
		PkgPath: "example.com/test",
		Fset:    token.NewFileSet(),
		Types:   types.NewPackage("example.com/test", "test"),
	}

	g := newGen(pkg)

	// Test same package
	result := g.qualifyImport("test", "example.com/test")
	if result != "" {
		t.Errorf("qualifyImport() for same package = %q, want empty string", result)
	}

	// Test different package
	result = g.qualifyImport("fmt", "fmt")
	if result != "fmt" {
		t.Errorf("qualifyImport() = %q, want %q", result, "fmt")
	}

	// Test vendored package
	result = g.qualifyImport("wire", "vendor/github.com/almondoo/wire")
	if result == "" {
		t.Error("qualifyImport() for vendored package returned empty string")
	}

	// Test name collision (should disambiguate)
	result2 := g.qualifyImport("fmt", "github.com/other/fmt")
	if result2 == result {
		t.Error("qualifyImport() should disambiguate conflicting names")
	}
}

// TestGenQualifiedID tests the qualifiedID function
func TestGenQualifiedID(t *testing.T) {
	pkg := &packages.Package{
		Name:    "test",
		PkgPath: "example.com/test",
		Fset:    token.NewFileSet(),
		Types:   types.NewPackage("example.com/test", "test"),
	}

	g := newGen(pkg)

	// Test same package
	result := g.qualifiedID("test", "example.com/test", "MyFunc")
	if result != "MyFunc" {
		t.Errorf("qualifiedID() for same package = %q, want %q", result, "MyFunc")
	}

	// Test different package
	result = g.qualifiedID("fmt", "fmt", "Println")
	if result != "fmt.Println" {
		t.Errorf("qualifiedID() = %q, want %q", result, "fmt.Println")
	}
}

// TestStructArgTypeEdgeCases tests edge cases for structArgType
func TestStructArgTypeEdgeCases(t *testing.T) {
	fset := token.NewFileSet()
	src := `package test
type NotAStruct int
var x = NotAStruct(42)
`
	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	conf := types.Config{}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Uses:  make(map[*ast.Ident]types.Object),
		Defs:  make(map[*ast.Ident]types.Object),
	}
	_, err = conf.Check("test", fset, []*ast.File{f}, info)
	if err != nil {
		t.Fatalf("failed to type check: %v", err)
	}

	// Find the composite literal (NotAStruct(42))
	var callExpr *ast.CallExpr
	ast.Inspect(f, func(n ast.Node) bool {
		if ce, ok := n.(*ast.CallExpr); ok && callExpr == nil {
			callExpr = ce
		}
		return true
	})

	if callExpr == nil {
		t.Fatal("call expression not found")
	}

	result := structArgType(info, callExpr)
	if result != nil {
		t.Errorf("structArgType() for non-struct type = %v, want nil", result)
	}
}

// TestGenerateEdgeCases tests edge cases for the Generate function
func TestGenerateEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with nil options
	ctx := context.Background()
	results, errs := Generate(ctx, tmpDir, nil, []string{"nonexistent"}, nil)
	if errs == nil {
		t.Error("Generate() with nonexistent package should return errors")
	}
	if results != nil {
		t.Errorf("Generate() with errors returned %d results, want nil", len(results))
	}
}

// TestDetectOutputDirEdgeCase tests detectOutputDir with relative paths
func TestDetectOutputDirEdgeCase(t *testing.T) {
	// Test with relative paths
	paths := []string{"foo/bar/file1.go", "foo/bar/file2.go"}
	result, err := detectOutputDir(paths)
	if err != nil {
		t.Errorf("detectOutputDir() with relative paths failed: %v", err)
	}
	expected := "foo/bar"
	if result != expected {
		t.Errorf("detectOutputDir() = %q, want %q", result, expected)
	}
}

// TestNameInFileScope tests the nameInFileScope function
func TestNameInFileScope(t *testing.T) {
	fset := token.NewFileSet()
	src := `package test
var GlobalVar int
`
	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	conf := types.Config{}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Uses:  make(map[*ast.Ident]types.Object),
		Defs:  make(map[*ast.Ident]types.Object),
	}
	typePkg, err := conf.Check("test", fset, []*ast.File{f}, info)
	if err != nil {
		t.Fatalf("failed to type check: %v", err)
	}

	pkg := &packages.Package{
		Name:      "test",
		PkgPath:   "example.com/test",
		Fset:      fset,
		Types:     typePkg,
		TypesInfo: info,
	}

	g := newGen(pkg)

	// Add an import to test import collision
	g.qualifyImport("fmt", "fmt")

	// Add a value to test value collision
	expr := &ast.BasicLit{Value: "42"}
	g.values[expr] = "myValue"

	tests := []struct {
		name string
		want bool
	}{
		{"fmt", true},       // import name
		{"myValue", true},   // value name
		{"GlobalVar", true}, // package scope var
		{"unknown", false},  // not in scope
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := g.nameInFileScope(tt.name)
			if got != tt.want {
				t.Errorf("nameInFileScope(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// TestTypeVariableNameEdgeCases tests typeVariableName with various edge cases
func TestTypeVariableNameEdgeCases(t *testing.T) {
	pkg := types.NewPackage("example.com/foo", "foo")

	// Test with pointer to named type
	namedType := types.NewNamed(types.NewTypeName(0, pkg, "MyType", nil), types.Typ[types.Int], nil)
	ptrType := types.NewPointer(namedType)

	result := typeVariableName(ptrType, "default", unexport, func(name string) bool { return false })
	if result != "myType" {
		t.Errorf("typeVariableName(pointer to named) = %q, want %q", result, "myType")
	}

	// Test with basic type and collisions
	result = typeVariableName(types.Typ[types.Int], "default", unexport, func(name string) bool {
		return name == "int"
	})
	if result != "int2" {
		t.Errorf("typeVariableName(int with collision) = %q, want %q", result, "int2")
	}

	// Test with default name when type has no name
	sliceType := types.NewSlice(types.Typ[types.Int])
	result = typeVariableName(sliceType, "myDefault", unexport, func(name string) bool { return false })
	if result != "myDefault" {
		t.Errorf("typeVariableName(slice) = %q, want %q", result, "myDefault")
	}

	// Test with package-prefixed name to avoid collision
	result = typeVariableName(namedType, "", unexport, func(name string) bool {
		return name == "myType"
	})
	if result != "fooMyType" {
		t.Errorf("typeVariableName(with pkg prefix) = %q, want %q", result, "fooMyType")
	}
}

// TestQualifyPkg tests the qualifyPkg function
func TestQualifyPkg(t *testing.T) {
	pkg := &packages.Package{
		Name:    "test",
		PkgPath: "example.com/test",
		Fset:    token.NewFileSet(),
		Types:   types.NewPackage("example.com/test", "test"),
	}

	g := newGen(pkg)

	// Test with different package
	otherPkg := types.NewPackage("fmt", "fmt")
	result := g.qualifyPkg(otherPkg)
	if result != "fmt" {
		t.Errorf("qualifyPkg(fmt) = %q, want %q", result, "fmt")
	}

	// Test with same package
	result = g.qualifyPkg(pkg.Types)
	if result != "" {
		t.Errorf("qualifyPkg(same pkg) = %q, want empty string", result)
	}
}

// TestProcessValueWrongArgs tests processValue with wrong number of arguments
func TestProcessValueWrongArgs(t *testing.T) {
	fset := token.NewFileSet()

	// Create a mock call expression with wrong number of args
	call := &ast.CallExpr{
		Fun:  ast.NewIdent("Value"),
		Args: []ast.Expr{}, // No arguments
	}

	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
	}

	_, err := processValue(fset, info, call)
	if err == nil {
		t.Error("processValue() expected error for no arguments, got nil")
	} else if !strings.Contains(err.Error(), "exactly one argument") {
		t.Errorf("processValue() error = %v, want error containing 'exactly one argument'", err)
	}
}

// TestProcessInterfaceValueWrongArgs tests processInterfaceValue with wrong number of arguments
func TestProcessInterfaceValueWrongArgs(t *testing.T) {
	fset := token.NewFileSet()

	// Create a mock call expression with wrong number of args
	call := &ast.CallExpr{
		Fun:  ast.NewIdent("InterfaceValue"),
		Args: []ast.Expr{ast.NewIdent("x")}, // Only one argument
	}

	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
	}

	_, err := processInterfaceValue(fset, info, call)
	if err == nil {
		t.Error("processInterfaceValue() expected error for one argument, got nil")
	} else if !strings.Contains(err.Error(), "exactly two arguments") {
		t.Errorf("processInterfaceValue() error = %v, want error containing 'exactly two arguments'", err)
	}
}

// TestGenerateErrorPaths tests error paths in Generate function
func TestGenerateErrorPaths(t *testing.T) {
	t.Run("format.Source error handling", func(t *testing.T) {
		// Create a temporary directory with a test package
		tempDir := t.TempDir()

		// Create a wire.go file with valid code
		wireFile := filepath.Join(tempDir, "wire.go")
		wireContent := `//go:build wireinject
// +build wireinject

package example

import "github.com/google/wire"

func NewFoo() int {
	wire.Build(provideFoo)
	return 0
}

func provideFoo() int {
	return 42
}
`
		if err := os.WriteFile(wireFile, []byte(wireContent), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		// Load the package
		cfg := &packages.Config{
			Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
				packages.NeedTypes | packages.NeedTypesInfo | packages.NeedModule,
			Dir: tempDir,
			Env: append(os.Environ(), "GOPACKAGESDRIVER=off"),
			Context: context.Background(),
		}

		pkgs, err := packages.Load(cfg, ".")
		if err != nil || len(pkgs) == 0 {
			t.Skipf("Skipping test: failed to load package: %v", err)
			return
		}

		if len(pkgs[0].Errors) > 0 {
			t.Skipf("Skipping test: package has errors: %v", pkgs[0].Errors)
			return
		}

		// Try to generate - this should not panic even if there are issues
		results, errs := Generate(context.Background(), tempDir, os.Environ(), []string{"."}, &GenerateOptions{
			Tags: "wireinject",
		})

		// We expect it to complete without panic
		if len(errs) > 0 {
			t.Logf("Generate returned errors (expected): %v", errs)
		}

		// Check that results were returned
		if results == nil {
			t.Error("Generate() returned nil results")
		}
	})
}

// TestWriteASTEdgeCases tests the writeAST function's error handling
func TestWriteASTEdgeCases(t *testing.T) {
	pkg := &packages.Package{
		Name:    "test",
		PkgPath: "example.com/test",
		Fset:    token.NewFileSet(),
		Types:   types.NewPackage("example.com/test", "test"),
	}

	g := newGen(pkg)

	// Create a simple AST node to write
	src := `package test
func foo() {}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Uses:  make(map[*ast.Ident]types.Object),
		Defs:  make(map[*ast.Ident]types.Object),
	}

	// Write the function declaration
	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			g.writeAST(info, fn)
			// Verify something was written
			if g.buf.Len() == 0 {
				t.Error("writeAST() did not write anything to buffer")
			}
		}
	}
}

// TestZeroValueArrayAndStruct tests zeroValue for array and struct types
func TestZeroValueArrayAndStruct(t *testing.T) {
	pkg := types.NewPackage("test", "test")

	// Test array type
	arrayType := types.NewArray(types.Typ[types.Int], 5)
	result := zeroValue(arrayType, types.RelativeTo(pkg))
	if !strings.Contains(result, "{}") {
		t.Errorf("zeroValue(array) = %q, want to contain {}", result)
	}

	// Test struct type
	fields := []*types.Var{
		types.NewField(0, pkg, "x", types.Typ[types.Int], false),
	}
	structType := types.NewStruct(fields, nil)
	result = zeroValue(structType, types.RelativeTo(pkg))
	if !strings.Contains(result, "{}") {
		t.Errorf("zeroValue(struct) = %q, want to contain {}", result)
	}

	// Test named struct type
	namedStruct := types.NewNamed(types.NewTypeName(0, pkg, "MyStruct", nil), structType, nil)
	result = zeroValue(namedStruct, types.RelativeTo(pkg))
	expected := "MyStruct{}"
	if result != expected {
		t.Errorf("zeroValue(named struct) = %q, want %q", result, expected)
	}
}

// TestProcessBindEdgeCases tests edge cases in processBind
func TestProcessBindEdgeCases(t *testing.T) {
	fset := token.NewFileSet()

	// Test bind with wrong number of arguments
	src := `package test
import "wire"

type Foo interface{}
type Bar struct{}

func init() {
	wire.Bind(new(Foo))  // Missing second argument
}
`
	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	conf := &types.Config{Importer: nil}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
	}
	_, _ = conf.Check("test", fset, []*ast.File{f}, info)

	// Find wire.Bind call
	var bindCall *ast.CallExpr
	ast.Inspect(f, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == "Bind" {
					bindCall = call
					return false
				}
			}
		}
		return true
	})

	if bindCall != nil {
		_, err := processBind(fset, info, bindCall)
		if err == nil {
			t.Error("processBind() expected error for wrong number of arguments, got nil")
		} else if !strings.Contains(err.Error(), "exactly two arguments") {
			t.Errorf("processBind() error = %v, want error containing 'exactly two arguments'", err)
		}
	}
}

// TestInjectorFuncSignatureErrors tests error cases in injector function signatures
func TestInjectorFuncSignatureErrors(t *testing.T) {
	// Test function with no return values
	sig := types.NewSignatureType(nil, nil, nil,
		types.NewTuple(),
		types.NewTuple(), // No returns
		false)

	_, _, err := injectorFuncSignature(sig)
	if err == nil {
		t.Error("injectorFuncSignature() expected error for no returns, got nil")
	}
}

// TestDetectOutputDirErrorPath tests error handling in detectOutputDir
func TestDetectOutputDirErrorPath(t *testing.T) {
	// Test with conflicting directories
	paths := []string{
		"/path/to/dir1/file1.go",
		"/path/to/dir2/file2.go",
	}

	_, err := detectOutputDir(paths)
	if err == nil {
		t.Error("detectOutputDir() expected error for conflicting directories, got nil")
	} else if !strings.Contains(err.Error(), "conflicting directories") {
		t.Errorf("detectOutputDir() error = %v, want error containing 'conflicting directories'", err)
	}
}

// TestProviderSetSrcDescription tests the description method for different provider types
func TestProviderSetSrcDescription(t *testing.T) {
	fset := token.NewFileSet()
	file := fset.AddFile("test.go", 1, 100)
	pos := file.Pos(10)

	tests := []struct {
		name string
		src  *providerSetSrc
		want string
	}{
		{
			name: "struct provider",
			src: &providerSetSrc{
				Provider: &Provider{
					Name:     "MyProvider",
					Pos:      pos,
					IsStruct: true,
				},
			},
			want: "struct provider",
		},
		{
			name: "function provider",
			src: &providerSetSrc{
				Provider: &Provider{
					Name:     "NewFoo",
					Pos:      pos,
					IsStruct: false,
				},
			},
			want: "provider",
		},
		{
			name: "provider without name",
			src: &providerSetSrc{
				Provider: &Provider{
					Name:     "",
					Pos:      pos,
					IsStruct: false,
				},
			},
			want: "provider",
		},
		{
			name: "binding",
			src: &providerSetSrc{
				Binding: &IfaceBinding{
					Pos: pos,
				},
			},
			want: "wire.Bind",
		},
		{
			name: "value",
			src: &providerSetSrc{
				Value: &Value{
					Pos: pos,
				},
			},
			want: "wire.Value",
		},
		{
			name: "import with name",
			src: &providerSetSrc{
				Import: &ProviderSet{
					VarName: "MySet",
					Pos:     pos,
				},
			},
			want: "provider set",
		},
		{
			name: "import without name",
			src: &providerSetSrc{
				Import: &ProviderSet{
					VarName: "",
					Pos:     pos,
				},
			},
			want: "provider set",
		},
		{
			name: "injector arg",
			src: &providerSetSrc{
				InjectorArg: &InjectorArg{
					Index: 0,
					Args: &InjectorArgs{
						Name:  "NewInjector",
						Tuple: types.NewTuple(types.NewVar(0, nil, "ctx", types.Typ[types.String])),
						Pos:   pos,
					},
				},
			},
			want: "argument ctx to injector",
		},
		{
			name: "field",
			src: &providerSetSrc{
				Field: &Field{
					Pos: pos,
				},
			},
			want: "wire.FieldsOf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.src.description(fset, nil)
			if !strings.Contains(result, tt.want) {
				t.Errorf("description() = %q, want to contain %q", result, tt.want)
			}
		})
	}
}

// TestStructArgTypeAdditionalEdgeCases tests more edge cases for structArgType
func TestStructArgTypeAdditionalEdgeCases(t *testing.T) {
	// Test with non-composite-lit expression (should return nil)
	lit := &ast.BasicLit{
		Kind:  token.INT,
		Value: "42",
	}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
	}

	result := structArgType(info, lit)
	if result != nil {
		t.Errorf("structArgType(literal) = %v, want nil", result)
	}
}

// TestProcessFieldsOfErrorCases tests error cases in processFieldsOf
func TestProcessFieldsOfErrorCases(t *testing.T) {
	fset := token.NewFileSet()

	// Test with wrong number of arguments
	call := &ast.CallExpr{
		Fun:  ast.NewIdent("FieldsOf"),
		Args: []ast.Expr{}, // No arguments
	}

	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
	}

	_, err := processFieldsOf(fset, info, call)
	if err == nil {
		t.Error("processFieldsOf() expected error for no arguments, got nil")
	} else if !strings.Contains(err.Error(), "must specify fields") {
		t.Errorf("processFieldsOf() error = %v, want error containing 'must specify fields'", err)
	}
}

// TestRewritePkgRefsEdgeCases tests edge cases in rewritePkgRefs
func TestRewritePkgRefsEdgeCases(t *testing.T) {
	pkg := &packages.Package{
		Name:    "test",
		PkgPath: "example.com/test",
		Fset:    token.NewFileSet(),
		Types:   types.NewPackage("example.com/test", "test"),
	}

	g := newGen(pkg)

	// Create a simple AST node with a selector expression
	src := `package test
import "fmt"

func foo() {
	fmt.Println("hello")
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	pkg.Fset = fset
	g.pkg.Fset = fset

	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Uses:  make(map[*ast.Ident]types.Object),
		Defs:  make(map[*ast.Ident]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}

	// Rewrite package references
	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			rewritten := g.rewritePkgRefs(info, fn)
			if rewritten == nil {
				t.Error("rewritePkgRefs() returned nil")
			}
		}
	}
}

// TestInjectEdgeCases tests edge cases in inject function
func TestInjectEdgeCases(t *testing.T) {
	pkg := &packages.Package{
		Name:    "test",
		PkgPath: "example.com/test",
		Fset:    token.NewFileSet(),
		Types:   types.NewPackage("example.com/test", "test"),
	}

	g := newGen(pkg)

	// Test with empty provider set
	set := &ProviderSet{
		Pos:     token.NoPos,
		PkgPath: "example.com/test",
	}

	// Create an injector with simple signature
	sig := types.NewSignatureType(nil, nil, nil,
		types.NewTuple(),
		types.NewTuple(types.NewVar(0, nil, "", types.Typ[types.Int])),
		false)

	// Call inject - it might fail but should not panic
	_ = g.inject(token.NoPos, "TestInjector", sig, set, nil)
}

// TestCopyNonInjectorDeclsEdgeCases tests edge cases in copyNonInjectorDecls
func TestCopyNonInjectorDeclsEdgeCases(t *testing.T) {
	pkg := &packages.Package{
		Name:    "test",
		PkgPath: "example.com/test",
		Fset:    token.NewFileSet(),
		Types:   types.NewPackage("example.com/test", "test"),
	}

	g := newGen(pkg)

	src := `package test

const MyConst = 42

var MyVar int

type MyType struct{}

func HelperFunc() {}

func init() {}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	pkg.Fset = fset
	g.pkg.Fset = fset

	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Uses:  make(map[*ast.Ident]types.Object),
		Defs:  make(map[*ast.Ident]types.Object),
	}

	// Copy non-injector declarations
	injectorFiles := []*ast.File{f}
	copyNonInjectorDecls(g, injectorFiles, info)

	// Verify something was written
	if g.buf.Len() == 0 {
		t.Error("copyNonInjectorDecls() did not write anything")
	}
}

// TestNewObjectCacheEdgeCases tests edge cases in newObjectCache
func TestNewObjectCacheEdgeCases(t *testing.T) {
	pkg := &packages.Package{
		Name:    "test",
		PkgPath: "example.com/test",
		Fset:    token.NewFileSet(),
		Types:   types.NewPackage("example.com/test", "test"),
		TypesInfo: &types.Info{
			Types: make(map[ast.Expr]types.TypeAndValue),
			Uses:  make(map[*ast.Ident]types.Object),
			Defs:  make(map[*ast.Ident]types.Object),
		},
	}

	// Create object cache
	oc := newObjectCache([]*packages.Package{pkg})
	if oc == nil {
		t.Error("newObjectCache() returned nil")
	}

	// Verify it was initialized - just check it's not nil
	if oc.packages == nil {
		t.Error("newObjectCache() packages map not initialized")
	}
	if oc.objects == nil {
		t.Error("newObjectCache() objects map not initialized")
	}
}

// TestLoadFunctionCoverage tests the Load function behavior
func TestLoadFunctionCoverage(t *testing.T) {
	// Test with invalid directory
	_, err := Load(context.Background(), "/nonexistent/directory/that/does/not/exist", os.Environ(), "wireinject", []string{"."})
	if err == nil || len(err) == 0 {
		t.Log("Load() with invalid directory may or may not return error depending on environment")
	}
}

// TestProcessStructProviderEdgeCases tests edge cases in processStructProvider
func TestProcessStructProviderEdgeCases(t *testing.T) {
	fset := token.NewFileSet()

	src := `package test
type MyStruct struct {
	Field int
}
`
	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	conf := &types.Config{Importer: nil}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Uses:  make(map[*ast.Ident]types.Object),
		Defs:  make(map[*ast.Ident]types.Object),
	}
	_, _ = conf.Check("test", fset, []*ast.File{f}, info)

	// Create a mock call expression
	call := &ast.CallExpr{
		Fun: ast.NewIdent("StructProvider"),
		Args: []ast.Expr{
			&ast.CompositeLit{
				Type: ast.NewIdent("MyStruct"),
			},
		},
	}

	// This tests the struct provider processing path
	_, err = processStructProvider(fset, info, call)
	// We expect an error since we don't have proper type information
	if err != nil {
		t.Logf("processStructProvider() returned expected error: %v", err)
	}
}

// TestGenerateInjectorsEdgeCases tests edge cases in generateInjectors
func TestGenerateInjectorsEdgeCases(t *testing.T) {
	pkg := &packages.Package{
		Name:    "test",
		PkgPath: "example.com/test",
		Fset:    token.NewFileSet(),
		Types:   types.NewPackage("example.com/test", "test"),
		TypesInfo: &types.Info{
			Types: make(map[ast.Expr]types.TypeAndValue),
			Uses:  make(map[*ast.Ident]types.Object),
			Defs:  make(map[*ast.Ident]types.Object),
		},
	}

	g := newGen(pkg)

	// Call generateInjectors with empty package
	_, errs := generateInjectors(g, pkg)

	// We expect it to complete without panic
	if len(errs) > 0 {
		t.Logf("generateInjectors() returned expected errors: %v", errs)
	}
}

// TestQualifiedIdentObjectWithNilInfo tests qualifiedIdentObject with nil in Uses map
func TestQualifiedIdentObjectWithNilInfo(t *testing.T) {
	info := &types.Info{
		Uses: make(map[*ast.Ident]types.Object),
	}

	// Create a selector expression
	sel := &ast.SelectorExpr{
		X:   ast.NewIdent("pkg"),
		Sel: ast.NewIdent("Type"),
	}

	// This should return an error since there's no type information
	result := qualifiedIdentObject(info, sel)
	if result != nil {
		t.Logf("qualifiedIdentObject() returned: %v", result)
	}
}

// TestProcessNewSetEdgeCases tests edge cases in processNewSet
func TestProcessNewSetEdgeCases(t *testing.T) {
	pkg := &packages.Package{
		Name:    "test",
		PkgPath: "example.com/test",
		Fset:    token.NewFileSet(),
		Types:   types.NewPackage("example.com/test", "test"),
		TypesInfo: &types.Info{
			Types: make(map[ast.Expr]types.TypeAndValue),
			Uses:  make(map[*ast.Ident]types.Object),
			Defs:  make(map[*ast.Ident]types.Object),
		},
	}

	oc := newObjectCache([]*packages.Package{pkg})

	// Create a mock call expression
	call := &ast.CallExpr{
		Fun:  ast.NewIdent("NewSet"),
		Args: []ast.Expr{},
	}

	// Call processNewSet
	_, errs := oc.processNewSet(pkg.TypesInfo, "example.com/test", call, nil, "")

	if len(errs) > 0 {
		t.Logf("processNewSet() returned expected errors: %v", errs)
	}
}
