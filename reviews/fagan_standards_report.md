# Fagan Inspection: Standards Report

**Inspector:** Standards Checker
**Date:** 2026-02-09
**Scope:** All `.go` files in `internal/`, `main.go`, and `go.mod`
**Files Inspected:** 38 files across 6 packages + tools

---

## Executive Summary

The codebase demonstrates strong overall adherence to Go and Terraform plugin framework standards. It is well-structured with clean package boundaries, consistent error handling, and thorough compile-time interface checks. The findings below are predominantly Minor or Style; only a handful of issues reach Major severity.

**Totals:** 5 Major, 12 Minor, 7 Style

---

## 1. Go Standards

### 1.1 Godoc Comments

| ID | Severity | File | Finding |
|----|----------|------|---------|
| STD-001 | Style | `internal/resource_form/crud_create.go:18` | Exported method `Create` has no godoc comment. |
| STD-002 | Style | `internal/resource_form/crud_read.go:17` | Exported method `Read` has no godoc comment. |
| STD-003 | Style | `internal/resource_form/crud_update.go:17` | Exported method `Update` has no godoc comment. |
| STD-004 | Style | `internal/resource_form/crud_delete.go:16` | Exported method `Delete` has no godoc comment. |

**Note:** All other exported types and functions are properly documented. The CRUD methods are interface implementations, so the omission is understandable but technically non-compliant with `golint`/`revive`.

### 1.2 Error Messages

| ID | Severity | File | Finding |
|----|----------|------|---------|
| STD-005 | Minor | `internal/resource_form/crud_create.go:45` | Diagnostic error summaries use Title Case (`"Error Creating Google Form"`) -- this is correct Terraform convention, but the `detail` strings mix sentence case and format-string style. E.g., line 45 uses `"Could not create form: %s"` while line 128 uses `"Form was created (ID: %s) but batchUpdate failed: %s"`. These detail messages are acceptable but should be normalized for user-facing consistency. |

The Go-level error messages in the `client` package correctly use lowercase (`"building token source: ..."`, `"creating forms service: ..."`) -- this is idiomatic Go.

### 1.3 Context Propagation

All functions receiving `context.Context` place it as the first parameter. No instances of dropped contexts found. The retry logic in `client/retry.go` correctly checks `ctx.Err()` before each attempt and uses `sleepWithContext` for cancellation-aware waiting.

**Verdict: PASS**

### 1.4 Error Wrapping with `%w`

| ID | Severity | File | Finding |
|----|----------|------|---------|
| STD-006 | Minor | `internal/client/retry.go:113` | `wrapContextError` wraps `ctxErr` with `%w` but formats `lastErr` with `%v`. This means the last API error is not unwrappable from the returned error. While this is likely intentional (the context error is the primary cause), it means callers cannot `errors.As` the original API error after context cancellation. |

### 1.5 Unused Imports/Variables

| ID | Severity | File | Finding |
|----|----------|------|---------|
| STD-007 | Style | `internal/resource_form/plan_modifiers_test.go:183` | `var _ attr.Value = types.StringNull()` exists solely to suppress an unused import warning for `attr`. The `attr` package is imported but never used by actual test code; the `planmodifier.String` interface check does not require it. This is a workaround that should be cleaned up. |
| STD-008 | Minor | `internal/client/retry_test.go:314` | `_ = gap2` exists only to avoid an "unused variable" error. The variable `gap2` is declared but not meaningfully asserted on. |

### 1.6 Variable Naming

All variable names are idiomatic Go. Abbreviations like `cfg`, `ctx`, `resp`, `req`, `mc`, `sa`, `opts` are standard in the Terraform provider ecosystem. No issues found.

**Verdict: PASS**

---

## 2. Terraform Plugin Framework Standards

### 2.1 Schema Attribute Types and Markings

