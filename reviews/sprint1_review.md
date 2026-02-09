# Sprint 1 Code Review

## Summary

**PASS WITH NOTES**

Sprint 1 delivers a solid client layer with clean interface abstractions, proper OAuth2 credential handling, exponential backoff retry logic, custom error types, and a well-structured Terraform provider configuration. The code is well-organized, idiomatic Go, and follows Terraform provider best practices. There is one confirmed bug (wrapContextError is a no-op), one compilation concern in the test file, and several minor issues that should be addressed.

---

## Findings

### [CRITICAL] wrapContextError is a no-op -- both branches return ctxErr, discarding lastErr

- **File**: internal/client/retry.go:108-113
- **Issue**: The `wrapContextError` function has identical branches. Both the `if lastErr != nil` path and the `else` path return `ctxErr`, meaning `lastErr` is always discarded. The intent was clearly to wrap `ctxErr` with `lastErr` context when a previous API error existed (e.g., `fmt.Errorf("context cancelled after API error: %v: %w", lastErr, ctxErr)`). As written, when a context cancellation occurs mid-retry, the caller loses all information about what API error triggered the retry in the first place. This makes debugging production issues significantly harder.
- **Fix**: Replace the function body with:
  ```go
  func wrapContextError(ctxErr error, lastErr error) error {
      if lastErr != nil {
          return fmt.Errorf("%w (last API error: %v)", ctxErr, lastErr)
      }
      return ctxErr
  }
  ```
  This preserves `ctxErr` as the primary unwrappable error (so `errors.Is(err, context.Canceled)` still works) while surfacing the last API error for diagnostics. Requires adding `"fmt"` to the imports.

---

### [WARNING] provider_test.go has a suspicious multi-value assignment that may not compile

- **File**: internal/provider/provider_test.go:96-99
- **Issue**: Lines 96-99 contain:
  ```go
  configState, err := tfsdk.Config{
      Schema: schemaResp.Schema,
      Raw:    configVal,
  }, error(nil)
  ```
  This is a comma-expression assigning a struct literal and `error(nil)` to two variables. While this is technically valid Go (a multi-value short variable declaration with two expressions), it is highly unidiomatic and confusing. It mimics the shape of a function call that returns `(T, error)`, but is actually just declaring `configState` as a `tfsdk.Config` and `err` as a typed nil `error`. The `err` variable is then checked on line 100 but can never be non-nil. This is dead code masquerading as error handling.
- **Fix**: Simplify to a direct assignment without the fake error check:
  ```go
  configState := tfsdk.Config{
      Schema: schemaResp.Schema,
      Raw:    configVal,
  }
  ```
  Remove the `if err != nil` block on lines 100-102.

---

### [WARNING] provider.go resolveCredentialValue reads arbitrary file paths without validation

- **File**: internal/provider/provider.go:135-157
- **Issue**: `resolveCredentialValue` calls `os.ReadFile(trimmed)` on any value that does not start with `{`. If a user supplies a credential value like `/etc/shadow` or `../../sensitive-file`, the provider will happily read it and pass its contents to the Google OAuth2 library. While Terraform providers typically run with the invoking user's permissions (so this is not a privilege escalation), it is still a potential information disclosure vector in shared environments or when provider config values are sourced from untrusted inputs. The official Google Cloud Terraform provider mitigates this by accepting only specific env vars and file paths with documented expectations.
- **Fix**: Consider restricting to absolute paths only, or validating that the file contains valid JSON with the expected `type` field before passing it along. At minimum, add a comment acknowledging this design decision. This is not urgent since Terraform providers inherently trust their config, but worth documenting.

---

### [WARNING] client.go resolveCredentials duplicates logic with provider.go resolveCredentials

- **File**: internal/client/client.go:75-86 and internal/provider/provider.go:115-130
- **Issue**: There are two functions named `resolveCredentials` -- one in the `client` package and one in the `provider` package. The client version checks `GOOGLE_CREDENTIALS` env var for raw JSON. The provider version also checks `GOOGLE_CREDENTIALS` env var, but additionally supports file paths via `resolveCredentialValue`. This means the env var is checked twice: once in `provider.resolveCredentials` (where it may be read as a file path) and again in `client.resolveCredentials` (where it is treated as raw JSON). Since `provider.Configure` passes the resolved JSON string to `client.NewClient`, the client-level env var check would only trigger if the provider passes an empty string. This works correctly but the double-resolution is confusing and fragile.
- **Fix**: Consider removing the `GOOGLE_CREDENTIALS` fallback from `client.resolveCredentials` (client.go:80-83) and documenting that the provider is responsible for all credential resolution. Alternatively, add a comment explaining the intentional layering.

