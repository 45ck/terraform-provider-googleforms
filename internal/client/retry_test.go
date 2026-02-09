// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetry_SuccessOnFirstAttempt(t *testing.T) {
	t.Parallel()

	var attempts int32

	err := WithRetry(context.Background(), DefaultRetryConfig(), func() error {
		atomic.AddInt32(&attempts, 1)
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Errorf("expected 1 attempt, got %d", got)
	}
}

func TestRetry_SuccessAfterRetries(t *testing.T) {
	t.Parallel()

	var attempts int32
	cfg := testRetryConfig()

	err := WithRetry(context.Background(), cfg, func() error {
		n := atomic.AddInt32(&attempts, 1)
		if n < 3 {
			return &APIError{StatusCode: 503, Message: "unavailable"}
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := atomic.LoadInt32(&attempts); got != 3 {
		t.Errorf("expected 3 attempts, got %d", got)
	}
}

func TestRetry_ExhaustsRetries(t *testing.T) {
	t.Parallel()

	var attempts int32
	cfg := testRetryConfig()
	cfg.MaxRetries = 3

	err := WithRetry(context.Background(), cfg, func() error {
		atomic.AddInt32(&attempts, 1)
		return &APIError{StatusCode: 500, Message: "server error"}
	})

	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}

	// MaxRetries=3 means 1 initial + 3 retries = 4 total attempts.
	if got := atomic.LoadInt32(&attempts); got != 4 {
		t.Errorf("expected 4 attempts, got %d", got)
	}
}

func TestRetry_NoRetryOn400(t *testing.T) {
	t.Parallel()

	var attempts int32
	cfg := testRetryConfig()

	err := WithRetry(context.Background(), cfg, func() error {
		atomic.AddInt32(&attempts, 1)
		return &APIError{StatusCode: 400, Message: "bad request"}
	})

	if err == nil {
		t.Fatal("expected error for 400")
	}

	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Errorf("expected 1 attempt (no retry on 400), got %d", got)
	}
}

func TestRetry_NoRetryOn404(t *testing.T) {
	t.Parallel()

	var attempts int32
	cfg := testRetryConfig()

	err := WithRetry(context.Background(), cfg, func() error {
		atomic.AddInt32(&attempts, 1)
		return &NotFoundError{Resource: "form", ID: "abc"}
	})

	if err == nil {
		t.Fatal("expected error for 404")
	}

	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Errorf("expected 1 attempt (no retry on 404), got %d", got)
	}

	if !IsNotFound(err) {
		t.Error("expected NotFoundError to be preserved")
	}
}

func TestRetry_NoRetryOn401(t *testing.T) {
	t.Parallel()

	var attempts int32
	cfg := testRetryConfig()

	err := WithRetry(context.Background(), cfg, func() error {
		atomic.AddInt32(&attempts, 1)
		return &APIError{StatusCode: 401, Message: "unauthorized"}
	})

	if err == nil {
		t.Fatal("expected error for 401")
	}

	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Errorf("expected 1 attempt (no retry on 401), got %d", got)
	}
}

func TestRetry_NoRetryOn403(t *testing.T) {
	t.Parallel()

	var attempts int32
	cfg := testRetryConfig()

	err := WithRetry(context.Background(), cfg, func() error {
		atomic.AddInt32(&attempts, 1)
		return &APIError{StatusCode: 403, Message: "forbidden"}
	})

	if err == nil {
		t.Fatal("expected error for 403")
	}

	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Errorf("expected 1 attempt (no retry on 403), got %d", got)
	}
}

