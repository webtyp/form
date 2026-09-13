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
	// Mirrors Form.Validate()'s own skip check (validate.go) — that one already
	// gates submission; this one gates the live per-keystroke error display.
	// Before this, a field with SetSkipValidation(true) still submitted fine
	// but showed a red error while typing, defeating the whole point of
	// skipping validation for a field that must not reveal its own format.
	if skipper, ok := fc.Input.(interface{ GetSkipValidation() bool }); ok && skipper.GetSkipValidation() {
		fc.err.Set("")
		return
	}
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

	var control *dom.Element
	var extra *dom.Element

	if r, ok := fc.Input.(Renderer); ok {
		control = r.RenderInput(fc.value, func(v string) {
			fc.value.Set(v)
			fc.validate(v)
		})
	} else {
		htmlName := fc.Input.HTMLName()
		switch htmlName {
		case "radio":
			control = fc.buildRadio()
		case "select":
			control = fc.buildSelect()
		case "datalist":
			control, extra = fc.buildDatalist()
		default:
			control, extra = fc.buildInput()
		}
	}

	// Field label. Rendered structurally for every titled field so a global form
	// skin (e.g. components/fieldset) can present it as a chip/legend; `for` ties
	// it to the input for click-to-focus. Form ships no styling for it — the look
	// is the consumer's skin.
	if lbl := fc.labelText(); lbl != "" && control != nil {
		container.Child(dom.NewElement("label").
			For(control).
			Class(widget.NameField.Class(widget.PartLabel).String()).
			Attr("title", lbl). // the untruncated text stays reachable
			Text(fmt.Convert(lbl).Truncate(labelChars).String()))
	}

	if control != nil {
		container.Child(control)
	}
	if extra != nil {
		container.Child(extra)
	}

	errSpan := dom.NewElement("span").
		ID(fc.Input.ErrorID()).
		Class(widget.NameField.Class(widget.PartError).String()).
		Attr("aria-live", "polite").
		BindText(fc.err)

	container.Child(errSpan)
	return container
}

