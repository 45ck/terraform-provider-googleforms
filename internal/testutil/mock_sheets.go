// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"context"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
	sheets "google.golang.org/api/sheets/v4"
)

// MockSheetsAPI is a configurable mock implementation of client.SheetsAPI.
type MockSheetsAPI struct {
	CreateFunc      func(ctx context.Context, s *sheets.Spreadsheet) (*sheets.Spreadsheet, error)
	GetFunc         func(ctx context.Context, spreadsheetID string) (*sheets.Spreadsheet, error)
	BatchUpdateFunc func(ctx context.Context, spreadsheetID string, req *sheets.BatchUpdateSpreadsheetRequest) (*sheets.BatchUpdateSpreadsheetResponse, error)
}

var _ client.SheetsAPI = &MockSheetsAPI{}

func (m *MockSheetsAPI) Create(ctx context.Context, s *sheets.Spreadsheet) (*sheets.Spreadsheet, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, s)
	}
	return &sheets.Spreadsheet{
		SpreadsheetId:  "mock-spreadsheet-id",
		SpreadsheetUrl: "https://docs.google.com/spreadsheets/d/mock-spreadsheet-id/edit",
		Properties:     s.Properties,
	}, nil
}

func (m *MockSheetsAPI) Get(ctx context.Context, spreadsheetID string) (*sheets.Spreadsheet, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, spreadsheetID)
	}
	return &sheets.Spreadsheet{
		SpreadsheetId:  spreadsheetID,
		SpreadsheetUrl: "https://docs.google.com/spreadsheets/d/" + spreadsheetID + "/edit",
	}, nil
}

func (m *MockSheetsAPI) BatchUpdate(
	ctx context.Context,
	spreadsheetID string,
	req *sheets.BatchUpdateSpreadsheetRequest,
) (*sheets.BatchUpdateSpreadsheetResponse, error) {
	if m.BatchUpdateFunc != nil {
		return m.BatchUpdateFunc(ctx, spreadsheetID, req)
	}
	return &sheets.BatchUpdateSpreadsheetResponse{}, nil
}
