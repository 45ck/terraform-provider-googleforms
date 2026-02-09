# Fagan Iteration 2 -- Tester Report

**Inspector:** Tester
**Date:** 2026-02-09
**Scope:** 6 modified/new test files from rework iteration

---

## 1. File: `internal/client/retry_test.go` (348 lines, 14 tests)

### 1.1 Correctness Verification

| Test | Verifies Claimed Behavior | False-Positive Risk | Verdict |
|------|--------------------------|---------------------|---------|
| `TestRetry_SuccessOnFirstAttempt` | 1 attempt, no error | Low -- checks both err==nil AND attempt==1 | PASS |
| `TestRetry_SuccessAfterRetries` | Transient 503 then success | Low -- checks err==nil AND attempts==3 | PASS |
| `TestRetry_ExhaustsRetries` | MaxRetries=3 => 4 total attempts | Low -- checks err!=nil AND attempts==4 | PASS |
| `TestRetry_NoRetryOn400` | Client errors not retried | Low -- checks attempts==1 | PASS |
| `TestRetry_NoRetryOn404` | NotFoundError not retried + preserved | Low -- also checks `IsNotFound(err)` | PASS |
| `TestRetry_NoRetryOn401` | 401 not retried | Low | PASS |
| `TestRetry_NoRetryOn403` | 403 not retried | Low | PASS |
| `TestRetry_RetriesOn429` | RateLimitError retried | Low -- checks success after retry | PASS |
| `TestRetry_RetriesOn503` | 503 retried | Low | PASS |
| `TestRetry_RetriesOn502` | 502 retried | Low | PASS |
| `TestRetry_RetriesOn504` | 504 retried | Low | PASS |
| `TestRetry_RespectsContextCancellation` | Context cancel stops retry | Low -- checks both err!=nil and ctx.Err() | PASS |
| `TestRetry_BackoffIncreases` | Exponential backoff growth | MODERATE -- see finding T2-001 | PASS with caveat |
| `TestRetry_NonAPIErrorNotRetried` | Plain `fmt.Errorf` not retried | Low | PASS |

### 1.2 Mock/Production Alignment

- Tests use `WithRetry`, `DefaultRetryConfig`, `APIError`, `NotFoundError`, `RateLimitError`, and `IsNotFound` -- all exist in `client/retry.go` and `client/errors.go`.
- `testRetryConfig()` helper uses very short durations (1ms/50ms) to keep tests fast. Correct pattern.

### 1.3 t.Parallel() / t.Helper()

- All 14 tests call `t.Parallel()`. Correct.
- `testRetryConfig()` is a plain function, not a test helper -- no `t.Helper()` needed. Correct.

### 1.4 Findings

**T2-001 (Observation, Low): `TestRetry_BackoffIncreases` backoff assertion is weak**
- The test checks `gap3 < gap1/2` and `gap2 < gap1/2` as failure conditions.
- With 25% jitter and 50ms initial backoff, gap1 ~ [37.5ms, 62.5ms], gap2 ~ [75ms, 125ms], gap3 ~ [150ms, 250ms]. The assertion `gap3 >= gap1/2` would pass even if backoff was constant (e.g., gap3=37ms > gap1/2=31ms).
- However, the test does verify general increase tendency, and the specific margin was chosen to tolerate jitter on CI. The property "exponential" is partially tested.
- **Risk:** A buggy implementation that uses constant backoff could still pass. Not a test correctness bug but a weaker-than-ideal assertion.
- **Recommendation:** Accept as-is given jitter makes tight bounds flaky. Optionally, add a test that checks `backoffDuration` directly with fixed seed.

---

## 2. File: `internal/convert/items_to_requests_test.go` (427 lines, 16 tests)

### 2.1 Correctness Verification

