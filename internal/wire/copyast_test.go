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
	"testing"
)

func TestCopyASTHelpers(t *testing.T) {
	// Test helper functions that extract nodes from map
	t.Run("identFromMap", func(t *testing.T) {
		m := make(map[ast.Node]ast.Node)

		// Test with nil
		if got := identFromMap(m, nil); got != nil {
			t.Errorf("identFromMap(m, nil) = %v, want nil", got)
		}

		// Test with valid ident
		src := &ast.Ident{Name: "foo"}
		dst := &ast.Ident{Name: "bar"}
		m[src] = dst
		if got := identFromMap(m, src); got != dst {
			t.Errorf("identFromMap() = %v, want %v", got, dst)
		}
	})

	t.Run("blockStmtFromMap", func(t *testing.T) {
		m := make(map[ast.Node]ast.Node)

		// Test with nil
		if got := blockStmtFromMap(m, nil); got != nil {
			t.Errorf("blockStmtFromMap(m, nil) = %v, want nil", got)
		}

		// Test with valid block statement
		src := &ast.BlockStmt{List: []ast.Stmt{}}
		dst := &ast.BlockStmt{List: []ast.Stmt{}}
		m[src] = dst
		if got := blockStmtFromMap(m, src); got != dst {
			t.Errorf("blockStmtFromMap() = %v, want %v", got, dst)
		}
	})

	t.Run("callExprFromMap", func(t *testing.T) {
		m := make(map[ast.Node]ast.Node)

		// Test with nil
		if got := callExprFromMap(m, nil); got != nil {
			t.Errorf("callExprFromMap(m, nil) = %v, want nil", got)
		}

		// Test with valid call expression
		src := &ast.CallExpr{Fun: &ast.Ident{Name: "foo"}}
		dst := &ast.CallExpr{Fun: &ast.Ident{Name: "bar"}}
		m[src] = dst
		if got := callExprFromMap(m, src); got != dst {
			t.Errorf("callExprFromMap() = %v, want %v", got, dst)
		}
	})

	t.Run("basicLitFromMap", func(t *testing.T) {
		m := make(map[ast.Node]ast.Node)

		// Test with nil
		if got := basicLitFromMap(m, nil); got != nil {
			t.Errorf("basicLitFromMap(m, nil) = %v, want nil", got)
		}

		// Test with valid basic literal
		src := &ast.BasicLit{Kind: token.INT, Value: "42"}
		dst := &ast.BasicLit{Kind: token.INT, Value: "42"}
		m[src] = dst
		if got := basicLitFromMap(m, src); got != dst {
			t.Errorf("basicLitFromMap() = %v, want %v", got, dst)
		}
	})

	t.Run("funcTypeFromMap", func(t *testing.T) {
		m := make(map[ast.Node]ast.Node)

		// Test with nil
		if got := funcTypeFromMap(m, nil); got != nil {
			t.Errorf("funcTypeFromMap(m, nil) = %v, want nil", got)
		}

		// Test with valid function type
		src := &ast.FuncType{Params: &ast.FieldList{}}
		dst := &ast.FuncType{Params: &ast.FieldList{}}
		m[src] = dst
		if got := funcTypeFromMap(m, src); got != dst {
			t.Errorf("funcTypeFromMap() = %v, want %v", got, dst)
		}
	})

	t.Run("fieldListFromMap", func(t *testing.T) {
		m := make(map[ast.Node]ast.Node)

		// Test with nil
		if got := fieldListFromMap(m, nil); got != nil {
			t.Errorf("fieldListFromMap(m, nil) = %v, want nil", got)
		}

		// Test with valid field list
		src := &ast.FieldList{List: []*ast.Field{}}
		dst := &ast.FieldList{List: []*ast.Field{}}
		m[src] = dst
		if got := fieldListFromMap(m, src); got != dst {
			t.Errorf("fieldListFromMap() = %v, want %v", got, dst)
		}
	})
}

