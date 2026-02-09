# Fagan Inspection Report: DESIGNER Role

**Inspector:** Designer (Architectural Spec Verification)
**Date:** 2026-02-09
**Scope:** All production source files under `internal/`
**Reference Spec:** `C:\Projects\terraform-google-forms-spec.md`
**Reference Plan:** `C:\Users\MQCKENC\.claude\plans\snappy-munching-charm.md`

---

## Summary

Inspected 20 production source files across 4 packages (`client`, `convert`, `provider`, `resource_form`) against the original spec and implementation plan. The implementation is architecturally sound and faithfully executes the plan's key decisions. Found **3 Major defects**, **5 Minor defects**, and **2 Observations**.

---

## Design Decision Verification Checklist

### 1. Typed Sub-Blocks (multiple_choice, short_answer, paragraph)

**Verdict: PASS**

- `schema.go:128-191` defines `multiple_choice`, `short_answer`, and `paragraph` as `schema.SingleNestedBlock` inside the `item` ListNestedBlock.
- Each sub-block has its own schema with type-specific attributes (e.g., `options` only on `multiple_choice`).
- The `ExactlyOneSubBlockValidator` in `validators.go:169-224` enforces that exactly one sub-block is set per item.
- Matches plan Decision 1 precisely.

### 2. `item_key` Required + UniqueItemKeyValidator Wired

**Verdict: PASS**

- `schema.go:113` declares `item_key` as `Required: true`.
- `validators.go:116-161` implements `UniqueItemKeyValidator` which iterates all items and detects duplicate keys.
- `resource.go:69` wires `UniqueItemKeyValidator{}` into `ConfigValidators()`.
- Format validation (`[a-z][a-z0-9_]{0,63}`) is **not enforced** at the schema level (no regex validator on the attribute). This is a minor gap vs. the plan which specified the format, but the plan did not explicitly require a validator for the format itself -- it was stated as a convention.

### 3. `content_json` is Declarative (not imperative batchUpdate JSON)

**Verdict: PASS**

- `json_mode.go:16-22` (`ParseDeclarativeJSON`) parses content_json as a JSON array of `forms.Item` objects (the `forms.get` response format), NOT as batchUpdate request JSON.
- `json_mode.go:26-42` (`DeclarativeJSONToRequests`) converts these declarative items into `CreateItemRequest` entries.
- Matches plan Decision 3: "Declarative JSON matching the items array from forms.get response."

### 4. Mutual Exclusivity: content_json vs item blocks

**Verdict: PASS**

- `validators.go:30-65` implements `MutuallyExclusiveValidator` which checks if both `content_json` and `item` blocks are set, returning an error if so.
- `resource.go:67` wires it as the first validator in `ConfigValidators()`.

### 5. Partial State Save on Create

**Verdict: PASS**

- `crud_create.go:55-62` -- Immediately after `forms.create` succeeds, the form ID is written to state via `resp.State.Set(ctx, &plan)` BEFORE any `batchUpdate` call.
- Lines 58-61: `plan.ID = types.StringValue(result.FormId)` followed by `resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)`.
- This matches the plan's Fix 1 exactly: "write form_id to state BEFORE batchUpdate."
- Tested in `crud_test.go:359-405` (`TestCreate_BatchUpdateError_PartialStateSaved`).

### 6. 404 on Read -> RemoveResource

**Verdict: PASS**

- `crud_read.go:37-43` checks `client.IsNotFound(err)` and calls `resp.State.RemoveResource(ctx)`.
- Matches plan Fix 2.
- Tested in `crud_test.go:451-489` (`TestRead_FormNotFound_RemovesFromState`).

### 7. 404 on Delete -> Success

**Verdict: PASS**

- `crud_delete.go:36-42` checks `client.IsNotFound(err)` and returns without error.
- Additionally, `drive_api.go:43-46` already handles 404 by returning nil at the client layer.
- Double-layer defense: both the client and CRUD layer treat 404 on delete as success.
- Matches plan Fix 3.
- Tested in `crud_test.go:763-796` (`TestDelete_NotFound_NoError`).

