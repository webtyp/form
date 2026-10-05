module webtyp.com/form

go 1.25.2

require (
	webtyp.com/css v0.4.27
	webtyp.com/dom v0.13.18
	webtyp.com/fmt v1.0.0
	webtyp.com/input v0.0.9
	webtyp.com/model v0.1.9
	webtyp.com/widget v0.6.34
)

require (
	webtyp.com/color v0.1.2 // indirect
	webtyp.com/font v0.0.5 // indirect
)

// widget.PartSubmit (submit button styling hook) — not published yet.

// TEMPORARY, local-only — webtyp/input's SetMasked/IsMasked and
// webtyp/widget's PartReveal (see docs/PLAN.md in veltylabs/mjosefa-cms)
// aren't tagged yet. Remove once they publish and `go get
// webtyp.com/input@latest webtyp.com/widget@latest`.
