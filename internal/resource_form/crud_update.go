// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceform

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	forms "google.golang.org/api/forms/v1"

	"github.com/45ck/terraform-provider-googleforms/internal/convert"
)

// Update replaces the form's settings and items with the planned configuration.
func (r *FormResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan FormResourceModel
	var state FormResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	formID := state.ID.ValueString()
	tflog.Debug(ctx, "updating Google Form", map[string]interface{}{
		"form_id": formID,
	})

	updateStrategy := "replace_all"
	if !plan.UpdateStrategy.IsNull() && !plan.UpdateStrategy.IsUnknown() && plan.UpdateStrategy.ValueString() != "" {
		updateStrategy = plan.UpdateStrategy.ValueString()
	}

	manageMode := "all"
	if !plan.ManageMode.IsNull() && !plan.ManageMode.IsUnknown() && plan.ManageMode.ValueString() != "" {
		manageMode = plan.ManageMode.ValueString()
	}

	dangerReplaceAll := false
	if !plan.DangerousReplaceAll.IsNull() && !plan.DangerousReplaceAll.IsUnknown() {
		dangerReplaceAll = plan.DangerousReplaceAll.ValueBool()
	}

	// Step 1: Fetch current form to know the existing item count.
	currentForm, err := r.client.Forms.Get(ctx, formID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Google Form Before Update",
			fmt.Sprintf("Could not read form %s: %s", formID, err),
		)
		return
	}

	// Step 2: Build and execute batch update requests.
	// Always update title/description. Items are updated based on update_strategy.
	var keyMap map[string]string
	switch updateStrategy {
	case "targeted":
		if manageMode == "partial" && !plan.ContentJSON.IsNull() && !plan.ContentJSON.IsUnknown() && plan.ContentJSON.ValueString() != "" {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"manage_mode = \"partial\" is not supported with content_json. Use item blocks for partial management.",
			)
			return
		}
		km, d := r.updateTargeted(ctx, plan, state, currentForm)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		keyMap = km
	case "replace_all":
		if manageMode == "partial" && (plan.ContentJSON.IsNull() || plan.ContentJSON.IsUnknown() || plan.ContentJSON.ValueString() == "") {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"manage_mode = \"partial\" cannot be used with update_strategy = \"replace_all\" because it would delete unmanaged items. Use update_strategy = \"targeted\".",
			)
			return
		}
		if !dangerReplaceAll {
			resp.Diagnostics.AddWarning(
				"Replace-All Item Updates Enabled",
				"update_strategy is 'replace_all'. This deletes and recreates all items on changes, which can break response mappings and integrations. Set update_strategy = \"targeted\" where possible, or set dangerously_replace_all_items = true to acknowledge this behavior.",
			)
		}

		km, d := r.updateReplaceAll(ctx, plan, state, currentForm)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		keyMap = km
	default:
		resp.Diagnostics.AddError("Invalid update_strategy", fmt.Sprintf("Unsupported update_strategy %q", updateStrategy))
		return
	}

	// Step 5: Handle publish settings changes.
	planPublished := plan.Published.ValueBool()
	planAccepting := plan.AcceptingResponses.ValueBool()
	statePublished := state.Published.ValueBool()
	stateAccepting := state.AcceptingResponses.ValueBool()

	if planPublished != statePublished || planAccepting != stateAccepting {
		tflog.Debug(ctx, "updating publish settings", map[string]interface{}{
			"published":           planPublished,
			"accepting_responses": planAccepting,
		})

		err := r.client.Forms.SetPublishSettings(ctx, formID, planPublished, planAccepting)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Setting Publish Settings",
				fmt.Sprintf("Could not update publish settings for form %s: %s", formID, err),
			)
			return
		}
	}

	// Step 6: Read back the final form state.
	finalForm, err := r.client.Forms.Get(ctx, formID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Google Form After Update",
			fmt.Sprintf("Could not read form %s after update: %s", formID, err),
		)
		return
	}

	// Step 7: Ensure keyMap is set for item correlation.
	// For targeted updates, it should include state mappings and any newly created items.
	// For replace_all, it should map newly created item IDs back to their plan item_keys.
	if keyMap == nil {
		keyMap = map[string]string{}
	}

	formModel, err := convert.FormToModel(finalForm, keyMap)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Converting Form Response",
			fmt.Sprintf("Could not convert API response for form %s: %s", formID, err),
		)
		return
	}

	if manageMode == "partial" {
		formModel.Items = filterItemsByKeyMap(formModel.Items, keyMap)
	}

	// Step 8: Convert to TF state and save.
	newState := convertFormModelToTFState(formModel, plan)

	if plan.ContentJSON.IsNull() || plan.ContentJSON.IsUnknown() || plan.ContentJSON.ValueString() == "" {
		itemList, diags := convertItemsToTFList(ctx, formModel.Items)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		newState.Items = itemList
	} else {
		newState.Items = plan.Items
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *FormResource) updateReplaceAll(
	ctx context.Context,
	plan FormResourceModel,
	state FormResourceModel,
	currentForm *forms.Form,
) (map[string]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	createdKeyMap := map[string]string{}

	conflictPolicy := "overwrite"
	if !plan.ConflictPolicy.IsNull() && !plan.ConflictPolicy.IsUnknown() && plan.ConflictPolicy.ValueString() != "" {
		conflictPolicy = plan.ConflictPolicy.ValueString()
	}
	if conflictPolicy == "fail" && !state.RevisionID.IsNull() && !state.RevisionID.IsUnknown() && state.RevisionID.ValueString() != "" {
		if currentForm != nil && currentForm.RevisionId != "" && currentForm.RevisionId != state.RevisionID.ValueString() {
			diags.AddError(
				"Update Conflict Detected",
				fmt.Sprintf("Form revision_id changed since last read (state=%s current=%s). Set conflict_policy = \"overwrite\" to force applying to the latest revision.", state.RevisionID.ValueString(), currentForm.RevisionId),
			)
			return nil, diags
		}
	} else if conflictPolicy == "fail" {
		diags.AddWarning(
			"Conflict Detection Not Available",
			"conflict_policy is \"fail\" but revision_id is not available in state. Proceeding without write control.",
		)
	}

	// Order: (1) UpdateFormInfo, (2) delete items, (3) quiz settings, (4) create items.
	// Item deletes MUST precede quiz settings changes so that disabling quiz
	// mode does not fail due to still-existing graded items.
	var requests []*forms.Request

	if currentForm == nil || currentForm.Info == nil ||
		currentForm.Info.Title != plan.Title.ValueString() ||
		currentForm.Info.Description != plan.Description.ValueString() {
		requests = append(requests, convert.BuildUpdateInfoRequest(
			plan.Title.ValueString(),
			plan.Description.ValueString(),
		))
	}

	existingItemCount := len(currentForm.Items)
	var createItemRequests []*forms.Request
	var createKeys []string

	if !plan.ContentJSON.IsNull() && !plan.ContentJSON.IsUnknown() && plan.ContentJSON.ValueString() != "" {
		tflog.Debug(ctx, "updating items via content_json mode")

		if existingItemCount > 0 {
			requests = append(requests, convert.BuildDeleteRequests(existingItemCount)...)
		}

		jsonRequests, err := convert.DeclarativeJSONToRequests(plan.ContentJSON.ValueString())
		if err != nil {
			diags.AddError("Error Parsing content_json", fmt.Sprintf("Could not parse content_json: %s", err))
			return nil, diags
		}
		createItemRequests = jsonRequests
	} else {
		convertItems, d := tfItemsToConvertItems(ctx, plan.Items)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		if len(convertItems) > 0 || existingItemCount > 0 {
			tflog.Debug(ctx, "replacing items", map[string]interface{}{
				"existing_count": existingItemCount,
				"new_count":      len(convertItems),
			})

			if existingItemCount > 0 {
				requests = append(requests, convert.BuildDeleteRequests(existingItemCount)...)
			}

			if len(convertItems) > 0 {
				// Collect item_keys in plan order so we can correlate CreateItem replies.
				var planItems []ItemModel
				diags.Append(plan.Items.ElementsAs(ctx, &planItems, false)...)
				if diags.HasError() {
					return nil, diags
				}
				createKeys = make([]string, len(planItems))
				for i := range planItems {
					createKeys[i] = planItems[i].ItemKey.ValueString()
				}

				itemRequests, err := convert.ItemsToCreateRequests(convertItems)
				if err != nil {
					diags.AddError("Error Building Item Requests", fmt.Sprintf("Could not build item create requests: %s", err))
					return nil, diags
				}
				createItemRequests = itemRequests
			}
		}
	}

	planQuiz := plan.Quiz.ValueBool()
	stateQuiz := state.Quiz.ValueBool()
	if planQuiz != stateQuiz {
		tflog.Debug(ctx, "updating quiz settings", map[string]interface{}{
			"quiz": planQuiz,
		})
		requests = append(requests, convert.BuildQuizSettingsRequest(planQuiz))
	}

	requests = append(requests, createItemRequests...)

	if len(requests) == 0 {
		return createdKeyMap, diags
	}

	tflog.Debug(ctx, "executing batchUpdate (replace_all)", map[string]interface{}{
		"request_count": len(requests),
	})

	batchReq := &forms.BatchUpdateFormRequest{
		Requests:              requests,
		IncludeFormInResponse: true,
	}
	if conflictPolicy == "fail" && !state.RevisionID.IsNull() && !state.RevisionID.IsUnknown() && state.RevisionID.ValueString() != "" {
		batchReq.WriteControl = &forms.WriteControl{RequiredRevisionId: state.RevisionID.ValueString()}
	}

	apiResp, err := r.client.Forms.BatchUpdate(ctx, state.ID.ValueString(), batchReq)
	if err != nil {
		diags.AddError("Error Updating Google Form", fmt.Sprintf("BatchUpdate failed for form %s: %s", state.ID.ValueString(), err))
		return nil, diags
	}

	// Build itemId -> item_key mapping for all created items (HCL item blocks only).
	if len(createKeys) > 0 && apiResp != nil {
		createReqIndexToKey, derr := buildCreateReqIndexToKey(requests, createKeys)
		if derr != nil {
			diags.AddWarning("Item Key Correlation Failed", derr.Error())
		} else {
			createdKeyMap, derr = extractCreateItemKeyMap(apiResp, requests, createReqIndexToKey)
			if derr != nil {
				diags.AddWarning("Item Key Correlation Failed", derr.Error())
			}
		}
	}

	return createdKeyMap, diags
}

