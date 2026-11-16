//go:build wasm

package form_test

import (
	"syscall/js"
	"testing"

	"webtyp.com/dom"
	"webtyp.com/form"
)

// pickRadio presses option 1 (by pointer when byPointer, else only the change a
// keyboard arrow causes) and reports which element document.activeElement is.
func pickRadio(t *testing.T, mountID string, byPointer bool) js.Value {
	doc := js.Global().Get("document")
	mount := doc.Call("createElement", "div")
	mount.Set("id", mountID)
	doc.Get("body").Call("appendChild", mount)
	f, err := form.New(mountID, &focusRecord{}, &testIDGen{})
	if err != nil {
		t.Fatal(err)
	}
	if err := dom.Render(mountID, f); err != nil {
		t.Fatal(err)
	}
	doc.Get("activeElement").Call("blur") // drop focus left by an earlier test
	radios := mount.Call("querySelectorAll", "input[type=radio]")
	radio := radios.Index(1)
	if byPointer {
		// a real press lands on the option's label, not on the hidden native input
		radio.Get("parentElement").Call("dispatchEvent", js.Global().Get("Event").New("pointerdown"))
	}
	radio.Set("checked", true)
	radio.Call("dispatchEvent", js.Global().Get("Event").New("change"))
	return doc.Get("activeElement")
}

func TestRadioChosenWithPointerMovesToNextField(t *testing.T) {
	active := pickRadio(t, "ra-ptr", true)
	if active.Get("name").String() != "Name" {
		t.Fatalf("focus on %q, want the next field", active.Get("name").String())
	}
}

func TestRadioChosenWithKeyboardStaysInGroup(t *testing.T) {
	active := pickRadio(t, "ra-key", false)
	if active.Get("name").String() == "Name" {
		t.Fatal("arrow-key change must not leave the radio group")
	}
}
