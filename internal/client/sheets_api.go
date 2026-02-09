// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/api/googleapi"

	sheets "google.golang.org/api/sheets/v4"
)

// SheetsAPIClient is the real implementation of SheetsAPI using Google's API.
type SheetsAPIClient struct {
	service *sheets.Service
	retry   RetryConfig
}

// NewSheetsAPIClient creates a new SheetsAPIClient.
func NewSheetsAPIClient(service *sheets.Service, retry RetryConfig) *SheetsAPIClient {
	return &SheetsAPIClient{service: service, retry: retry}
}

var _ SheetsAPI = &SheetsAPIClient{}

// Create creates a new spreadsheet via the Google Sheets API.
// Create is non-idempotent, so it must NOT be retried.
func (c *SheetsAPIClient) Create(
	ctx context.Context,
	s *sheets.Spreadsheet,
) (*sheets.Spreadsheet, error) {
	result, err := c.service.Spreadsheets.Create(s).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("sheets.Create: %w", wrapSheetsAPIError(err, "create spreadsheet"))
	}

	return result, nil
}

// Get retrieves a spreadsheet by ID via the Google Sheets API.
func (c *SheetsAPIClient) Get(
	ctx context.Context,
	spreadsheetID string,
) (*sheets.Spreadsheet, error) {
	var result *sheets.Spreadsheet

	err := WithRetry(ctx, c.retry, func() error {
		resp, apiErr := c.service.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
		if apiErr != nil {
			return wrapSheetsAPIError(apiErr, "get spreadsheet "+spreadsheetID)
		}
		result = resp
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("sheets.Get: %w", err)
	}

	return result, nil
}

// BatchUpdate applies a batch update to a spreadsheet.
func (c *SheetsAPIClient) BatchUpdate(
	ctx context.Context,
	spreadsheetID string,
	req *sheets.BatchUpdateSpreadsheetRequest,
) (*sheets.BatchUpdateSpreadsheetResponse, error) {
	var result *sheets.BatchUpdateSpreadsheetResponse

	err := WithRetry(ctx, c.retry, func() error {
		resp, apiErr := c.service.Spreadsheets.BatchUpdate(spreadsheetID, req).Context(ctx).Do()
		if apiErr != nil {
			return wrapSheetsAPIError(apiErr, "batch update spreadsheet "+spreadsheetID)
		}
		result = resp
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("sheets.BatchUpdate: %w", err)
	}

	return result, nil
}

// wrapSheetsAPIError converts a googleapi.Error from Sheets API into the
// appropriate custom error type.
func wrapSheetsAPIError(err error, operation string) error {
	var gErr *googleapi.Error
	if !errors.As(err, &gErr) {
		return fmt.Errorf("%s: %w", operation, err)
	}

	return mapStatusToError(gErr.Code, gErr.Message, operation, "spreadsheet")
}

// mapSheetsStatusToError is an alias for the shared mapStatusToError for clarity.
func mapSheetsStatusToError(code int, message, operation string) error {
	switch {
	case code == http.StatusNotFound:
		return &NotFoundError{Resource: "spreadsheet", ID: operation}
	case code == http.StatusTooManyRequests:
		return &RateLimitError{Message: message}
	default:
		return &APIError{StatusCode: code, Message: message}
	}
}
