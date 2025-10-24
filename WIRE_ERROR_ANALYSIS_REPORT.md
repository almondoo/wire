# Wire Codebase Error Handling Analysis Report

## Executive Summary

This report provides a comprehensive analysis of all error handling patterns in the Google Wire dependency injection framework. Through systematic exploration of the `/home/user/wire/internal/wire` package, I've identified:

- **27 distinct error types/categories** that Wire can detect and report
- **70 total test cases**, of which **27 test error conditions**
- **Multiple untested error paths** that should be considered for additional test coverage

---

## Part 1: Error Types That Wire Can Detect and Report

### Core Error Categories

#### 1. **Dependency Resolution Errors**

##### A. Missing Provider Errors
- **Message Pattern**: `"no provider found for TYPE"` or `"no provider found for TYPE, output of injector"`
- **Location**: `analyze.go:solve()` (lines 129-140)
- **Trigger Conditions**:
  - No provider exists for a required type
  - Type is requested as injector output but not available
  - Type is transitively needed but not available
- **Test Cases**: `MultipleMissingInputs`, `NoImplicitInterface`, `FieldsOfStructDoNotProvidePtrToField`
- **Status**: TESTED

##### B. Cyclic Dependency Errors
- **Message Pattern**: `"cycle for TYPE: PROV1 -> PROV2 -> ... -> TYPE"`
- **Location**: `analyze.go:verifyAcyclic()` (lines 432-507)
- **Trigger Conditions**:
  - Provider A requires type from B
  - Provider B requires type from C
  - Provider C requires type from A (or other cycle)
- **Test Cases**: `Cycle`, `FieldsOfCycle`
- **Status**: TESTED

##### C. Multiple Binding Conflicts
- **Message Pattern**: `"multiple bindings for TYPE\ncurrent:\n<- SOURCE1\nprevious:\n<- SOURCE2"`
- **Location**: `analyze.go:buildProviderMap()` (lines 346-425) and `bindingConflictError()` (lines 512-521)
- **Trigger Conditions**:
  - Two providers supply the same type
  - One argument and one provider for same type
  - Multiple values/bindings/fields for same type
  - Across imported provider sets
- **Test Cases**: `MultipleBindings`, `InjectInputConflict`, `MultipleArgsSameType`
- **Status**: TESTED

#### 2. **Unused Resource Errors**

##### Message Patterns (all from `analyze.go:verifyArgsUsed()` lines 259-326):
- `"unused provider set"` - Provider set imported but not used
- `"unused provider set \"NAME\""` - Named provider set unused
- `"unused provider \"PKG.NAME\""` - Provider function never called
- `"unused value of type TYPE"` - wire.Value never used
- `"unused interface binding to type TYPE"` - wire.Bind never used
- `"unused field \"TYPE.FIELDNAME\""` - wire.FieldsOf field never used

**Test Case**: `UnusedProviders`
**Status**: TESTED

#### 3. **Type Compatibility Errors**

##### A. Interface Implementation Errors
- **Message Pattern**: `"CONCRETE_TYPE does not implement INTERFACE_TYPE"`
- **Location**: `parse.go:processBind()` (line 920) and `processInterfaceValue()` (line 993)
- **Trigger Conditions**:
  - Type provided to wire.Bind() doesn't implement the interface
  - Type provided to wire.InterfaceValue() doesn't implement the interface
- **Test Cases**: `InterfaceBindingDoesntImplement`, `InterfaceValueDoesntImplement`
- **Status**: TESTED

##### B. Interface Type Requirement Errors
- **Message Pattern**: `"first argument to [Bind|InterfaceValue] must be a pointer to an interface type; found TYPE"`
- **Location**: `parse.go:processBind()` (lines 894-902) and `processInterfaceValue()` (lines 982-989)
- **Trigger Conditions**:
  - First argument to wire.Bind() is not a pointer to interface
  - First argument to wire.InterfaceValue() is not a pointer to interface
