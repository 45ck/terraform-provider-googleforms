// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"
)

var sweepFlag = flag.String("sweep", "", "run sweeper: all") //nolint:gochecknoglobals // test flag must be package-global for go test integration

func TestSweeper(t *testing.T) {
	if *sweepFlag == "" {
		t.Skip("skipping sweeper unless -sweep is set")
	}
	if *sweepFlag != "all" {
		t.Skip("unsupported sweep mode")
	}

	creds := os.Getenv("GOOGLE_CREDENTIALS")
	if creds == "" {
		t.Skip("skipping sweeper unless GOOGLE_CREDENTIALS is set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := RunSweeper(ctx, SweeperConfig{Credentials: creds}); err != nil {
		t.Fatalf("sweeper failed: %v", err)
	}
}
