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
	"testing"

	"golang.org/x/tools/go/types/typeutil"
)

func TestBuildProviderMap(t *testing.T) {
	fset := token.NewFileSet()
	hasher := typeutil.MakeHasher()
	pkg := testPkg("example.com/test", "test")
	intT := types.Typ[types.Int]
	stringT := types.Typ[types.String]

	t.Run("single provider", func(t *testing.T) {
		p := makeProvider(pkg, "NewFoo", nil, []types.Type{intT})
		pset := &ProviderSet{
			Pos:       token.NoPos,
			PkgPath:   pkg.Path(),
			Providers: []*Provider{p},
		}
		pm, sm, errs := buildProviderMap(fset, hasher, pset)
		assertNoErrors(t, errs)
		if pm.At(intT) == nil {
			t.Error("expected int type in provider map")
		}
		if sm.At(intT) == nil {
			t.Error("expected int type in source map")
		}
	})

	t.Run("duplicate providers for same type", func(t *testing.T) {
		p1 := makeProvider(pkg, "NewFoo", nil, []types.Type{intT})
		p2 := makeProvider(pkg, "NewBar", nil, []types.Type{intT})
		pset := &ProviderSet{
			Pos:       token.NoPos,
			PkgPath:   pkg.Path(),
			Providers: []*Provider{p1, p2},
		}
		_, _, errs := buildProviderMap(fset, hasher, pset)
		assertErrorContains(t, errs, "multiple bindings")
	})

	t.Run("binding without concrete provider", func(t *testing.T) {
		iface := types.NewInterfaceType(nil, nil).Complete()
		ifaceNamed := types.NewNamed(
			types.NewTypeName(token.NoPos, pkg, "MyIface", nil),
			iface,
			nil,
		)
		binding := &IfaceBinding{
			Iface:    ifaceNamed,
			Provided: intT,
			Pos:      token.NoPos,
		}
		pset := &ProviderSet{
			Pos:      token.NoPos,
			PkgPath:  pkg.Path(),
			Bindings: []*IfaceBinding{binding},
		}
		_, _, errs := buildProviderMap(fset, hasher, pset)
		assertErrorContains(t, errs, "does not include a provider")
	})

	t.Run("injector args", func(t *testing.T) {
		args := &InjectorArgs{
			Name:  "NewService",
			Tuple: vars(intT, stringT),
			Pos:   token.NoPos,
		}
		pset := &ProviderSet{
			Pos:          token.NoPos,
			PkgPath:      pkg.Path(),
			InjectorArgs: args,
		}
		pm, _, errs := buildProviderMap(fset, hasher, pset)
		assertNoErrors(t, errs)
		if pm.At(intT) == nil {
			t.Error("expected int type from injector args")
		}
		if pm.At(stringT) == nil {
			t.Error("expected string type from injector args")
		}
	})

	t.Run("duplicate injector args", func(t *testing.T) {
		args := &InjectorArgs{
			Name:  "NewService",
			Tuple: vars(intT, intT),
			Pos:   token.NoPos,
		}
		pset := &ProviderSet{
			Pos:          token.NoPos,
			PkgPath:      pkg.Path(),
			InjectorArgs: args,
		}
		_, _, errs := buildProviderMap(fset, hasher, pset)
		assertErrorContains(t, errs, "multiple bindings")
	})
}

