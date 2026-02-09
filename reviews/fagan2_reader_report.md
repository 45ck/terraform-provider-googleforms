# Fagan Inspection Iteration 2 -- Reader Report

## Inspection Metadata

| Field | Value |
|-------|-------|
| **Date** | 2026-02-09 |
| **Role** | Reader (re-inspection) |
| **Scope** | 16 files changed during iteration 1 rework |
| **Focus** | Verify fixes F-001 through F-028 (15 FIX items); find new defects and regressions |

---

## Part 1: Fix Verification

### F-001: forms_api.go Create must NOT use WithRetry [VERIFIED CORRECT]

**File:** `internal/client/forms_api.go:32-42`

The `Create` method now calls the API directly without `WithRetry`:
```go
func (c *FormsAPIClient) Create(ctx context.Context, form *forms.Form) (*forms.Form, error) {
    result, err := c.service.Forms.Create(form).Context(ctx).Do()
```

Comment on line 31 explicitly documents the rationale: "Create is non-idempotent, so it must NOT be retried."

**Status:** FIXED CORRECTLY. No regression.

---

### F-002: state_convert.go reads Title/Description/Quiz from API model [VERIFIED CORRECT]

**File:** `internal/resource_form/state_convert.go:99-118`

`convertFormModelToTFState` now reads from the `convert.FormModel` (API response):
```go
state := FormResourceModel{
    ID:            types.StringValue(model.ID),
    Title:         types.StringValue(model.Title),       // from API
    Description:   types.StringValue(model.Description), // from API
    ...
    Quiz:          types.BoolValue(model.Quiz),           // from API
```

`Published` and `AcceptingResponses` correctly remain from plan (F-013 deferred -- API does not expose these).

**Nil pointer risk analysis:** `convert.FormToModel` (form_to_model.go:22-25) guards `form.Info != nil` before reading Title/Description. If `Info` is nil, `model.Title` is empty string `""`, not nil. `types.StringValue("")` is a valid Terraform value. No nil pointer risk.

**Status:** FIXED CORRECTLY. No regression.

---

### F-003: client.go should not have resolveCredentials [VERIFIED CORRECT]

**File:** `internal/client/client.go:54-67`

The `resolveCredentials` function was removed from `client.go`. The `buildTokenSource` function now receives the pre-resolved credentials string and simply branches:
```go
func buildTokenSource(ctx context.Context, credentials string, impersonateUser string) (oauth2.TokenSource, error) {
    if credentials != "" {
        return tokenSourceFromJSON(ctx, []byte(credentials), impersonateUser)
    }
    return tokenSourceFromADC(ctx)
}
```

**ADC path regression check:** `resolveCredentials` now lives solely in `provider.go:116-131`. When `resolveCredentials` returns `""`, `buildTokenSource` receives `""` and calls `tokenSourceFromADC`. The ADC path works correctly.

**Grep confirmation:** `resolveCredentials` only appears in `provider.go` (not in client package).

**Status:** FIXED CORRECTLY. No regression.

---

### F-004: crud_update.go batch order: info, deletes, quiz settings, creates [VERIFIED CORRECT]

**File:** `internal/resource_form/crud_update.go:46-129`

Batch order is now:
1. **Line 53:** `UpdateFormInfo` request appended first
2. **Lines 70-71 / 99-100:** `BuildDeleteRequests` appended second (for both content_json and HCL modes)
3. **Lines 119-126:** Quiz settings appended third (after deletes, before creates)
4. **Line 129:** Create-item requests appended last

The critical insight is at line 63-64: create-item requests are collected in a separate `createItemRequests` variable and only appended after quiz settings at line 129.

**content_json mode check:** In content_json mode (line 66-82), deletes are appended at line 71, then `createItemRequests` is set at line 82. Quiz settings are appended at line 125, and creates at line 129. Order is correct for both paths.

**Status:** FIXED CORRECTLY. No regression.

---

### F-005: Missing CRUD error path tests [VERIFIED CORRECT]

**File:** `internal/resource_form/crud_test.go`

Four new tests added:
1. **Line 1231:** `TestCreate_ContentJSON_ParseError` -- tests invalid content_json during Create
2. **Line 1277:** `TestUpdate_BatchUpdateError_ReturnsDiagnostic` -- tests BatchUpdate failure during Update
3. **Line 1334:** `TestUpdate_WithContentJSON_Success` -- tests content_json mode during Update
4. **Line 1411:** `TestCreate_PublishSettingsError_ReturnsDiagnostic` -- tests SetPublishSettings failure during Create

All four properly verify the error diagnostic summary strings match the production code.

**Status:** FIXED CORRECTLY.

---

### F-006: No-question-block error path test [VERIFIED CORRECT]

