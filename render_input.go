package form

import (
	"webtyp.com/components/tapinfotip"
	"webtyp.com/country"
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/input"
	"webtyp.com/lang"
	"webtyp.com/widget"
)

// fieldComponent wraps an input.Input to implement dom.Component.
type fieldComponent struct {
	input.Input
	value *dom.SignalString
	err   *dom.SignalString
	label string
	// tip is the field's help (model's Field.Help): its button sits beside the
	// label and its panel under the control. Nil when the field has no Help.
	tip *tapinfotip.TapInfotip
	// locked mirrors the owning Form's whole-form read-only gate (Form.SetLocked).
	// Shared across every field, so toggling it re-locks/unlocks the entire form.
	locked *dom.SignalBool
	// hidden reports whether Field.ShownWhen leaves this field out right now.
	// Nil when the field is always shown.
	hidden func() bool
	// onValueSet repaints a control that keeps state derived from the value (the
	// phone). Nil for every other control; set when the control is built.
	onValueSet func(string)
	// focusEl is the element a label points at and aria-describedby names when the
	// control is a group (the phone row): its inner input. Nil otherwise.
	focusEl *dom.Element
	// onCommit fires when the user finishes editing this field (blur for
	// text/textarea/datalist, change for select/radio) — the auto-save hook set
	// via Form.OnFieldChange. Nil when the form has none registered.
	onCommit func()
	// options holds the live choice nodes of a select, radio group or datalist,
	// so Form.SetOptions after render repaints them: the choices usually come
	// from a Caller whose answer arrives once the form is already on screen.
	// Nil until a choice control is built.
	options      *dom.SignalNodes
	buildOptions func() []*dom.Element
}

// bindOptions binds el's choices to a signal that already holds the current
// options. The signal owns the children (dom): the first render, on the server
// and in the browser, draws exactly what it holds, and refreshOptions replaces
// them later.
func (fc *fieldComponent) bindOptions(el *dom.Element, build func() []*dom.Element) {
	fc.buildOptions = build
	fc.options = dom.NewNodes(build()...)
	el.BindChildren(fc.options)
}

// refreshOptions repaints the choice nodes from Input.GetOptions(). A no-op
// before the control is built: the first render reads the options directly.
func (fc *fieldComponent) refreshOptions() {
	if fc.options != nil && fc.buildOptions != nil {
		fc.options.Set(fc.buildOptions())
	}
}

// isDisabledOrLocked combines the field's own static disabled flag with the
// form-wide locked signal — either one disables the rendered control.
func (fc *fieldComponent) isDisabledOrLocked() bool {
	return fc.IsDisabled() || (fc.locked != nil && fc.locked.Get())
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
// validation: the widget must call onInput with the new value when the person
// finishes editing it (a commit, like blur: never per keystroke) — the form
// updates the value signal, validates it and fires the commit (auto-save). The value
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
	if val == "" {
		if req, ok := fc.Input.(interface{ IsRequired() bool }); ok && !req.IsRequired() {
			fc.err.Set("")
			return
		}
	}
	if err := fc.Validate(val); err != nil {
		fc.err.Set(err.Error())
	} else {
		fc.err.Set("")
	}
}

// labelText picks the human label for the field's chip: the title first, then
// the placeholder, then the raw field name as a last resort.
func (fc *fieldComponent) labelText() string {
	return lang.Translate(fc.label).String()
}