---

### [WARNING] retry_test.go TestRetry_BackoffIncreases has a weak assertion

- **File**: internal/client/retry_test.go:282-315
- **Issue**: The test checks `if gap3 < gap1` to verify exponential backoff. With jitter of +/-25%, the expected backoffs for attempts 0-3 are roughly: 50ms, 100ms, 200ms, 400ms. Gap1 (between attempt 0 and 1) is ~50ms and gap3 (between attempt 2 and 3) is ~200ms, so the assertion `gap3 < gap1` (i.e., gap3 should not be less than gap1) is a reasonable smoke test. However, it only compares the first and third gaps, not adjacent gaps. It would be a stronger test to verify `gap2 >= gap1 * 0.5` (allowing for jitter) AND `gap3 >= gap2 * 0.5`. Also, `_ = gap2` on line 314 explicitly discards gap2 with a misleading comment saying "used for diagnostic output above" -- but gap2 is not actually used in any output.
- **Fix**: Either strengthen the assertion to compare adjacent gaps (with jitter tolerance), or at least fix the misleading comment. The current test would pass even if backoff were constant (as long as it doesn't decrease from gap1 to gap3).

---

### [WARNING] NotFoundError.Unwrap creates a new APIError on every call

- **File**: internal/client/errors.go:37-39
- **Issue**: `NotFoundError.Unwrap()` returns `&APIError{StatusCode: 404, Message: e.Error()}` -- a freshly allocated `APIError` every time. This means `errors.Is` comparisons against a specific `APIError` instance will never match, only `errors.As` type assertions will work. This is not technically a bug since the code only uses `errors.As`, but it is semantically unusual. The same applies to `RateLimitError.Unwrap()` on lines 51-53. Each call to `Unwrap` produces a new pointer, so pointer equality checks will always fail.
- **Fix**: This is acceptable given the current usage pattern (only `errors.As` is used). Add a brief comment noting that `Unwrap` returns a new instance each time, to prevent future confusion.

---

### [NOTE] forms_api.go mutates the caller's BatchUpdateFormRequest

- **File**: internal/client/forms_api.go:82
- **Issue**: `BatchUpdate` sets `req.IncludeFormInResponse = true` directly on the passed-in request object. If the caller holds a reference to this request and inspects it afterward, they will see this mutation. In practice this is fine since the Terraform resource layer creates these requests fresh each time, but it is a subtle side-effect that could surprise future callers.
- **Fix**: Add a comment: `// Always include the updated form in the response for state reconciliation.` This documents the intentional mutation. Alternatively, make a shallow copy of the request before modifying it.

---

### [NOTE] drive_api.go wrapDriveAPIError duplicates wrapGoogleAPIError

- **File**: internal/client/drive_api.go:57-64 and internal/client/forms_api.go:132-139
- **Issue**: `wrapDriveAPIError` and `wrapGoogleAPIError` are identical functions in different files. Both convert `googleapi.Error` to custom error types using `mapStatusToError`. The duplication is small but could diverge over time.
- **Fix**: Consider consolidating into a single shared function (e.g., `wrapGoogleAPIError` in a shared file), or accept the duplication with a comment noting it is intentional for independent evolution.

---

### [NOTE] retry.go uses math/rand without seeding (Go < 1.20 compatibility)

- **File**: internal/client/retry.go:8
- **Issue**: `math/rand` is used for jitter without explicit seeding. In Go 1.20+, `math/rand` is automatically seeded, so this is fine for modern Go. However, if the project ever needs to support Go < 1.20, all goroutines would get the same jitter sequence. The `//nolint:gosec` comment correctly notes that crypto-grade randomness is unnecessary.
- **Fix**: No action needed if Go >= 1.20 is the minimum version (which is almost certainly the case for a 2026 project). Just confirming this is acceptable.

---

### [NOTE] provider_test.go testFakeCredentials contains a fake RSA key with AAAA padding

- **File**: internal/provider/provider_test.go:22-33
- **Issue**: The test credentials contain a clearly fake RSA private key (mostly `AAAA...` padding). This is appropriate for testing -- it will pass basic JSON structure validation but will fail actual JWT signing. The test correctly expects a "Client Creation Failed" error. However, depending on the version of `google.JWTConfigFromJSON`, it may fail during config parsing rather than during token acquisition, which could make the test brittle if the library changes validation behavior.
- **Fix**: No immediate action required. Consider adding a comment noting that the key is intentionally invalid and the test relies on the client construction failing at some point in the OAuth2 pipeline.

---

### [NOTE] provider_test.go imports resource/providerserver packages unused in some test paths

- **File**: internal/provider/provider_test.go:13-14
- **Issue**: The `providerserver` and `resource` imports from terraform-plugin-testing are used only for the acceptance test (`TestProviderAcceptance_ConfigIsValid`). The `helper/resource` import on line 16 is used for `resource.UnitTest`. This is fine but worth noting that the acceptance test is the only one using the testing framework -- all other tests use raw SDK calls.
- **Fix**: No action needed. This is a style observation.

---

### [GOOD] Clean interface design with compile-time verification

- **File**: internal/client/interfaces.go, forms_api.go:28, drive_api.go:27, testutil/mock_forms.go:22, testutil/mock_drive.go:17
- **Issue**: The `var _ FormsAPI = &FormsAPIClient{}` pattern is used consistently to verify interface satisfaction at compile time. This is excellent Go practice and catches interface drift early.

---

### [GOOD] Credentials marked as Sensitive in provider schema

- **File**: internal/provider/provider.go:63
- **Issue**: The `credentials` attribute is correctly marked `Sensitive: true`, which prevents Terraform from displaying credential values in plan output or state. This is a critical security control.

---

### [GOOD] Proper context propagation throughout

- **File**: All API files
- **Issue**: Context is consistently threaded through all API calls, retry loops, and sleep functions. The `sleepWithContext` function correctly uses `time.NewTimer` with a `select` on `ctx.Done()` for cancellation-aware sleeping. This is textbook correct.

---

### [GOOD] Drive Delete handles 404 as idempotent success

- **File**: internal/client/drive_api.go:43-46
- **Issue**: The `Delete` method correctly treats `NotFoundError` (404) as success, implementing idempotent deletion. This is the correct pattern for Terraform providers where `terraform destroy` may be re-run after a partial failure.

---

### [GOOD] Error type hierarchy with Unwrap support

- **File**: internal/client/errors.go
- **Issue**: The error types form a clean hierarchy: `NotFoundError` and `RateLimitError` unwrap to `APIError`, which unwraps to an inner error. This supports both `errors.Is` and `errors.As` patterns. The `ErrorStatusCode` helper provides a clean way to extract status codes without type-switching.

---

### [GOOD] Mock implementations with function injection pattern

- **File**: internal/testutil/mock_forms.go, internal/testutil/mock_drive.go
- **Issue**: The mocks use the function-field pattern (e.g., `CreateFunc func(...)`) with sensible defaults when the function is nil. This is a clean, flexible pattern that avoids heavy mocking frameworks and keeps tests readable.

---

### [GOOD] Comprehensive retry test coverage

- **File**: internal/client/retry_test.go
- **Issue**: Tests cover success on first attempt, success after retries, retry exhaustion, no-retry for 400/401/403/404, retry for 429/500/502/503/504, context cancellation, backoff increase, and non-API error handling. The use of `atomic.AddInt32` for thread-safe attempt counting is correct. The `testRetryConfig` helper with millisecond-scale durations keeps tests fast.

---

### [GOOD] Comprehensive error type test coverage

- **File**: internal/client/errors_test.go
- **Issue**: Tests cover error messages, unwrap chains, `IsNotFound`/`IsRateLimit` helpers (including wrapped errors, other errors, and nil), and `ErrorStatusCode` for all error types plus the generic case. Edge cases like nil errors are properly tested.

---

## Summary Table

| Severity | Count | Action Required |
|----------|-------|-----------------|
| CRITICAL | 1     | Must fix before merge |
| WARNING  | 4     | Should fix or explicitly document |
| NOTE     | 4     | Nice to have / informational |
| GOOD     | 7     | Positive observations |

## Recommendation

Fix the CRITICAL `wrapContextError` bug, clean up the unidiomatic test assignment in `provider_test.go`, and address the WARNING items (or add explicit documentation accepting the design decisions). The codebase is otherwise well-structured and ready to support Sprint 2 resource implementation.
