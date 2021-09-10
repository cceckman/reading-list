package entry_test

import (
	"bytes"
	"testing"

	"github.com/cceckman/reading-list/entry"
	"github.com/google/go-cmp/cmp"
)

// This represents how yaml.v3 encodes this document, at time of writing.
// This includes e.g. canonicalized timestamps.
const basicDoc = `---
title: Some Document
summary: Test case for decode
date: 2021-09-08T00:00:01Z
draft: true
reading-list:
  web-source:
    text: Github
    url: https://github.com/cceckman/reading-list
  author:
    text: cceckman
    url: https://github.com/cceckman
  read: 2021-09-09T10:02:04Z
  discovery:
    text: '@cceckman'
    url: https://github.com/cceckman
other-hugo-field:
  - Item 1
  - Item 2
  - Item 3

---
Contents go here
`

func TestDecodeBasicDoc(t *testing.T) {
	rd := bytes.NewBufferString(basicDoc)
	front, gotBody, err := entry.Read(rd)
	if err != nil {
		t.Fatal("unexpected error: ", err)
	}
	wantBody := []byte("Contents go here\n")
	if diff := cmp.Diff(gotBody, wantBody); diff != "" {
		t.Errorf("unexpected diff in contents: got: %q want: %q; diff: %v", gotBody, wantBody, diff)
	}

	match := func(got string, want string) {
		if got != want {
			t.Errorf("unexpected property: got: %q want: %q", got, want)
		}
	}

	match(front.Title, "Some Document")
	match(front.Summary, "Test case for decode")
	match(front.Date.Format("2006-01-02"), "2021-09-08")
	match(front.ReadingList.WebSource.Text, "Github")
	match(front.ReadingList.WebSource.RawUrl, "https://github.com/cceckman/reading-list")
	match(front.ReadingList.Read.Format("2006-01-02"), "2021-09-09")

	if !front.ReadingList.Reviewed.IsZero() {
		t.Errorf("unexpected reviewed date: got: %q want: (zero)", front.ReadingList.Reviewed)
	}
	if got, want := front.Draft, true; got != want {
		t.Errorf("unexpected draft status: got: %t want: %t", got, want)
	}

	wr := bytes.NewBuffer(nil)
	if _, err := front.WriteTo(wr); err != nil {
		t.Fatal(err)
	}
	if _, err := wr.Write(gotBody); err != nil {
		t.Fatal(err)
	}
	encoded := wr.String()
	if diff := cmp.Diff(encoded, basicDoc); diff != "" {
		t.Errorf("unexpected diff in docs: got:\n%v\nwant:\n%v\ndiff: %s", encoded, basicDoc, diff)
	}

}

func TestLaxDecode(t *testing.T) {
	rd := bytes.NewBufferString(`---
title: "Some document title"
date: 2021-09-08
---
`)
	front, _, err := entry.Read(rd)
	if err != nil {
		t.Fatal("unexpected error: ", err)
	}

	if got, want := front.Title, "Some document title"; got != want {
		t.Errorf("unexpected title: got: %q want: %q", got, want)
	}
	if got, want := front.Date.Format("2006-01-02"), "2021-09-08"; got != want {
		t.Errorf("unexpected date: got: %q want: %q", got, want)
	}
	if got, want := front.Draft, false; got != want {
		t.Errorf("unexpected draft status: got: %t want: %t", got, want)
	}
}

func TestEncodeRoundtrip(t *testing.T) {
	gave := entry.FrontMatter{
		Title:   "Some Document",
		Summary: "Test case for encode",
		Draft:   false,
		Other: map[string]interface{}{
			"some property": map[string]interface{}{
				"subproperty 1": 1,
				"subproperty 2": 2,
			},
		},
	}

	// We don't test that the yaml crate encodes a particular encoding / decoding;
	// we just check that we can recover the encoded object back from a buffer.
	const content = "Content goes here!"

	wr := bytes.NewBuffer(nil)
	if _, err := gave.WriteTo(wr); err != nil {
		t.Fatal(err)
	}
	if _, err := wr.WriteString(content); err != nil {
		t.Fatal(err)
	}

	gotFront, gotContent, err := entry.Read(wr)
	if err != nil {
		t.Fatal(err)
	}
	if string(gotContent) != content {
		t.Errorf("unexpected recovered contents: \n got: %q\nwant: %q", gotContent, content)
	}

	if diff := cmp.Diff(gotFront, &gave); diff != "" {
		t.Errorf("unexpected diff in recovered object: \n got: %+v\nwant: %+v\ndiff: %v", gotFront, gave, diff)
	}
}
