package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nwp/jenkins-cli/internal/auth"
	"github.com/nwp/jenkins-cli/internal/paths"
)

// Client is an authenticated Jenkins HTTP client.
type Client struct {
	baseURL    string
	authHeader string
	http       *http.Client

	crumb        string
	crumbField   string
	crumbFetched bool
}

// New creates a new Client for the given server URL and credentials.
func New(serverURL string, creds auth.Credentials) *Client {
	raw := fmt.Sprintf("%s:%s", creds.Username, creds.Token)
	return &Client{
		baseURL:    strings.TrimRight(serverURL, "/"),
		authHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte(raw)),
		http:       &http.Client{Timeout: 60 * time.Second},
	}
}

// JobURL returns the full URL for a job name, handling nested folder paths.
func (c *Client) JobURL(name string) string {
	return c.baseURL + paths.JobPath(name)
}

// NodeURL returns the full URL for a node/agent.
func (c *Client) NodeURL(name string) string {
	return c.baseURL + paths.NodeURL(name)
}

// ViewURL returns the full URL for a view.
func (c *Client) ViewURL(name string) string {
	return c.baseURL + paths.ViewURL(name)
}

func (c *Client) newRequest(method, fullURL string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.authHeader)
	return req, nil
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to %s failed: %w", req.URL, err)
	}
	return resp, nil
}

func apiError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 500))
	msg := strings.TrimSpace(string(body))
	if msg == "" {
		return fmt.Errorf("HTTP %d %s", resp.StatusCode, resp.Status)
	}
	return fmt.Errorf("HTTP %d %s: %s", resp.StatusCode, resp.Status, msg)
}

// Get performs a GET request to path (relative or full URL) and returns the response.
func (c *Client) Get(path string) (*http.Response, error) {
	fullURL := c.resolveURL(path)
	req, err := c.newRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		return nil, apiError(resp)
	}
	return resp, nil
}

// GetJSON performs a GET and decodes JSON into target.
func (c *Client) GetJSON(path string, target any) error {
	resp, err := c.Get(path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}

// GetText performs a GET and returns the response body as a string.
func (c *Client) GetText(path string) (string, error) {
	resp, err := c.Get(path)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return string(b), err
}

// GetRaw performs a GET and returns the response without error-checking the status code.
func (c *Client) GetRaw(path string) (*http.Response, error) {
	fullURL := c.resolveURL(path)
	req, err := c.newRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// Post performs a POST request with the given body and content type.
func (c *Client) Post(path string, body io.Reader, contentType string) (*http.Response, error) {
	if err := c.ensureCrumb(); err != nil {
		return nil, err
	}
	fullURL := c.resolveURL(path)
	req, err := c.newRequest(http.MethodPost, fullURL, body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if c.crumbField != "" {
		req.Header.Set(c.crumbField, c.crumb)
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		return nil, apiError(resp)
	}
	return resp, nil
}

// PostForm performs a POST with application/x-www-form-urlencoded body.
func (c *Client) PostForm(path string, params url.Values) (*http.Response, error) {
	return c.Post(path, strings.NewReader(params.Encode()), "application/x-www-form-urlencoded")
}

// PostXML performs a POST with an XML body.
func (c *Client) PostXML(path string, xml string) (*http.Response, error) {
	return c.Post(path, strings.NewReader(xml), "application/xml")
}

// PostEmpty performs a POST with no body (for action endpoints).
func (c *Client) PostEmpty(path string) (*http.Response, error) {
	return c.Post(path, nil, "")
}

type crumbResponse struct {
	CrumbRequestField string `json:"crumbRequestField"`
	Crumb             string `json:"crumb"`
}

func (c *Client) ensureCrumb() error {
	if c.crumbFetched {
		return nil
	}
	c.crumbFetched = true // mark early so we don't retry on failure

	fullURL := c.baseURL + "/crumbIssuer/api/json"
	req, err := c.newRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("crumb fetch failed: %w", err)
	}
	defer resp.Body.Close()

	// 404 means CSRF protection is disabled — that's fine.
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode >= 400 {
		return apiError(resp)
	}

	var cr crumbResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return fmt.Errorf("decode crumb response: %w", err)
	}
	c.crumbField = cr.CrumbRequestField
	c.crumb = cr.Crumb
	return nil
}

func (c *Client) resolveURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return c.baseURL + path
}

// BaseURL returns the normalized server base URL.
func (c *Client) BaseURL() string {
	return c.baseURL
}
