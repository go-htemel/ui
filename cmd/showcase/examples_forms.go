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
		ui.Checkbox("Send me the newsletter (checked)", datastar.Bind("newsletter"), htemel.Attr("checked", "")),
		ui.Checkbox("Locked by your admin", htemel.Attr("disabled", ""), htemel.Attr("checked", "")),
		ui.RadioGroup("Notify me by",
			ui.Radio("channel", "email", "Email", datastar.Bind("channel")),
			ui.Radio("channel", "sms", "SMS", datastar.Bind("channel")),
			ui.Radio("channel", "none", "Never", datastar.Bind("channel")),
			ui.Radio("channel", "fax", "Fax (unavailable)", htemel.Attr("disabled", "")),
		),
		ui.Inline(ui.Md,
			ui.Switch("Dark launches", datastar.Bind("darkLaunch")),
			ui.Badge(ui.Neutral, "off").Attrs(datastar.Text("$darkLaunch ? 'on' : 'off'")),
		),
		ui.Switch("Audit log (always on)", htemel.Attr("disabled", ""), htemel.Attr("checked", "")),
	)
}

func nativeInputs() htemel.FlowContent {
	// Plain ui.Input with other types: ui.css restyles the native parts (no spin buttons, a
	// themed search clear button, the picker icon following the color scheme, a file button
	// like a Secondary button, a field-height color swatch, a primary-colored slider).
	return html.Div(
		ui.Field("qty", "Quantity", ui.Input(html.InputTypeNumber, "qty").Value("3").Min("0").Step("1")),
		ui.Field("q", "Search", ui.Input(html.InputTypeSearch, "q").Value("htemel")),
		ui.Field("due", "Due date", ui.Input(html.InputTypeDate, "due").Value("2026-11-15")),
		ui.Field("at", "Start time", ui.Input(html.InputTypeTime, "at").Value("09:30")),
		ui.Field("avatar", "Avatar", ui.Input(html.InputTypeFile, "avatar").Accept("image/*")),
		ui.Field("accent", "Accent color", ui.Input(html.InputTypeColor, "accent").Value("#4f46e5")),
		ui.Field("volume", "Volume", ui.Input(html.InputTypeRange, "volume").Value("60").Min("0").Max("100")),
	).Class("grid gap-6 sm:grid-cols-2")
}
