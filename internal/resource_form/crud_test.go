// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceform

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	forms "google.golang.org/api/forms/v1"

	"github.com/your-org/terraform-provider-googleforms/internal/client"
	"github.com/your-org/terraform-provider-googleforms/internal/testutil"
)

// ---------------------------------------------------------------------------
// test helpers
// ---------------------------------------------------------------------------

// testResource creates a FormResource with mocked APIs.
func testResource(formsAPI client.FormsAPI, driveAPI client.DriveAPI) *FormResource {
	return &FormResource{
		client: &client.Client{
			Forms: formsAPI,
			Drive: driveAPI,
		},
	}
}

func buildPlan(t *testing.T, vals map[string]tftypes.Value) tfsdk.Plan {
	t.Helper()
	schemaResp := testSchemaResp()
	s := schemaResp.Schema
	tfType := s.Type().TerraformType(context.Background())
	objType := tfType.(tftypes.Object)
	merged := make(map[string]tftypes.Value)
	for k, v := range objType.AttributeTypes {
		merged[k] = tftypes.NewValue(v, nil)
	}
	for k, v := range vals {
		merged[k] = v
	}
	return tfsdk.Plan{
		Schema: s,
		Raw:    tftypes.NewValue(objType, merged),
	}
}

func buildState(t *testing.T, vals map[string]tftypes.Value) tfsdk.State {
	t.Helper()
	schemaResp := testSchemaResp()
	s := schemaResp.Schema
	tfType := s.Type().TerraformType(context.Background())
	objType := tfType.(tftypes.Object)
	merged := make(map[string]tftypes.Value)
	for k, v := range objType.AttributeTypes {
		merged[k] = tftypes.NewValue(v, nil)
	}
	for k, v := range vals {
		merged[k] = v
	}
	return tfsdk.State{
		Schema: s,
		Raw:    tftypes.NewValue(objType, merged),
	}
}

// emptyState returns a completely null/unset state for Create responses.
func emptyState(t *testing.T) tfsdk.State {
	t.Helper()
	schemaResp := testSchemaResp()
	s := schemaResp.Schema
	tfType := s.Type().TerraformType(context.Background())
	objType := tfType.(tftypes.Object)
	return tfsdk.State{
		Schema: s,
		Raw:    tftypes.NewValue(objType, nil),
	}
}

// basicFormResponse returns a minimal forms.Form suitable for mock Get returns.
func basicFormResponse(id, title string) *forms.Form {
	return &forms.Form{
		FormId:       id,
		Info:         &forms.Info{Title: title, DocumentTitle: title},
		ResponderUri: "https://docs.google.com/forms/d/" + id + "/viewform",
	}
}

// formWithItems returns a forms.Form that has short_answer/multiple_choice items.
func formWithItems(id, title string) *forms.Form {
	return &forms.Form{
		FormId:       id,
		Info:         &forms.Info{Title: title, DocumentTitle: title},
		ResponderUri: "https://docs.google.com/forms/d/" + id + "/viewform",
		Items: []*forms.Item{
			{
				ItemId: "gid_1",
				Title:  "Name?",
				QuestionItem: &forms.QuestionItem{
					Question: &forms.Question{
						TextQuestion: &forms.TextQuestion{Paragraph: false},
					},
				},
			},
			{
				ItemId: "gid_2",
				Title:  "Color?",
				QuestionItem: &forms.QuestionItem{
					Question: &forms.Question{
						ChoiceQuestion: &forms.ChoiceQuestion{
							Type: "RADIO",
							Options: []*forms.Option{
								{Value: "Red"},
								{Value: "Blue"},
							},
						},
					},
				},
			},
		},
	}
}

// stateFormID extracts the "id" attribute from the response state.
func stateFormID(t *testing.T, state tfsdk.State) string {
	t.Helper()
	var model FormResourceModel
	diags := state.Get(context.Background(), &model)
	if diags.HasError() {
		t.Fatalf("failed to read state: %v", diags.Errors())
	}
	return model.ID.ValueString()
}

