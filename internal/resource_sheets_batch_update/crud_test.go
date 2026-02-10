// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsbatchupdate

import "testing"

func TestDecodeBatchUpdateRequest_ArrayForm(t *testing.T) {
	t.Parallel()

	req, err := decodeBatchUpdateRequest(`[{"addSheet":{"properties":{"title":"RawTab"}}}]`)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(req.Requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(req.Requests))
	}
	if req.Requests[0].AddSheet == nil {
		t.Fatalf("expected AddSheet request, got %#v", req.Requests[0])
	}
}

func TestDecodeBatchUpdateRequest_ObjectForm(t *testing.T) {
	t.Parallel()

	req, err := decodeBatchUpdateRequest(`{"requests":[{"deleteSheet":{"sheetId":1}}]}`)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(req.Requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(req.Requests))
	}
	if req.Requests[0].DeleteSheet == nil {
		t.Fatalf("expected DeleteSheet request, got %#v", req.Requests[0])
	}
}

func TestHashID_IsDeterministic(t *testing.T) {
	t.Parallel()

	a := hashID("s1", "r1")
	b := hashID("s1", "r1")
	if a != b {
		t.Fatalf("expected deterministic hash, got %q vs %q", a, b)
	}
}
