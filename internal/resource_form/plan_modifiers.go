// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceform

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"

	"github.com/45ck/terraform-provider-googleforms/internal/convert"
)

// Compile-time interface check.
var _ planmodifier.String = ContentJSONHashModifier{}

// ContentJSONHashModifier suppresses diffs for content_json when the
// normalized JSON content is semantically equivalent. It computes a SHA-256
// hash of the compact, key-sorted JSON to compare config and state values.
type ContentJSONHashModifier struct{}

func (m ContentJSONHashModifier) Description(_ context.Context) string {
	return "Suppresses diff when normalized JSON hashes match."
}

func (m ContentJSONHashModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

// PlanModifyString compares the config and state values after JSON
// normalization. If they produce the same hash, the plan value is set
// to the state value to suppress the diff.
func (m ContentJSONHashModifier) PlanModifyString(
	_ context.Context,
	req planmodifier.StringRequest,
	resp *planmodifier.StringResponse,
) {
	// If there is no state value (create) or no config value, skip.
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	configHash, configErr := convert.HashJSON(req.ConfigValue.ValueString())
	stateHash, stateErr := convert.HashJSON(req.StateValue.ValueString())

	if configErr != nil || stateErr != nil {
		// If either value is not valid JSON, fall through to default
		// plan behavior (show a diff).
		return
	}

	if configHash == stateHash {
		resp.PlanValue = req.StateValue
	}
}
