# Fagan Inspection Report: Test Coverage & Quality

**Role:** Tester
**Inspector:** Claude Opus 4.6 (Agent)
**Date:** 2026-02-09
**Scope:** All test files and untested production code in terraform-provider-googleforms

---

## Summary

The test suite demonstrates solid coverage of the happy-path flows for all major subsystems. Tests are well-structured, use `t.Parallel()` appropriately, and the CRUD tests employ mock APIs cleanly. However, the inspection identified **12 Major**, **14 Minor**, and **4 Style** findings related to coverage gaps, weak assertions, missing edge cases, and entirely untested production files.

---

## 1. `internal/client/errors_test.go`

**Production file:** `errors.go`

### Correctness: GOOD
All tests verify the correct string output and error chain behavior. Assertions are direct string comparisons.

### Coverage: GOOD
All public functions tested: `APIError.Error()`, `APIError.Unwrap()`, `NotFoundError`, `RateLimitError`, `IsNotFound`, `IsRateLimit`, `ErrorStatusCode`.

### Findings

| ID | Severity | Finding |
|----|----------|---------|
| T-ERR-01 | Minor | `APIError.Error()` only tested with status 500. No test for status 0 or negative status codes to verify formatting edge cases. |
| T-ERR-02 | Minor | `APIError` with `Err: nil` -- `Unwrap()` returns nil, which is valid, but not explicitly tested. A test confirming `errors.Unwrap(err) == nil` when `Err` is nil would document intent. |
| T-ERR-03 | Minor | `ErrorStatusCode` tests a wrapped `NotFoundError` and `RateLimitError` through `errors.As`, but does not test double-wrapped errors (e.g., `fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", &NotFoundError{...}))`). This is important because CRUD code wraps errors. |
| T-ERR-04 | Minor | No test for `ErrorStatusCode` with a `nil` error. Would a nil error panic or return 0? The production code would get `errors.As(nil, ...)` which returns false, so it returns 0. Should be documented with a test. |

---

## 2. `internal/client/retry_test.go`

**Production file:** `retry.go`

### Correctness: GOOD
Tests properly verify attempt counts with `atomic.Int32`. Context cancellation test uses goroutine-based approach correctly.

### Coverage: GOOD
All retry/non-retry status codes covered: 400, 401, 403, 404, 429, 500, 502, 503, 504.

### Findings

| ID | Severity | Finding |
|----|----------|---------|
| T-RET-01 | **Major** | `TestRetry_BackoffIncreases` (line 282): The test only checks `gap3 > gap1` but with 25% jitter, this assertion can fail non-deterministically. With `InitialBackoff=50ms`, gap1 is ~[37.5ms, 62.5ms] and gap3 is ~[150ms, 250ms]. The assertion will *usually* pass but is a flaky test. The comparison should use a wider margin or check `gap3 > gap1*1.5` to account for jitter. |
| T-RET-02 | Minor | `TestRetry_RespectsContextCancellation`: The test does not verify the error message wrapping. The `wrapContextError` function includes the last API error in the message, but the test only checks `err != nil` and `ctx.Err() != nil`. Should verify the wrapped error message contains the last API error. |
| T-RET-03 | Minor | No test for `MaxRetries=0` (initial attempt only). This boundary condition should be documented with a test. |
| T-RET-04 | Minor | No test for `backoffDuration` capping at `MaxBackoff`. The `TestRetry_BackoffIncreases` uses `MaxBackoff=5s` with only 4 attempts, so the cap is never reached. A dedicated test with a small MaxBackoff and many attempts would verify the cap. |
| T-RET-05 | Minor | Private functions `isRetryable`, `backoffDuration`, `addJitter`, `sleepWithContext`, `wrapContextError` are only tested indirectly via `WithRetry`. Direct unit tests for `isRetryable` and `wrapContextError` would strengthen coverage. |

---

## 3. `internal/convert/items_to_requests_test.go`

**Production file:** `items_to_requests.go`

### Correctness: GOOD
Tests verify item title, question type, location index, and grading fields.

### Findings