- **Test Cases**: `InterfaceBindingInvalidArg0`, `InterfaceValueInvalidArg0`
- **Status**: TESTED

##### C. Self-Binding Errors
- **Message Pattern**: `"cannot bind interface to itself"`
- **Location**: `parse.go:processBind()` (line 916)
- **Trigger Conditions**:
  - wire.Bind(*I, I) - attempting to bind interface I to itself
- **Test Cases**: None
- **Status**: UNTESTED

##### D. Concrete Type Binding Errors
- **Message Pattern**: `"wire.Bind of concrete type \"CONCRETE\" to interface \"IFACE\", but SETNAME does not include a provider for \"CONCRETE\""`
- **Location**: `analyze.go:buildProviderMap()` (lines 414-421)
- **Trigger Conditions**:
  - wire.Bind() references a concrete type not provided by the set
  - Set lacks a provider for the concrete implementation
- **Test Cases**: `ProviderSetBindingMissingConcreteType`
- **Status**: TESTED

#### 4. **Struct/Field Errors**

##### A. Struct Type Errors
- **Message Pattern**: `"first argument to Struct must be a pointer to a named struct; found TYPE"`
- **Location**: `parse.go:processStructProvider()` (lines 809-818)
- **Trigger Conditions**:
  - wire.Struct() first arg is not a pointer to struct
  - wire.Struct() first arg is pointer to non-struct
  - wire.Struct() first arg is double pointer to struct
- **Test Cases**: `StructNotAStruct`
- **Status**: TESTED

##### B. Field Prevention Errors
- **Message Pattern**: `"FIELD_NAME is prevented from injecting by wire"`
- **Location**: `parse.go:checkField()` (line 1073) and `isPrevented()` (line 880)
- **Trigger Conditions**:
  - Field has `wire:"-"` struct tag
  - wire.Struct() or wire.FieldsOf() tries to use that field
- **Test Cases**: `StructWithPreventTag`
- **Status**: TESTED

##### C. Field Count Errors
- **Message Pattern**: `"fields number exceeds the number available in the struct which has N fields"`
- **Location**: `parse.go:processFieldsOf()` (lines 1035-1037)
- **Trigger Conditions**:
  - wire.FieldsOf() specifies more field names than struct has
- **Test Cases**: None
- **Status**: UNTESTED

##### D. Field Specification Errors
- **Message Patterns**:
  - `"EXPR must be a string with the field name"` - field name not string literal
  - `"\"FIELDNAME\" is not a field of STRUCT"` - field doesn't exist
- **Location**: `parse.go:checkField()` (lines 1065-1078)
- **Trigger Conditions**:
  - wire.FieldsOf(s, variable_name) - using variable instead of string
  - wire.FieldsOf(s, "NonExistent") - field doesn't exist in struct
- **Test Cases**: None
- **Status**: UNTESTED

#### 5. **Provider Function Signature Errors**

##### A. Return Type Validation Errors
- **Location**: `parse.go:funcOutput()` (lines 722-753)
- **Message Patterns**:
  - `"no return values"` - function has 0 return values
  - `"second return type is TYPE; must be error or func()"` - 2nd return is neither
  - `"second return type is TYPE; must be func()"` - in 3-return case, 2nd isn't cleanup func
  - `"third return type is TYPE; must be error"` - 3rd return isn't error
  - `"too many return values"` - function has 4+ returns
- **Test Cases**: None
- **Status**: UNTESTED

##### B. Parameter Duplication Errors
- **Message Pattern**: `"provider has multiple parameters of type TYPE"`
- **Location**: `parse.go:processFuncProvider()` (lines 698-702)
- **Trigger Conditions**:
  - Provider function takes same type as two or more parameters
- **Test Cases**: None (implicitly related to `MultipleArgsSameType`)
- **Status**: UNTESTED

##### C. Wrong Signature Errors
- **Message Pattern**: `"wrong signature for provider FUNC_NAME: ERROR_DETAILS"`
- **Location**: `parse.go:processFuncProvider()` (line 681)
- **Trigger Conditions**:
  - funcOutput() returns an error for the provider function signature
