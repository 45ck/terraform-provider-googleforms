---
page_title: "Importing Existing Resources"
description: |-
  Import existing Google Forms/Sheets/Drive resources into Terraform state.
---

# Importing Existing Resources

This provider supports importing existing resources so Terraform can take over lifecycle management.

## googleforms_form

Import format:

```bash
terraform import googleforms_form.example FORM_ID
```

Notes:

- After import, `item` blocks will be populated from the API response and `item_key` values are auto-generated as `item_0`, `item_1`, ...
- If you plan to use `update_strategy = "targeted"`, keep the imported `google_item_id` values so the provider can correlate items safely.

## googleforms_spreadsheet

```bash
terraform import googleforms_spreadsheet.example SPREADSHEET_ID
```

## googleforms_sheet

```bash
terraform import googleforms_sheet.example SPREADSHEET_ID#SHEET_ID
```

## googleforms_sheet_values

```bash
terraform import googleforms_sheet_values.example SPREADSHEET_ID#A1_RANGE
```

## googleforms_sheets_named_range

```bash
terraform import googleforms_sheets_named_range.example SPREADSHEET_ID#NAMED_RANGE_ID
```

## googleforms_sheets_protected_range

```bash
terraform import googleforms_sheets_protected_range.example SPREADSHEET_ID#PROTECTED_RANGE_ID
```

## googleforms_sheets_developer_metadata

```bash
terraform import googleforms_sheets_developer_metadata.example SPREADSHEET_ID#METADATA_ID
```

## googleforms_sheets_data_validation

```bash
terraform import googleforms_sheets_data_validation.example VALIDATION_ID
```

Note: `googleforms_sheets_data_validation` imports by `id` only. You must set `spreadsheet_id`, `range`, and `rule_json` in configuration after import.

## googleforms_sheets_conditional_format_rule

```bash
terraform import googleforms_sheets_conditional_format_rule.example SPREADSHEET_ID#SHEET_ID#INDEX
```

## googleforms_drive_permission

Import format:

```bash
terraform import googleforms_drive_permission.example FILE_ID#PERMISSION_ID
```

## googleforms_drive_folder

```bash
terraform import googleforms_drive_folder.example FOLDER_ID
```

## googleforms_drive_file

This resource does not create Drive files. To adopt an existing file:

```bash
terraform import googleforms_drive_file.example FILE_ID
```

