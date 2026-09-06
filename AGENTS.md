# Agent Guide — `webtyp/form`

Constraints for agents working on the form library. Read this before any change.

---

## Construction Harness — typed & explicit (the WebTyp approach)

This library is part of WebTyp's **construction harness**: the typed, explicit API is what keeps an
agent that doesn't know the library from building wrong code. The compiler must reject mistakes; what
it can't catch becomes a `devMode` warning — never a silent failure.

- **Typed over `any`** — no generic slots; typed builder methods (like `webtyp/json`), reusing `fmt` types. Anything reactive goes only through a signal binding (`BindText`/`Bind*`), which requires a signal.
- **Explicit names** — `Text` (static) vs `BindText` (reactive); reading the call states intent.
- **Illegal states unrepresentable** — dynamic content has ONE path, typed to require a signal.
- **Minimal public surface** — export only what the author types; engine plumbing stays unexported.
- **Docs are minimal "how" instructions, not long skills** — if a rule must be *remembered*, close it
  with types, not prose.

(Ecosystem rationale: `webtyp/app/docs/CONSTRUCTION_HARNESS.md`.)

---

## Component Contract — ONE way (signals)

The form implements **only** `Render() *dom.Element` (+ optional `Init(ctx dom.Ctx)`). There is
**NO** `OnMount`/`OnUpdate`/`OnUnmount` and **NO** manual `Update()` (unexported in `dom`).

Each field's value and error are **typed signals**; live validation patches only the bound error
node — the `<input>` is never re-rendered, so focus/cursor and IME composition are preserved.

```go
in := dom.Input("email").Bind(field.value).         // two-way: input <-> SignalString (cursor/IME safe)
    On("input", func(dom.Event){ field.err.Set(validate(field.value.Get())) })
errSpan := dom.Span().BindText(field.err)            // only this text node patches on error change
```

Do **not** re-introduce delegated `dom.Get(id).SetText(...)` listener wiring in a lifecycle hook —
that imperative model is removed.

## No Generics

Zero generic functions (follow `webtyp/fmt` codec rule "cero any, cero map"). Use concrete typed
signals: `SignalString`/`SignalBool`/`SignalNodes`, `DeriveString`/`DeriveBool`, and the `Bind*`
methods. Never `Signal[T]`.

## No Go `map` — no exceptions

**Never use a built-in `map[K]V`, in any file.** TinyGo's map runtime is heavy and this package
compiles into every WASM app that renders a form — a map here is a size tax every one of them
pays. Use `fmt.KeyValue{Key, Value string}` for a string pair (see `showFields`), or a small
slice-of-structs scanned linearly for anything else — every collection in this package (fields on
one form, one form's inputs) is small enough that a linear scan costs nothing measurable. Not a
"just this once" exception.

## Minimal Public Surface

Export only the form's user-facing API (`Bind`-based construction, config, `OnSubmit`). Unexport
helpers, the per-field model, and anything only this package uses. Struct fields stay unexported.

## WASM / TinyGo

- `//go:build wasm` for reactive code (`mount.go`); keep the `!wasm` stub (`mount_stub.go`) method
  set in sync (it shrinks — no `OnMount`/`OnUnmount`).
- No Go stdlib: use `webtyp.com/fmt`. DOM only via `webtyp.com/dom`, never
  `syscall/js`. No `defer/recover`.

## Testing

```bash
go install webtyp.com/devflow/cmd/gotest@latest
gotest
```

- `gotest`, never `go test`. Stdlib assertions only. Dual WASM/stdlib; real DOM in WASM tests.
- Cover frequent use cases: live validation patches a single node (input keeps identity + cursor),
  submit/loading/reset lifecycle, and an **IME check** (type `á`, `ñ`, a CJK char — composition
  must not break). Publish with `gopush 'message'`.

## Documentation First

Update docs **before** code and before `gopush`. Notably rewrite `docs/INTERACTIVITY_AND_MOUNTING.md`
(it describes the OBSOLETE delegated-mount model), record the decision in `docs/DESIGN.md`, update
`docs/API.md`, and re-index `README.md` so every `docs/` file is linked. Diagrams: `flowchart TD`,
no `subgraph`, `<br/>` for breaks.