func (fc *fieldComponent) Render() *dom.Element {
	container := dom.NewElement("div").
		Class(widget.NameField.Root().String()).
		BindStateFunc(widget.Invalid, func() bool { return fc.err.Get() != "" }).
		BindStateFunc(widget.Locked, fc.isDisabledOrLocked).
		BindStateFunc(widget.Hidden, func() bool { return fc.hidden != nil && fc.hidden() })

	if spanner, ok := fc.Input.(interface{ IsFullWidth() bool }); ok && spanner.IsFullWidth() {
		container.SetState(widget.SpanFull)
	}

	var control *dom.Element
	var extra *dom.Element

	if r, ok := fc.Input.(Renderer); ok {
		control = r.RenderInput(fc.value, func(v string) {
			fc.value.Set(v)
			fc.validate(v)
			if fc.onCommit != nil {
				fc.onCommit()
			}
		})
	} else {
		htmlName := fc.HTMLName()
		switch htmlName {
		case "radio":
			control = fc.buildRadio()
		case "checkbox":
			control = fc.buildCheckbox()
		case "select":
			control = fc.buildSelect()
		case "tel":
			control, extra = fc.buildPhone()
		case "datalist":
			control, extra = fc.buildDatalist()
		default:
			control, extra = fc.buildInput()
		}
	}

	// Field label, above the control. `for` ties it to the control for click-to-
	// focus and the accessible name — except where the control is not labelable:
	// a radio group is a div, so it names itself through role + aria-labelledby;
	// a lone checkbox carries its own text inside the option and has no label
	// above (it would repeat it), and takes the whole row so it never leaves a
	// gap beside a neighbour that does have a label.
	htmlName := ""
	if _, custom := fc.Input.(Renderer); !custom {
		htmlName = fc.HTMLName()
	}
	var label *dom.Element
	if htmlName == "checkbox" {
		container.SetState(widget.SpanFull)
	} else if lbl := fc.labelText(); lbl != "" && control != nil {
		label = dom.NewElement("label").
			Class(widget.NameField.Class(widget.PartLabel).String()).
			Text(lbl)
		if htmlName == "radio" {
			control.Attr("role", "radiogroup").Attr("aria-labelledby", label.GetID())
		} else {
			label.For(fc.labelTarget(control))
		}
	}

	// The help button shares the label's line. A field with Help always has
	// the line, even without a label, so the button never floats alone.
	// A lone checkbox has no label above, so its button shares the line of the
	// option itself (see where the control is added below).
	loneCheckbox := htmlName == "checkbox"
	if fc.tip != nil && control != nil && !loneCheckbox {
		row := dom.NewElement("div").Class(widget.NameField.Class(widget.PartLabelRow).String())
		if label != nil {
			row.Child(label)
		}
		row.Child(fc.tip.Button())
		container.Child(row)
	} else if label != nil {
		container.Child(label)
	}

	errSpan := dom.NewElement("span").
		ID(fc.ErrorID()).
		Class(widget.NameField.Class(widget.PartError).String()).
		Attr("aria-live", "polite").
		BindText(fc.err)

	// The help text opens under the control, only when its button is tapped.
	var panel *dom.Element
	if fc.tip != nil {
		panel = fc.tip.Panel()
	}

	if control != nil {
		if panel != nil {
			fc.labelTarget(control).DescribedBy(panel, errSpan)
		} else {
			fc.labelTarget(control).DescribedBy(errSpan)
		}
		if loneCheckbox && fc.tip != nil {
			container.Child(dom.NewElement("div").
				Class(widget.NameField.Class(widget.PartLabelRow).String()).
				Child(control, fc.tip.Button()))
		} else {
			container.Child(control)
		}
	}
	if extra != nil {
		container.Child(extra)
	}
	if panel != nil {
		container.Child(panel)
	}

	container.Child(errSpan)
	return container
}

