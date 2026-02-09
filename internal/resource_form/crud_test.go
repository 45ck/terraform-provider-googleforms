// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceform

import (
	"context"
	"fmt"
	"strings"
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
		if strings.Contains(d.Summary(), "Error Creating Google Form") {
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
		if strings.Contains(d.Summary(), "Error Reading Google Form") {
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
		if strings.Contains(d.Summary(), "Error Deleting Google Form") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected diagnostic with summary containing 'Error Deleting Google Form'")
	}
}

// ---------------------------------------------------------------------------
// Additional Create tests
// ---------------------------------------------------------------------------

func TestCreate_WithQuizGrading_Success(t *testing.T) {
	t.Parallel()

	var batchReqs []*forms.Request
	mockForms := &testutil.MockFormsAPI{
		CreateFunc: func(_ context.Context, form *forms.Form) (*forms.Form, error) {
			return &forms.Form{
				FormId: "quiz-form-001",
				Info:   form.Info,
			}, nil
		},
		BatchUpdateFunc: func(_ context.Context, _ string, req *forms.BatchUpdateFormRequest) (*forms.BatchUpdateFormResponse, error) {
			batchReqs = req.Requests
			return &forms.BatchUpdateFormResponse{}, nil
		},
		GetFunc: func(_ context.Context, formID string) (*forms.Form, error) {
			return &forms.Form{
				FormId:       formID,
				Info:         &forms.Info{Title: "Quiz Form", DocumentTitle: "Quiz Form"},
				ResponderUri: "https://docs.google.com/forms/d/" + formID + "/viewform",
				Settings: &forms.FormSettings{
					QuizSettings: &forms.QuizSettings{IsQuiz: true},
				},
				Items: []*forms.Item{
					{
						ItemId: "gid_quiz_1",
						Title:  "Capital of France?",
						QuestionItem: &forms.QuestionItem{
							Question: &forms.Question{
								ChoiceQuestion: &forms.ChoiceQuestion{
									Type: "RADIO",
									Options: []*forms.Option{
										{Value: "Paris"},
										{Value: "London"},
									},
								},
								Grading: &forms.Grading{
									PointValue: 10,
									CorrectAnswers: &forms.CorrectAnswers{
										Answers: []*forms.CorrectAnswer{{Value: "Paris"}},
									},
								},
							},
						},
					},
				},
			}, nil
		},
	}
	mockDrive := &testutil.MockDriveAPI{}

	r := testResource(mockForms, mockDrive)
	ctx := context.Background()

	grading := &map[string]tftypes.Value{
		"points":             tftypes.NewValue(tftypes.Number, 10),
		"correct_answer":     tftypes.NewValue(tftypes.String, "Paris"),
		"feedback_correct":   tftypes.NewValue(tftypes.String, nil),
		"feedback_incorrect": tftypes.NewValue(tftypes.String, nil),
	}

	plan := buildPlan(t, map[string]tftypes.Value{
		"title":               tftypes.NewValue(tftypes.String, "Quiz Form"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, true),
		"item": itemListVal(
			mcItem("q1", "Capital of France?", []string{"Paris", "London"}, grading),
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
		t.Fatal("expected no errors during Create with quiz grading")
	}

	// Verify BatchUpdate includes UpdateSettings for quiz mode.
	foundQuizSetting := false
	for _, req := range batchReqs {
		if req.UpdateSettings != nil &&
			req.UpdateSettings.Settings != nil &&
			req.UpdateSettings.Settings.QuizSettings != nil &&
			req.UpdateSettings.Settings.QuizSettings.IsQuiz {
			foundQuizSetting = true
		}
	}
	if !foundQuizSetting {
		t.Fatal("expected BatchUpdate to contain quiz settings request")
	}
}

func TestCreate_WithPublishSettings_Success(t *testing.T) {
	t.Parallel()

	var publishedCalled bool
	var gotPublished, gotAccepting bool

	mockForms := &testutil.MockFormsAPI{
		CreateFunc: func(_ context.Context, form *forms.Form) (*forms.Form, error) {
			return &forms.Form{
				FormId: "pub-form-002",
				Info:   form.Info,
			}, nil
		},
		SetPublishSettingsFunc: func(_ context.Context, _ string, isPublished bool, isAccepting bool) error {
			publishedCalled = true
			gotPublished = isPublished
			gotAccepting = isAccepting
			return nil
		},
		GetFunc: func(_ context.Context, formID string) (*forms.Form, error) {
			return basicFormResponse(formID, "Published Form"), nil
		},
	}
	mockDrive := &testutil.MockDriveAPI{}

	r := testResource(mockForms, mockDrive)
	ctx := context.Background()

	plan := buildPlan(t, map[string]tftypes.Value{
		"title":               tftypes.NewValue(tftypes.String, "Published Form"),
		"published":           tftypes.NewValue(tftypes.Bool, true),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, true),
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
		t.Fatal("expected no errors during Create with publish settings")
	}

	if !publishedCalled {
		t.Fatal("expected SetPublishSettings to be called")
	}
	if !gotPublished {
		t.Fatal("expected published=true passed to SetPublishSettings")
	}
	if !gotAccepting {
		t.Fatal("expected accepting=true passed to SetPublishSettings")
	}
}

// ---------------------------------------------------------------------------
// Additional Read tests
// ---------------------------------------------------------------------------

func TestRead_WithItems_CorrectMapping(t *testing.T) {
	t.Parallel()

	mockForms := &testutil.MockFormsAPI{
		GetFunc: func(_ context.Context, formID string) (*forms.Form, error) {
			return formWithItems(formID, "Items Read Form"), nil
		},
	}
	mockDrive := &testutil.MockDriveAPI{}

	r := testResource(mockForms, mockDrive)
	ctx := context.Background()

	// State includes items with google_item_id set so key map can be built.
	iType := itemBlockType()
	gType := iType.AttributeTypes["short_answer"].(tftypes.Object).AttributeTypes["grading"]

	sa := tftypes.NewValue(iType.AttributeTypes["short_answer"], map[string]tftypes.Value{
		"question_text": tftypes.NewValue(tftypes.String, "Name?"),
		"required":      tftypes.NewValue(tftypes.Bool, false),
		"grading":       tftypes.NewValue(gType, nil),
	})
	item1 := tftypes.NewValue(iType, map[string]tftypes.Value{
		"item_key":        tftypes.NewValue(tftypes.String, "q1"),
		"google_item_id":  tftypes.NewValue(tftypes.String, "gid_1"),
		"multiple_choice": tftypes.NewValue(iType.AttributeTypes["multiple_choice"], nil),
		"short_answer":    sa,
		"paragraph":       tftypes.NewValue(iType.AttributeTypes["paragraph"], nil),
	})

	mcType := iType.AttributeTypes["multiple_choice"]
	mc := tftypes.NewValue(mcType, map[string]tftypes.Value{
		"question_text": tftypes.NewValue(tftypes.String, "Color?"),
		"options": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "Red"),
			tftypes.NewValue(tftypes.String, "Blue"),
		}),
		"required": tftypes.NewValue(tftypes.Bool, false),
		"grading":  tftypes.NewValue(gType, nil),
	})
	item2 := tftypes.NewValue(iType, map[string]tftypes.Value{
		"item_key":        tftypes.NewValue(tftypes.String, "q2"),
		"google_item_id":  tftypes.NewValue(tftypes.String, "gid_2"),
		"multiple_choice": mc,
		"short_answer":    tftypes.NewValue(iType.AttributeTypes["short_answer"], nil),
		"paragraph":       tftypes.NewValue(iType.AttributeTypes["paragraph"], nil),
	})

	itemList := tftypes.NewValue(tftypes.List{ElementType: iType}, []tftypes.Value{item1, item2})

	state := buildState(t, map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, "items-read-id"),
		"title":               tftypes.NewValue(tftypes.String, "Items Read Form"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, false),
		"item":                itemList,
	})

	resp := &resource.ReadResponse{
		State: state,
	}

	r.Read(ctx, resource.ReadRequest{State: state}, resp)

	if resp.Diagnostics.HasError() {
		for _, d := range resp.Diagnostics.Errors() {
			t.Logf("  diagnostic: %s -- %s", d.Summary(), d.Detail())
		}
		t.Fatal("expected no errors during Read with items")
	}

	// Verify the state still has the correct form ID.
	gotID := stateFormID(t, resp.State)
	if gotID != "items-read-id" {
		t.Fatalf("expected form_id %q, got %q", "items-read-id", gotID)
	}

	// Verify items were mapped into state by reading the full model.
	var model FormResourceModel
	diags := resp.State.Get(ctx, &model)
	if diags.HasError() {
		t.Fatalf("failed to read state model: %v", diags.Errors())
	}
	if model.Items.IsNull() || model.Items.IsUnknown() {
		t.Fatal("expected items to be populated in state after Read")
	}
	if len(model.Items.Elements()) != 2 {
		t.Fatalf("expected 2 items in state, got %d", len(model.Items.Elements()))
	}
}

