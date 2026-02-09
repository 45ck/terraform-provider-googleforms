# Sprint 2 Code Review

## Summary

**Overall Assessment: PASS WITH NOTES**

Sprint 2 delivers a well-structured, production-quality Terraform provider resource layer. The code demonstrates strong familiarity with the Terraform Plugin Framework patterns, clean separation of concerns between the `resourceform` and `convert` packages, and careful handling of the Google Forms API's create-then-update workflow. There are no critical blocking issues, but several warnings and notes that should be addressed before release.

---

## Findings

### [GOOD] Clean package architecture avoids circular imports
- **Files**: `internal/convert/types.go`, `internal/resource_form/state_convert.go`
- **Observation**: The `convert` package defines its own plain-Go-type mirrors (`convert.ItemModel`, `convert.GradingBlock`, etc.) rather than importing Terraform framework types. The `resourceform` package does the translation between `types.String`/`types.Bool` and plain Go types in `state_convert.go`. This cleanly avoids circular import risks between the two packages and is a textbook approach for Terraform providers.

### [GOOD] Partial state save on Create
- **File**: `internal/resource_form/crud_create.go:55-62`
- **Observation**: The Create operation correctly saves the form ID to state immediately after the `forms.Create` call (line 59) and BEFORE the `batchUpdate` call. This ensures that if `batchUpdate` fails, the form ID is already in state and Terraform can track (and later destroy) the orphaned resource. This is a critical best practice for Terraform providers.

### [GOOD] 404 handling on Read
- **File**: `internal/resource_form/crud_read.go:36-43`
- **Observation**: Read correctly checks for 404 via `client.IsNotFound(err)` and calls `resp.State.RemoveResource(ctx)` to remove the resource from state. This handles the case where a form is deleted outside of Terraform.

### [GOOD] Delete idempotency
- **File**: `internal/resource_form/crud_delete.go:36-43`
- **Observation**: Delete correctly handles 404 as success, making the delete operation idempotent.

### [GOOD] Replace-all update strategy with reverse deletion
- **Files**: `internal/resource_form/crud_update.go:66-117`, `internal/convert/items_to_requests.go:112-124`
- **Observation**: The update strategy correctly deletes all existing items in reverse index order (highest first) before creating new ones. Reverse-order deletion avoids index-shifting bugs during batch operations. This is a pragmatic approach for V1 of the provider.

