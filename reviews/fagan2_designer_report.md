# Fagan Inspection Iteration 2 -- Designer Verification Report

## Metadata

| Field | Value |
|-------|-------|
| **Role** | Designer (spec compliance verifier) |
| **Date** | 2026-02-09 |
| **Scope** | Verify all 15 FIX-disposition defects from iteration 1 |

---

## Priority 1 -- Correctness Fixes

### F-001: Remove WithRetry from Create (forms_api.go)

**Status: VERIFIED**

- `Create` (line 32-41): Calls `c.service.Forms.Create(form).Context(ctx).Do()` directly -- no `WithRetry` wrapper. Comment on line 31 explicitly states: "Create is non-idempotent, so it must NOT be retried."
- `Get` (line 51): Still uses `WithRetry` -- correct.
- `BatchUpdate` (line 77): Still uses `WithRetry` -- correct.
- `SetPublishSettings` (line 100): Still uses `WithRetry` -- correct.

### F-002: State converter reads Title/Description/Quiz from API model (state_convert.go)

**Status: VERIFIED**

At `state_convert.go:99-118`, `convertFormModelToTFState`:
- Line 102: `Title: types.StringValue(model.Title)` -- reads from `model` (API).
- Line 103: `Description: types.StringValue(model.Description)` -- reads from `model` (API).
- Line 106: `Quiz: types.BoolValue(model.Quiz)` -- reads from `model` (API).
- Line 104-105: `Published` and `AcceptingResponses` still come from `plan.*` -- correct, since the Forms API does not return publish settings via `forms.get`.

### F-003: No duplicate credential resolution in client.go

**Status: VERIFIED**

- `client.go` has no `resolveCredentials` function. The credential resolution chain is:
  - `provider.go:resolveCredentials()` (lines 116-131) resolves credentials from config, env, or ADC.
  - `client.NewClient()` (line 26-52) receives the resolved `credentials` string and passes it to `buildTokenSource()`.
  - `buildTokenSource()` (lines 57-67) trusts the caller: if `credentials != ""`, uses JSON; else uses ADC.
  - Comment on line 55-56 explicitly states: "It trusts that the caller (provider.go) has already resolved credentials from config and environment variables."
- No env-var re-checking in the client package.

### F-004: Batch ordering in crud_update.go -- deletes before quiz settings

**Status: VERIFIED**

At `crud_update.go`, the batch request construction order is:
1. **Info** (line 53-56): `BuildUpdateInfoRequest` appended first.
2. **Deletes** (lines 70-71 for content_json, lines 99-101 for HCL): Delete requests appended to `requests` immediately.
3. **Quiz settings** (lines 119-126): `BuildQuizSettingsRequest` appended after deletes.
4. **Creates** (line 129): `createItemRequests` appended last.

The comment on lines 47-49 explicitly documents: "Order: (1) UpdateFormInfo, (2) delete items, (3) quiz settings, (4) create items. Item deletes MUST precede quiz settings changes so that disabling quiz mode does not fail due to still-existing graded items."

Both HCL and content_json paths follow this order correctly.

---

## Priority 2 -- Test Fixes

### F-005: Four new CRUD error path tests (crud_test.go)

**Status: VERIFIED**

All 4 tests found in `crud_test.go` under the "F-005" section header (line 1228):

1. **TestCreate_ContentJSON_ParseError** (line 1231): Tests Create with invalid `content_json` ("invalid json {{"). Verifies diagnostic with "Error Parsing content_json".
2. **TestUpdate_BatchUpdateError_ReturnsDiagnostic** (line 1277): Tests Update when `BatchUpdate` returns a 500 error. Verifies diagnostic with "Error Updating Google Form".
3. **TestUpdate_WithContentJSON_Success** (line 1334): Tests Update in `content_json` mode. Verifies BatchUpdate is called and completes without errors.
4. **TestCreate_PublishSettingsError_ReturnsDiagnostic** (line 1411): Tests Create when `SetPublishSettings` returns a 500 error. Verifies diagnostic with "Error Setting Publish Settings".

### F-006: No-question-block error test (items_to_requests_test.go)

**Status: VERIFIED**

`TestItemModelToCreateRequest_NoQuestionBlock` (line 405-421):
- Creates an `ItemModel` with all sub-blocks nil (`MultipleChoice: nil, ShortAnswer: nil, Paragraph: nil`).
- Verifies `ItemModelToCreateRequest` returns an error.
- Verifies the error message contains "no question block".

### F-007: MC/paragraph grading tests (validators_item_test.go)

**Status: VERIFIED**

Two new tests found:

1. **TestGradingRequiresQuiz_MultipleChoice_Error** (line 133-147): Creates a config with `quiz=false` and a `multiple_choice` item with grading. Verifies error containing "Grading requires quiz mode".
2. **TestGradingRequiresQuiz_Paragraph_Error** (line 149-163): Creates a config with `quiz=false` and a `paragraph` item with grading. Verifies error containing "Grading requires quiz mode".

