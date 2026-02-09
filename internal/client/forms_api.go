// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/api/googleapi"

	forms "google.golang.org/api/forms/v1"
)

// FormsAPIClient is the real implementation of FormsAPI using Google's API.
type FormsAPIClient struct {
	service *forms.Service
	retry   RetryConfig
}

// NewFormsAPIClient creates a new FormsAPIClient.
func NewFormsAPIClient(service *forms.Service, retry RetryConfig) *FormsAPIClient {
	return &FormsAPIClient{service: service, retry: retry}
}

var _ FormsAPI = &FormsAPIClient{}

// Create creates a new form via the Google Forms API.
// Create is non-idempotent, so it must NOT be retried.
func (c *FormsAPIClient) Create(
	ctx context.Context,
	form *forms.Form,
) (*forms.Form, error) {
	result, err := c.service.Forms.Create(form).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("forms.Create: %w", wrapGoogleAPIError(err, "create form"))
	}

	return result, nil
}

// Get retrieves a form by ID via the Google Forms API.
func (c *FormsAPIClient) Get(
	ctx context.Context,
	formID string,
) (*forms.Form, error) {
	var result *forms.Form

	err := WithRetry(ctx, c.retry, func() error {
		resp, apiErr := c.service.Forms.Get(formID).Context(ctx).Do()
		if apiErr != nil {
			return wrapGoogleAPIError(apiErr, "get form "+formID)
		}
		result = resp
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("forms.Get: %w", err)
	}

	return result, nil
}

// BatchUpdate applies a batch update to a form via the Google Forms API.
func (c *FormsAPIClient) BatchUpdate(
	ctx context.Context,
	formID string,
	req *forms.BatchUpdateFormRequest,
) (*forms.BatchUpdateFormResponse, error) {
	req.IncludeFormInResponse = true

	var result *forms.BatchUpdateFormResponse

	err := WithRetry(ctx, c.retry, func() error {
		resp, apiErr := c.service.Forms.BatchUpdate(formID, req).Context(ctx).Do()
		if apiErr != nil {
			return wrapGoogleAPIError(apiErr, "batch update form "+formID)
		}
		result = resp
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("forms.BatchUpdate: %w", err)
	}

	return result, nil
}

// SetPublishSettings updates the publish settings for a form.
func (c *FormsAPIClient) SetPublishSettings(
	ctx context.Context,
	formID string,
	isPublished bool,
	isAccepting bool,
) error {
	err := WithRetry(ctx, c.retry, func() error {
		pubReq := &forms.SetPublishSettingsRequest{
			PublishSettings: &forms.PublishSettings{
				PublishState: &forms.PublishState{
					IsPublished:          isPublished,
					IsAcceptingResponses: isAccepting,
					ForceSendFields:      []string{"IsPublished", "IsAcceptingResponses"},
				},
			},
		}
		_, apiErr := c.service.Forms.SetPublishSettings(formID, pubReq).Context(ctx).Do()
		if apiErr != nil {
			return wrapGoogleAPIError(apiErr, "set publish settings for "+formID)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("forms.SetPublishSettings: %w", err)
	}

	return nil
}

// wrapGoogleAPIError converts a googleapi.Error into the appropriate
// custom error type based on HTTP status code.
func wrapGoogleAPIError(err error, operation string) error {
	var gErr *googleapi.Error
	if !errors.As(err, &gErr) {
		return fmt.Errorf("%s: %w", operation, err)
	}

	return mapStatusToError(gErr.Code, gErr.Message, operation, "form")
}

// mapStatusToError creates the appropriate error type for an HTTP status code.
// The resource parameter identifies the API resource type (e.g. "form", "file").
func mapStatusToError(code int, message, operation, resource string) error {
	switch {
	case code == http.StatusNotFound:
		return &NotFoundError{Resource: resource, ID: operation}
	case code == http.StatusTooManyRequests:
		return &RateLimitError{Message: message}
	default:
		return &APIError{StatusCode: code, Message: message}
	}
}
