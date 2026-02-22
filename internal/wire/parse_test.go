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
	"go/token"
	"go/types"
	"testing"
)

func TestFuncOutput(t *testing.T) {
	intT := types.Typ[types.Int]
	stringT := types.Typ[types.String]

	tests := []struct {
		name        string
		sig         *types.Signature
		wantOut     types.Type
		wantCleanup bool
		wantErr     bool
		wantError   string
	}{
		{
			name:    "single return",
			sig:     makeSig(nil, []types.Type{intT}),
			wantOut: intT,
		},
		{
			name:    "return with error",
			sig:     makeSig(nil, []types.Type{intT, testErrorType}),
			wantOut: intT,
			wantErr: true,
		},
		{
			name:        "return with cleanup",
			sig:         makeSig(nil, []types.Type{intT, testCleanupType}),
			wantOut:     intT,
			wantCleanup: true,
		},
		{
			name:        "return with cleanup and error",
			sig:         makeSig(nil, []types.Type{intT, testCleanupType, testErrorType}),
			wantOut:     intT,
			wantCleanup: true,
			wantErr:     true,
		},
		{
			name:      "no return values",
			sig:       makeSig(nil, nil),
			wantError: "no return values",
		},
		{
			name:      "too many return values",
			sig:       makeSig(nil, []types.Type{intT, testCleanupType, testErrorType, stringT}),
			wantError: "too many return values",
		},
		{
			name:      "second return is string",
			sig:       makeSig(nil, []types.Type{intT, stringT}),
			wantError: "second return type",
		},
		{
			name:      "third return is string",
			sig:       makeSig(nil, []types.Type{intT, testCleanupType, stringT}),
			wantError: "third return type",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out, err := funcOutput(test.sig)
			if test.wantError != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", test.wantError)
				}
				assertErrorContains(t, []error{err}, test.wantError)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !types.Identical(out.out, test.wantOut) {
				t.Errorf("out type = %v; want %v", out.out, test.wantOut)
			}
			if out.cleanup != test.wantCleanup {
				t.Errorf("cleanup = %v; want %v", out.cleanup, test.wantCleanup)
			}
			if out.err != test.wantErr {
				t.Errorf("err = %v; want %v", out.err, test.wantErr)
			}
		})
	}
}

func TestProcessFuncProvider(t *testing.T) {
	fset := token.NewFileSet()
	pkg := testPkg("example.com/test", "test")
	intT := types.Typ[types.Int]
	stringT := types.Typ[types.String]
	boolT := types.Typ[types.Bool]

	tests := []struct {
		name      string
		fn        *types.Func
		wantArgs  int
		wantOut   types.Type
		wantErr   string
	}{
		{
			name:    "no params, single return",
			fn:      makeFunc(pkg, "NewFoo", makeSig(nil, []types.Type{intT})),
			wantOut: intT,
		},
		{
			name:     "two different params",
			fn:       makeFunc(pkg, "NewBar", makeSig([]types.Type{intT, stringT}, []types.Type{boolT})),
			wantArgs: 2,
			wantOut:  boolT,
		},
		{
			name:    "duplicate params",
			fn:      makeFunc(pkg, "NewBaz", makeSig([]types.Type{intT, intT}, []types.Type{stringT})),
			wantErr: "multiple parameters of type",
		},
		{
			name: "variadic provider",
			fn: makeFunc(pkg, "NewVariadic", makeVariadicSig(
				[]types.Type{types.NewSlice(intT)},
				[]types.Type{stringT},
			)),
			wantArgs: 1,
			wantOut:  stringT,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p, errs := processFuncProvider(fset, test.fn)
			if test.wantErr != "" {
				assertErrorContains(t, errs, test.wantErr)
				return
			}
			assertNoErrors(t, errs)
			if len(p.Args) != test.wantArgs {
				t.Errorf("got %d args; want %d", len(p.Args), test.wantArgs)
			}
			if !types.Identical(p.Out[0], test.wantOut) {
				t.Errorf("out = %v; want %v", p.Out[0], test.wantOut)
			}
		})
	}
}

func TestInjectorFuncSignature(t *testing.T) {
	intT := types.Typ[types.Int]

	tests := []struct {
		name    string
		sig     *types.Signature
		wantErr string
	}{
		{
			name: "valid single return",
			sig:  makeSig([]types.Type{intT}, []types.Type{intT}),
		},
		{
			name: "valid with error",
			sig:  makeSig(nil, []types.Type{intT, testErrorType}),
		},
		{
			name:    "no return values",
			sig:     makeSig(nil, nil),
			wantErr: "no return values",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := injectorFuncSignature(test.sig)
			if test.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", test.wantErr)
				}
				assertErrorContains(t, []error{err}, test.wantErr)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestIsWireImport(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"github.com/almondoo/wire", true},
		{"github.com/google/wire", false},
		{"example.com/foo", false},
		{"vendor/github.com/almondoo/wire", true},
		{"some/vendor/github.com/almondoo/wire", true},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			got := isWireImport(test.path)
			if got != test.want {
				t.Errorf("isWireImport(%q) = %v; want %v", test.path, got, test.want)
			}
		})
	}
}

