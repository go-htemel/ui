package ui_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

var (
	_ ui.CompositeControl = ui.MarkdownEditor("x")
	_ htemel.FlowContent  = ui.MarkdownEditor("x")
	_ htemel.FlowContent  = ui.MarkdownPreview("x")
)

func init() {
	addGolden(
		golden{"mdeditor", ui.MarkdownEditor("notes")},
		golden{"mdeditor-tabs", ui.MarkdownEditor("body", ui.WithPreview(datastar.Post("/preview")), ui.WithValue("Hello **world**"))},
		golden{"mdeditor-split", ui.MarkdownEditor("post-body", ui.WithPreview(datastar.Post("/preview")), ui.WithSplit(), ui.WithSignal("postBody"))},
		golden{"mdeditor-options", ui.MarkdownEditor("entry",
			ui.WithRows(4), ui.WithPlaceholder("Dear diary…"), ui.WithMaxLength(10), ui.WithValue("Nine char"),
			ui.WithAutosize(), ui.WithMonospace(), ui.WithEditorID("journal"),
			ui.WithToolbar(ui.MarkdownBold, ui.MarkdownItalic, ui.MarkdownSeparator, ui.MarkdownLink),
			ui.WithUploadAction("console.log(evt.detail.file.name)"),
			ui.WithHint("Formatting help", "/help/markdown"))},
		golden{"mdeditor-bare", ui.MarkdownEditor("plain", ui.WithToolbar(), ui.WithHint("", ""), ui.WithMaxLength(5), ui.WithValue("too long"))},
		golden{"mdeditor-field", ui.Field("body", "Entry", ui.MarkdownEditor("body"), ui.HelpText("Write freely."))},
		golden{"markdown-preview", ui.MarkdownPreview("body-preview", ui.Prose(html.P().Text("Rendered.")))},
		golden{"markdown-preview-empty", ui.MarkdownPreview("body-preview")},
	)
}

func TestMarkdownEditorField(t *testing.T) {
	got := render(t, ui.Field("entry-text", "Entry", ui.MarkdownEditor("body", ui.WithPreview(datastar.Post("/p"))),
		ui.HelpText("h"), ui.ErrorSignal("errors.body")))
	wantAll(t, got,
		`<label for="entry-text"`,
		`<textarea name="body" rows="8" id="entry-text"`,
		// Field's help and error first, then the editor's own hint and counter.
		`aria-describedby="entry-text-help entry-text-error-live body-hint body-count"`,
		`data-attr:aria-invalid="$errors.body ? 'true' : false"`,
		// The toolbar follows the id Field gave the textarea.
		`role="toolbar" aria-label="Formatting"`, `aria-controls="entry-text"`,
		// The rest is named after the editor id (the name), and PreviewID matches.
		`id="body-preview" role="region" aria-label="Preview" aria-live="polite"`,
		`id="body-write-tab"`, `aria-controls="body-preview-panel"`,
	)
	if id := ui.MarkdownEditor("body", ui.WithPreview("x")).PreviewID(); id != "body-preview" {
		t.Errorf("PreviewID = %q", id)
	}
	if id := ui.MarkdownEditor("body").PreviewID(); id != "" {
		t.Errorf("PreviewID without a preview = %q", id)
	}
}

func TestMarkdownEditorAttrsTargetTextarea(t *testing.T) {
	got := render(t, ui.MarkdownEditor("body").Attrs(htemel.Attr("required", ""), htemel.Attr("aria-describedby", "extra")))
	wantAll(t, got, `<textarea name="body" rows="8" id="body"`, `data-bind:body`, `required`, `aria-describedby="extra body-hint body-count"`)
	if strings.Count(got, "required") != 1 || !regexp.MustCompile(`<textarea[^>]*required`).MatchString(got) {
		t.Error("required is not on the textarea")
	}
}

var (
	toolbarRe  = regexp.MustCompile(`(?s)<div role="toolbar".*?</div></div>`)
	buttonRe   = regexp.MustCompile(`<button[^>]*>`)
	dataAttrRe = regexp.MustCompile(`\sdata-([a-z:-]+)=`)
)

