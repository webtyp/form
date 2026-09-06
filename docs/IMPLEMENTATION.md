# Implementation Notes

> Contributor-facing. See [README.md](../README.md) for usage,
> [DESIGN.md](DESIGN.md) for architecture.

## Library Rules (Non-Negotiable)

- **No stdlib imports** in library code. Never `errors`, `strconv`, `strings`, `reflect`.
- **Use `webtyp.com/fmt`**: `fmt.Err("Noun", "Adjective")` for errors,
  `fmt.Convert(s).Int()` instead of `strconv.Atoi`, `fmt.Contains(s, sub)`
  instead of `strings.Contains`.
- **No maps in inputs**: use slices — maps increase WASM binary size.
- **No `any`/`map` in public APIs.** No `reflect` at runtime.
- **The `input` package must stay free of `dom` imports** (edge-safe):
  HTML rendering lives in the `form` package.

## File Map

| File | Responsibility |
|------|---------------|
| `form.go` | `Form` struct, `New()`, `Input()`, `SetOptions()`, `SetValues()`, `Reset()`, `Namer` |
| `sync.go` | `SyncValues()`, pointer-based field sync |
| `forms.go` | `SetGlobalClass()`, global forms state |
| `render.go` | `Render()`, `String()`, `SetSSR()`, submit event wiring |
| `render_input.go` | Field rendering (input + error span; owns `dom` imports); `RenderInput()` helper |
| `css.go` | `RenderCSS()` — base `tw-*` styles (`!wasm`, additive `css.Stylesheet`) |
| `validate.go` | `Validate()` |
| `validate_struct.go` | `ValidateData()` (crudp.DataValidator) |
| `input/interface.go` | `Input` interface (embeds `model.Kind` + metadata getters; no `dom.Component`) |
| `input/base.go` | `Base` struct embedded by all inputs |
| `input/*.go` | 18 concrete input implementations |
| `tests/` | Black-box tests (`package form_test`, public API only) |

White-box tests (unexported internals) stay next to the code they test
(`render_input_test.go`, `submit.*_test.go`); everything testable through the
public API belongs in `tests/`.

## Adding a New Built-in Input

1. Create `input/mytype.go` — embed `Base`, configure `model.Permitted` rules.
2. Constructor `MyType() Input` returns a stateless prototype (no arguments).
3. Add `Clone(parentID, name string) Input` on the concrete type
   (copy + `InitBase`).
4. `Name()`, `Storage()`, `Validate()` are inherited from `Base`; override
   `Validate` only for specialized rules (delegate the baseline to
   `m.Permitted.Validate(m.FieldName(), value)`).
5. Add validation cases in `input/validation_test.go` and the compile-time
   assertion in `input/assertions_test.go`.

## Test Layout

- `tests/` — black-box (`package form_test`): binding, rendering contract,
  submit-adjacent behavior via public API.
- Module root / `input/` — white-box tests only, for unexported internals.
- Shared native/wasm cases use the three-file pattern: `x.shared_test.go`
  (helpers), `x.back_test.go` (`//go:build !wasm`), `x.front_test.go`
  (`//go:build wasm`). Run with `gotest ./...` (native + wasm).
