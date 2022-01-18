package dynamic_test

import (
	"bytes"
	"testing"

	"github.com/cceckman/reading-list/server/dynamic"
	"github.com/cceckman/reading-list/server/entry"
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

type paths struct{}

func (paths) Edit(id string) string {
	return "/edit"
}
func (paths) Save() string {
	return "/save"
}
func (paths) List() string {
	return "/"
}
func (paths) Share() string {
	return "/share"
}

func TestRenderList(t *testing.T) {
	var b bytes.Buffer

	if err := dynamic.List(&b, &paths{}, es); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRenderEdit(t *testing.T) {
	var b bytes.Buffer

	if err := dynamic.Edit(&b, &paths{}, &es[0]); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
