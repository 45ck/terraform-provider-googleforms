// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceform

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestContentJSONHashModifier_Description(t *testing.T) {
	t.Parallel()
	m := ContentJSONHashModifier{}
	if d := m.Description(context.Background()); d == "" {
		t.Error("Description should not be empty")
	}
	if md := m.MarkdownDescription(context.Background()); md == "" {
		t.Error("MarkdownDescription should not be empty")
	}
}

func TestContentJSONHashModifier_SemanticallyEqual_SuppressesDiff(t *testing.T) {
	t.Parallel()

	// State has compact JSON, config has whitespace/reordered keys — semantically equal.
	stateJSON := `{"title":"Q1","questionItem":{"question":{"required":true,"textQuestion":{}}}}`
	configJSON := `{
		"questionItem": { "question": { "textQuestion": {}, "required": true } },
		"title": "Q1"
	}`

	req := planmodifier.StringRequest{
		StateValue:  types.StringValue(stateJSON),
		ConfigValue: types.StringValue(configJSON),
		PlanValue:   types.StringValue(configJSON),
	}
	resp := &planmodifier.StringResponse{PlanValue: types.StringValue(configJSON)}

	ContentJSONHashModifier{}.PlanModifyString(context.Background(), req, resp)

	// Plan should be set to state value (diff suppressed).
	if resp.PlanValue.ValueString() != stateJSON {
		t.Errorf("expected plan to be state value %q, got %q", stateJSON, resp.PlanValue.ValueString())
	}
}

func TestContentJSONHashModifier_DifferentContent_ShowsDiff(t *testing.T) {
	t.Parallel()

	stateJSON := `{"title":"Q1"}`
	configJSON := `{"title":"Q2"}`

	req := planmodifier.StringRequest{
		StateValue:  types.StringValue(stateJSON),
		ConfigValue: types.StringValue(configJSON),
		PlanValue:   types.StringValue(configJSON),
	}
	resp := &planmodifier.StringResponse{PlanValue: types.StringValue(configJSON)}

	ContentJSONHashModifier{}.PlanModifyString(context.Background(), req, resp)

	// Plan should remain as config value (diff NOT suppressed).
	if resp.PlanValue.ValueString() != configJSON {
		t.Errorf("expected plan to remain %q, got %q", configJSON, resp.PlanValue.ValueString())
	}
}

func TestContentJSONHashModifier_NullState_Skips(t *testing.T) {
	t.Parallel()

	configJSON := `{"title":"Q1"}`

	req := planmodifier.StringRequest{
		StateValue:  types.StringNull(),
		ConfigValue: types.StringValue(configJSON),
		PlanValue:   types.StringValue(configJSON),
	}
	resp := &planmodifier.StringResponse{PlanValue: types.StringValue(configJSON)}

	ContentJSONHashModifier{}.PlanModifyString(context.Background(), req, resp)

	if resp.PlanValue.ValueString() != configJSON {
		t.Errorf("expected plan unchanged %q, got %q", configJSON, resp.PlanValue.ValueString())
	}
}

func TestContentJSONHashModifier_UnknownState_Skips(t *testing.T) {
	t.Parallel()

	configJSON := `{"title":"Q1"}`

	req := planmodifier.StringRequest{
		StateValue:  types.StringUnknown(),
		ConfigValue: types.StringValue(configJSON),
		PlanValue:   types.StringValue(configJSON),
	}
	resp := &planmodifier.StringResponse{PlanValue: types.StringValue(configJSON)}

	ContentJSONHashModifier{}.PlanModifyString(context.Background(), req, resp)

	if resp.PlanValue.ValueString() != configJSON {
		t.Errorf("expected plan unchanged %q, got %q", configJSON, resp.PlanValue.ValueString())
	}
}

func TestContentJSONHashModifier_NullConfig_Skips(t *testing.T) {
	t.Parallel()

	stateJSON := `{"title":"Q1"}`

	req := planmodifier.StringRequest{
		StateValue:  types.StringValue(stateJSON),
		ConfigValue: types.StringNull(),
		PlanValue:   types.StringNull(),
	}
	resp := &planmodifier.StringResponse{PlanValue: types.StringNull()}

	ContentJSONHashModifier{}.PlanModifyString(context.Background(), req, resp)

	if !resp.PlanValue.IsNull() {
		t.Errorf("expected plan to remain null, got %q", resp.PlanValue.ValueString())
	}
}

func TestContentJSONHashModifier_InvalidJSON_FallsThrough(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		state  string
		config string
	}{
		{
			name:   "invalid state JSON",
			state:  `{not valid}`,
			config: `{"title":"Q1"}`,
		},
		{
			name:   "invalid config JSON",
			state:  `{"title":"Q1"}`,
			config: `{not valid}`,
		},
		{
			name:   "both invalid",
			state:  `{bad}`,
			config: `{also bad}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := planmodifier.StringRequest{
				StateValue:  types.StringValue(tt.state),
				ConfigValue: types.StringValue(tt.config),
				PlanValue:   types.StringValue(tt.config),
			}
			resp := &planmodifier.StringResponse{PlanValue: types.StringValue(tt.config)}

			ContentJSONHashModifier{}.PlanModifyString(context.Background(), req, resp)

			// Should fall through without modification (show diff).
			if resp.PlanValue.ValueString() != tt.config {
				t.Errorf("expected plan unchanged %q, got %q", tt.config, resp.PlanValue.ValueString())
			}
		})
	}
}

// Ensure the type assertion in the compile-time check holds.
func TestContentJSONHashModifier_ImplementsInterface(t *testing.T) {
	t.Parallel()
	var _ planmodifier.String = ContentJSONHashModifier{}
}