- **Test Cases**: None
- **Status**: UNTESTED

#### 6. **wire.Value Errors**

##### A. Argument Count Errors
- **Message Pattern**: `"call to Value takes exactly one argument"`
- **Location**: `parse.go:processValue()` (line 934)
- **Trigger Conditions**:
  - wire.Value() called with 0, 2, or more arguments
- **Test Cases**: None
- **Status**: UNTESTED

##### B. Expression Complexity Errors
- **Message Pattern**: `"argument to Value is too complex"`
- **Location**: `parse.go:processValue()` (line 959)
- **Trigger Conditions**:
  - Expression contains channel send (<-), function calls, or other complex operations
- **Test Cases**: None
- **Status**: UNTESTED

##### C. Interface Value Errors
- **Message Pattern**: `"argument to Value may not be an interface value (found TYPE); use InterfaceValue instead"`
- **Location**: `parse.go:processValue()` (line 964)
- **Trigger Conditions**:
  - wire.Value(interfaceValue) - using interface type directly
- **Test Cases**: `ValueIsInterfaceValue`
- **Status**: TESTED

#### 7. **wire.InterfaceValue Errors**

##### Argument Count Error
- **Message Pattern**: `"call to InterfaceValue takes exactly two arguments"`
- **Location**: `parse.go:processInterfaceValue()` (line 979)
- **Trigger Conditions**:
  - wire.InterfaceValue() called with != 2 arguments
- **Test Cases**: `InterfaceValueNotEnoughArgs` (caught by Go compiler)
- **Status**: PARTIALLY TESTED

#### 8. **wire.FieldsOf Errors**

##### A. Argument Specification Errors
- **Message Pattern**: `"call to FieldsOf must specify fields to be extracted"`
- **Location**: `parse.go:processFieldsOf()` (lines 1007-1009)
- **Trigger Conditions**:
  - wire.FieldsOf() called with only 1 argument (the struct)
- **Test Cases**: None
- **Status**: UNTESTED

##### B. First Argument Type Errors
- **Message Pattern**: `"first argument to FieldsOf must be a pointer to a struct or a pointer to a pointer to a struct; found TYPE"`
- **Location**: `parse.go:processFieldsOf()` (lines 1015-1033)
- **Trigger Conditions**:
  - First argument is not a pointer
  - First argument is pointer to non-struct or non-pointer-to-struct
- **Test Cases**: None
- **Status**: UNTESTED

#### 9. **wire.Bind Additional Errors**

##### Second Argument Type Error
- **Message Pattern**: `"second argument to Bind must be a pointer or a pointer to a pointer; found TYPE"`
- **Location**: `parse.go:processBind()` (lines 908-910)
- **Trigger Conditions**:
  - Concrete type for wire.Bind() is neither * nor **
- **Test Cases**: None
- **Status**: UNTESTED

#### 10. **Injector Function Errors**

##### A. Invalid Injector Structure
- **Message Pattern**: `"a call to wire.Build indicates that this function is an injector, but injectors must consist of only the wire.Build call and an optional return"`
- **Location**: `parse.go:findInjectorBuild()` (line 1131)
- **Trigger Conditions**:
  - Injector function contains statements other than wire.Build call and return
  - Multiple statements before return
- **Test Cases**: `InvalidInjector`
- **Status**: TESTED

##### B. Return Type Validation
- **Message Patterns** (from `funcOutput()` - applies to injectors):
  - Same as provider return type validation (error, cleanup, error messages)
- **Location**: `parse.go:injectorFuncSignature()` (line 708)
- **Trigger Conditions**:
  - Injector has 0, 4+ returns, or wrong return types
- **Test Cases**: None (caught by Go type checker)
- **Status**: UNTESTED

##### C. Cleanup Mismatch Errors
- **Message Pattern**: `"inject INJECTOR: provider for TYPE returns cleanup but injection does not return cleanup function"`
- **Location**: `wire.go:inject()` (lines 336-340)
- **Trigger Conditions**:
  - Provider function returns cleanup func
  - Injector function doesn't return cleanup func
