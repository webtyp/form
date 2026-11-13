module webtyp.com/form

go 1.26.8

require (
	webtyp.com/components v0.0.0-00010101000000-000000000000
	webtyp.com/css v0.4.30
	webtyp.com/dom v0.13.29
	webtyp.com/fmt v1.0.0
	webtyp.com/input v0.0.36
	webtyp.com/lang v0.1.16
	webtyp.com/model v0.2.21
	webtyp.com/widget v0.6.46
)

require (
	webtyp.com/color v0.1.2 // indirect
	webtyp.com/escape v0.1.0 // indirect
	webtyp.com/font v0.0.5 // indirect
	webtyp.com/html v0.0.30 // indirect
	webtyp.com/icons v0.0.33 // indirect
	webtyp.com/json v0.5.40 // indirect
	webtyp.com/svg v0.3.30 // indirect
)

// TEMP local verification (picker) — remove before commit.

// TEMP local verification (tapinfotip, label-row) — remove before commit.
replace webtyp.com/components => ../components

replace webtyp.com/icons => ../icons

replace webtyp.com/widget => ../widget
