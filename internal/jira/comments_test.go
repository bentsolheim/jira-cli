package jira

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAddComment(t *testing.T) {
	body := "h3. Status\n* first\n* second"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.EscapedPath() != "/rest/api/2/issue/MUP-2010/comment" {
			t.Errorf("path = %s", r.URL.EscapedPath())
		}
		if got := r.Header.Get("Authorization"); got != "Bearer secret" {
			t.Errorf("Authorization = %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q", got)
		}
		var req CommentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decoding request: %v", err)
		}
		if req.Body != body {
			t.Errorf("body = %q, want %q", req.Body, body)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"10042","body":` + mustJSON(t, req.Body) + `,"created":"2026-11-13T11:00:00.000+0100","author":{"displayName":"Bent"}}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "secret", false)
	comment, err := c.AddComment("MUP-2010", body)
	if err != nil {
		t.Fatalf("AddComment: %v", err)
	}
	if comment.ID != "10042" || comment.Body != body || comment.Author == nil || comment.Author.DisplayName != "Bent" {
		t.Errorf("unexpected comment: %+v", comment)
	}
}

func TestAddCommentAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"errorMessages":["Issue Does Not Exist"]}`))
	}))
	defer srv.Close()

	_, err := NewClient(srv.URL, "secret", false).AddComment("MUP-0", "hi")
	if err == nil || !strings.Contains(err.Error(), "HTTP 404") {
		t.Fatalf("err = %v, want HTTP 404 error", err)
	}
}

func mustJSON(t *testing.T, v interface{}) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
