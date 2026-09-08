---
PLAN: "refactor(form): typed label/control association, stop reusing HandlerName as a DOM id"
TAG: v0.4.8
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 12706326592200080333
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> **GATE — do not dispatch until `webtyp.com/dom v0.13.11` is published.** It
> introduces `(*Element).Ref()` and makes a keyed element addressable. Run
> `go get webtyp.com/dom@v0.13.11` first.

# PLAN — `form` stops hand-wiring DOM ids

## Why `form` is a different case from the other consumers

`webtyp/components` and `webtyp/layout` invent ids per render, so two instances
collide. **`form` does not**: its ids come from `input.Input`'s stable model
identity (`"app.form.nombre.error"`), which is deliberate and meaningful.

So this plan does **not** strip ids from `form`. It fixes two narrower defects
that the audit found:

1. **Associations are hand-wired through strings** where `dom` already offers a
   typed method. `render_input.go:101` writes `Attr("for", fc.Input.GetID())` —
   a string round-trip that silently produces a dangling `for=` if the control's
   id ever changes. `(*Element).For(other)` exists precisely for this.
2. **`HandlerName()` is being used as a DOM id** (`render_input.go:179`, and as
   the prefix at `:215`). In this ecosystem `HandlerName()` means *the module's
   identity* — `crudp/interfaces.go:32` and `rbac/rbac.go:299` both declare it
   for routing and permissions. Reusing it as an element id conflates two
   concepts on one method, so a change made for routing silently moves DOM ids.

**`ID()` is not forbidden** and this plan adds no check. Stable, model-derived
ids on form controls are a legitimate use.

## Rules for this plan

- **Never invent a replacement upstream symbol.** If a site needs something
  `dom` does not expose (notably an `aria-describedby` / `list=` equivalent of
  `(*Element).For`), **stop and report it** in the PR description as a `dom`
  gap — do not declare a local helper. Per `CONSTRUCTION_HARNESS.md`: *"A missing
  contract at a boundary is a defect in the library, not in the consumer."*
- **Do NOT change `webtyp/input`'s public interface.** `ErrorID()` and
  `HandlerName()` stay declared in `input/interface.go` — other libraries depend
  on `HandlerName()`. Changing them is a separate `input` plan, if it turns out
  to be warranted; this plan only stops `form` from *misusing* one of them.
- **No `map`** (TinyGo binary budget) and **no standard library** — use
  `webtyp/fmt`.
- `gotest ./...` must stay green, `wasm ✅` included.

## Stage 1 — audit first, and write the findings down

Before editing, produce the inventory in the PR description. For each of the five
sites, state what actually consumes the id:

| File:line | Element | Question to answer |
|---|---|---|
| `render.go:47` | submit `<button>`, `ID(f.id + ".submit")` | does anything read this id — CSS, a test, a `Get`? |
| `render_input.go:127` | error `<span>`, `ID(fc.Input.ErrorID())` | is it referenced by an `aria-describedby` anywhere? (the audit found none) |
| `render_input.go:179` | `<select>`, `ID(fc.Input.HandlerName())` | is it the target of the `for=` written at `:101`? |
| `render_input.go:220` | radio `<input>`, `ID(optID)` | is it the target of a `<label for>`? |
| `render_input.go:275` | `<datalist>`, `ID(listID)` | consumed by `Attr("list", listID)` at `:259` |

`grep -rn "<the literal or method>" .` for each. **An id nothing reads is dead
weight** — say so, and Stage 3 deletes it.

## Stage 2 — typed label→control association

`render_input.go:101` currently writes:

```go
			Attr("for", fc.Input.GetID()).
```

Restructure so the control element is built **before** its label, then:

```go
	label := dom.NewElement("label").For(control)
```

`For` sets `for=` to the control's dom-assigned id and mints one if needed, so
the two can never disagree. Apply the same to the radio/label pairs around
`:215-225`: build the `radio` element, then `dom.NewElement("label").For(radio)`,
and delete the `optID` string entirely.