| ID | Severity | Finding |
|----|----------|---------|
| T-ITR-01 | **Major** | `ItemModelToCreateRequest` error path for "no question block set" is **never tested**. When `item.MultipleChoice`, `item.ShortAnswer`, and `item.Paragraph` are all nil, the function returns an error. No test exercises this. A buggy implementation that silently succeeds would not be caught. |
| T-ITR-02 | Minor | `TestShortAnswerToRequest_Basic`: Does not test the `Required` field (it defaults to false, but the test doesn't assert `q.Required == false`). |
| T-ITR-03 | Minor | `TestParagraphToRequest_Basic`: Does not test the `Required` field. |
| T-ITR-04 | Minor | No test for `ItemsToCreateRequests` when one item in the middle of the list has an error. The function wraps errors with `item[%d]` prefix but this wrapping is never verified. |
| T-ITR-05 | Minor | `BuildUpdateInfoRequest` always sets `UpdateMask: "title,description"`. No test verifies behavior when both title and description are empty strings (valid but edge case). |
| T-ITR-06 | Style | `TestFormsImportUsed` (line 404) is a no-op test to suppress unused import warnings. This should be removed; the forms import IS used by other tests. |

---

## 4. `internal/convert/form_to_model_test.go`

**Production file:** `form_to_model.go`

### Correctness: GOOD
Thorough testing of all conversion paths. Tests for grading, unsupported items, and order preservation.

### Findings

| ID | Severity | Finding |
|----|----------|---------|
| T-FTM-01 | **Major** | `resolveItemKey` function (line 49 of production) is **never directly tested**. It falls back to `item_N` when no key map entry exists. This is critical for import functionality. While it's indirectly tested by `TestFormToModel_WithShortAnswerItem` (no key map provided), there's no test that explicitly verifies the `item_N` fallback naming pattern. |
| T-FTM-02 | Minor | `FormItemToItemModel` with `QuestionItem != nil` but `Question == nil` -- This returns `(nil, nil)` per line 61 but only the `QuestionItem == nil` case is tested in `TestFormItemToItemModel_NilQuestionItem`. |
| T-FTM-03 | Minor | `convertChoiceQuestion`: No test for DROPDOWN or CHECKBOX type (non-RADIO). The production code only handles `RADIO` (line 73); a DROPDOWN question silently falls through to `return nil, nil`. Tests don't verify this behavior. |
| T-FTM-04 | Minor | `convertGrading`: No test for `Grading.CorrectAnswers` with an empty `Answers` slice (length 0). The production code checks `len(g.CorrectAnswers.Answers) > 0` (line 128) but this boundary is untested. |
| T-FTM-05 | Minor | `FormToModel` with `form.Info == nil`: No test for a form where Info is nil. The production code guards with `if form.Info != nil` (line 22). |

---

## 5. `internal/convert/json_mode_test.go`

**Production file:** `json_mode.go`

### Correctness: GOOD
Tests cover valid JSON, empty arrays, invalid JSON, and hash determinism.

### Findings

| ID | Severity | Finding |
|----|----------|---------|
| T-JSON-01 | Minor | `NormalizeJSON` and `HashJSON` error paths are only tested via invalid JSON in `ParseDeclarativeJSON`. `NormalizeJSON` with invalid JSON is never directly tested. `HashJSON` with invalid JSON is never tested either. |
| T-JSON-02 | Minor | `DeclarativeJSONToRequests` with invalid JSON is not tested. It delegates to `ParseDeclarativeJSON`, and the error should propagate, but this path is not verified. |
| T-JSON-03 | Style | `ParseDeclarativeJSON` tests don't verify the actual item structure -- only the title and count. The parsed items' nested `QuestionItem` fields are not inspected. |

---

## 6. `internal/provider/provider_test.go`

**Production file:** `provider.go`

### Correctness: MOSTLY GOOD, with one issue

### Findings

| ID | Severity | Finding |
|----|----------|---------|
| T-PROV-01 | **Major** | `TestProviderConfigure_WithCredentials` (line 77): The code on line 96-99 uses an unusual Go pattern: `configState, err := tfsdk.Config{...}, error(nil)`. This is a multi-value assignment where `err` is always `nil`, so the `if err != nil` check on line 100 is dead code. The test can never fail at that point. While not a false positive per se, it's misleading code that looks like it handles an error that can never occur. |
| T-PROV-02 | **Major** | `TestProviderConfigure_FallbackToEnvVar` and `TestProviderConfigure_FallbackToADC` are marked "Not parallel" and use `t.Setenv`, which is correct. However, they test the *provider's* env var fallback by accepting the stub "Client Creation Failed" error. This means they don't actually verify that credentials resolution happened correctly -- they only verify no "missing credentials" error occurred. A buggy `resolveCredentials` that always returns empty string would still pass these tests because the stub `NewClient` always fails. |
| T-PROV-03 | Minor | `resolveCredentialValue` (production line 135) is not directly tested. It handles inline JSON (starts with `{`) vs file paths. Only indirect testing via the `FallbackToEnvVar` test (which writes to a file). The inline JSON path is tested via `TestProviderConfigure_WithCredentials`. |
| T-PROV-04 | Minor | No test for `TestProviderConfigure_InvalidCredentials` verifying the specific error summary. The test only checks `HasError()` but does not validate the error message content. |
| T-PROV-05 | Style | The acceptance test `TestProviderAcceptance_ConfigIsValid` uses `PlanOnly: true` and `ExpectNonEmptyPlan: false`. This is minimal validation -- it only checks the schema is parseable, not that any configuration is actually processed. |

---

## 7. `internal/resource_form/validators_test.go`

**Production files:** `validators.go`

### Correctness: GOOD
Tests thoroughly exercise all 4 validators defined in `validators.go`, plus the `ConfigValidators` wiring.

### Findings

| ID | Severity | Finding |
|----|----------|---------|
| T-VAL-01 | **Major** | `MutuallyExclusiveValidator`: No test for when `content_json` is set to an empty string `""`. The production code (line 56) checks `!contentJSON.IsNull() && !contentJSON.IsUnknown()` but does NOT check for empty string. An empty string content_json with items present would NOT trigger the validator error. Is this intentional? If not, it's a validator bug. Either way, the edge case needs a test. |
| T-VAL-02 | Minor | `UniqueItemKeyValidator`: No test for items where `item_key` is unknown (e.g., computed from a variable). The production code (line 148) skips unknown keys but this is untested. |
| T-VAL-03 | Minor | `ExactlyOneSubBlockValidator`: No test for items in content_json mode (where items list may be empty/null). The validator correctly early-returns for null/unknown, but this path isn't explicitly documented with a test. |
| T-VAL-04 | Minor | `AcceptingResponsesRequiresPublishedValidator`: No test for both values being null/unknown (which is the default case). The production code returns early if `accepting` is null/unknown, so this would pass, but it's an untested path. |

---

## 8. `internal/resource_form/validators_item_test.go`

**Production file:** `validators_item.go`

### Correctness: GOOD
Tests cover positive and negative cases for all 3 item-level validators.

### Findings

| ID | Severity | Finding |
|----|----------|---------|
| T-IVAL-01 | **Major** | `GradingRequiresQuizValidator`: Only tests grading on `short_answer` items. Does NOT test grading on `multiple_choice` or `paragraph` items. The production code's `itemHasGrading` function (line 212) checks all three types, but only one branch is tested. A bug in the `mc.Grading` or `p.Grading` check would go undetected. |
| T-IVAL-02 | Minor | `OptionsRequiredForChoiceValidator`: No test for when the `options` field is null (not just empty). Production code checks `opts.IsNull()` (line 54). |
| T-IVAL-03 | Minor | `CorrectAnswerInOptionsValidator`: No test for non-multiple-choice items with grading. The production code only checks `item.MultipleChoice` (line 100), which is correct, but no test documents that short_answer/paragraph grading is ignored. |
| T-IVAL-04 | Minor | `TestConfigValidators_ReturnsAllSeven` verifies the count is 7 but does not verify the types of the validators in the slice. If two validators were duplicated and one removed, the count would be wrong but the type check would catch the actual semantic issue better. |

---

## 9. `internal/resource_form/plan_modifiers_test.go`

**Production file:** `plan_modifiers.go`

### Correctness: EXCELLENT
Comprehensive testing of all branches: null state, unknown state, null config, semantically equal JSON, different JSON, invalid JSON (3 sub-cases).

### Findings

| ID | Severity | Finding |
|----|----------|---------|
| T-PM-01 | Minor | No test for when config is Unknown (only null config is tested). The production code checks `req.ConfigValue.IsUnknown()` (line 42). |
| T-PM-02 | Style | `var _ attr.Value = types.StringNull()` (line 183) is a hack to suppress unused import. Should be removed if the `attr` import is genuinely unnecessary. |

---

## 10. `internal/resource_form/crud_test.go`

**Production files:** `crud_create.go`, `crud_read.go`, `crud_update.go`, `crud_delete.go`

### Correctness: GOOD
Tests exercise all 4 CRUD operations with success and error paths. Partial state save on Create is tested.

### Findings

| ID | Severity | Finding |
|----|----------|---------|
| T-CRUD-01 | **Major** | **Create: content_json parse error not tested.** When `content_json` contains invalid JSON, `DeclarativeJSONToRequests` returns an error. There is no test verifying the "Error Parsing content_json" diagnostic is produced. |
| T-CRUD-02 | **Major** | **Update: BatchUpdate error path not tested.** `TestUpdate_TitleChange_Success` and `TestUpdate_ItemsReplaced_Success` test happy paths, and `TestUpdate_APIError_ReturnsDiagnostic` tests the pre-read error path. But there is no test for when `BatchUpdate` itself fails during Update. This is a critical error path (the form exists but the update fails). |
| T-CRUD-03 | **Major** | **Update: content_json mode not tested.** No test exercises the Update path with `content_json` (lines 68-84 of `crud_update.go`). All update tests use HCL item blocks. The content_json update path includes delete-all + re-create, which is a distinct code path. |
| T-CRUD-04 | **Major** | **Create: SetPublishSettings error path not tested.** When `SetPublishSettings` fails during Create (line 143 of `crud_create.go`), the error "Error Setting Publish Settings" is produced. No test covers this. |
| T-CRUD-05 | Minor | **Read: content_json mode not tested.** The Read function has a branch (lines 73-85 of `crud_read.go`) where content_json is preserved from state. No test exercises this path. |
| T-CRUD-06 | Minor | **Create: error from `convertFormModelToTFState` not tested.** If `convert.FormToModel` returns an error during the final read-back (line 193 of `crud_create.go`), no test covers the "Error Converting Form Response" diagnostic. |
| T-CRUD-07 | Minor | **Delete: state not verified as removed.** `TestDelete_Success` verifies the Drive delete was called but does not verify the response state is null/removed. The framework handles this automatically, but explicit verification would document the behavior. |
| T-CRUD-08 | Minor | **Update: quiz settings change path not tested.** The `planQuiz != stateQuiz` branch (line 57 of `crud_update.go`) is not exercised by any test. Only Create tests exercise quiz settings. |
| T-CRUD-09 | Minor | Mock correctness: `MockFormsAPI.BatchUpdateFunc` default (line 46 of `mock_forms.go`) returns an empty `BatchUpdateFormResponse`. This is accurate for success, but the mock never validates that `IncludeFormInResponse: true` is set. A more accurate mock would verify this field. |

---

## 11. UNTESTED PRODUCTION FILES

### 11a. `internal/resource_form/state_convert.go` -- NO UNIT TESTS

| ID | Severity | Finding |
|----|----------|---------|
| T-SC-01 | **Major** | **`tfItemsToConvertItems`** -- This function converts TF types.List to []convert.ItemModel. It is only tested indirectly through CRUD tests. It has branches for MultipleChoice, ShortAnswer, and Paragraph that should have direct unit tests. Bugs in the conversion (e.g., forgetting to set `result[i].Title`) would only be caught if the CRUD test happens to check that specific field. |
| T-SC-02 | **Major** | **`convertFormModelToTFState`** -- Converts convert.FormModel to TF state. Sets `EditURI` by constructing a URL from the form ID. Has no direct test. If the `model.ID == ""` guard (line 113) were removed, the URI would be malformed for edge cases. |
| T-SC-03 | Minor | **`buildItemKeyMap`** -- Returns nil for null/unknown/empty items list. Not directly tested. While indirectly tested through Read, direct testing would verify the empty-string key filtering (line 139: `googleID != "" && itemKey != ""`). |
| T-SC-04 | Minor | **`convertItemsToTFList`** -- Returns `types.ListNull` for empty items. Not directly tested. |
| T-SC-05 | Minor | **`convertGradingToTF`** -- Converts GradingBlock with null-handling for empty strings. Not directly tested. If the empty-string-to-null logic (lines 222-237) were inverted, it would only be caught by end-to-end CRUD tests checking grading values. |
| T-SC-06 | Minor | **`itemObjectType` / `gradingObjectType`** -- These define the TF type structure. If they drift from `schema.go`, items would fail to serialize. No test verifies schema/type alignment. |

### 11b. `internal/resource_form/import.go` -- NO UNIT TESTS

| ID | Severity | Finding |
|----|----------|---------|
| T-IMP-01 | Minor | `ImportState` simply delegates to `resource.ImportStatePassthroughID`. While trivial, a test confirming the ID is correctly passed through would document the import behavior and catch accidental changes (e.g., someone changing the path from `"id"` to something else). |

### 11c. `internal/testutil/sweeper.go` -- STUB ONLY

| ID | Severity | Finding |
|----|----------|---------|
| T-SW-01 | Minor | The sweeper is entirely a TODO stub with no implementation. It has no test and no code. This is not a test gap per se, but should be tracked as incomplete test infrastructure. The comment documents the intended 5-step implementation. |

### 11d. `internal/client/client.go` -- NO UNIT TESTS

| ID | Severity | Finding |
|----|----------|---------|
| T-CLI-01 | Minor | `NewClient`, `buildTokenSource`, `resolveCredentials` (the client-package version), `tokenSourceFromJSON`, `tokenSourceFromADC` -- All untested. These require real Google API credentials so unit tests are not straightforward, but `resolveCredentials` (which reads env vars) could be tested in isolation. |

### 11e. `internal/client/forms_api.go` and `drive_api.go` -- NO UNIT TESTS

| ID | Severity | Finding |
|----|----------|---------|
| T-API-01 | Minor | `wrapGoogleAPIError`, `wrapDriveAPIError`, and `mapStatusToError` have no unit tests. These are the real-API-to-custom-error translation functions. `mapStatusToError` is critical -- it determines whether 404 becomes `NotFoundError` and 429 becomes `RateLimitError`. Direct unit tests would be simple and high-value. |
| T-API-02 | Minor | `DriveAPIClient.Delete` swallows 404 errors (lines 43-46). This behavior is documented but not tested. A unit test with a mock `drive.Service` would verify this. |

---

## 12. CROSS-CUTTING FINDINGS

| ID | Severity | Finding |
|----|----------|---------|
| T-XC-01 | Style | Tests in `crud_test.go` are in package `resourceform` (same package), giving access to unexported fields. This is appropriate for unit tests but means the tests could pass even if exported interfaces change. |
| T-XC-02 | Style | No table-driven tests used in `errors_test.go` or `retry_test.go` where many similar test cases exist (e.g., "no retry on 400/401/403/404" could be one table-driven test). |

---

## Defect Summary

| Severity | Count | Description |
|----------|-------|-------------|
| **Major** | 12 | False-positive-risk tests, untested critical paths |
| **Minor** | 14 | Missing edge cases, weak assertions, untested helpers |
| **Style** | 4 | Test organization, dead code |
| **Total** | 30 | |

### Top Priority Majors for Rework

1. **T-CRUD-01**: Create with invalid content_json untested
2. **T-CRUD-02**: Update BatchUpdate error path untested
3. **T-CRUD-03**: Update content_json mode entirely untested
4. **T-CRUD-04**: Create SetPublishSettings error path untested
5. **T-ITR-01**: ItemModelToCreateRequest error path for no-question-block untested
6. **T-IVAL-01**: GradingRequiresQuiz only tests short_answer, not MC/paragraph
7. **T-SC-01**: state_convert.go `tfItemsToConvertItems` has zero direct tests
8. **T-SC-02**: state_convert.go `convertFormModelToTFState` has zero direct tests
9. **T-VAL-01**: MutuallyExclusiveValidator empty string edge case
10. **T-RET-01**: Flaky backoff test due to jitter
11. **T-FTM-01**: resolveItemKey fallback pattern untested
12. **T-PROV-01/02**: Provider configure tests have dead code and weak assertions

---

*End of Tester Inspection Report*