func (fc *fieldComponent) buildInput() (*dom.Element, *dom.Element) {
	tag := "input"
	htmlName := fc.HTMLName()
	if htmlName == "textarea" {
		tag = "textarea"
	}

	el := dom.NewElement(tag).
		ID(fc.Input.GetID()).
		Class(widget.NameField.Class(widget.PartInput).String()).
		Attr("name", fc.FieldName())

	if htmlName == "textarea" {
		el.Class(widget.NameField.Class(widget.PartTextarea).String())
		el.Attr("rows", "3")
	}

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
		if fc.err.Get() != "" {
			fc.validate(val)
		}
	})
	el.OnBlur(func(dom.Event) {
		fc.validate(fc.value.Get())
		if fc.onCommit != nil {
			fc.onCommit()
		}
	})

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
		ID(fc.Input.GetID()+".reveal").
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
		Class(widget.NameField.Class(widget.PartInput).String(), widget.NameField.Class(widget.PartSelect).String()).
		Attr("name", fc.FieldName())

	if fc.IsRequired() {
		el.Attr("required", "")
	}
	el.BindAttrBoolFunc("disabled", fc.isDisabledOrLocked)

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

	fc.bindOptions(el, func() []*dom.Element {
		val := fc.value.Get()
		opts := fc.GetOptions()
		nodes := make([]*dom.Element, 0, len(opts)+1)
		// With no value chosen the select shows a prompt instead of silently
		// presenting its first option as if it had been picked. A required field
		// cannot go back to the prompt once it has a value (disabled), an optional
		// one can (that is how a value is cleared). Skipped when the options
		// already include an empty one.
		if !hasEmptyOption(opts) {
			prompt := dom.NewElement("option").Attr("value", "").Text(lang.Translate("Select an option").String())
			if val == "" {
				prompt.Attr("selected", "")
			}
			if fc.IsRequired() {
				prompt.Attr("disabled", "")
			}
			nodes = append(nodes, prompt)
		}
		for _, opt := range opts {
			option := dom.NewElement("option").Attr("value", opt.Key).Text(lang.Translate(opt.Value).String())
			if val != "" && opt.Key == val {
				option.Attr("selected", "")
			}
			nodes = append(nodes, option)
		}
		return nodes
	})
	return el
}

func hasEmptyOption(opts []fmt.KeyValue) bool {
	for _, o := range opts {
		if o.Key == "" {
			return true
		}
	}
	return false
}

func (fc *fieldComponent) buildRadio() *dom.Element {
	group := dom.NewElement("div").
		ID(fc.Input.GetID()).
		Class(widget.NameField.Class(widget.PartRadioGroup).String())
	fc.bindOptions(group, func() []*dom.Element {
		val := fc.value.Get()
		opts := fc.GetOptions()
		nodes := make([]*dom.Element, 0, len(opts))
		for _, opt := range opts {
			optKey := opt.Key
			optVal := opt.Value

			radio := dom.NewElement("input").
				Attr("type", "radio").
				Class(widget.NameField.Class(widget.PartRadioNative).String()).
				Attr("name", fc.FieldName()).
				Attr("value", optKey)

			if val != "" && optKey == val {
				radio.Attr("checked", "")
			}

			// Reactive checked state
			radio.BindAttrBoolFunc("checked", func() bool {
				return fc.value.Get() == optKey
			})
			radio.BindAttrBoolFunc("disabled", fc.isDisabledOrLocked)

			radio.OnChange(func(e dom.Event) {
				if e.TargetChecked() {
					fc.value.Set(optKey)
					fc.validate(optKey)
					if fc.onCommit != nil {
						fc.onCommit()
					}
				}
			})

			label := dom.NewElement("label").
				For(radio).
				Class(widget.NameField.Class(widget.PartRadioOption).String()).
				BindStateFunc(widget.Selected, func() bool {
					return fc.value.Get() == optKey
				})

			label.Child(radio)
			label.Child(dom.NewElement("span").Text(lang.Translate(optVal).String()))
			nodes = append(nodes, label)
		}
		return nodes
	})
	return group
}

