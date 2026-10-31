package ui_test

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

// Compile-time checks: each component is the content it claims to be.
var (
	_ htemel.PhrasingContent = ui.Button(ui.Primary, "x")
	_ htemel.PhrasingContent = ui.IconButton(ui.Ghost, ui.IconX(), "x")
	_ htemel.PhrasingContent = ui.ButtonGroup()
	_ htemel.PhrasingContent = ui.Badge(ui.Neutral, "x")
	_ htemel.PhrasingContent = ui.Kbd("K")
	_ htemel.PhrasingContent = ui.Icon("fa-solid fa-check")
	_ htemel.PhrasingContent = ui.IconCheck()
	_ htemel.EmbeddedContent = ui.IconCheck()
	_ htemel.PhrasingContent = ui.Avatar("a.png", "A")
	_ htemel.PhrasingContent = ui.AvatarInitials("AL", "Ada Lovelace")
	_ htemel.FlowContent     = ui.Card()
	_ htemel.FlowContent     = ui.Alert(ui.Info, "x")
	_ htemel.FlowContent     = ui.AlertDismissible("s", ui.Info, "x")
	_ htemel.FlowContent     = ui.Stack(ui.Md)
	_ htemel.FlowContent     = ui.Inline(ui.Md)
	_ htemel.FlowContent     = ui.Container()
	_ htemel.FlowContent     = ui.Divider()
	_ htemel.FlowContent     = ui.DividerLabel("or")
)

func init() {
	addGolden(
		golden{"button-primary", ui.Button(ui.Primary, "Save")},
		golden{"button-secondary-sm", ui.Button(ui.Secondary, "Cancel", ui.WithSize(ui.Sm))},
		golden{"button-outline", ui.Button(ui.Outline, "Export")},
		golden{"button-ghost", ui.Button(ui.Ghost, "More")},
		golden{"button-danger-icon", ui.Button(ui.Danger, "Delete", ui.WithIcon(ui.IconTrash()))},
		golden{"button-link", ui.Button(ui.Link, "Learn more")},
		golden{"button-zero-variant", ui.Button("", "Default")},
		golden{"button-xl-trailing", ui.Button(ui.Primary, "Options", ui.WithSize(ui.Xl), ui.WithTrailingIcon(ui.IconChevronDown()))},
		golden{"button-loading", ui.Button(ui.Primary, "Save", ui.WithLoading("saving")).
			Attrs(datastar.OnClick(datastar.Post("/save")))},
		golden{"button-loading-icon", ui.Button(ui.Primary, "Add", ui.WithIcon(ui.IconPlus()), ui.WithLoading("adding"))},
		golden{"button-submit", ui.Button(ui.Primary, "Send").Type(html.ButtonTypeSubmit)},
		golden{"icon-button", ui.IconButton(ui.Ghost, ui.IconPencil(), "Edit", ui.WithSize(ui.Sm))},
		golden{"button-group", ui.ButtonGroup(
			ui.Button(ui.Secondary, "Day"), ui.Button(ui.Secondary, "Week"), ui.Button(ui.Secondary, "Month"))},
		golden{"badge-tones", htemel.Group(
			ui.Badge(ui.Neutral, "Draft"), ui.Badge(ui.Brand, "Team"), ui.Badge(ui.Info, "New"), ui.Badge(ui.Success, "Paid"),
			ui.Badge(ui.Warning, "Due"), ui.Badge(ui.Danger, "Failed"))},
		golden{"kbd", ui.Kbd("⌘K")},
		golden{"icon-fontawesome", ui.Icon("fa-solid fa-check")},
		golden{"icons-svg", htemel.Group(
			ui.IconCheck(), ui.IconX(), ui.IconChevronDown(), ui.IconSearch(), ui.IconPlus(),
			ui.IconTrash(), ui.IconPencil(), ui.IconGripVertical(), ui.IconSpinner(),
			ui.IconSun(), ui.IconMoon(), ui.IconMonitor(), ui.IconCopy())},
		golden{"avatar", ui.Avatar("/ada.png", "Ada Lovelace", ui.WithSize(ui.Lg))},
		golden{"avatar-initials", ui.AvatarInitials("AL", "Ada Lovelace", ui.WithSize(ui.Sm))},
		golden{"card", ui.Card(
			ui.CardHeader("Team", ui.Badge(ui.Success, "Active")),
			ui.CardBody(html.P().Text("Three members.")),
			ui.CardFooter(ui.Button(ui.Secondary, "Cancel"), ui.Button(ui.Primary, "Save")),
		)},
		golden{"alert-info", ui.Alert(ui.Info, "Heads up", html.P().Text("Maintenance on Sunday."))},
		golden{"alert-danger-title-only", ui.Alert(ui.Danger, "Payment failed")},
		golden{"alert-dismissible", ui.AlertDismissible("welcome", ui.Success, "Saved")},
		golden{"stack", ui.Stack(ui.Sm, html.P().Text("a"), html.P().Text("b"))},
		golden{"inline", ui.Inline(ui.Lg, ui.Badge(ui.Info, "a"), ui.Badge(ui.Info, "b"))},
		golden{"container", ui.Container(html.P().Text("content"))},
		golden{"divider", ui.Divider()},
		golden{"divider-label", ui.DividerLabel("or")},
	)
}