If a control element is not reachable at the point the label is built, restructure
the function so it is — do **not** keep the string round-trip.

## Stage 3 — stop using `HandlerName()` as a DOM id

At `render_input.go:179` the `<select>` takes `ID(fc.Input.HandlerName())`.

- If Stage 2 made the label point at the element via `For`, the `<select>` no
  longer needs an author-chosen id at all: **delete the `ID(...)` line**. Keep
  `Attr("name", fc.Input.FieldName())` — `name` is the form-submission key and is
  unrelated to ids.
- If the audit in Stage 1 found a real reader of that id, replace it with
  `fc.Input.GetID()` (the field's own identity) rather than `HandlerName()`, and
  say in the PR description why the id must survive.

Either way, after this stage `grep -rn "HandlerName()" form/` returns **no site
that feeds a DOM id**.

Do the same judgement for `render.go:47` and `render_input.go:127`: if Stage 1
found no reader, delete the `ID(...)`; if the element needs to stay addressable
from Go, give it `Key(...)` and reach it with `Ref()`.

## Stage 4 — the `<datalist>` pairing

`render_input.go:252-275` builds `listID := fc.Input.GetID() + "-list"`, writes
`Attr("list", listID)` on the control and `ID(listID)` on the `<datalist>`. This
is the same string round-trip as `for=`, but `dom` has **no `List(other)`
counterpart to `For(other)`**.

**Do not write a local helper.** Keep this pairing as it is for now, add a
one-line comment marking it, and **report the missing `dom` contract in the PR
description**:

> `dom` exposes `(*Element).For(other)` for `for=` but nothing equivalent for
> `list=` / `aria-describedby` / `aria-labelledby`. `form` needs those; until
> `dom` provides them the pairing stays a string.

That report is a deliverable of this plan, not a failure of it.

## Tests

- Existing tests keep passing. `form/tests/render.shared_test.go:50` asserts
  `ErrorID()` returns `"app.form.nombre.error"` — that test is about
  `input.Base`'s naming and **must keep passing unchanged**; this plan does not
  touch `input`.
- Update only assertions that referenced an id this plan deleted; prefer
  selecting on the part class or `name=`.
- **New test `TestLabelFollowsControlID`**: render a field, extract the control's
  rendered `id`, and assert the label's `for=` equals it exactly. This is the
  regression for the string round-trip Stage 2 removes.
- **New WASM test `TestTwoFormsSameModel_NoIDCollision`**: build two forms from
  the same model and render them as siblings in one pass. Assert no
  `dom.claimID` panic. If it **does** panic, that is a real finding: report it in
  the PR description and leave the test skipped with a comment naming the
  colliding id — do not paper over it with a per-instance prefix.

## Acceptance criteria

- `grep -rn 'Attr("for"' --include=*.go form/` → **empty** (all pairings go
  through `For`).
- `grep -rn 'HandlerName()' --include=*.go form/` → no site feeding a DOM id.
- `webtyp/input` is untouched: `git diff --stat` shows no change outside this repo.
- `gotest ./...` green, `wasm ✅` and `race ✅` included.
- `gofmt -l .` → empty; `go vet ./...` clean.
- The PR description carries the Stage 1 inventory and the Stage 4 `dom` gap report.

## Out of scope

- `webtyp/input`'s public interface (`ErrorID`, `HandlerName`, `GetID`).
- `webtyp/dom` — upstream, already published.
- `webtyp/components`, `webtyp/layout` — their own plans.
- Any check or panic forbidding `ID()`.

## Stages

| # | Files | Change |
|---|---|---|
| 1 | — | audit the five id sites; inventory into the PR description |
| 2 | `render_input.go` | `Attr("for", …)` → `For(control)`; delete `optID` |
| 3 | `render_input.go`, `render.go` | stop feeding DOM ids from `HandlerName()`; delete ids nothing reads |
| 4 | `render_input.go` | keep the `list=` pairing, report the missing `dom` contract |
