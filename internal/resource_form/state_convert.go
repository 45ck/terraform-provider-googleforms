// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceform

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/45ck/terraform-provider-googleforms/internal/convert"
)

// tfItemsToConvertItems extracts items from the Terraform plan's types.List
// and converts them to []convert.ItemModel for use by the convert package.
func tfItemsToConvertItems(ctx context.Context, items types.List) ([]convert.ItemModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	if items.IsNull() || items.IsUnknown() || len(items.Elements()) == 0 {
		return nil, diags
	}

	var tfItems []ItemModel
	diags.Append(items.ElementsAs(ctx, &tfItems, false)...)
	if diags.HasError() {
		return nil, diags
	}

	result := make([]convert.ItemModel, len(tfItems))
	for i, tf := range tfItems {
		result[i] = convert.ItemModel{
			ItemKey:      tf.ItemKey.ValueString(),
			GoogleItemID: tf.GoogleItemID.ValueString(),
		}

		if tf.MultipleChoice != nil {
			mc := &convert.MultipleChoiceBlock{
				QuestionText: tf.MultipleChoice.QuestionText.ValueString(),
				Required:     tf.MultipleChoice.Required.ValueBool(),
				Shuffle:      tf.MultipleChoice.Shuffle.ValueBool(),
				HasOther:     tf.MultipleChoice.HasOther.ValueBool(),
			}

			opts, d := tfChoiceOptionsToConvert(ctx, tf.MultipleChoice.Options, tf.MultipleChoice.Option)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			mc.Options = opts

			if tf.MultipleChoice.Grading != nil {
				mc.Grading = tfGradingToConvert(tf.MultipleChoice.Grading)
			}

			result[i].MultipleChoice = mc
			result[i].Title = mc.QuestionText
		}

		if tf.ShortAnswer != nil {
			sa := &convert.ShortAnswerBlock{
				QuestionText: tf.ShortAnswer.QuestionText.ValueString(),
				Required:     tf.ShortAnswer.Required.ValueBool(),
			}
			if tf.ShortAnswer.Grading != nil {
				sa.Grading = tfGradingToConvert(tf.ShortAnswer.Grading)
			}
			result[i].ShortAnswer = sa
			result[i].Title = sa.QuestionText
		}

		if tf.Paragraph != nil {
			p := &convert.ParagraphBlock{
				QuestionText: tf.Paragraph.QuestionText.ValueString(),
				Required:     tf.Paragraph.Required.ValueBool(),
			}
			if tf.Paragraph.Grading != nil {
				p.Grading = tfGradingToConvert(tf.Paragraph.Grading)
			}
			result[i].Paragraph = p
			result[i].Title = p.QuestionText
		}

		if tf.Dropdown != nil {
			dd := &convert.DropdownBlock{
				QuestionText: tf.Dropdown.QuestionText.ValueString(),
				Required:     tf.Dropdown.Required.ValueBool(),
				Shuffle:      tf.Dropdown.Shuffle.ValueBool(),
			}
			opts, d := tfChoiceOptionsToConvert(ctx, tf.Dropdown.Options, tf.Dropdown.Option)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			dd.Options = opts
			if tf.Dropdown.Grading != nil {
				dd.Grading = tfGradingToConvert(tf.Dropdown.Grading)
			}
			result[i].Dropdown = dd
			result[i].Title = dd.QuestionText
		}

		if tf.Checkbox != nil {
			cb := &convert.CheckboxBlock{
				QuestionText: tf.Checkbox.QuestionText.ValueString(),
				Required:     tf.Checkbox.Required.ValueBool(),
				Shuffle:      tf.Checkbox.Shuffle.ValueBool(),
				HasOther:     tf.Checkbox.HasOther.ValueBool(),
			}
			opts, d := tfChoiceOptionsToConvert(ctx, tf.Checkbox.Options, tf.Checkbox.Option)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			cb.Options = opts
			if tf.Checkbox.Grading != nil {
				cb.Grading = tfGradingToConvert(tf.Checkbox.Grading)
			}
			result[i].Checkbox = cb
			result[i].Title = cb.QuestionText
		}

		if tf.MultipleChoiceGrid != nil {
			var rows []string
			diags.Append(tf.MultipleChoiceGrid.Rows.ElementsAs(ctx, &rows, false)...)
			if diags.HasError() {
				return nil, diags
			}
			var cols []string
			diags.Append(tf.MultipleChoiceGrid.Columns.ElementsAs(ctx, &cols, false)...)
			if diags.HasError() {
				return nil, diags
			}
			result[i].MultipleChoiceGrid = &convert.MultipleChoiceGridBlock{
				QuestionText:     tf.MultipleChoiceGrid.QuestionText.ValueString(),
				Rows:             rows,
				Columns:          cols,
				Required:         tf.MultipleChoiceGrid.Required.ValueBool(),
				ShuffleQuestions: tf.MultipleChoiceGrid.ShuffleQuestions.ValueBool(),
				ShuffleColumns:   tf.MultipleChoiceGrid.ShuffleColumns.ValueBool(),
			}
			result[i].Title = tf.MultipleChoiceGrid.QuestionText.ValueString()
		}

		if tf.CheckboxGrid != nil {
			var rows []string
			diags.Append(tf.CheckboxGrid.Rows.ElementsAs(ctx, &rows, false)...)
			if diags.HasError() {
				return nil, diags
			}
			var cols []string
			diags.Append(tf.CheckboxGrid.Columns.ElementsAs(ctx, &cols, false)...)
			if diags.HasError() {
				return nil, diags
			}
			result[i].CheckboxGrid = &convert.CheckboxGridBlock{
				QuestionText:     tf.CheckboxGrid.QuestionText.ValueString(),
				Rows:             rows,
				Columns:          cols,
				Required:         tf.CheckboxGrid.Required.ValueBool(),
				ShuffleQuestions: tf.CheckboxGrid.ShuffleQuestions.ValueBool(),
				ShuffleColumns:   tf.CheckboxGrid.ShuffleColumns.ValueBool(),
			}
			result[i].Title = tf.CheckboxGrid.QuestionText.ValueString()
		}

		if tf.Date != nil {
			result[i].Date = &convert.DateBlock{
				QuestionText: tf.Date.QuestionText.ValueString(),
				Required:     tf.Date.Required.ValueBool(),
				IncludeYear:  tf.Date.IncludeYear.ValueBool(),
			}
			result[i].Title = tf.Date.QuestionText.ValueString()
		}

		if tf.DateTime != nil {
			result[i].DateTime = &convert.DateTimeBlock{
				QuestionText: tf.DateTime.QuestionText.ValueString(),
				Required:     tf.DateTime.Required.ValueBool(),
				IncludeYear:  tf.DateTime.IncludeYear.ValueBool(),
			}
			result[i].Title = tf.DateTime.QuestionText.ValueString()
		}

		if tf.Scale != nil {
			result[i].Scale = &convert.ScaleBlock{
				QuestionText: tf.Scale.QuestionText.ValueString(),
				Required:     tf.Scale.Required.ValueBool(),
				Low:          tf.Scale.Low.ValueInt64(),
				High:         tf.Scale.High.ValueInt64(),
				LowLabel:     tf.Scale.LowLabel.ValueString(),
				HighLabel:    tf.Scale.HighLabel.ValueString(),
			}
			result[i].Title = tf.Scale.QuestionText.ValueString()
		}

		if tf.Time != nil {
			result[i].Time = &convert.TimeBlock{
				QuestionText: tf.Time.QuestionText.ValueString(),
				Required:     tf.Time.Required.ValueBool(),
				Duration:     tf.Time.Duration.ValueBool(),
			}
			result[i].Title = tf.Time.QuestionText.ValueString()
		}

		if tf.Rating != nil {
			result[i].Rating = &convert.RatingBlock{
				QuestionText:     tf.Rating.QuestionText.ValueString(),
				Required:         tf.Rating.Required.ValueBool(),
				IconType:         tf.Rating.IconType.ValueString(),
				RatingScaleLevel: tf.Rating.RatingScaleLevel.ValueInt64(),
			}
			result[i].Title = tf.Rating.QuestionText.ValueString()
		}

		if tf.FileUpload != nil {
			// Creation of file upload questions is blocked in the convert layer.
			// This is primarily for representing imported/existing items.
			var typesList []string
			if !tf.FileUpload.Types.IsNull() && !tf.FileUpload.Types.IsUnknown() {
				diags.Append(tf.FileUpload.Types.ElementsAs(ctx, &typesList, false)...)
				if diags.HasError() {
					return nil, diags
				}
			}
			result[i].FileUpload = &convert.FileUploadBlock{
				QuestionText: tf.FileUpload.QuestionText.ValueString(),
				Required:     tf.FileUpload.Required.ValueBool(),
				FolderID:     tf.FileUpload.FolderID.ValueString(),
				MaxFileSize:  tf.FileUpload.MaxFileSize.ValueInt64(),
				MaxFiles:     tf.FileUpload.MaxFiles.ValueInt64(),
				Types:        typesList,
			}
			result[i].Title = tf.FileUpload.QuestionText.ValueString()
		}

		if tf.TextItem != nil {
			result[i].TextItem = &convert.TextItemBlock{
				Title:       tf.TextItem.Title.ValueString(),
				Description: tf.TextItem.Description.ValueString(),
			}
			result[i].Title = tf.TextItem.Title.ValueString()
		}

		if tf.Image != nil {
			result[i].Image = &convert.ImageBlock{
				Title:       tf.Image.Title.ValueString(),
				Description: tf.Image.Description.ValueString(),
				SourceURI:   tf.Image.SourceURI.ValueString(),
				AltText:     tf.Image.AltText.ValueString(),
				ContentURI:  tf.Image.ContentURI.ValueString(),
			}
			result[i].Title = tf.Image.Title.ValueString()
		}

		if tf.Video != nil {
			result[i].Video = &convert.VideoBlock{
				Title:       tf.Video.Title.ValueString(),
				Description: tf.Video.Description.ValueString(),
				YoutubeURI:  tf.Video.YoutubeURI.ValueString(),
				Caption:     tf.Video.Caption.ValueString(),
			}
			result[i].Title = tf.Video.Title.ValueString()
		}

		if tf.SectionHeader != nil {
			result[i].SectionHeader = &convert.SectionHeaderBlock{
				Title:       tf.SectionHeader.Title.ValueString(),
				Description: tf.SectionHeader.Description.ValueString(),
			}
			result[i].Title = tf.SectionHeader.Title.ValueString()
		}
	}

	return result, diags
}

