package form_test

import (
	"testing"

	"webtyp.com/form"
	"webtyp.com/input"
	"webtyp.com/model"
)

// seqIDGen is the deterministic double for the hidden-PK tests: it mints
// "n1", "n2", … so each new record's id is distinguishable by eye.
type seqIDGen struct{ n int }

func (g *seqIDGen) NewID() string {
	g.n++
	if g.n == 1 {
		return "n1"
	}
	if g.n == 2 {
		return "n2"
	}
	return "n3"
}

// pkTextRecord carries a hidden text PK: the shape every real CRUD model has
// (an id the form never renders).
type pkTextRecord struct {
	Id   string
	Name string
}

func (r *pkTextRecord) Schema() []model.Field {
	return []model.Field{
		{Name: "id", Type: input.Text(), NotNull: true, DB: &model.FieldDB{PK: true}},
		{Name: "name", Type: input.Text()},
	}
}
func (r *pkTextRecord) Pointers() []any {
	if r == nil {
		return nil
	}
	return []any{&r.Id, &r.Name}
}
func (r *pkTextRecord) FormName() string { return "pktext" }
func (r *pkTextRecord) IsNil() bool      { return r == nil }

// pkIntRecord carries a hidden int PK (auto-increment lives in the DB).
type pkIntRecord struct {
	Id   int64
	Name string
}

func (r *pkIntRecord) Schema() []model.Field {
	return []model.Field{
		{Name: "id", Type: model.Int(), DB: &model.FieldDB{PK: true, AutoInc: true}},
		{Name: "name", Type: input.Text()},
	}
}
func (r *pkIntRecord) Pointers() []any {
	if r == nil {
		return nil
	}
	return []any{&r.Id, &r.Name}
}
func (r *pkIntRecord) FormName() string { return "pkint" }
func (r *pkIntRecord) IsNil() bool      { return r == nil }

// A loaded hidden PK travels with the record: syncing into a scratch object
// (what crudview does — it syncs into the shared Presenter.Record()) must
// write the LOADED id, never the scratch's own.
func TestHiddenPKLoadThenSyncIntoScratch(t *testing.T) {
	f, err := form.New("p", &pkTextRecord{}, &seqIDGen{})
	if err != nil {
		t.Fatal(err)
	}
	if err := f.LoadValues(&pkTextRecord{Id: "a", Name: "Alice"}); err != nil {
		t.Fatal(err)
	}
	scratch := &pkTextRecord{Id: "zzz", Name: ""}
	if err := f.SyncValues(scratch); err != nil {
		t.Fatal(err)
	}
	if scratch.Id != "a" {
		t.Errorf("expected the loaded id 'a' to win over the scratch's 'zzz', got %q", scratch.Id)
	}
	if scratch.Name != "Alice" {
		t.Errorf("expected name 'Alice', got %q", scratch.Name)
	}
}

// A reset form holds no record: syncing afterwards mints a fresh id instead
// of re-shipping whatever the scratch already carries.
func TestHiddenPKResetThenSyncMintsFresh(t *testing.T) {
	f, err := form.New("p", &pkTextRecord{}, &seqIDGen{})
	if err != nil {
		t.Fatal(err)
	}
	f.Reset()
	scratch := &pkTextRecord{Id: "a", Name: ""}
	if err := f.SyncValues(scratch); err != nil {
		t.Fatal(err)
	}
	if scratch.Id != "n1" {
		t.Errorf("expected a freshly minted id 'n1', got %q", scratch.Id)
	}
}

// Two syncs with no reset in between are a retry, not two records: the id
// minted on the first one is kept.
func TestHiddenPKRetryKeepsMintedID(t *testing.T) {
	f, err := form.New("p", &pkTextRecord{}, &seqIDGen{})
	if err != nil {
		t.Fatal(err)
	}
	f.Reset()
	first := &pkTextRecord{}
	if err := f.SyncValues(first); err != nil {
		t.Fatal(err)
	}
	second := &pkTextRecord{}
	if err := f.SyncValues(second); err != nil {
		t.Fatal(err)
	}
	if first.Id != "n1" || second.Id != "n1" {
		t.Errorf("expected both syncs to carry 'n1', got %q and %q", first.Id, second.Id)
	}
}

// Reset → sync → reset → sync mints two distinct ids: two drafts, two records.
func TestHiddenPKTwoDraftsGetDistinctIDs(t *testing.T) {
	f, err := form.New("p", &pkTextRecord{}, &seqIDGen{})
	if err != nil {
		t.Fatal(err)
	}
	f.Reset()
	first := &pkTextRecord{}
	if err := f.SyncValues(first); err != nil {
		t.Fatal(err)
	}
	f.Reset()
	second := &pkTextRecord{}
	if err := f.SyncValues(second); err != nil {
		t.Fatal(err)
	}
	if first.Id != "n1" || second.Id != "n2" {
		t.Errorf("expected 'n1' then 'n2', got %q and %q", first.Id, second.Id)
	}
}

// Loading nil is the "new record" case: it clears the remembered id exactly
// like Reset does.
func TestHiddenPKNilLoadClearsRememberedID(t *testing.T) {
	f, err := form.New("p", &pkTextRecord{}, &seqIDGen{})
	if err != nil {
		t.Fatal(err)
	}
	if err := f.LoadValues(&pkTextRecord{Id: "a", Name: "Alice"}); err != nil {
		t.Fatal(err)
	}
	if err := f.LoadValues(nil); err != nil {
		t.Fatal(err)
	}
	scratch := &pkTextRecord{Id: "a"}
	if err := f.SyncValues(scratch); err != nil {
		t.Fatal(err)
	}
	if scratch.Id != "n1" {
		t.Errorf("expected a freshly minted id 'n1' after a nil load, got %q", scratch.Id)
	}
}

// An int hidden PK is remembered the same way; a new one stays zero —
// auto-increment is the DB's job, not the form's.
func TestHiddenPKIntRoundTrip(t *testing.T) {
	f, err := form.New("p", &pkIntRecord{}, &seqIDGen{})
	if err != nil {
		t.Fatal(err)
	}
	if err := f.LoadValues(&pkIntRecord{Id: 7, Name: "Seven"}); err != nil {
		t.Fatal(err)
	}
	scratch := &pkIntRecord{Id: 99}
	if err := f.SyncValues(scratch); err != nil {
		t.Fatal(err)
	}
	if scratch.Id != 7 {
		t.Errorf("expected the loaded id 7 to win over the scratch's 99, got %d", scratch.Id)
	}
	f.Reset()
	fresh := &pkIntRecord{Id: 99}
	if err := f.SyncValues(fresh); err != nil {
		t.Fatal(err)
	}
	if fresh.Id != 0 {
		t.Errorf("expected a zero int PK after reset (DB assigns it), got %d", fresh.Id)
	}
}
