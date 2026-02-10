# Fagan Inspection -- Reader Report

**Role:** Reader (systematic code walkthrough)
**Date:** 2026-02-09
**Scope:** All production `.go` files in `internal/` and `main.go` (excluding `*_test.go`)

---

## 1. `main.go`

**Summary:** Entry point for the Terraform provider binary. Parses a `-debug` flag and starts the provider server using `providerserver.Serve`.

**Defects:** None

**Questions:** None

---

## 2. `internal/client/interfaces.go`

**Summary:** Defines the `FormsAPI` and `DriveAPI` interfaces and the `Client` struct that holds both. Interfaces enable mock-based testing.

**Defects:** None

**Questions:** None

---

## 3. `internal/client/client.go`

**Summary:** `NewClient` creates a real `Client` by building an OAuth2 token source (from service-account JSON, env-var, or ADC), then instantiating the Google Forms and Drive API services. Helper functions resolve credentials and create token sources.

**Defects:** None

**Questions:**
- Q-CLI-1: `resolveCredentials` in `client.go` duplicates the env-var lookup that also exists in `provider.go:resolveCredentials`. When the provider passes an explicit credential string, the client's `resolveCredentials` receives that string and returns it directly. But if the provider passes `""` (ADC mode), the client's `resolveCredentials` will also check `GOOGLE_CREDENTIALS` env var a second time. This double-lookup is redundant but not a bug -- the provider should always have resolved it first. However, if the provider resolves a file path while the client falls through to the env var, the two layers could disagree. **Verdict: Minor inconsistency, not a runtime bug given current call paths.**

---

## 4. `internal/client/errors.go`

**Summary:** Defines structured error types (`APIError`, `NotFoundError`, `RateLimitError`) with `Error()` and `Unwrap()` methods. Provides helper predicates `IsNotFound`, `IsRateLimit`, and `ErrorStatusCode`.

### Defects

| # | Severity | Description |
|---|----------|-------------|
| ERR-1 | **Minor** | `NotFoundError.Unwrap()` (line 37-39) creates a *new* `APIError` with `StatusCode: 404` each time it is called. The `Err` field of that `APIError` is zero-valued (`nil`). This means `errors.As(notFoundErr, &apiErr)` will succeed and give an `APIError` whose `.Err` is `nil` -- any code that calls `apiErr.Unwrap()` on it gets `nil`. This is not incorrect per se, but it means the error chain terminates: `errors.Is(notFoundErr, someWrappedErr)` will not traverse further. If a caller wraps a `NotFoundError` around another error, the inner error is lost. Same issue with `RateLimitError.Unwrap()` (line 51-53). |

**Questions:**
- Q-ERR-1: `ErrorStatusCode` checks for `NotFoundError` and `RateLimitError` before `APIError`. This is correct for precedence, but is there a scenario where an `APIError` wraps a `NotFoundError`? If so, the function would return the outer status code, not 404. Currently this cannot happen given the constructors in `forms_api.go`, so it is fine.

---

## 5. `internal/client/forms_api.go`

**Summary:** `FormsAPIClient` implements `FormsAPI` using the real Google Forms REST API. Methods: `Create`, `Get`, `BatchUpdate`, `SetPublishSettings`. Includes `wrapGoogleAPIError` and `mapStatusToError` for error classification.

### Defects

| # | Severity | Description |
|---|----------|-------------|
| FORMS-1 | **Minor** | `BatchUpdate` (line 82) mutates the caller's `*forms.BatchUpdateFormRequest` by setting `req.IncludeFormInResponse = true` on line 82. This is a side effect on the caller's struct. In the current codebase the caller constructs the request inline immediately before calling `BatchUpdate`, so this is harmless. But if the request were ever reused (e.g., for retry logic outside `WithRetry`), the mutation could cause confusion. |
| FORMS-2 | **Minor** | `mapStatusToError` (line 142-151) always sets `NotFoundError.Resource = "form"` regardless of whether the call came from the Forms API or the Drive API (via `wrapDriveAPIError` in `drive_api.go`, which delegates to the same `mapStatusToError`). For Drive API 404s, the resource label should be "file" or "drive file", not "form". This produces a misleading error message when `Drive.Delete` gets a 404 on a non-form file. |