// The toolbar carries no script: each button names its action, and <ui-mdeditor> runs it.
func TestMarkdownToolbarHasNoInlineJS(t *testing.T) {
	got := render(t, ui.MarkdownEditor("body"))
	bar := toolbarRe.FindString(got)
	if bar == "" {
		t.Fatalf("no toolbar in %s", got)
	}
	buttons := buttonRe.FindAllString(bar, -1)
	if len(buttons) != 12 {
		t.Errorf("%d toolbar buttons, want 12", len(buttons))
	}
	actions := map[string]bool{}
	for i, b := range buttons {
		for _, m := range dataAttrRe.FindAllStringSubmatch(b, -1) {
			if m[1] != "ui-md-action" {
				t.Errorf("button %d has data-%s: %s", i, m[1], b)
			}
		}
		if strings.Contains(b, " on") || strings.Contains(b, "javascript:") {
			t.Errorf("button %d has an inline handler: %s", i, b)
		}
		m := regexp.MustCompile(`data-ui-md-action="([a-z]+)"`).FindStringSubmatch(b)
		if m == nil {
			t.Errorf("button %d has no data-ui-md-action: %s", i, b)
			continue
		}
		actions[m[1]] = true
		wantAll(t, b, `aria-label="`, `title="`)
		if (i == 0) != strings.Contains(b, `tabindex="0"`) {
			t.Errorf("button %d: roving tabindex wrong: %s", i, b)
		}
	}
	for _, a := range []string{"bold", "italic", "strike", "heading", "quote", "code", "link", "image", "ul", "ol", "task", "table"} {
		if !actions[a] {
			t.Errorf("no %s button", a)
		}
	}
	wantAll(t, bar, `title="Bold (Ctrl+B)"`, `aria-keyshortcuts="Control+B Meta+B"`, `role="separator" aria-orientation="vertical"`)
}

func TestMarkdownEditorModes(t *testing.T) {
	plain := render(t, ui.MarkdownEditor("body"))
	if strings.Contains(plain, `role="tab"`) || strings.Contains(plain, "Preview") {
		t.Error("an editor without WithPreview has tabs or a preview")
	}
	tabs := render(t, ui.MarkdownEditor("body", ui.WithPreview("@post('/p')")))
	wantAll(t, tabs,
		`data-signals:_body-mode__ifmissing="'write'"`,
		`data-on:click="$_bodyMode = 'preview'; @post('/p')"`,
		`data-show="$_bodyMode === 'write'"`, // the toolbar and the write panel
		`role="tabpanel" id="body-preview-panel"`, `style="display: none"`,
		`starting:opacity-0 motion-reduce:transition-none`,
	)
	if strings.Contains(tabs, "data-on:input") || strings.Contains(tabs, "data-init") {
		t.Error("tabs mode should fetch the preview on tab switch only")
	}
	split := render(t, ui.MarkdownEditor("body", ui.WithPreview("@post('/p')"), ui.WithSplit()))
	wantAll(t, split, `data-on:input__debounce.400ms="@post('/p')"`, `data-init="@post('/p')"`, `@xl:grid-cols-2`, `@container`)
	if strings.Contains(split, `role="tab"`) {
		t.Error("split mode has tabs")
	}
	if strings.Contains(render(t, ui.MarkdownEditor("body", ui.WithSplit())), "@xl:grid-cols-2") {
		t.Error("WithSplit without WithPreview should be ignored")
	}
}

func TestMarkdownEditorCounter(t *testing.T) {
	for _, tc := range []struct {
		value, want, class string
	}{
		{"", "0 words · 0 / 10 characters", "text-neutral-600"},
		{"one", "1 word · 3 / 10 characters", "text-neutral-600"},
		{"nine char", "2 words · 9 / 10 characters", "text-warning-800"},
		{"eleven char", "2 words · 11 / 10 characters", "text-danger-600"},
		{"🙂🙂🙂🙂🙂", "1 word · 10 / 10 characters", "text-warning-800"}, // UTF-16 units, like String.length
	} {
		got, err := htemel.RenderString(ui.MarkdownEditor("b", ui.WithMaxLength(10), ui.WithValue(tc.value)))
		if err != nil {
			t.Fatal(err)
		}
		m := regexp.MustCompile(`<p id="b-count" class="tabular-nums ([^"]*)"[^>]*>([^<]*)</p>`).FindStringSubmatch(got)
		if m == nil {
			t.Fatalf("no counter in %s", got)
		}
		if m[2] != tc.want || !strings.Contains(m[1], tc.class) {
			t.Errorf("value %q: counter %q class %q, want %q class %s", tc.value, m[2], m[1], tc.want, tc.class)
		}
	}
	wantAll(t, render(t, ui.MarkdownEditor("b", ui.WithValue("a b c"))), `>3 words · 5 characters</p>`)
}

