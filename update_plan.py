import re

with open('docs/PLAN.md', 'r') as f:
    content = f.read()

stage4_update = """
| `PartRadioGroup` | `Row(Space2)` — no `As(Field)`, `ControlBox`, `PadEdge` |
| `PartRadioOption` | `Button(style.Panel)`, `Round(RadiusMd)` |

States:
- `When(widget.Selected, widget.PartRadioOption, style.As(style.Primary))`
- `PartRadioUnchecked` and `PartRadioChecked` get `Hide()`
- Delete the two `WhenWithin(Selected, ...)` glyph rules for radio
- Native radio stays `VisuallyHidden()`
- Locked: `WhenWithin(widget.Locked, "", widget.PartRadioOption, style.As(style.Inactive))` replaces the locked rule on `PartRadioGroup`.
"""
content = re.sub(
    r'\| `PartRadioGroup` \| as today without `PadEdge\(EdgeTop, Space2\)` \|',
    stage4_update.strip(),
    content
)

tests_update = """
- `radio_chips_test.go`: one option per choice with the option class, the selected one carries the selected state, and the stylesheet gives it the primary surface.
"""

content = re.sub(
    r'part has no `position: absolute`\.',
    'part has no `position: absolute`.\n' + tests_update.strip(),
    content
)

with open('docs/PLAN.md', 'w') as f:
    f.write(content)
