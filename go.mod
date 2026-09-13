module webtyp.com/form

go 1.25.2

require (
	webtyp.com/dom v0.13.16
	webtyp.com/fmt v1.0.0
	webtyp.com/input v0.0.6
	webtyp.com/model v0.1.9
	webtyp.com/widget v0.6.30
)

// widget.PartSubmit (submit button styling hook) — not published yet.

// TEMPORARY, local-only — webtyp/input's SetMasked/IsMasked and
// webtyp/widget's PartReveal (see docs/PLAN.md in veltylabs/mjosefa-cms)
// aren't tagged yet. Remove once they publish and `go get
// webtyp.com/input@latest webtyp.com/widget@latest`.
replace webtyp.com/input => ../input
