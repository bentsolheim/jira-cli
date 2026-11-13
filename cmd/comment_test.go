package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bentsolheim/jira-cli/internal/jira"
)

func TestReadCommentBody(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "c.txt")
	if err := os.WriteFile(path, []byte("h3. Title\n\nline 2\n\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		file    string
		body    string
		bodySet bool
		stdin   string
		want    string
		wantErr string
	}{
		{name: "file", file: path, want: "h3. Title\n\nline 2"},
		{name: "file dash reads stdin", file: "-", stdin: "from stdin\n", want: "from stdin"},
		{name: "body flag", body: "inline", bodySet: true, want: "inline"},
		{name: "stdin default", stdin: "a\nb\n", want: "a\nb"},
		{name: "both flags", file: path, body: "x", bodySet: true, wantErr: "not both"},
		{name: "empty body flag", body: "  \n", bodySet: true, wantErr: "empty"},
		{name: "empty stdin", stdin: "", wantErr: "empty"},
		{name: "missing file", file: filepath.Join(dir, "nope"), wantErr: "reading comment file"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readCommentBody(tt.file, tt.body, tt.bodySet, strings.NewReader(tt.stdin))
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

type fakeCommentClient struct {
	key, body string
}

func (f *fakeCommentClient) AddComment(key, body string) (*jira.Comment, error) {
	f.key, f.body = key, body
	return &jira.Comment{ID: "77", Body: body, Created: "now", Author: &jira.User{DisplayName: "Bent"}}, nil
}

func runCommentCmd(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()
	commentFile, commentBody, commentDryRun = "", "", false
	t.Cleanup(func() {
		commentFile, commentBody, commentDryRun = "", "", false
		commentCmd.Flags().Lookup("body").Changed = false
		commentCmd.Flags().Lookup("file").Changed = false
		commentCmd.Flags().Lookup("dry-run").Changed = false
		rootCmd.SetArgs(nil)
		rootCmd.SetIn(nil)
		rootCmd.SetOut(nil)
	})
	var out bytes.Buffer
	rootCmd.SetIn(strings.NewReader(stdin))
	rootCmd.SetOut(&out)
	rootCmd.SetArgs(append([]string{"comment", "--url", "https://jira.example", "-o", "json"}, args...))
	err := rootCmd.Execute()
	return out.String(), err
}

func TestCommentDryRunDoesNotContactJira(t *testing.T) {
	orig := newCommentClient
	newCommentClient = func() (commentAdder, error) {
		t.Fatal("dry run must not create a Jira client")
		return nil, nil
	}
	t.Cleanup(func() { newCommentClient = orig })

	out, err := runCommentCmd(t, "*bold*\n", "MUP-2010", "--dry-run")
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON %q: %v", out, err)
	}
	if got["dryRun"] != true || got["issueKey"] != "MUP-2010" || got["body"] != "*bold*" {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestCommentPostsBody(t *testing.T) {
	fake := &fakeCommentClient{}
	orig := newCommentClient
	newCommentClient = func() (commentAdder, error) { return fake, nil }
	t.Cleanup(func() { newCommentClient = orig })

	out, err := runCommentCmd(t, "", "MUP-2010", "--body", "hello\n")
	if err != nil {
		t.Fatal(err)
	}
	if fake.key != "MUP-2010" || fake.body != "hello" {
		t.Errorf("posted key=%q body=%q", fake.key, fake.body)
	}
	if !strings.Contains(out, `"url": "https://jira.example/browse/MUP-2010?focusedCommentId=77#comment-77"`) {
		t.Errorf("unexpected output: %s", out)
	}
}
