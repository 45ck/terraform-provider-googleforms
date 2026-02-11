// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type SweeperConfig struct {
	Credentials string

	// Prefix is matched against Drive file name.
	Prefix string

	// OlderThan filters out very recent resources to avoid racing active tests.
	OlderThan time.Duration

	SupportsAllDrives bool
}

func RunSweeper(ctx context.Context, cfg SweeperConfig) error {
	credJSON, err := resolveCredentialsValue(cfg.Credentials)
	if err != nil {
		return err
	}

	ts, err := tokenSourceFromJSON(ctx, []byte(credJSON))
	if err != nil {
		return err
	}

	svc, err := drive.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return fmt.Errorf("init drive service: %w", err)
	}

	prefix := strings.TrimSpace(cfg.Prefix)
	if prefix == "" {
		prefix = "tf-test-googleforms-"
	}
	olderThan := cfg.OlderThan
	if olderThan == 0 {
		olderThan = 1 * time.Hour
	}
	cutoff := time.Now().Add(-olderThan)

	// Drive query language: we keep it broad (by prefix) and then apply strict filtering client-side.
	q := fmt.Sprintf("name contains '%s' and trashed=false", escapeDriveQueryString(prefix))

	var deleted int
	pageToken := ""

	for {
		call := svc.Files.List().
			Context(ctx).
			Q(q).
			Fields("nextPageToken,files(id,name,mimeType,createdTime,trashed)").
			PageSize(1000).
			SupportsAllDrives(cfg.SupportsAllDrives).
			IncludeItemsFromAllDrives(cfg.SupportsAllDrives)

		if strings.TrimSpace(pageToken) != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("drive files.list: %w", err)
		}

		for _, f := range resp.Files {
			if f == nil || f.Id == "" {
				continue
			}
			if !strings.HasPrefix(f.Name, prefix) {
				continue
			}
			createdAt, perr := time.Parse(time.RFC3339, f.CreatedTime)
			if perr != nil {
				// If we cannot parse createdTime, skip rather than risk deleting something wrong.
				continue
			}
			if createdAt.After(cutoff) {
				continue
			}

			// Delete
			dcall := svc.Files.Delete(f.Id).
				Context(ctx).
				SupportsAllDrives(cfg.SupportsAllDrives)
			if derr := dcall.Do(); derr != nil {
				// Best-effort sweeper; keep going.
				continue
			}
			deleted++
		}

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	_ = deleted
	return nil
}

func tokenSourceFromJSON(ctx context.Context, credJSON []byte) (oauth2.TokenSource, error) {
	jwtCfg, err := google.JWTConfigFromJSON(credJSON, drive.DriveFileScope)
	if err != nil {
		return nil, fmt.Errorf("parse service account credentials: %w", err)
	}
	return jwtCfg.TokenSource(ctx), nil
}

func resolveCredentialsValue(val string) (string, error) {
	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return "", fmt.Errorf("missing credentials")
	}
	if strings.HasPrefix(trimmed, "{") {
		return trimmed, nil
	}
	// #nosec G304 -- path is an explicit test credential input.
	b, err := os.ReadFile(trimmed)
	if err != nil {
		return "", fmt.Errorf("read credentials file %q: %w", trimmed, err)
	}
	return string(b), nil
}

func escapeDriveQueryString(s string) string {
	// Drive query strings use single quotes; escape as \'
	return strings.ReplaceAll(s, "'", "\\'")
}
