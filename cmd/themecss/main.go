// Command themecss prints the default go-htemel/ui design tokens (ui.ThemeCSS): a Tailwind v4
// @theme block defining the primary, neutral, success, warning, danger and info color scales and
// the --radius-control and --radius-surface tokens. Write it next to your app.css and import it
// after tailwindcss:
//
//	go run github.com/go-htemel/ui/cmd/themecss > assets/ui.theme.css
//
// Rerun it after upgrading the module; override tokens in your own CSS, not in this file.
package main

import (
	"fmt"

	"github.com/go-htemel/ui"
)

func main() {
	fmt.Print(ui.ThemeCSS())
}