**Questions:**
- Q-FORMS-1: `Create` retries on transient errors. Is it safe to retry a Forms Create call? If the first call succeeds but the response is lost (network timeout), the retry will create a duplicate form. The Google Forms API `Create` is not idempotent. This is a general concern with retrying non-idempotent operations, mitigated by the timeout being relatively short.

---

## 6. `internal/client/drive_api.go`

**Summary:** `DriveAPIClient` implements `DriveAPI` with a single `Delete` method. Uses `WithRetry` and treats 404 as success (already deleted).

### Defects

| # | Severity | Description |
|---|----------|-------------|
| DRV-1 | **Minor** | `wrapDriveAPIError` (line 57-64) calls `mapStatusToError` from `forms_api.go`. As noted in FORMS-2, this sets `Resource = "form"` for all 404s, even for Drive operations where the resource is a file. |

**Questions:** None beyond FORMS-2 above.

---

## 7. `internal/client/retry.go`

**Summary:** Implements exponential backoff with jitter. `WithRetry` retries on 429/5xx status codes, respects context cancellation, and uses `sleepWithContext` for interruptible delays.

### Defects

| # | Severity | Description |
|---|----------|-------------|
| RET-1 | **Style** | `rand.Float64()` on line 91 uses the global `math/rand` source. In Go 1.20+, the global source is automatically seeded, so this is fine. The `//nolint:gosec` comment is appropriate. No issue. |

**Questions:**
- Q-RET-1: `isRetryable` (line 63-71) returns `false` for errors that have `ErrorStatusCode == 0` (non-API errors like network timeouts, DNS failures). These transient network errors are NOT retried. This may be intentional (only retry on known API error codes), but network-level errors like `connection reset` or `i/o timeout` are arguably retryable. **Verdict: Design choice, flag for discussion.**
- Q-RET-2: `wrapContextError` (line 110-115) wraps the context error with `%w` and the last API error with `%v`. This means `errors.Is(err, context.Canceled)` works, but the API error is only in the message string, not in the error chain. This is intentional (context error is the primary cause), but callers cannot use `errors.As` to extract the API error from the wrapped result.

---

## 8. `internal/convert/types.go`

**Summary:** Defines plain Go struct types (`ItemModel`, `MultipleChoiceBlock`, `ShortAnswerBlock`, `ParagraphBlock`, `GradingBlock`, `FormModel`) used as an intermediate representation between Terraform types and Google API types to avoid circular imports.

**Defects:** None

**Questions:** None

---

## 9. `internal/convert/items_to_requests.go`

**Summary:** Converts `ItemModel` instances to Google Forms API `Request` objects for creating items, deleting items, updating form info, and toggling quiz settings. Contains builders for each question type.

### Defects

| # | Severity | Description |
|---|----------|-------------|
| ITR-1 | **Major** | `buildShortAnswer` (line 58-65) sets `TextQuestion.Paragraph = false`. In the Google Forms API, `Paragraph` is a boolean that defaults to `false`. However, when sending JSON, Go's `omitempty` behavior means `false` values may be omitted by the Google API client library if the field has `json:",omitempty"`. Checking the `forms.TextQuestion` struct: the `Paragraph` field has tag `json:"paragraph,omitempty"`. This means `Paragraph: false` will be omitted from the JSON, which defaults to `false` on the server. So this works correctly. **Revised: Not a bug**, the omit-false matches the intended semantics. Downgraded to no defect. |
| ITR-2 | **Minor** | `BuildUpdateInfoRequest` (line 128-138) uses `UpdateMask: "title,description"`. If the user sets `description = ""` (empty string), the API will clear the description. This is correct behavior. But the mask always includes both fields even if only one changed. For Update operations, this means every update overwrites both title and description even if only one was modified. This is wasteful but not incorrect. |
| ITR-3 | **Minor** | `applyGrading` (line 78-96) sets `CorrectAnswers` only when `g.CorrectAnswer != ""`. If a user wants to clear a previously-set correct answer (set it to empty), this code will skip setting `CorrectAnswers`, effectively leaving the old answer in place. However, since the Update strategy is replace-all (delete all items then re-create), this is not a problem in practice. |

**Questions:**
- Q-ITR-1: `ItemModelToCreateRequest` uses `int64(index)` for the Location.Index. The Google Forms API uses 0-based indexing. The caller passes `i` from a range loop (0-based). This is correct.