- **Test Cases**: `InjectorMissingCleanup`
- **Status**: TESTED

##### D. Error Return Mismatch Errors
- **Message Pattern**: `"inject INJECTOR: provider for TYPE returns error but injection not allowed to fail"`
- **Location**: `wire.go:inject()` (lines 342-346)
- **Trigger Conditions**:
  - Provider function returns error
  - Injector function doesn't return error
- **Test Cases**: `InjectorMissingError`
- **Status**: TESTED

#### 11. **Visibility and Scope Errors**

##### A. Unexported Identifier Errors
- **Message Pattern**: `"uses unexported identifier IDENT"`
- **Location**: `wire.go:accessibleFrom()` (line 944)
- **Trigger Conditions**:
  - wire.Value() uses unexported identifier from different package
- **Test Cases**: `UnexportedValue`
- **Status**: TESTED

##### B. Package Scope Errors
- **Message Pattern**: `"IDENT is not declared in package scope"`
- **Location**: `wire.go:accessibleFrom()` (line 948)
- **Trigger Conditions**:
  - wire.Value() uses local variable from function scope
- **Test Cases**: `ValueFromFunctionScope`
- **Status**: TESTED

##### C. Value Accessibility Errors
- **Message Pattern**: `"inject INJECTOR: value TYPE can't be used: ERROR_DETAILS"`
- **Location**: `wire.go:inject()` (lines 349-354)
- **Trigger Conditions**:
  - Combined visibility/scope error wrapping
- **Test Cases**: `UnexportedValue`, `ValueFromFunctionScope`
- **Status**: TESTED

#### 12. **Variable/Object Errors**

##### Variable Not Provider/Set Errors
- **Message Pattern**: `"VAR is not a provider or a provider set"`
- **Location**: `parse.go:get()` (lines 485, 498)
- **Trigger Conditions**:
  - Variable assigned to non-provider value (e.g., struct{})
  - Variable declared as non-function, non-call expression
- **Test Cases**: `EmptyVar`, `FuncArgProvider`
- **Status**: TESTED

#### 13. **Package/Export Errors**

##### Export Errors
- **Message Pattern**: `"name IDENT not exported by package PKG"`
- **Location**: Go compiler (not directly Wire)
- **Trigger Conditions**:
  - Using unexported struct type from another package
- **Test Cases**: `UnexportedStruct`
- **Status**: TESTED (caught by Go compiler)

---

## Part 2: Test Coverage Analysis

### Tests WITH Error Coverage (27 tests)

| Test Name | Error Type | Error Pattern |
|-----------|-----------|---------------|
| Cycle | Cycle Detection | "cycle for..." |
| EmptyVar | Invalid Variable | "is not a provider..." |
| FieldsOfCycle | Cycle Detection | "cycle for..." |
| FieldsOfStructDoNotProvidePtrToField | Missing Provider | "no provider found..." |
| FuncArgProvider | Invalid Variable | "is not a provider..." |
| InjectInputConflict | Multiple Bindings | "multiple bindings..." |
| InjectorMissingCleanup | Cleanup Mismatch | "returns cleanup but injection does not" |
| InjectorMissingError | Error Mismatch | "returns error but injection not allowed" |
| InterfaceBindingDoesntImplement | Type Incompatibility | "does not implement" |
| InterfaceBindingInvalidArg0 | Invalid Argument | "must be pointer to interface..." |
| InterfaceBindingNotEnoughArgs | Signature Error | "not enough arguments..." (compiler) |
| InterfaceValueDoesntImplement | Type Incompatibility | "does not implement" |
| InterfaceValueInvalidArg0 | Invalid Argument | "must be pointer to interface..." |
| InterfaceValueNotEnoughArgs | Signature Error | "not enough arguments..." (compiler) |
| InvalidInjector | Invalid Injector | "must consist of only wire.Build" |
| MultipleArgsSameType | Multiple Bindings | "multiple bindings..." |
| MultipleBindings | Multiple Bindings | "multiple bindings..." |
| MultipleMissingInputs | Missing Provider | "no provider found..." |
| NoImplicitInterface | Missing Provider | "no provider found..." |
| ProviderSetBindingMissingConcreteType | Binding Error | "does not include a provider for" |
| StructNotAStruct | Struct Type Error | "must be pointer to named struct" |
| StructWithPreventTag | Field Prevention | "prevented from injecting" |
| UnexportedStruct | Export Error | "not exported by package" |
| UnexportedValue | Visibility Error | "uses unexported identifier" |
| UnusedProviders | Unused Resources | "unused provider/set/value/binding/field" |
| ValueFromFunctionScope | Scope Error | "not declared in package scope" |
| ValueIsInterfaceValue | Interface Value Error | "may not be interface value" |

