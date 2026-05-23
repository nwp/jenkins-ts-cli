package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nwp/jenkins-cli/internal/auth"
	"github.com/nwp/jenkins-cli/internal/client"
)

func newTestClient(t *testing.T, srv *httptest.Server) *client.Client {
	t.Helper()
	return client.New(srv.URL, auth.Credentials{Username: "admin", Token: "token"})
}

func TestGet_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/json" {
			json.NewEncoder(w).Encode(map[string]string{"hello": "world"})
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	var result map[string]string
	if err := c.GetJSON("/api/json", &result); err != nil {
		t.Fatal(err)
	}
	if result["hello"] != "world" {
		t.Errorf("got %v", result)
	}
}

func TestGet_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.Get("/missing")
	if err == nil {
		t.Error("expected error on 404")
	}
}

func TestPost_WithCrumb(t *testing.T) {
	crumbFetched := false
	postCalled := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/crumbIssuer/api/json":
			crumbFetched = true
			json.NewEncoder(w).Encode(map[string]string{
				"crumbRequestField": "Jenkins-Crumb",
				"crumb":             "testcrumb",
			})
		case "/action":
			postCalled = true
			if r.Header.Get("Jenkins-Crumb") != "testcrumb" {
				http.Error(w, "missing crumb", http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	resp, err := c.PostEmpty("/action")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if !crumbFetched {
		t.Error("expected crumb to be fetched")
	}
	if !postCalled {
		t.Error("expected POST to be called")
	}
}

func TestJobURL_NestedFolders(t *testing.T) {
	c := client.New("http://jenkins.example.com", auth.Credentials{})
	got := c.JobURL("folder/sub/myjob")
	want := "http://jenkins.example.com/job/folder/job/sub/job/myjob"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPost_NoCrumb_When404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/crumbIssuer/api/json" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	resp, err := c.PostEmpty("/action")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}