// ---------------------------------------------------------------------------
// Create tests
// ---------------------------------------------------------------------------

func TestCreate_BasicForm_Success(t *testing.T) {
	t.Parallel()

	mockForms := &testutil.MockFormsAPI{
		CreateFunc: func(_ context.Context, form *forms.Form) (*forms.Form, error) {
			return &forms.Form{
				FormId: "new-form-123",
				Info:   form.Info,
			}, nil
		},
		GetFunc: func(_ context.Context, formID string) (*forms.Form, error) {
			return basicFormResponse(formID, "My Form"), nil
		},
	}
	mockDrive := &testutil.MockDriveAPI{}

	r := testResource(mockForms, mockDrive)
	ctx := context.Background()

	plan := buildPlan(t, map[string]tftypes.Value{
		"title":               tftypes.NewValue(tftypes.String, "My Form"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, false),
	})

	resp := &resource.CreateResponse{
		State: emptyState(t),
	}

	r.Create(ctx, resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		for _, d := range resp.Diagnostics.Errors() {
			t.Logf("  diagnostic: %s -- %s", d.Summary(), d.Detail())
		}
		t.Fatal("expected no errors during Create")
	}

	gotID := stateFormID(t, resp.State)
	if gotID != "new-form-123" {
		t.Fatalf("expected form_id %q, got %q", "new-form-123", gotID)
	}
}

func TestCreate_WithItems_Success(t *testing.T) {
	t.Parallel()

	var batchCalled bool
	mockForms := &testutil.MockFormsAPI{
		CreateFunc: func(_ context.Context, form *forms.Form) (*forms.Form, error) {
			return &forms.Form{
				FormId: "items-form-456",
				Info:   form.Info,
			}, nil
		},
		BatchUpdateFunc: func(_ context.Context, _ string, req *forms.BatchUpdateFormRequest) (*forms.BatchUpdateFormResponse, error) {
			batchCalled = true
			// Expect at least: UpdateFormInfo + 2 CreateItem requests.
			createCount := 0
			for _, r := range req.Requests {
				if r.CreateItem != nil {
					createCount++
				}
			}
			if createCount != 2 {
				return nil, fmt.Errorf("expected 2 CreateItem requests, got %d", createCount)
			}
			return &forms.BatchUpdateFormResponse{}, nil
		},
		GetFunc: func(_ context.Context, formID string) (*forms.Form, error) {
			return formWithItems(formID, "Items Form"), nil
		},
	}
	mockDrive := &testutil.MockDriveAPI{}

	r := testResource(mockForms, mockDrive)
	ctx := context.Background()

	plan := buildPlan(t, map[string]tftypes.Value{
		"title":               tftypes.NewValue(tftypes.String, "Items Form"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, false),
		"item": itemListVal(
			saItem("q1", "Name?", nil),
			mcItem("q2", "Color?", []string{"Red", "Blue"}, nil),
		),
	})

	resp := &resource.CreateResponse{
		State: emptyState(t),
	}

	r.Create(ctx, resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		for _, d := range resp.Diagnostics.Errors() {
			t.Logf("  diagnostic: %s -- %s", d.Summary(), d.Detail())
		}
		t.Fatal("expected no errors during Create with items")
	}

	if !batchCalled {
		t.Fatal("expected BatchUpdate to be called for items")
	}
}