### Tests WITHOUT Error Coverage (43 tests)

These are successful test cases, not error cases:

`BindInjectorArg`, `BindInjectorArgPointer`, `BindInterfaceWithValue`, `BuildTagsAllPackages`, `Chain`, `Cleanup`, `CopyOtherDecls`, `DocComment`, `ExampleWithMocks`, `ExportedValue`, `ExportedValueDifferentPackage`, `FieldsOfImportedStruct`, `FieldsOfStruct`, `FieldsOfStructPointer`, `FieldsOfValueStruct`, `Header`, `ImportedInterfaceBinding`, `InjectInput`, `InjectWithPanic`, `InterfaceBinding`, `InterfaceBindingReuse`, `InterfaceValue`, `MultipleSimilarPackages`, `NamingWorstCase`, `NamingWorstCaseAllInOne`, `NiladicIdentity`, `NiladicValue`, `NoInjectParamNames`, `NoopBuild`, `PartialCleanup`, `PkgImport`, `RelativePkg`, `ReservedKeywords`, `ReturnArgumentAsInterface`, `ReturnError`, `Struct`, `StructPointer`, `TwoDeps`, `ValueChain`, `ValueConversion`, `ValueIsStruct`, `VarValue`, `Varargs`

---

## Part 3: Error Cases Without Dedicated Test Coverage

### High Priority (Common Scenarios)

1. **Provider Function Return Type Validation**
   - Function with 0 returns
   - Function with 4+ returns
   - Function with wrong 2nd/3rd return types
   - **Severity**: HIGH - Core validation logic
   - **Recommendation**: Add test cases like `ProviderNoReturn`, `ProviderTooManyReturns`, `ProviderWrongReturnType`

2. **Provider Function Parameter Duplication**
   - Function with duplicate parameter types
   - **Severity**: MEDIUM
   - **Recommendation**: Add test case `ProviderDuplicateParams`

3. **Self-Binding Error**
   - wire.Bind(*I, I) where I is interface
   - **Severity**: LOW - Rare scenario
   - **Recommendation**: Add test case `BindSelfInterface`

### Medium Priority

4. **wire.Value Errors**
   - Call with wrong number of arguments
   - Argument is too complex expression
   - **Severity**: MEDIUM
   - **Recommendation**: Add test cases `ValueWrongArgCount`, `ValueComplexExpr`

5. **wire.FieldsOf Errors**
   - No field names specified
   - Wrong first argument type
   - Field count exceeds struct fields
   - Wrong field specifications
   - **Severity**: MEDIUM
   - **Recommendation**: Add test cases `FieldsOfNoFields`, `FieldsOfWrongType`, `FieldsOfTooMany`, `FieldsOfNotFound`

6. **Struct Provider Field Duplication**
   - Multiple fields of same type in wire.Struct()
   - **Severity**: MEDIUM
   - **Recommendation**: Add test case `StructDuplicateFieldTypes`

### Low Priority (Edge Cases)

7. **Output Directory Errors**
   - Package with no Go files
   - Package files in different directories
   - **Severity**: LOW - Infrastructure level
   - **Recommendation**: Add test case `NoGoFiles`, `ConflictingDirectories`

