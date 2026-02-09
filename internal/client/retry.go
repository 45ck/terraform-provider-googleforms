// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// RetryConfig controls exponential backoff retry behavior.
type RetryConfig struct {
	// MaxRetries is the maximum number of retries after the initial attempt.
	MaxRetries int
	// InitialBackoff is the base delay before the first retry.
	InitialBackoff time.Duration
	// MaxBackoff is the maximum delay between retries.
	MaxBackoff time.Duration
}

// DefaultRetryConfig returns a RetryConfig with sensible defaults.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     5,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
	}
}

// WithRetry executes fn with exponential backoff retry for transient errors.
// It retries on 429, 500, 502, 503, 504 status codes.
// It does not retry on 400, 401, 403, 404, or non-API errors.
func WithRetry(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return wrapContextError(err, lastErr)
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if !isRetryable(lastErr) {
			return lastErr
		}

		if attempt < cfg.MaxRetries {
			if err := sleepWithContext(ctx, backoffDuration(cfg, attempt)); err != nil {
				return wrapContextError(err, lastErr)
			}
		}
	}

	return lastErr
}

// isRetryable determines if an error should be retried.
func isRetryable(err error) bool {
	code := ErrorStatusCode(err)
	switch code {
	case 429, 500, 502, 503, 504:
		return true
	default:
		return false
	}
}

// backoffDuration calculates the backoff duration for a given attempt
// with exponential increase and +-25% jitter.
func backoffDuration(cfg RetryConfig, attempt int) time.Duration {
	backoff := cfg.InitialBackoff
	for i := 0; i < attempt; i++ {
		backoff *= 2
		if backoff > cfg.MaxBackoff {
			backoff = cfg.MaxBackoff
			break
		}
	}

	return addJitter(backoff)
}

// addJitter adds +-25% random jitter to a duration.
func addJitter(d time.Duration) time.Duration {
	//nolint:gosec // jitter does not need cryptographic randomness
	jitter := 0.75 + rand.Float64()*0.5 // range [0.75, 1.25]
	return time.Duration(float64(d) * jitter)
}

// sleepWithContext sleeps for the given duration or until the context is cancelled.
func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// wrapContextError returns the context error, preserving the last API error
// for diagnostic purposes when a retry loop is interrupted by cancellation.
func wrapContextError(ctxErr error, lastErr error) error {
	if lastErr != nil {
		return fmt.Errorf("%w (last API error: %v)", ctxErr, lastErr)
	}
	return ctxErr
}
