package form

import (
	"webtyp.com/dom"
	"webtyp.com/widget"
)

// SetSSR enables or disables SSR mode for this form.
func (f *Form) SetSSR(enabled bool) *Form {
	f.ssrMode = enabled
	return f
}

// String serializes the form to its HTML string representation.
func (f *Form) String() string {
	return f.Render().String()
}

// Render returns a reactive dom.Element tree for the form.
func (f *Form) Render() *dom.Element {
	// The vy-field__form anatomy hook is the fieldset skin styles: it owns the
	// inter-field rhythm (a gap on this container, so the ends do not double
	// the way each field's own margin would). Any app classes from
	// SetClass/SetGlobalClass are layered on top.
	el := dom.NewElement("form").
		ID(f.GetID()).
		Class(widget.NameField.Class(widget.PartForm).String())

	if f.class != "" {
		el.Class(f.class)
	}

	// SSR mode: render method and action
	if f.ssrMode {
		el.Attr("method", f.method).Attr("action", f.action)
	}

	for _, child := range f.Children() {
		el.Child(child)
	}

	// Submit button
	if !f.noSubmit {
		btn := dom.NewElement("button").
			Attr("type", "submit").
			Class(widget.NameField.Class(widget.PartSubmit).String()).
			ID(f.id + ".submit")

		btn.BindAttrBool("disabled", f.submitting)

		btn.BindTextFunc(func() string {
			if f.submitting.Get() {
				label := f.submitLoadingLabel
				if label == "" {
					label = f.resolveSubmitLabel() + "..."
				}
				return label
			}
			return f.resolveSubmitLabel()
		})

		// Wrapped in the same field box every field gets, not appended
		// bare: a bare button skips the field's own inline inset and comes out
		// wider than the inputs above it on both edges — the one misaligned
		// edge in an otherwise squared-off stack. Sharing the wrapper aligns it
		// by construction, and the form's gap spaces it from the last field
		// like any other row.
		el.Child(dom.NewElement("div").
			Class(widget.NameField.Root().String()).
			Child(btn))
	}

	// Bind submit event
	el.On("submit", func(e dom.Event) {
		e.PreventDefault()
		f.Submit()
	})

	return el
}
