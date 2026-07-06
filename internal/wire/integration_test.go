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

// This file contains lightweight integration tests for the entry points of
// the code generation pipeline, Generate and Load. Unlike the rest of the
// tests in this package, which exercise internal helpers directly against
// synthetic *types values, these tests write a minimal real Go module to a
// temporary directory and drive Generate/Load exactly the way the wire CLI
// does: by loading real on-disk packages with golang.org/x/tools/go/packages
// and generating actual source code from them. This guards against
// regressions in the wiring between package loading, analysis, and code
// generation that unit tests of the individual stages cannot catch.
//
// The fixture module depends on github.com/almondoo/wire through a
// filesystem "replace" directive pointing back at the repository root, so
// no network access or module proxy is required: the root wire.go marker
// package has no external imports, so the Go toolchain never needs to
// resolve any transitive dependency to type-check it.

import (
	"bytes"
	"context"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// integrationTimeout bounds how long a single Generate/Load call may take.
// Loading packages spawns a `go list` subprocess, which is slow relative to
// the rest of this package's tests but should still complete quickly with a
// warm module/build cache.
const integrationTimeout = 60 * time.Second

// writeIntegrationModule writes go.mod for a throwaway module named
// example.com/wiretest rooted at dir. The module requires
// github.com/almondoo/wire and replaces it with the repository root
// (resolved relative to this test file's package directory, internal/wire)
// so the fixture always refers to the checkout under test, whether running
// on the host or inside the Docker dev container.
func writeIntegrationModule(t *testing.T, dir string) {
	t.Helper()

	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}

	goMod := "module example.com/wiretest\n\n" +
		"go 1.19\n\n" +
		"require github.com/almondoo/wire v0.0.0-00010101000000-000000000000\n\n" +
		"replace github.com/almondoo/wire => " + repoRoot + "\n"
	writeIntegrationFile(t, filepath.Join(dir, "go.mod"), goMod)
}

// writeIntegrationFile writes content to path, failing the test on error.
func writeIntegrationFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// writeValidInjectorFixture writes a module with a single struct provider
// (Greeter/NewGreeter) and a single injector (InitializeGreeter) that wires
// them together. This is the minimal shape that should generate valid,
// non-empty output.
func writeValidInjectorFixture(t *testing.T, dir string) {
	t.Helper()
	writeIntegrationModule(t, dir)

	writeIntegrationFile(t, filepath.Join(dir, "providers.go"), `package wiretest

type Greeter struct {
	Message string
}

func NewGreeter() *Greeter {
	return &Greeter{Message: "hello, wire"}
}
`)

	writeIntegrationFile(t, filepath.Join(dir, "wire.go"), `//go:build wireinject
// +build wireinject

package wiretest

import "github.com/almondoo/wire"

func InitializeGreeter() *Greeter {
	wire.Build(NewGreeter)
	return nil
}
`)
}

// writeMissingProviderFixture writes a module whose injector requires a
// type (*Widget) with no provider in scope. Generate/Load should still load
// the package successfully, but Generate must report the unresolved
// dependency as a generation error rather than panicking or silently
// emitting empty output.
func writeMissingProviderFixture(t *testing.T, dir string) {
	t.Helper()
	writeIntegrationModule(t, dir)

	writeIntegrationFile(t, filepath.Join(dir, "providers.go"), `package wiretest

type Widget struct {
	Name string
}
`)

	writeIntegrationFile(t, filepath.Join(dir, "wire.go"), `//go:build wireinject
// +build wireinject

package wiretest

import "github.com/almondoo/wire"

func InitializeWidget() *Widget {
	wire.Build()
	return nil
}
`)
}

// writeNoInjectorFixture writes a module with an ordinary provider but no
// injector function at all, exercising the "nothing to generate" path.
func writeNoInjectorFixture(t *testing.T, dir string) {
	t.Helper()
	writeIntegrationModule(t, dir)

	writeIntegrationFile(t, filepath.Join(dir, "providers.go"), `package wiretest

type Greeter struct {
	Message string
}

func NewGreeter() *Greeter {
	return &Greeter{Message: "hello, wire"}
}
`)
}

// integrationEnv returns the environment to load the fixture module with.
// It disables the module proxy so the test fails fast and loudly if it ever
// starts requiring network access, instead of hanging or flaking on it.
func integrationEnv() []string {
	return append(os.Environ(), "GOPROXY=off")
}

