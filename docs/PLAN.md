---
PLAN: "feat: form shows translated labels, help and options (Field.Label, Field.Help + lang)"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 11869723703476135862
PR: https://github.com/webtyp/form/pull/24
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan — form: every visible text goes through `lang`

Phase **T2** of the master plan `SOURCE_SELECTION_MASTER_PLAN.md` (orchestration only — everything
this plan needs is inline). **Depends on two published tags:**
- `webtyp.com/model` with `Field.Label` and `Field.Help` (`https://github.com/webtyp/model/blob/main/docs/PLAN.md`);
- `webtyp.com/lang` with the page-dictionary stage (Part 2)
  (`https://github.com/webtyp/lang/blob/main/docs/PLAN.md`).

Read [AGENTS.md](../AGENTS.md) first (signals-only component contract, typed harness). Rules that
matter here:
- This package compiles to WASM: use `webtyp.com/fmt`, never `strings`/`strconv`/stdlib `fmt`.
- Tests live in `tests/` (public API only). A root-level test needs a top-of-file
  `// Root-level test (justified): …` comment. **Never export a symbol so a test can reach it.**

## Why

In mjosefa-cms the device form reads `Name`, `Ip`, `Type`, `Location`, `Is_active` (labels), the
checkbox says `is_active`, the radio options say `Computer / Printer / Server / Other`, and the IP
placeholder says `example: 192.168.1.1`. All of it is English, and the labels are raw column names,
because `form` never calls `lang`:
- `labelText()` (`render_input.go` ~115) returns the input's title, else its **placeholder**, else
  the field name;
- option texts (`render_input.go` ~319, ~541) are written as is. The placeholder is already right:
  `input` keeps its parts (`SetPlaceholder("example:", "192.168.1.1")`) and `GetPlaceholder()`
  translates each part with `lang.Translate`. It only needs a dictionary, which this wave adds;
- the checkbox text (~466) is the placeholder or the label;
- the submit button defaults to `"Submit"` (`form.go` ~421).

From now on, translations are data. The page carries a dictionary, and `lang.Translate` looks
texts up in it. Each argument of `Translate` is one key, exactly as written: the author chooses
word or phrase with the comma. A library writes **English** and passes every visible text through
`lang.Translate`.

## Design gate

1. **Prior art.** Django forms render `verbose_name` through `gettext`; Angular Material labels go
   through the i18n pipe; React Hook Form plus i18next calls `t(label)`. The library translates at
   render time, and the schema carries the source-language text.
2. **Novice-name test.** No new public names. The behaviour is "a label comes from `Field.Label`".
3. **Complexity ledger.**
   ```
   Concepts the developer must learn   +0 (Label is model's)
   Files they must touch to do X       −1 (no per-app label overrides)
   Lines at the call site              +0
   Ways to do the same thing           −1 (the placeholder no longer doubles as a label)
   ```
4. **Where it belongs.** `form` renders the text, so `form` translates it at render.
5. **What it deletes.** The placeholder-as-label fallback in `labelText()`.

## Stage 1 — the label

1. `fieldComponent` gets an unexported `label string`, computed once in `New` (`form.go`, in the loop
   over `schema`) from the `model.Field`:
   - `inp.GetTitle()` if not empty (explicit per-input override, unchanged);
   - else `field.Label`;
   - else the humanised `field.Name`: every `_` becomes a space (`is_active` → `is active`). Write it
     as one unexported helper `humanize(name string) string`, and use it in both places that need
     it.
   `RenderInput` (the standalone helper, no `model.Field`) uses `GetTitle()`, else
   `humanize(inp.FieldName())`.
2. `labelText()` returns `lang.Translate(fc.label).String()`. **Delete** the placeholder fallback.
   The CSS `Capitalize` on the label part stays, so `is active` shows as `Is active` until it is
   translated.
3. Checkbox text (~466): the placeholder if set (already translated by `input`), else
   `lang.Translate(fc.label).String()`.

## Stage 1b — help text