func tfChoiceOptionsToConvert(ctx context.Context, options types.List, optionBlocks types.List) ([]convert.ChoiceOption, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Prefer option blocks when present.
	if !optionBlocks.IsNull() && !optionBlocks.IsUnknown() && len(optionBlocks.Elements()) > 0 {
		var tfOpts []ChoiceOptionModel
		diags.Append(optionBlocks.ElementsAs(ctx, &tfOpts, false)...)
		if diags.HasError() {
			return nil, diags
		}

		out := make([]convert.ChoiceOption, 0, len(tfOpts))
		for _, o := range tfOpts {
			out = append(out, convert.ChoiceOption{
				Value:          o.Value.ValueString(),
				GoToAction:     o.GoToAction.ValueString(),
				GoToSectionKey: o.GoToSectionKey.ValueString(),
				GoToSectionID:  o.GoToSectionID.ValueString(),
			})
		}
		return out, diags
	}

	// Backward-compatible path: plain string options list.
	if options.IsNull() || options.IsUnknown() || len(options.Elements()) == 0 {
		return nil, diags
	}

	var vals []string
	diags.Append(options.ElementsAs(ctx, &vals, false)...)
	if diags.HasError() {
		return nil, diags
	}
	out := make([]convert.ChoiceOption, 0, len(vals))
	for _, v := range vals {
		out = append(out, convert.ChoiceOption{Value: v})
	}
	return out, diags
}