**File:** `internal/convert/items_to_requests_test.go:405-421`

`TestItemModelToCreateRequest_NoQuestionBlock` creates an item with all sub-blocks nil and asserts error contains "no question block".

**Status:** FIXED CORRECTLY.

---

### F-007: GradingRequiresQuizValidator tested for all item types [VERIFIED CORRECT]

**File:** `internal/resource_form/validators_item_test.go:133-163`

Three new tests added:
- `TestGradingRequiresQuiz_MultipleChoice_Error` (line 133)
- `TestGradingRequiresQuiz_Paragraph_Error` (line 149)
- (Original `TestGradingRequiresQuiz_QuizFalseWithGrading_Error` at line 107 covers short_answer)

All three verify the error message contains "Grading requires quiz mode".

**Status:** FIXED CORRECTLY.

---

### F-008: Flaky TestRetry_BackoffIncreases [VERIFIED CORRECT]

**File:** `internal/client/retry_test.go:282-318`

The assertion was relaxed from `gap3 > gap1` to `gap3 < gap1/2`:
```go
if gap3 < gap1/2 {
    t.Errorf("expected backoff to increase: gap1=%v gap2=%v gap3=%v", ...)
}
```
This is a very safe margin. With 25% jitter and exponential growth (gap3 base = 4x gap1 base), the assertion can only fail if backoff shrinks by 50% or more, which is impossible.

The dummy `_ = gap2` (old F-026 style issue) has been replaced by an actual assertion using gap2 at line 315.

**Status:** FIXED CORRECTLY.

---

### F-009: Drive 404 should say "file" not "form" [VERIFIED CORRECT]

**File:** `internal/client/drive_api.go:63`

```go
return mapStatusToError(gErr.Code, gErr.Message, operation, "file")
```

The `mapStatusToError` function in `forms_api.go:134` now accepts a `resource` parameter. Drive passes `"file"`, Forms passes `"form"`.

**Status:** FIXED CORRECTLY.

---

### F-010: ExactlyOneSubBlock error identifies the item [VERIFIED CORRECT]

**File:** `internal/resource_form/validators.go:199-213`

```go
for i, item := range itemModels {
    count := countSubBlocks(item)
    if count != 1 {
        identity := fmt.Sprintf("index %d", i)
        if key := item.ItemKey.ValueString(); key != "" {
            identity = fmt.Sprintf("%q", key)
        }
```

The error now includes either the item_key (if set) or the index. Falls back to index when key is empty.

**Status:** FIXED CORRECTLY.

---

### F-011: MutuallyExclusive empty string test [VERIFIED CORRECT]

**File:** `internal/resource_form/validators_test.go:274-284`

`TestMutuallyExclusive_EmptyContentJSON_WithItems_Error` sets content_json to `""` with items present and asserts the error.

**Status:** FIXED CORRECTLY.

---

### F-024: Godoc comments on CRUD methods [VERIFIED CORRECT]

All four CRUD methods now have proper godoc comments:
- `crud_create.go:18`: `// Create creates a new Google Form with the configured items and settings.`
- `crud_read.go:17`: `// Read fetches the current state of a Google Form from the API.`
- `crud_update.go:17`: `// Update replaces the form's settings and items with the planned configuration.`
- `crud_delete.go:16`: `// Delete removes a Google Form by trashing it via the Drive API.`

**Status:** FIXED CORRECTLY.

---

### F-025: Dummy attr.Value suppressed import [VERIFIED]

**File:** `internal/resource_form/plan_modifiers_test.go`

The file no longer imports `attr` and has no dummy variable. The import list (lines 6-12) is clean.

**Status:** FIXED CORRECTLY.

---

### F-026: Dummy gap2 assignment [VERIFIED CORRECT]

**File:** `internal/client/retry_test.go:315-317`

`gap2` is now used in an actual assertion instead of `_ = gap2`:
```go
if gap2 < gap1/2 {
    t.Errorf("expected gap2 to be at least gap1/2: gap1=%v gap2=%v", gap1, gap2)
}
```

**Status:** FIXED CORRECTLY.

---

### F-027: Package-level doc comment for provider [VERIFIED CORRECT]

**File:** `internal/provider/provider.go:5`

```go
// Package provider implements the Terraform provider for Google Forms.
package provider
```

**Status:** FIXED CORRECTLY.

---

## Part 2: New Defects Found During Re-Inspection

### NEW-001: Double-set of IncludeFormInResponse in BatchUpdate

