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
	"testing"
)

func TestCopyAST(t *testing.T) {
	t.Run("BasicLit", func(t *testing.T) {
		orig := &ast.BasicLit{
			ValuePos: 10,
			Kind:     token.INT,
			Value:    "42",
		}
		copy := copyAST(orig).(*ast.BasicLit)

		if copy.Value != orig.Value {
			t.Errorf("Value = %q; want %q", copy.Value, orig.Value)
		}
		if copy.Kind != orig.Kind {
			t.Errorf("Kind = %v; want %v", copy.Kind, orig.Kind)
		}
		if copy == orig {
			t.Error("copy should be a different pointer than original")
		}

		// Verify memory independence.
		copy.Value = "99"
		if orig.Value == "99" {
			t.Error("modifying copy affected original")
		}
	})

	t.Run("Ident preserves identity", func(t *testing.T) {
		orig := &ast.Ident{Name: "foo"}
		copy := copyAST(orig).(*ast.Ident)

		// Idents preserve identity for *types.Info compatibility.
		if copy != orig {
			t.Error("Ident copy should be the same pointer as original")
		}
	})

	t.Run("BinaryExpr", func(t *testing.T) {
		x := &ast.BasicLit{Kind: token.INT, Value: "1"}
		y := &ast.BasicLit{Kind: token.INT, Value: "2"}
		orig := &ast.BinaryExpr{
			X:  x,
			Op: token.ADD,
			Y:  y,
		}
		copy := copyAST(orig).(*ast.BinaryExpr)

		if copy == orig {
			t.Error("copy should be a different pointer")
		}
		if copy.Op != orig.Op {
			t.Errorf("Op = %v; want %v", copy.Op, orig.Op)
		}

		// Children should be copies too.
		copyX := copy.X.(*ast.BasicLit)
		if copyX == x {
			t.Error("X child should be a copy")
		}
		if copyX.Value != "1" {
			t.Errorf("X.Value = %q; want %q", copyX.Value, "1")
		}

		copyY := copy.Y.(*ast.BasicLit)
		if copyY == y {
			t.Error("Y child should be a copy")
		}
		if copyY.Value != "2" {
			t.Errorf("Y.Value = %q; want %q", copyY.Value, "2")
		}

		// Verify memory independence.
		copyX.Value = "99"
		if x.Value == "99" {
			t.Error("modifying copy.X affected original")
		}
	})

	t.Run("CallExpr", func(t *testing.T) {
		fun := &ast.Ident{Name: "myFunc"}
		arg1 := &ast.BasicLit{Kind: token.STRING, Value: `"hello"`}
		arg2 := &ast.BasicLit{Kind: token.INT, Value: "42"}
		orig := &ast.CallExpr{
			Fun:  fun,
			Args: []ast.Expr{arg1, arg2},
		}
		copy := copyAST(orig).(*ast.CallExpr)

		if copy == orig {
			t.Error("copy should be a different pointer")
		}
		if len(copy.Args) != 2 {
			t.Fatalf("got %d args; want 2", len(copy.Args))
		}

		// Fun should be the same Ident (identity preserved).
		if copy.Fun != fun {
			t.Error("Fun Ident should preserve identity")
		}

		// Args should be copies.
		copyArg1 := copy.Args[0].(*ast.BasicLit)
		if copyArg1 == arg1 {
			t.Error("arg1 should be a copy")
		}
		if copyArg1.Value != `"hello"` {
			t.Errorf("arg1.Value = %q; want %q", copyArg1.Value, `"hello"`)
		}
	})

	t.Run("CompositeLit", func(t *testing.T) {
		elt1 := &ast.BasicLit{Kind: token.INT, Value: "1"}
		elt2 := &ast.BasicLit{Kind: token.INT, Value: "2"}
		typIdent := &ast.Ident{Name: "MyStruct"}
		orig := &ast.CompositeLit{
			Type: typIdent,
			Elts: []ast.Expr{elt1, elt2},
		}
		copy := copyAST(orig).(*ast.CompositeLit)

		if copy == orig {
			t.Error("copy should be a different pointer")
		}
		if len(copy.Elts) != 2 {
			t.Fatalf("got %d elts; want 2", len(copy.Elts))
		}

		// Type is an Ident, so identity is preserved.
		if copy.Type != typIdent {
			t.Error("Type Ident should preserve identity")
		}

		// Elements should be copies.
		copyElt1 := copy.Elts[0].(*ast.BasicLit)
		if copyElt1 == elt1 {
			t.Error("elt1 should be a copy")
		}

		// Verify memory independence.
		copyElt1.Value = "99"
		if elt1.Value == "99" {
			t.Error("modifying copy element affected original")
		}
	})

	t.Run("FuncDecl", func(t *testing.T) {
		name := &ast.Ident{Name: "myFunc"}
		body := &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.BasicLit{Kind: token.INT, Value: "0"},
					},
				},
			},
		}
		orig := &ast.FuncDecl{
			Name: name,
			Type: &ast.FuncType{},
			Body: body,
		}
		copy := copyAST(orig).(*ast.FuncDecl)

		if copy == orig {
			t.Error("copy should be a different pointer")
		}

		// Name Ident preserves identity.
		if copy.Name != name {
			t.Error("Name Ident should preserve identity")
		}

		// Body should be a copy.
		if copy.Body == body {
			t.Error("Body should be a copy")
		}
		if len(copy.Body.List) != 1 {
			t.Fatalf("got %d stmts; want 1", len(copy.Body.List))
		}

		// Verify deep copy of body.
		ret := copy.Body.List[0].(*ast.ReturnStmt)
		origRet := body.List[0].(*ast.ReturnStmt)
		if ret == origRet {
			t.Error("return stmt should be a copy")
		}
	})

	t.Run("UnaryExpr", func(t *testing.T) {
		x := &ast.BasicLit{Kind: token.INT, Value: "5"}
		orig := &ast.UnaryExpr{
			Op: token.SUB,
			X:  x,
		}
		copy := copyAST(orig).(*ast.UnaryExpr)

		if copy == orig {
			t.Error("copy should be a different pointer")
		}
		if copy.Op != token.SUB {
			t.Errorf("Op = %v; want %v", copy.Op, token.SUB)
		}
		copyX := copy.X.(*ast.BasicLit)
		if copyX == x {
			t.Error("X should be a copy")
		}
	})

	t.Run("StarExpr", func(t *testing.T) {
		inner := &ast.Ident{Name: "int"}
		orig := &ast.StarExpr{X: inner}
		copy := copyAST(orig).(*ast.StarExpr)

		if copy == orig {
			t.Error("copy should be a different pointer")
		}
		// Inner Ident preserves identity.
		if copy.X != inner {
			t.Error("X Ident should preserve identity")
		}
	})

	t.Run("SelectorExpr", func(t *testing.T) {
		x := &ast.Ident{Name: "pkg"}
		sel := &ast.Ident{Name: "Func"}
		orig := &ast.SelectorExpr{X: x, Sel: sel}
		copy := copyAST(orig).(*ast.SelectorExpr)

		if copy == orig {
			t.Error("copy should be a different pointer")
		}
		// Both X and Sel are Idents, so identity is preserved.
		if copy.X != x {
			t.Error("X Ident should preserve identity")
		}
		if copy.Sel != sel {
			t.Error("Sel Ident should preserve identity")
		}
	})

	t.Run("KeyValueExpr", func(t *testing.T) {
		key := &ast.Ident{Name: "key"}
		value := &ast.BasicLit{Kind: token.STRING, Value: `"val"`}
		orig := &ast.KeyValueExpr{Key: key, Value: value}
		copy := copyAST(orig).(*ast.KeyValueExpr)

		if copy == orig {
			t.Error("copy should be a different pointer")
		}
		copyVal := copy.Value.(*ast.BasicLit)
		if copyVal == value {
			t.Error("Value should be a copy")
		}
		if copyVal.Value != `"val"` {
			t.Errorf("Value = %q; want %q", copyVal.Value, `"val"`)
		}
	})
}