func (r *FormResource) updateTargeted(
	ctx context.Context,
	plan FormResourceModel,
	state FormResourceModel,
	currentForm *forms.Form,
) (map[string]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	keyMap, d := buildItemKeyMap(ctx, state.Items)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	conflictPolicy := "overwrite"
	if !plan.ConflictPolicy.IsNull() && !plan.ConflictPolicy.IsUnknown() && plan.ConflictPolicy.ValueString() != "" {
		conflictPolicy = plan.ConflictPolicy.ValueString()
	}
	if conflictPolicy == "fail" && !state.RevisionID.IsNull() && !state.RevisionID.IsUnknown() && state.RevisionID.ValueString() != "" {
		if currentForm != nil && currentForm.RevisionId != "" && currentForm.RevisionId != state.RevisionID.ValueString() {
			diags.AddError(
				"Update Conflict Detected",
				fmt.Sprintf("Form revision_id changed since last read (state=%s current=%s). Set conflict_policy = \"overwrite\" to force applying to the latest revision.", state.RevisionID.ValueString(), currentForm.RevisionId),
			)
			return nil, diags
		}
	} else if conflictPolicy == "fail" {
		diags.AddWarning(
			"Conflict Detection Not Available",
			"conflict_policy is \"fail\" but revision_id is not available in state. Proceeding without write control.",
		)
	}

	manageMode := "all"
	if !plan.ManageMode.IsNull() && !plan.ManageMode.IsUnknown() && plan.ManageMode.ValueString() != "" {
		manageMode = plan.ManageMode.ValueString()
	}

	createdKeyMap := map[string]string{}

	if !plan.ContentJSON.IsNull() && !plan.ContentJSON.IsUnknown() && plan.ContentJSON.ValueString() != "" {
		diags.AddError("Targeted Updates Not Supported With content_json", "content_json mode cannot be updated in-place. Set update_strategy = \"replace_all\" or switch to item blocks.")
		return nil, diags
	}

	// Decode plan/state items to correlate item_keys and detect adds/removes/reorders.
	var planItems []ItemModel
	var stateItems []ItemModel

	if !plan.Items.IsNull() && !plan.Items.IsUnknown() {
		diags.Append(plan.Items.ElementsAs(ctx, &planItems, false)...)
	}
	if !state.Items.IsNull() && !state.Items.IsUnknown() {
		diags.Append(state.Items.ElementsAs(ctx, &stateItems, false)...)
	}
	if diags.HasError() {
		return nil, diags
	}

	desiredItems, d := tfItemsToConvertItems(ctx, plan.Items)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	// Index current items by itemId to find existing items quickly.
	type itemRef struct {
		index int
		item  *forms.Item
	}
	byID := make(map[string]itemRef, len(currentForm.Items))
	currentOrder := make([]string, 0, len(currentForm.Items))
	for i, it := range currentForm.Items {
		if it != nil && it.ItemId != "" {
			byID[it.ItemId] = itemRef{index: i, item: it}
			currentOrder = append(currentOrder, it.ItemId)
		}
	}

	var requests []*forms.Request
	if currentForm == nil || currentForm.Info == nil ||
		currentForm.Info.Title != plan.Title.ValueString() ||
		currentForm.Info.Description != plan.Description.ValueString() {
		requests = append(requests, convert.BuildUpdateInfoRequest(
			plan.Title.ValueString(),
			plan.Description.ValueString(),
		))
	}

	planQuiz := plan.Quiz.ValueBool()
	stateQuiz := state.Quiz.ValueBool()
	if planQuiz && !stateQuiz {
		// Enabling quiz: must set quiz before applying grading.
		requests = append(requests, convert.BuildQuizSettingsRequest(true))
	}

	// Build state item_key -> google_item_id map for existing items.
	stateKeyToID := make(map[string]string, len(stateItems))
	for _, it := range stateItems {
		key := it.ItemKey.ValueString()
		gid := it.GoogleItemID.ValueString()
		if key != "" && gid != "" {
			stateKeyToID[key] = gid
		}
	}

	planKeySeen := make(map[string]bool, len(planItems))
	planExistingIDs := make([]string, 0, len(planItems))
	planNewIndices := make([]int, 0)
	for i, it := range planItems {
		key := it.ItemKey.ValueString()
		planKeySeen[key] = true
		if gid, ok := stateKeyToID[key]; ok {
			if _, exists := byID[gid]; exists {
				planExistingIDs = append(planExistingIDs, gid)
				continue
			}
			if manageMode == "partial" {
				// Previously-managed item was deleted out-of-band; treat as new.
				delete(stateKeyToID, key)
				planNewIndices = append(planNewIndices, i)
				continue
			}

			diags.AddError(
				"Targeted Update Failed",
				fmt.Sprintf("State tracked item %q (google_item_id=%s) not found in current form. Switch to update_strategy = \"replace_all\".", key, gid),
			)
			return nil, diags
		}
		planNewIndices = append(planNewIndices, i)
	}

	// Validate that all items in the current form are tracked by state (required for safe moves/deletes).
	if manageMode == "all" {
		for _, id := range currentOrder {
			if _, ok := keyMap[id]; !ok {
				diags.AddError(
					"Targeted Update Requires Full Item State",
					"One or more existing items are not present in Terraform state (missing google_item_id mapping). Import the form or switch to update_strategy = \"replace_all\".",
				)
				return nil, diags
			}
		}
	}

	// Step A: delete items that exist in state but are removed from plan.
	deleteIndices := make([]int, 0)
	for _, st := range stateItems {
		key := st.ItemKey.ValueString()
		if key == "" {
			continue
		}
		if !planKeySeen[key] {
			gid := st.GoogleItemID.ValueString()
			ref, ok := byID[gid]
			if !ok {
				continue
			}
			deleteIndices = append(deleteIndices, ref.index)
		}
	}
	// delete in reverse index order
	for i := 0; i < len(deleteIndices); i++ {
		for j := i + 1; j < len(deleteIndices); j++ {
			if deleteIndices[i] < deleteIndices[j] {
				deleteIndices[i], deleteIndices[j] = deleteIndices[j], deleteIndices[i]
			}
		}
	}
	for _, idx := range deleteIndices {
		requests = append(requests, &forms.Request{DeleteItem: &forms.DeleteItemRequest{Location: &forms.Location{Index: int64(idx)}}})
		currentOrder = append(currentOrder[:idx], currentOrder[idx+1:]...)
	}

	// Step B: reorder existing items (ignoring new items) to match plan order.
	if manageMode == "all" {
		if len(currentOrder) != len(planExistingIDs) {
			diags.AddError(
				"Targeted Update Structural Mismatch",
				"After applying deletions, the remaining item count does not match the number of existing items in the plan. Switch to update_strategy = \"replace_all\".",
			)
			return nil, diags
		}

		for targetIdx, wantID := range planExistingIDs {
			curIdx := indexOf(currentOrder, wantID)
			if curIdx < 0 {
				diags.AddError(
					"Targeted Update Failed To Locate Existing Item",
					fmt.Sprintf("Could not find existing item ID %s in the current form. Switch to update_strategy = \"replace_all\".", wantID),
				)
				return nil, diags
			}
			if curIdx != targetIdx {
				requests = append(requests, &forms.Request{
					MoveItem: &forms.MoveItemRequest{
						OriginalLocation: &forms.Location{Index: int64(curIdx)},
						NewLocation:      &forms.Location{Index: int64(targetIdx)},
					},
				})
				// simulate move
				id := currentOrder[curIdx]
				currentOrder = append(currentOrder[:curIdx], currentOrder[curIdx+1:]...)
				before := currentOrder[:targetIdx]
				after := currentOrder[targetIdx:]
				currentOrder = append(append(append([]string{}, before...), id), after...)
			}
		}
	} else {
		// partial: permute managed items within their existing slots so that
		// unmanaged items are left in their original order/positions.
		managedSet := make(map[string]bool, len(planExistingIDs))
		for _, gid := range planExistingIDs {
			managedSet[gid] = true
		}

		managedSlots := make([]int, 0, len(planExistingIDs))
		for idx, id := range currentOrder {
			if managedSet[id] {
				managedSlots = append(managedSlots, idx)
			}
		}
		if len(managedSlots) != len(planExistingIDs) {
			diags.AddError(
				"Targeted Update Failed",
				"Could not compute managed item slots for partial management mode. Switch to update_strategy = \"replace_all\" or set manage_mode = \"all\".",
			)
			return nil, diags
		}

		desiredOrder := append([]string{}, currentOrder...)
		for i, slot := range managedSlots {
			desiredOrder[slot] = planExistingIDs[i]
		}

		for targetIdx, wantID := range desiredOrder {
			curIdx := indexOf(currentOrder, wantID)
			if curIdx < 0 {
				diags.AddError(
					"Targeted Update Failed To Locate Existing Item",
					fmt.Sprintf("Could not find existing item ID %s in the current form. Switch to update_strategy = \"replace_all\".", wantID),
				)
				return nil, diags
			}
			if curIdx != targetIdx {
				requests = append(requests, &forms.Request{
					MoveItem: &forms.MoveItemRequest{
						OriginalLocation: &forms.Location{Index: int64(curIdx)},
						NewLocation:      &forms.Location{Index: int64(targetIdx)},
					},
				})
				// simulate move
				id := currentOrder[curIdx]
				currentOrder = append(currentOrder[:curIdx], currentOrder[curIdx+1:]...)
				before := currentOrder[:targetIdx]
				after := currentOrder[targetIdx:]
				currentOrder = append(append(append([]string{}, before...), id), after...)
			}
		}
	}

	// Step C: update existing items in-place.
	for i, pi := range planItems {
		key := pi.ItemKey.ValueString()
		gid, ok := stateKeyToID[key]
		if !ok {
			continue
		}
		ref, ok := byID[gid]
		if !ok || ref.item == nil {
			diags.AddError("Targeted Update Failed", fmt.Sprintf("Could not find API item %s for %q", gid, key))
			return nil, diags
		}
		idx := indexOf(currentOrder, gid)
		if idx < 0 {
			diags.AddError("Targeted Update Failed", fmt.Sprintf("Could not find item %s in current order for %q", gid, key))
			return nil, diags
		}

		updated, changed, needsReplace, err := convert.ApplyDesiredItem(ref.item, desiredItems[i])
		if err != nil {
			diags.AddError("Targeted Update Failed", err.Error())
			return nil, diags
		}
		if needsReplace {
			diags.AddError(
				"Targeted Update Requires Replace-All",
				fmt.Sprintf("Item %q requires a structural change (e.g. question type change) and cannot be updated in-place. Switch to update_strategy = \"replace_all\".", key),
			)
			return nil, diags
		}

		if changed {
			requests = append(requests, &forms.Request{
				UpdateItem: &forms.UpdateItemRequest{
					Item:       updated,
					Location:   &forms.Location{Index: int64(idx)},
					UpdateMask: "*",
				},
			})
		}
	}

	if !planQuiz && stateQuiz {
		// Disabling quiz: must clear grading first, then disable quiz.
		requests = append(requests, convert.BuildQuizSettingsRequest(false))
	}

	// Step D: create new items at their intended indices (in plan order).
	createReqIndexToKey := make(map[int]string)
	for i := range planItems {
		key := planItems[i].ItemKey.ValueString()
		if _, ok := stateKeyToID[key]; ok {
			continue
		}
		// New item: in partial mode, append by default to avoid shifting unmanaged items.
		insertIdx := i
		if manageMode == "partial" {
			insertIdx = len(currentOrder)
		}
		req, err := convert.ItemModelToCreateRequest(desiredItems[i], insertIdx)
		if err != nil {
			diags.AddError("Error Building Item Requests", err.Error())
			return nil, diags
		}
		createReqIndexToKey[len(requests)] = key
		requests = append(requests, req)
		// simulate insert so subsequent create indices line up
		currentOrder = append(currentOrder[:insertIdx], append([]string{"__new__"}, currentOrder[insertIdx:]...)...)
	}

	if len(requests) == 0 {
		return keyMap, diags
	}

	tflog.Debug(ctx, "executing batchUpdate (targeted)", map[string]interface{}{
		"request_count": len(requests),
	})

	batchReq := &forms.BatchUpdateFormRequest{
		Requests:              requests,
		IncludeFormInResponse: true,
	}
	if conflictPolicy == "fail" && !state.RevisionID.IsNull() && !state.RevisionID.IsUnknown() && state.RevisionID.ValueString() != "" {
		batchReq.WriteControl = &forms.WriteControl{RequiredRevisionId: state.RevisionID.ValueString()}
	}

	apiResp, err := r.client.Forms.BatchUpdate(ctx, state.ID.ValueString(), batchReq)
	if err != nil {
		diags.AddError("Error Updating Google Form", fmt.Sprintf("BatchUpdate failed for form %s: %s", state.ID.ValueString(), err))
		return nil, diags
	}

	createdKeyMap, derr := extractCreateItemKeyMap(apiResp, requests, createReqIndexToKey)
	if derr != nil {
		diags.AddWarning("Item Key Correlation Failed", derr.Error())
	} else {
		for gid, k := range createdKeyMap {
			keyMap[gid] = k
		}
	}

	return keyMap, diags
}

