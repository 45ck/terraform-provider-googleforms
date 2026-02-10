// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetvalues

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sheets "google.golang.org/api/sheets/v4"
)

func TestPlanToValueRange_ConvertsRows(t *testing.T) {
	t.Parallel()

	rowType := types.ObjectType{AttrTypes: map[string]attr.Type{"cells": types.ListType{ElemType: types.StringType}}}

	c1, d := types.ListValue(types.StringType, []attr.Value{types.StringValue("a"), types.StringValue("b")})
	if d.HasError() {
		t.Fatalf("cells1 diags: %v", d)
	}
	r1, d := types.ObjectValue(rowType.AttrTypes, map[string]attr.Value{"cells": c1})
	if d.HasError() {
		t.Fatalf("row1 diags: %v", d)
	}

	c2, d := types.ListValue(types.StringType, []attr.Value{types.StringValue("c"), types.StringValue("d")})
	if d.HasError() {
		t.Fatalf("cells2 diags: %v", d)
	}
	r2, d := types.ObjectValue(rowType.AttrTypes, map[string]attr.Value{"cells": c2})
	if d.HasError() {
		t.Fatalf("row2 diags: %v", d)
	}

	rows, d := types.ListValue(rowType, []attr.Value{r1, r2})
	if d.HasError() {
		t.Fatalf("rows diags: %v", d)
	}

	vr, diags := planToValueRange(context.Background(), rows, "Sheet1!A1:B2")
	if diags.HasError() {
		t.Fatalf("planToValueRange diags: %v", diags)
	}

	if vr.Range != "Sheet1!A1:B2" {
		t.Fatalf("expected range %q, got %q", "Sheet1!A1:B2", vr.Range)
	}
	if len(vr.Values) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(vr.Values))
	}
	if vr.Values[0][0] != "a" || vr.Values[0][1] != "b" {
		t.Fatalf("unexpected row0: %#v", vr.Values[0])
	}
	if vr.Values[1][0] != "c" || vr.Values[1][1] != "d" {
		t.Fatalf("unexpected row1: %#v", vr.Values[1])
	}
}

func TestValueRangeToRows_NormalizesToStrings(t *testing.T) {
	t.Parallel()

	vr := &sheets.ValueRange{
		Range: "Sheet1!A1:B2",
		Values: [][]interface{}{
			{"x", 1},
			{true, "y"},
		},
	}

	rows, diags := valueRangeToRows(context.Background(), vr)
	if diags.HasError() {
		t.Fatalf("valueRangeToRows diags: %v", diags)
	}

	var decoded []SheetValuesRowModel
	diags.Append(rows.ElementsAs(context.Background(), &decoded, false)...)
	if diags.HasError() {
		t.Fatalf("decode rows diags: %v", diags)
	}

	if len(decoded) != 2 {
		t.Fatalf("expected 2 decoded rows, got %d", len(decoded))
	}

	var cells0 []string
	diags.Append(decoded[0].Cells.ElementsAs(context.Background(), &cells0, false)...)
	if diags.HasError() {
		t.Fatalf("cells0 diags: %v", diags)
	}
	if cells0[0] != "x" || cells0[1] != "1" {
		t.Fatalf("unexpected cells0: %#v", cells0)
	}
}
