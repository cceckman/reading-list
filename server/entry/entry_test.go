package entry_test

import (
	"bytes"
	"testing"

	"github.com/cceckman/reading-list/server/entry"
)

// This represents how yaml.v3 encodes this document, at time of writing.
// This includes e.g. canonicalized timestamps.
const basicDoc = `---
title: Some Document
summary: Test case for decode
date: 2021-09-08T00:00:01Z
draft: true
reading-list:
  source:
    text: Github
    uri: https://github.com/cceckman/reading-list
  author:
    text: cceckman
    uri: https://github.com/cceckman
  read: 2021-09-09T10:02:04Z
  added: 2021-09-09T10:02:04Z
  discovery:
    text: '@cceckman'
    uri: https://github.com/cceckman
other-hugo-field:
  - Item 1
  - Item 2
  - Item 3

---
Contents go here
`

func TestDecodeBasicDoc(t *testing.T) {
	rd := bytes.NewBufferString(basicDoc)
	front, err := entry.Read("entryid", rd)
	if err != nil {
		t.Fatal("unexpected error: ", err)
	}

	match := func(got string, want string) {
		if got != want {
			t.Errorf("unexpected property: got: %q want: %q", got, want)
		}
	}

	match(front.Id, "entryid")
	match(front.Title, "Some Document")
	match(front.Summary, "Test case for decode")
	match(front.Added.Format("2006-01-02"), "2021-09-09")
	match(front.Source.Text, "Github")
	match(front.Source.UrlString(), "https://github.com/cceckman/reading-list")
	match(front.Read.Format("2006-01-02"), "2021-09-09")

	if !front.Reviewed.IsZero() {
		t.Errorf("unexpected reviewed date: got: %q want: (zero)", front.Reviewed)
	}
}

// TODO: Test reencoding

func TestLaxDecode(t *testing.T) {
	rd := bytes.NewBufferString(`---
title: "Some document title"
date: 2021-09-08
---
`)
	front, err := entry.Read("entryid", rd)
	if err != nil {
		t.Fatal("unexpected error: ", err)
	}

	if got, want := front.Title, "Some document title"; got != want {
		t.Errorf("unexpected title: got: %q want: %q", got, want)
	}
}
