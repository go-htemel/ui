// Command safelist prints the Tailwind classes the ui components use, one per line, for a theme.
// Run it from the module root (make tailwind-safelist does):
//
//	go run ./internal/cmd/safelist -primary indigo -radius lg > tailwind.safelist.txt
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/go-htemel/ui"
	"github.com/go-htemel/ui/internal/safelist"
)

func main() {
	t := ui.DefaultTheme
	dir := flag.String("dir", ".", "directory of the ui package sources")
	flag.StringVar(&t.Primary, "primary", t.Primary, "primary Tailwind color family")
	flag.StringVar(&t.Radius, "radius", t.Radius, "Tailwind radius suffix (none, sm, md, lg, xl, 2xl, full)")
	flag.StringVar(&t.Font, "font", t.Font, "font-family class")
	flag.Parse()

	classes, err := safelist.Classes(os.DirFS(*dir), t)
	if err != nil {
		fmt.Fprintln(os.Stderr, "safelist:", err)
		os.Exit(1)
	}
	fmt.Print(safelist.File(classes))
}
