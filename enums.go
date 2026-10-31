package ui

// Variant is the visual weight of a button.
type Variant string

// Button variants. The zero Variant renders as Primary.
const (
	Primary   Variant = "primary"
	Secondary Variant = "secondary"
	Outline   Variant = "outline"
	Ghost     Variant = "ghost"
	Link      Variant = "link"
)

// Danger is both a Variant (a destructive button) and a Tone (a red badge, alert or toast). It
// is an untyped constant so that it converts to either: ui.Button(ui.Danger, "Delete") and
// ui.Badge(ui.Danger, "Failed") both compile.
const Danger = "danger"

// Size scales a component. The zero Size renders as Md.
type Size string

// Sizes, from smallest to largest.
const (
	Xs Size = "xs"
	Sm Size = "sm"
	Md Size = "md"
	Lg Size = "lg"
	Xl Size = "xl"
)

// Tone is the semantic color of a badge, alert or toast. The zero Tone renders as Neutral. See
// Danger for the red tone.
type Tone string

// Tones.
const (
	Neutral Tone = "neutral"
	Info    Tone = "info"
	Success Tone = "success"
	Warning Tone = "warning"
)

// pick returns m[k], falling back to m[def] for an unknown or zero key.
func pick[K comparable](m map[K]string, k, def K) string {
	if v, ok := m[k]; ok {
		return v
	}
	return m[def]
}