| Test | Verifies | False-Positive Risk | Verdict |
|------|----------|---------------------|---------|
| `TestMultipleChoiceToRequest_Basic` | Title, location, RADIO type, options | Low -- checks 7+ fields | PASS |
| `TestMultipleChoiceToRequest_WithGrading` | Grading fields | Low -- checks points, answer, feedback | PASS |
| `TestMultipleChoiceToRequest_Required` | Required flag | Low | PASS |
| `TestShortAnswerToRequest_Basic` | TextQuestion.Paragraph=false | Low | PASS |
| `TestShortAnswerToRequest_WithGrading` | Points, feedback | Low | PASS |
| `TestParagraphToRequest_Basic` | TextQuestion.Paragraph=true | Low | PASS |
| `TestParagraphToRequest_WithGrading` | Points, feedback | Low | PASS |
| `TestItemsToCreateRequests_MultipleItems_CorrectOrder` | Location indices 0,1,2 | Low | PASS |
| `TestItemsToCreateRequests_EmptyList` | Empty input = empty output | Low | PASS |
| `TestItemsToCreateRequests_SingleItem` | Single item | Low | PASS |
| `TestBuildDeleteRequests_MultipleItems_ReverseOrder` | Indices 3,2,1,0 (F-004) | Low -- directly asserts order | PASS |
| `TestBuildDeleteRequests_EmptyList` | 0 items = 0 requests | Low | PASS |
| `TestBuildUpdateInfoRequest_TitleAndDescription` | Title, desc, update_mask | Low | PASS |
| `TestBuildUpdateInfoRequest_DescriptionOnly` | Empty title handled | Low | PASS |
| `TestBuildQuizSettingsRequest_EnableQuiz` | IsQuiz=true, mask | Low | PASS |
| `TestBuildQuizSettingsRequest_DisableQuiz` | IsQuiz=false | Low | PASS |
| `TestItemModelToCreateRequest_NoQuestionBlock` | Error for empty item | Low -- checks error string | PASS |
| `TestFormsImportUsed` | Compile-time check | N/A | PASS |

### 2.2 Mock/Production Alignment

- Tests use `convert.ItemModel`, `convert.MultipleChoiceBlock`, etc. -- all match `convert/types.go`.
- `ItemModelToCreateRequest`, `ItemsToCreateRequests`, `BuildDeleteRequests`, `BuildUpdateInfoRequest`, `BuildQuizSettingsRequest` are all public functions in the convert package.
- `forms.Form{}` import confirms Google API types are used correctly.

### 2.3 F-004 Cross-Reference (Batch Ordering)

- `TestBuildDeleteRequests_MultipleItems_ReverseOrder` directly verifies that delete requests are emitted in reverse index order (3, 2, 1, 0). This correctly tests the F-004 fix.
- `TestItemsToCreateRequests_MultipleItems_CorrectOrder` verifies create requests have ascending indices (0, 1, 2).

### 2.4 t.Parallel() / t.Helper()

- **Finding T2-002 (Minor): Tests in items_to_requests_test.go do NOT call `t.Parallel()`.**
  - All 16+ tests in this file omit `t.Parallel()`. Since they test pure functions with no shared state, adding `t.Parallel()` would be safe and consistent with other test files.
  - Not a correctness issue, only a consistency issue.

### 2.5 Findings

**T2-002 (Minor): Missing `t.Parallel()` in items_to_requests_test.go**
- None of the 16+ tests call `t.Parallel()`. Other test files consistently use it.
- No correctness impact but inconsistent with project conventions.

---

## 3. File: `internal/resource_form/crud_test.go` (1467 lines, 23 tests)

### 3.1 Correctness Verification

| Test | Verifies | False-Positive Risk | Verdict |
|------|----------|---------------------|---------|
| `TestCreate_BasicForm_Success` | Create + Get flow, form ID in state | Low | PASS |
| `TestCreate_WithItems_Success` | BatchUpdate called with 2 CreateItem | Low | PASS |
| `TestCreate_WithContentJSON_Success` | content_json path, BatchUpdate 1 item | Low | PASS |
| `TestCreate_APIError_ReturnsDiagnostic` | Error diagnostic on 500 | Low -- checks summary string | PASS |
| `TestCreate_BatchUpdateError_PartialStateSaved` | Partial state save pattern | Low -- checks both error AND form_id | PASS |
| `TestRead_ExistingForm_PopulatesState` | State populated correctly | Low | PASS |
| `TestRead_FormNotFound_RemovesFromState` | 404 -> state null | Low | PASS |
| `TestRead_APIError_ReturnsDiagnostic` | Error diagnostic on 500 | Low | PASS |
| `TestUpdate_TitleChange_Success` | UpdateFormInfo in batch, 2 Get calls | Low | PASS |
| `TestUpdate_ItemsReplaced_Success` | Delete+Create in batch requests | Low | PASS |
| `TestDelete_Success` | Drive.Delete called with correct ID | Low | PASS |
| `TestDelete_NotFound_NoError` | 404 on delete = no error | Low | PASS |
| `TestDelete_APIError_ReturnsDiagnostic` | Error diagnostic on 500 | Low | PASS |
| `TestCreate_WithQuizGrading_Success` | Quiz settings in batch | Low | PASS |
| `TestCreate_WithPublishSettings_Success` | SetPublishSettings called | Low | PASS |
| `TestRead_WithItems_CorrectMapping` | Items correctly mapped to state | Low -- checks 2 items, model | PASS |
| `TestUpdate_PublishSettingsChanged_Success` | SetPublishSettings on update | Low | PASS |
| `TestUpdate_APIError_ReturnsDiagnostic` | Pre-read error diagnostic | Low | PASS |
| `TestCreate_ContentJSON_ParseError` | Invalid JSON error diagnostic | Low | PASS |
| `TestUpdate_BatchUpdateError_ReturnsDiagnostic` | BatchUpdate error diagnostic | Low | PASS |
| `TestUpdate_WithContentJSON_Success` | content_json update path | Low | PASS |
| `TestCreate_PublishSettingsError_ReturnsDiagnostic` | Publish settings error | Low | PASS |

