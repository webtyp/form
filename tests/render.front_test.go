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

func TestTwoFormsSameModel_NoIDCollision(t *testing.T) {
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