---

## 10. `internal/convert/form_to_model.go`

**Summary:** Converts a Google Forms API `Form` response into a `FormModel`. Handles item-key resolution (from existing state or auto-generated), and converts each question type (RADIO choice, short answer, paragraph) with optional grading.

### Defects

| # | Severity | Description |
|---|----------|-------------|
| FTM-1 | **Major** | `convertChoiceQuestion` (line 87-99) assumes `q.ChoiceQuestion.Options` is non-nil and non-empty. If the Google API returns a RADIO question with zero options (e.g., a form created outside Terraform with no options yet), `len(q.ChoiceQuestion.Options)` is 0, which is safe (creates an empty slice). However, if `q.ChoiceQuestion` itself is nil while the switch case matched `q.ChoiceQuestion != nil`, that cannot happen. **Revised: No nil dereference risk** given the switch guard on line 73. Downgraded. |
| FTM-2 | **Minor** | `FormItemToItemModel` (line 60-84) returns `nil, nil` for unsupported question types (e.g., `CHECKBOX`, `DROP_DOWN`, `SCALE`). The caller `FormToModel` (line 38-39) skips `nil` items with `continue`. This means items in the API response are silently dropped. During `Read`, this causes the Terraform state to have fewer items than the actual form, leading to drift that Terraform cannot reconcile. For the supported subset this is intentional, but users who add items outside Terraform will see silent data loss in state. |
| FTM-3 | **Minor** | `convertGrading` (line 122-138) takes only the first element of `g.CorrectAnswers.Answers` (line 129). If the Google API returns multiple correct answers (e.g., for checkbox questions), all but the first are silently discarded. Since the provider currently only supports RADIO (single correct answer), this is acceptable but fragile for future extension. |

**Questions:**
- Q-FTM-1: `FormToModel` (line 15) receives `existingKeyMap` which can be `nil`. The `resolveItemKey` function (line 49-56) handles the nil case correctly by falling back to `"item_N"`. No issue.
- Q-FTM-2: `FormToModel` does not populate `FormModel.Quiz` when `form.Settings` is nil (line 28). This defaults `Quiz` to `false`, which is correct (Google Forms default is non-quiz).

---

## 11. `internal/convert/json_mode.go`

**Summary:** Provides functions for the `content_json` attribute: parsing declarative JSON into Forms API Items, normalizing JSON for comparison, and computing SHA-256 hashes for semantic equality checking.

### Defects

| # | Severity | Description |
|---|----------|-------------|
| JSON-1 | **Major** | `NormalizeJSON` (line 47-57) uses `json.Unmarshal` into `interface{}` then `json.Marshal`. Go's `encoding/json` does NOT guarantee sorted keys when marshaling a `map[string]interface{}`. The Go spec says map iteration order is randomized. Therefore, `NormalizeJSON` does NOT produce canonical output -- two calls with the same input can produce different key orderings. The doc comment on line 44-46 claims "sorted keys" but this is incorrect. As a result, `HashJSON` (which depends on `NormalizeJSON`) can produce different hashes for the same semantic JSON, causing spurious diffs in Terraform plans. |

**Questions:**
- Q-JSON-1: `ParseDeclarativeJSON` (line 16-22) deserializes directly into `[]*forms.Item`. If the JSON contains fields not in the `forms.Item` struct, they are silently ignored by `json.Unmarshal`. This is standard Go behavior but means typos in JSON field names will not produce errors.

---

## 12. `internal/provider/provider.go`

**Summary:** Implements the Terraform provider. Handles credentials resolution (inline JSON, file path, env var, or ADC), creates the API client, and registers the `googleforms_form` resource.

### Defects

| # | Severity | Description |
|---|----------|-------------|
| PROV-1 | **Minor** | `resolveCredentialValue` (line 135-157): If the value does not start with `{` and is not a valid file path, it falls through to returning the raw value (line 152). This means a typo in a file path (e.g., `/tmp/nonexistent.json`) will be silently passed as credential JSON, causing a confusing "parsing service account credentials" error downstream instead of a clear "file not found" error. The `tflog.Warn` on line 149 logs the issue, but the user gets no diagnostic error. |
| PROV-2 | **Style** | `resolveCredentials` in `provider.go` (line 115) shadows the function name `resolveCredentials` in `client/client.go` (line 75). They are in different packages so this is not a compile error, but it is confusing for maintainers. |

