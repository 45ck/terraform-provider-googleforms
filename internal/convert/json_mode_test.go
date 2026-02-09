// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package convert

import (
	"testing"
)

// ---------------------------------------------------------------------------
// ParseDeclarativeJSON
// ---------------------------------------------------------------------------

func TestParseDeclarativeJSON_ValidItems(t *testing.T) {
	jsonStr := `[
		{
			"title": "Favorite color?",
			"questionItem": {
				"question": {
					"choiceQuestion": {
						"type": "RADIO",
						"options": [{"value": "Red"}, {"value": "Blue"}]
					}
				}
			}
		},
		{
			"title": "Your name?",
			"questionItem": {
				"question": {
					"textQuestion": {
						"paragraph": false
					}
				}
			}
		}
	]`

	items, err := ParseDeclarativeJSON(jsonStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("count = %d, want 2", len(items))
	}
	if items[0].Title != "Favorite color?" {
		t.Errorf("item[0].Title = %q", items[0].Title)
	}
	if items[1].Title != "Your name?" {
		t.Errorf("item[1].Title = %q", items[1].Title)
	}
}

func TestParseDeclarativeJSON_EmptyArray(t *testing.T) {
	items, err := ParseDeclarativeJSON("[]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("count = %d, want 0", len(items))
	}
}

func TestParseDeclarativeJSON_InvalidJSON_Error(t *testing.T) {
	_, err := ParseDeclarativeJSON("{not valid json")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseDeclarativeJSON_SingleItem(t *testing.T) {
	jsonStr := `[
		{
			"title": "Only question",
			"questionItem": {
				"question": {
					"textQuestion": {
						"paragraph": true
					}
				}
			}
		}
	]`

	items, err := ParseDeclarativeJSON(jsonStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("count = %d, want 1", len(items))
	}
	if items[0].Title != "Only question" {
		t.Errorf("title = %q", items[0].Title)
	}
}

// ---------------------------------------------------------------------------
// NormalizeJSON
// ---------------------------------------------------------------------------

func TestNormalizeJSON_SortsKeys(t *testing.T) {
	input := `{"zebra": 1, "alpha": 2, "middle": 3}`
	result, err := NormalizeJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `{"alpha":2,"middle":3,"zebra":1}`
	if result != expected {
		t.Errorf("result = %q, want %q", result, expected)
	}
}

func TestNormalizeJSON_StripExtraWhitespace(t *testing.T) {
	input := `{  "key" :   "value"  }`
	result, err := NormalizeJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `{"key":"value"}`
	if result != expected {
		t.Errorf("result = %q, want %q", result, expected)
	}
}

func TestNormalizeJSON_IdenticalInputsProduceSameOutput(t *testing.T) {
	input := `[{"a":1,"b":2}]`
	r1, err := NormalizeJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r2, err := NormalizeJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r1 != r2 {
		t.Errorf("outputs differ: %q vs %q", r1, r2)
	}
}

// ---------------------------------------------------------------------------
// HashJSON
// ---------------------------------------------------------------------------

func TestHashJSON_DeterministicOutput(t *testing.T) {
	input := `{"key": "value"}`
	h1, err := HashJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	h2, err := HashJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h1 != h2 {
		t.Errorf("hashes differ: %q vs %q", h1, h2)
	}
	if len(h1) != 64 { // SHA256 hex output is 64 chars
		t.Errorf("hash length = %d, want 64", len(h1))
	}
}

func TestHashJSON_DifferentInputsDifferentHashes(t *testing.T) {
	h1, err := HashJSON(`{"a": 1}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	h2, err := HashJSON(`{"a": 2}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h1 == h2 {
		t.Error("expected different hashes for different inputs")
	}
}

func TestHashJSON_EquivalentJSONsSameHash(t *testing.T) {
	// Same content, different formatting
	h1, err := HashJSON(`{"key":  "value",  "num": 1}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	h2, err := HashJSON(`{"num":1,"key":"value"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h1 != h2 {
		t.Errorf("expected same hash for equivalent JSON: %q vs %q", h1, h2)
	}
}

// ---------------------------------------------------------------------------
// DeclarativeJSONToRequests
// ---------------------------------------------------------------------------

func TestDeclarativeJSONToRequests_ConvertsToCreateItems(t *testing.T) {
	jsonStr := `[
		{
			"title": "Q1",
			"questionItem": {
				"question": {
					"textQuestion": {"paragraph": false}
				}
			}
		},
		{
			"title": "Q2",
			"questionItem": {
				"question": {
					"textQuestion": {"paragraph": true}
				}
			}
		}
	]`

	reqs, err := DeclarativeJSONToRequests(jsonStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reqs) != 2 {
		t.Fatalf("count = %d, want 2", len(reqs))
	}

	for i, req := range reqs {
		if req.CreateItem == nil {
			t.Fatalf("request[%d]: expected CreateItem", i)
		}
	}
	if reqs[0].CreateItem.Item.Title != "Q1" {
		t.Errorf("request[0] title = %q, want Q1", reqs[0].CreateItem.Item.Title)
	}
	if reqs[1].CreateItem.Item.Title != "Q2" {
		t.Errorf("request[1] title = %q, want Q2", reqs[1].CreateItem.Item.Title)
	}
}

func TestDeclarativeJSONToRequests_AssignsCorrectIndices(t *testing.T) {
	jsonStr := `[
		{"title": "A", "questionItem": {"question": {"textQuestion": {"paragraph": false}}}},
		{"title": "B", "questionItem": {"question": {"textQuestion": {"paragraph": false}}}},
		{"title": "C", "questionItem": {"question": {"textQuestion": {"paragraph": false}}}}
	]`

	reqs, err := DeclarativeJSONToRequests(jsonStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reqs) != 3 {
		t.Fatalf("count = %d, want 3", len(reqs))
	}

	for i, req := range reqs {
		idx := req.CreateItem.Location.Index
		if idx != int64(i) {
			t.Errorf("request[%d] index = %d, want %d", i, idx, i)
		}
	}
}
