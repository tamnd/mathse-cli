// Package mathse is the library behind the mathse command line:
// the HTTP client, request shaping, and the typed data models for Math Stack Exchange.
//
// The Stack Exchange API at api.stackexchange.com/2.3 is open and requires no key
// for up to 300 requests per day (10k with a registered app key).
// All requests use &site=math.
package mathse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	defaultBaseURL   = "https://api.stackexchange.com/2.3"
	defaultUserAgent = "mathse/dev (+https://github.com/tamnd/mathse-cli)"
	site             = "math"
)

// ErrNotFound is returned when the API returns an empty items list for a specific id.
var ErrNotFound = errors.New("not found")

// Config holds constructor parameters.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Retries   int
	Timeout   time.Duration
}

// DefaultConfig returns sensible defaults for the Stack Exchange API.
func DefaultConfig() Config {
	return Config{
		BaseURL:   defaultBaseURL,
		UserAgent: defaultUserAgent,
		Rate:      200 * time.Millisecond,
		Retries:   5,
		Timeout:   30 * time.Second,
	}
}

// Client talks to the Stack Exchange API.
type Client struct {
	http      *http.Client
	baseURL   string
	userAgent string
	rate      time.Duration
	retries   int
	last      time.Time
}

// NewClient returns a Client with the given config.
func NewClient(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	return &Client{
		http:      &http.Client{Timeout: cfg.Timeout},
		baseURL:   cfg.BaseURL,
		userAgent: cfg.UserAgent,
		rate:      cfg.Rate,
		retries:   cfg.Retries,
	}
}

// ── HTTP plumbing ─────────────────────────────────────────────────────────────

func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, rawURL)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", rawURL, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	if c.rate <= 0 {
		return
	}
	if wait := c.rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}

// ── generic decode ────────────────────────────────────────────────────────────

type envelope[T any] struct {
	Items       []T  `json:"items"`
	HasMore     bool `json:"has_more"`
	QuotaMax    int  `json:"quota_max"`
	QuotaRemain int  `json:"quota_remaining"`
}

func getItems[T any](c *Client, ctx context.Context, rawURL string) ([]T, error) {
	body, err := c.get(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	var env envelope[T]
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("decode %s: %w", rawURL, err)
	}
	return env.Items, nil
}

// ── public methods ────────────────────────────────────────────────────────────

// SearchOptions controls the /search endpoint.
type SearchOptions struct {
	Query    string
	Sort     string
	PageSize int
}

// Search searches questions by title text.
func (c *Client) Search(ctx context.Context, opts SearchOptions) ([]Question, error) {
	params := url.Values{}
	params.Set("site", site)
	params.Set("intitle", opts.Query)
	params.Set("pagesize", strconv.Itoa(clamp(opts.PageSize, 1, 100)))
	if opts.Sort != "" {
		params.Set("sort", opts.Sort)
		params.Set("order", sortOrder(opts.Sort))
	}
	rawURL := c.baseURL + "/search?" + params.Encode()
	wires, err := getItems[wireQuestion](c, ctx, rawURL)
	if err != nil {
		return nil, err
	}
	out := make([]Question, len(wires))
	for i, w := range wires {
		out[i] = wireToQuestion(w)
	}
	return out, nil
}

// QuestionsOptions controls the /questions endpoint.
type QuestionsOptions struct {
	Sort     string
	Tag      string
	PageSize int
}

// Questions lists questions on Math SE.
func (c *Client) Questions(ctx context.Context, opts QuestionsOptions) ([]Question, error) {
	params := url.Values{}
	params.Set("site", site)
	params.Set("pagesize", strconv.Itoa(clamp(opts.PageSize, 1, 100)))
	if opts.Sort != "" {
		params.Set("sort", opts.Sort)
		params.Set("order", sortOrder(opts.Sort))
	}
	if opts.Tag != "" {
		params.Set("tagged", opts.Tag)
	}
	rawURL := c.baseURL + "/questions?" + params.Encode()
	wires, err := getItems[wireQuestion](c, ctx, rawURL)
	if err != nil {
		return nil, err
	}
	out := make([]Question, len(wires))
	for i, w := range wires {
		out[i] = wireToQuestion(w)
	}
	return out, nil
}

// Question fetches a single question by id.
func (c *Client) Question(ctx context.Context, id int) (Question, error) {
	params := url.Values{}
	params.Set("site", site)
	rawURL := fmt.Sprintf("%s/questions/%d?%s", c.baseURL, id, params.Encode())
	wires, err := getItems[wireQuestion](c, ctx, rawURL)
	if err != nil {
		return Question{}, err
	}
	if len(wires) == 0 {
		return Question{}, ErrNotFound
	}
	return wireToQuestion(wires[0]), nil
}

// AnswersOptions controls the /questions/{id}/answers endpoint.
type AnswersOptions struct {
	PageSize int
}

// Answers fetches answers for a question.
func (c *Client) Answers(ctx context.Context, questionID int, opts AnswersOptions) ([]Answer, error) {
	params := url.Values{}
	params.Set("site", site)
	params.Set("pagesize", strconv.Itoa(clamp(opts.PageSize, 1, 100)))
	params.Set("sort", "votes")
	params.Set("order", "desc")
	rawURL := fmt.Sprintf("%s/questions/%d/answers?%s", c.baseURL, questionID, params.Encode())
	wires, err := getItems[wireAnswer](c, ctx, rawURL)
	if err != nil {
		return nil, err
	}
	out := make([]Answer, len(wires))
	for i, w := range wires {
		out[i] = wireToAnswer(w, questionID)
	}
	return out, nil
}

// TagsOptions controls the /tags endpoint.
type TagsOptions struct {
	PageSize int
	Search   string
}

// Tags lists or searches tags on Math SE.
func (c *Client) Tags(ctx context.Context, opts TagsOptions) ([]Tag, error) {
	params := url.Values{}
	params.Set("site", site)
	params.Set("pagesize", strconv.Itoa(clamp(opts.PageSize, 1, 100)))
	params.Set("order", "desc")
	params.Set("sort", "popular")
	if opts.Search != "" {
		params.Set("inname", opts.Search)
	}
	rawURL := c.baseURL + "/tags?" + params.Encode()
	wires, err := getItems[wireTag](c, ctx, rawURL)
	if err != nil {
		return nil, err
	}
	out := make([]Tag, len(wires))
	for i, w := range wires {
		out[i] = wireToTag(w)
	}
	return out, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// sortOrder returns the API order param for a sort value.
// newest/unanswered use desc; votes and activity also use desc.
func sortOrder(sort string) string {
	return "desc"
}