// ---------------------------------------------------------------------------
// Additional Update tests
// ---------------------------------------------------------------------------

func TestUpdate_PublishSettingsChanged_Success(t *testing.T) {
	t.Parallel()

	var publishCalled bool
	var gotPublished, gotAccepting bool

	mockForms := &testutil.MockFormsAPI{
		GetFunc: func(_ context.Context, formID string) (*forms.Form, error) {
			return basicFormResponse(formID, "Pub Update Form"), nil
		},
		BatchUpdateFunc: func(_ context.Context, _ string, _ *forms.BatchUpdateFormRequest) (*forms.BatchUpdateFormResponse, error) {
			return &forms.BatchUpdateFormResponse{}, nil
		},
		SetPublishSettingsFunc: func(_ context.Context, _ string, isPublished bool, isAccepting bool) error {
			publishCalled = true
			gotPublished = isPublished
			gotAccepting = isAccepting
			return nil
		},
	}
	mockDrive := &testutil.MockDriveAPI{}

	r := testResource(mockForms, mockDrive)
	ctx := context.Background()

	// State: unpublished.
	state := buildState(t, map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, "pub-update-id"),
		"title":               tftypes.NewValue(tftypes.String, "Pub Update Form"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, false),
	})

	// Plan: now published and accepting.
	plan := buildPlan(t, map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, "pub-update-id"),
		"title":               tftypes.NewValue(tftypes.String, "Pub Update Form"),
		"published":           tftypes.NewValue(tftypes.Bool, true),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, true),
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
		t.Fatal("expected no errors during Update with publish settings change")
	}

	if !publishCalled {
		t.Fatal("expected SetPublishSettings to be called on update")
	}
	if !gotPublished {
		t.Fatal("expected published=true passed to SetPublishSettings")
	}
	if !gotAccepting {
		t.Fatal("expected accepting=true passed to SetPublishSettings")
	}
}

