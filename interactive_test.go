package ui_test

import (
	"strings"
	"testing"
	"time"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

var (
	_ htemel.FlowContent = ui.Dialog("d", "x")
	_ htemel.FlowContent = ui.DialogActions()
	_ htemel.FlowContent = ui.Dropdown("m", ui.Button(ui.Secondary, "x"))
	_ htemel.FlowContent = ui.MenuItem("x")
	_ htemel.FlowContent = ui.MenuLink("x", "/")
	_ htemel.FlowContent = ui.MenuDivider()
	_ htemel.FlowContent = ui.Tabs("t")
	_ htemel.FlowContent = ui.Disclosure("d", "x")
	_ htemel.FlowContent = ui.Accordion("a")
	_ htemel.FlowContent = ui.ToastRegion()
	_ htemel.FlowContent = ui.Toast("t", ui.Info, "x", "", 0)
	_ htemel.FlowContent = ui.Draggable("p")
	_ htemel.FlowContent = ui.DropZone("x")
)

func init() {
	addGolden(
		golden{"dialog", htemel.Group(
			ui.Button(ui.Danger, "Delete").Attrs(ui.Opens("confirmDelete")),
			ui.Dialog("confirmDelete", "Delete project?",
				html.P().Text("This cannot be undone."),
				ui.DialogActions(
					ui.Button(ui.Secondary, "Cancel").Attrs(datastar.OnClick(ui.Close("confirmDelete"))),
					ui.Button(ui.Danger, "Delete").Attrs(datastar.OnClick(datastar.Delete("/projects/1"))),
				),
			),
		)},
		golden{"dropdown", ui.Dropdown("rowMenu",
			ui.Button(ui.Secondary, "Options", ui.WithTrailingIcon(ui.IconChevronDown())),
			ui.MenuItem("Edit", datastar.OnClick(datastar.Get("/rows/1/edit"))),
			ui.MenuLink("Open", "/rows/1"),
			ui.MenuDivider(),
			ui.MenuItem("Delete", datastar.OnClick(datastar.Delete("/rows/1"))),
		)},
		golden{"tabs", ui.Tabs("settingsTab",
			ui.Tab("profile", "Profile", html.P().Text("Profile panel")),
			ui.Tab("billing", "Billing", html.P().Text("Billing panel")),
		)},
		golden{"tabs-empty", ui.Tabs("none")},
		golden{"disclosure", ui.Disclosure("showMore", "Show details", html.P().Text("Details"))},
		golden{"accordion", ui.Accordion("faq",
			ui.AccordionItem("ship", "When does it ship?", html.P().Text("Tomorrow.")),
			ui.AccordionItem("return", "Can I return it?", html.P().Text("Within 30 days.")),
		)},
		golden{"toast-region", ui.ToastRegion()},
		golden{"toast", ui.Toast("Saved-1", ui.Success, "Saved", "Your changes are live.", 5*time.Second)},
		golden{"toast-sticky", ui.Toast("42", ui.Danger, "Upload failed", "", 0)},
		golden{"draggable", ui.Draggable("card-42", html.P().Text("Card"))},
		golden{"dropzone", ui.DropZone(ui.DropInto("dragged", datastar.Post("/board/done")), html.P().Text("Done"))},
	)
}

func TestNextToastID(t *testing.T) {
	a, b := ui.NextToastID(), ui.NextToastID()
	if a == b || !strings.HasPrefix(a, "t") {
		t.Errorf("NextToastID returned %q then %q", a, b)
	}
}

func TestSignalExpressions(t *testing.T) {
	if got := ui.Toggle("open"); got != "$open = !$open" {
		t.Errorf("Toggle = %q", got)
	}
	if got := ui.Opens("x"); got.Name != "data-on:click__stop" || got.Value != "$x = true" {
		t.Errorf("Opens = %+v", got)
	}
}
