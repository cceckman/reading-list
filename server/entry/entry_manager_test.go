package entry_test

import (
	"bytes"
	"os"
	"path"
	"testing"
	"time"

	"github.com/cceckman/reading-list/server/entry"
	"github.com/cceckman/reading-list/server/fs"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

const fakeEntry string = `
---
title: "Efficient String Matching: An Aid to Bibliographic Search"
date: 2021-08-20
draft: false
summary: >
  The Aho-Corasick paper on string matching: a way to construct a state machine
  for finding all occurrences of "needles" in some "haystack".
reading-list:
  added: 2021-08-20
  read: 2021-08-20
  source:
    Text: Communications of the ACM
    url: https://doi.org/10.1145/360825.360855
  author:
    text: "Alfred V. Aho and Margaret J. Corasick"
---

If you're trying to match [regular] patterns - like a set of strings - it's hard
to beat a state machine. And these are some good state machines.
(So are [these].)

[regular]: https://en.wikipedia.org/wiki/Regular_language
[these]: https://twitter.com/happyautomata
`

const fakeId = "aho-corasick"

func fakeDirectory(t *testing.T) fs.CreateFS {
	dir, err := os.MkdirTemp(os.TempDir(), "")
	if err != nil {
		t.Fatal("failed to make temp directory: ", err)
	}
	if err := os.WriteFile(path.Join(dir, fakeId+".md"), []byte(fakeEntry), 0755); err != nil {
		t.Fatal("failed to write temp file: ", err)
	}
	t.Logf("temporary directory for test: %s", dir)

	return fs.NativeFS(dir)
}

func TestManagerRead(t *testing.T) {
	m, err := entry.NewManager(fakeDirectory(t))
	if err != nil {
		t.Fatal(err)
	}

	e, err := m.Read(fakeId)
	if err != nil {
		t.Fatal(err)
	}

	// Most of the logic is in entry.Read; compare with its outputs.
	rd := bytes.NewBufferString(fakeEntry)
	eRef, err := entry.Read(fakeId, rd)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(e, &eRef, cmpopts.IgnoreUnexported(entry.Entry{})); diff != "" {
		t.Error("unexpected diffs when read: ", diff)
	}

}

func TestManagerList(t *testing.T) {
	m, err := entry.NewManager(fakeDirectory(t))
	if err != nil {
		t.Fatal(err)
	}

	es, err := m.List(100)
	if err != nil {
		t.Fatal(err)
	}
	if len(es) != 1 {
		t.Fatalf("unexpected number of entries: got: %d", len(es))
	}
	rd := bytes.NewBufferString(fakeEntry)
	eRef, err := entry.Read(fakeId, rd)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(es[0], &eRef, cmpopts.IgnoreUnexported(entry.Entry{})); diff != "" {
		t.Error("unexpected diffs when read: ", diff)
	}
}

func TestManagerCreate(t *testing.T) {
	const newFakeId = "new-entry"
	now := time.Now()
	e := entry.Entry{
		Id:      newFakeId,
		Title:   "A new fake entry",
		Summary: "New fake entry for testing",
		Source: entry.Source{
			Text: "entry_manager_test.go",
		},
		Author: &entry.Source{
			Text: "cceckman",
			Uri:  "https://github.com/cceckman",
		},
		Added: entry.Date{now},
	}
	dir := fakeDirectory(t)
	m, err := entry.NewManager(dir)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := m.Update(e); err != nil {
		t.Fatal(err)
	}

	f, err := dir.Open(newFakeId + ".md")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	eRecovered, err := entry.Read(newFakeId, f)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(eRecovered, e, cmpopts.IgnoreUnexported(entry.Entry{})); diff != "" {
		t.Error("unexpected diffs when read: ", diff)
	}

}

func TestManagerUpdate(t *testing.T) {
	m, err := entry.NewManager(fakeDirectory(t))
	if err != nil {
		t.Fatal(err)
	}

	var eRef entry.Entry
	if e, err := m.Read(fakeId); err != nil {
		t.Fatal(err)
	} else {
		eRef = *e
	}

	now := entry.Date{time.Now()}
	newEntry := eRef
	newEntry.Title = "Modified title"
	newEntry.Summary = "Modified summary"
	newEntry.Read = now

	if _, err := m.Update(newEntry); err != nil {
		t.Error(err)
	}

	if e, err := m.Read(fakeId); err != nil {
		t.Error(err)
	} else if diff := cmp.Diff(e, &newEntry, cmpopts.IgnoreUnexported(entry.Entry{})); diff != "" {
		t.Error("unexpected diffs when read: ", diff)
	} else if diff := cmp.Diff(e.Content(), eRef.Content()); diff != "" {
		t.Error("unexpected diffs in content: ", diff)
	} else if got, want := e.Read.Format(entry.DateFormat), now.Format(entry.DateFormat); got != want {
		t.Errorf("unexpected diffs in returned time: got: %v want: %v", got, want)
	} else if diff := cmp.Diff(e.Author, eRef.Author); diff != "" {
		t.Errorf("unexpected diffs in returned author: got: %v want: %v", got, want)
	}
}