func (fc *fieldComponent) buildCheckbox() *dom.Element {
	val := fc.value.Get()
	isChecked := val == "true" || val == "1" || val == "on"

	chk := dom.NewElement("input").
		Attr("type", "checkbox").
		ID(fc.Input.GetID()).
		Class(widget.NameField.Class(widget.PartCheckNative).String()).
		Attr("name", fc.FieldName())

	if isChecked {
		chk.Attr("checked", "")
	}

	chk.BindAttrBoolFunc("checked", func() bool {
		v := fc.value.Get()
		return v == "true" || v == "1" || v == "on"
	})
	chk.BindAttrBoolFunc("disabled", fc.isDisabledOrLocked)

	chk.OnChange(func(e dom.Event) {
		checked := e.TargetChecked()
		v := "false"
		if checked {
			v = "true"
		}
		fc.value.Set(v)
		fc.validate(v)
		if fc.onCommit != nil {
			fc.onCommit()
		}
	})

	label := dom.NewElement("label").
		For(chk).
		Class(widget.NameField.Class(widget.PartCheckOption).String()).
		BindStateFunc(widget.Selected, func() bool {
			v := fc.value.Get()
			return v == "true" || v == "1" || v == "on"
		})

	label.Child(chk)
	label.Child(checkUncheckedGlyph())
	label.Child(checkCheckedGlyph())
	txt := fc.GetPlaceholder()
	if txt == "" {
		txt = lang.Translate(fc.label).String()
	}
	label.Child(dom.NewElement("span").Text(txt))
	return label
}

// checkUncheckedGlyph and checkCheckedGlyph are only the boxes' elements: their
// colours, outline and radius come from the stylesheet (css.go), so no colour
// is written here. The viewBox crops to the tick's own 18x18 square so the
// tick fills the CSS box.
func checkUncheckedGlyph() *dom.Element {
	return dom.NewElement("svg").
		Class(widget.NameField.Class(widget.PartCheckUnchecked).String()).
		Attr("aria-hidden", "true")
}

func checkCheckedGlyph() *dom.Element {
	return dom.NewElement("svg").
		Class(widget.NameField.Class(widget.PartCheckChecked).String()).
		Attr("viewBox", "3 3 18 18").
		Attr("aria-hidden", "true").
		Child(dom.NewElement("path").
			Attr("d", "M7 12.5l3.5 3.5 7-7").
			Attr("stroke", "currentColor").
			Attr("stroke-width", "2.5").
			Attr("stroke-linecap", "round").
			Attr("stroke-linejoin", "round").
			Attr("fill", "none"))
}

func (fc *fieldComponent) buildDatalist() (*dom.Element, *dom.Element) {
	listID := fc.Input.GetID() + "-list"

	el := dom.NewElement("input").
		Attr("type", "text").
		ID(fc.Input.GetID()).
		Class(widget.NameField.Class(widget.PartInput).String()).
		Attr("name", fc.FieldName()).
		Attr("list", listID)

	// Two-way binding
	el.Bind(fc.value)
	el.OnInput(func(e dom.Event) {
		val := e.TargetValue()
		fc.value.Set(val)
		if fc.err.Get() != "" {
			fc.validate(val)
		}
	})
	el.OnBlur(func(dom.Event) {
		fc.validate(fc.value.Get())
		if fc.onCommit != nil {
			fc.onCommit()
		}
	})

	applyCommonAttrs(el, fc)

	// dom exposes (*Element).For for for=, but nothing equivalent for list= / aria-describedby / aria-labelledby yet.
	datalist := dom.NewElement("datalist").ID(listID)
	fc.bindOptions(datalist, func() []*dom.Element {
		opts := fc.GetOptions()
		nodes := make([]*dom.Element, 0, len(opts))
		for _, opt := range opts {
			nodes = append(nodes, dom.NewElement("option").Attr("value", opt.Key).Text(lang.Translate(opt.Value).String()))
		}
		return nodes
	})
	return el, datalist
}