func TestVerifyAcyclic(t *testing.T) {
	hasher := typeutil.MakeHasher()
	pkg := testPkg("example.com/test", "test")

	makeNamedType := func(name string) types.Type {
		return types.NewNamed(
			types.NewTypeName(token.NoPos, pkg, name, nil),
			types.Typ[types.Int],
			nil,
		)
	}

	t.Run("linear chain (no cycle)", func(t *testing.T) {
		typeA := makeNamedType("A")
		typeB := makeNamedType("B")

		pm := new(typeutil.Map)
		pm.SetHasher(hasher)

		pA := makeProvider(pkg, "NewA", nil, []types.Type{typeA})
		pB := makeProvider(pkg, "NewB", []ProviderInput{{Type: typeA}}, []types.Type{typeB})

		pm.Set(typeA, &ProvidedType{t: typeA, p: pA})
		pm.Set(typeB, &ProvidedType{t: typeB, p: pB})

		errs := verifyAcyclic(pm, hasher)
		assertNoErrors(t, errs)
	})

	t.Run("diamond shape (no cycle)", func(t *testing.T) {
		typeA := makeNamedType("A")
		typeB := makeNamedType("B")
		typeC := makeNamedType("C")
		typeD := makeNamedType("D")

		pm := new(typeutil.Map)
		pm.SetHasher(hasher)

		pA := makeProvider(pkg, "NewA", nil, []types.Type{typeA})
		pB := makeProvider(pkg, "NewB", []ProviderInput{{Type: typeA}}, []types.Type{typeB})
		pC := makeProvider(pkg, "NewC", []ProviderInput{{Type: typeA}}, []types.Type{typeC})
		pD := makeProvider(pkg, "NewD", []ProviderInput{{Type: typeB}, {Type: typeC}}, []types.Type{typeD})

		pm.Set(typeA, &ProvidedType{t: typeA, p: pA})
		pm.Set(typeB, &ProvidedType{t: typeB, p: pB})
		pm.Set(typeC, &ProvidedType{t: typeC, p: pC})
		pm.Set(typeD, &ProvidedType{t: typeD, p: pD})

		errs := verifyAcyclic(pm, hasher)
		assertNoErrors(t, errs)
	})

	t.Run("self cycle", func(t *testing.T) {
		typeA := makeNamedType("A")

		pm := new(typeutil.Map)
		pm.SetHasher(hasher)

		pA := makeProvider(pkg, "NewA", []ProviderInput{{Type: typeA}}, []types.Type{typeA})
		pm.Set(typeA, &ProvidedType{t: typeA, p: pA})

		errs := verifyAcyclic(pm, hasher)
		assertErrorContains(t, errs, "cycle for")
	})

	t.Run("two-node cycle", func(t *testing.T) {
		typeA := makeNamedType("A")
		typeB := makeNamedType("B")

		pm := new(typeutil.Map)
		pm.SetHasher(hasher)

		pA := makeProvider(pkg, "NewA", []ProviderInput{{Type: typeB}}, []types.Type{typeA})
		pB := makeProvider(pkg, "NewB", []ProviderInput{{Type: typeA}}, []types.Type{typeB})

		pm.Set(typeA, &ProvidedType{t: typeA, p: pA})
		pm.Set(typeB, &ProvidedType{t: typeB, p: pB})

		errs := verifyAcyclic(pm, hasher)
		assertErrorContains(t, errs, "cycle for")
	})

	t.Run("three-node cycle", func(t *testing.T) {
		typeA := makeNamedType("A")
		typeB := makeNamedType("B")
		typeC := makeNamedType("C")

		pm := new(typeutil.Map)
		pm.SetHasher(hasher)

		pA := makeProvider(pkg, "NewA", []ProviderInput{{Type: typeC}}, []types.Type{typeA})
		pB := makeProvider(pkg, "NewB", []ProviderInput{{Type: typeA}}, []types.Type{typeB})
		pC := makeProvider(pkg, "NewC", []ProviderInput{{Type: typeB}}, []types.Type{typeC})

		pm.Set(typeA, &ProvidedType{t: typeA, p: pA})
		pm.Set(typeB, &ProvidedType{t: typeB, p: pB})
		pm.Set(typeC, &ProvidedType{t: typeC, p: pC})

		errs := verifyAcyclic(pm, hasher)
		assertErrorContains(t, errs, "cycle for")
	})

	t.Run("value provider (no cycle possible)", func(t *testing.T) {
		typeA := makeNamedType("A")

		pm := new(typeutil.Map)
		pm.SetHasher(hasher)
		pm.Set(typeA, &ProvidedType{t: typeA, v: &Value{Out: typeA}})

		errs := verifyAcyclic(pm, hasher)
		assertNoErrors(t, errs)
	})
}