**Questions:**
- Q-PROV-1: `Configure` (line 108-109) sets both `resp.DataSourceData` and `resp.ResourceData` to the same `*client.Client`. Currently there are no data sources, so `DataSourceData` is unused. Not a bug, just forward-looking.

---

## 13. `internal/resource_form/model.go`

**Summary:** Defines the Terraform state model structs (`FormResourceModel`, `ItemModel`, `MultipleChoiceModel`, `ShortAnswerModel`, `ParagraphModel`, `GradingModel`) using `terraform-plugin-framework` types.

**Defects:** None

**Questions:** None

---

## 14. `internal/resource_form/schema.go`

**Summary:** Defines the Terraform schema for `googleforms_form` including all attributes, nested blocks for item types (multiple_choice, short_answer, paragraph), grading, and plan modifiers.

### Defects

| # | Severity | Description |
|---|----------|-------------|
| SCH-1 | **Minor** | `description` attribute (line 44-47) is `Optional: true` but not `Computed: true`. If the Google API returns a non-empty description that was set outside Terraform (e.g., via the Forms UI), and the user's HCL has no `description`, the Terraform state will have `description = ""` (from plan), but the API has a non-empty description. On the next `Read`, `convertFormModelToTFState` (state_convert.go line 99-118) copies `plan.Description` (which is the state's description, `""`), so the API-side description is ignored. This means manual description changes are invisible to Terraform. This is arguably correct "Terraform owns the resource" behavior, but it also means `description` changes made outside Terraform cause silent drift that is never detected or reported. |

**Questions:**
- Q-SCH-1: `content_json` (line 66-72) has `ContentJSONHashModifier` as its only plan modifier. It does not have `UseStateForUnknown` -- is this intentional? During create, the config value is used directly, so unknown would only occur if it references another resource. This is fine.

---

## 15. `internal/resource_form/resource.go`

**Summary:** Defines `FormResource` struct, `NewFormResource` factory, `Metadata`, `Configure` (extracts the client from provider data), and `ConfigValidators` (returns all 7 validators).

**Defects:** None

**Questions:** None

---

## 16. `internal/resource_form/validators.go`

**Summary:** Implements 4 config validators: `MutuallyExclusiveValidator` (content_json vs items), `AcceptingResponsesRequiresPublishedValidator`, `UniqueItemKeyValidator`, `ExactlyOneSubBlockValidator`.

### Defects

| # | Severity | Description |
|---|----------|-------------|
| VAL-1 | **Minor** | `UniqueItemKeyValidator` (line 147-148): Checks `key == ""` before `item.ItemKey.IsNull()`. Since `ValueString()` returns `""` for null/unknown values, this short-circuits correctly. However, the ordering is misleading -- checking `IsNull()` first would be more idiomatic. No runtime impact. |
| VAL-2 | **Minor** | `ExactlyOneSubBlockValidator` (line 198-208): When it finds an item with `count != 1`, the error message does not indicate WHICH item has the problem. For a form with many items, the user cannot tell which item to fix. The error should include the item_key or index. |
| VAL-3 | **Minor** | `ExactlyOneSubBlockValidator`: `countSubBlocks` uses nil checks on the pointer fields (line 212-224). With `terraform-plugin-framework`, `SingleNestedBlock` fields are set to a non-nil zero-value struct when the block is present but empty in HCL. However, when the block is absent, the Go struct field will be `nil`. This is correct behavior for detecting block presence. |

**Questions:**
- Q-VAL-1: `MutuallyExclusiveValidator` (line 57) checks `len(items.Elements()) > 0` to determine if items are present. If the user writes `item {}` (an empty block), this would still count as having items and trigger the mutual exclusivity error. Is this the desired behavior? An empty item block is invalid anyway (fails ExactlyOneSubBlockValidator), so this is fine in practice.

---

## 17. `internal/resource_form/validators_item.go`

**Summary:** Implements 3 more validators: `OptionsRequiredForChoiceValidator`, `CorrectAnswerInOptionsValidator`, `GradingRequiresQuizValidator`.

### Defects