func convertChoiceOptionsToTF(ctx context.Context, opts []convert.ChoiceOption) (types.List, types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	// If any option has navigation config, represent in option blocks.
	needsBlocks := false
	for _, o := range opts {
		if o.GoToAction != "" || o.GoToSectionID != "" || o.GoToSectionKey != "" {
			needsBlocks = true
			break
		}
	}

	if !needsBlocks {
		values := make([]string, 0, len(opts))
		for _, o := range opts {
			values = append(values, o.Value)
		}
		lv, d := types.ListValueFrom(ctx, types.StringType, values)
		diags.Append(d...)
		return lv, types.ListNull(types.ObjectType{AttrTypes: choiceOptionAttrTypes()}), diags
	}

	// Use option blocks. Keep legacy options list null to satisfy mutual exclusion validators.
	tfOpts := make([]ChoiceOptionModel, 0, len(opts))
	for _, o := range opts {
		tfOpt := ChoiceOptionModel{
			Value:          types.StringValue(o.Value),
			GoToAction:     types.StringNull(),
			GoToSectionKey: types.StringNull(),
			GoToSectionID:  types.StringNull(),
		}
		if o.GoToAction != "" {
			tfOpt.GoToAction = types.StringValue(o.GoToAction)
		}
		if o.GoToSectionKey != "" {
			tfOpt.GoToSectionKey = types.StringValue(o.GoToSectionKey)
		}
		if o.GoToSectionID != "" {
			tfOpt.GoToSectionID = types.StringValue(o.GoToSectionID)
		}
		tfOpts = append(tfOpts, tfOpt)
	}
	ov, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: choiceOptionAttrTypes()}, tfOpts)
	diags.Append(d...)
	return types.ListNull(types.StringType), ov, diags
}

func choiceOptionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"value":             types.StringType,
		"go_to_action":      types.StringType,
		"go_to_section_key": types.StringType,
		"go_to_section_id":  types.StringType,
	}
}

// tfGradingToConvert converts a TF GradingModel to a convert.GradingBlock.
func tfGradingToConvert(g *GradingModel) *convert.GradingBlock {
	return &convert.GradingBlock{
		Points:            g.Points.ValueInt64(),
		CorrectAnswer:     g.CorrectAnswer.ValueString(),
		FeedbackCorrect:   g.FeedbackCorrect.ValueString(),
		FeedbackIncorrect: g.FeedbackIncorrect.ValueString(),
	}
}

// convertFormModelToTFState maps a convert.FormModel back into a
// FormResourceModel, preserving plan values for non-computed fields.
func convertFormModelToTFState(model *convert.FormModel, plan FormResourceModel) FormResourceModel {
	manageMode := plan.ManageMode
	if manageMode.IsNull() || manageMode.IsUnknown() || manageMode.ValueString() == "" {
		manageMode = types.StringValue("all")
	}
	partialNewItemPolicy := plan.PartialNewItemPolicy
	if partialNewItemPolicy.IsNull() || partialNewItemPolicy.IsUnknown() || partialNewItemPolicy.ValueString() == "" {
		partialNewItemPolicy = types.StringValue("append")
	}
	conflictPolicy := plan.ConflictPolicy
	if conflictPolicy.IsNull() || conflictPolicy.IsUnknown() || conflictPolicy.ValueString() == "" {
		conflictPolicy = types.StringValue("overwrite")
	}
	supportsAllDrives := plan.SupportsAllDrives
	if supportsAllDrives.IsNull() || supportsAllDrives.IsUnknown() {
		supportsAllDrives = types.BoolValue(false)
	}

	state := FormResourceModel{
		ID:                 types.StringValue(model.ID),
		Title:              types.StringValue(model.Title),
		Description:        types.StringValue(model.Description),
		Published:          plan.Published,
		AcceptingResponses: plan.AcceptingResponses,
		Quiz:               types.BoolValue(model.Quiz),
		EmailCollectionType: func() types.String {
			if model.EmailCollectionType == "" {
				return types.StringNull()
			}
			return types.StringValue(model.EmailCollectionType)
		}(),
		UpdateStrategy:       plan.UpdateStrategy,
		DangerousReplaceAll:  plan.DangerousReplaceAll,
		ManageMode:           manageMode,
		PartialNewItemPolicy: partialNewItemPolicy,
		ConflictPolicy:       conflictPolicy,
		FolderID:             plan.FolderID,
		SupportsAllDrives:    supportsAllDrives,
		ParentIDs:            plan.ParentIDs,
		ContentJSON:          plan.ContentJSON,
		ResponderURI:         types.StringValue(model.ResponderURI),
		DocumentTitle:        types.StringValue(model.DocumentTitle),
		RevisionID:           types.StringValue(model.RevisionID),
	}

	// edit_uri follows a known pattern
	if model.ID != "" {
		state.EditURI = types.StringValue("https://docs.google.com/forms/d/" + model.ID + "/edit")
	}

	return state
}