func TestSolve(t *testing.T) {
	fset := token.NewFileSet()
	pkg := testPkg("example.com/test", "test")

	makeNamedType := func(name string) types.Type {
		return types.NewNamed(
			types.NewTypeName(token.NoPos, pkg, name, nil),
			types.Typ[types.Int],
			nil,
		)
	}

	t.Run("single provider, no dependencies", func(t *testing.T) {
		typeA := makeNamedType("A")
		p := makeProvider(pkg, "NewA", nil, []types.Type{typeA})
		set := makeProviderSet(t, []*Provider{p}, nil, nil, nil, nil)

		calls, errs := solve(fset, typeA, vars(), set)
		assertNoErrors(t, errs)
		if len(calls) != 1 {
			t.Fatalf("got %d calls; want 1", len(calls))
		}
		if calls[0].name != "NewA" {
			t.Errorf("call name = %q; want %q", calls[0].name, "NewA")
		}
	})

	t.Run("linear chain A→B→C", func(t *testing.T) {
		typeA := makeNamedType("A")
		typeB := makeNamedType("B")
		typeC := makeNamedType("C")

		pA := makeProvider(pkg, "NewA", nil, []types.Type{typeA})
		pB := makeProvider(pkg, "NewB", []ProviderInput{{Type: typeA}}, []types.Type{typeB})
		pC := makeProvider(pkg, "NewC", []ProviderInput{{Type: typeB}}, []types.Type{typeC})

		set := makeProviderSet(t, []*Provider{pA, pB, pC}, nil, nil, nil, nil)
		calls, errs := solve(fset, typeC, vars(), set)
		assertNoErrors(t, errs)
		if len(calls) != 3 {
			t.Fatalf("got %d calls; want 3", len(calls))
		}
		// Calls should be in dependency order: A, B, C.
		if calls[0].name != "NewA" {
			t.Errorf("call[0] name = %q; want NewA", calls[0].name)
		}
		if calls[1].name != "NewB" {
			t.Errorf("call[1] name = %q; want NewB", calls[1].name)
		}
		if calls[2].name != "NewC" {
			t.Errorf("call[2] name = %q; want NewC", calls[2].name)
		}
	})

	t.Run("output type is injector argument", func(t *testing.T) {
		typeA := makeNamedType("A")
		args := &InjectorArgs{
			Name:  "Inject",
			Tuple: vars(typeA),
			Pos:   token.NoPos,
		}
		set := makeProviderSet(t, nil, nil, nil, nil, args)

		calls, errs := solve(fset, typeA, vars(typeA), set)
		assertNoErrors(t, errs)
		if len(calls) != 0 {
			t.Fatalf("got %d calls; want 0 (output is an input)", len(calls))
		}
	})

	t.Run("no provider for output type", func(t *testing.T) {
		typeA := makeNamedType("A")
		set := makeProviderSet(t, nil, nil, nil, nil, nil)

		_, errs := solve(fset, typeA, vars(), set)
		assertErrorContains(t, errs, "no provider found")
	})

	t.Run("no provider for transitive dependency", func(t *testing.T) {
		typeA := makeNamedType("A")
		typeB := makeNamedType("B")

		pB := makeProvider(pkg, "NewB", []ProviderInput{{Type: typeA}}, []types.Type{typeB})
		set := makeProviderSet(t, []*Provider{pB}, nil, nil, nil, nil)

		_, errs := solve(fset, typeB, vars(), set)
		assertErrorContains(t, errs, "no provider found")
	})

	t.Run("provider with cleanup", func(t *testing.T) {
		typeA := makeNamedType("A")
		p := makeProvider(pkg, "NewA", nil, []types.Type{typeA}, withCleanup())
		set := makeProviderSet(t, []*Provider{p}, nil, nil, nil, nil)

		calls, errs := solve(fset, typeA, vars(), set)
		assertNoErrors(t, errs)
		if len(calls) != 1 {
			t.Fatalf("got %d calls; want 1", len(calls))
		}
		if !calls[0].hasCleanup {
			t.Error("expected call to have cleanup")
		}
	})

	t.Run("provider with error", func(t *testing.T) {
		typeA := makeNamedType("A")
		p := makeProvider(pkg, "NewA", nil, []types.Type{typeA}, withErr())
		set := makeProviderSet(t, []*Provider{p}, nil, nil, nil, nil)

		calls, errs := solve(fset, typeA, vars(), set)
		assertNoErrors(t, errs)
		if len(calls) != 1 {
			t.Fatalf("got %d calls; want 1", len(calls))
		}
		if !calls[0].hasErr {
			t.Error("expected call to have error")
		}
	})
}

