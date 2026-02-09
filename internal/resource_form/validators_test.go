// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceform

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ---------------------------------------------------------------------------
// test helpers
// ---------------------------------------------------------------------------

func testSchemaResp() resource.SchemaResponse {
	var resp resource.SchemaResponse
	r := &FormResource{}
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

func buildConfig(t *testing.T, vals map[string]tftypes.Value) tfsdk.Config {
	t.Helper()
	schemaResp := testSchemaResp()
	s := schemaResp.Schema

	tfType := s.Type().TerraformType(context.Background())
	objType, ok := tfType.(tftypes.Object)
	if !ok {
		t.Fatalf("expected tftypes.Object, got %T", tfType)
	}

	merged := make(map[string]tftypes.Value)
	for k, v := range objType.AttributeTypes {
		merged[k] = tftypes.NewValue(v, nil)
	}
	for k, v := range vals {
		merged[k] = v
	}

	return tfsdk.Config{
		Schema: s,
		Raw:    tftypes.NewValue(objType, merged),
	}
}

func runValidators(
	t *testing.T,
	cfg tfsdk.Config,
	validators ...resource.ConfigValidator,
) diag.Diagnostics {
	t.Helper()
	var allDiags diag.Diagnostics
	for _, v := range validators {
		req := resource.ValidateConfigRequest{Config: cfg}
		resp := &resource.ValidateConfigResponse{}
		v.ValidateResource(context.Background(), req, resp)
		allDiags.Append(resp.Diagnostics...)
	}
	return allDiags
}

func expectError(t *testing.T, diags diag.Diagnostics) {
	t.Helper()
	if !diags.HasError() {
		t.Fatal("expected error diagnostic, got none")
	}
}

func expectNoError(t *testing.T, diags diag.Diagnostics) {
	t.Helper()
	if diags.HasError() {
		for _, d := range diags.Errors() {
			t.Logf("  error: %s -- %s", d.Summary(), d.Detail())
		}
		t.Fatal("expected no errors, got some")
	}
}

func expectErrorContains(t *testing.T, diags diag.Diagnostics, sub string) {
	t.Helper()
	expectError(t, diags)
	for _, d := range diags.Errors() {
		if strings.Contains(d.Detail(), sub) || strings.Contains(d.Summary(), sub) {
			return
		}
	}
	t.Fatalf("no error diagnostic contains %q", sub)
}

// ---------------------------------------------------------------------------
// item-block tftypes helpers
// ---------------------------------------------------------------------------

func itemBlockType() tftypes.Object {
	gradingType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"points":             tftypes.Number,
			"correct_answer":     tftypes.String,
			"feedback_correct":   tftypes.String,
			"feedback_incorrect": tftypes.String,
		},
	}
	mcType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"question_text": tftypes.String,
			"options":       tftypes.List{ElementType: tftypes.String},
			"required":      tftypes.Bool,
			"grading":       gradingType,
		},
	}
	saType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"question_text": tftypes.String,
			"required":      tftypes.Bool,
			"grading":       gradingType,
		},
	}
	paraType := saType
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"item_key":        tftypes.String,
			"google_item_id":  tftypes.String,
			"multiple_choice": mcType,
			"short_answer":    saType,
			"paragraph":       paraType,
		},
	}
}

func mcItem(
	key, questionText string, options []string, grading *map[string]tftypes.Value,
) tftypes.Value {
	iType := itemBlockType()
	gType := iType.AttributeTypes["multiple_choice"].(tftypes.Object).AttributeTypes["grading"]
	gVal := tftypes.NewValue(gType, nil)
	if grading != nil {
		gVal = tftypes.NewValue(gType, *grading)
	}

	optVals := make([]tftypes.Value, len(options))
	for i, o := range options {
		optVals[i] = tftypes.NewValue(tftypes.String, o)
	}

	mc := tftypes.NewValue(iType.AttributeTypes["multiple_choice"], map[string]tftypes.Value{
		"question_text": tftypes.NewValue(tftypes.String, questionText),
		"options":       tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, optVals),
		"required":      tftypes.NewValue(tftypes.Bool, false),
		"grading":       gVal,
	})
	return newItemVal(iType, key, &mc, nil, nil)
}

func saItem(key, questionText string, grading *map[string]tftypes.Value) tftypes.Value {
	iType := itemBlockType()
	gType := iType.AttributeTypes["short_answer"].(tftypes.Object).AttributeTypes["grading"]
	gVal := tftypes.NewValue(gType, nil)
	if grading != nil {
		gVal = tftypes.NewValue(gType, *grading)
	}

	sa := tftypes.NewValue(iType.AttributeTypes["short_answer"], map[string]tftypes.Value{
		"question_text": tftypes.NewValue(tftypes.String, questionText),
		"required":      tftypes.NewValue(tftypes.Bool, false),
		"grading":       gVal,
	})
	return newItemVal(iType, key, nil, &sa, nil)
}