func TestCopyAST(t *testing.T) {
	t.Run("copy simple function", func(t *testing.T) {
		source := `package test
func foo() int {
	return 42
}
`
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "test.go", source, 0)
		if err != nil {
			t.Fatal(err)
		}

		// Find function declaration
		var fnDecl *ast.FuncDecl
		for _, decl := range f.Decls {
			if fd, ok := decl.(*ast.FuncDecl); ok {
				fnDecl = fd
				break
			}
		}

		if fnDecl == nil {
			t.Fatal("function declaration not found")
		}

		// Copy the AST
		copied := copyAST(fnDecl).(*ast.FuncDecl)

		// Verify the copy
		if copied == fnDecl {
			t.Error("copyAST() returned same pointer, expected new instance")
		}
		if copied.Name.Name != fnDecl.Name.Name {
			t.Errorf("copied function name = %v, want %v", copied.Name.Name, fnDecl.Name.Name)
		}
	})

	t.Run("copy with nil comments", func(t *testing.T) {
		// Test copyAST with various comment scenarios
		source := `package test
// This is a comment
func bar() {
}
`
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}

		var fnDecl *ast.FuncDecl
		for _, decl := range f.Decls {
			if fd, ok := decl.(*ast.FuncDecl); ok {
				fnDecl = fd
				break
			}
		}

		if fnDecl == nil {
			t.Fatal("function declaration not found")
		}

		copied := copyAST(fnDecl).(*ast.FuncDecl)

		if copied == fnDecl {
			t.Error("copyAST() returned same pointer, expected new instance")
		}
	})

	t.Run("copy expression list", func(t *testing.T) {
		// Test copyExprList
		m := make(map[ast.Node]ast.Node)

		// Test with nil
		if got := copyExprList(m, nil); got != nil {
			t.Errorf("copyExprList(m, nil) = %v, want nil", got)
		}

		// Test with expressions
		expr1 := &ast.Ident{Name: "x"}
		expr2 := &ast.Ident{Name: "y"}
		copy1 := &ast.Ident{Name: "x_copy"}
		copy2 := &ast.Ident{Name: "y_copy"}

		m[expr1] = copy1
		m[expr2] = copy2

		exprs := []ast.Expr{expr1, expr2}
		copied := copyExprList(m, exprs)

		if len(copied) != 2 {
			t.Errorf("copyExprList() returned %d expressions, want 2", len(copied))
		}
		if copied[0] != copy1 || copied[1] != copy2 {
			t.Error("copyExprList() did not copy expressions correctly")
		}
	})

	t.Run("copy statement list", func(t *testing.T) {
		// Test copyStmtList
		m := make(map[ast.Node]ast.Node)

		// Test with nil
		if got := copyStmtList(m, nil); got != nil {
			t.Errorf("copyStmtList(m, nil) = %v, want nil", got)
		}

		// Test with statements
		stmt1 := &ast.ExprStmt{X: &ast.Ident{Name: "x"}}
		stmt2 := &ast.ExprStmt{X: &ast.Ident{Name: "y"}}
		copy1 := &ast.ExprStmt{X: &ast.Ident{Name: "x_copy"}}
		copy2 := &ast.ExprStmt{X: &ast.Ident{Name: "y_copy"}}

		m[stmt1] = copy1
		m[stmt2] = copy2

		stmts := []ast.Stmt{stmt1, stmt2}
		copied := copyStmtList(m, stmts)

		if len(copied) != 2 {
			t.Errorf("copyStmtList() returned %d statements, want 2", len(copied))
		}
		if copied[0] != copy1 || copied[1] != copy2 {
			t.Error("copyStmtList() did not copy statements correctly")
		}
	})

	t.Run("copy identifier list", func(t *testing.T) {
		// Test copyIdentList
		m := make(map[ast.Node]ast.Node)

		// Test with nil
		if got := copyIdentList(m, nil); got != nil {
			t.Errorf("copyIdentList(m, nil) = %v, want nil", got)
		}

		// Test with identifiers
		ident1 := &ast.Ident{Name: "x"}
		ident2 := &ast.Ident{Name: "y"}
		copy1 := &ast.Ident{Name: "x_copy"}
		copy2 := &ast.Ident{Name: "y_copy"}

		m[ident1] = copy1
		m[ident2] = copy2

		idents := []*ast.Ident{ident1, ident2}
		copied := copyIdentList(m, idents)

		if len(copied) != 2 {
			t.Errorf("copyIdentList() returned %d identifiers, want 2", len(copied))
		}
		if copied[0] != copy1 || copied[1] != copy2 {
			t.Error("copyIdentList() did not copy identifiers correctly")
		}
	})
}