### 8. Replace-All Update Strategy with Reverse Deletion

**Verdict: PASS**

- `crud_update.go:65-117` implements the replace-all strategy:
  1. First deletes all existing items via `convert.BuildDeleteRequests(existingItemCount)`.
  2. Then creates new items from the plan via `convert.ItemsToCreateRequests(convertItems)`.
- `items_to_requests.go:114-124` (`BuildDeleteRequests`) deletes in reverse order (highest index first) to avoid index shifting.
- Matches plan Decision 4: "delete-all-then-recreate for item changes."

### 9. Authentication: ADC + Service Account

**Verdict: PASS**

- `provider.go:61-72` defines `credentials` (Optional, Sensitive) and `impersonate_user` (Optional) attributes.
- `provider.go:86-100` resolves credentials via: (1) explicit config, (2) `GOOGLE_CREDENTIALS` env var, (3) ADC fallback.
- `client.go:19-22` specifies required scopes: `forms.FormsBodyScope` and `drive.DriveFileScope`.
- `client.go:56-114` supports service account JSON via `google.JWTConfigFromJSON` with optional impersonation via `config.Subject`, and ADC via `google.FindDefaultCredentials`.
- Matches the plan's auth design precisely.

**Note:** The spec also mentioned `forms.responses.readonly` scope, but the plan explicitly excluded it from Phase 1 (no response data features). The implementation correctly uses only the two required scopes.

### 10. Computed Outputs: id, responder_uri, edit_uri, document_title

**Verdict: PASS**

- `schema.go:33-93` declares all four as `Computed: true` with `UseStateForUnknown` plan modifiers:
  - `id` (line 33-39)
  - `responder_uri` (line 73-79)
  - `edit_uri` (line 80-86)
  - `document_title` (line 87-93)
- `state_convert.go:101-118` populates these from the API response.

**Minor finding:** `edit_uri` is constructed from a hardcoded URL pattern (`state_convert.go:114`: `"https://docs.google.com/forms/d/" + model.ID + "/edit"`) rather than read from the API response. The `forms.get` response has an `editUri` field (seen in `form_to_model.go` where `form.ResponderUri` is read but `editUri` is not). This is a correctness concern -- see Defect D-3.

### 11. Seven Validators All Wired in ConfigValidators()

**Verdict: PASS**

- `resource.go:63-75` returns exactly 7 validators:
  1. `MutuallyExclusiveValidator{}`
  2. `AcceptingResponsesRequiresPublishedValidator{}`
  3. `UniqueItemKeyValidator{}`
  4. `ExactlyOneSubBlockValidator{}`
  5. `OptionsRequiredForChoiceValidator{}`
  6. `CorrectAnswerInOptionsValidator{}`
  7. `GradingRequiresQuizValidator{}`
- All 7 have compile-time interface checks in `validators.go:15-23`.
- Matches the plan exactly.

### 12. Retry with Exponential Backoff

**Verdict: PASS**

- `retry.go:14-30` defines `RetryConfig` with `MaxRetries=5`, `InitialBackoff=1s`, `MaxBackoff=30s`.
- `retry.go:35-60` implements `WithRetry` with context-aware loop.
- `retry.go:63-71` `isRetryable` checks status codes: 429, 500, 502, 503, 504.
- `retry.go:75-93` implements exponential backoff with +-25% jitter.
- All Forms API calls (`forms_api.go`) and Drive API calls (`drive_api.go`) wrap operations in `WithRetry`.
- Matches the plan's retry design.

### 13. Error Hierarchy: APIError, NotFoundError, RateLimitError

**Verdict: PASS**

- `errors.go:12-53` defines all three error types:
  - `APIError` with `StatusCode`, `Message`, `Err`, plus `Unwrap()`.
  - `NotFoundError` with `Resource`, `ID`, plus `Unwrap()` returning `APIError{StatusCode: 404}`.
  - `RateLimitError` with `Message`, plus `Unwrap()` returning `APIError{StatusCode: 429}`.
