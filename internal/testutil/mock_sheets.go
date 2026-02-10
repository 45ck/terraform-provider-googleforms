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

	ValuesGetFunc    func(ctx context.Context, spreadsheetID, rng string) (*sheets.ValueRange, error)
	ValuesUpdateFunc func(ctx context.Context, spreadsheetID, rng string, vr *sheets.ValueRange, valueInputOption string) (*sheets.UpdateValuesResponse, error)
	ValuesClearFunc  func(ctx context.Context, spreadsheetID, rng string) error
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

func (m *MockSheetsAPI) ValuesGet(ctx context.Context, spreadsheetID, rng string) (*sheets.ValueRange, error) {
	if m.ValuesGetFunc != nil {
		return m.ValuesGetFunc(ctx, spreadsheetID, rng)
	}
	return &sheets.ValueRange{
		Range:  rng,
		Values: [][]interface{}{},
	}, nil
}

func (m *MockSheetsAPI) ValuesUpdate(
	ctx context.Context,
	spreadsheetID string,
	rng string,
	vr *sheets.ValueRange,
	valueInputOption string,
) (*sheets.UpdateValuesResponse, error) {
	if m.ValuesUpdateFunc != nil {
		return m.ValuesUpdateFunc(ctx, spreadsheetID, rng, vr, valueInputOption)
	}
	return &sheets.UpdateValuesResponse{UpdatedRange: rng}, nil
}

func (m *MockSheetsAPI) ValuesClear(ctx context.Context, spreadsheetID, rng string) error {
	if m.ValuesClearFunc != nil {
		return m.ValuesClearFunc(ctx, spreadsheetID, rng)
	}
	return nil
}
