# Interactivity and Reactive Binding

> Summary in `README.md` under "WASM Event Flow". This file has additional details.

## Signal-Bound Fields

`webtyp/form` uses a reactive binding model via `webtyp.com/dom` signals. Unlike traditional imperative mounting where a form must "attach" listeners after being added to the DOM, this form is reactive by construction.

Each field is bound to a `SignalString` for its value and another for its error state.

```go
// Simplified field binding in Render()
in := dom.NewElement("input").
    Bind(field.value).
    On("input", func(e dom.Event) {
        field.value.Set(e.TargetValue())
        if err := field.input.Validate(field.value.Get()); err != nil {
            field.err.Set(err.Error())
        } else {
            field.err.Set("")
        }
    })

errSpan := dom.NewElement("span").
    BindText(field.err)
```

## Advantages of Reactive Binding

1. **IME & Cursor Safety**: Two-way `Bind` in `webtyp/dom` is designed to be cursor-safe. It detects if the element is currently focused and active (e.g., during IME composition for accents or CJK characters) and avoids patching the `value` attribute if it matches the current signal value. This prevents cursor jumps and broken compositions.
2. **Surgical Updates**: When a user types, only the specific error text node and CSS classes are patched in the DOM. The `<input>` node itself is never replaced or re-rendered.
3. **No Lifecycle Hooks**: There is no `OnMount` or `OnUnmount`. The form is a `dom.Component` that provides a `Render() *dom.Element` method. The DOM tree it returns is already "alive" with signal bindings.

## Event Handlers

**Live Validation**:
Triggered on the `input` event. It updates the field's value signal and performs validation, updating the error signal immediately.

**Submission**:
Triggered on the form's `submit` event.
1. `e.PreventDefault()`
2. `f.SyncValues(f.data)` — pulls current values from signals into the struct.
3. `f.Validate()` — full validation check.
4. `f.submitting.Set(true)` — toggles the submit button's loading state.
5. `f.onSubmit(f.data, done)` — user callback.
6. `done(err)` — called by user to signal completion.
7. `f.submitting.Set(false)` and optional `f.reset()`.
