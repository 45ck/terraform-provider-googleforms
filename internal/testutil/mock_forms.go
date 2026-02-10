// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

// Package testutil provides testing utilities including mock API clients.
package testutil

import (
	"context"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
	forms "google.golang.org/api/forms/v1"
)

// MockFormsAPI is a configurable mock implementation of client.FormsAPI.
type MockFormsAPI struct {
	CreateFunc             func(ctx context.Context, form *forms.Form) (*forms.Form, error)
	GetFunc                func(ctx context.Context, formID string) (*forms.Form, error)
	BatchUpdateFunc        func(ctx context.Context, formID string, req *forms.BatchUpdateFormRequest) (*forms.BatchUpdateFormResponse, error)
	SetPublishSettingsFunc func(ctx context.Context, formID string, isPublished bool, isAccepting bool) error
}

var _ client.FormsAPI = &MockFormsAPI{}

func (m *MockFormsAPI) Create(ctx context.Context, form *forms.Form) (*forms.Form, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, form)
	}
	return &forms.Form{FormId: "mock-form-id", Info: form.Info}, nil
}

func (m *MockFormsAPI) Get(ctx context.Context, formID string) (*forms.Form, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, formID)
	}
	return &forms.Form{FormId: formID}, nil
}

func (m *MockFormsAPI) BatchUpdate(
	ctx context.Context,
	formID string,
	req *forms.BatchUpdateFormRequest,
) (*forms.BatchUpdateFormResponse, error) {
	if m.BatchUpdateFunc != nil {
		return m.BatchUpdateFunc(ctx, formID, req)
	}
	return &forms.BatchUpdateFormResponse{}, nil
}

func (m *MockFormsAPI) SetPublishSettings(
	ctx context.Context,
	formID string,
	isPublished bool,
	isAccepting bool,
) error {
	if m.SetPublishSettingsFunc != nil {
		return m.SetPublishSettingsFunc(ctx, formID, isPublished, isAccepting)
	}
	return nil
}
