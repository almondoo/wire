# Wire Codebase Guide for AI Assistants

This document provides a comprehensive guide to the Wire codebase for AI assistants working on this project. Wire is a compile-time dependency injection tool for Go that automates connecting components.

## Project Overview

**Repository**: `almondoo/wire`
**Language**: Go (1.19+)
**Type**: Code generation tool / CLI application
**Status**: Beta, feature complete (maintained fork of google/wire)

Wire operates without runtime state or reflection, generating code at compile-time. It analyzes provider functions and builds a directed acyclic graph (DAG) to resolve dependencies, generating initialization code automatically.

## Repository Structure

```
/
├── cmd/wire/              # Main CLI application
│   └── main.go           # Entry point with subcommands (gen, check, diff, show)
├── internal/wire/         # Core code generation logic
│   ├── wire.go           # Main generation logic
│   ├── parse.go          # AST parsing and provider analysis
│   ├── analyze.go        # Dependency graph analysis
│   ├── copyast.go        # AST manipulation utilities
│   ├── errors.go         # Error handling
│   └── testdata/         # 144 test case directories
├── wire.go               # Public API (NewSet, Build, Bind, etc.)
├── docs/                 # User documentation
│   ├── guide.md          # Complete user guide
│   ├── best-practices.md # Best practices
│   ├── faq.md            # Frequently asked questions
│   └── jp/               # Japanese translations
├── _tutorial/            # Tutorial with examples
├── .github/workflows/    # CI/CD configuration
│   └── tests.yml         # Test matrix (Go 1.19-1.25, Linux/macOS/Windows)
├── dockerfiles/          # Docker development environment
├── Makefile              # Development commands (Japanese comments)
└── go.mod                # Module dependencies
```

### Key Directories

- **cmd/wire**: CLI tool implementation using google/subcommands
- **internal/wire**: Core generation engine (not importable by users)
- **Root wire.go**: Public API that users import (marker types only)
- **testdata**: Extensive test suite with golden files

## Core Architecture

### 1. Wire API (wire.go)

The public API consists of marker types and functions:

- `NewSet(...interface{}) ProviderSet` - Groups providers together
- `Build(...interface{}) string` - Marks injector functions
- `Bind(iface, to interface{}) Binding` - Interface bindings
- `Value(interface{}) ProvidedValue` - Literal values
- `InterfaceValue(typ, x interface{}) ProvidedValue` - Interface values
- `Struct(structType interface{}, fieldNames ...string) StructProvider` - Struct injection
- `FieldsOf(structType interface{}, fieldNames ...string) StructFields` - Field extraction

These are **compile-time markers only** - the CLI tool analyzes them via AST parsing.

### 2. Code Generation Pipeline

```
User Code (wire.go)
    ↓
Parse AST → Find injector functions
    ↓
Extract provider sets → Build dependency graph
    ↓
Analyze dependencies → Topological sort
    ↓
Generate initialization code → Format
    ↓
Write wire_gen.go
```

**Key files:**
- `parse.go` (lines 1-1000+): AST parsing, provider extraction
- `analyze.go` (lines 1-400+): Dependency resolution, cycle detection
- `wire.go` in internal/ (lines 1-700+): Code generation, output formatting

### 3. CLI Commands

Defined in `cmd/wire/main.go`:

- `wire gen [packages]` - Generate wire_gen.go (default command)
- `wire check [packages]` - Validate without writing
- `wire diff [packages]` - Show diff between current and new
- `wire show [packages]` - Display dependency graph

Flags: `-header_file`, `-tags`, `-output_file_prefix`

## Development Workflows

### Local Development (Native)

```bash
# Install Wire locally
go install github.com/almondoo/wire/cmd/wire@latest

# Run tests
go test -mod=readonly -race ./...

# Run tests with coverage
go test -cover -mod=readonly -race ./...

# Format code (required before commit)
gofmt -s -w .

# Check dependencies
./internal/listdeps.sh > ./internal/alldeps
```

### Docker Development

The project includes a comprehensive Docker setup:

```bash
# Start dev environment
make dev

# Enter shell
make shell

# Run tests in container
make test

# Run verbose tests
make test-verbose

# Run with coverage
make test-cover

# Format and vet
make lint
```

