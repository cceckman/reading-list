package server_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cceckman/reading-list/server"
	"github.com/cceckman/reading-list/server/entry"
	"github.com/cceckman/reading-list/server/paths"
)

func TestEditById(t *testing.T) {
	em := entry.TestEntryManager{
		Items: make(map[string]*entry.Entry),
	}
	const id string = "id1"
	em.Items[id] = &entry.Entry{
		Id:    id,
		Title: "Entry title",
	}
	t.Logf("items: %+v", em)
	p := paths.Default

	m := server.New(p, &em)

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

	m := server.New(p, &em)

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

	m := server.New(p, &em)

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
