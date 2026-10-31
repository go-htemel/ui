package ui_test

import (
	stdhtml "html"
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
	_ htemel.FlowContent = ui.Drawer("d", ui.DrawerLeft, "x")
	_ htemel.FlowContent = ui.DrawerActions()
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
	_ htemel.FlowContent = ui.Sortable()
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
		golden{"drawer-right", htemel.Group(
			ui.Button(ui.Secondary, "Edit profile").Attrs(ui.Opens("profileDrawer")),
			ui.Drawer("profileDrawer", ui.DrawerRight, "Edit profile",
				ui.Field("name", "Name", ui.Input(html.InputTypeText, "name")),
				ui.DrawerActions(
					ui.Button(ui.Secondary, "Cancel").Attrs(datastar.OnClick(ui.Close("profileDrawer"))),
					ui.Button(ui.Primary, "Save").Attrs(datastar.OnClick(datastar.Post("/profile"))),
				),
			),
		)},
		golden{"drawer-left", ui.Drawer("nav", ui.DrawerLeft, "Navigation", html.P().Text("Links"))},
		golden{"drawer-bottom", ui.Drawer("share", ui.DrawerBottom, "Share", html.P().Text("Copy the link."))},
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

// render returns the HTML of n with attribute entities decoded, so Datastar expressions can be
// matched as written.
func render(t *testing.T, n htemel.Node) string {
	t.Helper()
	got, err := htemel.RenderString(n)
	if err != nil {
		t.Fatal(err)
	}
	return stdhtml.UnescapeString(got)
}

// wantAll fails for each substring of want missing from got.
func wantAll(t *testing.T, got string, want ...string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("output lacks %q:\n%s", w, got)
		}
	}
}

func TestDialogAccessibility(t *testing.T) {
	got := render(t, ui.Dialog("confirmDelete", "Delete?", html.P().Text("x"),
		ui.DialogActions(ui.Button(ui.Secondary, "Cancel"), ui.Button(ui.Danger, "Delete"))))
	wantAll(t, got,
		`<dialog id="confirmDelete-dialog" aria-labelledby="confirmDelete-title" closedby="any"`, `id="confirmDelete-title"`,
		`backdrop:bg-neutral-950/50`, `m-auto`,
		`data-signals:confirm-delete__ifmissing="false"`,
		`data-effect="if (!$confirmDelete) { if (el.open) el.close(); return } if (el.open) return; el.showModal(); `, // signal -> dialog
		`data-on:close="$confirmDelete = false"`,                                                                      // Escape, closedby, method=dialog -> signal
		`data-on:click__outside="if ('closedBy' in HTMLDialogElement.prototype) return; $confirmDelete = false"`,
		`aria-label="Close" title="Close" data-on:click="$confirmDelete = false"`,
		`autofocus>Cancel</button>`, // initial focus on the first action
	)
	for _, gone := range []string{`role="dialog"`, `aria-modal`, `data-show`, `display: none`, `keydown`, `overflow`, `_ret`, ` open>`, ` open `, ` open=`} {
		if strings.Contains(got, gone) {
			t.Errorf("native dialog output still contains %q:\n%s", gone, got)
		}
	}
	if n := strings.Count(got, " autofocus>"); n != 1 {
		t.Errorf("want exactly one autofocus, got %d:\n%s", n, got)
	}

	// An autofocus the caller set is kept, and the first action does not get another.
	own := render(t, ui.DialogActions(ui.Button(ui.Secondary, "Cancel"), ui.Button(ui.Primary, "Save").Autofocus(true)))
	if n := strings.Count(own, "autofocus"); n != 1 || !strings.Contains(own, "autofocus>Save") {
		t.Errorf("caller's autofocus must win:\n%s", own)
	}
	// Only buttons and links receive it.
	if plain := render(t, ui.DialogActions(html.Span().Text("x"))); strings.Contains(plain, "autofocus") {
		t.Errorf("a non-control first child must not get autofocus:\n%s", plain)
	}
}

