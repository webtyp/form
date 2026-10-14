---
PLAN: "feat(form)!: one field format — label above, single column, help and error below, validate on blur"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 4748136233166290427
---

> Read AGENTS.md at the repo root before any change.

# Plan — the one field format of the app

## Problem

`form` is the only field of every webtyp app, but its skin is a coloured label chip straddling the
input's top border, a two-column grid, an error docked in the input's top-right corner, and a label
truncated to 14 bytes ("Fecha de nacimiento" → "Fecha de na…"). Apps that found it hard to use wrote
their own fields by hand, so the app shows three field styles. The decided format (master plan
`UI_CONSISTENCY`, decisions D2, D3, D4, D10, P11) is:

```
 Full name *                    ← <label>, plain text, translated, never truncated
┌──────────────────────────────┐
│ Dra. Ana Rojas               │ ← control, full width, touch height (ControlBox)
└──────────────────────────────┘
 Format: 12.345.678-9            ← help (only when the model has Help)
 ⚠ Invalid RUT                   ← error; while present the help is hidden
```

Same format on phone and desktop; one column always; the field being edited shows its label in the
primary colour.

## Design gate

1. **Prior art.** GOV.UK Design System (label above, hint below label, error message, validate on
   submit/blur), Material 3 / Polaris / Atlassian (label above as default), Baymard Institute research
   (top-aligned labels fastest to complete; single column — multi-column forms cause skipped fields;
   inline validation after the field is left, not on every keystroke).
2. **Novice-name.** No exported name added or renamed. The format changes; the API does not.
3. **Ledger.** Concepts −1 (no chip/legend geometry). Ways to render a field in the ecosystem: the
   hand-made fields in apps are removed in their own plans; this plan makes `form` the one they move to.
4. **Where it belongs.** `form` owns the field: label, control, help, error.
5. **What it deletes.** `labelChars` and the `Truncate` call; the chip skin of `PartLabel`
   (`OnEdge`, `As(Primary)`, `Raise`, `Capitalize`, `PadInline` chip padding); `ChipSeat` on the field
   root; `Grid(2, ColumnMedium, …)` on `PartForm`; `Docked(...)` on `PartError`; the `PadEdge(EdgeTop,
   Space2/Space4)` that made room for the chip on `PartInput`, `PartTextarea`, `PartRadioGroup`;
   per-keystroke validation of a field that has no error yet.

## Stage 1 — markup order and no truncation (`render_input.go`)

1. Delete the `labelChars` constant and its comment block.
2. In `(*fieldComponent).Render`, the label text is `lbl` as is: replace
   `Text(fmt.Convert(lbl).Truncate(labelChars).String())` with `Text(lbl)`; keep `Attr("title", lbl)`
   only if another test relies on it, otherwise delete it too (the full text is now visible).
3. Child order becomes: **label → control → extra (reveal button / datalist) → help → error**
   (today help is appended before the control). `DescribedBy(helpEl, errSpan)` stays as is.
4. Remove the `webtyp.com/fmt` import if it becomes unused.

## Stage 2 — validate on blur, then live while an error is shown (`render_input.go`)

For every control that today calls `fc.validate(val)` inside `OnInput` (text/textarea in `buildInput`,
the datalist input in `buildDatalist`, and the `Renderer` callback in `Render`):

- On input: `fc.value.Set(val)`; then `if fc.err.Get() != "" { fc.validate(val) }` — an error already
  shown clears as soon as the user fixes it, but a field without an error is not judged mid-typing.
- On blur: `fc.validate(fc.value.Get())`, then the existing `fc.onCommit()` call if non-nil (one
  `OnBlur` handler doing both, in that order).
- Select, radio and checkbox keep validating in their `OnChange` (a choice is complete when made).
- `Form.Validate()` / `Submit()` are unchanged: they still paint every failing field.

## Stage 3 — a field without `Label` is a loud diagnostic (`form.go`)

In `New`, where the label falls back to `humanize(field.Name)`:

```go
if label == "" {
	label = humanize(field.Name)
	dom.Log(errFieldWithoutLabel, structName, field.Name)
}
```

with a package constant
`const errFieldWithoutLabel = "webtyp/form: field has a widget but no Label — set model.Field.Label:"`
and `structName` the value `resolveStructName(data)` already computes in `New` (reuse it, do not
recompute). The humanized fallback stays so the screen still renders.

