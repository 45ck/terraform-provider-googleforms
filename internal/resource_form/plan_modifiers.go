// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceform

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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

	configHash := normalizeJSONHash(req.ConfigValue.ValueString())
	stateHash := normalizeJSONHash(req.StateValue.ValueString())

	if configHash == stateHash {
		resp.PlanValue = req.StateValue
	}
}

// normalizeJSONHash returns a hex-encoded SHA-256 hash of the compact JSON
// representation of the input string. If the input is not valid JSON, the
// raw string is hashed instead.
func normalizeJSONHash(raw string) string {
	var parsed interface{}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return hashString(raw)
	}
	compact, err := json.Marshal(parsed)
	if err != nil {
		return hashString(raw)
	}
	return hashString(string(compact))
}

// hashString returns the hex-encoded SHA-256 hash of s.
func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}
