// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"errors"
	"fmt"

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

// wrapDriveAPIError converts a googleapi.Error from Drive API into the
// appropriate custom error type.
func wrapDriveAPIError(err error, operation string) error {
	var gErr *googleapi.Error
	if !errors.As(err, &gErr) {
		return fmt.Errorf("%s: %w", operation, err)
	}

	return mapStatusToError(gErr.Code, gErr.Message, operation, "file")
}
