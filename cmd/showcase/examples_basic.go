package main

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

func buttonVariants() htemel.FlowContent {
	return ui.Inline(ui.Sm,
		ui.Button(ui.Primary, "Primary"),
		ui.Button(ui.Secondary, "Secondary"),
		ui.Button(ui.Outline, "Outline"),
		ui.Button(ui.Ghost, "Ghost"),
		ui.Button(ui.Danger, "Danger"),
		ui.Button(ui.Link, "Link"),
	)
}

func buttonSizes() htemel.FlowContent {
	return ui.Inline(ui.Sm,
		ui.Button(ui.Primary, "Extra small", ui.WithSize(ui.Xs)),
		ui.Button(ui.Primary, "Small", ui.WithSize(ui.Sm)),
		ui.Button(ui.Primary, "Medium"),
		ui.Button(ui.Primary, "Large", ui.WithSize(ui.Lg)),
		ui.Button(ui.Primary, "Extra large", ui.WithSize(ui.Xl)),
	)
}

func buttonIcons() htemel.FlowContent {
	return ui.Inline(ui.Sm,
		ui.Button(ui.Primary, "Add item", ui.WithIcon(ui.IconPlus())),
		ui.Button(ui.Secondary, "Options", ui.WithTrailingIcon(ui.IconChevronDown())),
		ui.Button(ui.Danger, "Delete", ui.WithIcon(ui.Icon("fa-solid fa-trash"))),
	)
}

func buttonLoading() htemel.FlowContent {
	// The handler sleeps for a second, then pushes a toast over SSE.
	return ui.Button(ui.Primary, "Save", ui.WithIcon(ui.IconCheck()), ui.WithLoading("saving")).
		Attrs(datastar.OnClick(datastar.Post("/demo/slow")))
}

func iconButtons() htemel.FlowContent {
	return ui.Inline(ui.Sm,
		ui.IconButton(ui.Ghost, ui.IconPencil(), "Edit"),
		ui.IconButton(ui.Secondary, ui.IconTrash(), "Delete"),
		ui.IconButton(ui.Primary, ui.IconPlus(), "Add", ui.WithSize(ui.Lg)),
		ui.IconButton(ui.Outline, ui.IconSearch(), "Search", ui.WithSize(ui.Sm)),
	)
}

func buttonGroup() htemel.FlowContent {
	return html.Div(
		ui.ButtonGroup(
			ui.Button(ui.Secondary, "Day"),
			ui.Button(ui.Secondary, "Week"),
			ui.Button(ui.Secondary, "Month"),
		),
	)
}

func badges() htemel.FlowContent {
	return ui.Inline(ui.Sm,
		ui.Badge(ui.Neutral, "Draft"),
		ui.Badge(ui.Info, "New"),
		ui.Badge(ui.Success, "Paid"),
		ui.Badge(ui.Warning, "Due soon"),
		ui.Badge(ui.Danger, "Failed"),
	)
}

func avatars() htemel.FlowContent {
	return ui.Inline(ui.Md,
		ui.Avatar("/avatar.svg", "Grace Hopper", ui.WithSize(ui.Xl)),
		ui.AvatarInitials("AL", "Ada Lovelace", ui.WithSize(ui.Lg)),
		ui.AvatarInitials("KJ", "Katherine Johnson"),
		ui.AvatarInitials("MH", "Margaret Hamilton", ui.WithSize(ui.Sm)),
		ui.AvatarInitials("RF", "Rosalind Franklin", ui.WithSize(ui.Xs)),
	)
}

func keyboard() htemel.FlowContent {
	return html.P(
		htemel.Text("Press "), ui.Kbd("⌘"), htemel.Text(" "), ui.Kbd("K"),
		htemel.Text(" to search, or "), ui.Kbd("Esc"), htemel.Text(" to close a dialog."),
	).Class("text-sm text-neutral-700 dark:text-neutral-300")
}

func icons() htemel.FlowContent {
	return ui.Stack(ui.Md,
		ui.Inline(ui.Md,
			ui.IconCheck(), ui.IconX(), ui.IconChevronDown(), ui.IconSearch(),
			ui.IconPlus(), ui.IconTrash(), ui.IconPencil(), ui.IconGripVertical(), ui.IconSpinner(),
		).Class("flex flex-wrap items-center gap-4 text-neutral-700 dark:text-neutral-300"),
		ui.Inline(ui.Md,
			ui.Icon("fa-solid fa-house"), ui.Icon("fa-solid fa-gear"), ui.Icon("fa-regular fa-bell"),
		).Class("flex flex-wrap items-center gap-4 text-lg text-primary-600 dark:text-primary-400"),
	)
}

func dividers() htemel.FlowContent {
	return ui.Stack(ui.Lg,
		ui.Divider(),
		ui.DividerLabel("or continue with"),
	)
}

func stackAndInline() htemel.FlowContent {
	return ui.Stack(ui.Md,
		html.P().Class("text-sm font-medium").Text("Stack puts children in a column…"),
		ui.Inline(ui.Sm,
			ui.Badge(ui.Info, "…and Inline"),
			ui.Badge(ui.Info, "puts them"),
			ui.Badge(ui.Info, "in a wrapping row"),
		),
	)
}

func container() htemel.FlowContent {
	return ui.Container(
		html.P().Class("rounded-surface bg-neutral-100 p-4 text-center text-sm dark:bg-white/5").
			Text("Centred, at most max-w-7xl wide, with responsive padding."),
	)
}

func card() htemel.FlowContent {
	return ui.Card(
		ui.CardHeader("Team members", ui.Badge(ui.Success, "3 active")),
		ui.CardBody(
			ui.Stack(ui.Sm,
				ui.Inline(ui.Sm, ui.AvatarInitials("AL", "Ada Lovelace", ui.WithSize(ui.Sm)), htemel.Text("Ada Lovelace")),
				ui.Inline(ui.Sm, ui.AvatarInitials("GH", "Grace Hopper", ui.WithSize(ui.Sm)), htemel.Text("Grace Hopper")),
			),
		),
		ui.CardFooter(
			ui.Button(ui.Secondary, "Cancel"),
			ui.Button(ui.Primary, "Invite"),
		),
	)
}
