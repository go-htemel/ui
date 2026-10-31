package main

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

func dialog() htemel.FlowContent {
	return html.Div(
		ui.Button(ui.Danger, "Delete project", ui.WithIcon(ui.IconTrash())).Attrs(ui.Opens("confirmDelete")),
		ui.Dialog("confirmDelete", "Delete project?",
			html.P().Text("The project and its 12 deployments will be removed. This cannot be undone."),
			ui.DialogActions(
				ui.Button(ui.Secondary, "Cancel").Attrs(datastar.OnClick(ui.Close("confirmDelete"))),
				ui.Button(ui.Danger, "Delete").Attrs(datastar.OnClick(ui.Close("confirmDelete"))),
			),
		),
	)
}

func dropdown() htemel.FlowContent {
	// The menu is a native popover: it opens without the signal, which only mirrors its state.
	return ui.Inline(ui.Md,
		ui.Dropdown("projectMenu",
			ui.Button(ui.Secondary, "Options", ui.WithTrailingIcon(ui.IconChevronDown())),
			ui.MenuItem("Rename"),
			ui.MenuItem("Duplicate"),
			ui.MenuLink("Open settings", "#settings"),
			ui.MenuDivider(),
			ui.MenuItem("Archive", datastar.OnClick(datastar.Post("/demo/toast?title=Archived"))),
		),
		html.Span().Class("text-sm text-neutral-500 dark:text-neutral-400").
			Attrs(datastar.Text("$projectMenu ? 'Menu open' : 'Menu closed'")),
	)
}

func tabs() htemel.FlowContent {
	return ui.Tabs("settingsTab",
		ui.Tab("profile", "Profile", html.P().Class("text-sm").Text("Name, avatar and bio.")),
		ui.Tab("billing", "Billing", html.P().Class("text-sm").Text("Plan, invoices and payment methods.")),
		ui.Tab("team", "Team", html.P().Class("text-sm").Text("Members and their roles.")),
	)
}

func disclosure() htemel.FlowContent {
	return ui.Disclosure("showDetails", "Show technical details",
		html.P().Text("Rendered by htemel, styled by Tailwind, toggled by a Datastar signal."),
	)
}

func accordion() htemel.FlowContent {
	return ui.Accordion("faq",
		ui.AccordionItem("js", "Do I need to write JavaScript?", html.P().Text("No. Every interaction is a Datastar attribute.")),
		ui.AccordionItem("css", "How do I ship the CSS?", html.P().Text("Build it with the Tailwind CLI, using the ui safelist.")),
		ui.AccordionItem("ssr", "Is it server-rendered?", html.P().Text("Yes: components are Go functions that render HTML.")),
	)
}

func kanbanBoard() htemel.FlowContent {
	// Drag a card to another column: the drop posts {"dragged": "<card id>"} and the server
	// patches the whole board back. The board lives in server memory. Sortable adds touch
	// dragging when ui.js is loaded; the grip buttons (Pick up / Move here) work without it.
	return html.Div(
		ui.Sortable(demoBoard.view()),
		html.Div(ui.Button(ui.Ghost, "Reset board", ui.WithSize(ui.Sm)).
			Attrs(datastar.OnClick(datastar.Post("/demo/kanban/reset")))).Class("mt-4"),
	)
}
