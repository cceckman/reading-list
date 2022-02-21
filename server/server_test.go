package server_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/cceckman/reading-list/server"
	"github.com/cceckman/reading-list/server/dynamic"
	"github.com/cceckman/reading-list/server/entry"
	"github.com/cceckman/reading-list/server/paths"
	"github.com/cceckman/reading-list/server/static"
)

func TestList(t *testing.T) {
	em := entry.TestEntryManager{
		Items: make(map[string]*entry.Entry),
	}
	em.Items["id1"] = &entry.Entry{
		Id:    "id1",
		Title: "Entry title",
	}
	em.Items["id2"] = &entry.Entry{
		Id:    "id2",
		Title: "Entry 2 title",
	}

	p := paths.Default

	m := server.New(p, &em, dynamic.New(), static.Files)

	req := httptest.NewRequest("GET", "/", nil)

	var buf bytes.Buffer
	resp := httptest.ResponseRecorder{
		Body: &buf,
	}
	m.ServeHTTP(&resp, req)

	if got, want := resp.Code, http.StatusOK; got != want {
		t.Errorf("wrong response code: got: %v want: %v", got, want)
	}
	if !resp.Flushed {
		t.Error("response not flushed")
	}
	body := buf.String()
	for _, item := range em.Items {
		if !strings.Contains(body, item.Title) {
			t.Errorf("missing entry title: %q", item.Title)
			t.Logf("Contents:\n---\n%+v", body)
		}
	}

}

func TestEditById(t *testing.T) {
	em := entry.TestEntryManager{
		Items: make(map[string]*entry.Entry),
	}
	const id string = "id1"
	em.Items[id] = &entry.Entry{
		Id:    id,
		Title: "Entry title",
	}
	p := paths.Default

	m := server.New(p, &em, dynamic.New(), static.Files)

	req := httptest.NewRequest("GET", p.Edit(), nil)
	{
		q := req.URL.Query()
		q.Add("id", id)
		req.URL.RawQuery = q.Encode()
	}

	var buf bytes.Buffer
	resp := httptest.ResponseRecorder{
		Body: &buf,
	}
	m.ServeHTTP(&resp, req)

	if got, want := resp.Code, http.StatusOK; got != want {
		t.Errorf("wrong response code: got: %v want: %v", got, want)
	}
	if !resp.Flushed {
		t.Error("response not flushed")
	}
	body := buf.String()
	if !strings.Contains(body, em.Items[id].Title) {
		t.Error("edit contents does not contain entry title")
		t.Logf("Contents:\n---\n%+v", body)
	}
}

func TestEditMissingId(t *testing.T) {
	em := entry.TestEntryManager{
		Items: make(map[string]*entry.Entry),
	}
	const id = "id1"
	p := paths.Default

	m := server.New(p, &em, dynamic.New(), static.Files)

	req := httptest.NewRequest("GET", p.Edit(), nil)
	{
		q := req.URL.Query()
		q.Add("id", id)
		req.URL.RawQuery = q.Encode()
	}

	var resp httptest.ResponseRecorder
	m.ServeHTTP(&resp, req)

	if got, want := resp.Code, http.StatusNotFound; got != want {
		t.Errorf("wrong response code: got: %v want: %v", got, want)
	}
	if !resp.Flushed {
		t.Error("response not flushed")
	}
}

func TestEditNew(t *testing.T) {
	em := entry.TestEntryManager{
		Items: make(map[string]*entry.Entry),
	}
	p := paths.Default

	m := server.New(p, &em, dynamic.New(), static.Files)

	req := httptest.NewRequest("GET", p.Edit(), nil)

	var resp httptest.ResponseRecorder
	m.ServeHTTP(&resp, req)

	if got, want := resp.Code, http.StatusOK; got != want {
		t.Errorf("wrong response code: got: %v want: %v", got, want)
	}
	if !resp.Flushed {
		t.Error("response not flushed")
	}
}

func TestSaveOk(t *testing.T) {
	em := entry.TestEntryManager{
		Items: make(map[string]*entry.Entry),
	}

	const id = "id1"
	const newTitle = "New Entry Title"
	em.Items[id] = &entry.Entry{
		Id:    "id1",
		Title: "Entry title",
	}

	form := make(url.Values)
	form["id"] = []string{id}
	form["title"] = []string{newTitle}
	form["added"] = []string{"2023-02-06"}

	p := paths.Default
	m := server.New(p, &em, dynamic.New(), static.Files)
	s := httptest.NewServer(m)
	client := s.Client()

	// Tweak the client to avoid following redirects.
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	url := s.URL + p.Save()
	resp, err := client.PostForm(url, form)
	if err != nil {
		t.Fatalf("error in processing request: %v", err)
	}

	// Expect a redirect to the list.
	if got, want := resp.StatusCode, http.StatusSeeOther; got != want {
		t.Errorf("wrong response code: got: %v want: %v", got, want)
	}
	loc, err := resp.Location()
	if err != nil {
		t.Errorf("invalid location for response: %v", err)
	}
	if got, want := loc.Path, p.List(); got != want {
		t.Errorf("wrong resultant response path: got: %q want: %q", got, want)
	}
	if got, want := loc.Query().Get("done"), id; got != want {
		t.Errorf("missing or incorrect completion notification: got: %q want: %q", got, want)
	}

	if got, want := em.Items[id].Title, newTitle; got != want {
		t.Errorf("didn't save title: got: %q want: %q", got, want)
	}
}
