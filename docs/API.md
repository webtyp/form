# API Reference

See `README.md` for the consolidated API. This file contains additional detail.

## `form.New` — Widget Resolution Detail

```go
f, err := form.New("content", data, ids) // data implements model.Fielder; ids is a model.IDGenerator (e.g. unixid.NewUnixID())
// -> f.GetID() == "content." + resolveStructName(data)
// -> f.String() renders all fields that have a Widget in data.Schema()
```

For each field in `data.Schema()`:
1. `field.IsPK() && field.IsAutoInc()` → skip (auto-increment PKs not editable).
2. `field.Widget == nil` → skip (no UI binding).
3. `field.Widget.Clone(formID, fieldName).(input.Input)` → positioned input.
4. `field.NotNull` → `SetRequired(true)` on the input.
5. Current value bound via `fmt.ReadValues()` + `SetValues()`.

## `(*Form).Validate()` — Validation Detail

- Skips fields with `SkipValidation` set to true in the input.
- Pulls values from reactive signals.
- Calls `inp.Validate(val)` (promoted from `model.Kind`).
- Returns the **first** error encountered.
- Paints every failing field that holds a value into its own error signal
  (an empty failing field stays unpainted — its live validation paints it
  when the user types); passing fields are cleared.

## `(*Form).LoadValues(data model.Fielder)` — Load Detail

Populates every input from data, the inverse of `SyncValues`. A nil record
resets the form (the "new record" case). Loading also remembers each hidden
PK's id as form state, and clears stale validation errors.

## `(*Form).SyncValues(data model.Fielder)` — Binding Detail

Synchronizes input values back to the struct pointers provided by `data.Pointers()`.
Supports `model.FieldText`, `model.FieldInt`, `model.FieldFloat`, and `model.FieldBool`.
The hidden PK is written from the id `LoadValues` remembered (`Reset` cleared
it); the target's own PK is never read. An empty remembered text PK mints a
new id (kept across retries until `Reset`); an empty int PK is zeroed for the
DB to auto-increment.

## `(*Form).ValidateData(action byte, data model.Fielder)` — Server-side Validation

Validates the provided `data` using the form's input rules. Satisfies `crudp.DataValidator`.

## `(*Form).IsDirty()` and `(*Form).DirtyFields()` — Dirty State & Bulk Edit

- `f.IsDirty() bool`: Reports whether any field's current value differs from the baseline captured at creation or the last `LoadValues`/`Reset`/`MarkPristine`.
- `f.DirtyFields() []string`: Returns the field names (in schema order) whose current value differs from baseline. Returns `nil` (not an empty slice) when nothing is dirty, so `len(f.DirtyFields()) > 0` and `f.IsDirty()` always agree.
- **Bulk Edit Note**: A host applying a form to multiple records must write ONLY the fields returned by `DirtyFields()`. Writing all fields would silently revert untouched columns on other records to whatever the form was holding.

## `(*Form).Submit()`

Runs the full submit pipeline programmatically:
1. `SyncValues(f.data)`: copies values from signals to struct.
2. `Validate()`: final validation check.
3. If valid and `OnSubmit` is set:
   - Sets `submitting` signal to true.
   - Calls the `OnSubmit` callback.
   - When the callback's `done` function is called:
     - Sets `submitting` signal back to false.
     - Resets the form (unless `NoResetOnSuccess` was called).

Returns the first validation error, or nil if the submission was dispatched.
The DOM `submit` event handler delegates to this method.

## `form.Renderer`

Optional capability interface for custom inputs that own their markup.

```go
type Renderer interface {
    RenderInput(value *dom.SignalString, onInput func(string)) *dom.Element
}
```

- **Markup ownership**: The form still owns the field wrapper (`div.tw-field`), the error span, and the field ID. The widget provides the inner control.
- **Contract**: The widget must call `onInput` with the new value on user input. The form updates the value signal and runs live validation.
- **Location**: Lives in package `form` because it references `*dom.Element` (the `input` package is dom-free).

## `fmt.Permitted` — Validation Engine

```go
type Permitted struct {
	Letters    bool     // a-z, A-Z (and ñ/Ñ)
	Tilde      bool     // á, é, í, ó, ú
	Numbers    bool     // 0-9
	Spaces     bool     // ' '
	BreakLine  bool     // '\n'
	Tab        bool     // '\t'
	Extra      []rune   // additional allowed characters
	NotAllowed []string // disallowed substrings
	Minimum    int      // minimum length
	Maximum    int      // maximum length
}
```

Error messages from `Permitted.Validate(name, text)`:
- `"{name} minimum {min} chars"` — value shorter than Minimum
- `"{name} maximum {max} chars"` — value longer than Maximum
- `"space not allowed"` — space when Spaces=false
- `"character {X} not allowed"` — disallowed character

## Namer Interface

Fielder types can optionally implement `Namer` to provide a custom form name:

```go
type Namer interface {
    FormName() string
}
```

If not implemented, defaults to `"form"`.