func TestCreate_WithContentJSON_Success(t *testing.T) {
	t.Parallel()

	var batchCalled bool
	contentJSON := `[{"title":"Q1","questionItem":{"question":{"textQuestion":{"paragraph":false}}}}]`

	mockForms := &testutil.MockFormsAPI{
		CreateFunc: func(_ context.Context, form *forms.Form) (*forms.Form, error) {
			return &forms.Form{
				FormId: "json-form-789",
				Info:   form.Info,
			}, nil
		},
		BatchUpdateFunc: func(_ context.Context, _ string, req *forms.BatchUpdateFormRequest) (*forms.BatchUpdateFormResponse, error) {
			batchCalled = true
			createCount := 0
			for _, r := range req.Requests {
				if r.CreateItem != nil {
					createCount++
				}
			}
			if createCount != 1 {
				return nil, fmt.Errorf("expected 1 CreateItem from content_json, got %d", createCount)
			}
			return &forms.BatchUpdateFormResponse{}, nil
		},
		GetFunc: func(_ context.Context, formID string) (*forms.Form, error) {
			return basicFormResponse(formID, "JSON Form"), nil
		},
	}
	mockDrive := &testutil.MockDriveAPI{}

	r := testResource(mockForms, mockDrive)
	ctx := context.Background()

	plan := buildPlan(t, map[string]tftypes.Value{
		"title":               tftypes.NewValue(tftypes.String, "JSON Form"),
		"content_json":        tftypes.NewValue(tftypes.String, contentJSON),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, false),
	})

	resp := &resource.CreateResponse{
		State: emptyState(t),
	}

	r.Create(ctx, resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		for _, d := range resp.Diagnostics.Errors() {
			t.Logf("  diagnostic: %s -- %s", d.Summary(), d.Detail())
		}
		t.Fatal("expected no errors during Create with content_json")
	}

	if !batchCalled {
		t.Fatal("expected BatchUpdate to be called for content_json items")
	}
}

func TestCreate_APIError_ReturnsDiagnostic(t *testing.T) {
	t.Parallel()

	mockForms := &testutil.MockFormsAPI{
		CreateFunc: func(_ context.Context, _ *forms.Form) (*forms.Form, error) {
			return nil, &client.APIError{
				StatusCode: 500,
				Message:    "internal server error",
			}
		},
	}
	mockDrive := &testutil.MockDriveAPI{}

	r := testResource(mockForms, mockDrive)
	ctx := context.Background()

	plan := buildPlan(t, map[string]tftypes.Value{
		"title":               tftypes.NewValue(tftypes.String, "Fail Form"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, false),
	})

	resp := &resource.CreateResponse{
		State: emptyState(t),
	}

	r.Create(ctx, resource.CreateRequest{Plan: plan}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostic when Create API fails")
	}

	found := false
	for _, d := range resp.Diagnostics.Errors() {
		if strContains(d.Summary(), "Error Creating Google Form") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected diagnostic with summary containing 'Error Creating Google Form'")
	}
}

func TestCreate_BatchUpdateError_PartialStateSaved(t *testing.T) {
	t.Parallel()

	mockForms := &testutil.MockFormsAPI{
		CreateFunc: func(_ context.Context, form *forms.Form) (*forms.Form, error) {
			return &forms.Form{
				FormId: "partial-form-abc",
				Info:   form.Info,
			}, nil
		},
		BatchUpdateFunc: func(_ context.Context, _ string, _ *forms.BatchUpdateFormRequest) (*forms.BatchUpdateFormResponse, error) {
			return nil, &client.APIError{
				StatusCode: 400,
				Message:    "batch update failed",
			}
		},
	}
	mockDrive := &testutil.MockDriveAPI{}

	r := testResource(mockForms, mockDrive)
	ctx := context.Background()

	plan := buildPlan(t, map[string]tftypes.Value{
		"title":               tftypes.NewValue(tftypes.String, "Partial Form"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, false),
		"item":                itemListVal(saItem("q1", "Name?", nil)),
	})

	resp := &resource.CreateResponse{
		State: emptyState(t),
	}

	r.Create(ctx, resource.CreateRequest{Plan: plan}, resp)

	// Should have an error from BatchUpdate failure.
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostic when BatchUpdate fails")
	}

	// The form ID should still be in state (partial save).
	gotID := stateFormID(t, resp.State)
	if gotID != "partial-form-abc" {
		t.Fatalf("expected partial state to have form_id %q, got %q", "partial-form-abc", gotID)
	}
}

