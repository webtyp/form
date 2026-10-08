---
PLAN: "fix(form): SetOptions finds the field by name, not by == between interfaces"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 8050262862025602113
PR: https://github.com/webtyp/form/pull/25
---

# Plan — `Form.SetOptions` sin `==` entre interfaces

> Master: `webtyp/docs/NO_REFLECTION_MASTER_PLAN.md` (ola 1). Sin API pública nueva.

## 1. El problema

`form.go:401-405`:

```go
for _, c := range f.children {
	if fc, ok := c.(*fieldComponent); ok && fc.Input == inp {
		fc.refreshOptions()
	}
}
```

`fc.Input` e `inp` son `input.Input` (interfaz). En TinyGo, `==` entre interfaces compila a
`runtime.interfaceEqual` → `reflectValueEqual` → mete `internal/reflectlite` (~7 KB) en el binario
wasm de toda app que use formularios.

## 2. La corrección

`inp` viene de `f.Input(fieldName)`, que devuelve el input cuyo `FieldName()` es `fieldName`
(`form.go:382-391`). La condición equivalente, comparando strings:

```go
for _, c := range f.children {
	if fc, ok := c.(*fieldComponent); ok && fc.Input.FieldName() == fieldName {
		fc.refreshOptions()
	}
}
```

Luego buscar en todo el módulo (`*.go` que compilan a wasm, sin tests) otros `==`/`!=`/`switch` entre
valores de interfaz con operandos no nil, y resolverlos igual: comparando un campo concreto (string,
puntero de tipo concreto), nunca la interfaz.

## 3. Tests

- `tests/`: un formulario con dos campos con opciones; `SetOptions("b", …)` refresca solo el campo
  `b` (las opciones de `a` no cambian). Debe pasar antes y después (protege la equivalencia).
- `gotest` verde (vet, race, tests, wasm).

## 4. Criterios de aceptación

- `grep -n 'fc.Input == inp' form.go` → vacío.
- Ningún símbolo exportado nuevo: `git diff | grep '^+func [A-Z]'` → vacío.
- `gotest` verde.

## 5. Restricciones

Las de `AGENTS.md`, más: nada de `reflect`, nada de `unsafe`, ningún `==`/`!=`/`switch` entre
valores de interfaz con operandos no nil.