## Stage 4 — the skin (`css.go`)

Rewrite `RenderCSS()` with these parts (keep every rule not listed here — reveal, radio/check glyphs,
locked, submit, SpanFull — exactly as it is, except removing the chip padding listed in the gate):

| Part | Options |
|---|---|
| Root (field) | `Stack(Space1)`, `Anchor()`, `KeepSize()` — no `ChipSeat` |
| `PartForm` | `Stack(Space4)`, `Width(Readable)` — one column; no `Grid`, no `PadInline` |
| `PartLabel` | `FontSize(TextSm)`, `FontWeight(WeightMedium)`, `StartContent()` — no chip options |
| `PartInput` | `As(Field)`, `Round(RadiusMd)`, `PadInline(Space4)`, `ControlBox()` — no `PadEdge` |
| `PartTextarea` | `PadEdge(EdgeTop, Space2)`, `PadInline(Space4)` |
| `PartRadioGroup` | `Row(Space2)` (wraps, packed to the start) — no `As(Field)`, no `ControlBox`, no `PadEdge` |
| `PartRadioOption` | `Button(style.Panel)`, `Round(RadiusMd)` — each option is a chip, as wide as its label |
| `PartHelp` | `Glyph(Inactive)`, `FontSize(TextXs)` — no `PadInline` |
| `PartError` | `Glyph(Danger)`, `FontSize(TextXs)` — in the flow, no `Docked` |

States:

- Radio as chips (same look as `webtyp/components/chipgrid`): `When(widget.Selected, widget.PartRadioOption, style.As(style.Primary))`; `PartRadioUnchecked` and `PartRadioChecked` get `style.Hide()` and the two `WhenWithin(widget.Selected, …)` glyph rules are deleted; the native radio stays `VisuallyHidden()` (keyboard and screen readers keep working). Locked: `WhenWithin(widget.Locked, "", widget.PartRadioOption, style.As(style.Inactive))` replaces the locked rule on `PartRadioGroup`.

- `WhenWithin(widget.Invalid, "", widget.PartHelp, style.Hide())` — the error replaces the help.
- Keep `When(widget.Invalid, widget.PartInput, style.Glyph(style.Danger))`.
- Focus: `CueWithin(widget.FocusWithin, "", widget.PartLabel, style.Glyph(style.Primary))` — the label
  of the field being edited takes the primary colour. The input's frame keeps the focus treatment
  `As(style.Field)` already emits. If `CueWithin` with an empty container does not target the field
  root, use the form the existing `WhenWithin(..., "", ...)` calls use; if no existing option can
  express it, **stop and report it in the PR** — never hand-compose CSS.
- Delete the comments that describe the chip/legend geometry; one short comment per part is enough.

## Stage 5 — tests: behaviour only, never the look

- **Delete** the existing tests whose only assertion is the look: emitted CSS declarations, chip/legend
  classes, label truncation, child order (`styling_test.go`, `css_test.go` and the look-only cases in
  `label.*_test.go`, `render_input_test.go`, `anatomy*_test.go`). Keep every test that asserts
  behaviour (values, validation, submit, dirty state, locked, options).
- Add only:
  - `validate_on_blur.front_test.go` (wasm, like `onfieldchange.front_test.go`): typing an invalid
    value shows no error; blur shows it; typing a valid value then clears it without a blur.
  - `label_missing_test.go`: `form.New` over a model whose widget field has no `Label` calls the logger
    set with `dom.SetLog` once with `errFieldWithoutLabel` and the field name.

## Stage 6 — docs

`README.md`, `docs/DESIGN.md`, `docs/API.md`: replace any description of the chip/legend, the two
columns or the 14-byte truncation with the format of this plan (the ASCII sketch above), and state that
a field without `Label` logs `webtyp/form: field has a widget but no Label`.

## Acceptance

- `grep -rn "labelChars\|Truncate(\|ChipSeat\|OnEdge\|Grid(2" --include=*.go .` → empty (outside `tests/`).
- `gotest ./...` green (stdlib + wasm).

| Stage | Files |
|---|---|
| 1, 2 | `render_input.go` |
| 3 | `form.go` |
| 4 | `css.go` |
| 5 | the look-only tests deleted, `tests/validate_on_blur.front_test.go`, `tests/label_missing_test.go` |
| 6 | `README.md`, `docs/DESIGN.md`, `docs/API.md` |
