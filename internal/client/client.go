// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	drive "google.golang.org/api/drive/v3"
	forms "google.golang.org/api/forms/v1"
	"google.golang.org/api/option"
	sheets "google.golang.org/api/sheets/v4"
)

// requiredScopes lists the OAuth2 scopes needed for the provider.
var requiredScopes = []string{
	forms.FormsBodyScope,
	drive.DriveFileScope,
	sheets.SpreadsheetsScope,
}

// NewClient creates a new Client with real Google API implementations.
// credentials is the service account JSON content or empty for ADC.
// impersonateUser is the email to impersonate via domain-wide delegation.
func NewClient(
	ctx context.Context,
	credentials string,
	impersonateUser string,
) (*Client, error) {
	tokenSource, err := buildTokenSource(ctx, credentials, impersonateUser)
	if err != nil {
		return nil, fmt.Errorf("building token source: %w", err)
	}

	formsService, err := createFormsService(ctx, tokenSource)
	if err != nil {
		return nil, fmt.Errorf("creating forms service: %w", err)
	}

	driveService, err := createDriveService(ctx, tokenSource)
	if err != nil {
		return nil, fmt.Errorf("creating drive service: %w", err)
	}

	sheetsService, err := createSheetsService(ctx, tokenSource)
	if err != nil {
		return nil, fmt.Errorf("creating sheets service: %w", err)
	}

	retryCfg := DefaultRetryConfig()

	return &Client{
		Forms:  NewFormsAPIClient(formsService, retryCfg),
		Drive:  NewDriveAPIClient(driveService, retryCfg),
		Sheets: NewSheetsAPIClient(sheetsService, retryCfg),
	}, nil
}

// buildTokenSource creates an OAuth2 token source from credentials or ADC.
// It trusts that the caller (provider.go) has already resolved credentials
// from config and environment variables.
func buildTokenSource(
	ctx context.Context,
	credentials string,
	impersonateUser string,
) (oauth2.TokenSource, error) {
	if credentials != "" {
		return tokenSourceFromJSON(ctx, []byte(credentials), impersonateUser)
	}

	return tokenSourceFromADC(ctx)
}

// tokenSourceFromJSON creates a token source from service account JSON.
func tokenSourceFromJSON(
	ctx context.Context,
	credJSON []byte,
	impersonateUser string,
) (oauth2.TokenSource, error) {
	config, err := google.JWTConfigFromJSON(credJSON, requiredScopes...)
	if err != nil {
		return nil, fmt.Errorf("parsing service account credentials: %w", err)
	}

	if impersonateUser != "" {
		config.Subject = impersonateUser
	}

	return config.TokenSource(ctx), nil
}

// tokenSourceFromADC creates a token source from application default credentials.
func tokenSourceFromADC(ctx context.Context) (oauth2.TokenSource, error) {
	creds, err := google.FindDefaultCredentials(ctx, requiredScopes...)
	if err != nil {
		return nil, fmt.Errorf("finding default credentials: %w", err)
	}

	return creds.TokenSource, nil
}

// createFormsService creates a Google Forms API service.
func createFormsService(
	ctx context.Context,
	ts oauth2.TokenSource,
) (*forms.Service, error) {
	svc, err := forms.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("initializing forms service: %w", err)
	}

	return svc, nil
}

// createDriveService creates a Google Drive API service.
func createDriveService(
	ctx context.Context,
	ts oauth2.TokenSource,
) (*drive.Service, error) {
	svc, err := drive.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("initializing drive service: %w", err)
	}

	return svc, nil
}

// createSheetsService creates a Google Sheets API service.
func createSheetsService(
	ctx context.Context,
	ts oauth2.TokenSource,
) (*sheets.Service, error) {
	svc, err := sheets.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("initializing sheets service: %w", err)
	}

	return svc, nil
}
