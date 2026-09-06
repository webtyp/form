package form_test

import (
	"testing"

	"webtyp.com/fmt"
	"webtyp.com/form"
	"webtyp.com/widget"
)

// formHook is the class the <form> element itself carries — see
// anatomy_classes_test.go for why every one of these is derived.
var formHook = widget.NameField.Class(widget.PartForm).String()

func TestForm_SetClass(t *testing.T) {
	s := &submitStruct{}
	f, _ := form.New("app", s, &testIDGen{})

	f.SetClass("cms-form")

	html := f.String()
	// The <form> always carries the field anatomy hook; app classes from
	// SetClass are layered after it.
	expected := "class='" + formHook + " cms-form'"
	if !fmt.Contains(html, expected) {
		t.Errorf("Expected html to contain %q, got: %s", expected, html)
	}
}

func TestForm_SetClass_Append(t *testing.T) {
	form.SetGlobalClass("global-class")
	defer form.SetGlobalClass("") // Reset global state

	s := &submitStruct{}
	f, _ := form.New("app", s, &testIDGen{})
	f.SetClass("local-class")

	html := f.String()
	// New() uses globalClass as initial f.class. SetClass appends.
	// Initial f.class = "global-class"
	// After SetClass("local-class"), f.class = "global-class local-class"
	// The anatomy hook always comes first.
	expected := "class='" + formHook + " global-class local-class'"
	if !fmt.Contains(html, expected) {
		t.Errorf("Expected html to contain %q, got: %s", expected, html)
	}
}