| ID | Severity | File | Finding |
|----|----------|------|---------|
| STD-009 | **Major** | `internal/resource_form/schema.go:33-38` | `id` attribute is `Computed: true` with `UseStateForUnknown` -- **correct**. |
| STD-010 | **Pass** | `internal/resource_form/schema.go:73-79` | `responder_uri` is `Computed: true` with `UseStateForUnknown` -- **correct**. |
| STD-011 | **Pass** | `internal/resource_form/schema.go:80-86` | `edit_uri` is `Computed: true` with `UseStateForUnknown` -- **correct**. |
| STD-012 | **Pass** | `internal/resource_form/schema.go:87-93` | `document_title` is `Computed: true` with `UseStateForUnknown` -- **correct**. |

All computed attributes are correctly marked. No attributes are incorrectly marked as both `Required` and `Computed`.

### 2.2 Sensitive Attribute Marking

| ID | Severity | File | Finding |
|----|----------|------|---------|
| STD-013 | **Pass** | `internal/provider/provider.go:62` | `credentials` is marked `Sensitive: true` -- **correct**. |

### 2.3 Plan Modifiers and Validators

| ID | Severity | File | Finding |
|----|----------|------|---------|
| STD-014 | **Pass** | `internal/resource_form/schema.go` | All computed ID-like attributes use `stringplanmodifier.UseStateForUnknown()`. |
| STD-015 | **Pass** | `internal/resource_form/plan_modifiers.go` | `ContentJSONHashModifier` correctly implements `planmodifier.String` with compile-time check. |
| STD-016 | **Pass** | `internal/resource_form/resource.go:14-19` | Compile-time checks for `Resource`, `ResourceWithImportState`, and `ResourceWithConfigValidators`. |

### 2.4 Diagnostics Usage

All error conditions use `resp.Diagnostics.AddError()` for fatal errors. The `tflog.Warn` in `crud_read.go:38` and `crud_delete.go:38` is appropriate for non-fatal 404 conditions. No misuse of `AddWarning` for fatal errors or vice versa.

**Verdict: PASS**

### 2.5 Required Interfaces

| ID | Severity | File | Finding |
|----|----------|------|---------|
| STD-017 | Minor | `internal/provider/provider.go:22` | Provider implements `provider.Provider`. No `provider.ProviderWithValidateConfig` is implemented, which means provider-level config validation relies solely on schema constraints. This is acceptable given the current simple schema. |

---

## 3. Consistency

### 3.1 Error Message Format Consistency Across CRUD

| ID | Severity | File | Finding |
|----|----------|------|---------|
| STD-018 | Minor | Various CRUD files | Error diagnostic summaries follow a consistent pattern (`"Error Creating/Reading/Updating/Deleting Google Form"`) -- **good**. However, the detail message format varies: |
| | | `crud_create.go:46` | `"Could not create form: %s"` |
| | | `crud_create.go:128` | `"Form was created (ID: %s) but batchUpdate failed: %s"` |
| | | `crud_read.go:47` | `"Could not read form %s: %s"` |
| | | `crud_update.go:135` | `"BatchUpdate failed for form %s: %s"` |
| | | `crud_delete.go:46` | `"Could not delete form %s: %s"` |

The variation is reasonable given the different contexts, but the Create details could be made slightly more consistent.

### 3.2 Logging Consistency

| ID | Severity | File | Finding |
|----|----------|------|---------|
| STD-019 | Minor | CRUD files | Logging levels are used consistently across operations: `tflog.Debug` for entry/progress, `tflog.Info` for significant actions (create/delete), `tflog.Warn` for 404 removals. Good pattern. |

### 3.3 Function Signatures

All CRUD methods follow the standard Terraform plugin framework signature: `(ctx context.Context, req *Request, resp *Response)`. All validators follow the `resource.ConfigValidator` interface. Consistent.

**Verdict: PASS**

### 3.4 Convert Function Naming