The Makefile uses Japanese comments but commands are in English.

### Testing Strategy

1. **Unit Tests**: `wire_test.go` in internal/wire
2. **Integration Tests**: 144+ testdata directories with golden files
3. **Test Pattern**: Each testdata/ subdirectory contains:
   - `foo/*.go` - Input source files
   - `want/wire_gen.go` - Expected output
4. **Version Testing**: Scripts for testing across Go versions:
   - `test_all_versions.sh`
   - `generate_version_errs.sh`

### CI/CD Pipeline

**File**: `.github/workflows/tests.yml`

**Matrix Strategy**:
- OS: ubuntu-latest, macos-latest, windows-latest
- Go versions: 1.19.x through 1.25.x
- Test script: `internal/runtests.sh`

**Checks performed**:
1. Run all tests with race detector
2. Verify gofmt formatting (Linux only)
3. Check dependency list matches `./internal/alldeps` (Linux only)

## Code Conventions

### 1. Go Style

- **Formatting**: Use `gofmt -s` (simplify mode required)
- **Naming**: Standard Go conventions (CamelCase exports)
- **Errors**: Return errors explicitly, no panics in library code
- **Comments**: All exported symbols must have doc comments
- **Testing**: Table-driven tests preferred

### 2. Provider Conventions

```go
// Provider functions should:
// - Be exported if in provider sets
// - Have clear names (ProvideX, NewX)
// - Return (Type, error) for fallible operations
// - Return (Type, func(), error) for cleanup functions
```

### 3. AST Manipulation

- Use `golang.org/x/tools/go/ast/astutil` for AST operations
- Preserve source formatting where possible
- Use `go/format` for final output
- Handle both `GOPATH` and module mode

### 4. Error Messages

- Wire has detailed error messages for common issues
- Error files: `COMPLETE_ERROR_MESSAGES_LIST.txt`, `ERROR_SUMMARY.txt`
- Analysis: `WIRE_ERROR_ANALYSIS_REPORT.md`
- Errors should include:
  - Clear description of the problem
  - Location (file:line)
  - Suggested fixes when possible

### 5. Dependencies

**Direct dependencies** (from go.mod):
```
github.com/google/go-cmp v0.6.0         # Testing comparisons
github.com/google/subcommands v1.2.0    # CLI framework
github.com/pmezard/go-difflib v1.0.0    # Diff generation
golang.org/x/tools v0.24.1              # Go tooling (AST, packages)
```

**Indirect**:
```
golang.org/x/mod v0.20.0    # Module support
golang.org/x/sync v0.8.0    # Concurrency primitives
```

## Important Implementation Details

### 1. Injector Function Detection

Wire identifies injector functions by:
- Looking for `wire.Build()` calls in function bodies
- Function must have exactly one Build call
- Build arguments define the provider set
- Return types determine what's provided

### 2. Provider Types

Wire supports multiple provider types:
- **Functions**: Standard providers
- **Structs**: Field-based initialization (`wire.Struct`)
- **Values**: Compile-time constants (`wire.Value`)
- **Bindings**: Interface implementations (`wire.Bind`)
- **Field extraction**: From structs (`wire.FieldsOf`)

### 3. Dependency Resolution

Process:
1. Build type-to-provider map
2. Start from injector return type
3. Walk dependency graph backwards
4. Detect cycles and missing providers
5. Topologically sort providers
6. Generate initialization code

### 4. Code Generation

Output format (`wire_gen.go`):
```go
// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/almondoo/wire/cmd/wire
//+build !wireinject

package example

// Generated injector functions follow...
```

## Common Tasks for AI Assistants

### Adding a New Feature

1. **File an issue first** (per CONTRIBUTING.md) - project is feature complete
2. Understand that new features are generally not accepted
3. Bug fixes are welcome

### Fixing a Bug

1. **Reproduce**: Create test case in `internal/wire/testdata/`
2. **Diagnose**: Debug using `wire show` command
3. **Fix**: Modify parsing, analysis, or generation code
4. **Test**: Run full test suite across versions
5. **Format**: Ensure `gofmt -s` compliance

### Working with Tests

