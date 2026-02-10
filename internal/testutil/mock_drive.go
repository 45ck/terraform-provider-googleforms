// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"context"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
	drive "google.golang.org/api/drive/v3"
)

// MockDriveAPI is a configurable mock implementation of client.DriveAPI.
type MockDriveAPI struct {
	DeleteFunc func(ctx context.Context, fileID string) error

	GetParentsFunc   func(ctx context.Context, fileID string, supportsAllDrives bool) ([]string, error)
	MoveToFolderFunc func(ctx context.Context, fileID string, folderID string, supportsAllDrives bool) error

	CreatePermissionFunc func(ctx context.Context, fileID string, p *drive.Permission, sendNotificationEmail bool, emailMessage string, supportsAllDrives bool) (*drive.Permission, error)
	GetPermissionFunc    func(ctx context.Context, fileID, permissionID string, supportsAllDrives bool) (*drive.Permission, error)
	DeletePermissionFunc func(ctx context.Context, fileID, permissionID string, supportsAllDrives bool) error
}

var _ client.DriveAPI = &MockDriveAPI{}

func (m *MockDriveAPI) Delete(ctx context.Context, fileID string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, fileID)
	}
	return nil
}

func (m *MockDriveAPI) GetParents(
	ctx context.Context,
	fileID string,
	supportsAllDrives bool,
) ([]string, error) {
	if m.GetParentsFunc != nil {
		return m.GetParentsFunc(ctx, fileID, supportsAllDrives)
	}
	return []string{}, nil
}

func (m *MockDriveAPI) MoveToFolder(
	ctx context.Context,
	fileID string,
	folderID string,
	supportsAllDrives bool,
) error {
	if m.MoveToFolderFunc != nil {
		return m.MoveToFolderFunc(ctx, fileID, folderID, supportsAllDrives)
	}
	return nil
}

func (m *MockDriveAPI) CreatePermission(
	ctx context.Context,
	fileID string,
	p *drive.Permission,
	sendNotificationEmail bool,
	emailMessage string,
	supportsAllDrives bool,
) (*drive.Permission, error) {
	if m.CreatePermissionFunc != nil {
		return m.CreatePermissionFunc(ctx, fileID, p, sendNotificationEmail, emailMessage, supportsAllDrives)
	}
	return &drive.Permission{Id: "mock-permission-id"}, nil
}

func (m *MockDriveAPI) GetPermission(
	ctx context.Context,
	fileID string,
	permissionID string,
	supportsAllDrives bool,
) (*drive.Permission, error) {
	if m.GetPermissionFunc != nil {
		return m.GetPermissionFunc(ctx, fileID, permissionID, supportsAllDrives)
	}
	return &drive.Permission{Id: permissionID}, nil
}

func (m *MockDriveAPI) DeletePermission(
	ctx context.Context,
	fileID string,
	permissionID string,
	supportsAllDrives bool,
) error {
	if m.DeletePermissionFunc != nil {
		return m.DeletePermissionFunc(ctx, fileID, permissionID, supportsAllDrives)
	}
	return nil
}