// ---------------------------------------------------------------------------
// Read tests
// ---------------------------------------------------------------------------

func TestRead_ExistingForm_PopulatesState(t *testing.T) {
	t.Parallel()

	mockForms := &testutil.MockFormsAPI{
		GetFunc: func(_ context.Context, formID string) (*forms.Form, error) {
			return basicFormResponse(formID, "Existing Form"), nil
		},
	}
	mockDrive := &testutil.MockDriveAPI{}

	r := testResource(mockForms, mockDrive)
	ctx := context.Background()

	state := buildState(t, map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, "existing-form-id"),
		"title":               tftypes.NewValue(tftypes.String, "Existing Form"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, false),
	})

	resp := &resource.ReadResponse{
		State: state,
	}

	r.Read(ctx, resource.ReadRequest{State: state}, resp)

	if resp.Diagnostics.HasError() {
		for _, d := range resp.Diagnostics.Errors() {
			t.Logf("  diagnostic: %s -- %s", d.Summary(), d.Detail())
		}
		t.Fatal("expected no errors during Read")
	}

	gotID := stateFormID(t, resp.State)
	if gotID != "existing-form-id" {
		t.Fatalf("expected form_id %q, got %q", "existing-form-id", gotID)
	}
}

func TestRead_FormNotFound_RemovesFromState(t *testing.T) {
	t.Parallel()

	mockForms := &testutil.MockFormsAPI{
		GetFunc: func(_ context.Context, formID string) (*forms.Form, error) {
			return nil, &client.NotFoundError{
				Resource: "Form",
				ID:       formID,
			}
		},
	}
	mockDrive := &testutil.MockDriveAPI{}

	r := testResource(mockForms, mockDrive)
	ctx := context.Background()

	state := buildState(t, map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, "gone-form-id"),
		"title":               tftypes.NewValue(tftypes.String, "Gone Form"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, false),
	})

	resp := &resource.ReadResponse{
		State: state,
	}

	r.Read(ctx, resource.ReadRequest{State: state}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatal("expected no errors when form is not found (should just remove from state)")
	}

	// After RemoveResource, the state Raw should be null.
	if !resp.State.Raw.IsNull() {
		t.Fatal("expected state to be null after form not found")
	}
}

func TestRead_APIError_ReturnsDiagnostic(t *testing.T) {
	t.Parallel()

	mockForms := &testutil.MockFormsAPI{
		GetFunc: func(_ context.Context, _ string) (*forms.Form, error) {
			return nil, &client.APIError{
				StatusCode: 500,
				Message:    "server error",
			}
		},
	}
	mockDrive := &testutil.MockDriveAPI{}

	r := testResource(mockForms, mockDrive)
	ctx := context.Background()

	state := buildState(t, map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, "error-form-id"),
		"title":               tftypes.NewValue(tftypes.String, "Error Form"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, false),
	})

	resp := &resource.ReadResponse{
		State: state,
	}

	r.Read(ctx, resource.ReadRequest{State: state}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostic when Read API fails")
	}

	found := false
	for _, d := range resp.Diagnostics.Errors() {
		if strContains(d.Summary(), "Error Reading Google Form") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected diagnostic with summary containing 'Error Reading Google Form'")
	}
}

// ---------------------------------------------------------------------------
// Update tests
// ---------------------------------------------------------------------------