| ID | Severity | File | Finding |
|----|----------|------|---------|
| STD-020 | Minor | `internal/resource_form/state_convert.go` | Conversion functions use mixed naming conventions: `tfItemsToConvertItems`, `convertFormModelToTFState`, `convertItemModelToTF`, `convertGradingToTF`. The prefix/suffix pattern switches between `tf...ToConvert...` and `convert...ToTF...`. These should consistently use one pattern (e.g., `convertTFItemsToModels` / `convertModelToTFState`). |

---

## 4. Security

### 4.1 Credentials Handling

| ID | Severity | File | Finding |
|----|----------|------|---------|
| STD-021 | **Pass** | `internal/provider/provider.go:62` | Credentials marked `Sensitive: true` in schema -- logs won't leak them. |
| STD-022 | **Pass** | `internal/provider/provider.go:95` | Credentials presence logged as boolean (`has_credentials: true/false`), not the actual value. |
| STD-023 | Minor | `internal/provider/provider.go:147` | `resolveCredentialValue` reads a file from a user-supplied path via `os.ReadFile(trimmed)`. The path is sanitized only by `strings.TrimSpace`. There is no path traversal protection. However, this is standard practice in Terraform providers (the Google provider does the same) and the credentials value comes from the Terraform config, which is trusted input. **Acceptable risk** for a provider. |

### 4.2 Input Validation

| ID | Severity | File | Finding |
|----|----------|------|---------|
| STD-024 | **Major** | `internal/resource_form/schema.go:113-116` | `item_key` description says format is `[a-z][a-z0-9_]{0,63}` but **no validator enforces this format**. A malicious or accidental `item_key` value could contain special characters. While this does not create a security vulnerability (the key is only used internally for state mapping), the documented contract is not enforced. |

### 4.3 Content JSON Injection

| ID | Severity | File | Finding |
|----|----------|------|---------|
| STD-025 | Minor | `internal/convert/json_mode.go:17` | `ParseDeclarativeJSON` deserializes arbitrary JSON into `[]*forms.Item`. The JSON is then sent directly to the Google Forms API. This means a user could craft JSON with any valid Forms API fields, including those not exposed by the HCL schema. This is by design (the `content_json` attribute is a power-user escape hatch), but it bypasses all validators (grading, quiz mode checks, etc.). This should be documented as a known limitation. |

---

## 5. Cross-Cutting Concerns

### 5.1 Context Propagation

All API calls pass the `ctx` parameter through. The retry logic respects context cancellation. No dropped contexts found.

**Verdict: PASS**

### 5.2 Resource Cleanup

| ID | Severity | File | Finding |
|----|----------|------|---------|
| STD-026 | **Pass** | `internal/resource_form/crud_create.go:57-62` | Partial state save after Create ensures the form ID is persisted before batchUpdate. If batchUpdate fails, the form can still be destroyed on subsequent `terraform destroy`. |
| STD-027 | **Pass** | `internal/resource_form/crud_delete.go:37` | Delete handles 404 as success, preventing orphaned state entries. |

### 5.3 Compile-Time Interface Checks

| ID | Severity | File | Finding |
|----|----------|------|---------|
| STD-028 | **Pass** | `internal/resource_form/resource.go:15-19` | `FormResource` checks: `Resource`, `ResourceWithImportState`, `ResourceWithConfigValidators`. |
| STD-029 | **Pass** | `internal/client/forms_api.go:28` | `FormsAPIClient` checks: `FormsAPI`. |
| STD-030 | **Pass** | `internal/client/drive_api.go:27` | `DriveAPIClient` checks: `DriveAPI`. |
| STD-031 | **Pass** | `internal/resource_form/plan_modifiers.go:15` | `ContentJSONHashModifier` checks: `planmodifier.String`. |
| STD-032 | **Pass** | `internal/resource_form/validators.go:16-23` | All 7 validators check: `resource.ConfigValidator`. |
| STD-033 | **Pass** | `internal/testutil/mock_forms.go:22` | `MockFormsAPI` checks: `FormsAPI`. |
| STD-034 | **Pass** | `internal/testutil/mock_drive.go:17` | `MockDriveAPI` checks: `DriveAPI`. |