| # | Severity | Description |
|---|----------|-------------|
| VALI-1 | **Minor** | `GradingRequiresQuizValidator` (line 170-172): When `quiz` is null or unknown, the validator returns early without checking for grading blocks. This means during `terraform plan` with a `quiz` value that depends on another resource (unknown), grading blocks will not be validated. The validation only occurs when `quiz` has a known `false` value. This is acceptable behavior (you cannot validate against unknowns), but worth documenting. |

**Questions:** None

---

## 18. `internal/resource_form/plan_modifiers.go`

**Summary:** `ContentJSONHashModifier` suppresses Terraform diffs for `content_json` when the normalized JSON hashes match, by comparing SHA-256 hashes of the config and state values.

### Defects

| # | Severity | Description |
|---|----------|-------------|
| PM-1 | **Major** | This modifier depends on `convert.HashJSON` which depends on `convert.NormalizeJSON`. As documented in JSON-1 above, `NormalizeJSON` does NOT produce deterministic output because Go's `json.Marshal` does not sort map keys. Therefore, the hash comparison in `ContentJSONHashModifier` may produce false negatives (hashes differ for semantically identical JSON), causing spurious diffs. It will never produce false positives (hashes match for different JSON), so this is a UX problem, not a correctness/data-loss problem. |

**Questions:** None

---

## 19. `internal/resource_form/state_convert.go`

**Summary:** Bidirectional conversion between Terraform framework types and `convert` package types. Includes `tfItemsToConvertItems`, `convertFormModelToTFState`, `buildItemKeyMap`, `convertItemsToTFList`, and type definitions for `itemObjectType`/`gradingObjectType`.

### Defects

| # | Severity | Description |
|---|----------|-------------|
| SC-1 | **Major** | `convertFormModelToTFState` (line 99-118) copies `plan.Published` and `plan.AcceptingResponses` directly to the state without reading them from the API. The Google Forms API does not return publish state in the `Form` response (publish settings are managed via a separate endpoint). This means the state always reflects the *plan* values, not the actual API state. If someone changes publish settings outside Terraform (via the UI), Terraform will never detect the drift. This is arguably intentional (the API does not expose this in `Get`), but it should be documented. |
| SC-2 | **Minor** | `convertFormModelToTFState` (line 113-115) constructs `edit_uri` by concatenating the form ID into a URL pattern. If the Google Forms URL pattern ever changes, this hardcoded URL will be wrong. More importantly, this is a computed field that should ideally come from the API response. The Forms API does not return an edit URI, so this is the best available approach. |
| SC-3 | **Minor** | `convertItemsToTFList` (line 149-167): When `items` is empty (length 0), returns `types.ListNull(itemObjectType())`. This sets the item list to null in state. However, in Terraform, a null list and an empty list have different semantics. If the user's config has no `item` blocks (null), this is correct. But if the API returns zero items after the user deleted all items, setting null could cause a plan diff on the next apply if the user still has `item` blocks in their config (the plan would have an empty list, not null). In practice, the replace-all strategy means the plan and API item counts should always match. |
| SC-4 | **Style** | `tfItemsToConvertItems` (line 38-81) uses sequential `if` statements instead of `switch`/`case` for MultipleChoice/ShortAnswer/Paragraph. Since `ExactlyOneSubBlockValidator` guarantees exactly one is set, only one branch will execute. But the code does not `else if`, meaning all three conditions are evaluated even after one matches. This is harmless but slightly wasteful. |

**Questions:**
- Q-SC-1: `convertGradingToTF` (line 217-241) sets `CorrectAnswer`, `FeedbackCorrect`, `FeedbackIncorrect` to `types.StringNull()` when the corresponding convert field is empty string. This ensures optional null-vs-empty semantics are preserved. However, if the API returns an explicit empty string for these fields, the state will have null, and the plan (with explicit `""` in config) will have `""`, causing a perpetual diff. Is this a real scenario?

---

## 20. `internal/resource_form/crud_create.go`

**Summary:** Implements the `Create` CRUD operation. Creates a bare form (title only), saves partial state (form ID), then batch-updates settings/items, sets publish state, reads back final state, and saves.

### Defects

