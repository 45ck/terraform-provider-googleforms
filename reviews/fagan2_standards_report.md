# Fagan Inspection Iteration 2: Standards Report

**Inspector:** Standards Checker
**Date:** 2026-02-09
**Scope:** Re-inspection of 16 reworked files for standards compliance
**Focus:** NEW/changed code only; iteration 1 DEFER items not re-reported

---

## Checklist Results

### 1. Godoc Comments Follow Go Conventions

**Verdict: PASS**

All iteration 1 findings (F-024) have been resolved:

| File | Method | Status |
|------|--------|--------|
| `crud_create.go:18` | `// Create creates a new Google Form with the configured items and settings.` | Correct. Starts with function name. |
| `crud_read.go:17` | `// Read fetches the current state of a Google Form from the API.` | Correct. |
| `crud_update.go:17` | `// Update replaces the form's settings and items with the planned configuration.` | Correct. |
| `crud_delete.go:16` | `// Delete removes a Google Form by trashing it via the Drive API.` | Correct. |

Additionally:
- `provider.go:4` now has a package-level doc: `// Package provider implements the Terraform provider for Google Forms.` (fixes F-027).
- All new/changed functions (`resolveCredentials`, `resolveCredentialValue`, `buildTokenSource`, `tokenSourceFromJSON`, `tokenSourceFromADC`, `mapStatusToError`, `wrapDriveAPIError`) have proper godoc comments starting with the function name.

No new godoc violations found.

---

### 2. Error Messages Consistent with Existing Patterns

**Verdict: PASS**

Error message patterns across the reworked files are consistent:

- **Go-level errors (client package):** Lowercase, wrapping with `%w`. Examples:
  - `"building token source: %w"` (client.go:33)
  - `"parsing service account credentials: %w"` (client.go:77)
  - `"drive.Delete: %w"` (drive_api.go:49)

- **Terraform diagnostics (resource_form package):** Title Case summaries with detail messages. Examples:
  - `"Error Creating Google Form"` / `"Could not create form: %s"`
  - `"Error Updating Google Form"` / `"BatchUpdate failed for form %s: %s"`
  - `"Error Setting Publish Settings"` / `"Could not update publish settings for form %s: %s"`

- The `mapStatusToError` now correctly parameterizes the `resource` field (F-009 fix):
  - Forms operations pass `"form"` (forms_api.go:129)
  - Drive operations pass `"file"` (drive_api.go:63)

No new inconsistencies introduced.

---

### 3. Imports Necessary and Correctly Ordered

**Verdict: PASS**

All 16 files follow standard Go import grouping (stdlib, external, internal):

| File | Import Groups | Status |
|------|---------------|--------|
| `client.go` | stdlib (`context`, `fmt`) / external (`golang.org/x/oauth2`, `google.golang.org/api`) | Correct |
| `drive_api.go` | stdlib (`context`, `errors`, `fmt`) / external (`google.golang.org/api`) | Correct |
| `forms_api.go` | stdlib (`context`, `errors`, `fmt`, `net/http`) / external (`google.golang.org/api`) | Correct |
| `retry_test.go` | stdlib (`context`, `fmt`, `sync/atomic`, `testing`, `time`) | Correct |
| `items_to_requests_test.go` | stdlib (`strings`, `testing`) / external (`google.golang.org/api/forms/v1`) | Correct |
| `provider.go` | stdlib (`context`, `os`, `strings`) / external (hashicorp) / internal | Correct |
| `crud_create.go` | stdlib (`context`, `fmt`) / external (hashicorp, google) / internal | Correct |
| `crud_delete.go` | stdlib (`context`, `fmt`) / external (hashicorp) / internal | Correct |
| `crud_read.go` | stdlib (`context`, `fmt`) / external (hashicorp) / internal | Correct |
| `crud_update.go` | stdlib (`context`, `fmt`) / external (hashicorp, google) / internal | Correct |
| `plan_modifiers_test.go` | stdlib (`context`, `testing`) / external (hashicorp) | Correct |
| `state_convert.go` | stdlib (`context`) / external (hashicorp) / internal | Correct |
| `validators.go` | stdlib (`context`, `fmt`) / external (hashicorp) | Correct |
| `validators_item_test.go` | stdlib (`context`, `testing`) / external (hashicorp) | Correct |
| `validators_test.go` | stdlib (`context`, `strings`, `testing`) / external (hashicorp) | Correct |
| `crud_test.go` | stdlib (`context`, `fmt`, `strings`, `testing`) / external (hashicorp, google) / internal | Correct |

