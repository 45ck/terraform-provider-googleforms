// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/api/googleapi"

	drive "google.golang.org/api/drive/v3"
)

// DriveAPIClient is the real implementation of DriveAPI using Google's API.
type DriveAPIClient struct {
	service *drive.Service
	retry   RetryConfig
}

// NewDriveAPIClient creates a new DriveAPIClient.
func NewDriveAPIClient(service *drive.Service, retry RetryConfig) *DriveAPIClient {
	return &DriveAPIClient{service: service, retry: retry}
}

var _ DriveAPI = &DriveAPIClient{}

// Delete deletes a file by ID via the Google Drive API.
// Returns nil if the file is already deleted (404 treated as success).
func (c *DriveAPIClient) Delete(
	ctx context.Context,
	fileID string,
) error {
	err := WithRetry(ctx, c.retry, func() error {
		apiErr := c.service.Files.Delete(fileID).Context(ctx).Do()
		if apiErr != nil {
			return wrapDriveAPIError(apiErr, "delete file "+fileID)
		}
		return nil
	})

	if err != nil && IsNotFound(err) {
		// File already deleted; treat as success.
		return nil
	}

	if err != nil {
		return fmt.Errorf("drive.Delete: %w", err)
	}

	return nil
}

// CreatePermission creates a permission on a Drive file.
func (c *DriveAPIClient) CreatePermission(
	ctx context.Context,
	fileID string,
	p *drive.Permission,
	sendNotificationEmail bool,
	emailMessage string,
	supportsAllDrives bool,
) (*drive.Permission, error) {
	var result *drive.Permission

	err := WithRetry(ctx, c.retry, func() error {
		call := c.service.Permissions.Create(fileID, p).
			Context(ctx).
			SupportsAllDrives(supportsAllDrives).
			SendNotificationEmail(sendNotificationEmail)
		if strings.TrimSpace(emailMessage) != "" {
			call = call.EmailMessage(emailMessage)
		}

		resp, apiErr := call.Do()
		if apiErr != nil {
			return wrapDriveAPIError(apiErr, "create permission on file "+fileID)
		}
		result = resp
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("drive.CreatePermission: %w", err)
	}

	return result, nil
}

// GetPermission retrieves a permission by ID from a Drive file.
func (c *DriveAPIClient) GetPermission(
	ctx context.Context,
	fileID string,
	permissionID string,
	supportsAllDrives bool,
) (*drive.Permission, error) {
	var result *drive.Permission

	err := WithRetry(ctx, c.retry, func() error {
		resp, apiErr := c.service.Permissions.Get(fileID, permissionID).
			Context(ctx).
			SupportsAllDrives(supportsAllDrives).
			Do()
		if apiErr != nil {
			return wrapDriveAPIError(apiErr, "get permission "+permissionID+" on file "+fileID)
		}
		result = resp
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("drive.GetPermission: %w", err)
	}

	return result, nil
}

// DeletePermission deletes a permission from a Drive file.
// Returns nil if the permission or file is already gone (404 treated as success).
func (c *DriveAPIClient) DeletePermission(
	ctx context.Context,
	fileID string,
	permissionID string,
	supportsAllDrives bool,
) error {
	err := WithRetry(ctx, c.retry, func() error {
		apiErr := c.service.Permissions.Delete(fileID, permissionID).
			Context(ctx).
			SupportsAllDrives(supportsAllDrives).
			Do()
		if apiErr != nil {
			return wrapDriveAPIError(apiErr, "delete permission "+permissionID+" on file "+fileID)
		}
		return nil
	})

	if err != nil && IsNotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("drive.DeletePermission: %w", err)
	}

	return nil
}

// wrapDriveAPIError converts a googleapi.Error from Drive API into the
// appropriate custom error type.
func wrapDriveAPIError(err error, operation string) error {
	var gErr *googleapi.Error
	if !errors.As(err, &gErr) {
		return fmt.Errorf("%s: %w", operation, err)
	}

	return mapStatusToError(gErr.Code, gErr.Message, operation, "file")
}