Combined with the existing `short_answer` test (`TestGradingRequiresQuiz_QuizFalseWithGrading_Error`, line 107), all 3 item types are now covered.

### F-008: Flaky backoff test fixed (retry_test.go)

**Status: VERIFIED**

`TestRetry_BackoffIncreases` (line 282-318):
- The flaky assertion `gap3 > gap1` has been replaced with a safe margin: `gap3 < gap1/2` (line 309). This is extremely tolerant -- even with 25% jitter, exponential backoff means gap3 will always be well above gap1/2.
- Additionally, `gap2` is now asserted (line 315-317): `gap2 < gap1/2` provides a secondary sanity check.
- The dummy `_ = gap2` from F-026 is also gone -- `gap2` is now used in a real assertion.

### F-011: Empty-string content_json test (validators_test.go)

**Status: VERIFIED**

`TestMutuallyExclusive_EmptyContentJSON_WithItems_Error` (line 274-284):
- Creates a config with `content_json: ""` (empty string, not null) and items.
- Comment explains: "Empty string is not null/unknown, so the validator treats it as 'set'."
- Verifies the "Cannot use both" error is raised.

---

## Priority 3 -- Polish Fixes

### F-009: Resource labels in mapStatusToError (forms_api.go, drive_api.go)

**Status: VERIFIED**

- `forms_api.go:134`: `mapStatusToError` now accepts a `resource` parameter (4th argument). Signature: `func mapStatusToError(code int, message, operation, resource string) error`.
- `forms_api.go:129`: `wrapGoogleAPIError` calls `mapStatusToError(..., "form")`.
- `drive_api.go:63`: `wrapDriveAPIError` calls `mapStatusToError(..., "file")`.
- `forms_api.go:137`: The `NotFoundError` uses `Resource: resource` (parameterized, not hardcoded).

### F-010: Item identity in ExactlyOneSubBlockValidator error (validators.go)

**Status: VERIFIED**

At `validators.go:199-214`:
- Line 202: `identity := fmt.Sprintf("index %d", i)` -- defaults to index.
- Lines 203-205: If `item.ItemKey.ValueString()` is not empty, overrides to `identity = fmt.Sprintf("%q", key)`.
- Line 210: Error message uses `identity`: `"Item %s must have exactly one question type..."`.

### F-024: Godoc comments on CRUD methods (crud_*.go)

**Status: VERIFIED**

- `crud_create.go:18`: `// Create creates a new Google Form with the configured items and settings.`
- `crud_read.go:17`: `// Read fetches the current state of a Google Form from the API.`
- `crud_update.go:17`: `// Update replaces the form's settings and items with the planned configuration.`
- `crud_delete.go:16`: `// Delete removes a Google Form by trashing it via the Drive API.`

All four exported CRUD methods have godoc comments.

### F-025: Dummy var removed in plan_modifiers_test.go

**Status: VERIFIED**

The entire `plan_modifiers_test.go` file (180 lines) has been reviewed. There is no dummy `var _ attr.Value = types.StringNull()` statement. The `attr` package is not imported. The compile-time interface check on line 177 uses a proper pattern: `var _ planmodifier.String = ContentJSONHashModifier{}`.

### F-026: gap2 fixed in retry_test.go

**Status: VERIFIED**

At `retry_test.go:304-317`:
- `gap2` is declared on line 305.
- `gap2` is used in a real assertion on line 315: `if gap2 < gap1/2 { ... }`.
- The dummy `_ = gap2` assignment is gone.

### F-027: Package doc on provider.go

**Status: VERIFIED**

At `provider.go:4`: `// Package provider implements the Terraform provider for Google Forms.`
This is a proper package-level doc comment directly preceding the `package provider` declaration.

---

## Summary

| Defect ID | Description | Verdict |
|-----------|-------------|---------|
| F-001 | Remove WithRetry from Create | **VERIFIED** |
| F-002 | Read Title/Description/Quiz from API model | **VERIFIED** |
| F-003 | Remove duplicate credential resolution | **VERIFIED** |
| F-004 | Batch ordering: deletes before quiz settings | **VERIFIED** |
| F-005 | 4 new CRUD error path tests | **VERIFIED** |
| F-006 | No-question-block error test | **VERIFIED** |
| F-007 | MC/paragraph grading tests | **VERIFIED** |
| F-008 | Flaky backoff test fixed | **VERIFIED** |
| F-011 | Empty-string content_json test | **VERIFIED** |
| F-009 | Resource labels in mapStatusToError | **VERIFIED** |
| F-010 | Item identity in error message | **VERIFIED** |
| F-024 | Godoc comments on CRUD methods | **VERIFIED** |
| F-025 | Dummy var removed in plan_modifiers_test | **VERIFIED** |
| F-026 | gap2 used in real assertion | **VERIFIED** |
| F-027 | Package doc on provider package | **VERIFIED** |

**Result: All 15 FIX-disposition defects are VERIFIED as correctly resolved.**

No regressions detected. No new defects introduced by the fixes.

---

*Designer: Claude Opus 4.6*
*Date: 2026-02-09*
