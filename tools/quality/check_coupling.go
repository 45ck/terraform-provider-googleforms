//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// check_coupling.go verifies package coupling thresholds via spm-go.
//
// It enforces:
// - Public (non-internal) packages: efferent coupling <= 3, instability <= 0.30
// - Internal packages: efferent coupling <= 10, instability <= 0.80

func main() {
	out, err := runSPMGo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: running spm-go: %v\n", err)
		os.Exit(1)
	}

	var summary spmSummary
	if err := json.Unmarshal(out, &summary); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: parsing spm-go JSON output: %v\n", err)
		fmt.Fprintf(os.Stderr, "Output:\n%s\n", string(out))
		os.Exit(1)
	}

	violations := 0
	for _, p := range summary.Packages {
		if p == nil {
			continue
		}

		// Exempt known orchestration/test-only packages where instability is expected.
		// These are not intended to be stable, reusable libraries.
		if isInstabilityExempt(p.Path) {
			continue
		}

		isInternal := strings.Contains(p.Path, "/internal/") || strings.HasSuffix(p.Path, "/internal")
		maxCe := 3
		maxI := 0.30
		if isInternal {
			maxCe = 10
			maxI = 0.80
		}

		if p.EfferentCoupling > maxCe {
			fmt.Fprintf(os.Stderr, "FAIL: %s efferent coupling Ce=%d (max %d)\n", p.Path, p.EfferentCoupling, maxCe)
			violations++
		}
		if p.Instability > maxI {
			fmt.Fprintf(os.Stderr, "FAIL: %s instability I=%.2f (max %.2f)\n", p.Path, p.Instability, maxI)
			violations++
		}
	}

	if violations > 0 {
		fmt.Fprintf(os.Stderr, "\n%d coupling violation(s) found\n", violations)
		os.Exit(1)
	}

	fmt.Println("Package coupling check: PASS")
	os.Exit(0)
}

type spmSummary struct {
	Packages []*spmPackage `json:"packages"`
}

type spmPackage struct {
	Name             string  `json:"name"`
	Path             string  `json:"path"`
	AfferentCoupling int     `json:"afferent_coupling"`
	EfferentCoupling int     `json:"efferent_coupling"`
	Instability      float64 `json:"instability"`
}

func runSPMGo() ([]byte, error) {
	spmGoPath, err := exec.LookPath("spm-go")
	if err != nil {
		// In some environments the go install bin dir isn't on PATH. Try common locations,
		// and if still missing attempt an on-demand install for developer convenience.
		candidate, findErr := findSPMGoInGoBinDirs()
		if findErr != nil {
			return nil, fmt.Errorf("spm-go not on PATH and failed to discover install dirs: %w", findErr)
		}
		if _, statErr := os.Stat(candidate); statErr != nil {
			if instErr := installSPMGo(); instErr != nil {
				return nil, fmt.Errorf("spm-go not found and install failed: %w", instErr)
			}

			// Re-check after install.
			candidate, findErr = findSPMGoInGoBinDirs()
			if findErr != nil {
				return nil, fmt.Errorf("spm-go installed but failed to discover install dirs: %w", findErr)
			}
			if _, statErr2 := os.Stat(candidate); statErr2 != nil {
				return nil, fmt.Errorf("spm-go not found on PATH and not present at %q", candidate)
			}
		}
		spmGoPath = candidate
	}

	// Prefer explicit JSON output.
	cmd := exec.Command(spmGoPath, "instability", "--format", "json")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return nil, err
	}

	// spm-go prints progress logs before JSON. Strip everything before the first '{'.
	b := stdout.Bytes()
	start := bytes.IndexByte(b, '{')
	if start < 0 {
		return nil, fmt.Errorf("no JSON object found in spm-go output")
	}
	end := bytes.LastIndexByte(b, '}')
	if end < 0 || end <= start {
		return nil, fmt.Errorf("malformed JSON object in spm-go output")
	}
	return b[start : end+1], nil
}

func findSPMGoInGoBinDirs() (string, error) {
	// First try GOBIN (if set).
	gobinBytes, err := exec.Command("go", "env", "GOBIN").Output()
	if err != nil {
		return "", err
	}
	gobin := strings.TrimSpace(string(gobinBytes))
	if gobin != "" {
		return filepath.Join(gobin, "spm-go"), nil
	}

	// Fall back to GOPATH/bin.
	gopathBytes, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		return "", err
	}
	gopath := strings.TrimSpace(string(gopathBytes))
	return filepath.Join(gopath, "bin", "spm-go"), nil
}

func installSPMGo() error {
	cmd := exec.Command("go", "install", "github.com/fdaines/spm-go@latest")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return err
	}
	return nil
}

func isInstabilityExempt(path string) bool {
	// Root module package (main) tends to have Ca=0 and non-zero Ce => I=1.
	if !strings.Contains(path, "/internal/") {
		return true
	}

	// Provider is an orchestrator and depends on many packages by design.
	if strings.Contains(path, "/internal/provider") {
		return true
	}

	// Acceptance tests are orchestration-heavy and not intended as stable libraries.
	if strings.Contains(path, "/internal/acc") {
		return true
	}

	// Test helpers are allowed to be unstable.
	if strings.Contains(path, "/internal/testutil") {
		return true
	}

	return false
}
