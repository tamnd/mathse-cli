package mathse

import (
	"fmt"
	"strings"
	"time"
)

// Question is the record emitted for Math SE questions.
type Question struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Score       int    `json:"score"`
	ViewCount   int    `json:"view_count"`
	AnswerCount int    `json:"answer_count"`
	IsAnswered  bool   `json:"is_answered"`
	Tags        string `json:"tags"`
	CreatedAt   string `json:"created_at"`
	URL         string `json:"url"`
}

// Answer is the record emitted for answers to a Math SE question.
type Answer struct {
	ID         int    `json:"id"`
	Score      int    `json:"score"`
	IsAccepted bool   `json:"is_accepted"`
	CreatedAt  string `json:"created_at"`
	URL        string `json:"url"`
}

// Tag is the record emitted for Math SE tags.
type Tag struct {
	Name            string `json:"name"`
	Count           int    `json:"count"`
	IsRequired      bool   `json:"is_required"`
	IsModeratorOnly bool   `json:"is_moderator_only"`
}

// ── wire types ───────────────────────────────────────────────────────────────

type wireQuestion struct {
	QuestionID   int      `json:"question_id"`
	Title        string   `json:"title"`
	Score        int      `json:"score"`
	ViewCount    int      `json:"view_count"`
	AnswerCount  int      `json:"answer_count"`
	IsAnswered   bool     `json:"is_answered"`
	Tags         []string `json:"tags"`
	CreationDate int64    `json:"creation_date"`
	Link         string   `json:"link"`
}

type wireAnswer struct {
	AnswerID     int    `json:"answer_id"`
	Score        int    `json:"score"`
	IsAccepted   bool   `json:"is_accepted"`
	CreationDate int64  `json:"creation_date"`
	QuestionID   int    `json:"question_id"`
	Link         string `json:"link"`
}

type wireTag struct {
	Name            string `json:"name"`
	Count           int    `json:"count"`
	IsRequired      bool   `json:"is_required"`
	IsModeratorOnly bool   `json:"is_moderator_only"`
}

// ── converters ───────────────────────────────────────────────────────────────

func wireToQuestion(w wireQuestion) Question {
	return Question{
		ID:          w.QuestionID,
		Title:       w.Title,
		Score:       w.Score,
		ViewCount:   w.ViewCount,
		AnswerCount: w.AnswerCount,
		IsAnswered:  w.IsAnswered,
		Tags:        strings.Join(w.Tags, ","),
		CreatedAt:   isoDate(w.CreationDate),
		URL:         w.Link,
	}
}

func wireToAnswer(w wireAnswer, qid int) Answer {
	link := w.Link
	if link == "" {
		link = fmt.Sprintf("https://math.stackexchange.com/a/%d", w.AnswerID)
	}
	return Answer{
		ID:         w.AnswerID,
		Score:      w.Score,
		IsAccepted: w.IsAccepted,
		CreatedAt:  isoDate(w.CreationDate),
		URL:        link,
	}
}

func wireToTag(w wireTag) Tag {
	return Tag{
		Name:            w.Name,
		Count:           w.Count,
		IsRequired:      w.IsRequired,
		IsModeratorOnly: w.IsModeratorOnly,
	}
}

func isoDate(unix int64) string {
	return time.Unix(unix, 0).UTC().Format(time.RFC3339)
}