### 3.2 F-002 Cross-Reference (State reads Title/Description from API)

**Critical check: Do mock Get responses return correct Title/Description values?**

- `TestCreate_BasicForm_Success`: Mock `GetFunc` returns `basicFormResponse(formID, "My Form")` which has `Info: &forms.Info{Title: "My Form", DocumentTitle: "My Form"}`. Since `convertFormModelToTFState` now reads Title from `model.Title` (via `FormToModel` which reads `form.Info.Title`), the state will get Title="My Form". This matches the plan Title="My Form". **Consistent.**

- `TestCreate_WithItems_Success`: Mock Get returns `formWithItems(formID, "Items Form")` with Title="Items Form". Plan has Title="Items Form". **Consistent.**

- `TestRead_ExistingForm_PopulatesState`: Mock Get returns `basicFormResponse(formID, "Existing Form")`. State has Title="Existing Form". **Consistent.**

- `TestUpdate_TitleChange_Success`: Mock Get always returns Title="New Title". Plan has Title="New Title". **Consistent.**

- All mock Get responses provide matching titles. **F-002 rework is correctly reflected in tests.**

### 3.3 F-004 Cross-Reference (Batch Request Ordering)

- `TestUpdate_ItemsReplaced_Success` verifies deleteCount==1 and createCount==1, but does NOT assert the order of requests (deletes before creates).
- However, the ordering is enforced by the production code in `crud_update.go` (lines 98-129): deletes are appended first, then quiz settings, then creates. The unit test for `BuildDeleteRequests` already verifies reverse-order indices.
- **Finding T2-003 (Observation): No test directly verifies the full batch request order in the update path (UpdateFormInfo first, then deletes, then quiz settings, then creates).** The individual pieces are tested but not the composite ordering.

### 3.4 Mock/Production Alignment

- `testResource()` creates `FormResource` with `client.Client{Forms: formsAPI, Drive: driveAPI}`. This matches `resource.go` line 23.
- `MockFormsAPI` implements `client.FormsAPI` (compile-time check at `mock_forms.go:22`).
- `MockDriveAPI` implements `client.DriveAPI` (compile-time check at `mock_drive.go:17`).
- `FormResourceModel` used in `stateFormID()` matches `model.go`.
- Error diagnostic summary strings match production code exactly:
  - `"Error Creating Google Form"` matches `crud_create.go:46`
  - `"Error Reading Google Form"` matches `crud_read.go:47`
  - `"Error Deleting Google Form"` matches `crud_delete.go:46`
  - `"Error Updating Google Form"` matches `crud_update.go:144`
  - `"Error Reading Google Form Before Update"` matches `crud_update.go:40`
  - `"Error Parsing content_json"` matches `crud_create.go:85`
  - `"Error Setting Publish Settings"` matches `crud_create.go:148`

### 3.5 t.Parallel() / t.Helper()

- All 23 tests call `t.Parallel()`. Correct.
- Helper functions `buildPlan`, `buildState`, `emptyState`, `stateFormID` all call `t.Helper()`. Correct.
- `testResource`, `basicFormResponse`, `formWithItems` do not call `t.Helper()` -- they are not assertion helpers so this is fine.

