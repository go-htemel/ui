package main

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

func alerts() htemel.FlowContent {
	return ui.Stack(ui.Sm,
		ui.Alert(ui.Neutral, "Draft saved", html.P().Text("Nobody else can see it yet.")),
		ui.Alert(ui.Brand, "New: saved views", html.P().Text("Keep your filters for next time.")),
		ui.Alert(ui.Info, "Scheduled maintenance", html.P().Text("Sunday 02:00–03:00 UTC.")),
		ui.Alert(ui.Success, "Payment received"),
		ui.Alert(ui.Warning, "Your trial ends in 3 days"),
		ui.Alert(ui.Danger, "Deployment failed", html.P().Text("Check the build log for details.")),
	)
}

func dismissibleAlert() htemel.FlowContent {
	return ui.Stack(ui.Sm,
		ui.AlertDismissible("welcomeOpen", ui.Info, "Welcome back!", html.P().Text("Close me; the signal hides me.")),
		html.Div(ui.Button(ui.Secondary, "Show again", ui.WithSize(ui.Sm)).
			Attrs(datastar.OnClick(ui.Open("welcomeOpen")))),
	)
}

func serverToast() htemel.FlowContent {
	// POST /demo/toast answers with an SSE event that appends a ui.Toast to ui.ToastRegion.
	return ui.Inline(ui.Sm,
		ui.Button(ui.Primary, "Push a toast", ui.WithLoading("pushing")).
			Attrs(datastar.OnClick(datastar.Post("/demo/toast"))),
		ui.Button(ui.Danger, "Push an error", ui.WithLoading("pushingError")).
			Attrs(datastar.OnClick(datastar.Post("/demo/toast?tone=danger"))),
	)
}
