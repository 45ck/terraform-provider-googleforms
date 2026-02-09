// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceform

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ---------------------------------------------------------------------------
// 5. OptionsRequiredForChoiceValidator
// ---------------------------------------------------------------------------

func TestOptionsRequiredForChoice_WithOptions_Passes(t *testing.T) {
	t.Parallel()
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title": tftypes.NewValue(tftypes.String, "T"),
		"item":  itemListVal(mcItem("q1", "Pick one?", []string{"A", "B"}, nil)),
	})
	diags := runValidators(t, cfg, OptionsRequiredForChoiceValidator{})
	expectNoError(t, diags)
}

func TestOptionsRequiredForChoice_EmptyOptions_Error(t *testing.T) {
	t.Parallel()
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title": tftypes.NewValue(tftypes.String, "T"),
		"item":  itemListVal(mcItem("q1", "Pick one?", []string{}, nil)),
	})
	diags := runValidators(t, cfg, OptionsRequiredForChoiceValidator{})
	expectErrorContains(t, diags, "requires at least one option")
}

// ---------------------------------------------------------------------------
// 6. CorrectAnswerInOptionsValidator
// ---------------------------------------------------------------------------

func TestCorrectAnswerInOptions_ValidAnswer_Passes(t *testing.T) {
	t.Parallel()
	grading := &map[string]tftypes.Value{
		"points":             tftypes.NewValue(tftypes.Number, 10),
		"correct_answer":     tftypes.NewValue(tftypes.String, "B"),
		"feedback_correct":   tftypes.NewValue(tftypes.String, nil),
		"feedback_incorrect": tftypes.NewValue(tftypes.String, nil),
	}
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title": tftypes.NewValue(tftypes.String, "T"),
		"quiz":  tftypes.NewValue(tftypes.Bool, true),
		"item":  itemListVal(mcItem("q1", "Pick?", []string{"A", "B", "C"}, grading)),
	})
	diags := runValidators(t, cfg, CorrectAnswerInOptionsValidator{})
	expectNoError(t, diags)
}

func TestCorrectAnswerInOptions_InvalidAnswer_Error(t *testing.T) {
	t.Parallel()
	grading := &map[string]tftypes.Value{
		"points":             tftypes.NewValue(tftypes.Number, 10),
		"correct_answer":     tftypes.NewValue(tftypes.String, "D"),
		"feedback_correct":   tftypes.NewValue(tftypes.String, nil),
		"feedback_incorrect": tftypes.NewValue(tftypes.String, nil),
	}
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title": tftypes.NewValue(tftypes.String, "T"),
		"quiz":  tftypes.NewValue(tftypes.Bool, true),
		"item":  itemListVal(mcItem("q1", "Pick?", []string{"A", "B", "C"}, grading)),
	})
	diags := runValidators(t, cfg, CorrectAnswerInOptionsValidator{})
	expectErrorContains(t, diags, `correct_answer "D"`)
}

func TestCorrectAnswerInOptions_NoCorrectAnswer_Passes(t *testing.T) {
	t.Parallel()
	grading := &map[string]tftypes.Value{
		"points":             tftypes.NewValue(tftypes.Number, 10),
		"correct_answer":     tftypes.NewValue(tftypes.String, nil),
		"feedback_correct":   tftypes.NewValue(tftypes.String, nil),
		"feedback_incorrect": tftypes.NewValue(tftypes.String, nil),
	}
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title": tftypes.NewValue(tftypes.String, "T"),
		"quiz":  tftypes.NewValue(tftypes.Bool, true),
		"item":  itemListVal(mcItem("q1", "Pick?", []string{"A", "B"}, grading)),
	})
	diags := runValidators(t, cfg, CorrectAnswerInOptionsValidator{})
	expectNoError(t, diags)
}

// ---------------------------------------------------------------------------
// 7. GradingRequiresQuizValidator
// ---------------------------------------------------------------------------

