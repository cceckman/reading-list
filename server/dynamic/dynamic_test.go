package dynamic_test

import (
	"bytes"
	"testing"

	"github.com/cceckman/reading-list/server/dynamic"
	"github.com/cceckman/reading-list/server/entry"
	"github.com/cceckman/reading-list/server/paths"
)

var es []entry.Entry = []entry.Entry{
	{
		Id:    "first-entry",
		Title: "Something",
	},
	{
		Id:    "second-entry",
		Title: "Something else",
	}}

func TestRenderList(t *testing.T) {
	var b bytes.Buffer

	if err := dynamic.New().List(&b, paths.Default, es); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRenderEdit(t *testing.T) {
	var b bytes.Buffer

	if err := dynamic.New().Edit(&b, paths.Default, &es[0]); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TODO: test NewFromFs