func TestVerifyArgsUsed(t *testing.T) {
	pkg := testPkg("example.com/test", "test")

	makeNamedType := func(name string) types.Type {
		return types.NewNamed(
			types.NewTypeName(token.NoPos, pkg, name, nil),
			types.Typ[types.Int],
			nil,
		)
	}

	t.Run("all providers used", func(t *testing.T) {
		typeA := makeNamedType("A")
		p := makeProvider(pkg, "NewA", nil, []types.Type{typeA})
		src := &providerSetSrc{Provider: p}
		set := &ProviderSet{Providers: []*Provider{p}}

		errs := verifyArgsUsed(set, []*providerSetSrc{src})
		assertNoErrors(t, errs)
	})

	t.Run("unused provider", func(t *testing.T) {
		typeA := makeNamedType("A")
		p := makeProvider(pkg, "NewA", nil, []types.Type{typeA})
		set := &ProviderSet{Providers: []*Provider{p}}

		errs := verifyArgsUsed(set, nil)
		assertErrorContains(t, errs, "unused provider")
	})

	t.Run("unused binding", func(t *testing.T) {
		typeA := makeNamedType("A")
		b := &IfaceBinding{Iface: typeA, Provided: typeA, Pos: token.NoPos}
		set := &ProviderSet{Bindings: []*IfaceBinding{b}}

		errs := verifyArgsUsed(set, nil)
		assertErrorContains(t, errs, "unused interface binding")
	})

	t.Run("unused value", func(t *testing.T) {
		typeA := makeNamedType("A")
		v := &Value{Out: typeA}
		set := &ProviderSet{Values: []*Value{v}}

		errs := verifyArgsUsed(set, nil)
		assertErrorContains(t, errs, "unused value")
	})

	t.Run("unused provider set (named)", func(t *testing.T) {
		imp := &ProviderSet{VarName: "MySet"}
		set := &ProviderSet{Imports: []*ProviderSet{imp}}

		errs := verifyArgsUsed(set, nil)
		assertErrorContains(t, errs, "unused provider set")
		assertErrorContains(t, errs, "MySet")
	})

	t.Run("unused provider set (unnamed)", func(t *testing.T) {
		imp := &ProviderSet{VarName: ""}
		set := &ProviderSet{Imports: []*ProviderSet{imp}}

		errs := verifyArgsUsed(set, nil)
		assertErrorContains(t, errs, "unused provider set")
	})

	t.Run("unused field", func(t *testing.T) {
		typeA := makeNamedType("A")
		f := &Field{
			Parent: typeA,
			Name:   "Foo",
			Pkg:    pkg,
			Pos:    token.NoPos,
			Out:    []types.Type{types.Typ[types.Int]},
		}
		set := &ProviderSet{Fields: []*Field{f}}

		errs := verifyArgsUsed(set, nil)
		assertErrorContains(t, errs, "unused field")
	})
}

func TestBindingConflictError(t *testing.T) {
	fset := token.NewFileSet()
	pkg := testPkg("example.com/test", "test")
	intT := types.Typ[types.Int]

	p1 := makeProvider(pkg, "NewFoo", nil, []types.Type{intT})
	p2 := makeProvider(pkg, "NewBar", nil, []types.Type{intT})

	src1 := &providerSetSrc{Provider: p1}
	src2 := &providerSetSrc{Provider: p2}

	set := &ProviderSet{
		Pos:     token.NoPos,
		VarName: "TestSet",
	}

	err := bindingConflictError(fset, intT, set, src1, src2)
	if err == nil {
		t.Fatal("expected error")
	}
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("expected non-empty error message")
	}
}