// buildItemKeyMap creates a mapping from google_item_id -> item_key from
// current state items, enabling correlation of API items back to TF config.
func buildItemKeyMap(ctx context.Context, items types.List) (map[string]string, diag.Diagnostics) {
	var diags diag.Diagnostics

	if items.IsNull() || items.IsUnknown() || len(items.Elements()) == 0 {
		return nil, diags
	}

	var tfItems []ItemModel
	diags.Append(items.ElementsAs(ctx, &tfItems, false)...)
	if diags.HasError() {
		return nil, diags
	}

	keyMap := make(map[string]string, len(tfItems))
	for _, item := range tfItems {
		googleID := item.GoogleItemID.ValueString()
		itemKey := item.ItemKey.ValueString()
		if googleID != "" && itemKey != "" {
			keyMap[googleID] = itemKey
		}
	}

	return keyMap, diags
}

// convertItemsToTFList converts []convert.ItemModel back to types.List
// for setting in Terraform state.
func convertItemsToTFList(ctx context.Context, items []convert.ItemModel) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(items) == 0 {
		return types.ListNull(itemObjectType()), diags
	}

	tfItems := make([]ItemModel, len(items))
	for i, item := range items {
		tfItems[i] = convertItemModelToTF(ctx, item, &diags)
		if diags.HasError() {
			return types.ListNull(itemObjectType()), diags
		}
	}

	result, d := types.ListValueFrom(ctx, itemObjectType(), tfItems)
	diags.Append(d...)
	return result, diags
}

