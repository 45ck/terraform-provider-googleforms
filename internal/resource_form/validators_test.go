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

func itemBlockType(t *testing.T) tftypes.Object {
	t.Helper()

	schemaResp := testSchemaResp()
	s := schemaResp.Schema

	tfType := s.Type().TerraformType(context.Background())
	objType, ok := tfType.(tftypes.Object)
	if !ok {
		t.Fatalf("expected tftypes.Object, got %T", tfType)
	}

	itemListType, ok := objType.AttributeTypes["item"].(tftypes.List)
	if !ok {
		t.Fatalf("expected item to be tftypes.List, got %T", objType.AttributeTypes["item"])
	}

	iType, ok := itemListType.ElementType.(tftypes.Object)
	if !ok {
		t.Fatalf("expected item element type to be tftypes.Object, got %T", itemListType.ElementType)
	}

	return iType
}

func newObjectValue(objType tftypes.Object, overrides map[string]tftypes.Value) tftypes.Value {
	merged := make(map[string]tftypes.Value)
	for k, v := range objType.AttributeTypes {
		merged[k] = tftypes.NewValue(v, nil)
	}
	for k, v := range overrides {
		merged[k] = v
	}
	return tftypes.NewValue(objType, merged)
}

func mcItem(
	t *testing.T,
	key, questionText string,
	options []string,
	grading *map[string]tftypes.Value,
) tftypes.Value {
	iType := itemBlockType(t)
	mcType, ok := iType.AttributeTypes["multiple_choice"].(tftypes.Object)
	if !ok {
		t.Fatalf("expected multiple_choice to be tftypes.Object, got %T", iType.AttributeTypes["multiple_choice"])
	}
	gType := mcType.AttributeTypes["grading"]
	gVal := tftypes.NewValue(gType, nil)
	if grading != nil {
		gVal = tftypes.NewValue(gType, *grading)
	}

	optVals := make([]tftypes.Value, len(options))
	for i, o := range options {
		optVals[i] = tftypes.NewValue(tftypes.String, o)
	}

	mc := newObjectValue(mcType, map[string]tftypes.Value{
		"question_text": tftypes.NewValue(tftypes.String, questionText),
		"options":       tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, optVals),
		"required":      tftypes.NewValue(tftypes.Bool, false),
		"grading":       gVal,
	})
	return newItemVal(iType, key, map[string]tftypes.Value{"multiple_choice": mc})
}

func saItem(t *testing.T, key, questionText string, grading *map[string]tftypes.Value) tftypes.Value {
	iType := itemBlockType(t)
	saType, ok := iType.AttributeTypes["short_answer"].(tftypes.Object)
	if !ok {
		t.Fatalf("expected short_answer to be tftypes.Object, got %T", iType.AttributeTypes["short_answer"])
	}
	gType := saType.AttributeTypes["grading"]
	gVal := tftypes.NewValue(gType, nil)
	if grading != nil {
		gVal = tftypes.NewValue(gType, *grading)
	}

	sa := newObjectValue(saType, map[string]tftypes.Value{
		"question_text": tftypes.NewValue(tftypes.String, questionText),
		"required":      tftypes.NewValue(tftypes.Bool, false),
		"grading":       gVal,
	})
	return newItemVal(iType, key, map[string]tftypes.Value{"short_answer": sa})
}

func paraItem(t *testing.T, key, questionText string, grading *map[string]tftypes.Value) tftypes.Value {
	iType := itemBlockType(t)
	paraType, ok := iType.AttributeTypes["paragraph"].(tftypes.Object)
	if !ok {
		t.Fatalf("expected paragraph to be tftypes.Object, got %T", iType.AttributeTypes["paragraph"])
	}
	gType := paraType.AttributeTypes["grading"]
	gVal := tftypes.NewValue(gType, nil)
	if grading != nil {
		gVal = tftypes.NewValue(gType, *grading)
	}

	para := newObjectValue(paraType, map[string]tftypes.Value{
		"question_text": tftypes.NewValue(tftypes.String, questionText),
		"required":      tftypes.NewValue(tftypes.Bool, false),
		"grading":       gVal,
	})
	return newItemVal(iType, key, map[string]tftypes.Value{"paragraph": para})
}

func bareItem(t *testing.T, key string) tftypes.Value {
	iType := itemBlockType(t)
	return newItemVal(iType, key, nil)
}