func TestUpdate_APIError_ReturnsDiagnostic(t *testing.T) {
	t.Parallel()

	mockForms := &testutil.MockFormsAPI{
		GetFunc: func(_ context.Context, _ string) (*forms.Form, error) {
			return nil, &client.APIError{
				StatusCode: 500,
				Message:    "server error during update read",
			}
		},
	}
	mockDrive := &testutil.MockDriveAPI{}

	r := testResource(mockForms, mockDrive)
	ctx := context.Background()

	state := buildState(t, map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, "update-err-id"),
		"title":               tftypes.NewValue(tftypes.String, "Update Err Form"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, false),
	})

	plan := buildPlan(t, map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, "update-err-id"),
		"title":               tftypes.NewValue(tftypes.String, "Updated Title"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, false),
	})

	resp := &resource.UpdateResponse{
		State: state,
	}

	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: state}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostic when Update pre-read fails")
	}

	found := false
	for _, d := range resp.Diagnostics.Errors() {
		if strings.Contains(d.Summary(), "Error Reading Google Form Before Update") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected diagnostic with summary containing 'Error Reading Google Form Before Update'")
	}
}

// ---------------------------------------------------------------------------
// F-005: Additional CRUD error path tests
// ---------------------------------------------------------------------------

func TestCreate_ContentJSON_ParseError(t *testing.T) {
	t.Parallel()

	mockForms := &testutil.MockFormsAPI{
		CreateFunc: func(_ context.Context, form *forms.Form) (*forms.Form, error) {
			return &forms.Form{
				FormId: "json-err-form-001",
				Info:   form.Info,
			}, nil
		},
	}
	mockDrive := &testutil.MockDriveAPI{}

	r := testResource(mockForms, mockDrive)
	ctx := context.Background()

	plan := buildPlan(t, map[string]tftypes.Value{
		"title":               tftypes.NewValue(tftypes.String, "Bad JSON Form"),
		"content_json":        tftypes.NewValue(tftypes.String, "invalid json {{"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, false),
	})

	resp := &resource.CreateResponse{
		State: emptyState(t),
	}

	r.Create(ctx, resource.CreateRequest{Plan: plan}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostic when content_json is invalid")
	}

	found := false
	for _, d := range resp.Diagnostics.Errors() {
		if strings.Contains(d.Summary(), "Error Parsing content_json") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected diagnostic with summary containing 'Error Parsing content_json'")
	}
}