| # | Severity | Description |
|---|----------|-------------|
| CR-1 | **Major** | Positional key map construction (line 176-191): When `keyMap == nil` (first create, no existing google_item_ids), the code builds a positional map from `finalForm.Items`. But `finalForm` is read back after the batchUpdate, and `FormToModel` (called on line 193) uses this key map. The issue is: `finalForm.Items` may include items that were skipped by `FormToModel` (unsupported types return `nil`). If the form has unsupported items (e.g., page breaks inserted by the API), the positional indices will be off, assigning the wrong `item_key` to the wrong question. However, since this is a Create from Terraform (no external items), the risk is low. |
| CR-2 | **Minor** | `BatchUpdate` on line 124 receives a request object, but `FormsAPIClient.BatchUpdate` (forms_api.go line 82) mutates this object by setting `IncludeFormInResponse = true`. The Create code also sets `IncludeFormInResponse: true` on line 122. This double-set is harmless but redundant. |
| CR-3 | **Minor** | If the batchUpdate fails (line 125-131), the function returns with an error, but the partial state (with only the form ID) was already saved on line 59. This is the intentional partial-state-save pattern. However, the state at this point has the *plan* values for title, description, quiz, etc., even though only the bare form was actually created. On a subsequent `terraform apply`, the Read will fetch the actual state (which only has the title), and Terraform will detect drift and re-apply. This is acceptable but could confuse users who see state values that don't match reality after a partial failure. |

**Questions:**
- Q-CR-1: Line 82 sets `IncludeFormInResponse = true` in the batch request, but the response on line 124 is discarded (assigned to `_`). The final form state is fetched via a separate `Get` call on line 154. This is correct but wasteful -- the BatchUpdate response already contains the updated form.

---

## 21. `internal/resource_form/crud_read.go`

**Summary:** Implements the `Read` CRUD operation. Fetches the form, handles 404 by removing from state, converts API response to TF state using key maps.

### Defects

| # | Severity | Description |
|---|----------|-------------|
| RD-1 | **Minor** | In content_json mode (line 80-84), `Read` preserves the existing `state.ContentJSON` without comparing it to the actual API response. This means if someone modifies the form items outside Terraform (via the UI), the state's `content_json` will still show the old value. Drift is completely invisible in content_json mode. The plan modifier (hash comparison) only compares config vs state, not state vs API. |

