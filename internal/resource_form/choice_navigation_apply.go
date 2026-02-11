// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceform

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	forms "google.golang.org/api/forms/v1"

	"github.com/45ck/terraform-provider-googleforms/internal/convert"
)

func (r *FormResource) applyChoiceNavigationUpdates(
	ctx context.Context,
	formID string,
	desiredItems []convert.ItemModel,
	keyMap map[string]string,
) diag.Diagnostics {
	var diags diag.Diagnostics
	if len(desiredItems) == 0 || keyMap == nil {
		return diags
	}

	// Read current form state to get accurate indexes and item IDs.
	currentForm, err := r.client.Forms.Get(ctx, formID)
	if err != nil {
		diags.AddError("Read Form Failed", err.Error())
		return diags
	}

	// Build item_key -> item_id map.
	keyToID := make(map[string]string)
	for _, it := range currentForm.Items {
		if it == nil || it.ItemId == "" {
			continue
		}
		if k, ok := keyMap[it.ItemId]; ok && k != "" {
			keyToID[k] = it.ItemId
		}
	}

	if err := convert.ResolveChoiceOptionSectionIDs(desiredItems, keyToID, false); err != nil {
		diags.AddError("Resolve Choice Navigation Failed", err.Error())
		return diags
	}

	type itemRef struct {
		index int
		item  *forms.Item
	}
	byID := make(map[string]itemRef, len(currentForm.Items))
	for i, it := range currentForm.Items {
		if it == nil || it.ItemId == "" {
			continue
		}
		byID[it.ItemId] = itemRef{index: i, item: it}
	}

	var requests []*forms.Request
	for _, desired := range desiredItems {
		if desired.MultipleChoice == nil && desired.Dropdown == nil && desired.Checkbox == nil {
			continue
		}
		itemID := keyToID[desired.ItemKey]
		if itemID == "" {
			// Not tracked / not present (partial mode) - skip.
			continue
		}
		ref, ok := byID[itemID]
		if !ok || ref.item == nil {
			diags.AddError("Choice Navigation Apply Failed", fmt.Sprintf("could not find current item %s for %q", itemID, desired.ItemKey))
			return diags
		}

		updated, changed, needsReplace, err := convert.ApplyDesiredItem(ref.item, desired)
		if err != nil {
			diags.AddError("Choice Navigation Apply Failed", err.Error())
			return diags
		}
		if needsReplace {
			diags.AddError("Choice Navigation Apply Failed", fmt.Sprintf("item %q requires replace_all", desired.ItemKey))
			return diags
		}
		if changed {
			requests = append(requests, &forms.Request{
				UpdateItem: &forms.UpdateItemRequest{
					Item:       updated,
					Location:   &forms.Location{Index: int64(ref.index)},
					UpdateMask: "*",
				},
			})
		}
	}

	if len(requests) == 0 {
		return diags
	}

	_, err = r.client.Forms.BatchUpdate(ctx, formID, &forms.BatchUpdateFormRequest{
		Requests:              requests,
		IncludeFormInResponse: false,
	})
	if err != nil {
		diags.AddError("Choice Navigation Update Failed", err.Error())
		return diags
	}

	return diags
}

