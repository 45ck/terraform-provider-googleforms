// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"errors"
	"fmt"

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

// ValuesGet retrieves a bounded range of values from a spreadsheet.
func (c *SheetsAPIClient) ValuesGet(
	ctx context.Context,
	spreadsheetID string,
	rng string,
) (*sheets.ValueRange, error) {
	var result *sheets.ValueRange

	err := WithRetry(ctx, c.retry, func() error {
		resp, apiErr := c.service.Spreadsheets.Values.Get(spreadsheetID, rng).Context(ctx).Do()
		if apiErr != nil {
			return wrapSheetsAPIError(apiErr, "get values "+rng+" from spreadsheet "+spreadsheetID)
		}
		result = resp
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("sheets.ValuesGet: %w", err)
	}

	return result, nil
}

// ValuesUpdate writes a bounded range of values into a spreadsheet.
func (c *SheetsAPIClient) ValuesUpdate(
	ctx context.Context,
	spreadsheetID string,
	rng string,
	vr *sheets.ValueRange,
	valueInputOption string,
) (*sheets.UpdateValuesResponse, error) {
	var result *sheets.UpdateValuesResponse

	err := WithRetry(ctx, c.retry, func() error {
		call := c.service.Spreadsheets.Values.Update(spreadsheetID, rng, vr).Context(ctx)
		if valueInputOption != "" {
			call = call.ValueInputOption(valueInputOption)
		}
		resp, apiErr := call.Do()
		if apiErr != nil {
			return wrapSheetsAPIError(apiErr, "update values "+rng+" in spreadsheet "+spreadsheetID)
		}
		result = resp
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("sheets.ValuesUpdate: %w", err)
	}

	return result, nil
}

// ValuesClear clears values in a bounded range.
func (c *SheetsAPIClient) ValuesClear(
	ctx context.Context,
	spreadsheetID string,
	rng string,
) error {
	err := WithRetry(ctx, c.retry, func() error {
		_, apiErr := c.service.Spreadsheets.Values.Clear(spreadsheetID, rng, &sheets.ClearValuesRequest{}).Context(ctx).Do()
		if apiErr != nil {
			return wrapSheetsAPIError(apiErr, "clear values "+rng+" in spreadsheet "+spreadsheetID)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("sheets.ValuesClear: %w", err)
	}

	return nil
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

// mapSheetsStatusToError was kept for backward clarity but is no longer used.
