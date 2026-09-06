# webtyp/form
<img src="docs/img/badges.svg">

HTML forms generated from your model schema — validation included, no
reflection, TinyGo/WASM-ready. You don't declare forms: you declare a
[model.Definition](https://github.com/webtyp/model) once, and fields whose
kind is an `input.*` type become form inputs automatically.

This package is part of the [webtyp ecosystem](https://github.com/webtyp)
— Go libraries for building full-stack web apps compiled to WebAssembly.

## Install

```bash
go get webtyp.com/form   # brings webtyp/model as dependency
```

The code generator tooling is covered in step 2 of the Quick Start.

## Quick Start

### 1. Define your model

Two imports: [`model`](https://github.com/webtyp/model) provides the schema
types and base kinds (`model.Text()`, `model.Int()`, …); `input` is this
package's sub-package with the form kinds (`input.Text()`, `input.Email()`, …).
The kind you choose per field decides everything: `input.*` = form input +
validation; `model.*` = validation only (never rendered).

```go
import (
    "webtyp.com/model"
    "webtyp.com/input"
)

var UserModel = model.Definition{
    Name: "user",
    Fields: model.Fields{
        {Name: "id",    Type: model.Int(), DB: &model.FieldDB{PK: true, AutoInc: true}},
        {Name: "name",  Type: input.Text(),  NotNull: true},
        {Name: "email", Type: input.Email(), NotNull: true},
        {Name: "sku",   Type: SKU()},        // custom input — defined in your own package
        {Name: "notes", Type: model.Text()}, // validation only — never rendered
    },
}
```

Auto-increment PKs are skipped automatically (not editable).

A **custom input** like `SKU()` is a type in your own package that embeds
`input.Base` and configures its validation rules — it then works as a kind
like any built-in. Minimal shape (full pattern:
[input/README.md](input/README.md)):

```go
type sku struct{ input.Base }

func SKU() input.Input {
    s := &sku{}
    s.Letters, s.Numbers = true, true
    s.Maximum = 12
    s.InitBase("", "", "text")
    return s
}

func (s *sku) Clone(parentID, name string) input.Input {
    c := *s
    c.InitBase(parentID, name, "text")
    return &c
}
```

### 2. Generate the Fielder

From your `model.Definition` vars, the generator emits `<file>_orm.go` next
to each model file: the row struct plus the `model.Fielder` methods
(`Schema()`, `Pointers()`, `Values()`). You never write these by hand.

**Recommended**: use the [`webtyp`](https://github.com/webtyp/app) dev
environment — it watches your model files and regenerates `*_orm.go`
automatically with hot reload:

```bash
go install webtyp.com/app/cmd/webtyp@latest
webtyp -tui    # interactive dev server (or -mcp for AI agents)
```

**Manual alternative**: run [`ormc`](https://github.com/webtyp/ormc)
(the standalone generator that `webtyp` uses internally) once in your
module:

```bash
go install webtyp.com/ormc/cmd/ormc@latest
ormc
```

### 3. Create and render the form

Note you pass the generated **struct instance**, not the Definition: the form
needs a data holder — it reads initial values from it and writes submitted
values back into it. The schema travels along anyway: the generated
`Schema()` method returns `UserModel.Fields`.

```go
import "webtyp.com/form"

f, err := form.New("parent-id", &User{Name: "John"}, ids)  // ids: model.IDGenerator (e.g. unixid.NewUnixID())
html := f.String()          // SSR: render to HTML string
```

### 4. Make it interactive (WASM)

[`dom`](https://github.com/webtyp/dom) mounts components in the browser
(compiled with TinyGo):

```go
import "webtyp.com/dom"

f.LoadValues(record) // registro → formulario (el usuario selecciona)
f.Validate()         // el usuario edita y guarda
f.SyncValues(record) // formulario → registro (listo para enviar)

f.OnSubmit(func(data model.Fielder, done func(error)) {
    // send data to your API, then:
    done(nil) // nil = success → form resets (see NoResetOnSuccess)
})
dom.Mount("root", f)
```

Mounted forms get live per-field validation on input, a submit button bound
to the submitting state, and IME-safe reactive updates. Detail:
[Interactivity & Mounting](docs/INTERACTIVITY_AND_MOUNTING.md).

Runtime tweaks: `f.Input("Field").SetPlaceholder(...)`,
`f.SetOptions("Field", ...)`, `f.SetValues("Field", ...)`.

### Dirty State & Bulk Edit

- `IsDirty()` reports whether *any* field's current value differs from the baseline captured at creation or last load/reset (`LoadValues`, `Reset`, `MarkPristine`). Use this to gate persistence (e.g. auto-save).
- `DirtyFields()` returns the names of fields (in schema order) whose value differs from baseline, or `nil` when pristine. Use this for bulk edit to write *only* the fields the user actually changed, preventing untouched columns from being overwritten.

## Styling

The library ships structure, the project owns the look (CSS-first doctrine).
The forms do not embed CSS or rely on custom styling libraries, but instead emit a standard semantic anatomy defined by `webtyp/widget`:

1. **Stable anatomy and class contract** — every bound field renders with the classes:
   - Root container: `tw-field`
   - Label: `tw-field__label`
   - Inputs/Textareas/Selects/Datalists: `tw-field__input`
   - Error spans: `tw-field__error`
   - Radio group wrappers: `tw-field__radio-group`

2. **State attributes** — validation and lock states are published as standard reactive `data-*` attributes on the root container of each field:
   - Invalid state: `data-invalid="true"`
   - Locked / Read-only state: `data-locked="true"`

   Global form skins (e.g., `components/fieldset`) use these selectors to style fields dynamically, e.g., `.tw-field[data-invalid="true"] .tw-field__error { ... }`.

3. **`form.SetGlobalClass("my-app-form")`** and **`f.SetClass("local-class")`**
   — adds classes to the `<form>`, useful for scoping:
   `.my-app-form .tw-field { ... }`.

## Custom Inputs

Custom markup for custom inputs is possible by implementing `form.Renderer`;
see [input/README.md](input/README.md).

## Built-in Input Types

18 types in `input/` ([full reference](docs/STANDARD_TYPES.md)):

| Input | HTML type | | Input | HTML type |
|-------|-----------|-|-------|-----------|
| `Address` | `text` | | `Password` | `password` |
| `Checkbox` | `checkbox` | | `Phone` | `tel` |
| `Datalist` | `text` | | `Radio` | `radio` |
| `Date` | `date` | | `Rut` | `text` |
| `Email` | `email` | | `Search` | `search` |
| `Filepath` | `text` | | `Select` | `select` |
| `Gender` | `radio` | | `Text` | `text` |
| `Hour` | `time` | | `Textarea` | `textarea` |
| `IP` | `text` | | `Number` | `number` |

Need one that isn't here? **Custom inputs** live in your own package: embed
`input.Base`, configure the `Permitted` rules, override `Validate` if needed.
Full pattern: [input/README.md](input/README.md).

## API Reference

### `form.New(parentID string, data model.Fielder, idGen model.IDGenerator) (*Form, error)`

Creates a `Form` from any `Fielder`. `idGen` is required — form never
constructs its own ID generator (see `model.IDGenerator`); pass
`unixid.NewUnixID()` at your composition root, or a test double in tests.
Form `id` = `parentID + "." + name`, where the name comes from the optional
`Namer` interface (`FormName() string`, default `"form"`).

### Form Methods

| Method | Description |
|--------|-------------|
| `String() string` | Generates form HTML |
| `Render() *dom.Element` | **WASM** — reactive DOM tree (`dom.ViewRenderer`) |
| `SetSSR(bool) *Form` | SSR mode: adds `method`/`action` attributes |
| `OnSubmit(func(model.Fielder, func(error))) *Form` | WASM submit callback |
| `Validate() error` | Validates all inputs, returns first error |
| `LoadValues(model.Fielder) error` | Populates every input from data, the inverse of SyncValues |
| `SyncValues(model.Fielder) error` | Copies input values back into the data struct |
| `ValidateData(byte, model.Fielder) error` | Server-side validation (crudp.DataValidator) |
| `Input(fieldName string) input.Input` | Returns the input for a field name |
| `SetOptions(fieldName, ...fmt.KeyValue) *Form` | Options for select/radio/datalist |
| `SetValues(fieldName, ...string) *Form` | Sets a value programmatically |
| `IsDirty() bool` | Reports whether any field's value differs from baseline |
| `DirtyFields() []string` | Returns names of fields differing from baseline (nil if none) |
| `Submit() error` | Runs sync + validate + OnSubmit callback programmatically; returns first validation error |
| `Reset()` | Clears all values and error messages |
| `NoResetOnSuccess() *Form` | Keeps values after a successful submit |
| `SubmitLabel(string) *Form` | Submit button text (default "Submit") |
| `SubmitLoadingLabel(string) *Form` | Button text while submitting (default label + "...") |
| `HideSubmit() *Form` | Renders without a submit button |
| `SetClass(...string) *Form` | Appends CSS classes to this form (on top of SetGlobalClass) |
| `GetID() string` | Form's HTML id |

Package-level: `form.SetGlobalClass(classes ...string)` — CSS classes for all
forms created afterwards.

## How It Works

`form.New()` iterates `data.Schema()`: a field becomes a form input **iff its
`Type` (a `model.Kind`) also implements `input.Input`** — capability by
interface, no registry, no name matching. Base kinds (`model.Text()`, …) are
skipped for rendering but still validate (fail-closed). Bound inputs are
positioned clones (`Clone(formID, fieldName)`) with constraint defaults
applied (`NotNull` → required) and current values bound from `Pointers()`.

The `input` package stays free of `dom` imports (edge-safe); HTML rendering
is owned by `form`. Validation baselines come from `model.Permitted`
character whitelists — see [API Reference](docs/API.md).

## Documentation

- [API Reference](docs/API.md) — Validate, SyncValues, Permitted detail
- [Design & Architecture](docs/DESIGN.md) — core layers and philosophy
- [Standard Types](docs/STANDARD_TYPES.md) — the 18 input types in detail
- [Interactivity & Mounting](docs/INTERACTIVITY_AND_MOUNTING.md) — WASM event handling
- [Implementation Notes](docs/IMPLEMENTATION.md) — **contributors**: library rules, file map, adding built-in inputs, test layout
- [input/README.md](input/README.md) — input package: custom inputs, composition, `Base`

---

## [Contributing](https://github.com/webtyp/contributing)

---

## [License](LICENSE)