### 3.6 Findings

**T2-003 (Observation, Low): No composite batch-ordering test for Update**
- `TestUpdate_ItemsReplaced_Success` checks counts of DeleteItem and CreateItem requests but does not verify their relative order in the `batchReqs` slice.
- The ordering is: (1) UpdateFormInfo, (2) DeleteItem(s), (3) UpdateSettings (quiz), (4) CreateItem(s). A bug that reordered these in `crud_update.go` would NOT be caught by existing tests.
- **Risk:** Low, since the production code assembles them sequentially. But a refactor could break ordering without failing tests.
- **Recommendation:** Add a test or assertion that verifies relative request type positions in the batch.

---

## 4. File: `internal/resource_form/plan_modifiers_test.go` (179 lines, 8 tests)

### 4.1 Correctness Verification

| Test | Verifies | False-Positive Risk | Verdict |
|------|----------|---------------------|---------|
| `TestContentJSONHashModifier_Description` | Non-empty descriptions | Low | PASS |
| `TestContentJSONHashModifier_SemanticallyEqual_SuppressesDiff` | Reordered JSON = same hash | Low -- checks plan==state | PASS |
| `TestContentJSONHashModifier_DifferentContent_ShowsDiff` | Different JSON keeps diff | Low | PASS |
| `TestContentJSONHashModifier_NullState_Skips` | No state = no suppression | Low | PASS |
| `TestContentJSONHashModifier_UnknownState_Skips` | Unknown state = no suppression | Low | PASS |
| `TestContentJSONHashModifier_NullConfig_Skips` | Null config = plan stays null | Low | PASS |
| `TestContentJSONHashModifier_InvalidJSON_FallsThrough` | Invalid JSON = no suppression | Low -- 3 sub-cases | PASS |
| `TestContentJSONHashModifier_ImplementsInterface` | Compile-time check | N/A | PASS |

### 4.2 Mock/Production Alignment

- Tests directly instantiate `ContentJSONHashModifier{}` and call `PlanModifyString`. Matches `plan_modifiers.go`.
- Uses `convert.HashJSON` indirectly through the modifier. Correct.
- Tests `planmodifier.StringRequest/StringResponse` from the framework. Correct API.

### 4.3 t.Parallel() / t.Helper()

- All 8 tests (including subtests in `TestContentJSONHashModifier_InvalidJSON_FallsThrough`) call `t.Parallel()`. Correct.

### 4.4 Findings

No defects found. Test coverage is thorough for all edge cases (null, unknown, invalid JSON, equal, different).

---

## 5. File: `internal/resource_form/validators_item_test.go` (176 lines, 10 tests)

### 5.1 Correctness Verification

| Test | Verifies | False-Positive Risk | Verdict |
|------|----------|---------------------|---------|
| `TestOptionsRequiredForChoice_WithOptions_Passes` | Options present = no error | Low | PASS |
| `TestOptionsRequiredForChoice_EmptyOptions_Error` | Empty options = error | Low | PASS |
| `TestCorrectAnswerInOptions_ValidAnswer_Passes` | Answer "B" in ["A","B","C"] | Low | PASS |
| `TestCorrectAnswerInOptions_InvalidAnswer_Error` | Answer "D" not in options | Low | PASS |
| `TestCorrectAnswerInOptions_NoCorrectAnswer_Passes` | Null answer = ok | Low | PASS |
| `TestGradingRequiresQuiz_QuizTrueWithGrading_Passes` | Quiz + grading = ok | Low | PASS |
| `TestGradingRequiresQuiz_QuizFalseWithGrading_Error` | No quiz + grading = error | Low | PASS |
| `TestGradingRequiresQuiz_NoGrading_Passes` | No grading = ok | Low | PASS |
| `TestGradingRequiresQuiz_MultipleChoice_Error` | MC grading without quiz | Low | PASS |
| `TestGradingRequiresQuiz_Paragraph_Error` | Paragraph grading without quiz | Low | PASS |
| `TestConfigValidators_ReturnsAllSeven` | 7 validators registered | Low | PASS |

### 5.2 Error String Cross-Reference

| Test assertion | Production code string | Match? |
|---------------|----------------------|--------|
| `"requires at least one option"` | `"A multiple_choice question requires at least one option."` | YES (substring) |
| `correct_answer "D"` | `The correct_answer "D" is not in the options list.` | YES (substring) |
| `"Grading requires quiz mode"` | `"Grading requires quiz mode. Add quiz = true..."` | YES (substring) |

