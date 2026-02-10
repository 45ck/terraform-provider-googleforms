// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

// Package client provides interfaces and implementations for Google API access.
// Interfaces enable mock-based testing without real API calls.
package client

import (
	"context"

	drive "google.golang.org/api/drive/v3"
	forms "google.golang.org/api/forms/v1"
	sheets "google.golang.org/api/sheets/v4"
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

	// GetParents returns the current parent folder IDs for a Drive file.
	GetParents(ctx context.Context, fileID string, supportsAllDrives bool) ([]string, error)

	// MoveToFolder moves a Drive file into the given folder.
	// If folderID is empty, the file is moved to the user's root.
	MoveToFolder(ctx context.Context, fileID string, folderID string, supportsAllDrives bool) error

	// CreatePermission creates a permission on a Drive file.
	CreatePermission(ctx context.Context, fileID string, p *drive.Permission, sendNotificationEmail bool, emailMessage string, supportsAllDrives bool) (*drive.Permission, error)

	// GetPermission retrieves a permission by ID from a Drive file.
	GetPermission(ctx context.Context, fileID, permissionID string, supportsAllDrives bool) (*drive.Permission, error)

	// DeletePermission deletes a permission from a Drive file.
	// Returns nil if the permission or file is already gone (404 treated as success).
	DeletePermission(ctx context.Context, fileID, permissionID string, supportsAllDrives bool) error

	// GetFile retrieves metadata for a Drive file.
	GetFile(ctx context.Context, fileID string, supportsAllDrives bool) (*drive.File, error)

	// CreateFile creates a new Drive file (including folders) and returns its metadata.
	CreateFile(ctx context.Context, f *drive.File, supportsAllDrives bool) (*drive.File, error)

	// UpdateFile updates a Drive file's metadata and/or parents.
	// addParents/removeParents should be comma-separated folder IDs (or empty).
	UpdateFile(ctx context.Context, fileID string, f *drive.File, addParents string, removeParents string, supportsAllDrives bool) (*drive.File, error)

	// ListFiles lists Drive files matching a query.
	// q uses Drive Files.list query syntax (e.g. "mimeType='...'" and "name contains '...'" etc).
	ListFiles(ctx context.Context, q string, supportsAllDrives bool) ([]*drive.File, error)
}

// SheetsAPI defines the interface for Google Sheets API operations.
type SheetsAPI interface {
	// Create creates a new spreadsheet.
	Create(ctx context.Context, s *sheets.Spreadsheet) (*sheets.Spreadsheet, error)

	// Get retrieves a spreadsheet by ID.
	Get(ctx context.Context, spreadsheetID string) (*sheets.Spreadsheet, error)

	// BatchUpdate applies a batch of update requests to a spreadsheet.
	BatchUpdate(ctx context.Context, spreadsheetID string, req *sheets.BatchUpdateSpreadsheetRequest) (*sheets.BatchUpdateSpreadsheetResponse, error)

	// ValuesGet retrieves a value range (bounded) from a spreadsheet.
	ValuesGet(ctx context.Context, spreadsheetID, rng string) (*sheets.ValueRange, error)

	// ValuesUpdate writes a value range (bounded) into a spreadsheet.
	ValuesUpdate(ctx context.Context, spreadsheetID, rng string, vr *sheets.ValueRange, valueInputOption string) (*sheets.UpdateValuesResponse, error)

	// ValuesClear clears values in a range.
	ValuesClear(ctx context.Context, spreadsheetID, rng string) error
}

// Client holds the API clients used by the provider.
type Client struct {
	Forms  FormsAPI
	Drive  DriveAPI
	Sheets SheetsAPI
}
