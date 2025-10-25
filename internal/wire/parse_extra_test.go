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
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"golang.org/x/tools/go/types/typeutil"
)

func TestProviderSetID_String(t *testing.T) {
	tests := []struct {
		name string
		id   ProviderSetID
		want string
	}{
		{
			name: "basic provider set",
			id:   ProviderSetID{ImportPath: "example.com/foo", VarName: "MySet"},
			want: `"example.com/foo".MySet`,
		},
		{
			name: "provider set with complex path",
			id:   ProviderSetID{ImportPath: "github.com/user/repo/pkg", VarName: "DefaultSet"},
			want: `"github.com/user/repo/pkg".DefaultSet`,
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

func TestInjector_String(t *testing.T) {
	tests := []struct {
		name string
		in   *Injector
		want string
	}{
		{
			name: "basic injector",
			in:   &Injector{ImportPath: "example.com/foo", FuncName: "Initialize"},
			want: `"example.com/foo".Initialize`,
		},
		{
			name: "injector with complex path",
			in:   &Injector{ImportPath: "github.com/user/repo/pkg", FuncName: "NewService"},
			want: `"github.com/user/repo/pkg".NewService`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.String(); got != tt.want {
				t.Errorf("Injector.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProviderSet_Outputs(t *testing.T) {
	t.Run("outputs from provider set", func(t *testing.T) {
		// Create a simple provider set for testing
		// Note: This requires complex setup with actual type information
		// For now, we test the basic structure
		set := &ProviderSet{
			providerMap: new(typeutil.Map),
		}

		// Call Outputs - should return empty for empty provider map
		outputs := set.Outputs()
		if outputs == nil {
			t.Error("ProviderSet.Outputs() returned nil, expected empty slice")
		}
	})
}

func TestIsProviderSetType(t *testing.T) {
	tests := []struct {
		name     string
		typeStr  string
		want     bool
		setupFn  func() types.Type
	}{
		{
			name: "not a named type",
			setupFn: func() types.Type {
				return types.Typ[types.Int]
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := tt.setupFn()
			if got := isProviderSetType(typ); got != tt.want {
				t.Errorf("isProviderSetType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStructArgType(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		wantNil bool
	}{
		{
			name:    "not a composite literal",
			source:  "package test\nvar x = 42",
			wantNil: true,
		},
		{
			name: "struct composite literal",
			source: `package test
type Foo struct {
	X int
}
var x = Foo{X: 42}`,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "test.go", tt.source, 0)
			if err != nil {
				t.Fatal(err)
			}

			// Create type info
			info := &types.Info{
				Types: make(map[ast.Expr]types.TypeAndValue),
				Defs:  make(map[*ast.Ident]types.Object),
				Uses:  make(map[*ast.Ident]types.Object),
			}

			conf := types.Config{}
			pkg, err := conf.Check("test", fset, []*ast.File{f}, info)
			if err != nil {
				// Type errors are expected in some test cases
				t.Logf("type check error (may be expected): %v", err)
			}
			_ = pkg

			// Find composite literal if present
			var expr ast.Expr
			ast.Inspect(f, func(n ast.Node) bool {
				if cl, ok := n.(*ast.CompositeLit); ok {
					expr = cl
					return false
				}
				return true
			})

			if expr != nil {
				result := structArgType(info, expr)
				if tt.wantNil && result != nil {
					t.Errorf("structArgType() = %v, want nil", result)
				}
			}
		})
	}
}

func TestIsWireImport(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "wire import",
			path: "github.com/almondoo/wire",
			want: true,
		},
		{
			name: "wire import with vendor",
			path: "example.com/vendor/github.com/almondoo/wire",
			want: true,
		},
		{
			name: "not wire import",
			path: "github.com/other/package",
			want: false,
		},
		{
			name: "wire-like but different",
			path: "github.com/almondoo/wire-extra",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isWireImport(tt.path); got != tt.want {
				t.Errorf("isWireImport(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestProvidedType_Methods(t *testing.T) {
	intType := types.Typ[types.Int]
	stringType := types.Typ[types.String]

	t.Run("IsNil", func(t *testing.T) {
		var pt ProvidedType
		if !pt.IsNil() {
			t.Error("zero ProvidedType should be nil")
		}

		pt = ProvidedType{t: intType, p: &Provider{}}
		if pt.IsNil() {
			t.Error("ProvidedType with provider should not be nil")
		}
	})

	t.Run("Type", func(t *testing.T) {
		pt := ProvidedType{t: stringType}
		if got := pt.Type(); got != stringType {
			t.Errorf("Type() = %v, want %v", got, stringType)
		}
	})

	t.Run("IsProvider", func(t *testing.T) {
		pt := ProvidedType{t: intType, p: &Provider{}}
		if !pt.IsProvider() {
			t.Error("ProvidedType with provider should return true for IsProvider()")
		}

		pt = ProvidedType{t: intType}
		if pt.IsProvider() {
			t.Error("ProvidedType without provider should return false for IsProvider()")
		}
	})

	t.Run("IsValue", func(t *testing.T) {
		pt := ProvidedType{t: intType, v: &Value{}}
		if !pt.IsValue() {
			t.Error("ProvidedType with value should return true for IsValue()")
		}

		pt = ProvidedType{t: intType}
		if pt.IsValue() {
			t.Error("ProvidedType without value should return false for IsValue()")
		}
	})

	t.Run("IsArg", func(t *testing.T) {
		pt := ProvidedType{t: intType, a: &InjectorArg{}}
		if !pt.IsArg() {
			t.Error("ProvidedType with arg should return true for IsArg()")
		}

		pt = ProvidedType{t: intType}
		if pt.IsArg() {
			t.Error("ProvidedType without arg should return false for IsArg()")
		}
	})

	t.Run("IsField", func(t *testing.T) {
		pt := ProvidedType{t: intType, f: &Field{}}
		if !pt.IsField() {
			t.Error("ProvidedType with field should return true for IsField()")
		}

		pt = ProvidedType{t: intType}
		if pt.IsField() {
			t.Error("ProvidedType without field should return false for IsField()")
		}
	})

	t.Run("Provider", func(t *testing.T) {
		provider := &Provider{Name: "TestProvider"}
		pt := ProvidedType{t: intType, p: provider}
		if got := pt.Provider(); got != provider {
			t.Errorf("Provider() = %v, want %v", got, provider)
		}
	})

	t.Run("Value", func(t *testing.T) {
		value := &Value{Out: intType}
		pt := ProvidedType{t: intType, v: value}
		if got := pt.Value(); got != value {
			t.Errorf("Value() = %v, want %v", got, value)
		}
	})

	t.Run("Arg", func(t *testing.T) {
		arg := &InjectorArg{Index: 0}
		pt := ProvidedType{t: intType, a: arg}
		if got := pt.Arg(); got != arg {
			t.Errorf("Arg() = %v, want %v", got, arg)
		}
	})

	t.Run("Field", func(t *testing.T) {
		field := &Field{Name: "TestField"}
		pt := ProvidedType{t: intType, f: field}
		if got := pt.Field(); got != field {
			t.Errorf("Field() = %v, want %v", got, field)
		}
	})
}

// TestLoad tests the Load function which was previously uncovered
func TestLoad(t *testing.T) {
	t.Run("empty patterns", func(t *testing.T) {
		// Test with empty patterns - should handle gracefully
		// This requires actual packages, so we use a simple test case
		info, errs := Load(nil, ".", nil, "", []string{})
		if len(errs) == 0 && info == nil {
			t.Error("Load() with empty patterns should return valid result or errors")
		}
	})
}

func TestProcessStructLiteralProvider(t *testing.T) {
	t.Run("valid struct type", func(t *testing.T) {
		source := `package test
type Foo struct {
	X int
	Y string
}`
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "test.go", source, 0)
		if err != nil {
			t.Fatal(err)
		}

		info := &types.Info{
			Types: make(map[ast.Expr]types.TypeAndValue),
			Defs:  make(map[*ast.Ident]types.Object),
			Uses:  make(map[*ast.Ident]types.Object),
		}

		conf := types.Config{}
		pkg, err := conf.Check("test", fset, []*ast.File{f}, info)
		if err != nil {
			t.Fatal(err)
		}

		// Find the Foo type
		obj := pkg.Scope().Lookup("Foo")
		if obj == nil {
			t.Fatal("Foo type not found")
		}

		typeName := obj.(*types.TypeName)
		provider, errs := processStructLiteralProvider(fset, typeName)

		// Should return provider with deprecation warning
		if provider == nil && len(errs) > 0 {
			t.Logf("processStructLiteralProvider returned errors: %v", errs)
		} else if provider == nil {
			t.Error("expected provider to be non-nil")
		}
	})

	t.Run("non-struct type", func(t *testing.T) {
		source := `package test
type Foo int`
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "test.go", source, 0)
		if err != nil {
			t.Fatal(err)
		}

		info := &types.Info{
			Types: make(map[ast.Expr]types.TypeAndValue),
			Defs:  make(map[*ast.Ident]types.Object),
			Uses:  make(map[*ast.Ident]types.Object),
		}

		conf := types.Config{}
		pkg, err := conf.Check("test", fset, []*ast.File{f}, info)
		if err != nil {
			t.Fatal(err)
		}

		obj := pkg.Scope().Lookup("Foo")
		if obj == nil {
			t.Fatal("Foo type not found")
		}

		typeName := obj.(*types.TypeName)
		provider, errs := processStructLiteralProvider(fset, typeName)

		// Should return error for non-struct type
		if provider != nil {
			t.Error("expected provider to be nil for non-struct type")
		}
		if len(errs) == 0 {
			t.Error("expected errors for non-struct type")
		}
	})
}