func TestUpdate_BatchUpdateError_ReturnsDiagnostic(t *testing.T) {
	t.Parallel()

	mockForms := &testutil.MockFormsAPI{
		GetFunc: func(_ context.Context, formID string) (*forms.Form, error) {
			return basicFormResponse(formID, "Update Batch Err"), nil
		},
		BatchUpdateFunc: func(_ context.Context, _ string, _ *forms.BatchUpdateFormRequest) (*forms.BatchUpdateFormResponse, error) {
			return nil, &client.APIError{
				StatusCode: 500,
				Message:    "batch update server error",
			}
		},
	}
	mockDrive := &testutil.MockDriveAPI{}

	r := testResource(mockForms, mockDrive)
	ctx := context.Background()

	state := buildState(t, map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, "batch-err-id"),
		"title":               tftypes.NewValue(tftypes.String, "Update Batch Err"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, false),
	})

	plan := buildPlan(t, map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, "batch-err-id"),
		"title":               tftypes.NewValue(tftypes.String, "Updated Batch Err"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, false),
	})

	resp := &resource.UpdateResponse{
		State: state,
	}

	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: state}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostic when BatchUpdate fails during Update")
	}

	found := false
	for _, d := range resp.Diagnostics.Errors() {
		if strings.Contains(d.Summary(), "Error Updating Google Form") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected diagnostic with summary containing 'Error Updating Google Form'")
	}
}

func TestUpdate_WithContentJSON_Success(t *testing.T) {
	t.Parallel()

	var batchCalled bool
	contentJSON := `[{"title":"Q1","questionItem":{"question":{"textQuestion":{"paragraph":false}}}}]`

	getCalls := 0
	mockForms := &testutil.MockFormsAPI{
		GetFunc: func(_ context.Context, formID string) (*forms.Form, error) {
			getCalls++
			if getCalls == 1 {
				// Pre-update: form has 1 existing item.
				return &forms.Form{
					FormId:       formID,
					Info:         &forms.Info{Title: "JSON Update Form", DocumentTitle: "JSON Update Form"},
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
			// Post-update: return updated form.
			return basicFormResponse(formID, "JSON Update Form"), nil
		},
		BatchUpdateFunc: func(_ context.Context, _ string, _ *forms.BatchUpdateFormRequest) (*forms.BatchUpdateFormResponse, error) {
			batchCalled = true
			return &forms.BatchUpdateFormResponse{}, nil
		},
	}
	mockDrive := &testutil.MockDriveAPI{}

	r := testResource(mockForms, mockDrive)
	ctx := context.Background()

	state := buildState(t, map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, "json-update-id"),
		"title":               tftypes.NewValue(tftypes.String, "JSON Update Form"),
		"published":           tftypes.NewValue(tftypes.Bool, false),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, false),
	})

	plan := buildPlan(t, map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, "json-update-id"),
		"title":               tftypes.NewValue(tftypes.String, "JSON Update Form"),
		"content_json":        tftypes.NewValue(tftypes.String, contentJSON),
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
		t.Fatal("expected no errors during Update with content_json")
	}

	if !batchCalled {
		t.Fatal("expected BatchUpdate to be called for content_json update")
	}
}

func TestCreate_PublishSettingsError_ReturnsDiagnostic(t *testing.T) {
	t.Parallel()

	mockForms := &testutil.MockFormsAPI{
		CreateFunc: func(_ context.Context, form *forms.Form) (*forms.Form, error) {
			return &forms.Form{
				FormId: "pub-err-form-001",
				Info:   form.Info,
			}, nil
		},
		BatchUpdateFunc: func(_ context.Context, _ string, _ *forms.BatchUpdateFormRequest) (*forms.BatchUpdateFormResponse, error) {
			return &forms.BatchUpdateFormResponse{}, nil
		},
		GetFunc: func(_ context.Context, formID string) (*forms.Form, error) {
			return basicFormResponse(formID, "Pub Err Form"), nil
		},
		SetPublishSettingsFunc: func(_ context.Context, _ string, _ bool, _ bool) error {
			return &client.APIError{
				StatusCode: 500,
				Message:    "publish settings failed",
			}
		},
	}
	mockDrive := &testutil.MockDriveAPI{}

	r := testResource(mockForms, mockDrive)
	ctx := context.Background()

	plan := buildPlan(t, map[string]tftypes.Value{
		"title":               tftypes.NewValue(tftypes.String, "Pub Err Form"),
		"published":           tftypes.NewValue(tftypes.Bool, true),
		"accepting_responses": tftypes.NewValue(tftypes.Bool, false),
		"quiz":                tftypes.NewValue(tftypes.Bool, false),
	})

	resp := &resource.CreateResponse{
		State: emptyState(t),
	}

	r.Create(ctx, resource.CreateRequest{Plan: plan}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostic when SetPublishSettings fails")
	}

	found := false
	for _, d := range resp.Diagnostics.Errors() {
		if strings.Contains(d.Summary(), "Error Setting Publish Settings") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected diagnostic with summary containing 'Error Setting Publish Settings'")
	}
}
