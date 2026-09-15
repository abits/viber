// Package ghfetch centralizes the pieces of talking to GitHub that
// internal/templates and internal/updater used to each implement
// separately: an HTTP client carrying the optional GITHUB_TOKEN bearer
// header, owner/repo[@ref] parsing, and gzip+tar extraction (see UntarGz in
// untar.go).
package ghfetch

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// DefaultTimeout bounds every request made through a zero-value Client.
const DefaultTimeout = 30 * time.Second

// Client issues authenticated GET requests against GitHub.
//
// The zero value is ready to use: HTTP defaults to a client with
// DefaultTimeout, and Token defaults to the GITHUB_TOKEN environment
// variable read at call time.
type Client struct {
	// HTTP is the client used to send requests. nil uses a client with
	// DefaultTimeout.
	HTTP *http.Client
	// Token is sent as a bearer token when set. Empty reads GITHUB_TOKEN
	// from the environment at call time, so tests can use t.Setenv.
	Token string
}

func (c *Client) httpClient() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: DefaultTimeout}
}

func (c *Client) token() string {
	if c.Token != "" {
		return c.Token
	}
	return os.Getenv("GITHUB_TOKEN")
}

// NewRequest builds a GET request against url, setting the bearer token
// header when a token is available. Callers may add further headers (as
// Latest does for the GitHub REST API's Accept header) before calling Do.
func (c *Client) NewRequest(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if tok := c.token(); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	return req, nil
}

// Do sends req using the configured HTTP client, or a client with
// DefaultTimeout when none was set. The caller must close resp.Body.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.httpClient().Do(req)
}

// Get is NewRequest followed by Do. The caller must close resp.Body.
func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := c.NewRequest(ctx, url)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// ParseRepoSpec parses a GitHub repository reference of the form
// "owner/repo" or "owner/repo@ref". ref is "" when the input names none;
// callers that don't have a concept of a ref (such as viber update's fixed
// --repo flag) may simply discard it.
func ParseRepoSpec(s string) (owner, repo, ref string, err error) {
	if i := strings.Index(s, "@"); i >= 0 {
		ref = s[i+1:]
		s = s[:i]
		if ref == "" {
			return "", "", "", fmt.Errorf("%q: empty ref after '@'", s+"@")
		}
	}
	owner, repo, found := strings.Cut(s, "/")
	if !found || owner == "" || repo == "" {
		return "", "", "", fmt.Errorf("%q: want owner/repo[@ref]", s)
	}
	return owner, repo, ref, nil
}