Already published, use them: `webtyp.com/widget` ≥ v0.6.36 has `widget.PartHelp`, and
`webtyp.com/dom` ≥ v0.13.20 has `(*Element).DescribedBy(others ...*Element)`. It sets
`aria-describedby` from the minted IDs, the same contract as `For`. Never compose an id by hand.
1. `fieldComponent` gets an unexported `help string` = `field.Help`, set in `New` like `label`.
2. In `fieldComponent.Render()`, when `help != ""`, add right after the label element:
   `helpEl := dom.NewElement("small").Class(widget.NameField.Class(widget.PartHelp).String()).Text(lang.Translate(fc.help).String())`.
3. The control gets `DescribedBy(helpEl, errSpan)` (help first, then the error span already
   built for `fc.Input.ErrorID()`). With no help, use `DescribedBy(errSpan)`, so the error
   message is announced by screen readers too.
4. Skin (`css.go`): add `Part(widget.PartHelp, …)` — small, muted text in normal flow under the
   control: `style.Glyph(style.Inactive)`, `style.FontSize(style.TextXs)`,
   `style.PadInline(style.Space4)`. Compose only existing recipes. If no recipe expresses "muted
   small text", that is a defect in `webtyp/widget/style`: stop, and report it in the PR under
   `## Executor notes`. Never hand-write CSS here.

## Stage 2 — options, submit

- Placeholder (~550): **no change.** `GetPlaceholder()` already returns the translated text, built
  part by part by `input`. Translating it again here would look up the joined text as one key.
- Option texts (~319 select, ~541 datalist, and the radio builder): `lang.Translate(opt.Value).String()`.
  The option's `value` attribute (`opt.Key`) is never translated.
- `resolveSubmitLabel()`: the default `"Submit"` becomes `lang.Translate("Submit").String()`, and an
  explicit `SubmitLabel(text)` is translated too.
- `go get webtyp.com/lang@latest webtyp.com/model@latest`.

## Stage 3 — tests (`tests/`)

- **Backend** (`tests/label_test.go`): there is no dictionary on the backend, so `Translate` passes
  English through. Render a form from a `model.Definition` whose fields are:
  - `{Name: "ip", Label: "IP address", Type: input.IP()}`: the label text is `IP address`;
  - `{Name: "is_active", Type: input.Checkbox()}` with no Label: the label and the checkbox text are
    `is active`;
  - a field whose input has a placeholder and no Label: the label is the humanised name, **not** the
    placeholder. This is the red test for the deleted fallback; it fails on today's code;
  - `{Name: "rut", Help: "Format: 12.345.678-9", Type: input.Rut()}`: a `small` element with class
    `field__help` and that text, and the control's `aria-describedby` lists the help element's ID
    first and then the error span's ID; a field without Help has `aria-describedby` = the error
    span's ID only.
- **WASM** (`tests/label_translate_test.go`, `//go:build wasm`): `TestMain` inserts
  `<script type="application/json" id="` + lang.ScriptID + `">` with
  `{"default":"es","languages":["es"],"keys":{"Computer":["Computador"],"IP address":["Dirección IP"],"Submit":["Guardar"],"example:":["ejemplo:"],"is active":["Activo"]}}`
  (positional lists in `languages` order; every key is one whole label, option or placeholder part,
  exactly as written)
  before any lookup. Then, under `lang.OutLang(lang.ES)`:
  - the label is `Dirección IP`, and the field with no `Label` (`is_active`) shows `Activo`;
  - the radio option text is `Computador` while its value stays the key;
  - the placeholder is `ejemplo: 192.168.1.1`;
  - the submit button reads `Guardar`.
- Root-level tests (`state_attrs_internal_test.go` and any other `package form` file at the root):
  if they only use exported identifiers, `git mv` them to `tests/`. Otherwise add the
  `// Root-level test (justified): …` comment.

## Acceptance

- `gotest` passes (includes WASM).
- `grep -n "GetPlaceholder" render_input.go`: no hit inside `labelText`.
- `grep -rn "Text(opt.Value)" --include='*.go' .` → empty.

## Stages

| # | Stage | Files |
|---|---|---|
| 1 | Label | `form.go`, `render_input.go` |
| 1b | Help | `form.go`, `render_input.go`, `css.go`, `go.mod` (widget ≥ v0.6.36, dom ≥ v0.13.20) |
| 2 | Options, submit | `render_input.go`, `form.go`, `go.mod`, `go.sum` |
| 3 | Tests | `tests/label_test.go`, `tests/label_translate_test.go`, root `*_test.go` |
