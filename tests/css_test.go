//go:build !wasm

package form_test

import (
	"testing"

	"webtyp.com/fmt"
	"webtyp.com/form"
	"webtyp.com/widget"
)

func TestForm_RenderCSS(t *testing.T) {
	sheet := form.RenderCSS()
	if sheet == nil {
		t.Fatal("RenderCSS() returned nil stylesheet")
	}

	cssText := sheet.String()
	if cssText == "" {
		t.Fatal("RenderCSS().String() returned empty string")
	}

	expectedClasses := []string{
		widget.NameField.Root().String(),
		widget.NameField.Class(widget.PartForm).String(),
		widget.NameField.Class(widget.PartLabel).String(),
		widget.NameField.Class(widget.PartInput).String(),
		widget.NameField.Class(widget.PartError).String(),
		widget.NameField.Class(widget.PartSubmit).String(),
		widget.NameField.Class(widget.PartRadioGroup).String(),
		widget.NameField.Class(widget.PartRadioOption).String(),
		widget.NameField.Class(widget.PartRadioNative).String(),
		widget.NameField.Class(widget.PartRadioUnchecked).String(),
		widget.NameField.Class(widget.PartRadioChecked).String(),
		widget.NameField.Class(widget.PartCheckOption).String(),
		widget.NameField.Class(widget.PartCheckNative).String(),
		widget.NameField.Class(widget.PartCheckUnchecked).String(),
		widget.NameField.Class(widget.PartCheckChecked).String(),
		widget.NameField.Class(widget.PartReveal).String(),
	}

	for _, cls := range expectedClasses {
		if !fmt.Contains(cssText, cls) {
			t.Errorf("expected stylesheet to contain class %q", cls)
		}
	}

	// Verify that PartRadioGroup and PartCheckOption carry Panel frame styles
	for _, partClass := range []string{
		widget.NameField.Class(widget.PartRadioGroup).String(),
		widget.NameField.Class(widget.PartCheckOption).String(),
	} {
		if !fmt.Contains(cssText, "."+partClass+" {") && !fmt.Contains(cssText, "."+partClass+"{") {
			t.Errorf("missing rule for .%s", partClass)
		}
	}

	// Also verify method on Form receiver
	f := &form.Form{}
	if f.WidgetName() != widget.NameField {
		t.Errorf("WidgetName() = %q, want %q", f.WidgetName(), widget.NameField)
	}
	if f.WidgetKind() != widget.Form {
		t.Errorf("WidgetKind() = %v, want %v", f.WidgetKind(), widget.Form)
	}
	methodSheet := f.RenderCSS()
	if methodSheet == nil {
		t.Fatal("f.RenderCSS() returned nil")
	}
}
