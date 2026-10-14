//go:build wasm

package form_test

import (
	"syscall/js"
	"testing"
	"webtyp.com/dom"
	"webtyp.com/form"
	"webtyp.com/input"
	"webtyp.com/model"
)

type blurStruct struct {
	model.Fielder
	Val string
}

func (s *blurStruct) Schema() []model.Field {
	inp := input.Text()
	if c, ok := inp.(interface{ SetRequired(bool) }); ok {
		c.SetRequired(true)
	}
	return []model.Field{{Name: "val", Type: inp}}
}
func (s *blurStruct) Pointers() []any { return []any{&s.Val} }
func (s *blurStruct) Values() []any   { return []any{s.Val} }

func TestValidateOnBlur(t *testing.T) {
	s := &blurStruct{}
	f, _ := form.New("app", s, &testIDGen{})

	doc := js.Global().Get("document")
	mountEl := doc.Call("createElement", "div")
	mountEl.Set("id", "app")
	doc.Get("body").Call("appendChild", mountEl)
	defer mountEl.Call("remove")

	if err := dom.Render("app", f); err != nil {
		t.Fatalf("dom.Render: %v", err)
	}

	inputEl := doc.Call("getElementById", "app.form.val")
	if inputEl.IsNull() {
		t.Fatalf("could not find input")
	}

	errSpan := doc.Call("getElementById", "app.form.val.error")
	if errSpan.IsNull() {
		t.Fatalf("could not find error span")
	}

	// 1. Typing an invalid value shows no error initially
	inputEl.Set("value", "")
	inputEl.Call("dispatchEvent", js.Global().Get("Event").New("input"))

	if errSpan.Get("textContent").String() != "" {
		t.Errorf("expected no error while typing initially, got %q", errSpan.Get("textContent").String())
	}

	// 2. Blur shows the error
	inputEl.Call("dispatchEvent", js.Global().Get("Event").New("blur"))
	if errSpan.Get("textContent").String() == "" {
		t.Errorf("expected error after blur, got none")
	}

	// 3. Typing a valid value clears it without a blur
	inputEl.Set("value", "abc")
	inputEl.Call("dispatchEvent", js.Global().Get("Event").New("input"))
	if errSpan.Get("textContent").String() != "" {
		t.Errorf("expected error to clear on valid input, got %q", errSpan.Get("textContent").String())
	}
}