func TestMarkdownEditorUpload(t *testing.T) {
	wantAll(t, render(t, ui.MarkdownEditor("b", ui.WithUploadAction("upload(evt.detail.file)"))),
		`data-on:ui:md-upload="upload(evt.detail.file)"`, `<ui-mdeditor data-ui-upload>`)
	if strings.Contains(render(t, ui.MarkdownEditor("b")), "data-ui-upload") {
		t.Error("data-ui-upload without WithUploadAction")
	}
}

// The <ui-mdeditor> element and the rule that hides its toolbar until it is defined.
func TestMarkdownEditorAssets(t *testing.T) {
	wantAll(t, ui.ThemeCSS(), "ui-mdeditor {\n  display: contents;", "ui-mdeditor:not(:defined) .ui-mdeditor-toolbar {\n  display: none;")
	wantAll(t, render(t, ui.MarkdownEditor("b")), `<ui-mdeditor>`, `class="ui-mdeditor-toolbar`)
}

// mdeditorRows are the contrast pairs of the editor, on the white / white-5% editor surface
// over a card.
func mdeditorRows() []spec {
	var m []spec
	add := func(state, src string, k kind, light, dark side) {
		m = append(m, spec{component: "MarkdownEditor", state: state, src: src, kind: k, light: light, dark: dark})
	}
	const surfL, surfD = "bg-white", "dark:bg-white/5"
	add("text", "mdeditor.go:222,mdeditor.go:223", text, on("text-neutral-900", surfL).over(cardL), on("dark:text-white", surfD).over(cardD))
	add("placeholder", "mdeditor.go:223", text, on("placeholder:text-neutral-500", surfL).over(cardL), on("dark:placeholder:text-neutral-400", surfD).over(cardD))
	add("frame border vs card (SC 1.4.11)", "mdeditor.go:218", nonText, on("outline-field-border", "").over(cardL), on("outline-field-border", "").over(cardD))
	add("frame border vs page (SC 1.4.11)", "mdeditor.go:218", nonText, on("outline-field-border", ""), on("outline-field-border", ""))
	add("focus-within ring", "mdeditor.go:219,mdeditor.go:220", nonText, on("focus-within:outline-primary-600", "").over(cardL), on("dark:focus-within:outline-primary-500", "").over(cardD))
	add("toolbar icon", "button.go:20", nonText, on("text-neutral-700", surfL).over(cardL), on("dark:text-neutral-300", surfD).over(cardD))
	add("toolbar icon, hover", "mdeditor.go:224,button.go:20", nonText, on("hover:text-neutral-900", "hover:bg-neutral-100").over(cardL), on("dark:hover:text-white", "dark:hover:bg-white/10").over(cardD))
	add("hint link", "mdeditor.go:228,mdeditor.go:229", text, on("text-primary-600", surfL).over(cardL), on("dark:text-primary-400", surfD).over(cardD))
	add("hint link, hover", "mdeditor.go:228,mdeditor.go:229", text, on("hover:text-primary-700", surfL).over(cardL), on("dark:hover:text-primary-300", surfD).over(cardD))
	add("counter", "mdeditor.go:232", text, on("text-neutral-600", surfL).over(cardL), on("dark:text-neutral-400", surfD).over(cardD))
	add("counter near the limit", "mdeditor.go:233", text, on("text-warning-800", surfL).over(cardL), on("dark:text-warning-300", surfD).over(cardD))
	add("counter over the limit", "mdeditor.go:234", text, on("text-danger-600", surfL).over(cardL), on("dark:text-danger-400", surfD).over(cardD))
	add("empty preview text", "mdeditor.go:231", text, on("text-neutral-600", surfL).over(cardL), on("dark:text-neutral-400", surfD).over(cardD))
	add("counter on the page", "mdeditor.go:232", text, on("text-neutral-600", surfL), on("dark:text-neutral-400", surfD))
	add("counter over the limit on the page", "mdeditor.go:234", text, on("text-danger-600", surfL), on("dark:text-danger-400", surfD))
	return m
}