// convertItemModelToTF converts a single convert.ItemModel to the TF ItemModel.
func convertItemModelToTF(ctx context.Context, item convert.ItemModel, diags *diag.Diagnostics) ItemModel {
	tf := ItemModel{
		ItemKey:      types.StringValue(item.ItemKey),
		GoogleItemID: types.StringValue(item.GoogleItemID),
	}

	if item.MultipleChoice != nil {
		mc := item.MultipleChoice
		optsList, optionBlocks, d := convertChoiceOptionsToTF(ctx, mc.Options)
		diags.Append(d...)

		tf.MultipleChoice = &MultipleChoiceModel{
			QuestionText: types.StringValue(mc.QuestionText),
			Options:      optsList,
			Option:       optionBlocks,
			Required:     types.BoolValue(mc.Required),
			Shuffle:      types.BoolValue(mc.Shuffle),
			HasOther:     types.BoolValue(mc.HasOther),
		}
		if mc.Grading != nil {
			tf.MultipleChoice.Grading = convertGradingToTF(mc.Grading)
		}
	}

	if item.ShortAnswer != nil {
		sa := item.ShortAnswer
		tf.ShortAnswer = &ShortAnswerModel{
			QuestionText: types.StringValue(sa.QuestionText),
			Required:     types.BoolValue(sa.Required),
		}
		if sa.Grading != nil {
			tf.ShortAnswer.Grading = convertGradingToTF(sa.Grading)
		}
	}

	if item.Paragraph != nil {
		p := item.Paragraph
		tf.Paragraph = &ParagraphModel{
			QuestionText: types.StringValue(p.QuestionText),
			Required:     types.BoolValue(p.Required),
		}
		if p.Grading != nil {
			tf.Paragraph.Grading = convertGradingToTF(p.Grading)
		}
	}

	if item.Dropdown != nil {
		dd := item.Dropdown
		optsList, optionBlocks, d := convertChoiceOptionsToTF(ctx, dd.Options)
		diags.Append(d...)
		tf.Dropdown = &DropdownModel{
			QuestionText: types.StringValue(dd.QuestionText),
			Options:      optsList,
			Option:       optionBlocks,
			Required:     types.BoolValue(dd.Required),
			Shuffle:      types.BoolValue(dd.Shuffle),
		}
		if dd.Grading != nil {
			tf.Dropdown.Grading = convertGradingToTF(dd.Grading)
		}
	}

	if item.Checkbox != nil {
		cb := item.Checkbox
		optsList, optionBlocks, d := convertChoiceOptionsToTF(ctx, cb.Options)
		diags.Append(d...)
		tf.Checkbox = &CheckboxModel{
			QuestionText: types.StringValue(cb.QuestionText),
			Options:      optsList,
			Option:       optionBlocks,
			Required:     types.BoolValue(cb.Required),
			Shuffle:      types.BoolValue(cb.Shuffle),
			HasOther:     types.BoolValue(cb.HasOther),
		}
		if cb.Grading != nil {
			tf.Checkbox.Grading = convertGradingToTF(cb.Grading)
		}
	}

	if item.MultipleChoiceGrid != nil {
		g := item.MultipleChoiceGrid
		rows, d := types.ListValueFrom(ctx, types.StringType, g.Rows)
		diags.Append(d...)
		cols, d := types.ListValueFrom(ctx, types.StringType, g.Columns)
		diags.Append(d...)
		tf.MultipleChoiceGrid = &MultipleChoiceGridModel{
			QuestionText:     types.StringValue(g.QuestionText),
			Rows:             rows,
			Columns:          cols,
			Required:         types.BoolValue(g.Required),
			ShuffleQuestions: types.BoolValue(g.ShuffleQuestions),
			ShuffleColumns:   types.BoolValue(g.ShuffleColumns),
		}
	}

	if item.CheckboxGrid != nil {
		g := item.CheckboxGrid
		rows, d := types.ListValueFrom(ctx, types.StringType, g.Rows)
		diags.Append(d...)
		cols, d := types.ListValueFrom(ctx, types.StringType, g.Columns)
		diags.Append(d...)
		tf.CheckboxGrid = &CheckboxGridModel{
			QuestionText:     types.StringValue(g.QuestionText),
			Rows:             rows,
			Columns:          cols,
			Required:         types.BoolValue(g.Required),
			ShuffleQuestions: types.BoolValue(g.ShuffleQuestions),
			ShuffleColumns:   types.BoolValue(g.ShuffleColumns),
		}
	}
	if item.Date != nil {
		tf.Date = &DateModel{
			QuestionText: types.StringValue(item.Date.QuestionText),
			Required:     types.BoolValue(item.Date.Required),
			IncludeYear:  types.BoolValue(item.Date.IncludeYear),
		}
	}

	if item.DateTime != nil {
		tf.DateTime = &DateTimeModel{
			QuestionText: types.StringValue(item.DateTime.QuestionText),
			Required:     types.BoolValue(item.DateTime.Required),
			IncludeYear:  types.BoolValue(item.DateTime.IncludeYear),
		}
	}

	if item.Scale != nil {
		s := item.Scale
		tf.Scale = &ScaleModel{
			QuestionText: types.StringValue(s.QuestionText),
			Required:     types.BoolValue(s.Required),
			Low:          types.Int64Value(s.Low),
			High:         types.Int64Value(s.High),
		}
		if s.LowLabel != "" {
			tf.Scale.LowLabel = types.StringValue(s.LowLabel)
		} else {
			tf.Scale.LowLabel = types.StringNull()
		}
		if s.HighLabel != "" {
			tf.Scale.HighLabel = types.StringValue(s.HighLabel)
		} else {
			tf.Scale.HighLabel = types.StringNull()
		}
	}

	if item.Time != nil {
		t := item.Time
		tf.Time = &TimeModel{
			QuestionText: types.StringValue(t.QuestionText),
			Required:     types.BoolValue(t.Required),
			Duration:     types.BoolValue(t.Duration),
		}
	}

	if item.Rating != nil {
		r := item.Rating
		tf.Rating = &RatingModel{
			QuestionText:     types.StringValue(r.QuestionText),
			Required:         types.BoolValue(r.Required),
			IconType:         types.StringValue(r.IconType),
			RatingScaleLevel: types.Int64Value(r.RatingScaleLevel),
		}
	}

	if item.FileUpload != nil {
		u := item.FileUpload
		typesList, d := types.ListValueFrom(ctx, types.StringType, u.Types)
		diags.Append(d...)
		tf.FileUpload = &FileUploadModel{
			QuestionText: types.StringValue(u.QuestionText),
			Required:     types.BoolValue(u.Required),
			FolderID:     types.StringValue(u.FolderID),
			MaxFileSize:  types.Int64Value(u.MaxFileSize),
			MaxFiles:     types.Int64Value(u.MaxFiles),
			Types:        typesList,
		}
	}
	if item.TextItem != nil {
		ti := item.TextItem
		tf.TextItem = &TextItemModel{
			Title: types.StringValue(ti.Title),
		}
		if ti.Description != "" {
			tf.TextItem.Description = types.StringValue(ti.Description)
		} else {
			tf.TextItem.Description = types.StringNull()
		}
	}

	if item.Image != nil {
		img := item.Image
		tf.Image = &ImageModel{
			SourceURI:  types.StringValue(img.SourceURI),
			AltText:    types.StringValue(img.AltText),
			ContentURI: types.StringValue(img.ContentURI),
		}
		if img.Title != "" {
			tf.Image.Title = types.StringValue(img.Title)
		} else {
			tf.Image.Title = types.StringNull()
		}
		if img.Description != "" {
			tf.Image.Description = types.StringValue(img.Description)
		} else {
			tf.Image.Description = types.StringNull()
		}
	}

	if item.Video != nil {
		v := item.Video
		tf.Video = &VideoModel{
			YoutubeURI: types.StringValue(v.YoutubeURI),
		}
		if v.Title != "" {
			tf.Video.Title = types.StringValue(v.Title)
		} else {
			tf.Video.Title = types.StringNull()
		}
		if v.Description != "" {
			tf.Video.Description = types.StringValue(v.Description)
		} else {
			tf.Video.Description = types.StringNull()
		}
		if v.Caption != "" {
			tf.Video.Caption = types.StringValue(v.Caption)
		} else {
			tf.Video.Caption = types.StringNull()
		}
	}

	if item.SectionHeader != nil {
		sh := item.SectionHeader
		tf.SectionHeader = &SectionHeaderModel{
			Title: types.StringValue(sh.Title),
		}
		if sh.Description != "" {
			tf.SectionHeader.Description = types.StringValue(sh.Description)
		} else {
			tf.SectionHeader.Description = types.StringNull()
		}
	}

	return tf
}

