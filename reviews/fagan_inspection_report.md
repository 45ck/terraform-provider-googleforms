# Fagan Inspection Report — Final Defect Log

## Inspection Metadata

| Field | Value |
|-------|-------|
| **Date** | 2026-02-09 |
| **Moderator** | Team Lead (Claude Opus 4.6) |
| **Materials** | 38 Go source files, 70 total project files |
| **Scope** | All production and test code in `internal/`, `main.go`, `go.mod` |
| **Reference Spec** | `terraform-google-forms-spec.md`, implementation plan |

### Inspection Participants

| Role | Agent | Report |
|------|-------|--------|
| **Reader** | reader | `fagan_reader_report.md` — systematic line-by-line walkthrough |
| **Designer** | designer | `fagan_designer_report.md` — spec compliance verification |
| **Tester** | tester | `fagan_tester_report.md` — test coverage and quality |
| **Standards** | standards-checker | `fagan_standards_report.md` — coding standards and consistency |

---

## Defect Log

### Legend
- **Disposition**: FIX (must fix), DEFER (document for future), REJECT (not a real defect)
- **Severity**: Major / Minor / Style
- **Found by**: R=Reader, D=Designer, T=Tester, S=Standards

---

### MAJOR DEFECTS

| ID | Severity | Found By | File(s) | Description | Disposition |
|----|----------|----------|---------|-------------|-------------|
| F-001 | **Major** | D, R | `forms_api.go:37-44` | **Non-idempotent `forms.Create` wrapped in `WithRetry`.** If the first call succeeds but the response is lost (network timeout on return), the retry creates a DUPLICATE form, producing an orphaned resource. The spec explicitly warns: "creating a form twice should be avoided." | **FIX** |
| F-002 | **Major** | R, D | `state_convert.go:100-106` | **`convertFormModelToTFState` copies Title, Description, Quiz from plan/state instead of API model.** The `convert.FormModel` has these fields populated from the API, but the state converter ignores them, using the old state values instead. External changes to title, description, or quiz mode are never detected as drift. | **FIX** |
| F-003 | **Major** | S | `provider.go` + `client.go` | **Duplicate credential resolution.** Both `provider.resolveCredentials()` and `client.resolveCredentials()` independently check `GOOGLE_CREDENTIALS` env var. If provider passes empty string (ADC mode), client re-checks the env var, potentially using credentials the provider explicitly skipped. | **FIX** |
| F-004 | **Major** | R | `crud_update.go:48-61` | **Batch ordering when disabling quiz mode.** In the batchUpdate request list, quiz settings change (disable) comes BEFORE item deletes. If existing items have grading, disabling quiz mode may fail because graded items still exist. Deletes should come first. | **FIX** |
| F-005 | **Major** | T | `crud_test.go` | **Missing tests for 4 critical error paths:** (1) Create with invalid content_json, (2) Update BatchUpdate failure, (3) Update content_json mode, (4) Create SetPublishSettings failure. These are code paths that handle real API failures but have zero test coverage. | **FIX** |
| F-006 | **Major** | T | `items_to_requests_test.go` | **No-question-block error path untested.** `ItemModelToCreateRequest` returns an error when all sub-blocks are nil. No test exercises this. A buggy implementation that silently succeeds would not be caught. | **FIX** |
| F-007 | **Major** | T | `validators_item_test.go` | **`GradingRequiresQuizValidator` only tested with short_answer.** Production code checks all 3 item types. Bugs in `multiple_choice` or `paragraph` grading detection would go undetected. | **FIX** |
| F-008 | **Major** | T | `retry_test.go:282` | **Flaky `TestRetry_BackoffIncreases`.** Compares `gap3 > gap1` but 25% jitter can cause non-deterministic failures. With `InitialBackoff=50ms`, gap1 range is [37.5ms, 62.5ms] and gap3 range is [150ms, 250ms]. The assertion usually passes but can fail. | **FIX** |

### REJECTED MAJOR CLAIMS

| ID | Claimed By | Claim | Rejection Rationale |
|----|------------|-------|---------------------|
| REJ-001 | Reader (JSON-1/PM-1) | `NormalizeJSON` does not produce deterministic output because Go's `json.Marshal` does not sort map keys. | **WRONG.** Since Go 1.12 (2019), `encoding/json.Marshal` explicitly sorts map keys for `map[string]interface{}`. This project requires Go 1.22. `NormalizeJSON` IS deterministic. |
| REJ-002 | Standards (STD-035/036) | Missing `go.sum` and indirect dependencies is a "Major" defect. | **EXPECTED.** Go is not installed on this machine; the project has never been compiled. Running `go mod tidy` will generate both. This is an environment limitation, not a code defect. Downgraded to NOTE. |
| REJ-003 | Standards (STD-024) | `item_key` format not validated is "Major". | **Downgraded to Minor.** The key is internal-only (never sent to Google API). Non-conforming keys work correctly. Format validation is defensive, not critical. |

---

### MINOR DEFECTS