All implementations have compile-time interface checks. Excellent.

### 5.4 Error Wrapping Consistency

The `client` package consistently uses `fmt.Errorf("description: %w", err)` for error wrapping. The CRUD layer uses `resp.Diagnostics.AddError()` with `fmt.Sprintf` for detail messages, which is the correct Terraform pattern. Consistent.

**Verdict: PASS**

---

## 6. Package Dependencies

### 6.1 go.mod Analysis

```
go 1.22

require (
    github.com/hashicorp/terraform-plugin-framework v1.12.0
    github.com/hashicorp/terraform-plugin-go v0.25.0
    github.com/hashicorp/terraform-plugin-log v0.9.0
    github.com/hashicorp/terraform-plugin-testing v1.10.0
    golang.org/x/oauth2 v0.24.0
    google.golang.org/api v0.209.0
)
```

| ID | Severity | File | Finding |
|----|----------|------|---------|
| STD-035 | **Major** | `go.mod` | **Missing `require` block for indirect dependencies.** The `go.mod` file only lists direct dependencies but has no indirect dependency entries. A real module would have a `require ( ... ) // indirect` section generated by `go mod tidy`. This indicates the module has never been built/compiled. |
| STD-036 | **Major** | `go.mod` | **Missing `go.sum` file.** No `go.sum` file exists. This is required for reproducible builds and dependency verification. Must be generated by running `go mod tidy`. |
| STD-037 | Minor | `go.mod` | The `go 1.22` directive is acceptable. All dependencies are at stable, release versions. No pre-release or fork versions. |

### 6.2 Internal Package Boundaries

| ID | Severity | File | Finding |
|----|----------|------|---------|
| STD-038 | **Pass** | Package structure | Clean dependency graph: `main` -> `provider` -> `resource_form` + `client`, `resource_form` -> `convert` + `client`, `convert` uses only `google.golang.org/api/forms/v1` and stdlib. No circular dependencies. |
| STD-039 | **Pass** | `internal/convert/types.go` | Correctly uses plain Go types to avoid circular dependency with Terraform framework types. The `resource_form/state_convert.go` handles the TF-type conversion. |
| STD-040 | Minor | `internal/testutil/sweeper.go` | Sweeper is a TODO stub with no actual implementation. It should either be implemented or removed to avoid dead code. |

### 6.3 Duplicate Logic

| ID | Severity | File | Finding |
|----|----------|------|---------|
| STD-041 | **Major** | `internal/provider/provider.go:115-130` and `internal/client/client.go:73-86` | **Duplicate credential resolution logic.** Both `provider.resolveCredentials()` and `client.resolveCredentials()` read the `GOOGLE_CREDENTIALS` env var and resolve credentials. The provider resolves credentials and passes the JSON string to `client.NewClient`, but `client.NewClient` calls its own `resolveCredentials` which **again** checks the env var. This means if the provider passes an empty string (ADC path), the client will still check the env var, potentially using credentials the provider explicitly skipped. This is a logic bug where the credential resolution at two layers could produce inconsistent behavior. |

---

## 7. File-Level Issues

### 7.1 File Size

| ID | Severity | File | Finding |
|----|----------|------|---------|
| STD-042 | Style | `internal/resource_form/crud_test.go` | At 1226 lines, this exceeds the project's own `maxLinesPerFile = 400` limit from `tools/quality/check_file_limits.go`. The test file should be split. |
| STD-043 | Style | `internal/provider/provider_test.go` | At 329 lines, this is within limits. |

### 7.2 Copyright Headers

All 38 `.go` files consistently include the SPDX copyright header. Quality tools use `//go:build ignore` correctly.

**Verdict: PASS**

### 7.3 Package Documentation