### [GOOD] Validator suite is thorough
- **Files**: `internal/resource_form/validators.go`, `internal/resource_form/validators_item.go`
- **Observation**: All 7 validators are well-implemented with proper null/unknown checks. The `CorrectAnswerInOptionsValidator` correctly cross-references grading answers against the options list. The `GradingRequiresQuizValidator` correctly allows null/unknown quiz values to pass (since defaults haven't been applied yet during validation).

### [GOOD] Schema-to-model type alignment
- **Files**: `internal/resource_form/schema.go`, `internal/resource_form/state_convert.go:245-286`
- **Observation**: The `itemObjectType()` function at `state_convert.go:245` correctly mirrors the schema defined in `schema.go`. All attribute names and types match: `item_key` (String), `google_item_id` (String), `multiple_choice` (Object with `question_text`, `options` as `ListType{ElemType: StringType}`, `required`, `grading`), `short_answer` and `paragraph` (Object with `question_text`, `required`, `grading`). The `gradingObjectType()` matches the grading block schema exactly.

### [GOOD] `types.StringType` as ListAttribute ElementType works correctly
- **File**: `internal/resource_form/schema.go:138-140`
- **Observation**: Using `types.StringType` as the `ElementType` for the `options` ListAttribute is correct. `types.StringType` implements `attr.Type` and is the standard way to define a list of strings in the Terraform Plugin Framework.

---

### [WARNING] Positional key mapping assumes API returns items in creation order
- **File**: `internal/resource_form/crud_create.go:173-188`
- **Issue**: When creating a new form, `buildItemKeyMap` returns nil because no items have `google_item_id` set yet. The fallback (lines 173-188) builds a positional key map by iterating `finalForm.Items` and matching by index to `planItems`. This assumes the Google Forms API returns items in the exact order they were created. While this is currently the observed behavior, the Google Forms API documentation does not explicitly guarantee item ordering in the response.
- **Fix**: Document this assumption with a comment. Consider adding a defensive check that logs a warning if `len(finalForm.Items) != len(planItems)` to help diagnose mismatches. The same pattern appears in `crud_update.go:184-189`.

### [WARNING] `NormalizeJSON` does not guarantee key ordering in Go's `encoding/json`
- **Files**: `internal/convert/json_mode.go:47-57`, `internal/resource_form/plan_modifiers.go:58-68`
- **Issue**: Go's `encoding/json` marshals map keys in sorted order (as of Go 1.12+), so `json.Marshal` of an unmarshalled `interface{}` DOES produce sorted keys. However, both `json_mode.go:NormalizeJSON` and `plan_modifiers.go:normalizeJSONHash` rely on this behavior without documenting it. The `NormalizeJSON` function comment says "sorted keys" which is correct but should note the Go version dependency. Additionally, there is duplication: `plan_modifiers.go` has its own `normalizeJSONHash` function that does the same thing as `convert.NormalizeJSON` + `convert.HashJSON`. This duplication could lead to divergent behavior if one is updated but not the other.
- **Fix**: Have `ContentJSONHashModifier.PlanModifyString` call `convert.HashJSON()` instead of the local `normalizeJSONHash`. Remove the duplicated `normalizeJSONHash` and `hashString` functions from `plan_modifiers.go`.

### [WARNING] `convertFormModelToTFState` does not set `Items` field
- **File**: `internal/resource_form/state_convert.go:99-118`
- **Issue**: The `convertFormModelToTFState` function constructs a `FormResourceModel` but never sets the `Items` field. The caller is responsible for setting `state.Items` separately (as seen in `crud_create.go:202-212`, `crud_read.go:73-85`, `crud_update.go:204-213`). This split responsibility is a source of potential bugs -- if a future caller forgets to set Items, they will be silently null.
- **Fix**: Consider having `convertFormModelToTFState` accept an optional items list, or at minimum add a prominent comment on the function documenting that the caller MUST set `Items` separately.

### [WARNING] `EditURI` is computed from ID rather than read from API
- **File**: `internal/resource_form/state_convert.go:112-115`
- **Issue**: The `edit_uri` is constructed as `"https://docs.google.com/forms/d/" + model.ID + "/edit"` rather than being read from the API response. The `convert.FormModel` does not have an `EditURI` field. If Google ever changes the URL format, this would produce incorrect URIs.
- **Fix**: Investigate whether the Google Forms API returns an edit URL. If it does, read it from the API response. If not, this pattern is acceptable but should be documented as a known limitation. Low priority since the URL format has been stable.

### [WARNING] `AcceptingResponsesRequiresPublishedValidator` error message case mismatch
- **File**: `internal/resource_form/validators.go:104-107`
- **Issue**: The error detail says "A form cannot accept responses while unpublished." but the test at `validators_test.go:287` checks for `expectErrorContains(t, diags, "cannot accept responses while unpublished")` with lowercase "c". The `strContains` helper checks both `Summary()` and `Detail()`. The Detail starts with "A form cannot..." so the substring "cannot accept responses while unpublished" IS found, so the test passes. However, this is fragile -- if the message changes slightly, the test could break unexpectedly.
- **Fix**: Consider using a constant or searching for a shorter, more stable substring.

### [WARNING] `strContains` reimplements `strings.Contains`
- **File**: `internal/resource_form/validators_test.go:96-103`
- **Issue**: The test file includes a manual implementation of substring search (`strContains`) instead of using the standard library `strings.Contains`. This is unnecessary and slightly harder to read.
- **Fix**: Replace with `strings.Contains` from the standard library and add `"strings"` to the import block.

### [WARNING] Validators tests missing coverage for validators 5-7
- **File**: `internal/resource_form/validators_test.go`
- **Issue**: The test file covers validators 1-4 (MutuallyExclusive, AcceptingResponsesRequiresPublished, UniqueItemKey, ExactlyOneSubBlock) but does NOT include tests for validators 5 (OptionsRequiredForChoice), 6 (CorrectAnswerInOptions), or 7 (GradingRequiresQuiz). These are declared in `validators_item.go` but have no corresponding test cases.
- **Fix**: Add test cases for all three missing validators. Priority is high for `CorrectAnswerInOptions` since it has cross-field logic.

### [WARNING] `QuestionText` field set on `MultipleChoiceModel` from `item.Title` in `state_convert.go`
- **File**: `internal/resource_form/state_convert.go:56`
- **Issue**: In `tfItemsToConvertItems`, the code sets `result[i].Title = mc.QuestionText` (line 56). The `Title` field on `convert.ItemModel` is used as the Forms API `Item.Title`, which is mapped to `QuestionText` in the TF model. When reading back, `form_to_model.go:87` maps `item.Title` -> `MultipleChoiceBlock.QuestionText`. This round-trip works correctly, but it means the `convert.ItemModel.Title` field is an alias for the question text. This is correct but could be confusing to future maintainers.
- **Fix**: Add a comment on `convert.ItemModel.Title` explaining it maps to both the API's `Item.Title` and the TF model's `question_text`.

---

### [NOTE] `ContentJSONHashModifier` does not detect structural changes when content_json changes type
- **File**: `internal/resource_form/plan_modifiers.go:39-52`
- **Issue**: The hash modifier only compares when both state and config values are non-null/unknown. If the user switches from `content_json` to `item` blocks (or vice versa), the modifier correctly skips comparison (lines 40-45) since `content_json` will become null. This is correct behavior. However, there is no plan modifier or validation that detects when a user switches modes between applies. The `MutuallyExclusiveValidator` handles the case where both are set simultaneously, but switching from one to the other between applies could leave stale items in state.
- **Fix**: Consider adding a note in documentation about switching modes. Low priority as the replace-all update strategy handles this correctly.

### [NOTE] `Read` preserves `content_json` from state without drift detection
- **File**: `internal/resource_form/crud_read.go:80-84`
- **Issue**: In `content_json` mode, Read preserves the existing `content_json` value from state (line 83) without comparing it to the actual API response. This means external changes to the form (e.g., someone editing via the Google Forms UI) will NOT be detected as drift. Drift detection relies solely on the hash-based plan modifier comparing the user's config against the stored state value, which remains unchanged.
- **Fix**: Document this as a known limitation of `content_json` mode. Consider reconstructing the JSON from the API response and comparing hashes in a future iteration.

### [NOTE] No `item_key` format validation
- **File**: `internal/resource_form/schema.go:113-115`
- **Issue**: The schema description says `item_key` format should be `[a-z][a-z0-9_]{0,63}` but no validator enforces this regex pattern. Users could provide `item_key = "123-invalid!"` and it would be accepted.
- **Fix**: Add a `stringvalidator.RegexMatches` validator on the `item_key` attribute, or add a dedicated ConfigValidator. Low priority as item_key is an internal identifier.

### [NOTE] Missing error item on `ItemModelToCreateRequest` for no-block items
- **File**: `internal/convert/items_to_requests.go:28`
- **Issue**: The error message uses `item.Title` which could be empty: `item %q has no question block set`. If `Title` is empty, the error reads `item "" has no question block set` which is not very helpful.
- **Fix**: Consider including the item index or item key in the error message for better debugging.

### [NOTE] `BuildDeleteRequests` in update: delete requests and create requests in same batch
- **File**: `internal/resource_form/crud_update.go:100-116`
- **Issue**: Delete requests (reversed) and create requests (sequential) are appended to the same batch and sent in one `batchUpdate` call. The Google Forms API processes requests sequentially within a batch, so this should work correctly -- all deletes happen before all creates. This is efficient (one API call instead of two) but worth verifying against API documentation.
- **Fix**: Add a comment explaining the sequential processing guarantee.

### [NOTE] `forms.Grading.PointValue` is int64 in API, matching `GradingBlock.Points`
- **File**: `internal/convert/items_to_requests.go:82`
- **Issue**: The `forms.Grading` struct uses `PointValue int64` and `convert.GradingBlock` uses `Points int64`. These match correctly. No issue, just confirming the type alignment.

### [NOTE] Missing `GeneralFeedback` field from Forms API Grading
- **File**: `internal/convert/items_to_requests.go:78-96`
- **Issue**: The Google Forms API `Grading` struct has a `GeneralFeedback *Feedback` field (shown to all respondents regardless of answer correctness) that is not mapped in `applyGrading` or `convertGrading`. Only `WhenRight` (feedback_correct) and `WhenWrong` (feedback_incorrect) are supported.
- **Fix**: Document this as an unsupported field or add it to the schema. Low priority for V1.

### [NOTE] `Import` auto-generates item_keys that may not be recognized after plan
- **File**: `internal/resource_form/import.go:14-24`
- **Issue**: After import, items get auto-generated keys like `item_0`, `item_1`, etc. The user's HCL config must use these exact keys or Terraform will report a diff. The comment on line 17 mentions users should "review and rename" but there is no mechanism for renaming without recreating.
- **Fix**: Document the import workflow clearly in provider docs, including how to update item_keys using `moved` blocks or state manipulation.

### [NOTE] `Update` always sends `UpdateFormInfo` even if title/description unchanged
- **File**: `internal/resource_form/crud_update.go:48-52`
- **Issue**: The update always appends an `UpdateFormInfo` request even when title and description have not changed. This is a minor inefficiency but not a correctness issue since the API is idempotent for info updates.
- **Fix**: Compare plan vs state title/description and skip if unchanged. Low priority.

### [NOTE] `convertGradingToTF` uses empty string check rather than zero-value awareness
- **File**: `internal/resource_form/state_convert.go:222-238`
- **Issue**: `convertGradingToTF` checks `g.CorrectAnswer != ""` to decide between `StringValue` and `StringNull`. If the API ever returns an empty string for a field that was intentionally set to empty (unlikely but possible), it would be converted to null in state.
- **Fix**: Acceptable for V1 given Forms API behavior. Document the assumption.

---

## File-by-File Compilation Assessment

| File | Compiles? | Notes |
|------|-----------|-------|
| `model.go` | Yes | Clean, minimal imports |
| `schema.go` | Yes | All imports used, types correct |
| `resource.go` | Yes | Interface checks will catch mismatches at compile time |
| `validators.go` | Yes | Clean validator implementations |
| `validators_item.go` | Yes | Clean validator implementations |
| `validators_test.go` | Yes | Correct tftypes usage |
| `plan_modifiers.go` | Yes | Correct planmodifier.String implementation |
| `state_convert.go` | Yes | Correct use of attr.Type and types package |
| `crud_create.go` | Yes | Forms API types used correctly |
| `crud_read.go` | Yes | Client interfaces used correctly |
| `crud_update.go` | Yes | Forms API types used correctly |
| `crud_delete.go` | Yes | Drive API interface used correctly |
| `import.go` | Yes | Standard import pattern |
| `convert/types.go` | Yes | Plain Go types only |
| `convert/items_to_requests.go` | Yes | Forms API types used correctly |
| `convert/form_to_model.go` | Yes | Forms API types used correctly |
| `convert/json_mode.go` | Yes | Standard library only |
| `convert/items_to_requests_test.go` | Yes | Correct test patterns |
| `convert/form_to_model_test.go` | Yes | Comprehensive test coverage |
| `convert/json_mode_test.go` | Yes | Good edge case coverage |

---

## Test Coverage Summary

| Area | Coverage | Notes |
|------|----------|-------|
| `convert/items_to_requests` | Good | All 3 question types + grading, delete requests, info/quiz requests |
| `convert/form_to_model` | Good | All question types, grading round-trip, unsupported item skip, empty form, multiple items |
| `convert/json_mode` | Good | Parse, normalize, hash, declarative-to-requests |
| `validators 1-4` | Good | Multiple test cases per validator |
| `validators 5-7` | **Missing** | No tests for OptionsRequired, CorrectAnswerInOptions, GradingRequiresQuiz |
| `state_convert` | Indirect | Tested via CRUD integration; no unit tests |
| `plan_modifiers` | **Missing** | No unit tests for ContentJSONHashModifier |
| `CRUD operations` | **Missing** | No unit tests (requires mock client) |

---

## Priority Summary

**Should fix before release:**
1. Add tests for validators 5-7 (WARNING)
2. Deduplicate `normalizeJSONHash` / use `convert.HashJSON` in plan modifier (WARNING)

**Should fix soon:**
3. Add unit tests for `ContentJSONHashModifier`
4. Replace custom `strContains` with `strings.Contains`
5. Document positional key mapping assumption

**Nice to have:**
6. Add `item_key` format validation
7. Document `content_json` drift detection limitations
8. Add `GeneralFeedback` support

---

*Reviewed by: Senior Go Engineer (Sprint 2 Review)*
*Date: 2026-02-09*
*Reviewer model: Claude Opus 4.6*
