// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"errors"
	"fmt"
	"testing"
)

func TestAPIError_ErrorMessage(t *testing.T) {
	t.Parallel()

	err := &APIError{StatusCode: 500, Message: "internal server error"}
	expected := "API error (status 500): internal server error"

	if err.Error() != expected {
		t.Errorf("got %q, want %q", err.Error(), expected)
	}
}

func TestAPIError_Unwrap(t *testing.T) {
	t.Parallel()

	inner := fmt.Errorf("underlying cause")
	err := &APIError{StatusCode: 500, Message: "server error", Err: inner}

	if !errors.Is(err, inner) {
		t.Error("expected APIError.Unwrap to return the wrapped error")
	}
}

func TestNotFoundError_ErrorMessage(t *testing.T) {
	t.Parallel()

	err := &NotFoundError{Resource: "form", ID: "abc123"}
	msg := err.Error()

	if msg != "form not found: abc123" {
		t.Errorf("got %q, want %q", msg, "form not found: abc123")
	}
}

func TestNotFoundError_StatusCode(t *testing.T) {
	t.Parallel()

	err := &NotFoundError{Resource: "form", ID: "abc123"}
	var apiErr *APIError

	if !errors.As(err, &apiErr) {
		t.Fatal("expected NotFoundError to be unwrappable as APIError")
	}

	if apiErr.StatusCode != 404 {
		t.Errorf("got status %d, want 404", apiErr.StatusCode)
	}
}

func TestRateLimitError_ErrorMessage(t *testing.T) {
	t.Parallel()

	err := &RateLimitError{Message: "quota exceeded"}
	msg := err.Error()

	if msg != "rate limit exceeded: quota exceeded" {
		t.Errorf("got %q, want %q", msg, "rate limit exceeded: quota exceeded")
	}
}

func TestRateLimitError_StatusCode(t *testing.T) {
	t.Parallel()

	err := &RateLimitError{Message: "quota exceeded"}
	var apiErr *APIError

	if !errors.As(err, &apiErr) {
		t.Fatal("expected RateLimitError to be unwrappable as APIError")
	}

	if apiErr.StatusCode != 429 {
		t.Errorf("got status %d, want 429", apiErr.StatusCode)
	}
}

func TestIsNotFound_WithNotFoundError(t *testing.T) {
	t.Parallel()

	err := &NotFoundError{Resource: "form", ID: "abc123"}
	if !IsNotFound(err) {
		t.Error("expected IsNotFound to return true for NotFoundError")
	}
}

func TestIsNotFound_WithWrappedNotFoundError(t *testing.T) {
	t.Parallel()

	inner := &NotFoundError{Resource: "form", ID: "abc123"}
	wrapped := fmt.Errorf("operation failed: %w", inner)

	if !IsNotFound(wrapped) {
		t.Error("expected IsNotFound to return true for wrapped NotFoundError")
	}
}

func TestIsNotFound_WithOtherError(t *testing.T) {
	t.Parallel()

	err := fmt.Errorf("something went wrong")
	if IsNotFound(err) {
		t.Error("expected IsNotFound to return false for non-NotFoundError")
	}
}

func TestIsNotFound_WithNil(t *testing.T) {
	t.Parallel()

	if IsNotFound(nil) {
		t.Error("expected IsNotFound to return false for nil error")
	}
}

func TestIsRateLimit_WithRateLimitError(t *testing.T) {
	t.Parallel()

	err := &RateLimitError{Message: "quota exceeded"}
	if !IsRateLimit(err) {
		t.Error("expected IsRateLimit to return true for RateLimitError")
	}
}

func TestIsRateLimit_WithWrappedRateLimitError(t *testing.T) {
	t.Parallel()

	inner := &RateLimitError{Message: "quota exceeded"}
	wrapped := fmt.Errorf("operation failed: %w", inner)

	if !IsRateLimit(wrapped) {
		t.Error("expected IsRateLimit to return true for wrapped RateLimitError")
	}
}

func TestIsRateLimit_WithOtherError(t *testing.T) {
	t.Parallel()

	err := fmt.Errorf("something went wrong")
	if IsRateLimit(err) {
		t.Error("expected IsRateLimit to return false for non-RateLimitError")
	}
}

func TestIsRateLimit_WithNil(t *testing.T) {
	t.Parallel()

	if IsRateLimit(nil) {
		t.Error("expected IsRateLimit to return false for nil error")
	}
}

func TestErrorStatusCode_WithAPIError(t *testing.T) {
	t.Parallel()

	err := &APIError{StatusCode: 503, Message: "unavailable"}
	code := ErrorStatusCode(err)

	if code != 503 {
		t.Errorf("got %d, want 503", code)
	}
}

func TestErrorStatusCode_WithNotFoundError(t *testing.T) {
	t.Parallel()

	err := &NotFoundError{Resource: "form", ID: "abc"}
	code := ErrorStatusCode(err)

	if code != 404 {
		t.Errorf("got %d, want 404", code)
	}
}

func TestErrorStatusCode_WithRateLimitError(t *testing.T) {
	t.Parallel()

	err := &RateLimitError{Message: "slow down"}
	code := ErrorStatusCode(err)

	if code != 429 {
		t.Errorf("got %d, want 429", code)
	}
}

func TestErrorStatusCode_WithGenericError(t *testing.T) {
	t.Parallel()

	err := fmt.Errorf("generic error")
	code := ErrorStatusCode(err)

	if code != 0 {
		t.Errorf("got %d, want 0", code)
	}
}