func newItemVal(
	iType tftypes.Object,
	key string,
	overrides map[string]tftypes.Value,
) tftypes.Value {
	merged := make(map[string]tftypes.Value)
	for k, v := range iType.AttributeTypes {
		merged[k] = tftypes.NewValue(v, nil)
	}

	merged["item_key"] = tftypes.NewValue(tftypes.String, key)
	if _, ok := merged["google_item_id"]; ok {
		merged["google_item_id"] = tftypes.NewValue(tftypes.String, nil)
	}

	for k, v := range overrides {
		merged[k] = v
	}

	return tftypes.NewValue(iType, merged)
}

func itemListVal(t *testing.T, items ...tftypes.Value) tftypes.Value {
	iType := itemBlockType(t)
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
		"item":         itemListVal(t, saItem(t, "q1", "Question?", nil)),
	})
	diags := runValidators(t, cfg, MutuallyExclusiveValidator{})
	expectErrorContains(t, diags, "Cannot use both")
}

func TestMutuallyExclusiveValidator_OnlyItems_Passes(t *testing.T) {
	t.Parallel()
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title": tftypes.NewValue(tftypes.String, "T"),
		"item":  itemListVal(t, saItem(t, "q1", "Question?", nil)),
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
		"item":         itemListVal(t, saItem(t, "q1", "Question?", nil)),
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
		"item": itemListVal(t,
			saItem(t, "q1", "Q1?", nil),
			saItem(t, "q2", "Q2?", nil),
		),
	})
	diags := runValidators(t, cfg, UniqueItemKeyValidator{})
	expectNoError(t, diags)
}

func TestUniqueItemKeyValidator_DuplicateKeys_Error(t *testing.T) {
	t.Parallel()
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title": tftypes.NewValue(tftypes.String, "T"),
		"item": itemListVal(t,
			saItem(t, "q1", "Q1?", nil),
			saItem(t, "q1", "Q2?", nil),
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
		"item":  itemListVal(t, saItem(t, "q1", "Q?", nil)),
	})
	diags := runValidators(t, cfg, ExactlyOneSubBlockValidator{})
	expectNoError(t, diags)
}

func TestExactlyOneSubBlockValidator_NoneSet_Error(t *testing.T) {
	t.Parallel()
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title": tftypes.NewValue(tftypes.String, "T"),
		"item":  itemListVal(t, bareItem(t, "q1")),
	})
	diags := runValidators(t, cfg, ExactlyOneSubBlockValidator{})
	expectErrorContains(t, diags, "exactly one question type")
}

func TestExactlyOneSubBlockValidator_TwoSet_Error(t *testing.T) {
	t.Parallel()
	iType := itemBlockType(t)
	saType, ok := iType.AttributeTypes["short_answer"].(tftypes.Object)
	if !ok {
		t.Fatalf("expected short_answer to be tftypes.Object, got %T", iType.AttributeTypes["short_answer"])
	}
	gType := saType.AttributeTypes["grading"]

	mcType, ok := iType.AttributeTypes["multiple_choice"].(tftypes.Object)
	if !ok {
		t.Fatalf("expected multiple_choice to be tftypes.Object, got %T", iType.AttributeTypes["multiple_choice"])
	}
	mc := newObjectValue(mcType, map[string]tftypes.Value{
		"question_text": tftypes.NewValue(tftypes.String, "MC?"),
		"options": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "A"),
		}),
		"required": tftypes.NewValue(tftypes.Bool, false),
		"grading":  tftypes.NewValue(gType, nil),
	})
	sa := newObjectValue(saType, map[string]tftypes.Value{
		"question_text": tftypes.NewValue(tftypes.String, "SA?"),
		"required":      tftypes.NewValue(tftypes.Bool, false),
		"grading":       tftypes.NewValue(gType, nil),
	})
	twoBlock := newItemVal(iType, "q1", map[string]tftypes.Value{
		"multiple_choice": mc,
		"short_answer":    sa,
	})

	cfg := buildConfig(t, map[string]tftypes.Value{
		"title": tftypes.NewValue(tftypes.String, "T"),
		"item":  itemListVal(t, twoBlock),
	})
	diags := runValidators(t, cfg, ExactlyOneSubBlockValidator{})
	expectErrorContains(t, diags, "exactly one question type")
}