func TestDrawer(t *testing.T) {
	got := render(t, ui.Drawer("filters", ui.DrawerBottom, "Filters",
		html.P().Text("body"),
		ui.DrawerActions(ui.Button(ui.Primary, "Apply")),
		html.P().Text("more"),
	))
	wantAll(t, got,
		`<dialog id="filters-dialog" aria-labelledby="filters-title" closedby="any"`, `<h2 id="filters-title"`,
		`data-effect="if (!$filters) { if (el.open) el.close(); return } if (el.open) return; el.showModal(); if (el.querySelector('[autofocus]')) return; `,
		`[id='filters-body'] :is(input:not([type=hidden]), select, textarea):not(:disabled)`, // first control, else the dialog
		`f ? f.focus() : el.focus({preventScroll: true})`,
		`tabindex="-1"`, `outline-none`,
		`data-on:close="$filters = false"`,
		`aria-label="Close" title="Close" data-on:click="$filters = false"`,
		`m-0 mt-auto`, `translate-y-full open:translate-y-0 starting:open:translate-y-full`,
		`overflow-y-auto`, `motion-reduce:transition-none`,
		// The actions are pinned after the scrolling body, whatever their position among children.
		`<p>body</p><p>more</p></div><div class="flex shrink-0 justify-end`,
	)
	if strings.Contains(got, " autofocus") {
		t.Errorf("DrawerActions must not set autofocus:\n%s", got)
	}
	for side, want := range map[ui.DrawerSide]string{
		ui.DrawerRight: "m-0 ml-auto h-dvh", ui.DrawerLeft: "m-0 mr-auto h-dvh", ui.DrawerSide(99): "m-0 ml-auto h-dvh",
	} {
		if s := render(t, ui.Drawer("d", side, "x")); !strings.Contains(s, want) {
			t.Errorf("side %d: want %q in %s", side, want, s)
		}
	}
}

func TestDropdownAccessibility(t *testing.T) {
	got := render(t, ui.Dropdown("rowMenu", ui.Button(ui.Secondary, "Options"), ui.MenuItem("Edit"), ui.MenuLink("Open", "/x")))
	wantAll(t, got,
		`id="rowMenu-trigger" popovertarget="rowMenu-menu" popovertargetaction="toggle"`,
		`aria-haspopup="menu"`, `aria-controls="rowMenu-menu"`, `aria-expanded="false"`,
		`style="anchor-name: --rowMenu-anchor"`,
		`data-attr:aria-expanded="($rowMenu) ? 'true' : 'false'"`,
		`<div id="rowMenu-menu" popover="auto" role="menu" aria-labelledby="rowMenu-trigger" style="position-anchor: --rowMenu-anchor; position-area: bottom span-left; position-try-fallbacks: flip-block"`,
		`data-on:toggle="$rowMenu = evt.newState === 'open'"`, // popover -> signal
		`data-effect="if ($rowMenu !== el.matches(":popover-open")) { if (!$rowMenu) el.hidePopover(); else if (el.parentElement.checkVisibility()) el.showPopover() }"`, // signal -> popover, unless not rendered
		`data-on:beforetoggle="if (evt.newState !== 'open' || CSS.supports('position-anchor: --a')) return;`,
		`document.getElementById('rowMenu-trigger').getBoundingClientRect()`,
		`<button type="button" role="menuitem" tabindex="-1"`, `<a href="/x" role="menuitem" tabindex="-1"`,
		`data-on:keydown="`, `'ArrowDown', 'ArrowUp', 'Home', 'End'`, `[role='menuitem']`, `m.showPopover()`,
		`if (k === 'Tab') { if (open) m.hidePopover(); return }`,
		`data-on:click="el.hidePopover(); el.previousElementSibling.focus()"`, // close on item click, focus back
	)
	for _, gone := range []string{`display: none`, `data-show`, `__outside`, `__window`, `$rowMenu = !$rowMenu`} {
		if strings.Contains(got, gone) {
			t.Errorf("popover dropdown output still contains %q:\n%s", gone, got)
		}
	}
}

func TestTooltipAccessibility(t *testing.T) {
	got := render(t, ui.Tooltip(ui.Button(ui.Secondary, "x"), "Help"))
	wantAll(t, got, `aria-describedby="tooltip-`, `role="tooltip"`, `group-hover:opacity-100`, `group-focus-within:opacity-100`)
	if strings.Contains(got, "popover") {
		t.Errorf("Tooltip is CSS only:\n%s", got)
	}
}

func TestTabsAccessibility(t *testing.T) {
	got := render(t, ui.Tabs("settingsTab", ui.Tab("profile", "Profile"), ui.Tab("billing", "Billing")))
	wantAll(t, got,
		`role="tablist"`, `data-on:keydown="`, `'ArrowRight'`, `'ArrowLeft'`, `'Home'`, `'End'`,
		`id="settingsTab-profile-tab"`, `aria-controls="settingsTab-profile-panel"`,
		`id="settingsTab-profile-panel"`, `aria-labelledby="settingsTab-profile-tab"`, `role="tabpanel"`,
		`aria-selected="true"`, `aria-selected="false"`,
		`tabindex="0"`, `tabindex="-1"`, // roving tabindex, server-rendered
		`data-attr:tabindex="($settingsTab === 'profile') ? '0' : '-1'"`, // and kept in step
		`data-attr:aria-selected="($settingsTab === 'billing') ? 'true' : 'false'"`,
	)
}

