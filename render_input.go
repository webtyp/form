package form

import (
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/input"
	"webtyp.com/widget"
)

// labelChars is the label's budget, calibrated against the --chip-width the
// form skin gives the chip. Truncate counts the three-byte ellipsis inside this
// number, so the visible text is labelChars-3 at most.
//
// The count is BYTES, not runes: an accented character costs two, and a cut can
// land mid-character. Field names are identifiers in practice, so this is
// tolerable here; a title with accents long enough to trip it belongs in the
// input's title, not its label.
const labelChars = 14

// fieldComponent wraps an input.Input to implement dom.Component.
type fieldComponent struct {
	input.Input
	value *dom.SignalString
	err   *dom.SignalString
	// locked mirrors the owning Form's whole-form read-only gate (Form.SetLocked).
	// Shared across every field, so toggling it re-locks/unlocks the entire form.
	locked *dom.SignalBool
	// onCommit fires when the user finishes editing this field (blur for
	// text/textarea/datalist, change for select/radio) — the auto-save hook set
	// via Form.OnFieldChange. Nil when the form has none registered.
	onCommit func()
}

// isDisabledOrLocked combines the field's own static disabled flag with the
// form-wide locked signal — either one disables the rendered control.
func (fc *fieldComponent) isDisabledOrLocked() bool {
	return fc.Input.IsDisabled() || (fc.locked != nil && fc.locked.Get())
}

func (fc *fieldComponent) String() string {
	return fc.Render().String()
}

// GetID must differ from the input's own id. The framework injects a component's
// id onto its root element (here the field wrapper div); if that equalled the
// input's id, getElementById would resolve the wrapper instead of the input and
// the value binding would write to the div, leaving the input empty.
func (fc *fieldComponent) GetID() string {
	return fc.Input.GetID() + ".field"
}

func (fc *fieldComponent) SetID(id string) {
	fc.Input.SetID(id)
}

func (fc *fieldComponent) Children() []dom.Component {
	return nil
}

// Renderer is an optional capability for custom inputs that own their markup.
// The form still owns the field wrapper (the field wrapper div), the error span, and
// validation: the widget must call onInput with the new value on user input —
// the form updates the value signal and runs live validation. The value
// signal carries the initial value and programmatic updates (SetValues).
type Renderer interface {
	RenderInput(value *dom.SignalString, onInput func(string)) *dom.Element
}

func (fc *fieldComponent) validate(val string) {
	if err := fc.Input.Validate(val); err != nil {
		fc.err.Set(err.Error())
	} else {
		fc.err.Set("")
	}
}

// labelText picks the human label for the field's chip: the title first, then
// the placeholder, then the raw field name as a last resort.
func (fc *fieldComponent) labelText() string {
	if t := fc.Input.GetTitle(); t != "" {
		return t
	}
	if p := fc.Input.GetPlaceholder(); p != "" {
		return p
	}
	return fc.Input.FieldName()
}

func (fc *fieldComponent) Render() *dom.Element {
	container := dom.NewElement("div").
		Class(widget.NameField.Root().String()).
		BindStateFunc(widget.Invalid, func() bool { return fc.err.Get() != "" }).
		BindStateFunc(widget.Locked, fc.isDisabledOrLocked)

	// Field label. Rendered structurally for every titled field so a global form
	// skin (e.g. components/fieldset) can present it as a chip/legend; `for` ties
	// it to the input for click-to-focus. Form ships no styling for it — the look
	// is the consumer's skin.
	if lbl := fc.labelText(); lbl != "" {
		container.Child(dom.NewElement("label").
			Attr("for", fc.Input.GetID()).
			Class(widget.NameField.Class(widget.PartLabel).String()).
			Attr("title", lbl). // the untruncated text stays reachable
			Text(fmt.Convert(lbl).Truncate(labelChars).String()))
	}

	if r, ok := fc.Input.(Renderer); ok {
		container.Child(r.RenderInput(fc.value, func(v string) {
			fc.value.Set(v)
			fc.validate(v)
		}))
	} else {
		htmlName := fc.Input.HTMLName()
		switch htmlName {
		case "radio":
			fc.renderRadio(container)
		case "select":
			fc.renderSelect(container)
		case "datalist":
			fc.renderDatalist(container)
		default:
			fc.renderInput(container)
		}
	}

	errSpan := dom.NewElement("span").
		ID(fc.Input.ErrorID()).
		Class(widget.NameField.Class(widget.PartError).String()).
		Attr("aria-live", "polite").
		BindText(fc.err)

	container.Child(errSpan)
	return container
}

