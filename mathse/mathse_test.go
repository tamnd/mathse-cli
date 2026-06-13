package mathse_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tamnd/mathse-cli/mathse"
)

func newTestClient(baseURL string) *mathse.Client {
	cfg := mathse.DefaultConfig()
	cfg.BaseURL = baseURL
	cfg.Rate = 0
	return mathse.NewClient(cfg)
}

func TestGetSendsUserAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		env := map[string]any{
			"items":            []any{},
			"has_more":         false,
			"quota_max":        300,
			"quota_remaining":  299,
		}
		_ = json.NewEncoder(w).Encode(env)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	_, err := c.Tags(context.Background(), mathse.TagsOptions{PageSize: 1})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetRetriesOn503(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		env := map[string]any{
			"items":            []any{},
			"has_more":         false,
			"quota_max":        300,
			"quota_remaining":  299,
		}
		_ = json.NewEncoder(w).Encode(env)
	}))
	defer srv.Close()

	cfg := mathse.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	cfg.Retries = 5
	c := mathse.NewClient(cfg)

	start := time.Now()
	_, err := c.Tags(context.Background(), mathse.TagsOptions{PageSize: 1})
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
	if time.Since(start) < 500*time.Millisecond {
		t.Error("retries did not back off")
	}
}

func TestSearchReturnsQuestions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("intitle") == "" {
			t.Error("search request missing intitle param")
		}
		env := map[string]any{
			"items": []map[string]any{
				{
					"question_id":   1234,
					"title":         "Is every prime greater than 2 odd?",
					"score":         42,
					"view_count":    1000,
					"answer_count":  3,
					"is_answered":   true,
					"tags":          []string{"prime-numbers", "number-theory"},
					"creation_date": int64(1609459200),
					"link":          "https://math.stackexchange.com/questions/1234",
				},
			},
			"has_more":        false,
			"quota_max":       300,
			"quota_remaining": 299,
		}
		_ = json.NewEncoder(w).Encode(env)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	qs, err := c.Search(context.Background(), mathse.SearchOptions{Query: "prime numbers", PageSize: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(qs) != 1 {
		t.Fatalf("got %d questions, want 1", len(qs))
	}
	q := qs[0]
	if q.ID != 1234 {
		t.Errorf("id = %d, want 1234", q.ID)
	}
	if q.Tags != "prime-numbers,number-theory" {
		t.Errorf("tags = %q, want %q", q.Tags, "prime-numbers,number-theory")
	}
}

func TestQuestionNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env := map[string]any{
			"items":           []any{},
			"has_more":        false,
			"quota_max":       300,
			"quota_remaining": 299,
		}
		_ = json.NewEncoder(w).Encode(env)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	_, err := c.Question(context.Background(), 99999)
	if err == nil {
		t.Fatal("expected ErrNotFound, got nil")
	}
}

func TestAnswers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env := map[string]any{
			"items": []map[string]any{
				{
					"answer_id":     5678,
					"score":         15,
					"is_accepted":   true,
					"creation_date": int64(1609459200),
					"question_id":   1234,
					"link":          "https://math.stackexchange.com/a/5678",
				},
			},
			"has_more":        false,
			"quota_max":       300,
			"quota_remaining": 299,
		}
		_ = json.NewEncoder(w).Encode(env)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	answers, err := c.Answers(context.Background(), 1234, mathse.AnswersOptions{PageSize: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(answers) != 1 {
		t.Fatalf("got %d answers, want 1", len(answers))
	}
	a := answers[0]
	if a.ID != 5678 {
		t.Errorf("id = %d, want 5678", a.ID)
	}
	if !a.IsAccepted {
		t.Error("expected IsAccepted=true")
	}
}

func TestTags(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env := map[string]any{
			"items": []map[string]any{
				{
					"name":               "calculus",
					"count":              50000,
					"is_required":        false,
					"is_moderator_only":  false,
				},
			},
			"has_more":        false,
			"quota_max":       300,
			"quota_remaining": 299,
		}
		_ = json.NewEncoder(w).Encode(env)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	tags, err := c.Tags(context.Background(), mathse.TagsOptions{PageSize: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 1 {
		t.Fatalf("got %d tags, want 1", len(tags))
	}
	if tags[0].Name != "calculus" {
		t.Errorf("name = %q, want %q", tags[0].Name, "calculus")
	}
}