func applyCommonAttrs(el *dom.Element, fc *fieldComponent) {
	inp := fc.Input
	if ph := inp.GetPlaceholder(); ph != "" {
		el.Attr("placeholder", ph)
	}
	// A title equal to the field name is the generated default, not a tooltip
	// someone wrote: it would show the technical name on hover and become the
	// control's accessible name. The visible label already names the control.
	if title := inp.GetTitle(); title != "" && title != inp.FieldName() {
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
	lbl := inp.GetTitle()
	if lbl == "" || lbl == inp.FieldName() {
		lbl = fmt.Convert(inp.FieldName()).Replace("_", " ").String()
	}
	fc := &fieldComponent{
		Input: inp,
		value: dom.NewString(""),
		err:   dom.NewString(""),
		label: lbl,
	}
	// Note: this Render() returns the field wrapper div containing the input + error span.
	return fc.Render()
}

// labelTarget is the element a label points at: the control, or the inner input
// of a group control (the phone row).
func (fc *fieldComponent) labelTarget(control *dom.Element) *dom.Element {
	if fc.focusEl != nil {
		return fc.focusEl
	}
	return control
}

// buildPhone draws the phone field: a select of dial prefixes and the national
// number. The choices are the app's (view.WithOptions on this field, built with
// country.Catalog.PrefixOptions) and the first one is the default country. The
// value the field holds is always the international form, "+56911111111"; the
// person never types the prefix. With no choices the field is one box where the
// number is typed with its "+".
func (fc *fieldComponent) buildPhone() (*dom.Element, *dom.Element) {
	opts := fc.GetOptions()
	if len(opts) == 0 {
		return fc.buildInput()
	}
	defaultCode := opts[0].Key
	code, national := splitPhone(fc.value.Get(), defaultCode)
	prefix := dom.NewString(code)
	number := dom.NewString(national)
	fc.onValueSet = func(v string) {
		c, n := splitPhone(v, defaultCode)
		prefix.Set(c)
		number.Set(n)
	}
	compose := func() {
		next := ""
		if digits := onlyDigits(number.Get()); digits != "" {
			if c, ok := country.Find(prefix.Get()); ok {
				next = "+" + c.Dial + digits
			}
		}
		fc.value.Set(next)
		fc.validate(next)
		if fc.onCommit != nil {
			fc.onCommit()
		}
	}

	sel := dom.NewElement("select").
		Class(widget.NameField.Class(widget.PartInput).String(), widget.NameField.Class(widget.PartSelect).String(), widget.NameField.Class(widget.PartPrefix).String()).
		Attr("aria-label", lang.Translate("Country prefix").String()).
		BindAttrBoolFunc("disabled", fc.isDisabledOrLocked).
		Bind(prefix).
		OnChange(func(e dom.Event) {
			prefix.Set(e.TargetValue())
			compose()
		})
	for _, opt := range opts {
		sel.Child(dom.NewElement("option").Attr("value", opt.Key).Text(opt.Value))
	}

	num := dom.NewElement("input").
		ID(fc.Input.GetID()).
		Attr("type", "tel").
		Attr("inputmode", "tel").
		Attr("autocomplete", "tel-national").
		Attr("name", fc.FieldName()).
		Class(widget.NameField.Class(widget.PartInput).String(), widget.NameField.Class(widget.PartNumber).String()).
		BindAttrBoolFunc("disabled", fc.isDisabledOrLocked).
		Bind(number).
		OnChange(func(e dom.Event) {
			number.Set(e.TargetValue())
			compose()
		})
	fc.focusEl = num

	return dom.NewElement("div").Class(widget.NameField.Class(widget.PartPhone).String()).Child(sel, num), nil
}

// splitPhone reads a stored value into the prefix's country code and the
// national digits. An empty value, or one that is not international (an old
// record typed without "+"), shows the default country with the digits it has.
func splitPhone(value, defaultCode string) (code, national string) {
	if c, rest, ok := country.Split(value); ok {
		return c.Code, rest
	}
	return defaultCode, onlyDigits(value)
}

// onlyDigits keeps the digits of what the person typed.
func onlyDigits(typed string) string {
	out := make([]byte, 0, len(typed))
	for i := 0; i < len(typed); i++ {
		if typed[i] >= '0' && typed[i] <= '9' {
			out = append(out, typed[i])
		}
	}
	return string(out)
}