func paraItem(key, questionText string, grading *map[string]tftypes.Value) tftypes.Value {
	iType := itemBlockType()
	gType := iType.AttributeTypes["paragraph"].(tftypes.Object).AttributeTypes["grading"]
	gVal := tftypes.NewValue(gType, nil)
	if grading != nil {
		gVal = tftypes.NewValue(gType, *grading)
	}

	para := tftypes.NewValue(iType.AttributeTypes["paragraph"], map[string]tftypes.Value{
		"question_text": tftypes.NewValue(tftypes.String, questionText),
		"required":      tftypes.NewValue(tftypes.Bool, false),
		"grading":       gVal,
	})
	return newItemVal(iType, key, nil, nil, &para)
}

func bareItem(key string) tftypes.Value {
	iType := itemBlockType()
	return newItemVal(iType, key, nil, nil, nil)
}

func newItemVal(
	iType tftypes.Object,
	key string,
	mc *tftypes.Value,
	sa *tftypes.Value,
	para *tftypes.Value,
) tftypes.Value {
	mcVal := tftypes.NewValue(iType.AttributeTypes["multiple_choice"], nil)
	if mc != nil {
		mcVal = *mc
	}
	saVal := tftypes.NewValue(iType.AttributeTypes["short_answer"], nil)
	if sa != nil {
		saVal = *sa
	}
	paraVal := tftypes.NewValue(iType.AttributeTypes["paragraph"], nil)
	if para != nil {
		paraVal = *para
	}
	return tftypes.NewValue(iType, map[string]tftypes.Value{
		"item_key":        tftypes.NewValue(tftypes.String, key),
		"google_item_id":  tftypes.NewValue(tftypes.String, nil),
		"multiple_choice": mcVal,
		"short_answer":    saVal,
		"paragraph":       paraVal,
	})
}

func itemListVal(items ...tftypes.Value) tftypes.Value {
	iType := itemBlockType()
	if len(items) == 0 {
		return tftypes.NewValue(tftypes.List{ElementType: iType}, []tftypes.Value{})
	}
	return tftypes.NewValue(tftypes.List{ElementType: iType}, items)
}

// ---------------------------------------------------------------------------
// 1. MutuallyExclusiveValidator
// ---------------------------------------------------------------------------

func TestMutuallyExclusiveValidator_BothSet_ReturnsError(t *testing.T) {
	t.Parallel()
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title":        tftypes.NewValue(tftypes.String, "T"),
		"content_json": tftypes.NewValue(tftypes.String, `[{"type":"short_answer"}]`),
		"item":         itemListVal(saItem("q1", "Question?", nil)),
	})
	diags := runValidators(t, cfg, MutuallyExclusiveValidator{})
	expectErrorContains(t, diags, "Cannot use both")
}

func TestMutuallyExclusiveValidator_OnlyItems_Passes(t *testing.T) {
	t.Parallel()
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title": tftypes.NewValue(tftypes.String, "T"),
		"item":  itemListVal(saItem("q1", "Question?", nil)),
	})
	diags := runValidators(t, cfg, MutuallyExclusiveValidator{})
	expectNoError(t, diags)
}

func TestMutuallyExclusiveValidator_OnlyContentJSON_Passes(t *testing.T) {
	t.Parallel()
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title":        tftypes.NewValue(tftypes.String, "T"),
		"content_json": tftypes.NewValue(tftypes.String, `[{"type":"short_answer"}]`),
	})
	diags := runValidators(t, cfg, MutuallyExclusiveValidator{})
	expectNoError(t, diags)
}

func TestMutuallyExclusiveValidator_NeitherSet_Passes(t *testing.T) {
	t.Parallel()
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title": tftypes.NewValue(tftypes.String, "T"),
	})
	diags := runValidators(t, cfg, MutuallyExclusiveValidator{})
	expectNoError(t, diags)
}

func TestMutuallyExclusive_EmptyContentJSON_WithItems_Error(t *testing.T) {
	t.Parallel()
	// Empty string is not null/unknown, so the validator treats it as "set".
	// Combined with items, this should trigger the mutually exclusive error.
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title":        tftypes.NewValue(tftypes.String, "T"),
		"content_json": tftypes.NewValue(tftypes.String, ""),
		"item":         itemListVal(saItem("q1", "Question?", nil)),
	})
	diags := runValidators(t, cfg, MutuallyExclusiveValidator{})
	expectErrorContains(t, diags, "Cannot use both")
}

// ---------------------------------------------------------------------------
// 2. AcceptingResponsesRequiresPublishedValidator
// ---------------------------------------------------------------------------