**F-025 fix verified:** `plan_modifiers_test.go` no longer has the dummy `var _ attr.Value` import workaround. The `attr` import has been removed entirely.

No unused imports detected. All imports are consumed by code in the file.

---

### 4. Test Function Names Follow Conventions

**Verdict: PASS**

All test functions follow the `Test<Type>_<Scenario>_<Expected>` or `Test<Type>_<Scenario>` convention:

**New tests added (F-005, F-006, F-007, F-008, F-011):**

| File | Test Function | Convention |
|------|---------------|------------|
| `crud_test.go` | `TestCreate_ContentJSON_ParseError` | Correct |
| `crud_test.go` | `TestUpdate_BatchUpdateError_ReturnsDiagnostic` | Correct |
| `crud_test.go` | `TestUpdate_WithContentJSON_Success` | Correct |
| `crud_test.go` | `TestCreate_PublishSettingsError_ReturnsDiagnostic` | Correct |
| `crud_test.go` | `TestCreate_WithQuizGrading_Success` | Correct |
| `crud_test.go` | `TestCreate_WithPublishSettings_Success` | Correct |
| `crud_test.go` | `TestUpdate_PublishSettingsChanged_Success` | Correct |
| `crud_test.go` | `TestUpdate_APIError_ReturnsDiagnostic` | Correct |
| `crud_test.go` | `TestRead_WithItems_CorrectMapping` | Correct |
| `items_to_requests_test.go` | `TestItemModelToCreateRequest_NoQuestionBlock` | Correct (F-006) |
| `validators_item_test.go` | `TestGradingRequiresQuiz_MultipleChoice_Error` | Correct (F-007) |
| `validators_item_test.go` | `TestGradingRequiresQuiz_Paragraph_Error` | Correct (F-007) |
| `validators_test.go` | `TestMutuallyExclusive_EmptyContentJSON_WithItems_Error` | Correct (F-011) |

All new tests use `t.Parallel()` where appropriate. Helper functions like `testResource`, `buildPlan`, `buildState`, `emptyState` use `t.Helper()` correctly.

---

### 5. Security Concerns from Credential Resolution Changes

**Verdict: PASS -- No new security issues**

The F-003 fix restructured credential resolution:

- **Before:** Both `provider.go` and `client.go` independently checked `GOOGLE_CREDENTIALS` env var, creating a dual-resolution bug.
- **After:** Only `provider.go` resolves credentials (via `resolveCredentials` + `resolveCredentialValue`). `client.go:NewClient` receives the already-resolved JSON string and passes it through to `buildTokenSource`.

Security analysis of the new flow:
1. `resolveCredentials` (provider.go:116): Checks config attr first, then `GOOGLE_CREDENTIALS` env var, then falls back to empty (ADC).
2. `resolveCredentialValue` (provider.go:136): Distinguishes JSON (`{`-prefixed) from file paths, reads file if path.
3. `buildTokenSource` (client.go:57): Trusts the resolved string. Comment explicitly states: "It trusts that the caller (provider.go) has already resolved credentials from config and environment variables."
4. Credential values are never logged (only `has_credentials: true/false` at provider.go:96).
5. `credentials` attribute remains `Sensitive: true` in schema.

No path traversal, credential leakage, or injection risks introduced. The single-layer resolution eliminates the iteration 1 dual-resolution bug.

---

### 6. Batch Reordering Well-Commented

**Verdict: PASS**

The F-004 fix in `crud_update.go` reordered batch operations. The new order is:

```
(1) UpdateFormInfo
(2) Delete existing items (reverse order)
(3) Quiz settings change
(4) Create new items
```

This is well-documented with comments:

- Lines 47-49: `// Order: (1) UpdateFormInfo, (2) delete items, (3) quiz settings, (4) create items.`
- Lines 48-49: `// Item deletes MUST precede quiz settings changes so that disabling quiz`
- Line 49: `// mode does not fail due to still-existing graded items.`
- Lines 62-63: `// Collect create-item requests separately so we can insert quiz settings between deletes and creates.`
- Line 118: `// Update quiz settings if changed (after deletes, before creates).`
- Line 128-129: `// Append create-item requests after quiz settings.`

The ordering rationale is clearly explained. No ambiguity remains.

---

