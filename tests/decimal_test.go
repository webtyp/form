package form_test

import (
	"testing"

	"webtyp.com/form"
	"webtyp.com/input"
	"webtyp.com/model"
)

// TestNumberAndDecimal_Storage pins the storage of each numeric widget: Number()
// must stay FieldInt (nothing about it changes) and Decimal() must be FieldFloat —
// the gap that let a service_catalog price field silently truncate to int64.
func TestNumberAndDecimal_Storage(t *testing.T) {
	if got := input.Number().Storage(); got != model.FieldInt {
		t.Errorf("input.Number().Storage() = %v, want FieldInt", got)
	}
	if got := input.Decimal().Storage(); got != model.FieldFloat {
		t.Errorf("input.Decimal().Storage() = %v, want FieldFloat", got)
	}
}

// priceRecord mirrors a monetary field declared with input.Decimal().
type priceRecord struct {
	Price float64
}

func (m *priceRecord) Schema() []model.Field {
	return []model.Field{
		{Name: "Price", Type: input.Decimal()},
	}
}

func (m *priceRecord) Pointers() []any { return []any{&m.Price} }

// TestDecimal_RoundTripsFractionalValue is the consumer-shaped test: a fractional
// value survives New -> LoadValues -> SyncValues without truncation.
func TestDecimal_RoundTripsFractionalValue(t *testing.T) {
	f, err := form.New("parent", &priceRecord{}, &testIDGen{})
	if err != nil {
		t.Fatalf("form.New: %v", err)
	}

	src := &priceRecord{Price: 49.99}
	if err := f.LoadValues(src); err != nil {
		t.Fatalf("LoadValues: %v", err)
	}

	dst := &priceRecord{}
	if err := f.SyncValues(dst); err != nil {
		t.Fatalf("SyncValues: %v", err)
	}

	// webtyp/fmt's float<->string conversion documents 6-decimal-place precision
	// (see fmt.Conv.wrFloatBase), not bit-exact float64 round-trip — a ~1e-13 diff
	// here is that library's known rounding, not a truncation to int. What THIS
	// test guards against is the real bug: Price silently becoming a whole number.
	const epsilon = 1e-6
	diff := dst.Price - 49.99
	if diff < 0 {
		diff = -diff
	}
	if diff > epsilon {
		t.Errorf("Price round-trip = %v, want ~49.99 (truncated to int?)", dst.Price)
	}
}
