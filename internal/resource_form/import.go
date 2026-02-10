// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceform

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// ImportState handles terraform import for existing Google Forms.
// Usage: terraform import googleforms_form.example FORM_ID
//
// After import, items receive auto-generated item_keys (item_0, item_1, ...).
// Users should review and rename these in their configuration.
func (r *FormResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