### 5.3 t.Parallel() / t.Helper()

- **Finding T2-004 (Minor): Tests in validators_item_test.go do NOT call `t.Parallel()`.**
  - Same as T2-002 -- inconsistent with other test files. No correctness impact.

### 5.4 Findings

**T2-004 (Minor): Missing `t.Parallel()` in validators_item_test.go**
- None of the 10 tests call `t.Parallel()`.

---

## 6. File: `internal/resource_form/validators_test.go` (404 lines, 13 tests + helpers)

### 6.1 Correctness Verification

| Test | Verifies | False-Positive Risk | Verdict |
|------|----------|---------------------|---------|
| `TestMutuallyExclusiveValidator_BothSet_ReturnsError` | Both set = error | Low | PASS |
| `TestMutuallyExclusiveValidator_OnlyItems_Passes` | Items only = ok | Low | PASS |
| `TestMutuallyExclusiveValidator_OnlyContentJSON_Passes` | JSON only = ok | Low | PASS |
| `TestMutuallyExclusiveValidator_NeitherSet_Passes` | Neither = ok | Low | PASS |
| `TestMutuallyExclusive_EmptyContentJSON_WithItems_Error` | Empty string + items = error | Low | PASS |
| `TestAcceptingResponsesRequiresPublished_BothTrue_Passes` | Published+accepting = ok | Low | PASS |
| `TestAcceptingResponsesRequiresPublished_AcceptingTruePublishedFalse_Error` | Unpublished+accepting = error | Low | PASS |
| `TestAcceptingResponsesRequiresPublished_AcceptingFalse_Passes` | Not accepting = ok | Low | PASS |
| `TestUniqueItemKeyValidator_UniqueKeys_Passes` | Unique keys = ok | Low | PASS |
| `TestUniqueItemKeyValidator_DuplicateKeys_Error` | Duplicate "q1" = error | Low | PASS |
| `TestUniqueItemKeyValidator_EmptyItems_Passes` | No items = ok | Low | PASS |
| `TestExactlyOneSubBlockValidator_OneSet_Passes` | One sub-block = ok | Low | PASS |
| `TestExactlyOneSubBlockValidator_NoneSet_Error` | Zero sub-blocks = error | Low | PASS |
| `TestExactlyOneSubBlockValidator_TwoSet_Error` | Two sub-blocks = error | Low | PASS |

### 6.2 Error String Cross-Reference

| Test assertion | Production code string | Match? |
|---------------|----------------------|--------|
| `"Cannot use both"` | `Cannot use both "content_json" and "item" blocks...` | YES |
| `"cannot accept responses while unpublished"` | `"A form cannot accept responses while unpublished."` | YES (substring) |
| `Duplicate item_key "q1"` | `Duplicate item_key "q1" found.` | YES (substring) |
| `"exactly one question type"` | `"must have exactly one question type..."` | YES (substring) |

### 6.3 Helper Quality

- `testSchemaResp()`: Gets real schema from `FormResource.Schema()`. Ensures test types match actual schema. Excellent.
- `buildConfig()`: Properly merges user-supplied values with null defaults for all schema attributes. Prevents `tftypes.Object` panics.
- `runValidators()`: Iterates over validators, collects all diagnostics. Uses `t.Helper()`. Correct.
- `expectError()`, `expectNoError()`, `expectErrorContains()`: All use `t.Helper()`. `expectErrorContains` checks both Summary and Detail. Correct.
- `itemBlockType()`: Manually constructs `tftypes.Object` matching the schema. Verified against `model.go` and `state_convert.go:itemObjectType()` -- field names match.
- `mcItem()`, `saItem()`, `paraItem()`, `bareItem()`: Correctly construct item values. Null sub-blocks use `tftypes.NewValue(type, nil)`. Correct.

### 6.4 t.Parallel() / t.Helper()

- **Finding T2-005 (Minor): Tests in validators_test.go do NOT call `t.Parallel()`.**
  - Same consistency issue as T2-002 and T2-004.

### 6.5 Findings

**T2-005 (Minor): Missing `t.Parallel()` in validators_test.go**
- None of the 13 tests call `t.Parallel()`.

---

## 7. Duplicate Function Name Check