8. **Bind Second Argument Error**
   - Concrete type is neither pointer nor **
   - **Severity**: LOW
   - **Recommendation**: Add test case `BindInvalidConcreteType`

---

## Part 4: Error Message Location Reference

### By Source File

#### `parse.go` (52 error messages/locations)
- Variable/object resolution: get() (line 485, 498)
- Unknown pattern handling: processExpr() (lines 536, 540, 543, 580, 590)
- Function provider: processFuncProvider() (lines 681, 700)
- Function output validation: funcOutput() (lines 726, 737, 741, 744, 752)
- Struct literal provider: processStructLiteralProvider() (lines 766, 791)
- Struct provider: processStructProvider() (lines 805, 812, 858)
- wire.Bind: processBind() (lines 888, 896, 902, 910, 916, 920)
- wire.Value: processValue() (lines 934, 959, 964)
- wire.InterfaceValue: processInterfaceValue() (lines 979, 984, 989, 993)
- wire.FieldsOf: processFieldsOf() (lines 1009, 1016, 1037)
- Field specification: checkField() (lines 1068, 1073, 1078)
- Injector validation: findInjectorBuild() (line 1131)

#### `analyze.go` (14 error messages/locations)
- Dependency solving: solve() (lines 129, 134, 138)
- Provider map building: buildProviderMap() (lines 347, 359, 375, 385, 395, 420)
- Cyclic check: verifyAcyclic() (line 492)
- Argument usage: verifyArgsUsed() (lines 272, 274, 287, 299, 311, 323)

#### `wire.go` (13 error messages/locations)
- Output directory: detectOutputDir() (lines 126, 131)
- Injector generation: generateInjectors() (lines 167, 169)
- Inject function: inject() (lines 314, 322, 324, 340, 346, 354)
- Value accessibility: accessibleFrom() (lines 944, 948)

#### `errors.go` (0 error messages - only error infrastructure)
- Error collection and position tracking

---

## Part 5: Error Detection Workflow

### When Errors Are Detected

1. **Parse Time** (Wire analyzes user's provider definitions)
   - Variable type checking
   - Function signature validation
   - wire.Bind/Value/FieldsOf argument validation
   - Struct type checking

2. **Build Time** (Wire constructs provider graph)
   - Provider map construction
   - Binding conflict detection
   - Concrete type availability checking

3. **Analysis Time** (Wire solves dependency graph)
   - Cycle detection (verifyAcyclic)
   - Missing provider detection (solve)
   - Unused resource detection (verifyArgsUsed)

4. **Code Generation Time** (Wire generates injector code)
   - Cleanup/error return type matching
   - Value visibility checking (accessibleFrom)
   - Import accessibility

5. **Compile Time** (Go compiler checks generated code)
   - Type errors in generated code
   - Function signature errors
   - Export/visibility errors

---

## Part 6: Error Context Information

Wire captures rich context for errors:

1. **Position Information**
   - File path and line number via `notePosition()` and `wireErr` type
   - Source code locations for providers and uses

2. **Dependency Chain Information**
   - "needed by X in Y" chain for missing providers
   - Cycle paths showing the full cycle

3. **Type Information**
   - Type strings using `types.TypeString()`
   - Concrete vs. interface types
   - Package paths

4. **Source Identification**
   - Provider name and location
   - Variable name
   - wire.Bind/Value/FieldsOf call location
   - Provider set name (if named)

---

## Recommendations

### For Test Coverage
1. Add tests for provider function signature errors (high priority)
2. Add tests for wire.Value/FieldsOf argument validation (medium priority)
3. Add tests for edge cases like self-binding (low priority)

### For Error Detection Enhancement
1. Consider validating provider structs at analysis time rather than code generation time
2. Add more granular error types vs. current string-based messages
3. Consider batching similar errors in reporting

### For Documentation
1. Create error reference guide mapping all error messages to causes
2. Document error detection workflow and when users should expect each error
3. Provide examples of how to fix each error category

