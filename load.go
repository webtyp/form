package form

import (
	"webtyp.com/fmt"
	"webtyp.com/model"
)

// LoadValues populates every input from data, the inverse of SyncValues.
// It is the operation a CRUD view needs when the user selects a record: one call,
// no per-field string conversion at the call site.
//
// A nil data (including a typed-nil pointer inside the interface) resets the form —
// that is the "new record" case, not an error.
func (f *Form) LoadValues(data model.Fielder) error {
	if model.IsNil(data) {
		f.reset()
		return nil
	}

	values := model.ReadValues(data.Schema(), data.Pointers())

	for i, inp := range f.Inputs {
		idx := f.fieldIndices[i]
		if idx < 0 || idx >= len(values) {
			continue
		}

		val := fmt.Convert(values[idx]).String()

		// Signal is the source of truth in WASM mode.
		f.valueSignals[i].Set(val)
		f.errorSignals[i].Set("") // loading a record clears stale validation errors
		f.baseline[i] = val       // a freshly loaded record is pristine — see IsDirty

		// Keep input internal state in sync for SSR mode.
		if setter, ok := inp.(interface{ SetValues(...string) }); ok {
			setter.SetValues(val)
		}
	}

	// The hidden PK has no Input and never reaches the loop above — remember
	// it here, or SyncValues would have no identity to write (see sync.go).
	// A nil record reset the whole form at the top; there is nothing to keep.
	for k, idx := range f.hiddenPKIndices {
		if idx >= 0 && idx < len(values) {
			f.loadedPK[k] = fmt.Convert(values[idx]).String()
		}
	}

	return nil
}
