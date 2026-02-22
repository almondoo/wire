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
	"go/token"
	"go/types"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/types/typeutil"
)

// Common types used across tests.
var (
	testErrorType   = types.Universe.Lookup("error").Type()
	testCleanupType = types.NewSignature(nil, nil, nil, false) // func()
)

// vars constructs a *types.Tuple from a list of types.
// Each type gets a synthetic variable name like "_p0", "_p1", etc.
func vars(ts ...types.Type) *types.Tuple {
	vs := make([]*types.Var, len(ts))
	for i, t := range ts {
		vs[i] = types.NewVar(token.NoPos, nil, "", t)
	}
	return types.NewTuple(vs...)
}

// namedVars constructs a *types.Tuple from alternating name/type pairs.
func namedVars(pairs ...interface{}) *types.Tuple {
	if len(pairs)%2 != 0 {
		panic("namedVars requires alternating name/type pairs")
	}
	vs := make([]*types.Var, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		name := pairs[i].(string)
		typ := pairs[i+1].(types.Type)
		vs[i/2] = types.NewVar(token.NoPos, nil, name, typ)
	}
	return types.NewTuple(vs...)
}

// makeSig constructs a *types.Signature from param and result type slices.
func makeSig(params, results []types.Type) *types.Signature {
	return types.NewSignature(nil, vars(params...), vars(results...), false)
}

// makeVariadicSig constructs a variadic *types.Signature.
func makeVariadicSig(params, results []types.Type) *types.Signature {
	return types.NewSignature(nil, vars(params...), vars(results...), true)
}

// makeFunc constructs a *types.Func with the given package, name, and signature.
func makeFunc(pkg *types.Package, name string, sig *types.Signature) *types.Func {
	return types.NewFunc(token.NoPos, pkg, name, sig)
}

// makeProvider constructs a *Provider for testing.
func makeProvider(pkg *types.Package, name string, args []ProviderInput, out []types.Type, opts ...providerOpt) *Provider {
	p := &Provider{
		Pkg:  pkg,
		Name: name,
		Pos:  token.NoPos,
		Args: args,
		Out:  out,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

type providerOpt func(*Provider)

func withCleanup() providerOpt {
	return func(p *Provider) { p.HasCleanup = true }
}

func withErr() providerOpt {
	return func(p *Provider) { p.HasErr = true }
}

func withStruct() providerOpt {
	return func(p *Provider) { p.IsStruct = true }
}

func withVarargs() providerOpt {
	return func(p *Provider) { p.Varargs = true }
}

// makeProviderSet constructs a *ProviderSet and builds its providerMap and srcMap.
// Returns the set and any errors from buildProviderMap/verifyAcyclic.
func makeProviderSet(t *testing.T, providers []*Provider, bindings []*IfaceBinding, values []*Value, fields []*Field, injectorArgs *InjectorArgs) *ProviderSet {
	t.Helper()
	fset := token.NewFileSet()
	hasher := typeutil.MakeHasher()
	pset := &ProviderSet{
		Pos:          token.NoPos,
		PkgPath:      "example.com/test",
		Providers:    providers,
		Bindings:     bindings,
		Values:       values,
		Fields:       fields,
		InjectorArgs: injectorArgs,
	}
	var errs []error
	pset.providerMap, pset.srcMap, errs = buildProviderMap(fset, hasher, pset)
	if len(errs) > 0 {
		t.Fatalf("buildProviderMap failed: %v", errs)
	}
	if errs := verifyAcyclic(pset.providerMap, hasher); len(errs) > 0 {
		t.Fatalf("verifyAcyclic failed: %v", errs)
	}
	return pset
}

// makeProviderSetRaw constructs a *ProviderSet and returns errors from buildProviderMap.
// Unlike makeProviderSet, it does not fail the test on errors.
func makeProviderSetRaw(providers []*Provider, bindings []*IfaceBinding, values []*Value, fields []*Field, injectorArgs *InjectorArgs) (*ProviderSet, []error) {
	fset := token.NewFileSet()
	hasher := typeutil.MakeHasher()
	pset := &ProviderSet{
		Pos:          token.NoPos,
		PkgPath:      "example.com/test",
		Providers:    providers,
		Bindings:     bindings,
		Values:       values,
		Fields:       fields,
		InjectorArgs: injectorArgs,
	}
	var errs []error
	pset.providerMap, pset.srcMap, errs = buildProviderMap(fset, hasher, pset)
	if len(errs) > 0 {
		return pset, errs
	}
	if errs := verifyAcyclic(pset.providerMap, hasher); len(errs) > 0 {
		return pset, errs
	}
	return pset, nil
}

// makeChain creates a linear dependency chain of n providers.
// Each provider takes the output of the previous one as input.
// Returns the providers and the final output type.
func makeChain(pkg *types.Package, n int) ([]*Provider, types.Type) {
	if n < 1 {
		panic("makeChain requires n >= 1")
	}
	chainTypes := make([]types.Type, n+1)
	for i := range chainTypes {
		name := string(rune('A' + i))
		chainTypes[i] = types.NewNamed(
			types.NewTypeName(token.NoPos, pkg, name, nil),
			types.Typ[types.Int],
			nil,
		)
	}
	providers := make([]*Provider, n)
	for i := 0; i < n; i++ {
		var args []ProviderInput
		if i > 0 {
			args = []ProviderInput{{Type: chainTypes[i]}}
		}
		providers[i] = &Provider{
			Pkg:  pkg,
			Name: "New" + string(rune('A'+i+1)),
			Pos:  token.NoPos,
			Args: args,
			Out:  []types.Type{chainTypes[i+1]},
		}
	}
	return providers, chainTypes[n]
}

// assertErrorContains checks that at least one error in errs contains substr.
func assertErrorContains(t *testing.T, errs []error, substr string) {
	t.Helper()
	if len(errs) == 0 {
		t.Fatalf("expected error containing %q, got no errors", substr)
	}
	for _, err := range errs {
		if strings.Contains(err.Error(), substr) {
			return
		}
	}
	msgs := make([]string, len(errs))
	for i, err := range errs {
		msgs[i] = err.Error()
	}
	t.Errorf("expected error containing %q, got:\n%s", substr, strings.Join(msgs, "\n"))
}

// assertNoErrors checks that errs is empty.
func assertNoErrors(t *testing.T, errs []error) {
	t.Helper()
	if len(errs) > 0 {
		msgs := make([]string, len(errs))
		for i, err := range errs {
			msgs[i] = err.Error()
		}
		t.Fatalf("unexpected errors:\n%s", strings.Join(msgs, "\n"))
	}
}

// testPkg creates a test *types.Package.
func testPkg(path, name string) *types.Package {
	return types.NewPackage(path, name)
}

// isIdent reports whether s is a valid Go identifier.
func isIdent(s string) bool {
	if len(s) == 0 {
		return false
	}
	r, i := utf8.DecodeRuneInString(s)
	if !unicode.IsLetter(r) && r != '_' {
		return false
	}
	for i < len(s) {
		r, sz := utf8.DecodeRuneInString(s[i:])
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
		i += sz
	}
	return true
}