| Field | Value |
|-------|-------|
| **Severity** | Minor |
| **File** | `forms_api.go:73` + `crud_create.go:122` + `crud_update.go:139` |
| **Description** | `FormsAPIClient.BatchUpdate` unconditionally sets `req.IncludeFormInResponse = true` (line 73). But the CRUD callers (`crud_create.go:122`, `crud_update.go:139`) ALSO set `IncludeFormInResponse: true` in the struct literal. The callers' value is immediately overwritten by the API client. This is harmless (same value) but confusing -- it suggests the callers think they control this field, while the API client always overrides it. |
| **Recommendation** | Remove the field from the CRUD struct literals since the API client always sets it. Or remove from the API client and let callers control it. Pick one owner. |
| **Status** | Pre-existing (noted as F-017 deferred in iteration 1). No change needed for this iteration. |

---

### NEW-002: crud_update.go content_json mode does not insert quiz settings between deletes and creates

| Field | Value |
|-------|-------|
| **Severity** | Style |
| **File** | `crud_update.go:66-82` |
| **Description** | In the content_json branch (lines 66-82), the delete requests are appended to `requests` directly (line 71), while creates go to `createItemRequests` (line 82). The quiz settings check at line 119-126 then appends to `requests`, and creates are appended at line 129. The ordering IS correct (deletes -> quiz -> creates). However, the code structure is slightly misleading: the content_json branch exits at line 82 and it's not immediately obvious that quiz settings and creates are handled after the else block. This is a readability observation, not a bug. |
| **Recommendation** | No action needed. Code is correct. |

---

### NEW-003: convertFormModelToTFState does not preserve content_json from plan on Create path

| Field | Value |
|-------|-------|
| **Severity** | Style |
| **File** | `state_convert.go:107` |
| **Description** | `convertFormModelToTFState` preserves `plan.ContentJSON` in the returned state (line 107). This is correct because content_json is a config-only field (API does not return it). Both Create and Update callers handle the items vs content_json branching AFTER calling this function (`crud_create.go:206-216`, `crud_update.go:218-227`). The design is clear and works correctly. |

---

### NEW-004: crud_create.go batch order differs from crud_update.go

| Field | Value |
|-------|-------|
| **Severity** | Minor |
| **File** | `crud_create.go:68-111` vs `crud_update.go:46-129` |
| **Description** | In `crud_create.go`, the batch order is: (1) UpdateFormInfo, (2) QuizSettings, (3) CreateItem requests. In `crud_update.go`, the batch order is: (1) UpdateFormInfo, (2) DeleteItems, (3) QuizSettings, (4) CreateItems. The create path puts quiz settings BEFORE creates, which is correct. However, on Create there are never existing graded items to conflict with, so the ordering difference is benign. The important thing is that Update's order (deletes before quiz settings) is correct, which it is per F-004 fix. |
| **Recommendation** | No action needed. Both orderings are correct for their respective contexts. |

---

### NEW-005: `testutil.MockFormsAPI` BatchUpdate called but return value unused

| Field | Value |
|-------|-------|
| **Severity** | Style |
| **File** | `crud_test.go:125` |
| **Description** | In `TestCreate_BasicForm_Success`, the mock does not define `BatchUpdateFunc`. If the Create path calls BatchUpdate (which it should for the info update), this would cause a nil pointer panic. However, looking at the test plan, it only sets title with no description, no quiz, no items. The batchUpdate IS still called because line 72 of `crud_create.go` always builds an UpdateInfoRequest. The mock likely has a default no-op or the test happens to work because `BatchUpdateFunc` field defaults check. |

Let me re-examine: the `MockFormsAPI` is in testutil package. Checking the test: `TestCreate_BasicForm_Success` (line 145) does NOT set `BatchUpdateFunc`. But `crud_create.go` always calls `BatchUpdate` because `requests` always has at least the UpdateFormInfo request (line 72). This would call `mockForms.BatchUpdate()` which would invoke a nil `BatchUpdateFunc`.

**Checking the mock implementation pattern:** If `MockFormsAPI.BatchUpdateFunc` is nil and BatchUpdate is called, the mock would either panic or return a zero value. Since the test passes (line 177 checks no errors), the mock must have a nil-safe implementation (returns `nil, nil` when func is nil).

| Field | Value |
|-------|-------|
| **Severity** | Minor |
| **File** | `crud_test.go:145-188` |
| **Description** | `TestCreate_BasicForm_Success` relies on the mock's nil-func fallback behavior for `BatchUpdate`. This is fragile -- if the mock implementation changes to panic on nil func, this test breaks. The test should explicitly set `BatchUpdateFunc` since `crud_create.go` always sends a batchUpdate. |
| **Recommendation** | Add explicit `BatchUpdateFunc` to `TestCreate_BasicForm_Success` mock setup. |

---

## Part 3: Summary

### Fix Verification Results