Scanned all `func Test*` declarations across all 4 test files in `resource_form/`:
- `crud_test.go`: 23 unique test functions
- `plan_modifiers_test.go`: 8 unique test functions
- `validators_test.go`: 13 unique test functions
- `validators_item_test.go`: 11 unique test functions

**No duplicate function names found.** All 55 test functions have unique names.

Also checked `retry_test.go` (14 tests) and `items_to_requests_test.go` (17 tests) -- all unique within their respective packages.

---

## 8. Cross-Package Type Reference Check

| Test file | References | Exists in production? |
|-----------|-----------|----------------------|
| `retry_test.go` | `WithRetry`, `DefaultRetryConfig`, `RetryConfig`, `APIError`, `NotFoundError`, `RateLimitError`, `IsNotFound` | All exist in `client/retry.go` and `client/errors.go` |
| `items_to_requests_test.go` | `ItemModel`, `MultipleChoiceBlock`, `ShortAnswerBlock`, `ParagraphBlock`, `GradingBlock`, `ItemModelToCreateRequest`, `ItemsToCreateRequests`, `BuildDeleteRequests`, `BuildUpdateInfoRequest`, `BuildQuizSettingsRequest` | All exist in `convert/types.go` and `convert/items_to_requests.go` |
| `crud_test.go` | `FormResource`, `FormResourceModel`, `client.Client`, `client.FormsAPI`, `client.DriveAPI`, `client.APIError`, `client.NotFoundError`, `testutil.MockFormsAPI`, `testutil.MockDriveAPI` | All verified |
| `plan_modifiers_test.go` | `ContentJSONHashModifier` | Exists in `plan_modifiers.go` |
| `validators_test.go` | `MutuallyExclusiveValidator`, `AcceptingResponsesRequiresPublishedValidator`, `UniqueItemKeyValidator`, `ExactlyOneSubBlockValidator`, `ItemModel`, `MultipleChoiceModel`, etc. | All exist |
| `validators_item_test.go` | `OptionsRequiredForChoiceValidator`, `CorrectAnswerInOptionsValidator`, `GradingRequiresQuizValidator`, `FormResource` | All exist |

**No dangling references found.**

---

## 9. Summary of Findings

| ID | Severity | File | Description |
|----|----------|------|-------------|
| T2-001 | Observation | `retry_test.go` | Backoff increase assertion is weak due to jitter tolerance; a constant-backoff bug could still pass |
| T2-002 | Minor | `items_to_requests_test.go` | Missing `t.Parallel()` on all 16+ tests (inconsistent with project convention) |
| T2-003 | Observation | `crud_test.go` | No test verifies composite batch request ordering in Update path (deletes before quiz before creates) |
| T2-004 | Minor | `validators_item_test.go` | Missing `t.Parallel()` on all 10 tests |
| T2-005 | Minor | `validators_test.go` | Missing `t.Parallel()` on all 13 tests |

### Defect Counts by Severity

| Severity | Count |
|----------|-------|
| Critical | 0 |
| Major | 0 |
| Minor | 3 (T2-002, T2-004, T2-005) |
| Observation | 2 (T2-001, T2-003) |

### Overall Assessment

The test suite is **well-constructed and thorough**. Key positives:

1. **F-002 (state_convert.go Title/Description from API):** All CRUD mock `GetFunc` responses correctly return titles matching plan values. The `convertFormModelToTFState` -> `FormToModel` -> `form.Info.Title` chain is properly exercised.

2. **F-004 (batch ordering):** `TestBuildDeleteRequests_MultipleItems_ReverseOrder` directly verifies reverse-order delete indices. Individual ordering pieces are tested, though composite Update ordering is not explicitly asserted.

3. **Error path coverage is excellent:** Every CRUD operation has both success and error tests. Error message assertions check the correct diagnostic summary strings using substring matching.

4. **Mock setup matches production code paths:** `MockFormsAPI` and `MockDriveAPI` implement the correct interfaces with compile-time checks. All mock function signatures match the interface definitions.

5. **Partial state save pattern (F-005) is tested:** `TestCreate_BatchUpdateError_PartialStateSaved` verifies that the form ID is saved to state even when BatchUpdate fails.

6. **No duplicate test function names.** No dangling type/function references. Tests will compile.

The 3 minor findings (missing `t.Parallel()`) are cosmetic consistency issues. The 2 observations are areas where tests could be strengthened but do not represent correctness defects.