func extractCreateItemKeyMap(
	resp *forms.BatchUpdateFormResponse,
	requests []*forms.Request,
	requestIndexToKey map[int]string,
) (map[string]string, error) {
	out := make(map[string]string)
	if resp == nil {
		return out, fmt.Errorf("nil BatchUpdateFormResponse")
	}
	if len(resp.Replies) != len(requests) {
		return out, fmt.Errorf("reply count mismatch: got %d replies for %d requests", len(resp.Replies), len(requests))
	}

	for i := range requests {
		key, ok := requestIndexToKey[i]
		if !ok {
			continue
		}
		r := requests[i]
		if r.CreateItem == nil {
			continue
		}
		rep := resp.Replies[i]
		if rep == nil || rep.CreateItem == nil {
			return out, fmt.Errorf("missing createItem reply for request[%d]", i)
		}
		itemID := rep.CreateItem.ItemId
		if strings.TrimSpace(itemID) == "" {
			return out, fmt.Errorf("empty itemId in createItem reply for request[%d]", i)
		}
		out[itemID] = key
	}

	return out, nil
}

func buildCreateReqIndexToKey(requests []*forms.Request, createKeys []string) (map[int]string, error) {
	out := make(map[int]string)
	next := 0
	for i := range requests {
		if requests[i] == nil || requests[i].CreateItem == nil {
			continue
		}
		if next >= len(createKeys) {
			return nil, fmt.Errorf("more CreateItem requests than keys: want %d, got extra at request[%d]", len(createKeys), i)
		}
		out[i] = createKeys[next]
		next++
	}
	if next != len(createKeys) {
		return nil, fmt.Errorf("key count mismatch: expected %d CreateItem requests, saw %d", len(createKeys), next)
	}
	return out, nil
}

func indexOf(s []string, want string) int {
	for i := range s {
		if s[i] == want {
			return i
		}
	}
	return -1
}
