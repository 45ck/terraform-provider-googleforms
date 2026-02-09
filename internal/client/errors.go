// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"errors"
	"fmt"
)

// APIError represents a general Google API error with an HTTP status code.
type APIError struct {
	StatusCode int
	Message    string
	Err        error
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

func (e *APIError) Unwrap() error {
	return e.Err
}

// NotFoundError represents a 404 response from the API.
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

// Unwrap returns an APIError with status 404 so errors.As works.
func (e *NotFoundError) Unwrap() error {
	return &APIError{StatusCode: 404, Message: e.Error()}
}

// RateLimitError represents a 429 response from the API.
type RateLimitError struct {
	Message string
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded: %s", e.Message)
}

// Unwrap returns an APIError with status 429 so errors.As works.
func (e *RateLimitError) Unwrap() error {
	return &APIError{StatusCode: 429, Message: e.Error()}
}

// IsNotFound reports whether err is or wraps a NotFoundError.
func IsNotFound(err error) bool {
	var target *NotFoundError
	return errors.As(err, &target)
}

// IsRateLimit reports whether err is or wraps a RateLimitError.
func IsRateLimit(err error) bool {
	var target *RateLimitError
	return errors.As(err, &target)
}

// ErrorStatusCode extracts the HTTP status code from an error.
// Returns 0 if the error does not contain a status code.
func ErrorStatusCode(err error) int {
	var nf *NotFoundError
	if errors.As(err, &nf) {
		return 404
	}

	var rl *RateLimitError
	if errors.As(err, &rl) {
		return 429
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode
	}

	return 0
}
