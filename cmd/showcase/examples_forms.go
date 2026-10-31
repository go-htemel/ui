package main

import (
	"time"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

func fieldInput() htemel.FlowContent {
	return ui.Field("email", "Email address",
		ui.Input(html.InputTypeEmail, "email").Placeholder("you@example.com").Attrs(datastar.Bind("email")),
		ui.HelpText("We will only use it to send receipts."),
	)
}

func fieldValidation() htemel.FlowContent {
	// Each keystroke (debounced) posts the signals; the server patches $errors.username.
	return ui.Field("username", "Username",
		ui.Input(html.InputTypeText, "username").Attrs(
			datastar.Bind("username"),
			datastar.OnInput(datastar.Post("/demo/validate"), datastar.Debounce(300*time.Millisecond)),
		),
		ui.HelpText("At least 3 characters. Try \"admin\"."),
		ui.ErrorSignal("errors.username"),
	)
}

func fieldServerError() htemel.FlowContent {
	return ui.Field("company", "Company",
		ui.Input(html.InputTypeText, "company").Value("ACME"),
		ui.ErrorText("A company with this name already exists."),
	)
}

func textareaAndSelect() htemel.FlowContent {
	return ui.Stack(ui.Md,
		ui.Field("bio", "Bio", ui.Textarea("bio", 3).Placeholder("A few words about you").Attrs(datastar.Bind("bio"))),
		ui.Field("plan", "Plan", ui.Select("plan",
			ui.Choice("free", "Free"),
			ui.Choice("team", "Team").Selected(true),
			ui.Choice("enterprise", "Enterprise"),
		).Attrs(datastar.Bind("plan"))),
		html.P().Class("text-sm text-neutral-500 dark:text-neutral-400").Children(
			htemel.Text("Plan signal: "),
			html.Code().Attrs(datastar.Text("$plan")),
		),
	)
}

func choices() htemel.FlowContent {
	return ui.Stack(ui.Lg,
		ui.Checkbox("Remember me on this device", datastar.Bind("remember")),
		ui.RadioGroup("Notify me by",
			ui.Radio("channel", "email", "Email", datastar.Bind("channel")),
			ui.Radio("channel", "sms", "SMS", datastar.Bind("channel")),
			ui.Radio("channel", "none", "Never", datastar.Bind("channel")),
		),
		ui.Inline(ui.Md,
			ui.Switch("Dark launches", datastar.Bind("darkLaunch")),
			ui.Badge(ui.Neutral, "off").Attrs(datastar.Text("$darkLaunch ? 'on' : 'off'")),
		),
	)
}