func TestDisclosureAccessibility(t *testing.T) {
	got := render(t, ui.Disclosure("showMore", "More", html.P().Text("x")))
	wantAll(t, got,
		`id="showMore-button"`, `aria-controls="showMore-panel"`, `aria-expanded="false"`,
		`data-attr:aria-expanded="($showMore) ? 'true' : 'false'"`,
		`id="showMore-panel"`, `role="region"`, `aria-labelledby="showMore-button"`,
		// The panel follows the button's aria-expanded in CSS: collapsed grid row and
		// visibility: hidden until the peer button is expanded, then 1fr and visible.
		`class="peer `, `invisible grid grid-rows-[0fr]`, `peer-aria-expanded:visible peer-aria-expanded:grid-rows-[1fr]`,
		`<div class="min-h-0 overflow-hidden">`,
	)
	if strings.Contains(got, "data-show") || strings.Contains(got, "display: none") {
		t.Errorf("the panel must not be hidden with data-show (no height transition then):\n%s", got)
	}
	if strings.Contains(got, " hidden") {
		t.Errorf("panel must not use the hidden attribute (it would override the CSS state):\n%s", got)
	}
	acc := render(t, ui.Accordion("faq", ui.AccordionItem("a", "A"), ui.AccordionItem("b", "B")))
	wantAll(t, acc, `data-on:keydown="`, `'ArrowDown'`, `'ArrowUp'`, `aria-controls="faq-b-panel"`, `aria-labelledby="faq-b-button"`)
}

func TestDragAndDropAccessibility(t *testing.T) {
	drag := render(t, ui.Draggable("c1", html.P().Text("x")))
	wantAll(t, drag, `data-ui-drag `, `data-ui-payload="c1"`, `draggable="true"`, `aria-grabbed="false"`,
		`data-on:dragstart="`, `el.setAttribute('aria-grabbed', 'true')`, `el.toggleAttribute('data-ui-dragging', true)`,
		`data-on:dragend="`, `el.setAttribute('aria-grabbed', 'false')`)

	zone := render(t, ui.DropZone(ui.DropInto("dragged", "@post('/m')")))
	wantAll(t, zone,
		`data-on:dragenter="evt.preventDefault(); el.toggleAttribute('data-ui-over', true)"`,
		`data-on:dragover="evt.preventDefault();`,
		`data-on:dragleave="if (!el.contains(evt.relatedTarget)) el.toggleAttribute('data-ui-over', false)"`,
		`data-on:drop__prevent="el.toggleAttribute('data-ui-over', false); $dragged = evt.dataTransfer.getData('text/plain'); @post('/m')`,
		`data-ui-over:outline-2`, ` data-ui-dropzone `,
		`data-on:ui:drop="el.toggleAttribute('data-ui-over', false); $dragged = evt.detail.payload; @post('/m')`)

	move := render(t, ui.MoveHere("dragged", "Move here", "@post('/m')"))
	wantAll(t, move, `data-signals:dragged-via__ifmissing="''"`, `data-show="$draggedVia === 'pickup'"`, `style="display: none"`, `data-on:click="@post('/m'); setTimeout(() => { $dragged = ''; $draggedVia = '' })"`, ">Move here<")
	if got := ui.PickUp("dragged", "c1"); got.Name != "data-on:click" || got.Value != "$dragged = 'c1'; $draggedVia = 'pickup'" {
		t.Errorf("PickUp = %+v", got)
	}
}

func TestToastAccessibility(t *testing.T) {
	region := render(t, ui.ToastRegion())
	wantAll(t, region, `aria-live="polite"`, `aria-relevant="additions"`)

	got := render(t, ui.Toast("a", ui.Info, "Saved", "", 4*time.Second))
	wantAll(t, got, `role="status"`, `aria-label="Dismiss"`,
		`data-on-interval__duration.250ms="if (el.matches('*:hover, *:focus-within')) return;`, // pause on hover and focus
		`el._left ?? 4000`, `if (el._left <= 0) $toasts.a = false`)

	short := render(t, ui.Toast("b", ui.Info, "Hi", "", 100*time.Millisecond))
	wantAll(t, short, `data-on-interval__duration.100ms=`, `el._left ?? 100) - 100`)
	if sticky := render(t, ui.Toast("c", ui.Info, "Hi", "", 0)); strings.Contains(sticky, "data-on-interval") {
		t.Errorf("a toast without timeout must not tick:\n%s", sticky)
	}
}