func TestIsProviderSetType(t *testing.T) {
	t.Run("non-named type returns false", func(t *testing.T) {
		if isProviderSetType(types.Typ[types.Int]) {
			t.Error("expected false for basic int type")
		}
	})

	t.Run("named type in wrong package returns false", func(t *testing.T) {
		pkg := testPkg("example.com/foo", "foo")
		tn := types.NewTypeName(token.NoPos, pkg, "ProviderSet", nil)
		named := types.NewNamed(tn, types.Typ[types.Int], nil)
		if isProviderSetType(named) {
			t.Error("expected false for ProviderSet in wrong package")
		}
	})
}

func TestCheckField(t *testing.T) {
	// Build a struct type with fields.
	pkg := testPkg("example.com/test", "test")
	f1 := types.NewField(token.NoPos, pkg, "Name", types.Typ[types.String], false)
	f2 := types.NewField(token.NoPos, pkg, "Age", types.Typ[types.Int], false)
	st := types.NewStruct([]*types.Var{f1, f2}, []string{"", `wire:"-"`})

	t.Run("valid field", func(t *testing.T) {
		expr := &ast.BasicLit{Kind: token.STRING, Value: `"Name"`}
		v, err := checkField(expr, st)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Name() != "Name" {
			t.Errorf("field name = %q; want %q", v.Name(), "Name")
		}
	})

	t.Run("prevented field", func(t *testing.T) {
		expr := &ast.BasicLit{Kind: token.STRING, Value: `"Age"`}
		_, err := checkField(expr, st)
		if err == nil {
			t.Fatal("expected error for prevented field")
		}
		assertErrorContains(t, []error{err}, "prevented")
	})

	t.Run("non-existent field", func(t *testing.T) {
		expr := &ast.BasicLit{Kind: token.STRING, Value: `"Missing"`}
		_, err := checkField(expr, st)
		if err == nil {
			t.Fatal("expected error for non-existent field")
		}
		assertErrorContains(t, []error{err}, "is not a field")
	})

	t.Run("non-string expression", func(t *testing.T) {
		expr := &ast.Ident{Name: "x"}
		_, err := checkField(expr, st)
		if err == nil {
			t.Fatal("expected error for non-string expression")
		}
		assertErrorContains(t, []error{err}, "must be a string")
	})
}

func TestIsPrevented(t *testing.T) {
	tests := []struct {
		tag  string
		want bool
	}{
		{"", false},
		{`json:"name"`, false},
		{`wire:"-"`, true},
		{`json:"name" wire:"-"`, true},
		{`wire:"inject"`, false},
	}
	for _, test := range tests {
		t.Run(test.tag, func(t *testing.T) {
			got := isPrevented(test.tag)
			if got != test.want {
				t.Errorf("isPrevented(%q) = %v; want %v", test.tag, got, test.want)
			}
		})
	}
}

func TestAllFields(t *testing.T) {
	t.Run("wildcard", func(t *testing.T) {
		call := &ast.CallExpr{
			Args: []ast.Expr{
				&ast.Ident{Name: "dummy"},
				&ast.BasicLit{Kind: token.STRING, Value: `"*"`},
			},
		}
		if !allFields(call) {
			t.Error("expected true for wildcard")
		}
	})

	t.Run("not wildcard", func(t *testing.T) {
		call := &ast.CallExpr{
			Args: []ast.Expr{
				&ast.Ident{Name: "dummy"},
				&ast.BasicLit{Kind: token.STRING, Value: `"Name"`},
			},
		}
		if allFields(call) {
			t.Error("expected false for non-wildcard")
		}
	})

	t.Run("too many args", func(t *testing.T) {
		call := &ast.CallExpr{
			Args: []ast.Expr{
				&ast.Ident{Name: "dummy"},
				&ast.BasicLit{Kind: token.STRING, Value: `"*"`},
				&ast.BasicLit{Kind: token.STRING, Value: `"extra"`},
			},
		}
		if allFields(call) {
			t.Error("expected false for too many args")
		}
	})
}

func TestQualifiedIdentObject(t *testing.T) {
	pkg := testPkg("example.com/test", "test")
	obj := types.NewVar(token.NoPos, pkg, "Foo", types.Typ[types.Int])

	t.Run("simple ident", func(t *testing.T) {
		ident := &ast.Ident{Name: "Foo"}
		info := &types.Info{
			Uses: map[*ast.Ident]types.Object{ident: obj},
		}
		got := qualifiedIdentObject(info, ident)
		if got != obj {
			t.Errorf("got %v; want %v", got, obj)
		}
	})

	t.Run("selector expr", func(t *testing.T) {
		pkgName := types.NewPkgName(token.NoPos, pkg, "test", pkg)
		xIdent := &ast.Ident{Name: "test"}
		selIdent := &ast.Ident{Name: "Foo"}
		info := &types.Info{
			Uses: map[*ast.Ident]types.Object{
				xIdent:  pkgName,
				selIdent: obj,
			},
		}
		expr := &ast.SelectorExpr{X: xIdent, Sel: selIdent}
		got := qualifiedIdentObject(info, expr)
		if got != obj {
			t.Errorf("got %v; want %v", got, obj)
		}
	})

	t.Run("non-ident expr", func(t *testing.T) {
		info := &types.Info{}
		expr := &ast.BasicLit{Kind: token.INT, Value: "42"}
		got := qualifiedIdentObject(info, expr)
		if got != nil {
			t.Errorf("got %v; want nil", got)
		}
	})
}