| ID | Severity | File | Finding |
|----|----------|------|---------|
| STD-044 | Minor | `internal/resource_form/model.go:5` | Has package doc `// Package resourceform implements...`. |
| STD-045 | **Pass** | `internal/client/interfaces.go:6` | Has package doc `// Package client provides...`. |
| STD-046 | **Pass** | `internal/convert/items_to_requests.go:6` | Has package doc `// Package convert provides...`. |
| STD-047 | **Pass** | `internal/testutil/mock_forms.go:5` | Has package doc `// Package testutil provides...`. |
| STD-048 | Minor | `internal/provider/provider.go` | **No package-level doc comment** on the `provider` package. |

---

## Defect Summary

### Major (5)

| ID | Description | Recommendation |
|----|-------------|----------------|
| STD-024 | `item_key` format documented but not validated | Add a regex validator for `item_key` |
| STD-035 | Missing indirect dependencies in `go.mod` | Run `go mod tidy` when Go is available |
| STD-036 | Missing `go.sum` file | Run `go mod tidy` when Go is available |
| STD-041 | Duplicate credential resolution in provider and client | Remove env-var check from `client.resolveCredentials`; have client trust the string passed by provider |
| STD-042 | `crud_test.go` exceeds 400-line file limit | Split into separate test files per CRUD operation |

### Minor (12)

| ID | Description | Recommendation |
|----|-------------|----------------|
| STD-005 | Diagnostic detail message format inconsistency | Normalize to `"Could not <verb> form %s: %s"` pattern |
| STD-006 | `wrapContextError` does not wrap lastErr with `%w` | Consider wrapping both errors or using `errors.Join` |
| STD-007 | Unused `attr` import kept alive with package-level var | Remove the dummy var or reorganize imports |
| STD-008 | Dummy `_ = gap2` to suppress unused variable | Either assert on gap2 or remove it |
| STD-017 | No `ProviderWithValidateConfig` | Consider implementing for provider-level validation |
| STD-020 | Inconsistent convert function naming | Adopt consistent `convert<Source>To<Dest>` naming |
| STD-023 | No path traversal protection on credential file path | Document that path is from trusted config |
| STD-025 | `content_json` bypasses validators | Document as known limitation |
| STD-037 | Go 1.22 acceptable but check for 1.23+ features | No action needed now |
| STD-040 | Sweeper is a TODO stub | Implement or remove |
| STD-048 | Missing package doc for `provider` package | Add package-level comment |
| STD-018 | Error detail message inconsistency across CRUD | Minor normalization |

### Style (7)

| ID | Description | Recommendation |
|----|-------------|----------------|
| STD-001 | `Create` method missing godoc | Add `// Create ...` comment |
| STD-002 | `Read` method missing godoc | Add `// Read ...` comment |
| STD-003 | `Update` method missing godoc | Add `// Update ...` comment |
| STD-004 | `Delete` method missing godoc | Add `// Delete ...` comment |
| STD-007 | Dummy var for unused import | Clean up |
| STD-008 | Dummy `_ = gap2` assignment | Clean up |
| STD-042 | Test file exceeds project line limit | Split file |

---

## Positive Findings (Notable Good Practices)

1. **Compile-time interface checks** on every implementation (resource, client, mocks, validators, plan modifiers) -- 10/10.
2. **Partial state save** on Create before batchUpdate -- prevents orphaned resources.
3. **404 handling** in Read (RemoveResource) and Delete (treat as success) -- standard Terraform pattern done correctly.
4. **Credentials marked Sensitive** and logged as boolean presence only.
5. **Clean package boundaries** -- convert uses plain Go types to avoid circular imports.
6. **Consistent copyright/SPDX headers** on all files.
7. **Thorough test coverage** for validators, plan modifiers, errors, retry logic, and CRUD operations.
8. **Context-aware retry** with jitter and configurable backoff.
9. **Quality tooling** (`check_file_limits.go`, `check_coupling.go`) built into the project.
10. **t.Parallel()** used consistently in tests for parallelism.
