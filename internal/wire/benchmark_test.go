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
	"testing"

	"golang.org/x/tools/go/types/typeutil"
)

func BenchmarkBuildProviderMap(b *testing.B) {
	pkg := testPkg("example.com/bench", "bench")

	for _, n := range []int{1, 5, 10, 25, 50} {
		b.Run(fmt.Sprintf("providers=%d", n), func(b *testing.B) {
			providers := make([]*Provider, n)
			for i := 0; i < n; i++ {
				outType := types.NewNamed(
					types.NewTypeName(token.NoPos, pkg, fmt.Sprintf("T%d", i), nil),
					types.Typ[types.Int],
					nil,
				)
				providers[i] = &Provider{
					Pkg:  pkg,
					Name: fmt.Sprintf("New%d", i),
					Pos:  token.NoPos,
					Out:  []types.Type{outType},
				}
			}

			fset := token.NewFileSet()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				hasher := typeutil.MakeHasher()
				pset := &ProviderSet{
					Pos:       token.NoPos,
					PkgPath:   pkg.Path(),
					Providers: providers,
				}
				buildProviderMap(fset, hasher, pset)
			}
		})
	}
}

func BenchmarkSolve(b *testing.B) {
	pkg := testPkg("example.com/bench", "bench")
	fset := token.NewFileSet()

	for _, depth := range []int{1, 5, 10, 20} {
		b.Run(fmt.Sprintf("depth=%d", depth), func(b *testing.B) {
			providers, outType := makeChain(pkg, depth)
			// Build the provider set once.
			hasher := typeutil.MakeHasher()
			pset := &ProviderSet{
				Pos:       token.NoPos,
				PkgPath:   pkg.Path(),
				Providers: providers,
			}
			var errs []error
			pset.providerMap, pset.srcMap, errs = buildProviderMap(fset, hasher, pset)
			if len(errs) > 0 {
				b.Fatalf("buildProviderMap failed: %v", errs)
			}

			given := vars()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				solve(fset, outType, given, pset)
			}
		})
	}
}

func BenchmarkVerifyAcyclic(b *testing.B) {
	pkg := testPkg("example.com/bench", "bench")

	for _, n := range []int{5, 10, 25, 50, 100} {
		b.Run(fmt.Sprintf("nodes=%d", n), func(b *testing.B) {
			hasher := typeutil.MakeHasher()
			pm := new(typeutil.Map)
			pm.SetHasher(hasher)

			nodeTypes := make([]types.Type, n)
			for i := 0; i < n; i++ {
				nodeTypes[i] = types.NewNamed(
					types.NewTypeName(token.NoPos, pkg, fmt.Sprintf("N%d", i), nil),
					types.Typ[types.Int],
					nil,
				)
			}

			// Build a linear chain (no cycles).
			for i := 0; i < n; i++ {
				var args []ProviderInput
				if i > 0 {
					args = []ProviderInput{{Type: nodeTypes[i-1]}}
				}
				p := &Provider{
					Pkg:  pkg,
					Name: fmt.Sprintf("New%d", i),
					Pos:  token.NoPos,
					Args: args,
					Out:  []types.Type{nodeTypes[i]},
				}
				pm.Set(nodeTypes[i], &ProvidedType{t: nodeTypes[i], p: p})
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				verifyAcyclic(pm, hasher)
			}
		})
	}
}

func BenchmarkDisambiguate(b *testing.B) {
	for _, n := range []int{10, 100} {
		b.Run(fmt.Sprintf("collisions=%d", n), func(b *testing.B) {
			collides := make(map[string]bool, n)
			collides["name"] = true
			for i := 2; i <= n; i++ {
				collides[fmt.Sprintf("name%d", i)] = true
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				disambiguate("name", func(s string) bool { return collides[s] })
			}
		})
	}
}

func BenchmarkCopyAST(b *testing.B) {
	b.Run("simple", func(b *testing.B) {
		node := &ast.BasicLit{Kind: token.INT, Value: "42"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			copyAST(node)
		}
	})

	b.Run("medium", func(b *testing.B) {
		node := &ast.BinaryExpr{
			X:  &ast.BasicLit{Kind: token.INT, Value: "1"},
			Op: token.ADD,
			Y: &ast.BinaryExpr{
				X:  &ast.BasicLit{Kind: token.INT, Value: "2"},
				Op: token.MUL,
				Y:  &ast.BasicLit{Kind: token.INT, Value: "3"},
			},
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			copyAST(node)
		}
	})

	b.Run("complex", func(b *testing.B) {
		// A function declaration with a body.
		stmts := make([]ast.Stmt, 10)
		for i := range stmts {
			stmts[i] = &ast.ExprStmt{
				X: &ast.CallExpr{
					Fun: &ast.Ident{Name: "foo"},
					Args: []ast.Expr{
						&ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", i)},
					},
				},
			}
		}
		node := &ast.FuncDecl{
			Name: &ast.Ident{Name: "myFunc"},
			Type: &ast.FuncType{},
			Body: &ast.BlockStmt{List: stmts},
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			copyAST(node)
		}
	})
}