| ID | Severity | Found By | File(s) | Description | Disposition |
|----|----------|----------|---------|-------------|-------------|
| F-009 | Minor | R, D | `forms_api.go:145`, `drive_api.go:57` | `mapStatusToError` hardcodes `Resource: "form"` for all 404s, including Drive operations. Drive 404 should say "file". | **FIX** |
| F-010 | Minor | R | `validators.go:198-208` | `ExactlyOneSubBlockValidator` error message does not identify WHICH item has the problem. Should include `item_key` or index. | **FIX** |
| F-011 | Minor | T | `validators_test.go` | `MutuallyExclusiveValidator` not tested with empty string `content_json`. Empty string bypasses null/unknown check but may not be a valid config. | **FIX** |
| F-012 | Minor | R | `provider.go:135-157` | `resolveCredentialValue`: invalid file path silently falls through as raw credential JSON, producing confusing downstream errors. Should add a warning or early error. | **DEFER** |
| F-013 | Minor | D | `state_convert.go:99-118` | `Published`/`AcceptingResponses` always from plan. The Forms API `forms.get` does NOT return publish settings (separate endpoint). This is an API limitation, not a code bug. | **DEFER** (document as known limitation) |
| F-014 | Minor | D | `state_convert.go:112-115` | `edit_uri` hardcoded from form ID. The Forms API does not return an edit URI in `forms.get`. Acceptable workaround. | **DEFER** |
| F-015 | Minor | D | `schema.go:113-115` | `item_key` format `[a-z][a-z0-9_]{0,63}` documented but not enforced. | **DEFER** |
| F-016 | Minor | R | `errors.go:37-39, 51-53` | `NotFoundError.Unwrap()` creates orphaned `APIError`. Error chain terminates at synthetic 404/429 error. | **DEFER** |
| F-017 | Minor | R | `forms_api.go:82` | `BatchUpdate` mutates caller's request by setting `IncludeFormInResponse`. | **DEFER** |
| F-018 | Minor | R | `crud_read.go:80-84` | `content_json` drift completely invisible. State never compared to API. | **DEFER** (documented limitation) |
| F-019 | Minor | R | `crud_create.go:176-191` | Positional key map assumes API returns items in creation order. | **DEFER** (documented with ASSUMPTION comment) |
| F-020 | Minor | T | (multiple) | Missing tests for: `state_convert.go` (no direct tests), `import.go` (no tests), `client.go` (no tests), `mapStatusToError` (no tests). | **DEFER** (tracked for next sprint) |
| F-021 | Minor | S | `state_convert.go` | Inconsistent convert function naming: `tfItemsToConvertItems` vs `convertFormModelToTFState`. | **DEFER** |
| F-022 | Minor | S | `sweeper.go` | Sweeper is TODO stub with no code. Should be implemented or removed. | **DEFER** |
| F-023 | Minor | S | `crud_test.go` | At ~1226 lines, exceeds project's own 400-line file limit. | **DEFER** |

---

### STYLE DEFECTS

| ID | Severity | Found By | File(s) | Description | Disposition |
|----|----------|----------|---------|-------------|-------------|
| F-024 | Style | S | `crud_create.go`, `crud_read.go`, `crud_update.go`, `crud_delete.go` | Missing godoc comments on exported CRUD methods (interface implementations). | **FIX** |
| F-025 | Style | S | `plan_modifiers_test.go:183` | Dummy `var _ attr.Value = types.StringNull()` to suppress unused import. | **FIX** |
| F-026 | Style | S | `retry_test.go:314` | Dummy `_ = gap2` assignment to suppress unused variable. | **FIX** |
| F-027 | Style | S | `provider.go` | Missing package-level doc comment for `provider` package. | **FIX** |
| F-028 | Style | R | `state_convert.go:38-81` | Sequential `if` instead of `switch` for mutually exclusive sub-blocks. | **DEFER** |

---

## Defect Statistics

| Severity | Total | FIX | DEFER | REJECT |
|----------|-------|-----|-------|--------|
| Major | 8 (+3 rejected) | 8 | 0 | 3 |
| Minor | 15 | 3 | 12 | 0 |
| Style | 5 | 4 | 1 | 0 |
| **Total** | **28** | **15** | **13** | **3** |

---

## FIX List (Rework Items)

### Priority 1 — Correctness (must fix)

1. **F-001**: Remove `WithRetry` from `FormsAPIClient.Create` in `forms_api.go`
2. **F-002**: In `convertFormModelToTFState`, read `Title`, `Description`, `Quiz` from `model` (API) instead of `plan`
3. **F-003**: Remove env-var credential check from `client.resolveCredentials`; trust the string passed by provider
4. **F-004**: Reorder batch requests in `crud_update.go` — put item deletes BEFORE settings changes

### Priority 2 — Test gaps (must fix)

5. **F-005**: Add 4 missing CRUD error path tests
6. **F-006**: Add no-question-block error test for `ItemModelToCreateRequest`
7. **F-007**: Add grading validation tests for `multiple_choice` and `paragraph`
8. **F-008**: Fix flaky backoff test (use wider margin or deterministic approach)
9. **F-011**: Add empty-string `content_json` test for `MutuallyExclusiveValidator`

### Priority 3 — Polish (should fix)

10. **F-009**: Add `resource` parameter to `mapStatusToError` for correct 404 resource label
11. **F-010**: Include item_key/index in `ExactlyOneSubBlockValidator` error message
12. **F-024**: Add godoc to CRUD methods
13. **F-025**: Remove dummy attr var in plan_modifiers_test
14. **F-026**: Either assert on gap2 or remove it in retry_test
15. **F-027**: Add package doc to provider package

---

## Re-Inspection Decision

**Major defect count: 8** (threshold for re-inspection: typically >3 per KLOC)

Given the codebase is ~3000 lines of production code, 8 Major defects is 2.7 per KLOC — borderline. However:
- F-001 (Create retry) is a genuine data integrity risk
- F-002 (drift detection) undermines core Terraform functionality
- F-003 (duplicate credentials) is a latent bug
- F-004 (batch ordering) can cause runtime failures
- F-005 through F-008 are test gaps, not production bugs

**Recommendation: Fix all FIX items, then perform a TARGETED re-inspection of the changed files only (not full re-inspection).**

---

*Moderator: Team Lead (Claude Opus 4.6)*
*Date: 2026-02-09*
