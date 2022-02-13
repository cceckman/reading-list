package entry_test

import (
	"bytes"
	"io/fs"
	"os"
	"path"
	"testing"

	"github.com/cceckman/reading-list/server/entry"
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

func fakeDirectory(t *testing.T) fs.FS {
	dir, err := os.MkdirTemp(os.TempDir(), "")
	if err != nil {
		t.Fatal("failed to make temp directory: ", err)
	}
	if err := os.WriteFile(path.Join(dir, fakeId+".md"), []byte(fakeEntry), 0755); err != nil {
		t.Fatal("failed to write temp file: ", err)
	}

	return os.DirFS(dir)
}

func TestManagerRead(t *testing.T) {
	m := entry.NewManager(fakeDirectory(t))
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

	if diff := cmp.Diff(e, eRef, cmpopts.IgnoreUnexported(entry.Entry{})); diff != "" {
		t.Error("unexpected diffs when read: ", diff)
	}

}