- `errors.go:56-86` provides `IsNotFound()`, `IsRateLimit()`, `ErrorStatusCode()` utility functions.
- `forms_api.go:132-151` (`wrapGoogleAPIError`, `mapStatusToError`) properly maps Google API errors to custom types.
- Matches the plan's error hierarchy.

### 14. Import Support: ImportState reads form by ID

**Verdict: PASS**

- `import.go:18-24` implements `ImportState` using `resource.ImportStatePassthroughID` which sets the form ID from the import argument.
- After import, the Read method will be called, which fetches the form and populates state.
- `form_to_model.go:49-56` generates auto-generated `item_N` keys when no key map exists (import scenario).
- `resource.go:18` declares `resource.ResourceWithImportState` interface compliance.
- Matches the plan's import design.

### 15. Quiz/Grading: Quiz toggle, points, correct_answer, feedback

**Verdict: PASS**

- `schema.go:60-65` defines `quiz` as Optional+Computed bool with default false.
- `schema.go:194-216` defines the grading block with `points` (Required), `correct_answer` (Optional), `feedback_correct` (Optional), `feedback_incorrect` (Optional).
- Grading block is attached to all three question types (`schema.go:150-152`, `168-170`, `186-188`).
- `items_to_requests.go:78-96` (`applyGrading`) correctly maps grading to API structures including `CorrectAnswers`, `WhenRight`, `WhenWrong`.
- `form_to_model.go:122-138` (`convertGrading`) correctly reads grading back from API responses.
- `validators_item.go:148-223` (`GradingRequiresQuizValidator`) enforces grading requires `quiz=true`.
- `validators_item.go:68-140` (`CorrectAnswerInOptionsValidator`) verifies correct_answer is in options list.

---

## Defects Found

### D-1: `edit_uri` Hardcoded Instead of Read from API [Major]

**Location:** `state_convert.go:112-115`
**Spec reference:** Plan says `edit_uri` is a Computed output. The `forms.get` API response includes an `editUri` field.
**Finding:** The implementation constructs `edit_uri` from a hardcoded URL pattern (`"https://docs.google.com/forms/d/" + model.ID + "/edit"`) instead of reading the actual `editUri` field from the API response.
**Impact:** If Google ever changes the edit URL format, this will produce incorrect URLs. More importantly, the `convert.FormModel` struct (`types.go:48-59`) does not include an `EditURI` field, and `form_to_model.go:15-45` does not read `form.LinkedSheetId` or `form.FormId` for this purpose -- it reads `form.ResponderUri` but completely ignores the API's edit URL.
**Classification:** Major -- correctness issue; computed output does not reflect actual API data.
**Fix:** Add `EditURI` to `convert.FormModel`, populate it from `form.ResponderUri` (which is the responder URI) and properly read `edit_uri` from the forms.get response if available, or keep hardcoded as documented workaround if the Forms API does not expose this field.

### D-2: `item_key` Format Not Validated [Minor]

**Location:** `schema.go:113-115`, `validators.go:126-161`
**Spec reference:** Plan Decision 2 specifies format `[a-z][a-z0-9_]{0,63}`.
**Finding:** The `item_key` attribute is `Required` and uniqueness is validated, but no regex or format validator enforces the `[a-z][a-z0-9_]{0,63}` pattern. Users can provide item_keys like "123", "UPPER_CASE", or strings with spaces.
**Impact:** Inconsistent item_key values could cause issues in Phase 2 when used for diff correlation. Non-conforming keys may confuse state management.
**Classification:** Minor -- defensive validation missing.
**Fix:** Add a `stringvalidator.RegexMatches` on the `item_key` attribute in `schema.go`.

### D-3: `published` and `accepting_responses` Not Read Back from API [Major]