func TestGenerateIntegration(t *testing.T) {
	dir := t.TempDir()
	writeValidInjectorFixture(t, dir)

	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	defer cancel()

	results, errs := Generate(ctx, dir, integrationEnv(), []string{"."}, &GenerateOptions{})
	if len(errs) > 0 {
		t.Fatalf("Generate returned load errors: %v", errs)
	}
	if len(results) != 1 {
		t.Fatalf("got %d GenerateResults, want 1: %+v", len(results), results)
	}

	res := results[0]
	if len(res.Errs) > 0 {
		t.Fatalf("GenerateResult.Errs is non-empty: %v", res.Errs)
	}
	if len(res.Content) == 0 {
		t.Fatal("GenerateResult.Content is empty, want generated source")
	}

	content := string(res.Content)
	if !strings.Contains(content, "func InitializeGreeter()") {
		t.Errorf("generated content missing injector InitializeGreeter():\n%s", content)
	}
	if !strings.Contains(content, "NewGreeter()") {
		t.Errorf("generated content missing call to provider NewGreeter():\n%s", content)
	}
	if !strings.Contains(content, "Code generated by Wire. DO NOT EDIT.") {
		t.Errorf("generated content missing Wire header comment:\n%s", content)
	}
	if !strings.Contains(content, "!wireinject") {
		t.Errorf("generated content missing !wireinject build tag:\n%s", content)
	}

	// The generated source must be valid, already-gofmt'd Go source, since
	// Generate runs format.Source over it before returning.
	formatted, err := format.Source(res.Content)
	if err != nil {
		t.Fatalf("generated content is not valid Go source: %v\n%s", err, content)
	}
	if !bytes.Equal(formatted, res.Content) {
		t.Errorf("generated content is not gofmt'd:\ngot:\n%s\nwant:\n%s", res.Content, formatted)
	}
}

func TestGenerateIntegrationMissingProvider(t *testing.T) {
	dir := t.TempDir()
	writeMissingProviderFixture(t, dir)

	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	defer cancel()

	results, errs := Generate(ctx, dir, integrationEnv(), []string{"."}, &GenerateOptions{})
	if len(errs) > 0 {
		t.Fatalf("Generate returned load errors: %v, want package to load and fail at generation instead", errs)
	}
	if len(results) != 1 {
		t.Fatalf("got %d GenerateResults, want 1: %+v", len(results), results)
	}

	res := results[0]
	if len(res.Errs) == 0 {
		t.Fatalf("GenerateResult.Errs is empty, want an unresolved-provider error; Content:\n%s", res.Content)
	}
	assertErrorContains(t, res.Errs, "no provider found")
	if len(res.Content) != 0 {
		t.Errorf("GenerateResult.Content = %q, want empty when generation fails", res.Content)
	}
}

func TestGenerateIntegrationNoInjectors(t *testing.T) {
	dir := t.TempDir()
	writeNoInjectorFixture(t, dir)

	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	defer cancel()

	results, errs := Generate(ctx, dir, integrationEnv(), []string{"."}, &GenerateOptions{})
	if len(errs) > 0 {
		t.Fatalf("Generate returned load errors: %v", errs)
	}
	if len(results) != 1 {
		t.Fatalf("got %d GenerateResults, want 1: %+v", len(results), results)
	}
	if len(results[0].Errs) > 0 {
		t.Errorf("GenerateResult.Errs is non-empty for a package with no injectors: %v", results[0].Errs)
	}
}

func TestLoadIntegration(t *testing.T) {
	dir := t.TempDir()
	writeValidInjectorFixture(t, dir)

	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	defer cancel()

	info, errs := Load(ctx, dir, integrationEnv(), "", []string{"."})
	if len(errs) > 0 {
		t.Fatalf("Load returned errors: %v", errs)
	}
	if info == nil {
		t.Fatal("Load returned nil Info")
	}
	if len(info.Injectors) != 1 {
		t.Fatalf("got %d injectors, want 1: %+v", len(info.Injectors), info.Injectors)
	}

	inj := info.Injectors[0]
	if inj.FuncName != "InitializeGreeter" {
		t.Errorf("FuncName = %q, want %q", inj.FuncName, "InitializeGreeter")
	}
	if !strings.HasSuffix(inj.ImportPath, "wiretest") {
		t.Errorf("ImportPath = %q, want suffix %q", inj.ImportPath, "wiretest")
	}
}

func TestLoadIntegrationNoInjectors(t *testing.T) {
	dir := t.TempDir()
	writeNoInjectorFixture(t, dir)

	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	defer cancel()

	info, errs := Load(ctx, dir, integrationEnv(), "", []string{"."})
	if len(errs) > 0 {
		t.Fatalf("Load returned errors: %v", errs)
	}
	if info == nil {
		t.Fatal("Load returned nil Info")
	}
	if len(info.Injectors) != 0 {
		t.Errorf("got %d injectors, want 0: %+v", len(info.Injectors), info.Injectors)
	}
}

func TestLoadIntegrationMissingProvider(t *testing.T) {
	dir := t.TempDir()
	writeMissingProviderFixture(t, dir)

	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	defer cancel()

	// Unlike a pure signature check, Load runs the same dependency solver
	// as Generate (see parse.go's call to solve()), so an unresolved
	// provider surfaces directly as a Load error and the injector is
	// omitted from Info.Injectors.
	info, errs := Load(ctx, dir, integrationEnv(), "", []string{"."})
	if len(errs) == 0 {
		t.Fatalf("Load returned no errors, want an unresolved-provider error; Injectors: %+v", info.Injectors)
	}
	assertErrorContains(t, errs, "no provider found")
	if info != nil && len(info.Injectors) != 0 {
		t.Errorf("got %d injectors, want 0 since the injector failed to solve: %+v", len(info.Injectors), info.Injectors)
	}
}
