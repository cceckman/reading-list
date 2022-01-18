package entry_test

/*
import (
	"bytes"
	"errors"
	"testing"

	"github.com/cceckman/reading-list/entry"
	"github.com/spf13/afero"
)

const testTitle = "Article title"
const testUrl = "https://github.com/cceckman/reading-list"

func TestEntryManagerCreate(t *testing.T) {
	fs := &afero.Afero{Fs: afero.NewMemMapFs()}
	em := entry.NewManager(fs)

	key, err := em.Create(testTitle, testUrl)
	if err != nil {
		t.Fatal(err)
	}

	file, err := fs.ReadFile(key + ".md")
	if err != nil {
		t.Fatal(err)
	}

	fm, body, err := entry.Read(bytes.NewBuffer(file))
	if err != nil {
		t.Fatal(err)
	}
	if len(body) != 0 {
		t.Error("unexpected body for entry: ", body)
	}
	if got, want := fm.Title, testTitle; got != want {
		t.Errorf("unexpected title: got: %q want: %q", got, want)
	}
	if fm.Date.IsZero() {
		t.Errorf("unexpected timestamp: got: %v want: (zero)", fm.Date)
	}
	if got, want := fm.ReadingList.WebSource.RawUrl, testUrl; got != want {
		t.Errorf("unexpected URL: got: %q want: %q", got, want)
	}
	if len(fm.ReadingList.WebSource.Text) == 0 {
		t.Errorf("unexpected source text: got: %q want: (nonempty string)", fm.ReadingList.WebSource.Text)
	}

}

func TestEntryManagerCreateDoesntClobber(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Create a collision
	const filename = "article-title.md"
	const contents = "This is some arbitrary data"
	if f, err := fs.Create(filename); err != nil {
		t.Fatal(err)
	} else if _, err := f.WriteString(contents); err != nil {
		t.Fatal(err)
	} else if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	em := entry.NewManager(fs)
	_, err := em.Create(testTitle, testUrl)
	if !errors.Is(err, afero.ErrFileExists) {
		t.Errorf("unexpected Create return; got: %v, wanted: %v", err, afero.ErrFileExists)
	}

	// Check that the file has the same contents
	got, err := afero.ReadFile(fs, filename)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != contents {
		t.Errorf("unexpected file contents: got: %q want: %q", string(got), contents)
	}
}

func TestEntryManagerGet(t *testing.T) {
	fs := afero.NewMemMapFs()
	const filename = "article-title.md"
	const contents = `---
title: This is an important article
reading-list:
  web-source:
    text: github.com
    url: https://github.com/cceckman
---
Some notes go here.
`
	if f, err := fs.Create(filename); err != nil {
		t.Fatal(err)
	} else if _, err := f.WriteString(contents); err != nil {
		t.Fatal(err)
	} else if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	em := entry.NewManager(fs)
	fm, err := em.Read("article-title")
	if err != nil {
		t.Fatal(err)
	}

	if got, want := fm.Title, "This is an important article"; got != want {
		t.Errorf("unexpected article title: got: %q want: %q", got, want)
	}

}

func TestEntryManagerGetBadData(t *testing.T) {
	fs := afero.NewMemMapFs()
	const filename = "article-title.md"
	const contents = "No front matter here."
	if f, err := fs.Create(filename); err != nil {
		t.Fatal(err)
	} else if _, err := f.WriteString(contents); err != nil {
		t.Fatal(err)
	} else if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	em := entry.NewManager(fs)
	_, err := em.Read("article-title")
	if err == nil {
		t.Fatalf("unexpected result from Read; got: %v want: some parse error", err)
	}
}

func TestEntryManagerGetNotFound(t *testing.T) {
	fs := afero.NewMemMapFs()

	em := entry.NewManager(fs)
	_, err := em.Read("article-title")
	if err == nil {
		t.Fatalf("unexpected result from Read; got: %v want: some parse error", err)
	}
}
*/