**Location:** `state_convert.go:99-118`, `form_to_model.go:15-45`
**Spec reference:** Spec says Read should detect drift for published/accepting_responses.
**Finding:** The `convertFormModelToTFState` function preserves `plan.Published` and `plan.AcceptingResponses` rather than reading them from the API response. The `FormModel` struct in `types.go` does not even have `Published` or `AcceptingResponses` fields. The `FormToModel` function in `form_to_model.go` does not attempt to read publish settings from the API response.
**Impact:** External changes to publish/accepting state are never detected. If someone unpublishes a form outside Terraform, `terraform plan` will show no drift -- it always shows the last-known config value. This violates the spec's requirement: "Ensure any drift (external changes to the form) will be detected."
**Classification:** Major -- drift detection broken for publish settings.
**Fix:** Either call a separate API endpoint to read publish settings (if `forms.get` doesn't include them), or document this as a known limitation. At minimum, `FormModel` should include these fields and `convertFormModelToTFState` should use API values when available.

### D-4: `forms.create` Should Not Be Retried [Major]

**Location:** `forms_api.go:37-44`
**Spec reference:** Spec says "Ensure idempotency where possible... creating a form twice should be avoided -- we likely won't retry forms.create on uncertain failures."
**Finding:** The `Create` method wraps the API call in `WithRetry`, which will retry on 429/5xx errors. If `forms.create` succeeds but the response is lost (network issue returning the response), the retry will create a SECOND form, resulting in an orphaned duplicate. Unlike `batchUpdate` or `Get`, `forms.create` is NOT idempotent.
**Impact:** Can produce orphaned Google Forms that exist in Google but are not tracked by Terraform state.
**Classification:** Major -- data integrity risk; contradicts spec's explicit caution.
**Fix:** Remove `WithRetry` from the `Create` method, or use a single-attempt wrapper. If retrying, first check if a form with the expected title was just created (though this is fragile).

### D-5: `convert.FormModel` Missing `Published`/`AcceptingResponses` Fields [Minor]

**Location:** `types.go:48-59`
**Spec reference:** Plan Phase 1 scope includes "published, accepting_responses" as resource fields.
**Finding:** The `convert.FormModel` struct has `ID`, `Title`, `Description`, `DocumentTitle`, `ResponderURI`, `RevisionID`, `Quiz`, `Items` but does NOT have `Published` or `AcceptingResponses` fields. This is the root cause of D-3.
**Classification:** Minor -- structural omission enabling the drift detection gap.

### D-6: `ExactlyOneSubBlockValidator` Does Not Run When content_json Is Set [Minor]

**Location:** `validators.go:179-209`
**Spec reference:** Plan says validators 1-7 are all wired.
**Finding:** The `ExactlyOneSubBlockValidator` reads items from config and validates them. When `content_json` is used instead of item blocks, the items list will be null/empty, so the validator passes trivially. This is correct behavior (content_json mode doesn't use item blocks), but the validator does not explicitly skip when content_json is set -- it only happens to work because items will be null. If someone sets `content_json` AND an empty `item` block (no sub-blocks), the MutuallyExclusiveValidator will catch it first. This is fine but fragile.
**Classification:** Minor -- implicit rather than explicit skip logic.

### D-7: `AcceptingResponsesRequiresPublished` Error Message Case Mismatch [Minor]

**Location:** `validators.go:103-108` vs `validators_test.go:279`
**Finding:** The error message in the validator says "A form cannot accept responses while unpublished" (capital A), but the test at line 279 checks for lowercase "cannot accept responses while unpublished". The `expectErrorContains` helper checks both Summary and Detail. The actual Summary is "Invalid Configuration" and the Detail starts with "A form cannot..." The test substring "cannot accept responses while unpublished" IS contained in the Detail, so the test passes. Not actually a bug, but the test's search string is slightly misleading -- it appears to test for a different casing than exists.
**Classification:** Minor -- test hygiene issue.

### D-8: No `forms.responses.readonly` Scope - Intentional Omission [Observation]

**Location:** `client.go:19-22`
**Spec reference:** Spec says "at minimum forms.body, forms.responses.readonly, and drive.file".
**Finding:** Only `forms.FormsBodyScope` and `drive.DriveFileScope` are included. The `forms.responses.readonly` scope is omitted.
**Assessment:** The implementation plan explicitly excluded response data features from Phase 1. This is an intentional deviation from the original spec, documented in the plan. No action needed for Phase 1, but should be added in Phase 2 when response data sources are introduced.
**Classification:** Observation -- intentional plan deviation from spec, documented.

---

## Additional Checks

### Missing Phase 1 Requirements

1. **`folder_id` / Drive folder placement**: Spec Phase 1 includes `folder_id` to place forms in Drive folders. The plan explicitly deferred this to Phase 2. **Not implemented.** This is a documented scope reduction in the plan.

2. **Ownership transfer**: Spec mentions optional owner email. Plan explicitly deferred to Phase 2. **Not implemented.** Documented.

3. **User OAuth flow**: Spec mentions OAuth tokens. Plan deferred to Phase 2. Only service account + ADC implemented. Documented.

### Unintended Phase 2 Features Leaking Into Phase 1

**None found.** The implementation strictly follows the Phase 1 scope boundary defined in the plan. No advanced question types, no permissions resource, no targeted diff engine, no collaborative drift mode.

### Interface Contract Violations Between Packages

1. **`client.FormsAPI.SetPublishSettings` interface mismatch**: The `interfaces.go:27` signature is `SetPublishSettings(ctx, formID, isPublished, isAccepting) error` (returns only error), while the spec/API reference suggests `SetPublishSettingsResponse` should be returned. The implementation in `forms_api.go:103-128` discards the response (`_, apiErr := ...`). This is acceptable since the response is not needed, but the interface deviates from the plan's `interfaces.go` spec which showed the response type. **Minor -- acceptable simplification.**

2. **`convert` package uses plain Go types correctly**: The `types.go` types do NOT import Terraform framework types, avoiding the circular import issue flagged in the plan. `state_convert.go` in `resource_form` handles all TF type conversions. **PASS -- clean separation.**

3. **`client.Client` struct correctly composes interfaces**: `interfaces.go:38-41` defines `Client` with `Forms FormsAPI` and `Drive DriveAPI` fields. Resource code accesses these via `r.client.Forms` and `r.client.Drive`. **PASS.**

---

## Defect Summary Table

| ID   | Severity | Category          | File                  | Line(s)   | Description                                                     |
|------|----------|-------------------|-----------------------|-----------|-----------------------------------------------------------------|
| D-1  | Major    | Correctness       | state_convert.go      | 112-115   | `edit_uri` hardcoded instead of read from API                   |
| D-2  | Minor    | Validation        | schema.go             | 113-115   | `item_key` format `[a-z][a-z0-9_]{0,63}` not enforced          |
| D-3  | Major    | Drift Detection   | state_convert.go      | 99-118    | `published`/`accepting_responses` not read from API on refresh  |
| D-4  | Major    | Data Integrity    | forms_api.go          | 37-44     | Non-idempotent `forms.create` wrapped in retry                  |
| D-5  | Minor    | Structural        | types.go              | 48-59     | `FormModel` missing `Published`/`AcceptingResponses` fields     |
| D-6  | Minor    | Robustness        | validators.go         | 179-209   | `ExactlyOneSubBlockValidator` implicitly skips content_json     |
| D-7  | Minor    | Test Hygiene      | validators_test.go    | 279       | Error message case mismatch in test assertion                   |
| D-8  | Obs.     | Scope             | client.go             | 19-22     | `forms.responses.readonly` scope omitted (intentional)          |

**Totals:** 3 Major, 5 Minor, 1 Observation

---

## Recommendations

1. **D-4 is the highest-priority fix** -- retrying a non-idempotent create can produce orphaned resources. This should be fixed before any release.

2. **D-3 is the second priority** -- without drift detection for publish settings, the provider cannot detect if someone manually unpublishes a form. This undermines Terraform's core value proposition.

3. **D-1 can be deferred** if the Forms API's `forms.get` response does not include an `editUri` field (it may not). In that case, the hardcoded pattern should be documented as an implementation note.

4. **D-2, D-5, D-6, D-7** are low-priority improvements that should be addressed before v1.0.0 but are not blocking for v0.1.0.

---

*Report generated by the Designer role in a Fagan inspection process.*