**Questions:**
- Q-RD-1: `convertFormModelToTFState` receives `state` (the old state) as the `plan` parameter. This means fields like `Published`, `AcceptingResponses`, `Quiz`, `ContentJSON` are copied from the OLD state, not from the API. For `Published` and `AcceptingResponses`, this is necessary (API doesn't return these). For `Quiz`, the API does return this via `form.Settings.QuizSettings.IsQuiz`, but `convertFormModelToTFState` ignores it and uses the old state value. **This means quiz mode changes made outside Terraform are not detected.** This should be classified as a defect.

---

## 22. `internal/resource_form/crud_update.go`

**Summary:** Implements the `Update` CRUD operation. Uses a "replace-all" strategy: fetches current form, builds batch requests to delete all existing items and re-create from plan, updates settings, and reads back final state.

### Defects

| # | Severity | Description |
|---|----------|-------------|
| UP-1 | **Major** | The replace-all strategy (delete all items then create new ones) is executed in a single `batchUpdate` call. This is atomic from the API's perspective. However, the delete requests use index-based locations (from `BuildDeleteRequests`), and the create requests also use index-based locations. In a single batch, the Google Forms API processes requests sequentially. After all deletes execute (removing items from highest index to lowest), the form has 0 items. Then creates add items at indices 0, 1, 2, etc. This sequencing is correct. **No defect** -- the reverse-order delete followed by 0-indexed create is sound. |
| UP-2 | **Minor** | Similar to CR-1: positional key map construction (line 178-193) after update. The same risk of index misalignment exists if unsupported items are present. Since Update replaces all items, and the items come from the plan, unsupported items should not be present. Low risk. |
| UP-3 | **Minor** | `BatchUpdate` response is discarded (line 131, `_`), and a separate `Get` call is made (line 164). Same redundancy as CR-2/Q-CR-1. |

**Questions:**
- Q-UP-1: If the form has quiz grading on items, and the user changes `quiz = false` in the plan, the quiz settings request is added to the batch (line 61). But the item delete+create requests are also in the same batch. Does the Google Forms API allow removing quiz mode while items still have grading? The ordering within the batch is: (1) update info, (2) quiz settings change, (3) delete items, (4) create items. If disabling quiz mode fails because existing items have grading, the entire batch fails. The delete-first approach would avoid this, but deletes come after the settings change in the request list. **This could cause a batch failure when disabling quiz mode on a form with graded items.**

---

## 23. `internal/resource_form/crud_delete.go`

**Summary:** Implements the `Delete` CRUD operation. Deletes the form via Drive API. Handles 404 (already deleted) gracefully.

**Defects:** None

**Questions:** None

---

## 24. `internal/resource_form/import.go`

**Summary:** Implements `ImportState` using the framework's `ImportStatePassthroughID`, which passes the import ID to the `id` attribute and triggers a `Read`.

**Defects:** None

**Questions:**
- Q-IMP-1: After import, `Read` will execute. The state will have no `ContentJSON` and no `Items` (only `id`). The `Read` function will build a `keyMap` from `state.Items` (which is null/empty), resulting in `keyMap == nil`. Then `FormToModel` will auto-generate keys as `item_0`, `item_1`, etc. Then `convertFormModelToTFState` copies `plan.Published` from the state (which is null after import). This means `Published` and `AcceptingResponses` will be null in state after import, which defaults to `false`. If the actual form is published, this will show a diff on the next plan. This is expected import behavior but worth noting.

---

## 25. `internal/testutil/mock_forms.go`

**Summary:** Configurable mock implementation of `FormsAPI` for unit testing. Each method delegates to a function field if set, otherwise returns sensible defaults.

**Defects:** None

**Questions:** None

---

## 26. `internal/testutil/mock_drive.go`

**Summary:** Configurable mock implementation of `DriveAPI` for unit testing. Single `Delete` method with function field.

**Defects:** None

**Questions:** None

---

## 27. `internal/testutil/sweeper.go`

**Summary:** Placeholder file for a test resource sweeper. Contains only comments and a TODO. No executable code.

### Defects

| # | Severity | Description |
|---|----------|-------------|
| SWP-1 | **Style** | File contains no Go declarations (no functions, no init, no vars). It only has comments. While this compiles (a file with just a package declaration is valid Go), it serves no purpose at runtime. This should either be implemented or removed. |

**Questions:** None

---

## Defect Summary

### Major Defects

| ID | File | Line(s) | Description |
|----|------|---------|-------------|
| JSON-1 | `convert/json_mode.go` | 47-57 | `NormalizeJSON` does not produce deterministic output; Go's `json.Marshal` does not sort map keys. This breaks the hash-based diff suppression in `ContentJSONHashModifier`. |
| PM-1 | `resource_form/plan_modifiers.go` | 33-58 | Direct consequence of JSON-1: hash comparison for `content_json` is unreliable, causing spurious Terraform plan diffs. |
| Q-RD-1 (promoted) | `resource_form/crud_read.go` + `state_convert.go` | 70, 99-106 | `Read` copies `Quiz`, `Title`, `Description` from old state instead of API response via `convertFormModelToTFState`. Changes to `quiz` mode made outside Terraform are never detected. `Title` and `Description` are similarly not refreshed from the API. |

### Minor Defects

| ID | File | Line(s) | Description |
|----|------|---------|-------------|
| ERR-1 | `client/errors.go` | 37-39, 51-53 | `NotFoundError.Unwrap()` and `RateLimitError.Unwrap()` create new orphaned `APIError` instances; inner error chain is lost. |
| FORMS-1 | `client/forms_api.go` | 82 | `BatchUpdate` mutates caller's request struct by setting `IncludeFormInResponse`. |
| FORMS-2 | `client/forms_api.go` | 145 | `mapStatusToError` hardcodes `Resource: "form"` for all 404s, including Drive operations. |
| DRV-1 | `client/drive_api.go` | 57-64 | Same as FORMS-2, surfaced through `wrapDriveAPIError`. |
| ITR-2 | `convert/items_to_requests.go` | 128-138 | `UpdateMask` always includes both `title` and `description` even when only one changed. |
| ITR-3 | `convert/items_to_requests.go` | 78-96 | Cannot clear a previously-set correct answer (mitigated by replace-all strategy). |
| FTM-2 | `convert/form_to_model.go` | 60-84 | Unsupported item types silently dropped; causes invisible state drift if form has non-question items. |
| FTM-3 | `convert/form_to_model.go` | 129 | Only first correct answer preserved for multi-answer questions. |
| PROV-1 | `provider/provider.go` | 135-157 | Invalid file path silently falls through as raw credential JSON, producing confusing errors. |
| SCH-1 | `resource_form/schema.go` | 44-47 | `description` not marked `Computed`, preventing drift detection for externally-set descriptions. |
| VAL-2 | `resource_form/validators.go` | 198-208 | `ExactlyOneSubBlockValidator` error message does not identify which item. |
| VALI-1 | `resource_form/validators_item.go` | 170-172 | `GradingRequiresQuizValidator` skips validation when `quiz` is unknown. |
| SC-1 | `resource_form/state_convert.go` | 99-118 | `Published`, `AcceptingResponses` always from plan, never from API (API limitation). |
| SC-3 | `resource_form/state_convert.go` | 149-167 | Empty items list sets null instead of empty list, potential plan diff. |
| RD-1 | `resource_form/crud_read.go` | 80-84 | `content_json` drift is completely invisible -- state never compared to API. |
| CR-1 | `resource_form/crud_create.go` | 176-191 | Positional key map could misalign if unsupported item types are present. |
| CR-3 | `resource_form/crud_create.go` | 59-62 | Partial state save includes unverified plan values for non-computed fields. |
| UP-2 | `resource_form/crud_update.go` | 178-193 | Same positional key map risk as CR-1. |

### Style Defects

| ID | File | Line(s) | Description |
|----|------|---------|-------------|
| PROV-2 | `provider/provider.go` | 115 | Function name `resolveCredentials` shadows same name in `client/client.go`. |
| SC-4 | `resource_form/state_convert.go` | 38-81 | Sequential `if` instead of `else if` or `switch` for mutually exclusive blocks. |
| SWP-1 | `testutil/sweeper.go` | 1-20 | Placeholder file with no executable code; should be implemented or removed. |

---

## Open Questions Summary

| ID | File | Question |
|----|------|----------|
| Q-CLI-1 | `client/client.go` | Double env-var lookup for `GOOGLE_CREDENTIALS` between provider and client layers. |
| Q-ERR-1 | `client/errors.go` | Can `APIError` ever wrap `NotFoundError` causing precedence issues in `ErrorStatusCode`? |
| Q-FORMS-1 | `client/forms_api.go` | Is retrying non-idempotent `Create` calls safe? Risk of duplicate forms. |
| Q-RET-1 | `client/retry.go` | Network-level errors (connection reset, timeouts) are not retried. Design choice? |
| Q-RET-2 | `client/retry.go` | Context-wrapped errors lose the API error from the chain (only in message string). |
| Q-JSON-1 | `convert/json_mode.go` | Silent field name typo tolerance in declarative JSON. |
| Q-VAL-1 | `resource_form/validators.go` | Empty item block triggers mutual exclusivity. Correct but surprising? |
| Q-SC-1 | `resource_form/state_convert.go` | Empty string vs null semantics for optional grading fields. |
| Q-CR-1 | `resource_form/crud_create.go` | BatchUpdate response discarded; redundant Get call. |
| Q-UP-1 | `resource_form/crud_update.go` | Batch ordering when disabling quiz mode -- settings change before item delete may fail. |
| Q-IMP-1 | `resource_form/import.go` | Post-import state has null published/accepting, may show unexpected diff. |

---

## Reader's Summary

The codebase is well-structured with clear separation of concerns (client, convert, resource, testutil). Error handling follows consistent patterns, and the partial-state-save on Create is a good practice.

The most significant defect is **JSON-1/PM-1**: the `NormalizeJSON` function does not produce deterministic output, which undermines the `content_json` hash-based diff suppression. This will cause user-visible spurious diffs on every `terraform plan`.

The second most significant area of concern is **drift detection**: the `convertFormModelToTFState` function copies several fields from the plan/state rather than the API response (`Quiz`, `Title`, `Description`, `Published`, `AcceptingResponses`). For `Published` and `AcceptingResponses`, this is forced by the API limitation. But for `Quiz`, `Title`, and `Description`, the API does return these values and they should be used for drift detection.

The **Q-UP-1** question about batch ordering when disabling quiz mode deserves investigation -- it could cause runtime failures.

---

*End of Reader Report*

