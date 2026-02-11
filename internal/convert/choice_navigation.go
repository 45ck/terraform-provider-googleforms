// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package convert

import "fmt"

// ResolveChoiceOptionSectionIDs populates GoToSectionID for any choice option that
// references a section by GoToSectionKey.
//
// If allowMissing is true, unresolved keys are left as-is. This is useful when
// the target section is being created in the same apply and its item ID is not
// yet known.
func ResolveChoiceOptionSectionIDs(items []ItemModel, keyToID map[string]string, allowMissing bool) error {
	if len(items) == 0 {
		return nil
	}
	if keyToID == nil {
		keyToID = map[string]string{}
	}

	resolve := func(itemKey string, opts []ChoiceOption) error {
		for i := range opts {
			if opts[i].GoToSectionID != "" || opts[i].GoToSectionKey == "" {
				continue
			}
			id, ok := keyToID[opts[i].GoToSectionKey]
			if !ok || id == "" {
				if allowMissing {
					continue
				}
				return fmt.Errorf("item %q: unresolved go_to_section_key %q", itemKey, opts[i].GoToSectionKey)
			}
			opts[i].GoToSectionID = id
		}
		return nil
	}

	for i := range items {
		key := items[i].ItemKey
		switch {
		case items[i].MultipleChoice != nil:
			if err := resolve(key, items[i].MultipleChoice.Options); err != nil {
				return err
			}
		case items[i].Dropdown != nil:
			if err := resolve(key, items[i].Dropdown.Options); err != nil {
				return err
			}
		case items[i].Checkbox != nil:
			if err := resolve(key, items[i].Checkbox.Options); err != nil {
				return err
			}
		}
	}
	return nil
}