func TestUpdate_TitleChange_Success(t *testing.T) {
	t.Parallel()

	var batchReqs []*forms.Request
	getCalls := 0
	mockForms := &testutil.MockFormsAPI{
		GetFunc: func(_ context.Context, formID string) (*forms.Form, error) {
			getCalls++
			// First call: current form state (pre-update).
			// Second call: final form state (post-update).
			return basicFormResponse(formID, "New Title"), nil
		},
		BatchUpdateFunc: func(_ context.Context, _ string, req *forms.BatchUpdateFormRequest) (*forms.BatchUpdateFormResponse, error) {
			batchReqs = req.Requests
			return &forms.BatchUpdateFormResponse{}, nil
		},
	}
	mockDrive := &testutil.MockDriveAPI{}

	r := testResource(mockForms, mockDrive)
	ctx := context.Background()

	// Prior state: old title.
	state := buildState(t, map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, "update-form-id"),
		"title":               tftypes.NewValue(tftypes.String, "Old Title"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, false),
	})

	// New plan: new title.
	plan := buildPlan(t, map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, "update-form-id"),
		"title":               tftypes.NewValue(tftypes.String, "New Title"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, false),
	})

	resp := &resource.UpdateResponse{
		State: state,
	}

	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: state}, resp)

	if resp.Diagnostics.HasError() {
		for _, d := range resp.Diagnostics.Errors() {
			t.Logf("  diagnostic: %s -- %s", d.Summary(), d.Detail())
		}
		t.Fatal("expected no errors during Update")
	}

	// Verify BatchUpdate was called with an UpdateFormInfo request.
	foundUpdateInfo := false
	for _, req := range batchReqs {
		if req.UpdateFormInfo != nil {
			foundUpdateInfo = true
			if req.UpdateFormInfo.Info.Title != "New Title" {
				t.Fatalf("expected UpdateFormInfo title %q, got %q", "New Title", req.UpdateFormInfo.Info.Title)
			}
		}
	}
	if !foundUpdateInfo {
		t.Fatal("expected BatchUpdate to contain an UpdateFormInfo request")
	}

	if getCalls != 2 {
		t.Fatalf("expected 2 Get calls (pre+post update), got %d", getCalls)
	}
}

func TestUpdate_ItemsReplaced_Success(t *testing.T) {
	t.Parallel()

	var batchReqs []*forms.Request
	getCalls := 0
	mockForms := &testutil.MockFormsAPI{
		GetFunc: func(_ context.Context, formID string) (*forms.Form, error) {
			getCalls++
			if getCalls == 1 {
				// Pre-update: form has 1 existing item.
				return &forms.Form{
					FormId:       formID,
					Info:         &forms.Info{Title: "Update Items", DocumentTitle: "Update Items"},
					ResponderUri: "https://docs.google.com/forms/d/" + formID + "/viewform",
					Items: []*forms.Item{
						{
							ItemId: "old_gid",
							Title:  "Old Q?",
							QuestionItem: &forms.QuestionItem{
								Question: &forms.Question{
									TextQuestion: &forms.TextQuestion{Paragraph: false},
								},
							},
						},
					},
				}, nil
			}
			// Post-update: form has the new item.
			return &forms.Form{
				FormId:       formID,
				Info:         &forms.Info{Title: "Update Items", DocumentTitle: "Update Items"},
				ResponderUri: "https://docs.google.com/forms/d/" + formID + "/viewform",
				Items: []*forms.Item{
					{
						ItemId: "new_gid",
						Title:  "New Q?",
						QuestionItem: &forms.QuestionItem{
							Question: &forms.Question{
								TextQuestion: &forms.TextQuestion{Paragraph: false},
							},
						},
					},
				},
			}, nil
		},
		BatchUpdateFunc: func(_ context.Context, _ string, req *forms.BatchUpdateFormRequest) (*forms.BatchUpdateFormResponse, error) {
			batchReqs = req.Requests
			return &forms.BatchUpdateFormResponse{}, nil
		},
	}
	mockDrive := &testutil.MockDriveAPI{}

	r := testResource(mockForms, mockDrive)
	ctx := context.Background()

	state := buildState(t, map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, "replace-form-id"),
		"title":               tftypes.NewValue(tftypes.String, "Update Items"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, false),
		"item":                itemListVal(saItem("q_old", "Old Q?", nil)),
	})

	plan := buildPlan(t, map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, "replace-form-id"),
		"title":               tftypes.NewValue(tftypes.String, "Update Items"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, false),
		"item":                itemListVal(saItem("q_new", "New Q?", nil)),
	})

	resp := &resource.UpdateResponse{
		State: state,
	}

	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: state}, resp)

	if resp.Diagnostics.HasError() {
		for _, d := range resp.Diagnostics.Errors() {
			t.Logf("  diagnostic: %s -- %s", d.Summary(), d.Detail())
		}
		t.Fatal("expected no errors during Update with item replacement")
	}

	// Verify: should have DeleteItem + CreateItem requests (plus UpdateFormInfo).
	deleteCount := 0
	createCount := 0
	for _, req := range batchReqs {
		if req.DeleteItem != nil {
			deleteCount++
		}
		if req.CreateItem != nil {
			createCount++
		}
	}
	if deleteCount != 1 {
		t.Fatalf("expected 1 DeleteItem request, got %d", deleteCount)
	}
	if createCount != 1 {
		t.Fatalf("expected 1 CreateItem request, got %d", createCount)
	}
}

