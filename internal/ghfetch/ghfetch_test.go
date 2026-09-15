package ghfetch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientGetSetsAuthorizationFromToken(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client(), Token: "secret"}
	resp, err := c.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if want := "Bearer secret"; gotAuth != want {
		t.Errorf("Authorization = %q, want %q", gotAuth, want)
	}
}

func TestClientGetFallsBackToEnvToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "from-env")
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client()}
	resp, err := c.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if want := "Bearer from-env"; gotAuth != want {
		t.Errorf("Authorization = %q, want %q", gotAuth, want)
	}
}

func TestClientGetOmitsAuthorizationWithoutToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client()}
	resp, err := c.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if gotAuth != "" {
		t.Errorf("Authorization = %q, want empty", gotAuth)
	}
}

func TestClientExplicitTokenWinsOverEnv(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "from-env")
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client(), Token: "explicit"}
	resp, err := c.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if want := "Bearer explicit"; gotAuth != want {
		t.Errorf("Authorization = %q, want %q", gotAuth, want)
	}
}

func TestNewRequestAllowsExtraHeaders(t *testing.T) {
	var gotAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAccept = r.Header.Get("Accept")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client()}
	req, err := c.NewRequest(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if want := "application/vnd.github+json"; gotAccept != want {
		t.Errorf("Accept = %q, want %q", gotAccept, want)
	}
}

func TestClientZeroValueUsesDefaultTimeout(t *testing.T) {
	var c Client
	if got := c.httpClient().Timeout; got != DefaultTimeout {
		t.Errorf("zero-value Client timeout = %v, want %v", got, DefaultTimeout)
	}
}

func TestClientCustomHTTPIsUsedAsIs(t *testing.T) {
	custom := &http.Client{Timeout: 5 * time.Second}
	c := &Client{HTTP: custom}
	if c.httpClient() != custom {
		t.Error("Client did not use the injected *http.Client")
	}
}

func TestParseRepoSpec(t *testing.T) {
	cases := []struct {
		name             string
		in               string
		owner, repo, ref string
		wantErr          bool
	}{
		{"owner/repo, no ref", "me/tpl", "me", "tpl", "", false},
		{"explicit ref", "me/tpl@v2", "me", "tpl", "v2", false},
		{"ref containing slashes", "me/tpl@feat/x", "me", "tpl", "feat/x", false},
		{"repo containing slashes, no ref", "o/r/extra", "o", "r/extra", "", false},
		{"empty ref after @ is rejected", "me/tpl@", "", "", "", true},
		{"missing slash", "metpl", "", "", "", true},
		{"empty owner", "/tpl", "", "", "", true},
		{"empty repo", "me/", "", "", "", true},
		{"empty string", "", "", "", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			owner, repo, ref, err := ParseRepoSpec(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseRepoSpec(%q) = %q,%q,%q; want error", tc.in, owner, repo, ref)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseRepoSpec(%q): %v", tc.in, err)
			}
			if owner != tc.owner || repo != tc.repo || ref != tc.ref {
				t.Errorf("ParseRepoSpec(%q) = %q,%q,%q; want %q,%q,%q",
					tc.in, owner, repo, ref, tc.owner, tc.repo, tc.ref)
			}
		})
	}
}
