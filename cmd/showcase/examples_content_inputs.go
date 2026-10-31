package main

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

func tagInputs() htemel.FlowContent {
	// Enter or a comma adds a tag, Backspace removes the last; each chip holds a hidden
	// <input name="tags">, and TagSignal mirrors the list into $postTags.
	return ui.Stack(ui.Md,
		ui.Field("tags", "Tags", ui.TagInput("tags", []string{"go", "htmx"},
			ui.TagSuggestions("go", "css", "datastar", "tailwind", "htmx", "writing"),
			ui.TagMax(5), ui.TagSignal("postTags")),
			ui.HelpText("Up to five tags.")),
		html.P().Class("text-sm text-neutral-600 dark:text-neutral-400").
			Attrs(datastar.Text("'Signal: ' + JSON.stringify($postTags)")).Text("Signal: [\"go\",\"htmx\"]"),
	)
}

func fileUploads() htemel.FlowContent {
	// A label-wrapped file input (works without script) with a drop area, a file list and image
	// previews from ui.js. The second picker simulates an upload: FileUploadAction runs per file
	// and drives a ProgressSignal; a real app would post evt.detail.file.
	simulate := "$uploadName = evt.detail.file.name; $uploadPct = 0; " +
		"const t = setInterval(() => { $uploadPct = Math.min(100, $uploadPct + 20); if ($uploadPct >= 100) clearInterval(t) }, 150)"
	return ui.Stack(ui.Lg,
		ui.Field("cover", "Cover image", ui.FileUpload("cover", ui.FileAccept("image/*"), ui.FileHint("PNG, JPG or SVG, up to 5 MB"))),
		html.Div(
			ui.Field("attachments", "Attachments", ui.FileUpload("attachments", ui.FileMultiple(), ui.FileUploadAction(simulate))),
			html.Div(
				ui.ProgressSignal("uploadPct", 100, ui.ProgressLabel("Upload")),
				html.P().Class("mt-1 text-xs text-neutral-600 dark:text-neutral-400").Attrs(datastar.Text("$uploadName ? 'Uploading ' + $uploadName : 'Nothing uploaded yet'")).Text("Nothing uploaded yet"),
			).Class("mt-3").Attrs(datastar.Signals(map[string]any{"uploadName": "", "uploadPct": 0}, datastar.IfMissing())),
		),
	)
}

func dateTimeInputs() htemel.FlowContent {
	// Native pickers styled like Input; min, max and step are options.
	return html.Div(html.Div(
		ui.Field("publishOn", "Publish on", ui.DateInput("publishOn", ui.DateMin("2026-01-01"), ui.DateValue("2026-10-31"))),
		ui.Field("eventAt", "Event starts", ui.DateTimeInput("eventAt", ui.DateStep("900"))),
		ui.Field("reminder", "Daily reminder", ui.TimeInput("reminder", ui.DateValue("08:30"))),
	).Class("grid gap-4 @xl:grid-cols-3")).Class("@container")
}

func slugField() htemel.FlowContent {
	// The slug follows the title until you type in it; clear it to follow again. Pure Datastar:
	// a data-effect and an input handler, no script file.
	return ui.Stack(ui.Md,
		ui.Field("postTitle", "Title", ui.Input(html.InputTypeText, "title").Placeholder("Hello, Wörld!").Attrs(datastar.Bind("postTitle"))),
		ui.Field("postSlug", "Slug", ui.InputGroup(htemel.Text("/blog/"), ui.SlugField("postSlug", "postTitle"), nil),
			ui.HelpText("Follows the title until you edit it.")),
	).Attrs(datastar.Signal("postTitle", "''", datastar.IfMissing()))
}

func filterBar() htemel.FlowContent {
	// FilterSearch GETs /demo/people 300ms after the last keystroke with $q; the team select and
	// Reset do the same. The server filters, sorts and pages, and patches the body and the pager.
	get := datastar.Get(peopleURL)
	return html.Div(
		ui.FilterBar(
			ui.FilterSearch("q", "$page = 1; "+get),
			ui.Select("team",
				ui.Choice("", "All teams"), ui.Choice("Platform", "Platform"), ui.Choice("Growth", "Growth"),
				ui.Choice("Core", "Core"), ui.Choice("Design", "Design"),
			).Attrs(datastar.Bind("team"), datastar.OnChange("$page = 1; "+get), htemel.Attr("aria-label", "Team")),
			ui.Button(ui.Ghost, "Reset").Attrs(datastar.OnClick(ui.FilterReset("q", "team")+"; $page = 1; "+get)),
		),
		html.Div(ui.Table(
			ui.TableHead(
				ui.SortableTh("Name", "name", "sort", get),
				ui.SortableTh("Role", "role", "sort", get),
				ui.SortableTh("Team", "team", "sort", get),
			),
			peopleBody(peopleQuery{SortDir: "asc", Page: 1}),
			ui.TableCaption("People, filtered on the server"),
		)).Class("mt-4"),
		html.Div(peoplePager(peopleQuery{Page: 1})).Class("mt-4"),
	).Attrs(datastar.Signals(map[string]any{"q": "", "team": "", "sort": "", "sortDir": "asc", "page": 1}))
}
