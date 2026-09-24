package form

import (
	"webtyp.com/fmt"
	"webtyp.com/model"
)

// SyncValues copies all input values back into the bound struct
// via the Fielder's Pointers() method.
func (f *Form) SyncValues(data model.Fielder) error {
	pointers := data.Pointers()
	schema := data.Schema()

	for i, inp := range f.Inputs {
		idx := f.fieldIndices[i]
		if idx < 0 || idx >= len(pointers) {
			continue
		}

		// Signal is the source of truth in WASM mode.
		val := f.valueSignals[i].Get()

		// Fallback only if we are somehow in SSR mode where signals might be empty
		// but input state is populated (though f.Render() should handle this).
		if val == "" && f.ssrMode {
			if getter, ok := inp.(interface{ GetValues() []string }); ok {
				vals := getter.GetValues()
				if len(vals) > 0 {
					val = vals[0]
				}
			}
		}

		values := []string{val}

		ptr := pointers[idx]
		field := schema[idx]

		if val == "" {
			// Zero the field
			zeroField(ptr, field.Type.Storage())
			continue
		}

		writeField(ptr, field.Type.Storage(), values)
	}

	// The hidden PK is form state: LoadValues remembered it, Reset cleared it.
	// The target's own PK is never read — the target may be a scratch record
	// shared across edits (crudview syncs into Presenter.Record()), and trusting
	// it wrote edits onto whichever record was synced last.
	for k, idx := range f.hiddenPKIndices {
		ptr, st := pointers[idx], schema[idx].Type.Storage()
		if f.loadedPK[k] == "" && st == model.FieldText {
			f.loadedPK[k] = f.idGen.NewID() // a new record keeps this id across retries until Reset
		}
		if f.loadedPK[k] == "" {
			zeroField(ptr, st) // int PK of a new record: auto-increment is the DB's job
			continue
		}
		writeField(ptr, st, []string{f.loadedPK[k]})
	}

	return nil
}

// zeroField sets a field to its zero value via its pointer.
func zeroField(ptr any, ft model.FieldType) {
	switch ft {
	case model.FieldText:
		if p, ok := ptr.(*string); ok {
			*p = ""
		}
	case model.FieldInt:
		if p, ok := ptr.(*int64); ok {
			*p = 0
		}
	case model.FieldFloat:
		if p, ok := ptr.(*float64); ok {
			*p = 0
		}
	case model.FieldBool:
		if p, ok := ptr.(*bool); ok {
			*p = false
		}
	}
}

// writeField writes string values into a field via its pointer.
func writeField(ptr any, ft model.FieldType, values []string) {
	switch ft {
	case model.FieldText:
		if p, ok := ptr.(*string); ok {
			*p = values[0]
		}
	case model.FieldInt:
		if p, ok := ptr.(*int64); ok {
			val, _ := fmt.Convert(values[0]).Int64()
			*p = val
		}
	case model.FieldFloat:
		if p, ok := ptr.(*float64); ok {
			val, _ := fmt.Convert(values[0]).Float64()
			*p = val
		}
	case model.FieldBool:
		if p, ok := ptr.(*bool); ok {
			val, _ := fmt.Convert(values[0]).Bool()
			*p = val
		}
	}
}