### 7. `mapStatusToError` Change Backward Compatible

**Verdict: PASS**

The F-009 fix added a `resource` parameter to `mapStatusToError`:

**Before:** `mapStatusToError(code int, message, operation string) error`
**After:** `mapStatusToError(code int, message, operation, resource string) error` (forms_api.go:134)

Callers updated:
- `wrapGoogleAPIError` (forms_api.go:129): passes `"form"`
- `wrapDriveAPIError` (drive_api.go:63): passes `"file"`

These are both unexported functions within the `client` package, so no external callers exist. The `NotFoundError` struct field `Resource` is populated with the correct value per API type. This change is fully backward compatible since:
- `mapStatusToError` is unexported
- All callers are within the `client` package
- The `NotFoundError.Resource` field was already in the struct; it now gets a meaningful value instead of a hardcoded `"form"`

The `ErrorStatusCode`, `IsNotFound`, and `IsRateLimit` helper functions are unaffected.

---

### 8. New Unused Imports or Variables

**Verdict: PASS**

- **F-025 verified:** `plan_modifiers_test.go` -- the dummy `attr` import and `var _ attr.Value` are gone. Only `context`, `testing`, and hashicorp framework packages remain, all used.
- **F-026 verified:** `retry_test.go:303-316` -- the `gap2` variable is now actively asserted (`if gap2 < gap1/2`) instead of being suppressed with `_ = gap2`. All three gap variables are used in assertions.
- No new unused imports or variables found across all 16 files.

---

### 9. Compile-Time Interface Checks

**Verdict: PASS**

All existing compile-time interface checks remain intact:

| File | Check |
|------|-------|
| `drive_api.go:27` | `var _ DriveAPI = &DriveAPIClient{}` |
| `forms_api.go:28` | `var _ FormsAPI = &FormsAPIClient{}` |
| `validators.go:16-23` | All 7 validators check `resource.ConfigValidator` |
| `mock_forms.go:22` | `var _ client.FormsAPI = &MockFormsAPI{}` |
| `mock_drive.go:17` | `var _ client.DriveAPI = &MockDriveAPI{}` |
| `plan_modifiers_test.go:177` | `var _ planmodifier.String = ContentJSONHashModifier{}` |

No new implementations introduced without interface checks. No violations.

---

## New Issues Found in Iteration 2

### STD2-001: `crud_test.go` Still Exceeds File Limit (DEFER-carried)

**Severity:** Style
**File:** `crud_test.go` (now ~1467 lines)
**Description:** With the addition of ~240 lines of new tests (F-005), the file has grown further beyond the project's 400-line limit (previously ~1226 lines, tracked as F-023/DEFER). This is not a new defect -- it is the same DEFER item from iteration 1, now slightly worse due to the necessary test additions.
**Disposition:** Remains DEFER (same as iteration 1 F-023).

### STD2-002: `convertFormModelToTFState` Title/Description/Quiz from Model vs Plan

**Severity:** N/A -- Verification item
**Description:** Verified that F-002 fix is correct. `state_convert.go:101-106` now reads:
```go
Title:       types.StringValue(model.Title),
Description: types.StringValue(model.Description),
Quiz:        types.BoolValue(model.Quiz),
```
These values come from `model` (the API response), not `plan`. Drift detection now works for these fields.
**Disposition:** FIX VERIFIED.

---

## Summary

| Checklist Item | Result |
|----------------|--------|
| 1. Godoc comments follow conventions | PASS |
| 2. Error messages consistent | PASS |
| 3. Imports necessary and ordered | PASS |
| 4. Test function names follow conventions | PASS |
| 5. Security from credential changes | PASS |
| 6. Batch reordering well-commented | PASS |
| 7. `mapStatusToError` backward compatible | PASS |
| 8. No new unused imports/variables | PASS |
| 9. Compile-time interface checks | PASS |

**New defects found: 0**
**Existing DEFER items worsened: 1** (STD2-001, file size -- not a new defect)

### Conclusion

All 15 FIX items from iteration 1 that touch the 16 reworked files have been addressed in a standards-compliant manner. No new standards violations have been introduced. The reworked code maintains consistent style, proper godoc documentation, correct import ordering, idiomatic error handling, and sound security practices.

**Recommendation: PASS -- No standards-related rework needed for iteration 2.**

---

*Inspector: Standards Checker (Claude Opus 4.6)*
*Date: 2026-02-09*
