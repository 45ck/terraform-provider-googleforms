// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package convert

import (
	"strings"
	"testing"

	forms "google.golang.org/api/forms/v1"
)

// ---------------------------------------------------------------------------
// ItemModelToCreateRequest - multiple_choice
// ---------------------------------------------------------------------------

func TestMultipleChoiceToRequest_Basic(t *testing.T) {
	t.Parallel()
	item := ItemModel{
		Title: "Favorite color?",
		MultipleChoice: &MultipleChoiceBlock{
			Options: []string{"Red", "Blue", "Green"},
		},
	}

	req, err := ItemModelToCreateRequest(item, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ci := req.CreateItem
	if ci == nil {
		t.Fatal("expected CreateItem to be set")
	}
	if ci.Item.Title != "Favorite color?" {
		t.Errorf("title = %q, want %q", ci.Item.Title, "Favorite color?")
	}
	if ci.Location == nil || ci.Location.Index != 0 {
		t.Errorf("location index = %v, want 0", ci.Location)
	}

	qi := ci.Item.QuestionItem
	if qi == nil {
		t.Fatal("expected QuestionItem to be set")
	}
	cq := qi.Question.ChoiceQuestion
	if cq == nil {
		t.Fatal("expected ChoiceQuestion to be set")
	}
	if cq.Type != "RADIO" {
		t.Errorf("type = %q, want RADIO", cq.Type)
	}
	if len(cq.Options) != 3 {
		t.Fatalf("options count = %d, want 3", len(cq.Options))
	}
	for i, want := range []string{"Red", "Blue", "Green"} {
		if cq.Options[i].Value != want {
			t.Errorf("option[%d] = %q, want %q", i, cq.Options[i].Value, want)
		}
	}
}

func TestMultipleChoiceToRequest_WithGrading(t *testing.T) {
	t.Parallel()
	item := ItemModel{
		Title: "Capital of France?",
		MultipleChoice: &MultipleChoiceBlock{
			Options: []string{"London", "Paris", "Berlin"},
			Grading: &GradingBlock{
				Points:            5,
				CorrectAnswer:     "Paris",
				FeedbackCorrect:   "Well done!",
				FeedbackIncorrect: "Try again.",
			},
		},
	}

	req, err := ItemModelToCreateRequest(item, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	q := req.CreateItem.Item.QuestionItem.Question
	g := q.Grading
	if g == nil {
		t.Fatal("expected Grading to be set")
	}
	if g.PointValue != 5 {
		t.Errorf("point_value = %d, want 5", g.PointValue)
	}
	if g.CorrectAnswers == nil || len(g.CorrectAnswers.Answers) != 1 {
		t.Fatal("expected one correct answer")
	}
	if g.CorrectAnswers.Answers[0].Value != "Paris" {
		t.Errorf("correct answer = %q, want Paris", g.CorrectAnswers.Answers[0].Value)
	}
	if g.WhenRight == nil || g.WhenRight.Text != "Well done!" {
		t.Errorf("feedback_correct = %v, want 'Well done!'", g.WhenRight)
	}
	if g.WhenWrong == nil || g.WhenWrong.Text != "Try again." {
		t.Errorf("feedback_incorrect = %v, want 'Try again.'", g.WhenWrong)
	}
}

func TestMultipleChoiceToRequest_Required(t *testing.T) {
	t.Parallel()
	item := ItemModel{
		Title: "Pick one",
		MultipleChoice: &MultipleChoiceBlock{
			Options:  []string{"A", "B"},
			Required: true,
		},
	}

	req, err := ItemModelToCreateRequest(item, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	q := req.CreateItem.Item.QuestionItem.Question
	if !q.Required {
		t.Error("expected Required to be true")
	}
}

// ---------------------------------------------------------------------------
// ItemModelToCreateRequest - short_answer
// ---------------------------------------------------------------------------

func TestShortAnswerToRequest_Basic(t *testing.T) {
	t.Parallel()
	item := ItemModel{
		Title:       "Your name?",
		ShortAnswer: &ShortAnswerBlock{},
	}

	req, err := ItemModelToCreateRequest(item, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	qi := req.CreateItem.Item.QuestionItem
	if qi == nil {
		t.Fatal("expected QuestionItem")
	}
	tq := qi.Question.TextQuestion
	if tq == nil {
		t.Fatal("expected TextQuestion")
	}
	if tq.Paragraph {
		t.Error("expected Paragraph to be false for short_answer")
	}
	if req.CreateItem.Item.Title != "Your name?" {
		t.Errorf("title = %q, want 'Your name?'", req.CreateItem.Item.Title)
	}
}

func TestShortAnswerToRequest_WithGrading(t *testing.T) {
	t.Parallel()
	item := ItemModel{
		Title: "Capital of Italy?",
		ShortAnswer: &ShortAnswerBlock{
			Grading: &GradingBlock{
				Points:          3,
				FeedbackCorrect: "Correct!",
			},
		},
	}

	req, err := ItemModelToCreateRequest(item, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g := req.CreateItem.Item.QuestionItem.Question.Grading
	if g == nil {
		t.Fatal("expected Grading to be set")
	}
	if g.PointValue != 3 {
		t.Errorf("point_value = %d, want 3", g.PointValue)
	}
	if g.WhenRight == nil || g.WhenRight.Text != "Correct!" {
		t.Errorf("feedback_correct = %v, want 'Correct!'", g.WhenRight)
	}
}

// ---------------------------------------------------------------------------
// ItemModelToCreateRequest - paragraph
// ---------------------------------------------------------------------------

func TestParagraphToRequest_Basic(t *testing.T) {
	t.Parallel()
	item := ItemModel{
		Title:     "Tell us about yourself",
		Paragraph: &ParagraphBlock{},
	}

	req, err := ItemModelToCreateRequest(item, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	qi := req.CreateItem.Item.QuestionItem
	if qi == nil {
		t.Fatal("expected QuestionItem")
	}
	tq := qi.Question.TextQuestion
	if tq == nil {
		t.Fatal("expected TextQuestion")
	}
	if !tq.Paragraph {
		t.Error("expected Paragraph to be true for paragraph type")
	}
}

func TestParagraphToRequest_WithGrading(t *testing.T) {
	t.Parallel()
	item := ItemModel{
		Title: "Essay question",
		Paragraph: &ParagraphBlock{
			Grading: &GradingBlock{
				Points:            10,
				FeedbackIncorrect: "Needs more detail.",
			},
		},
	}

	req, err := ItemModelToCreateRequest(item, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g := req.CreateItem.Item.QuestionItem.Question.Grading
	if g == nil {
		t.Fatal("expected Grading")
	}
	if g.PointValue != 10 {
		t.Errorf("point_value = %d, want 10", g.PointValue)
	}
	if g.WhenWrong == nil || g.WhenWrong.Text != "Needs more detail." {
		t.Errorf("feedback_incorrect = %v, want 'Needs more detail.'", g.WhenWrong)
	}
}

// ---------------------------------------------------------------------------
// ItemModelToCreateRequest - time
// ---------------------------------------------------------------------------

func TestTimeToRequest_Basic(t *testing.T) {
	t.Parallel()
	item := ItemModel{
		Title: "How long?",
		Time: &TimeBlock{
			Required: true,
			Duration: true,
		},
	}

	req, err := ItemModelToCreateRequest(item, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	q := req.CreateItem.Item.QuestionItem.Question
	if q.TimeQuestion == nil {
		t.Fatal("expected TimeQuestion")
	}
	if !q.Required {
		t.Error("expected Required=true")
	}
	if !q.TimeQuestion.Duration {
		t.Error("expected Duration=true")
	}
}

// ---------------------------------------------------------------------------
// ItemModelToCreateRequest - rating
// ---------------------------------------------------------------------------

func TestRatingToRequest_Basic(t *testing.T) {
	t.Parallel()
	item := ItemModel{
		Title: "Rate us",
		Rating: &RatingBlock{
			Required:         true,
			IconType:         "STAR",
			RatingScaleLevel: 5,
		},
	}

	req, err := ItemModelToCreateRequest(item, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	q := req.CreateItem.Item.QuestionItem.Question
	if q.RatingQuestion == nil {
		t.Fatal("expected RatingQuestion")
	}
	if !q.Required {
		t.Error("expected Required=true")
	}
	if q.RatingQuestion.IconType != "STAR" {
		t.Errorf("icon_type=%q, want STAR", q.RatingQuestion.IconType)
	}
	if q.RatingQuestion.RatingScaleLevel != 5 {
		t.Errorf("rating_scale_level=%d, want 5", q.RatingQuestion.RatingScaleLevel)
	}
}

// ---------------------------------------------------------------------------
// ItemModelToCreateRequest - text_item
// ---------------------------------------------------------------------------

func TestTextItemToRequest_Basic(t *testing.T) {
	t.Parallel()
	item := ItemModel{
		TextItem: &TextItemBlock{
			Title:       "Welcome",
			Description: "Please answer the questions below.",
		},
	}

	req, err := ItemModelToCreateRequest(item, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ci := req.CreateItem
	if ci == nil {
		t.Fatal("expected CreateItem")
	}
	if ci.Item.TextItem == nil {
		t.Fatal("expected TextItem")
	}
	if ci.Item.QuestionItem != nil {
		t.Fatal("expected QuestionItem to be nil for text_item")
	}
	if ci.Item.Title != "Welcome" {
		t.Errorf("title=%q, want Welcome", ci.Item.Title)
	}
	if ci.Item.Description != "Please answer the questions below." {
		t.Errorf("description=%q, want expected", ci.Item.Description)
	}
}

// ---------------------------------------------------------------------------
// ItemModelToCreateRequest - image
// ---------------------------------------------------------------------------

func TestImageToRequest_Basic(t *testing.T) {
	t.Parallel()
	item := ItemModel{
		Image: &ImageBlock{
			Title:       "Logo",
			Description: "Company logo",
			SourceURI:   "https://example.com/logo.png",
			AltText:     "Logo",
		},
	}

	req, err := ItemModelToCreateRequest(item, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ci := req.CreateItem
	if ci == nil {
		t.Fatal("expected CreateItem")
	}
	if ci.Item.ImageItem == nil || ci.Item.ImageItem.Image == nil {
		t.Fatal("expected ImageItem.Image")
	}
	if ci.Item.ImageItem.Image.SourceUri != "https://example.com/logo.png" {
		t.Errorf("source_uri=%q, want expected", ci.Item.ImageItem.Image.SourceUri)
	}
	if ci.Item.ImageItem.Image.AltText != "Logo" {
		t.Errorf("alt_text=%q, want Logo", ci.Item.ImageItem.Image.AltText)
	}
	if ci.Item.Title != "Logo" {
		t.Errorf("title=%q, want Logo", ci.Item.Title)
	}
	if ci.Item.Description != "Company logo" {
		t.Errorf("description=%q, want expected", ci.Item.Description)
	}
	if ci.Item.QuestionItem != nil {
		t.Fatal("expected QuestionItem to be nil for image")
	}
}

// ---------------------------------------------------------------------------
// ItemModelToCreateRequest - video
// ---------------------------------------------------------------------------

func TestVideoToRequest_Basic(t *testing.T) {
	t.Parallel()
	item := ItemModel{
		Video: &VideoBlock{
			Title:       "Intro",
			Description: "Watch this",
			YoutubeURI:  "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			Caption:     "Intro video",
		},
	}

	req, err := ItemModelToCreateRequest(item, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ci := req.CreateItem
	if ci == nil {
		t.Fatal("expected CreateItem")
	}
	if ci.Item.VideoItem == nil || ci.Item.VideoItem.Video == nil {
		t.Fatal("expected VideoItem.Video")
	}
	if ci.Item.VideoItem.Video.YoutubeUri != "https://www.youtube.com/watch?v=dQw4w9WgXcQ" {
		t.Errorf("youtube_uri=%q, want expected", ci.Item.VideoItem.Video.YoutubeUri)
	}
	if ci.Item.VideoItem.Caption != "Intro video" {
		t.Errorf("caption=%q, want expected", ci.Item.VideoItem.Caption)
	}
	if ci.Item.Title != "Intro" {
		t.Errorf("title=%q, want Intro", ci.Item.Title)
	}
	if ci.Item.Description != "Watch this" {
		t.Errorf("description=%q, want expected", ci.Item.Description)
	}
	if ci.Item.QuestionItem != nil {
		t.Fatal("expected QuestionItem to be nil for video")
	}
}

// ---------------------------------------------------------------------------
// ItemsToCreateRequests
// ---------------------------------------------------------------------------

func TestItemsToCreateRequests_MultipleItems_CorrectOrder(t *testing.T) {
	t.Parallel()
	items := []ItemModel{
		{Title: "Q1", ShortAnswer: &ShortAnswerBlock{}},
		{Title: "Q2", Paragraph: &ParagraphBlock{}},
		{Title: "Q3", MultipleChoice: &MultipleChoiceBlock{Options: []string{"A"}}},
	}

	reqs, err := ItemsToCreateRequests(items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reqs) != 3 {
		t.Fatalf("count = %d, want 3", len(reqs))
	}

	for i, req := range reqs {
		idx := req.CreateItem.Location.Index
		if idx != int64(i) {
			t.Errorf("request[%d] location index = %d, want %d", i, idx, i)
		}
	}

	if reqs[0].CreateItem.Item.Title != "Q1" {
		t.Errorf("first item title = %q, want Q1", reqs[0].CreateItem.Item.Title)
	}
	if reqs[2].CreateItem.Item.Title != "Q3" {
		t.Errorf("last item title = %q, want Q3", reqs[2].CreateItem.Item.Title)
	}
}

func TestItemsToCreateRequests_EmptyList(t *testing.T) {
	t.Parallel()
	reqs, err := ItemsToCreateRequests(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reqs) != 0 {
		t.Errorf("count = %d, want 0", len(reqs))
	}
}

func TestItemsToCreateRequests_SingleItem(t *testing.T) {
	t.Parallel()
	items := []ItemModel{
		{Title: "Only question", ShortAnswer: &ShortAnswerBlock{}},
	}

	reqs, err := ItemsToCreateRequests(items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reqs) != 1 {
		t.Fatalf("count = %d, want 1", len(reqs))
	}
	if reqs[0].CreateItem.Item.Title != "Only question" {
		t.Errorf("title = %q, want 'Only question'", reqs[0].CreateItem.Item.Title)
	}
}

// ---------------------------------------------------------------------------
// BuildDeleteRequests
// ---------------------------------------------------------------------------

func TestBuildDeleteRequests_MultipleItems_ReverseOrder(t *testing.T) {
	t.Parallel()
	reqs := BuildDeleteRequests(4)

	if len(reqs) != 4 {
		t.Fatalf("count = %d, want 4", len(reqs))
	}

	// Expect indices 3, 2, 1, 0 (reverse order to avoid index shifting)
	expected := []int64{3, 2, 1, 0}
	for i, req := range reqs {
		di := req.DeleteItem
		if di == nil {
			t.Fatalf("request[%d]: expected DeleteItem to be set", i)
		}
		if di.Location.Index != expected[i] {
			t.Errorf("request[%d] index = %d, want %d", i, di.Location.Index, expected[i])
		}
	}
}

func TestBuildDeleteRequests_EmptyList(t *testing.T) {
	t.Parallel()
	reqs := BuildDeleteRequests(0)
	if len(reqs) != 0 {
		t.Errorf("count = %d, want 0", len(reqs))
	}
}

// ---------------------------------------------------------------------------
// BuildUpdateInfoRequest
// ---------------------------------------------------------------------------

func TestBuildUpdateInfoRequest_TitleAndDescription(t *testing.T) {
	t.Parallel()
	req := BuildUpdateInfoRequest("My Form", "A description")

	ui := req.UpdateFormInfo
	if ui == nil {
		t.Fatal("expected UpdateFormInfo to be set")
	}
	if ui.Info.Title != "My Form" {
		t.Errorf("title = %q, want 'My Form'", ui.Info.Title)
	}
	if ui.Info.Description != "A description" {
		t.Errorf("description = %q, want 'A description'", ui.Info.Description)
	}
	if ui.UpdateMask != "title,description" {
		t.Errorf("update_mask = %q, want 'title,description'", ui.UpdateMask)
	}
}

func TestBuildUpdateInfoRequest_DescriptionOnly(t *testing.T) {
	t.Parallel()
	req := BuildUpdateInfoRequest("", "Only desc")

	ui := req.UpdateFormInfo
	if ui == nil {
		t.Fatal("expected UpdateFormInfo to be set")
	}
	if ui.Info.Title != "" {
		t.Errorf("title = %q, want empty", ui.Info.Title)
	}
	if ui.Info.Description != "Only desc" {
		t.Errorf("description = %q, want 'Only desc'", ui.Info.Description)
	}
}

// ---------------------------------------------------------------------------
// BuildQuizSettingsRequest
// ---------------------------------------------------------------------------

func TestBuildQuizSettingsRequest_EnableQuiz(t *testing.T) {
	t.Parallel()
	req := BuildQuizSettingsRequest(true)

	us := req.UpdateSettings
	if us == nil {
		t.Fatal("expected UpdateSettings to be set")
	}
	if us.Settings.QuizSettings == nil {
		t.Fatal("expected QuizSettings to be set")
	}
	if !us.Settings.QuizSettings.IsQuiz {
		t.Error("expected IsQuiz to be true")
	}
	if us.UpdateMask != "quizSettings.isQuiz" {
		t.Errorf("update_mask = %q, want 'quizSettings.isQuiz'", us.UpdateMask)
	}
}

func TestBuildQuizSettingsRequest_DisableQuiz(t *testing.T) {
	t.Parallel()
	req := BuildQuizSettingsRequest(false)

	us := req.UpdateSettings
	if us == nil {
		t.Fatal("expected UpdateSettings to be set")
	}
	if us.Settings.QuizSettings.IsQuiz {
		t.Error("expected IsQuiz to be false")
	}
}

func TestBuildEmailCollectionTypeRequest(t *testing.T) {
	req := BuildEmailCollectionTypeRequest("RESPONDER_INPUT")
	if req.UpdateSettings == nil {
		t.Fatal("expected UpdateSettings to be set")
	}
	if req.UpdateSettings.Settings == nil {
		t.Fatal("expected Settings to be set")
	}
	if req.UpdateSettings.Settings.EmailCollectionType != "RESPONDER_INPUT" {
		t.Fatalf("got %q, want %q", req.UpdateSettings.Settings.EmailCollectionType, "RESPONDER_INPUT")
	}
	if req.UpdateSettings.UpdateMask != "emailCollectionType" {
		t.Fatalf("got %q, want %q", req.UpdateSettings.UpdateMask, "emailCollectionType")
	}
}

func TestItemModelToCreateRequest_MultipleChoiceGrid(t *testing.T) {
	item := ItemModel{
		Title: "Grid",
		MultipleChoiceGrid: &MultipleChoiceGridBlock{
			QuestionText:     "Grid",
			Rows:             []string{"r1", "r2"},
			Columns:          []string{"c1", "c2"},
			Required:         true,
			ShuffleQuestions: true,
			ShuffleColumns:   true,
		},
	}

	req, err := ItemModelToCreateRequest(item, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.CreateItem == nil || req.CreateItem.Item == nil {
		t.Fatal("expected CreateItem to be set")
	}
	if req.CreateItem.Item.QuestionGroupItem == nil || req.CreateItem.Item.QuestionGroupItem.Grid == nil {
		t.Fatal("expected QuestionGroupItem grid to be set")
	}
	if req.CreateItem.Item.QuestionGroupItem.Grid.Columns == nil || req.CreateItem.Item.QuestionGroupItem.Grid.Columns.Type != "RADIO" {
		t.Fatalf("expected columns type RADIO, got %#v", req.CreateItem.Item.QuestionGroupItem.Grid.Columns)
	}
	if !req.CreateItem.Item.QuestionGroupItem.Grid.ShuffleQuestions {
		t.Fatal("expected ShuffleQuestions=true")
	}
}

func TestItemModelToCreateRequest_CheckboxGrid(t *testing.T) {
	item := ItemModel{
		Title: "Grid",
		CheckboxGrid: &CheckboxGridBlock{
			QuestionText:     "Grid",
			Rows:             []string{"r1"},
			Columns:          []string{"c1"},
			Required:         false,
			ShuffleQuestions: false,
			ShuffleColumns:   false,
		},
	}

	req, err := ItemModelToCreateRequest(item, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.CreateItem == nil || req.CreateItem.Item == nil {
		t.Fatal("expected CreateItem to be set")
	}
	if req.CreateItem.Item.QuestionGroupItem == nil || req.CreateItem.Item.QuestionGroupItem.Grid == nil {
		t.Fatal("expected QuestionGroupItem grid to be set")
	}
	if req.CreateItem.Item.QuestionGroupItem.Grid.Columns == nil || req.CreateItem.Item.QuestionGroupItem.Grid.Columns.Type != "CHECKBOX" {
		t.Fatalf("expected columns type CHECKBOX, got %#v", req.CreateItem.Item.QuestionGroupItem.Grid.Columns)
	}
}

func TestItemModelToCreateRequest_FileUpload_Errors(t *testing.T) {
	item := ItemModel{
		Title: "Upload",
		FileUpload: &FileUploadBlock{
			QuestionText: "Upload",
			Required:     true,
		},
	}

	_, err := ItemModelToCreateRequest(item, 0)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// helpers - ensure types compile (verify forms import used)
// ---------------------------------------------------------------------------

func TestItemModelToCreateRequest_NoQuestionBlock(t *testing.T) {
	t.Parallel()
	item := ItemModel{
		Title:          "Empty item",
		MultipleChoice: nil,
		ShortAnswer:    nil,
		Paragraph:      nil,
	}

	_, err := ItemModelToCreateRequest(item, 0)
	if err == nil {
		t.Fatal("expected error when no question block is set")
	}

	if !strings.Contains(err.Error(), "no question block") {
		t.Errorf("expected error to contain 'no question block', got: %v", err)
	}
}

func TestFormsImportUsed(t *testing.T) {
	t.Parallel()
	// This test simply ensures the forms import is used in tests.
	_ = &forms.Form{}
}