```bash
# Run specific test
go test -v -run TestSpecificCase ./internal/wire

# Update golden files (carefully!)
# 1. Make changes
# 2. Run tests to generate new output
# 3. Manually verify correctness
# 4. Copy to want/ directory

# Test across Go versions
cd internal/wire && ./test_all_versions.sh
```

### Understanding Errors

When Wire reports errors:
1. Check error location (file:line in user code)
2. Review provider set composition
3. Use `wire show` to visualize dependency graph
4. Common issues:
   - Cycle in dependency graph
   - Missing provider for type
   - Type mismatch in binding
   - Multiple providers for same type

## Git Workflow

### Branch Strategy

- **main**: Stable branch
- **Feature branches**: Use `claude/` prefix for AI assistant work
- **PR Requirements**:
  - All tests pass on all platforms
  - Code formatted with `gofmt -s`
  - Dependencies match `./internal/alldeps`
  - Commits will be squashed on merge

### Commit Guidelines

- Clear, descriptive messages
- Reference issues: `Fixes #123`
- Use conventional commits style
- One logical change per commit
- Will be squashed to single commit on merge

### Review Process

From CONTRIBUTING.md:
- All submissions require review
- Almost all PRs require some changes
- Add "PTAL" comment when ready for re-review
- Assignees must approve before merge
- Use `git merge` not `git rebase` after creating PR
- Never use `git push --force`

## Language-Specific Notes

### Japanese Documentation

The project includes Japanese documentation:
- `JP_README.md` - Japanese README
- `docs/jp/` - Japanese versions of guides
- Makefile comments are in Japanese
- `_tutorial/JP_README.md` - Japanese tutorial

## Module Information

**Module path**: `github.com/almondoo/wire`
**Import path for users**: `github.com/almondoo/wire`
**Tool installation**: `go install github.com/almondoo/wire/cmd/wire@latest`

## Key Constraints

1. **No runtime reflection** - all code generation at compile-time
2. **Type safety** - leverages Go type system
3. **No global state** - explicit initialization
4. **Feature complete** - new features generally not accepted
5. **Backward compatibility** - must maintain API stability

## Testing Requirements

Before submitting changes:

```bash
# 1. Format code
gofmt -s -w .

# 2. Run tests
go test -mod=readonly -race ./...

# 3. Check dependencies
./internal/listdeps.sh | diff ./internal/alldeps -

# 4. (Optional) Test across versions
cd internal/wire && ./test_all_versions.sh
```

## Useful References

- **User Guide**: `docs/guide.md` - Complete Wire usage guide
- **Best Practices**: `docs/best-practices.md` - Patterns and anti-patterns
- **FAQ**: `docs/faq.md` - Common questions
- **Tutorial**: `_tutorial/README.md` - Step-by-step walkthrough
- **Original blog post**: https://blog.golang.org/wire
- **GoDoc**: https://godoc.org/github.com/almondoo/wire

## Support and Community

- **Issues**: https://github.com/almondoo/wire/issues
- **Discussions**: https://github.com/almondoo/wire/discussions
- **Code of Conduct**: `CODE_OF_CONDUCT.md` (Go CoC)

## Quick Reference Card

```go
// Define providers
func ProvideDB() *sql.DB { ... }
func ProvideRepo(db *sql.DB) *Repo { ... }

// Group into set
var Set = wire.NewSet(ProvideDB, ProvideRepo)

// Create injector (in wire.go with //+build wireinject)
//+build wireinject

func InitializeRepo() (*Repo, error) {
    wire.Build(Set)
    return nil, nil  // Wire generates actual implementation
}

// Run: wire gen
// Output: wire_gen.go with actual implementation
```

## Notes for AI Assistants

1. **Always run tests** before suggesting changes
2. **Respect feature freeze** - focus on bugs, not features
3. **Understand AST manipulation** - core of Wire
4. **Test across Go versions** - compatibility is critical
5. **Read error analysis docs** - helps understand common issues
6. **Use Docker for consistency** - matches CI environment
7. **Follow review conventions** from CONTRIBUTING.md
8. **Never skip gofmt** - CI will fail
9. **Update dependency list** if adding imports
10. **Preserve backward compatibility** - users rely on stable API

---

**Last Updated**: 2025-11-15
**Wire Version**: v0.3.0+
**Minimum Go Version**: 1.19