func TestRetry_RetriesOn429(t *testing.T) {
	t.Parallel()

	var attempts int32
	cfg := testRetryConfig()

	err := WithRetry(context.Background(), cfg, func() error {
		n := atomic.AddInt32(&attempts, 1)
		if n < 2 {
			return &RateLimitError{Message: "rate limited"}
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := atomic.LoadInt32(&attempts); got != 2 {
		t.Errorf("expected 2 attempts, got %d", got)
	}
}

func TestRetry_RetriesOn503(t *testing.T) {
	t.Parallel()

	var attempts int32
	cfg := testRetryConfig()

	err := WithRetry(context.Background(), cfg, func() error {
		n := atomic.AddInt32(&attempts, 1)
		if n < 2 {
			return &APIError{StatusCode: 503, Message: "unavailable"}
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := atomic.LoadInt32(&attempts); got != 2 {
		t.Errorf("expected 2 attempts, got %d", got)
	}
}

func TestRetry_RetriesOn502(t *testing.T) {
	t.Parallel()

	var attempts int32
	cfg := testRetryConfig()

	err := WithRetry(context.Background(), cfg, func() error {
		n := atomic.AddInt32(&attempts, 1)
		if n < 2 {
			return &APIError{StatusCode: 502, Message: "bad gateway"}
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := atomic.LoadInt32(&attempts); got != 2 {
		t.Errorf("expected 2 attempts, got %d", got)
	}
}

func TestRetry_RetriesOn504(t *testing.T) {
	t.Parallel()

	var attempts int32
	cfg := testRetryConfig()

	err := WithRetry(context.Background(), cfg, func() error {
		n := atomic.AddInt32(&attempts, 1)
		if n < 2 {
			return &APIError{StatusCode: 504, Message: "gateway timeout"}
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := atomic.LoadInt32(&attempts); got != 2 {
		t.Errorf("expected 2 attempts, got %d", got)
	}
}

func TestRetry_RespectsContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	var attempts int32
	cfg := testRetryConfig()
	// Use longer backoff so cancellation triggers during wait.
	cfg.InitialBackoff = 500 * time.Millisecond

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := WithRetry(ctx, cfg, func() error {
		atomic.AddInt32(&attempts, 1)
		return &APIError{StatusCode: 503, Message: "unavailable"}
	})

	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}

	if ctx.Err() == nil {
		t.Error("expected context to be cancelled")
	}
}

func TestRetry_BackoffIncreases(t *testing.T) {
	t.Parallel()

	var timestamps []time.Time
	cfg := testRetryConfig()
	cfg.InitialBackoff = 50 * time.Millisecond
	cfg.MaxBackoff = 5 * time.Second
	cfg.MaxRetries = 4

	_ = WithRetry(context.Background(), cfg, func() error {
		timestamps = append(timestamps, time.Now())
		return &APIError{StatusCode: 500, Message: "error"}
	})

	if len(timestamps) < 4 {
		t.Fatalf("expected at least 4 timestamps, got %d", len(timestamps))
	}

	// Verify that each gap is longer than or approximately equal to the previous.
	// With jitter, we allow some tolerance. The key property is that
	// later gaps should generally be larger than earlier gaps.
	gap1 := timestamps[1].Sub(timestamps[0])
	gap2 := timestamps[2].Sub(timestamps[1])
	gap3 := timestamps[3].Sub(timestamps[2])

	// Verify backoff generally increases. With jitter, use a very safe margin:
	// gap3 should be greater than gap1/2 (well within exponential growth).
	if gap3 < gap1/2 {
		t.Errorf("expected backoff to increase: gap1=%v gap2=%v gap3=%v",
			gap1, gap2, gap3)
	}

	// Also assert gap2 is positive and within a reasonable range.
	if gap2 < gap1/2 {
		t.Errorf("expected gap2 to be at least gap1/2: gap1=%v gap2=%v", gap1, gap2)
	}
}

func TestRetry_NonAPIErrorNotRetried(t *testing.T) {
	t.Parallel()

	var attempts int32
	cfg := testRetryConfig()

	err := WithRetry(context.Background(), cfg, func() error {
		atomic.AddInt32(&attempts, 1)
		return fmt.Errorf("non-api error")
	})

	if err == nil {
		t.Fatal("expected error")
	}

	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Errorf("expected 1 attempt (non-API errors not retried), got %d", got)
	}
}

// testRetryConfig returns a RetryConfig with short durations for testing.
func testRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     5,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     50 * time.Millisecond,
	}
}
