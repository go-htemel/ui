package ui_test

import (
	"os"
	"strings"
	"testing"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"

	"github.com/go-htemel/ui"
)

var (
	_ ui.CompositeControl = ui.TagInput("tags", nil)
	_ ui.CompositeControl = ui.FileUpload("file")
	_ ui.FormControl      = ui.DateInput("d")
	_ ui.FormControl      = ui.DateTimeInput("d")
	_ ui.FormControl      = ui.TimeInput("d")
	_ ui.FormControl      = ui.SlugField("slug", "title")
	_ ui.FormControl      = ui.FilterSearch("q", "")
	_ htemel.FlowContent  = ui.FilterBar()
	_ htemel.FlowContent  = ui.Field("tags", "Tags", ui.TagInput("tags", nil))
)

func init() {
	addGolden(
		golden{"tag-input", ui.Field("tags", "Tags", ui.TagInput("tags", []string{"go", "css"},
			ui.TagSuggestions("htmx", "datastar"), ui.TagMax(5), ui.TagSignal("tags")), ui.HelpText("Up to five."))},
		golden{"tag-input-plain", ui.TagInput("labels", nil, ui.TagPlaceholder("Add a label"))},
		golden{"file-upload", ui.Field("cover", "Cover image", ui.FileUpload("cover", ui.FileAccept("image/*"), ui.FileHint("PNG or JPG, up to 5 MB")))},
		golden{"file-upload-multiple", ui.FileUpload("files", ui.FileMultiple(), ui.FileUploadAction("upload(evt.detail.file)"), ui.FileBrowseLabel("Add attachments"))},
		golden{"date-input", ui.DateInput("published", ui.DateMin("2026-01-01"), ui.DateMax("2026-12-31"), ui.DateValue("2026-10-31"))},
		golden{"datetime-input", ui.DateTimeInput("at", ui.DateStep("900"))},
		golden{"time-input", ui.TimeInput("t", ui.DateStep("1"))},
		golden{"slug-field", ui.SlugField("slug", "title")},
		golden{"filter-bar", ui.FilterBar(
			ui.FilterSearch("q", datastar.Get("/people")),
			ui.Select("team", ui.Choice("", "All teams")).Attrs(datastar.Bind("team")),
			ui.Button(ui.Ghost, "Reset").Attrs(datastar.OnClick(ui.FilterReset("q", "team")+"; "+datastar.Get("/people"))),
		)},
	)
}

func TestTagInput(t *testing.T) {
	got := render(t, ui.Field("tags", "Tags", ui.TagInput("tags", []string{"go", "a&b"}, ui.TagMax(3), ui.TagSignal("postTags")), ui.HelpText("h")))
	wantAll(t, got,
		`<ui-taginput class="flex`, `data-max="3" data-name="tags">`,
		`<ul role="list" aria-label="Selected tags" class="contents">`,
		`data-ui-tag="go"`, `<input type="hidden" name="tags" value="go">`, `aria-label="Remove go"`,
		`value="a&b"`,
		// Field's label names the text field, and Field's help joins the component's own hint.
		`<label for="tags"`, `id="tags"`, `aria-describedby="tags-help tags-tag-hint"`,
		`Up to 3 tags.`, `aria-live="polite" data-ui-tag-status`,
		`<template data-ui-tag-template>`,
		`data-signals:post-tags__ifmissing="["go","a\u0026b"]"`, `data-on:ui:tags="$postTags = evt.detail.tags"`,
	)
	if strings.Count(got, `type="hidden"`) != 3 { // two chips + the template chip
		t.Errorf("want 2 hidden inputs plus the template's: %s", got)
	}
	if plain := render(t, ui.TagInput("t", nil)); strings.Contains(plain, "data-on:") || strings.Contains(plain, "datalist") {
		t.Errorf("no signal or datalist unless asked: %s", plain)
	}
}

func TestFileUpload(t *testing.T) {
	got := render(t, ui.Field("cover", "Cover", ui.FileUpload("cover", ui.FileAccept("image/*"), ui.FileHint("PNG"))))
	wantAll(t, got,
		`<ui-fileupload class="block">`, `data-ui-dropzone`,
		`<label class="relative cursor-pointer`, `<span>Choose a file</span><input type="file" name="cover" class="sr-only" accept="image/*" aria-describedby="cover-file-hint" id="cover">`,
		`<label for="cover"`, ` or drag it here`,
		`<template data-ui-file-template>`, `data-ui-file-remove`, `aria-live="polite" data-ui-file-status`,
	)
	if strings.Contains(got, "data-ui-upload") {
		t.Error("data-ui-upload without FileUploadAction")
	}
	multi := render(t, ui.FileUpload("f", ui.FileMultiple(), ui.FileUploadAction("go(evt.detail.file)")))
	wantAll(t, multi, `data-on:ui:upload="go(evt.detail.file)"`, `<ui-fileupload class="block" data-ui-upload>`, ` multiple`, `Choose files`, `drag them here`)
}

func TestDateInputs(t *testing.T) {
	wantAll(t, render(t, ui.DateInput("d", ui.DateMin("2026-01-01"), ui.DateMax("2026-02-01"))), `type="date"`, `min="2026-01-01"`, `max="2026-02-01"`, `ui-control`)
	wantAll(t, render(t, ui.DateTimeInput("d")), `type="datetime-local"`)
	wantAll(t, render(t, ui.TimeInput("d", ui.DateStep("900"))), `type="time"`, `step="900"`)
}

