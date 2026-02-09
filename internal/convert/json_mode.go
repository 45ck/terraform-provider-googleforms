// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package convert

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	forms "google.golang.org/api/forms/v1"
)

// ParseDeclarativeJSON parses a JSON string representing a list of Forms API
// Item objects. This is used for the content_json attribute.
func ParseDeclarativeJSON(jsonStr string) ([]*forms.Item, error) {
	var items []*forms.Item
	if err := json.Unmarshal([]byte(jsonStr), &items); err != nil {
		return nil, fmt.Errorf("parsing declarative JSON: %w", err)
	}
	return items, nil
}

// DeclarativeJSONToRequests parses declarative JSON and converts each item
// into a CreateItemRequest with sequential Location indices.
func DeclarativeJSONToRequests(jsonStr string) ([]*forms.Request, error) {
	items, err := ParseDeclarativeJSON(jsonStr)
	if err != nil {
		return nil, err
	}

	reqs := make([]*forms.Request, len(items))
	for i, item := range items {
		reqs[i] = &forms.Request{
			CreateItem: &forms.CreateItemRequest{
				Item:     item,
				Location: &forms.Location{Index: int64(i)},
			},
		}
	}
	return reqs, nil
}

// NormalizeJSON produces a canonical JSON string with sorted keys and no
// extraneous whitespace. This ensures two semantically identical JSON
// documents produce the same output.
func NormalizeJSON(jsonStr string) (string, error) {
	var v interface{}
	if err := json.Unmarshal([]byte(jsonStr), &v); err != nil {
		return "", fmt.Errorf("normalizing JSON: %w", err)
	}
	out, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("re-marshaling JSON: %w", err)
	}
	return string(out), nil
}

// HashJSON returns a hex-encoded SHA-256 hash of the normalized JSON. Two
// semantically identical JSON documents (differing only in whitespace or key
// order) will produce the same hash.
func HashJSON(jsonStr string) (string, error) {
	normalized, err := NormalizeJSON(jsonStr)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(normalized))
	return fmt.Sprintf("%x", sum), nil
}
