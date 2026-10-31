package main

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

func progressBars() htemel.FlowContent {
	return ui.Stack(ui.Md,
		ui.Progress(25, 100, ui.ProgressLabel("Storage used")),
		ui.Progress(60, 100, ui.ProgressLabel("Sync progress"), ui.ProgressTone(ui.Info), ui.ProgressSize(ui.Sm)),
		ui.Progress(35, 100, ui.ProgressLabel("Onboarding"), ui.ProgressTone(ui.Brand), ui.ProgressSize(ui.Sm)),
		ui.Progress(100, 100, ui.ProgressLabel("Backup complete"), ui.ProgressTone(ui.Success), ui.ProgressSize(ui.Lg)),
		ui.Progress(9, 10, ui.ProgressLabel("Error budget used"), ui.ProgressTone(ui.Danger)),
	)
}

func progressSignal() htemel.FlowContent {
	// The bar follows $upload. The first button sets it in the browser; the second posts it to
	// the server, which patches the signal back 25 higher.
	return ui.Stack(ui.Md,
		ui.ProgressSignal("upload", 100, ui.ProgressLabel("Upload")),
		html.P().Class("text-sm text-neutral-600 dark:text-neutral-400").
			Children(html.Span().Attrs(datastar.Text("$upload")), htemel.Text("% uploaded")),
		ui.Inline(ui.Sm,
			ui.Button(ui.Secondary, "+10 in the browser", ui.WithSize(ui.Sm)).
				Attrs(datastar.OnClick("$upload = Math.min(100, $upload + 10)")),
			ui.Button(ui.Primary, "+25 from the server", ui.WithSize(ui.Sm), ui.WithLoading("advancing")).
				Attrs(datastar.OnClick(datastar.Post("/demo/progress"))),
			ui.Button(ui.Ghost, "Reset", ui.WithSize(ui.Sm)).Attrs(datastar.OnClick("$upload = 0")),
		),
	)
}

func skeletonAndSpinner() htemel.FlowContent {
	return ui.Stack(ui.Lg,
		ui.Skeleton(4),
		ui.Inline(ui.Md,
			ui.Spinner(ui.Xs), ui.Spinner(ui.Sm), ui.Spinner(ui.Md), ui.Spinner(ui.Lg), ui.Spinner(ui.Xl),
		),
	)
}

func tooltips() htemel.FlowContent {
	return ui.Inline(ui.Lg,
		ui.Tooltip(ui.IconButton(ui.Secondary, ui.IconTrash(), "Delete"), "Delete this row"),
		ui.Tooltip(ui.Button(ui.Secondary, "Hover or focus me"), "Appears on hover and on keyboard focus"),
	)
}

func inputGroups() htemel.FlowContent {
	// Field labels the input inside the group: the id and aria-describedby go on the <input>.
	return ui.Stack(ui.Md,
		ui.Field("site", "Website",
			ui.InputGroup(htemel.Text("https://"), ui.Input(html.InputTypeText, "site"), htemel.Text(".com"))),
		ui.Field("price", "Price",
			ui.InputGroup(htemel.Text("$"), ui.Input(html.InputTypeText, "price"), htemel.Text("USD"))),
		ui.Field("search", "Search",
			ui.InputGroup(ui.IconSearch(), ui.Input(html.InputTypeSearch, "search"), nil)),
	)
}