func TestSlugField(t *testing.T) {
	got := render(t, ui.SlugField("slug", "title"))
	wantAll(t, got,
		`data-signals:slug__ifmissing="''"`, `data-signals:slug-touched__ifmissing="false"`, `data-bind:slug`,
		`data-effect="if (!$slugTouched) $slug = String($title ?? '').toLowerCase()`,
		`.split('-').filter(Boolean).join('-')"`,
		`data-on:input="$slugTouched = el.value !== ''"`,
	)
	// slugify uses no regex anchor: Datastar rewrites "$" followed by a word into a signal read.
	_, eff, _ := strings.Cut(got, `data-effect="`)
	eff, _, _ = strings.Cut(eff, `" `)
	if strings.Count(eff, "$") != 3 { // $slugTouched, $slug, $title
		t.Errorf("slugify expression has a stray $: %s", eff)
	}
}

func TestFilter(t *testing.T) {
	got := render(t, ui.FilterBar(ui.FilterSearch("q", "@get('/people')")))
	wantAll(t, got, `role="search" aria-label="Filters"`, `type="search" name="q"`, `aria-label="Search"`,
		`data-on:input__debounce.300ms="@get('/people')"`, `data-bind:q`)
	if got := ui.FilterReset("q", "team"); got != "$q = ''; $team = ''" {
		t.Errorf("FilterReset = %q", got)
	}
}

// contentInputRows are the contrast pairs of TagInput and FileUpload.
func contentInputRows() []spec {
	var m []spec
	add := func(component, state, src string, k kind, light, dark side) {
		m = append(m, spec{component: component, state: state, src: src, kind: k, light: light, dark: dark})
	}
	// TagInput chips (taginput.go:69-72) on the white field / white-5% field; the dark field is
	// approximated by the slightly lighter neutral-800 under the chip (white/10 over white/5 over a
	// neutral-900 card is darker), so the light chip text is measured on a lighter background.
	add("TagInput", "chip text", "taginput.go:69", text, on("text-neutral-800", "bg-neutral-100").over(cardL), on("dark:text-neutral-100", "dark:bg-white/10").over("dark:bg-neutral-800"))
	add("TagInput", "chip remove icon", "taginput.go:70,taginput.go:72", nonText, on("text-neutral-600", "bg-neutral-100").over(cardL), on("dark:text-neutral-300", "dark:bg-white/10").over("dark:bg-neutral-800"))
	add("TagInput", "chip remove icon, hover", "taginput.go:71,taginput.go:72", nonText, on("hover:text-neutral-900", "bg-neutral-200").over(cardL), on("dark:hover:text-white", "dark:hover:bg-white/15").over("dark:bg-neutral-800"))
	add("TagInput", "focus-within outline", "taginput.go:67,taginput.go:68", nonText, on("focus-within:outline-primary-600", "").over(cardL), on("dark:focus-within:outline-primary-500", "").over(cardD))

	// FileUpload (fileupload.go:59-71) on its white / white-5% drop area over a card.
	const zoneL, zoneD = "bg-white", "dark:bg-white/5"
	add("FileUpload", "drop area text", "fileupload.go:65", text, on("text-neutral-600", zoneL).over(cardL), on("dark:text-neutral-400", zoneD).over(cardD))
	add("FileUpload", "browse link", "fileupload.go:62,fileupload.go:64", text, on("text-primary-600", zoneL).over(cardL), on("dark:text-primary-400", zoneD).over(cardD))
	add("FileUpload", "browse link, hover", "fileupload.go:62,fileupload.go:64", text, on("hover:text-primary-700", zoneL).over(cardL), on("dark:hover:text-primary-300", zoneD).over(cardD))
	add("FileUpload", "browse focus ring", "fileupload.go:63,fileupload.go:64", nonText, on("has-focus-visible:outline-primary-600", zoneL).over(cardL), on("dark:has-focus-visible:outline-primary-500", zoneD).over(cardD))
	add("FileUpload", "hint", "fileupload.go:66", text, on("text-neutral-600", zoneL).over(cardL), on("dark:text-neutral-400", zoneD).over(cardD))
	add("FileUpload", "drop outline (data-over)", "fileupload.go:61", nonText, on("data-over:outline-primary-600", "").over(cardL), on("dark:data-over:outline-primary-500", "").over(cardD))
	add("FileUpload", "file name", "fileupload.go:68", text, on("text-neutral-900", "").over(cardL), on("dark:text-white", "").over(cardD))
	add("FileUpload", "file size", "fileupload.go:69", text, on("text-neutral-600", "").over(cardL), on("dark:text-neutral-400", "").over(cardD))
	return m
}

// The elements the content inputs rely on are defined by ui.js, with the events they document.
func TestContentInputScripts(t *testing.T) {
	b, err := os.ReadFile("assets/ui.js")
	if err != nil {
		t.Fatal(err)
	}
	wantAll(t, string(b),
		"customElements.define('ui-taginput'", "customElements.define('ui-fileupload'",
		"new CustomEvent('ui:tags'", "new CustomEvent('ui:upload'",
		"addEventListener('formdata'", "URL.revokeObjectURL", "new DataTransfer()",
	)
}