// convertGradingToTF converts a convert.GradingBlock to a TF GradingModel.
func convertGradingToTF(g *convert.GradingBlock) *GradingModel {
	gm := &GradingModel{
		Points: types.Int64Value(g.Points),
	}

	if g.CorrectAnswer != "" {
		gm.CorrectAnswer = types.StringValue(g.CorrectAnswer)
	} else {
		gm.CorrectAnswer = types.StringNull()
	}

	if g.FeedbackCorrect != "" {
		gm.FeedbackCorrect = types.StringValue(g.FeedbackCorrect)
	} else {
		gm.FeedbackCorrect = types.StringNull()
	}

	if g.FeedbackIncorrect != "" {
		gm.FeedbackIncorrect = types.StringValue(g.FeedbackIncorrect)
	} else {
		gm.FeedbackIncorrect = types.StringNull()
	}

	return gm
}

// itemObjectType returns the types.ObjectType for a single item block element.
// This must match the schema defined in schema.go.
func itemObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"item_key":       types.StringType,
			"google_item_id": types.StringType,
			"multiple_choice": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"question_text": types.StringType,
					"options":       types.ListType{ElemType: types.StringType},
					"option":        types.ListType{ElemType: types.ObjectType{AttrTypes: choiceOptionAttrTypes()}},
					"required":      types.BoolType,
					"shuffle":       types.BoolType,
					"has_other":     types.BoolType,
					"grading":       gradingObjectType(),
				},
			},
			"short_answer": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"question_text": types.StringType,
					"required":      types.BoolType,
					"grading":       gradingObjectType(),
				},
			},
			"paragraph": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"question_text": types.StringType,
					"required":      types.BoolType,
					"grading":       gradingObjectType(),
				},
			},
			"dropdown": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"question_text": types.StringType,
					"options":       types.ListType{ElemType: types.StringType},
					"option":        types.ListType{ElemType: types.ObjectType{AttrTypes: choiceOptionAttrTypes()}},
					"required":      types.BoolType,
					"shuffle":       types.BoolType,
					"grading":       gradingObjectType(),
				},
			},
			"checkbox": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"question_text": types.StringType,
					"options":       types.ListType{ElemType: types.StringType},
					"option":        types.ListType{ElemType: types.ObjectType{AttrTypes: choiceOptionAttrTypes()}},
					"required":      types.BoolType,
					"shuffle":       types.BoolType,
					"has_other":     types.BoolType,
					"grading":       gradingObjectType(),
				},
			},
			"multiple_choice_grid": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"question_text":     types.StringType,
					"rows":              types.ListType{ElemType: types.StringType},
					"columns":           types.ListType{ElemType: types.StringType},
					"required":          types.BoolType,
					"shuffle_questions": types.BoolType,
					"shuffle_columns":   types.BoolType,
				},
			},
			"checkbox_grid": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"question_text":     types.StringType,
					"rows":              types.ListType{ElemType: types.StringType},
					"columns":           types.ListType{ElemType: types.StringType},
					"required":          types.BoolType,
					"shuffle_questions": types.BoolType,
					"shuffle_columns":   types.BoolType,
				},
			},
			"date": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"question_text": types.StringType,
					"required":      types.BoolType,
					"include_year":  types.BoolType,
				},
			},
			"date_time": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"question_text": types.StringType,
					"required":      types.BoolType,
					"include_year":  types.BoolType,
				},
			},
			"scale": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"question_text": types.StringType,
					"required":      types.BoolType,
					"low":           types.Int64Type,
					"high":          types.Int64Type,
					"low_label":     types.StringType,
					"high_label":    types.StringType,
				},
			},
			"time": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"question_text": types.StringType,
					"required":      types.BoolType,
					"duration":      types.BoolType,
				},
			},
			"rating": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"question_text":      types.StringType,
					"required":           types.BoolType,
					"icon_type":          types.StringType,
					"rating_scale_level": types.Int64Type,
				},
			},
			"file_upload": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"question_text": types.StringType,
					"required":      types.BoolType,
					"folder_id":     types.StringType,
					"max_file_size": types.Int64Type,
					"max_files":     types.Int64Type,
					"types":         types.ListType{ElemType: types.StringType},
				},
			},
			"text_item": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"title":       types.StringType,
					"description": types.StringType,
				},
			},
			"image": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"title":       types.StringType,
					"description": types.StringType,
					"source_uri":  types.StringType,
					"alt_text":    types.StringType,
					"content_uri": types.StringType,
				},
			},
			"video": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"title":       types.StringType,
					"description": types.StringType,
					"youtube_uri": types.StringType,
					"caption":     types.StringType,
				},
			},
			"section_header": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"title":       types.StringType,
					"description": types.StringType,
				},
			},
		},
	}
}

