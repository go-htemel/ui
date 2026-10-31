package ui_test

import (
	stdhtml "html"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

var _ htemel.Node = ui.ColorSchemeBootstrap()
var _ htemel.FlowContent = ui.ColorSchemeToggle()

func init() {
	addGolden(
		golden{"color-scheme-toggle", ui.ColorSchemeToggle()},
		golden{"color-scheme-toggle-labels", ui.ColorSchemeToggle(ui.WithLabels())},
		golden{"color-scheme-bootstrap", ui.ColorSchemeBootstrap()},
		golden{"color-scheme-bootstrap-nonce", ui.ColorSchemeBootstrap(ui.WithNonce("r4nd0m"))},
	)
}

// TestColorSchemeBootstrapScript checks the inline script is safe to embed and valid JavaScript.
func TestColorSchemeBootstrapScript(t *testing.T) {
	src := ui.ColorSchemeBootstrapSource()
	if strings.Contains(strings.ToLower(src), "</script") || strings.Contains(src, "<!--") {
		t.Errorf("the script contains a sequence that could end the <script> element")
	}
	if !strings.Contains(src, "'"+ui.ColorSchemeKey+"'") {
		t.Errorf("the script does not use ColorSchemeKey %q", ui.ColorSchemeKey)
	}
	out, err := htemel.RenderString(html.Head(ui.ColorSchemeBootstrap()))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, src) {
		t.Errorf("rendered script differs from ColorSchemeBootstrapSource (it was escaped or rewritten):\n%s", out)
	}
	if len(src) > 700 {
		t.Errorf("the script is %d bytes; it is inlined in every page, keep it small", len(src))
	}

	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not installed; skipping the syntax check")
	}
	f := filepath.Join(t.TempDir(), "bootstrap.js")
	if err := os.WriteFile(f, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if b, err := exec.Command(node, "--check", f).CombinedOutput(); err != nil {
		t.Errorf("node --check: %v\n%s", err, b)
	}
}

func TestColorSchemeBootstrapNonce(t *testing.T) {
	plain, _ := htemel.RenderString(ui.ColorSchemeBootstrap())
	if strings.Contains(plain, "nonce") {
		t.Errorf("no nonce was given, yet: %s", plain)
	}
	withNonce, _ := htemel.RenderString(ui.ColorSchemeBootstrap(ui.WithNonce("abc123"), nil))
	if !strings.HasPrefix(withNonce, `<script nonce="abc123">`) {
		t.Errorf("nonce not on the script tag: %.60s", withNonce)
	}
}

var ariaPressed = regexp.MustCompile(`aria-pressed="(true|false)"`)

func TestColorSchemeToggle(t *testing.T) {
	out, err := htemel.RenderString(ui.ColorSchemeToggle())
	if err != nil {
		t.Fatal(err)
	}
	out = stdhtml.UnescapeString(out) // attribute values escape quotes
	for _, want := range []string{
		`role="group"`, `aria-label="Color scheme"`,
		`data-signals:color-scheme__ifmissing="'system'"`,
		`data-init=`,
		`aria-label="System"`, `aria-label="Light"`, `aria-label="Dark"`,
		`data-attr:aria-pressed="$colorScheme === 'dark' ? 'true' : 'false'"`,
		`localStorage.setItem('color-scheme', 'dark')`,
		`localStorage.removeItem('color-scheme')`,
		`transition-colors duration-150`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}
	// Server-rendered state: exactly one pressed button (System) until the init expression runs.
	if got := ariaPressed.FindAllString(out, -1); strings.Join(got, ",") != `aria-pressed="true",aria-pressed="false",aria-pressed="false"` {
		t.Errorf("server-rendered aria-pressed = %v", got)
	}
	if strings.Count(out, "<button") != 3 {
		t.Errorf("want 3 buttons, got %d", strings.Count(out, "<button"))
	}
	labeled, _ := htemel.RenderString(ui.ColorSchemeToggle(ui.WithLabels()))
	for _, w := range []string{">System<", ">Light<", ">Dark<"} {
		if !strings.Contains(labeled, w) {
			t.Errorf("WithLabels: missing visible %q", w)
		}
	}
}
