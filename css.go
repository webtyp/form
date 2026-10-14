//go:build !wasm

package form

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// RenderCSS returns the standard form fieldset stylesheet.
func (f *Form) RenderCSS() *css.Stylesheet {
	return RenderCSS()
}

// RenderCSS returns the standard form fieldset stylesheet.
// It can be called directly or via Form.RenderCSS().
func RenderCSS() *css.Stylesheet {
	f := &Form{}
	return style.For(f).
		// The field owns NO vertical rhythm of its own — the gap between
		// consecutive fields is one value, on the form container (PartForm
		// below).
		Root(
			style.Anchor(),
			style.Stack(style.Space1),
			style.KeepSize(),
		).
		// The form: the one place the inter-field rhythm lives. A gap on the
		// container spaces every field the same and — unlike a per-field
		// margin — does not double up at the first and last field against the
		// card's own inset.
		Part(widget.PartForm,
			style.Stack(style.Space4),
			style.Width(style.Readable),
		).
		// Field label above the control.
		Part(widget.PartLabel,
			style.FontSize(style.TextSm),
			style.FontWeight(style.WeightMedium),
			style.StartContent(),
		).
		// The error message below the control.
		Part(widget.PartError,
			style.Glyph(style.Danger),
			style.FontSize(style.TextXs),
		).
		Part(widget.PartHelp,
			style.Glyph(style.Inactive),
			style.FontSize(style.TextXs),
		).
		// Field.
		Part(widget.PartInput,
			style.As(style.Field),
			style.Round(style.RadiusMd),
			style.PadInline(style.Space4),
			style.ControlBox(),
		).
		Part(widget.PartTextarea,
			style.PadEdge(style.EdgeTop, style.Space2),
			style.PadInline(style.Space4),
		).
		// Bottom-right corner of the field, not top-right: PartError already
		// owns the top-right corner (the two would overlap the moment a
		// masked field also has an error to show). Docked has no
		// vertical-center primitive today — a real gap in the DSL, not
		// solved here — so bottom-right is the closest fit that never
		// collides with the label chip (top) or the error message (top,
		// trailing). ControlBox matches PartInput/PartSubmit's own tap
		// target instead of a bespoke size. Inactive by default (a quiet
		// affordance, not a second call to action next to the field it
		// sits in) — When(Selected, …) below brightens it once revealed.
		Part(widget.PartReveal,
			style.Docked(style.Parent, style.EdgeBottom, style.SideEnd, style.Space2),
			style.Glyph(style.Inactive),
			style.ControlBox(),
		).
		Part(widget.PartRadioGroup,
			style.Row(style.Space2),
		).
		Part(widget.PartRadioOption,
			style.Button(style.Panel),
			style.Round(style.RadiusMd),
		).
		Part(widget.PartRadioNative,
			style.VisuallyHidden(),
		).
		Part(widget.PartRadioUnchecked,
			style.Hide(),
		).
		Part(widget.PartRadioChecked,
			style.Hide(),
		).
		When(widget.Selected, widget.PartRadioOption,
			style.As(style.Primary),
		).
		Part(widget.PartCheckOption,
			style.As(style.Field),
			style.Round(style.RadiusMd),
			style.ControlBox(),
			style.Row(style.Space2),
			style.PadInline(style.Space4),
			style.Interactive(style.Bare),
		).
		Part(widget.PartCheckNative,
			style.VisuallyHidden(),
		).
		Part(widget.PartCheckUnchecked,
			style.IconBox(style.IconMd),
		).
		Part(widget.PartCheckChecked,
			style.IconBox(style.IconMd),
			style.Hide(),
		).
		WhenWithin(widget.Selected, widget.PartCheckOption, widget.PartCheckUnchecked,
			style.Hide(),
		).
		WhenWithin(widget.Selected, widget.PartCheckOption, widget.PartCheckChecked,
			style.Show(),
		).
		// Filled and full-width, matching crudview's own primary action
		// button exactly (As(Primary), Pad(Space3), ControlBox) -- a form's
		// submit is the same weight of commitment as crudview's own save
		// action, and a bare unstyled <button> next to a skinned field stack
		// was the one piece of every form (login included) that still read
		// as unfinished.
		Part(widget.PartSubmit,
			style.Button(style.Primary),
			style.Width(style.Full),
			style.CenterContent(),
		).
		// Locked repaints the INPUT, not the field root — historically the
		// wrapper got the treatment and the resulting ring crossed the legend
		// chip straddling the top border line. Nothing about a locked field
		// needs a second box around the pair; the control itself is what is
		// locked, so the control is what changes. (Secondary drops the base
		// Panel's hairline frame and repaints the fill — no border, no ring;
		// the only box-shadow in this sheet stays the label chip's Raise
		// elevation.)
		//
		// This gate is PER FIELD, not per form. dom writes data-locked on the
		// wrapper whenever form's isDisabledOrLocked() says so, and that check
		// ORs the whole-form lock (SetLocked — no longer called by crudview,
		// but still public API) with the field's OWN
		// fc.Input.IsDisabled(), a per-field property declared on the model.
		// Without this rule, a genuinely disabled field would be
		// indistinguishable from an editable one — which is why the rule
		// survives the form-lock's retirement.
		//
		// Secondary is Panel's fill with no border: the control stops reading
		// as a writing area and settles back into the card, which is what a
		// value you are only reading should do. The text stays ColorOnSurface
		// at full contrast — a read-only value is still there to BE read, so
		// this deliberately does not reach for Inactive's muted text.
		// WhenWithin, not When: dom writes data-locked on the field wrapper
		// that owns the gate, so When(Locked, PartInput) would emit
		// `.NAME__input[data-locked="true"]` (NAME = widget.NameField) and match
		// nothing.
		// Round(RadiusMd) is not redundant: a surface with no explicit radius
		// falls back to its own default, and Secondary's is RadiusSm, so
		// locking silently shrank the control's corners from the 8px the base
		// rule gives it to 4px. Restating the base radius keeps the box the
		// same shape in both states — only its fill is supposed to change.
		WhenWithin(widget.Locked, "", widget.PartInput,
			style.As(style.Secondary),
			style.Round(style.RadiusMd),
		).
		WhenWithin(widget.Locked, "", widget.PartRadioOption,
			style.As(style.Inactive),
		).
		WhenWithin(widget.Locked, "", widget.PartCheckOption,
			style.As(style.Secondary),
			style.Round(style.RadiusMd),
		).
		// Hide help text when error is present.
		WhenWithin(widget.Invalid, "", widget.PartHelp,
			style.Hide(),
		).
		// No filled state on the field itself: the red message in the corner is
		// the signal. Painting the whole box danger buried the value.
		When(widget.Invalid, widget.PartInput,
			style.Glyph(style.Danger),
		).
		// Highlight label text color when field is focused.
		CueWithin(widget.FocusWithin, "", widget.PartLabel,
			style.Glyph(style.Primary),
		).
		// dom writes data-selected on the reveal button itself when the
		// value is showing (see form's wireMaskToggle) — When targets the
		// same element the state is written on, unlike WhenWithin above.
		When(widget.Selected, widget.PartReveal,
			style.Glyph(style.Primary),
		).
		When(widget.SpanFull, "",
			style.SpanFull(),
		).
		Stylesheet()
}