func (fc *fieldComponent) renderInput(container *dom.Element) {
	tag := "input"
	htmlName := fc.Input.HTMLName()
	if htmlName == "textarea" {
		tag = "textarea"
	}

	el := dom.NewElement(tag).
		ID(fc.Input.GetID()).
		Class(widget.NameField.Class(widget.PartInput).String()).
		Attr("name", fc.Input.FieldName())

	if tag == "input" {
		el.Attr("type", htmlName)
	}

	// Initial value for SSR
	val := fc.value.Get()
	if val != "" {
		if htmlName == "textarea" {
			el.Text(val)
		} else {
			el.Attr("value", val)
		}
	}

	// Two-way binding
	el.Bind(fc.value)
	el.On("input", func(e dom.Event) {
		val := e.TargetValue()
		fc.value.Set(val)
		fc.validate(val)
	})
	if fc.onCommit != nil {
		el.On("blur", func(dom.Event) { fc.onCommit() })
	}

	applyCommonAttrs(el, fc)
	container.Child(el)
}

func (fc *fieldComponent) renderSelect(container *dom.Element) {
	el := dom.NewElement("select").
		ID(fc.Input.HandlerName()).
		Class(widget.NameField.Class(widget.PartInput).String()).
		Attr("name", fc.Input.FieldName())

	if fc.Input.IsRequired() {
		el.Attr("required", "")
	}
	el.BindAttrBoolFunc("disabled", fc.isDisabledOrLocked)

	val := fc.value.Get()

	// Two-way binding for select
	el.Bind(fc.value)
	el.On("change", func(e dom.Event) {
		val := e.TargetValue()
		fc.value.Set(val)
		fc.validate(val)
		if fc.onCommit != nil {
			fc.onCommit()
		}
	})

	for _, opt := range fc.Input.GetOptions() {
		option := dom.NewElement("option").Attr("value", opt.Key).Text(opt.Value)
		if val != "" && opt.Key == val {
			option.Attr("selected", "")
		}
		el.Child(option)
	}
	container.Child(el)
}

func (fc *fieldComponent) renderRadio(container *dom.Element) {
	group := dom.NewElement("div").Class(widget.NameField.Class(widget.PartRadioGroup).String())
	val := fc.value.Get()
	for _, opt := range fc.Input.GetOptions() {
		optID := fc.Input.HandlerName() + "." + opt.Key
		label := dom.NewElement("label")

		radio := dom.NewElement("input").
			Attr("type", "radio").
			ID(optID).
			Attr("name", fc.Input.FieldName()).
			Attr("value", opt.Key)

		if val != "" && opt.Key == val {
			radio.Attr("checked", "")
		}

		// Reactive checked state
		radio.BindAttrBoolFunc("checked", func() bool {
			return fc.value.Get() == opt.Key
		})
		radio.BindAttrBoolFunc("disabled", fc.isDisabledOrLocked)

		radio.On("change", func(e dom.Event) {
			if e.TargetChecked() {
				fc.value.Set(opt.Key)
				fc.validate(opt.Key)
				if fc.onCommit != nil {
					fc.onCommit()
				}
			}
		})

		label.Child(radio)
		label.Child(dom.NewElement("span").Text(opt.Value))
		group.Child(label)
	}
	container.Child(group)
}

func (fc *fieldComponent) renderDatalist(container *dom.Element) {
	listID := fc.Input.GetID() + "-list"

	el := dom.NewElement("input").
		Attr("type", "text").
		ID(fc.Input.GetID()).
		Class(widget.NameField.Class(widget.PartInput).String()).
		Attr("name", fc.Input.FieldName()).
		Attr("list", listID)

	// Two-way binding
	el.Bind(fc.value)
	el.On("input", func(e dom.Event) {
		val := e.TargetValue()
		fc.value.Set(val)
		fc.validate(val)
	})
	if fc.onCommit != nil {
		el.On("blur", func(dom.Event) { fc.onCommit() })
	}

	applyCommonAttrs(el, fc)
	container.Child(el)

	datalist := dom.NewElement("datalist").ID(listID)
	for _, opt := range fc.Input.GetOptions() {
		datalist.Child(dom.NewElement("option").Attr("value", opt.Key).Text(opt.Value))
	}
	container.Child(datalist)
}

func applyCommonAttrs(el *dom.Element, fc *fieldComponent) {
	inp := fc.Input
	if ph := inp.GetPlaceholder(); ph != "" {
		el.Attr("placeholder", ph)
	}
	if title := inp.GetTitle(); title != "" {
		el.Attr("title", title)
	}
	for _, attr := range inp.GetAttributes() {
		if attr.Value != "" {
			el.Attr(attr.Key, attr.Value)
		}
	}
	if inp.IsRequired() {
		el.Attr("required", "")
	}
	el.BindAttrBoolFunc("disabled", fc.isDisabledOrLocked)
	if inp.IsReadonly() {
		el.Attr("readonly", "")
	}
}

// RenderInput is kept for backward compatibility and as a standalone helper
func RenderInput(inp input.Input) *dom.Element {
	fc := &fieldComponent{
		Input: inp,
		value: dom.NewString(""),
		err:   dom.NewString(""),
	}
	// Note: this Render() returns the field wrapper div containing the input + error span.
	return fc.Render()
}
