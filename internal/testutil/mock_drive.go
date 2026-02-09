// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"context"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

// MockDriveAPI is a configurable mock implementation of client.DriveAPI.
type MockDriveAPI struct {
	DeleteFunc func(ctx context.Context, fileID string) error
}

var _ client.DriveAPI = &MockDriveAPI{}

func (m *MockDriveAPI) Delete(ctx context.Context, fileID string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, fileID)
	}
	return nil
}
