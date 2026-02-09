// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

// Package client provides interfaces and implementations for Google API access.
// Interfaces enable mock-based testing without real API calls.
package client

import (
	"context"

	forms "google.golang.org/api/forms/v1"
)

// FormsAPI defines the interface for Google Forms API operations.
// Implementations include the real API client and mock for testing.
type FormsAPI interface {
	// Create creates a new form.
	Create(ctx context.Context, form *forms.Form) (*forms.Form, error)

	// Get retrieves the current state of a form by ID.
	Get(ctx context.Context, formID string) (*forms.Form, error)

	// BatchUpdate applies a batch of update requests to a form.
	BatchUpdate(ctx context.Context, formID string, req *forms.BatchUpdateFormRequest) (*forms.BatchUpdateFormResponse, error)

	// SetPublishSettings updates the publish state of a form.
	SetPublishSettings(ctx context.Context, formID string, isPublished bool, isAccepting bool) error
}

// DriveAPI defines the interface for Google Drive API operations on forms.
type DriveAPI interface {
	// Delete deletes a form (Drive file) by ID.
	// Returns nil if the file is already deleted (404).
	Delete(ctx context.Context, fileID string) error
}

// Client holds the API clients used by the provider.
type Client struct {
	Forms FormsAPI
	Drive DriveAPI
}
