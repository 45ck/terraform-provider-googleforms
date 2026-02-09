// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package testutil

// Sweeper cleans up orphaned test resources.
//
// Usage: go test ./internal/testutil/... -v -sweep=all
//
// The sweeper lists all forms owned by the test service account
// and deletes those with a "tf-test-" title prefix that are
// older than 1 hour.
//
// TODO: Implement:
// 1. Initialize Drive API client from GOOGLE_CREDENTIALS
// 2. List files with mimeType = "application/vnd.google-apps.form"
// 3. Filter by name prefix "tf-test-"
// 4. Filter by createdTime older than 1 hour
// 5. Delete matching files
