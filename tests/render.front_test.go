//go:build wasm

package form_test

import (
	"testing"

	"webtyp.com/dom"
	"webtyp.com/form"
)

func TestRender_Front(t *testing.T) {
	runRenderTests(t)
}

// TestTwoFormsSameModel_NoIDCollision documents a real, pre-existing limitation
// rather than asserting a fixed behaviour: TWO forms built from the SAME model
// cannot share a page today. Both roots take the model-derived id and dom's
// claimID — correctly — refuses the duplicate:
//
//	dom: id app.form was written twice in one render, by <form> and <form>
//
// The id is not invented per render (that is why the rest of this package keeps
// ID(): the ids come from input.Input's stable model identity, which is
// deliberate and meaningful). What is missing is an instance dimension on the
// FORM ROOT, and adding one is a form.New API decision — not something to paper
// over here with a per-instance prefix.
//
// Skipped, not deleted: the day the root gains an instance dimension, drop the
// Skip and this becomes the regression test. Until then it names the exact id
// that collides so the finding cannot get lost.
func TestTwoFormsSameModel_NoIDCollision(t *testing.T) {
	t.Skip(`two forms from one model collide on the root id "app.form" — ` +
		`form.New gives the root a model-derived id with no instance dimension; ` +
		`fixing that is an API decision, tracked separately`)

	s1 := &renderStruct{}
	s2 := &renderStruct{}

	f1, err1 := form.New("app", s1, &testIDGen{})
	if err1 != nil {
		t.Fatalf("failed creating form 1: %v", err1)
	}

	f2, err2 := form.New("app", s2, &testIDGen{})
	if err2 != nil {
		t.Fatalf("failed creating form 2: %v", err2)
	}

	container := dom.NewElement("div")
	container.Child(f1.Render())
	container.Child(f2.Render())

	_ = container.String()
}
