// Command safelist prints the Tailwind classes the ui components use, one per line. Tailwind
// generates only the classes it finds in your sources, and the ui sources live in the Go module
// cache, so add this list to your Tailwind sources instead:
//
//	go run github.com/go-htemel/ui/cmd/safelist > tailwind.safelist.txt
//
// The classes use design tokens (bg-primary-600, rounded-control, ...), so one list serves every
// theme; see cmd/themecss for the tokens.
//
// It reads the ui sources from -dir; by default from the current directory when that is the ui
// module, and otherwise from the ui module your go.mod requires.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-htemel/ui/internal/safelist"
)

func main() {
	dir := flag.String("dir", "", "directory of the ui package sources (default: detected)")
	flag.Parse()

	if *dir == "" {
		d, err := uiDir()
		if err != nil {
			fail(err)
		}
		*dir = d
	}
	classes, err := safelist.Classes(os.DirFS(*dir))
	if err != nil {
		fail(err)
	}
	fmt.Print(safelist.File(classes))
}

// uiDir returns the current directory when it holds the ui package, and otherwise asks the go
// command where the required ui module is.
func uiDir() (string, error) {
	if _, err := os.Stat(filepath.Join(".", "theme.go")); err == nil {
		if b, err := os.ReadFile("go.mod"); err == nil && strings.Contains(string(b), "module github.com/go-htemel/ui\n") {
			return ".", nil
		}
	}
	out, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/go-htemel/ui").Output()
	if err != nil {
		return "", fmt.Errorf("locating github.com/go-htemel/ui (pass -dir): %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "safelist:", err)
	os.Exit(1)
}
