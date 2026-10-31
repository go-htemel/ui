package main

import (
	"fmt"
	"strings"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

// demoThemes are the alternative token sets of the theming page. Each one only overrides CSS
// variables under [data-theme="<name>"]; the Go code and the generated utilities stay the same.
// Radii are CSS values: Tailwind v4 has --radius-xs to --radius-4xl but no --radius-none or
// --radius-full, so square and pill corners are written as 0 and calc(infinity * 1px).
var demoThemes = []struct {
	name, primary, neutral, control, surface, font string
}{
	{name: "indigo", primary: "indigo", control: "var(--radius-lg)", surface: "var(--radius-xl)"},
	{name: "emerald", primary: "emerald", neutral: "stone", control: "calc(infinity * 1px)", surface: "var(--radius-2xl)"},
	{name: "editorial", primary: "rose", neutral: "slate", control: "0", surface: "0",
		font: `ui-serif, Georgia, Cambria, "Times New Roman", serif`},
}

// themeDemoCSS is the page CSS that defines demoThemes. The showcase puts it in the
// <style type="text/tailwindcss"> block of every page, after the default tokens.
var themeDemoCSS = func() string {
	var b strings.Builder
	steps := []int{50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950}
	for _, t := range demoThemes {
		fmt.Fprintf(&b, "[data-theme=%q] {\n", t.name)
		for _, fam := range [][2]string{{"primary", t.primary}, {"neutral", t.neutral}} {
			if fam[1] == "" {
				continue
			}
			for _, s := range steps {
				fmt.Fprintf(&b, "  --color-%s-%d: var(--color-%s-%d);\n", fam[0], s, fam[1], s)
			}
		}
		fmt.Fprintf(&b, "  --radius-control: %s;\n  --radius-surface: %s;\n", t.control, t.surface)
		if t.font != "" {
			fmt.Fprintf(&b, "  --font-sans: %s;\n", t.font)
		}
		b.WriteString("}\n")
	}
	return b.String()
}()

func themeSwitcher() htemel.FlowContent {
	// The select writes $theme; the effect copies it to <html data-theme="...">, and the CSS
	// variables defined under [data-theme="..."] (shown below) retheme every component on the page.
	return ui.Stack(ui.Lg,
		ui.Field("theme", "Theme", ui.Select("theme",
			ui.Choice("", "Default (sky, zinc, md corners)"),
			ui.Choice("indigo", "Indigo (lg and xl corners)"),
			ui.Choice("emerald", "Emerald on stone (pill controls)"),
			ui.Choice("editorial", "Editorial (rose on slate, square, serif)"),
		).Attrs(
			datastar.Bind("theme"),
			datastar.Effect("$theme ? (document.documentElement.dataset.theme = $theme) : delete document.documentElement.dataset.theme"),
		)),
		ui.Inline(ui.Sm,
			ui.Button(ui.Primary, "Primary"),
			ui.Button(ui.Secondary, "Secondary"),
			ui.Button(ui.Outline, "Outline"),
			ui.Button(ui.Danger, "Delete"),
			ui.Badge(ui.Info, "Info"),
			ui.Badge(ui.Success, "Success"),
			ui.Badge(ui.Warning, "Warning"),
		),
		ui.Field("theme-email", "Email", ui.Input(html.InputTypeEmail, "theme-email").Placeholder("you@example.com")),
		ui.Alert(ui.Info, "Tokens, not rewrites", htemel.Text("The markup is identical in every theme; only CSS variables change.")),
		ui.Card(
			ui.CardHeader("Card", ui.Button(ui.Ghost, "Edit", ui.WithSize(ui.Sm))),
			ui.CardBody(ui.Progress(60, 100, ui.ProgressLabel("Upload"))),
		),
	)
}

func tokenSwatches() htemel.FlowContent {
	// One row per color token family, steps 50 to 950. The classes name tokens (bg-primary-600),
	// so the swatches follow the theme selected above.
	rows := ui.Stack(ui.Sm)
	for _, fam := range []string{"primary", "neutral", "success", "warning", "danger", "info"} {
		row := html.Div().Class("grid grid-cols-11 gap-1")
		for _, step := range []int{50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950} {
			row.Children(html.Div().
				Class(fmt.Sprintf("h-8 rounded-control bg-%s-%d ring-1 ring-black/5 ring-inset dark:ring-white/10", fam, step)).
				Title(fmt.Sprintf("--color-%s-%d", fam, step)))
		}
		rows.Children(html.Div(
			html.P().Class("mb-1 font-mono text-xs text-neutral-500 dark:text-neutral-400").Text(fam),
			row,
		))
	}
	return rows
}