// ---------------------------------------------------------------------------
// Delete tests
// ---------------------------------------------------------------------------

func TestDelete_Success(t *testing.T) {
	t.Parallel()

	var deletedID string
	mockForms := &testutil.MockFormsAPI{}
	mockDrive := &testutil.MockDriveAPI{
		DeleteFunc: func(_ context.Context, fileID string) error {
			deletedID = fileID
			return nil
		},
	}

	r := testResource(mockForms, mockDrive)
	ctx := context.Background()

	state := buildState(t, map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, "delete-form-id"),
		"title":               tftypes.NewValue(tftypes.String, "Delete Me"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, false),
	})

	resp := &resource.DeleteResponse{
		State: state,
	}

	r.Delete(ctx, resource.DeleteRequest{State: state}, resp)

	if resp.Diagnostics.HasError() {
		for _, d := range resp.Diagnostics.Errors() {
			t.Logf("  diagnostic: %s -- %s", d.Summary(), d.Detail())
		}
		t.Fatal("expected no errors during Delete")
	}

	if deletedID != "delete-form-id" {
		t.Fatalf("expected Drive.Delete called with %q, got %q", "delete-form-id", deletedID)
	}
}

func TestDelete_NotFound_NoError(t *testing.T) {
	t.Parallel()

	mockForms := &testutil.MockFormsAPI{}
	mockDrive := &testutil.MockDriveAPI{
		DeleteFunc: func(_ context.Context, fileID string) error {
			return &client.NotFoundError{
				Resource: "Form",
				ID:       fileID,
			}
		},
	}

	r := testResource(mockForms, mockDrive)
	ctx := context.Background()

	state := buildState(t, map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, "already-gone-id"),
		"title":               tftypes.NewValue(tftypes.String, "Already Gone"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, false),
	})

	resp := &resource.DeleteResponse{
		State: state,
	}

	r.Delete(ctx, resource.DeleteRequest{State: state}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatal("expected no errors when deleting an already-deleted form")
	}
}

func TestDelete_APIError_ReturnsDiagnostic(t *testing.T) {
	t.Parallel()

	mockForms := &testutil.MockFormsAPI{}
	mockDrive := &testutil.MockDriveAPI{
		DeleteFunc: func(_ context.Context, _ string) error {
			return &client.APIError{
				StatusCode: 500,
				Message:    "drive server error",
			}
		},
	}

	r := testResource(mockForms, mockDrive)
	ctx := context.Background()

	state := buildState(t, map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, "error-delete-id"),
		"title":               tftypes.NewValue(tftypes.String, "Error Delete"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, false),
	})

	resp := &resource.DeleteResponse{
		State: state,
	}

	r.Delete(ctx, resource.DeleteRequest{State: state}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostic when Delete API fails")
	}

	found := false
	for _, d := range resp.Diagnostics.Errors() {
		if strContains(d.Summary(), "Error Deleting Google Form") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected diagnostic with summary containing 'Error Deleting Google Form'")
	}
}