// gradingObjectType returns the types.ObjectType for the grading block.
func gradingObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"points":             types.Int64Type,
			"correct_answer":     types.StringType,
			"feedback_correct":   types.StringType,
			"feedback_incorrect": types.StringType,
		},
	}
}

func filterItemsByKeyMap(items []convert.ItemModel, keyMap map[string]string) []convert.ItemModel {
	if len(items) == 0 || keyMap == nil {
		return items
	}

	out := make([]convert.ItemModel, 0, len(items))
	for _, it := range items {
		if it.GoogleItemID == "" {
			continue
		}
		if _, ok := keyMap[it.GoogleItemID]; ok {
			out = append(out, it)
		}
	}
	return out
}

func overlayConvertItemInputsFromTF(
	ctx context.Context,
	items []convert.ItemModel,
	tfItems types.List,
) ([]convert.ItemModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if tfItems.IsNull() || tfItems.IsUnknown() || len(tfItems.Elements()) == 0 || len(items) == 0 {
		return items, diags
	}

	var models []ItemModel
	diags.Append(tfItems.ElementsAs(ctx, &models, false)...)
	if diags.HasError() {
		return items, diags
	}

	byKey := make(map[string]ItemModel, len(models))
	for _, it := range models {
		k := it.ItemKey.ValueString()
		if k != "" {
			byKey[k] = it
		}
	}

	for i := range items {
		k := items[i].ItemKey
		tf, ok := byKey[k]
		if !ok {
			continue
		}

		// The Forms API does not reliably return some input-only fields (notably
		// Image.SourceUri). Preserve configured values from TF.
		if items[i].Image != nil && tf.Image != nil {
			if items[i].Image.SourceURI == "" && !tf.Image.SourceURI.IsNull() && !tf.Image.SourceURI.IsUnknown() {
				items[i].Image.SourceURI = tf.Image.SourceURI.ValueString()
			}
			if items[i].Image.AltText == "" && !tf.Image.AltText.IsNull() && !tf.Image.AltText.IsUnknown() {
				items[i].Image.AltText = tf.Image.AltText.ValueString()
			}
		}
		if items[i].Video != nil && tf.Video != nil {
			if items[i].Video.YoutubeURI == "" && !tf.Video.YoutubeURI.IsNull() && !tf.Video.YoutubeURI.IsUnknown() {
				items[i].Video.YoutubeURI = tf.Video.YoutubeURI.ValueString()
			}
			if items[i].Video.Caption == "" && !tf.Video.Caption.IsNull() && !tf.Video.Caption.IsUnknown() {
				items[i].Video.Caption = tf.Video.Caption.ValueString()
			}
		}
	}

	return items, diags
}