func (fc *fieldComponent) buildInput() (*dom.Element, *dom.Element) {
	tag := "input"
	htmlName := fc.Input.HTMLName()
	if htmlName == "textarea" {
		tag = "textarea"
	}

	el := dom.NewElement(tag).
		ID(fc.Input.GetID()).
		Class(widget.NameField.Class(widget.PartInput).String()).
		Attr("name", fc.Input.FieldName())

	var reveal *dom.Element
	if tag == "input" {
		masker, isMasked := fc.Input.(interface{ IsMasked() bool })
		if isMasked && masker.IsMasked() {
			reveal = fc.wireMaskToggle(el)
		} else {
			el.Attr("type", htmlName)
		}
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
	el.OnInput(func(e dom.Event) {
		val := e.TargetValue()
		fc.value.Set(val)
		fc.validate(val)
	})
	if fc.onCommit != nil {
		el.OnBlur(func(dom.Event) { fc.onCommit() })
	}

	applyCommonAttrs(el, fc)
	return el, reveal
}

// wireMaskToggle binds el's type attribute to a revealed/hidden signal
// (instead of the fixed "password" a masked field would otherwise get) and
// returns the button that flips it — the show/hide toggle NIST SP 800-63B
// §5.1.1.2 recommends, on by default for every masked field (see
// docs/PLAN.md in veltylabs/mjosefa-cms for why this replaced an app-local
// button built with syscall/js and a class no stylesheet defined).
func (fc *fieldComponent) wireMaskToggle(el *dom.Element) *dom.Element {
	revealed := dom.NewBool(false)
	el.BindAttrFunc("type", func() string {
		if revealed.Get() {
			return "text"
		}
		return "password"
	})
	return dom.NewElement("button").
		Attr("type", "button"). // never submit — revealing must not send the form
		ID(fc.Input.GetID() + ".reveal").
		Class(widget.NameField.Class(widget.PartReveal).String()).
		BindState(widget.Selected, revealed).
		BindAttrFunc("aria-label", func() string {
			if revealed.Get() {
				return "Ocultar"
			}
			return "Mostrar"
		}).
		BindAttrFunc("aria-pressed", func() string {
			if revealed.Get() {
				return "true"
			}
			return "false"
		}).
		OnClick(func(dom.Event) { revealed.Set(!revealed.Get()) }).
		Child(eyeGlyph())
}

// eyeGlyph is a real inline SVG (fill="currentColor", so components/fieldset's
// Glyph() colors it through the cascade like any other text) — never an emoji:
// a "👁" renders a different bitmap per OS/font and ignores color entirely.
// One shape for both states; components/fieldset's When(Selected, …) is what
// changes how it reads as revealed vs hidden, not a second icon to keep in
// sync.
func eyeGlyph() *dom.Element {
	return dom.NewElement("svg").
		Attr("viewBox", "0 0 24 24").
		Attr("width", "18").
		Attr("height", "18").
		Attr("aria-hidden", "true").
		Attr("fill", "currentColor").
		Child(dom.NewElement("path").
			Attr("d", "M12 5C5 5 1 12 1 12s4 7 11 7 11-7 11-7-4-7-11-7zm0 12a5 5 0 1 1 0-10 5 5 0 0 1 0 10zm0-2a3 3 0 1 0 0-6 3 3 0 0 0 0 6z"))
}

func (fc *fieldComponent) buildSelect() *dom.Element {
	el := dom.NewElement("select").
		ID(fc.Input.GetID()).
		Class(widget.NameField.Class(widget.PartInput).String()).
		Attr("name", fc.Input.FieldName())

	if fc.Input.IsRequired() {
		el.Attr("required", "")
	}
	el.BindAttrBoolFunc("disabled", fc.isDisabledOrLocked)

	val := fc.value.Get()

	// Two-way binding for select
	el.Bind(fc.value)
	el.OnChange(func(e dom.Event) {
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
	return el
}

func (fc *fieldComponent) buildRadio() *dom.Element {
	group := dom.NewElement("div").
		ID(fc.Input.GetID()).
		Class(widget.NameField.Class(widget.PartRadioGroup).String())
	val := fc.value.Get()
	for _, opt := range fc.Input.GetOptions() {
		radio := dom.NewElement("input").
			Attr("type", "radio").
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

		radio.OnChange(func(e dom.Event) {
			if e.TargetChecked() {
				fc.value.Set(opt.Key)
				fc.validate(opt.Key)
				if fc.onCommit != nil {
					fc.onCommit()
				}
			}
		})

		label := dom.NewElement("label").For(radio)
		label.Child(radio)
		label.Child(dom.NewElement("span").Text(opt.Value))
		group.Child(label)
	}
	return group
}

func (fc *fieldComponent) buildDatalist() (*dom.Element, *dom.Element) {
	listID := fc.Input.GetID() + "-list"

	el := dom.NewElement("input").
		Attr("type", "text").
		ID(fc.Input.GetID()).
		Class(widget.NameField.Class(widget.PartInput).String()).
		Attr("name", fc.Input.FieldName()).
		Attr("list", listID)

	// Two-way binding
	el.Bind(fc.value)
	el.OnInput(func(e dom.Event) {
		val := e.TargetValue()
		fc.value.Set(val)
		fc.validate(val)
	})
	if fc.onCommit != nil {
		el.OnBlur(func(dom.Event) { fc.onCommit() })
	}

	applyCommonAttrs(el, fc)

	// dom exposes (*Element).For for for=, but nothing equivalent for list= / aria-describedby / aria-labelledby yet.
	datalist := dom.NewElement("datalist").ID(listID)
	for _, opt := range fc.Input.GetOptions() {
		datalist.Child(dom.NewElement("option").Attr("value", opt.Key).Text(opt.Value))
	}
	return el, datalist
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