func TestGradingRequiresQuiz_QuizTrueWithGrading_Passes(t *testing.T) {
	t.Parallel()
	grading := &map[string]tftypes.Value{
		"points":             tftypes.NewValue(tftypes.Number, 5),
		"correct_answer":     tftypes.NewValue(tftypes.String, nil),
		"feedback_correct":   tftypes.NewValue(tftypes.String, nil),
		"feedback_incorrect": tftypes.NewValue(tftypes.String, nil),
	}
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title": tftypes.NewValue(tftypes.String, "T"),
		"quiz":  tftypes.NewValue(tftypes.Bool, true),
		"item":  itemListVal(saItem("q1", "Q?", grading)),
	})
	diags := runValidators(t, cfg, GradingRequiresQuizValidator{})
	expectNoError(t, diags)
}

func TestGradingRequiresQuiz_QuizFalseWithGrading_Error(t *testing.T) {
	t.Parallel()
	grading := &map[string]tftypes.Value{
		"points":             tftypes.NewValue(tftypes.Number, 5),
		"correct_answer":     tftypes.NewValue(tftypes.String, nil),
		"feedback_correct":   tftypes.NewValue(tftypes.String, nil),
		"feedback_incorrect": tftypes.NewValue(tftypes.String, nil),
	}
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title": tftypes.NewValue(tftypes.String, "T"),
		"quiz":  tftypes.NewValue(tftypes.Bool, false),
		"item":  itemListVal(saItem("q1", "Q?", grading)),
	})
	diags := runValidators(t, cfg, GradingRequiresQuizValidator{})
	expectErrorContains(t, diags, "Grading requires quiz mode")
}

func TestGradingRequiresQuiz_NoGrading_Passes(t *testing.T) {
	t.Parallel()
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title": tftypes.NewValue(tftypes.String, "T"),
		"quiz":  tftypes.NewValue(tftypes.Bool, false),
		"item":  itemListVal(saItem("q1", "Q?", nil)),
	})
	diags := runValidators(t, cfg, GradingRequiresQuizValidator{})
	expectNoError(t, diags)
}

func TestGradingRequiresQuiz_MultipleChoice_Error(t *testing.T) {
	t.Parallel()
	grading := &map[string]tftypes.Value{
		"points":             tftypes.NewValue(tftypes.Number, 5),
		"correct_answer":     tftypes.NewValue(tftypes.String, "A"),
		"feedback_correct":   tftypes.NewValue(tftypes.String, nil),
		"feedback_incorrect": tftypes.NewValue(tftypes.String, nil),
	}
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title": tftypes.NewValue(tftypes.String, "T"),
		"quiz":  tftypes.NewValue(tftypes.Bool, false),
		"item":  itemListVal(mcItem("q1", "Pick?", []string{"A", "B"}, grading)),
	})
	diags := runValidators(t, cfg, GradingRequiresQuizValidator{})
	expectErrorContains(t, diags, "Grading requires quiz mode")
}

func TestGradingRequiresQuiz_Paragraph_Error(t *testing.T) {
	t.Parallel()
	grading := &map[string]tftypes.Value{
		"points":             tftypes.NewValue(tftypes.Number, 5),
		"correct_answer":     tftypes.NewValue(tftypes.String, nil),
		"feedback_correct":   tftypes.NewValue(tftypes.String, nil),
		"feedback_incorrect": tftypes.NewValue(tftypes.String, nil),
	}
	cfg := buildConfig(t, map[string]tftypes.Value{
		"title": tftypes.NewValue(tftypes.String, "T"),
		"quiz":  tftypes.NewValue(tftypes.Bool, false),
		"item":  itemListVal(paraItem("q1", "Essay?", grading)),
	})
	diags := runValidators(t, cfg, GradingRequiresQuizValidator{})
	expectErrorContains(t, diags, "Grading requires quiz mode")
}

// ---------------------------------------------------------------------------
// ConfigValidators wiring
// ---------------------------------------------------------------------------

func TestConfigValidators_ReturnsAllSeven(t *testing.T) {
	t.Parallel()
	r := &FormResource{}
	validators := r.ConfigValidators(context.Background())
	if len(validators) != 7 {
		t.Fatalf("expected 7 ConfigValidators, got %d", len(validators))
	}
}
