package ui

import (
	"github.com/go-htemel/htemel/html"
	"github.com/go-htemel/htemel/svg"
)

// Icon returns a Font Awesome icon, <i class="..." aria-hidden="true">. Load Font Awesome
// yourself; class is the icon's classes, such as "fa-solid fa-check".
func Icon(class string) *html.IElement {
	return html.I().Class(class).Aria("hidden", "true")
}

// Inline SVG icons, drawn on a 24x24 grid with a 2px currentColor stroke. Each returns the <svg>
// element, sized size-5 and hidden from assistive technology. Replace the size with SetClass (for
// example .SetClass("size-4 shrink-0")); Class would add to size-5 instead. Put a label on the
// surrounding control.

// stroked builds an outline icon from one path.
func stroked(d string) *svg.SvgElement {
	return svg.Svg(
		svg.Path().D(d),
	).ViewBox("0 0 24 24").
		Fill("none").
		Stroke("currentColor").
		StrokeWidth("2").
		StrokeLinecap(svg.StrokeLinecapRound).
		StrokeLinejoin(svg.StrokeLinejoinRound).
		Class("size-5 shrink-0").
		Aria("hidden", "true")
}

// IconCheck is a check mark.
func IconCheck() *svg.SvgElement { return stroked("M5 12.5l4.5 4.5L19 7.5") }

// IconCopy is two overlapping sheets, for copy-to-clipboard buttons.
func IconCopy() *svg.SvgElement {
	return stroked("M9 8h10a1 1 0 0 1 1 1v10a1 1 0 0 1-1 1H9a1 1 0 0 1-1-1V9a1 1 0 0 1 1-1zM16 8V5a1 1 0 0 0-1-1H5a1 1 0 0 0-1 1v10a1 1 0 0 0 1 1h3")
}

// IconX is a cross, for close and remove buttons.
func IconX() *svg.SvgElement { return stroked("M6 6l12 12M18 6L6 18") }

// IconChevronDown is a downward chevron, for disclosure and menu triggers.
func IconChevronDown() *svg.SvgElement { return stroked("M6 9l6 6 6-6") }

// IconSearch is a magnifying glass.
func IconSearch() *svg.SvgElement {
	return stroked("M11 4a7 7 0 1 0 0 14a7 7 0 1 0 0-14zM20 20l-4.2-4.2")
}

// IconPlus is a plus sign.
func IconPlus() *svg.SvgElement { return stroked("M12 5v14M5 12h14") }

// IconTrash is a waste bin.
func IconTrash() *svg.SvgElement {
	return stroked("M4 7h16M10 11v6M14 11v6M6 7l1 12a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2l1-12M9 7V4h6v3")
}

// IconPencil is a pencil, for edit buttons.
func IconPencil() *svg.SvgElement { return stroked("M4 20h4L19 9l-4-4L4 16v4zM13.5 6.5l4 4") }

// IconGripVertical is a two-column grip of dots, the handle of a draggable item.
func IconGripVertical() *svg.SvgElement {
	dot := func(cx, cy string) *svg.CircleElement { return svg.Circle().Cx(cx).Cy(cy).R("1.5") }
	return svg.Svg(
		dot("9", "6"), dot("15", "6"),
		dot("9", "12"), dot("15", "12"),
		dot("9", "18"), dot("15", "18"),
	).ViewBox("0 0 24 24").
		Fill("currentColor").
		Class("size-5 shrink-0").
		Aria("hidden", "true")
}

// IconSpinner is a spinning progress ring, used by WithLoading.
func IconSpinner() *svg.SvgElement {
	return svg.Svg(
		svg.Circle().Cx("12").Cy("12").R("9").Opacity(0.25),
		svg.Path().D("M21 12a9 9 0 0 0-9-9"),
	).ViewBox("0 0 24 24").
		Fill("none").
		Stroke("currentColor").
		StrokeWidth("3").
		StrokeLinecap(svg.StrokeLinecapRound).
		Class("size-5 shrink-0 animate-spin").
		Aria("hidden", "true")
}

// IconSun is a sun, for the light color scheme.
func IconSun() *svg.SvgElement {
	return stroked("M12 8a4 4 0 1 0 0 8a4 4 0 1 0 0-8zM12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4")
}

// IconMoon is a crescent moon, for the dark color scheme.
func IconMoon() *svg.SvgElement { return stroked("M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8z") }

// IconMonitor is a desktop monitor, for "follow the system" color scheme.
func IconMonitor() *svg.SvgElement {
	return stroked("M4 4h16a1 1 0 0 1 1 1v10a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V5a1 1 0 0 1 1-1zM8 20h8M12 16v4")
}