func TestAcceptingResponsesRequiresPublished_BothTrue_Passes(t *testing.T) {
	t.Parallel()
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title":               tftypes.NewValue(tftypes.String, "T"),
		"published":           tftypes.NewValue(tftypes.Bool, true),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, true),
	})
	diags := runValidators(t, cfg, AcceptingResponsesRequiresPublishedValidator{})
	expectNoError(t, diags)
}

func TestAcceptingResponsesRequiresPublished_AcceptingTruePublishedFalse_Error(t *testing.T) {
	t.Parallel()
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title":               tftypes.NewValue(tftypes.String, "T"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, true),
	})
	diags := runValidators(t, cfg, AcceptingResponsesRequiresPublishedValidator{})
	expectErrorContains(t, diags, "cannot accept responses while unpublished")
}

func TestAcceptingResponsesRequiresPublished_AcceptingFalse_Passes(t *testing.T) {
	t.Parallel()
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title":               tftypes.NewValue(tftypes.String, "T"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
	})
	diags := runValidators(t, cfg, AcceptingResponsesRequiresPublishedValidator{})
	expectNoError(t, diags)
}

// ---------------------------------------------------------------------------
// 3. UniqueItemKeyValidator
// ---------------------------------------------------------------------------

func TestUniqueItemKeyValidator_UniqueKeys_Passes(t *testing.T) {
	t.Parallel()
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title": tftypes.NewValue(tftypes.String, "T"),
		"item": itemListVal(
			saItem("q1", "Q1?", nil),
			saItem("q2", "Q2?", nil),
		),
	})
	diags := runValidators(t, cfg, UniqueItemKeyValidator{})
	expectNoError(t, diags)
}

func TestUniqueItemKeyValidator_DuplicateKeys_Error(t *testing.T) {
	t.Parallel()
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title": tftypes.NewValue(tftypes.String, "T"),
		"item": itemListVal(
			saItem("q1", "Q1?", nil),
			saItem("q1", "Q2?", nil),
		),
	})
	diags := runValidators(t, cfg, UniqueItemKeyValidator{})
	expectErrorContains(t, diags, `Duplicate item_key "q1"`)
}

func TestUniqueItemKeyValidator_EmptyItems_Passes(t *testing.T) {
	t.Parallel()
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title": tftypes.NewValue(tftypes.String, "T"),
	})
	diags := runValidators(t, cfg, UniqueItemKeyValidator{})
	expectNoError(t, diags)
}

// ---------------------------------------------------------------------------
// 4. ExactlyOneSubBlockValidator
// ---------------------------------------------------------------------------

func TestExactlyOneSubBlockValidator_OneSet_Passes(t *testing.T) {
	t.Parallel()
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title": tftypes.NewValue(tftypes.String, "T"),
		"item":  itemListVal(saItem("q1", "Q?", nil)),
	})
	diags := runValidators(t, cfg, ExactlyOneSubBlockValidator{})
	expectNoError(t, diags)
}

func TestExactlyOneSubBlockValidator_NoneSet_Error(t *testing.T) {
	t.Parallel()
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title": tftypes.NewValue(tftypes.String, "T"),
		"item":  itemListVal(bareItem("q1")),
	})
	diags := runValidators(t, cfg, ExactlyOneSubBlockValidator{})
	expectErrorContains(t, diags, "exactly one question type")
}

func TestExactlyOneSubBlockValidator_TwoSet_Error(t *testing.T) {
	t.Parallel()
	iType := itemBlockType()
	gType := iType.AttributeTypes["short_answer"].(tftypes.Object).AttributeTypes["grading"]

	mc := tftypes.NewValue(iType.AttributeTypes["multiple_choice"], map[string]tftypes.Value{
		"question_text": tftypes.NewValue(tftypes.String, "MC?"),
		"options": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "A"),
		}),
		"required": tftypes.NewValue(tftypes.Bool, false),
		"grading":  tftypes.NewValue(gType, nil),
	})
	sa := tftypes.NewValue(iType.AttributeTypes["short_answer"], map[string]tftypes.Value{
		"question_text": tftypes.NewValue(tftypes.String, "SA?"),
		"required":      tftypes.NewValue(tftypes.Bool, false),
		"grading":       tftypes.NewValue(gType, nil),
	})
	twoBlock := newItemVal(iType, "q1", &mc, &sa, nil)

	cfg := buildConfig(t, map[string]tftypes.Value{
		"title": tftypes.NewValue(tftypes.String, "T"),
		"item":  itemListVal(twoBlock),
	})
	diags := runValidators(t, cfg, ExactlyOneSubBlockValidator{})
	expectErrorContains(t, diags, "exactly one question type")
}
