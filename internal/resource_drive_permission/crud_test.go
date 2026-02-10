// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcedrivepermission

import "testing"

func TestSplit_Valid(t *testing.T) {
	t.Parallel()

	fileID, permID, diags := split("file123#perm456")
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if fileID != "file123" || permID != "perm456" {
		t.Fatalf("unexpected parts: %q %q", fileID, permID)
	}
}

func TestSplit_Invalid(t *testing.T) {
	t.Parallel()

	_, _, diags := split("nope")
	if !diags.HasError() {
		t.Fatal("expected error diagnostics")
	}
}