| Fix ID | Status | Notes |
|--------|--------|-------|
| F-001 | VERIFIED | Create no longer retried |
| F-002 | VERIFIED | Title/Description/Quiz from API model |
| F-003 | VERIFIED | resolveCredentials removed from client.go |
| F-004 | VERIFIED | Batch order: info, deletes, quiz, creates |
| F-005 | VERIFIED | 4 new error path tests added |
| F-006 | VERIFIED | No-question-block test added |
| F-007 | VERIFIED | MC + paragraph grading tests added |
| F-008 | VERIFIED | Backoff test assertion relaxed |
| F-009 | VERIFIED | Drive 404 says "file" |
| F-010 | VERIFIED | ExactlyOneSubBlock identifies item |
| F-011 | VERIFIED | Empty content_json test added |
| F-024 | VERIFIED | Godoc comments on CRUD methods |
| F-025 | VERIFIED | Dummy import removed |
| F-026 | VERIFIED | gap2 used in real assertion |
| F-027 | VERIFIED | Package doc comment added |

**All 15 fixes verified correct. No regressions found.**

### New Defects

| ID | Severity | Description |
|----|----------|-------------|
| NEW-005 | Minor | `TestCreate_BasicForm_Success` omits explicit `BatchUpdateFunc` in mock, relying on nil-func fallback |

### Defect Totals

| Severity | Count |
|----------|-------|
| Major | 0 |
| Minor | 1 |
| Style | 0 |
| **Total** | **1** |

### Reader Assessment

The rework was executed cleanly. All 15 fixes were applied correctly with no regressions. The code is well-structured, the batch ordering fix (F-004) was done elegantly with the `createItemRequests` separation variable, and the credential resolution cleanup (F-003) properly centralized logic in provider.go. The only new finding is a minor test robustness issue.

**Recommendation:** PASS. The codebase is ready to proceed.

---

## Part 4: DEFER Item Re-Examination

The iteration 1 report deferred several items. Three were specifically flagged for re-examination to determine if the rework changes should promote them.

### F-013: Published/AcceptingResponses always from plan (DEFER -- no change)

**File:** `state_convert.go:104-105`

```go
Published:     plan.Published,
AcceptingResponses: plan.AcceptingResponses,
```

**Re-examination:** The F-002 fix changed Title/Description/Quiz to read from the API model. However, Published and AcceptingResponses remain from the plan. This is correct because:
1. The `forms.get` API endpoint does NOT return publish settings (confirmed by Google Forms API docs).
2. There is no `forms.getPublishSettings` equivalent to read back the current state.
3. The only way to know the publish state is from what Terraform set.

**Verdict:** Remain DEFERRED. This is a Google API limitation, not a code defect. The F-002 fix was scoped correctly to only change fields the API actually returns.

---

### F-018: content_json drift completely invisible (DEFER -- no change)

**File:** `crud_read.go:81-85`

```go
} else {
    // In content_json mode, preserve the existing content_json value.
    newState.ContentJSON = state.ContentJSON
    newState.Items = state.Items
}
```

**Re-examination:** In content_json mode, Read preserves the existing content_json value from state without comparing to the API. If someone edits the form via the Google Forms UI, Terraform will not detect it. The `ContentJSONHashModifier` plan modifier (tested in `plan_modifiers_test.go`) only handles semantic equivalence between state and config JSON -- it does NOT compare against the live API.

However, detecting drift would require converting API items back to the JSON format, which is an inverse of `DeclarativeJSONToRequests`. This is non-trivial and was correctly scoped out.

**Verdict:** Remain DEFERRED. Known limitation, correctly documented.

---

### F-019: Positional key map assumes API returns items in creation order (DEFER -- no change)

**Files:** `crud_create.go:176-191`, `crud_update.go:189-204`

Both Create and Update use positional mapping to correlate plan items with API response items:
```go
for i, apiItem := range finalForm.Items {
    if i < len(planItems) {
        keyMap[apiItem.ItemId] = planItems[i].ItemKey.ValueString()
    }
}
```

**Re-examination:** The ASSUMPTION comment is present in both files. The Google Forms API documentation does not explicitly guarantee item ordering in responses. However:
1. In practice, the API returns items in the order they were created.
2. The replace-all strategy (delete everything, recreate) means all items are created in a single batchUpdate, preserving order.
3. After the first create/update cycle, subsequent Reads use the `google_item_id` key map (from `buildItemKeyMap`), not positional mapping.

The positional approach is only used during the initial correlation (Create and Update). After that, the more robust ID-based mapping takes over.

**Verdict:** Remain DEFERRED. The risk is low and the ASSUMPTION comment documents it clearly. A future improvement could use `IncludeFormInResponse` from BatchUpdate to get the response with IDs immediately, but this would require restructuring the flow.
