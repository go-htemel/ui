package ui

import (
	"io"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
	"github.com/go-htemel/htemel/svg"
)

// MarkdownTool is a button of the MarkdownEditor toolbar; see WithToolbar.
type MarkdownTool string

// Toolbar tools. The value is the button's data-ui-md-action, which the <ui-mdeditor> element of
// ui.js reads; MarkdownSeparator draws a divider between groups.
const (
	MarkdownBold          MarkdownTool = "bold"
	MarkdownItalic        MarkdownTool = "italic"
	MarkdownStrikethrough MarkdownTool = "strike"
	MarkdownHeading       MarkdownTool = "heading"
	MarkdownQuote         MarkdownTool = "quote"
	MarkdownCode          MarkdownTool = "code"
	MarkdownLink          MarkdownTool = "link"
	MarkdownImage         MarkdownTool = "image"
	MarkdownBulletList    MarkdownTool = "ul"
	MarkdownNumberedList  MarkdownTool = "ol"
	MarkdownTaskList      MarkdownTool = "task"
	MarkdownTable         MarkdownTool = "table"
	MarkdownSeparator     MarkdownTool = "|"
)

// DefaultMarkdownToolbar is the toolbar MarkdownEditor shows unless WithToolbar replaces it.
var DefaultMarkdownToolbar = []MarkdownTool{
	MarkdownBold, MarkdownItalic, MarkdownStrikethrough, MarkdownSeparator,
	MarkdownHeading, MarkdownQuote, MarkdownCode, MarkdownSeparator,
	MarkdownLink, MarkdownImage, MarkdownSeparator,
	MarkdownBulletList, MarkdownNumberedList, MarkdownTaskList, MarkdownTable,
}

type markdownToolInfo struct {
	label, keys, aria string // label; shortcut as shown in the tooltip; aria-keyshortcuts
	icon              func() *svg.SvgElement
}

var markdownTools = map[MarkdownTool]markdownToolInfo{
	MarkdownBold:          {"Bold", "Ctrl+B", "Control+B Meta+B", IconBold},
	MarkdownItalic:        {"Italic", "Ctrl+I", "Control+I Meta+I", IconItalic},
	MarkdownStrikethrough: {"Strikethrough", "Ctrl+Shift+X", "Control+Shift+X Meta+Shift+X", IconStrikethrough},
	MarkdownHeading:       {"Heading", "", "", IconHeading},
	MarkdownQuote:         {"Quote", "", "", IconQuote},
	MarkdownCode:          {"Code", "Ctrl+E", "Control+E Meta+E", IconCode},
	MarkdownLink:          {"Link", "Ctrl+K", "Control+K Meta+K", IconLink},
	MarkdownImage:         {"Image", "", "", IconImage},
	MarkdownBulletList:    {"Bulleted list", "", "", IconListBullet},
	MarkdownNumberedList:  {"Numbered list", "", "", IconListNumbered},
	MarkdownTaskList:      {"Task list", "", "", IconListTask},
	MarkdownTable:         {"Table", "", "", IconTable},
}

// MarkdownEditorOption configures MarkdownEditor.
type MarkdownEditorOption func(*mdConfig)

type mdConfig struct {
	id, signal, value, placeholder string
	rows, maxLength                int
	preview, previewLabel          string
	split, autosize, mono          bool
	tools                          []MarkdownTool
	upload                         string
	hint, hintHref                 string
}

// WithPreview adds a preview: action is a Datastar expression that asks the server to render
// the Markdown, typically datastar.Post("/preview"). The request carries the editor's signal
// (see WithSignal); the server answers with an SSE patch of the element whose id is
// MarkdownEditorElement.PreviewID(), most simply by sending MarkdownPreview(id,
// markdown.Render(src)). In tabs mode (the default) the action runs when the Preview tab is
// chosen; with WithSplit it runs on load and 400ms after the last keystroke.
func WithPreview(action string) MarkdownEditorOption {
	return func(c *mdConfig) { c.preview = action }
}

// WithPreviewLabel names the preview region (default "Preview"). Give each editor on a page its
// own, such as "Post preview", so that assistive technology can tell the regions apart.
func WithPreviewLabel(label string) MarkdownEditorOption {
	return func(c *mdConfig) { c.previewLabel = label }
}

// WithSplit shows the preview beside the textarea instead of behind a tab, once the editor is
// at least 36rem wide (an @xl container query); narrower, the preview stacks under the
// textarea. It needs WithPreview.
func WithSplit() MarkdownEditorOption { return func(c *mdConfig) { c.split = true } }

// WithSignal names the signal bound to the text (default: the name passed to MarkdownEditor).
// Use it when the form field name is not a valid signal name, such as "post[body]".
func WithSignal(name string) MarkdownEditorOption { return func(c *mdConfig) { c.signal = name } }

// WithEditorID sets the id the editor's parts are named after (default: the name with anything
// but letters and digits turned into hyphens): the textarea (until Field sets its own), the
// preview (<id>-preview), the tabs and the footer texts. Give two editors with the same name
// on one page distinct ids.
func WithEditorID(id string) MarkdownEditorOption { return func(c *mdConfig) { c.id = id } }

// WithValue sets the initial text, and the initial counts in the footer.
func WithValue(text string) MarkdownEditorOption { return func(c *mdConfig) { c.value = text } }

// WithRows sets the visible rows of the textarea (default 8).
func WithRows(n int) MarkdownEditorOption {
	return func(c *mdConfig) {
		if n > 0 {
			c.rows = n
		}
	}
}

// WithPlaceholder sets the textarea placeholder.
func WithPlaceholder(text string) MarkdownEditorOption {
	return func(c *mdConfig) { c.placeholder = text }
}

// WithMaxLength shows "used / max" characters in the footer: the counter turns to the warning
// color at 90% and to the danger color past the limit. It sets no maxlength attribute, so a
// paste is never cut silently; validate the length on the server.
func WithMaxLength(n int) MarkdownEditorOption {
	return func(c *mdConfig) {
		if n > 0 {
			c.maxLength = n
		}
	}
}

// WithAutosize grows the textarea with its content (CSS field-sizing: content, Chromium 123+;
// other browsers keep the rows and the resize handle), from about 10rem up to 70% of the
// viewport height.
func WithAutosize() MarkdownEditorOption { return func(c *mdConfig) { c.autosize = true } }

// WithMonospace sets the textarea in the monospace font (font-mono). The default is the sans
// font, which reads better for prose.
func WithMonospace() MarkdownEditorOption { return func(c *mdConfig) { c.mono = true } }

// WithToolbar replaces the toolbar buttons (DefaultMarkdownToolbar) with tools, in order;
// MarkdownSeparator draws a divider. No tools removes the toolbar.
func WithToolbar(tools ...MarkdownTool) MarkdownEditorOption {
	return func(c *mdConfig) { c.tools = append([]MarkdownTool{}, tools...) }
}

// WithUploadAction handles files pasted or dropped onto the textarea. The <ui-mdeditor> element
// stops the browser from inserting them and dispatches a ui:md-upload event per file, which runs
// action, a Datastar expression with evt.detail.file (the File) and evt.detail.insert(text),
// which inserts Markdown at the caret (or replaces the selection) as if typed. Uploading is the
// application's job, for example:
//
//	ui.WithUploadAction("fetch('/upload', {method: 'POST', body: evt.detail.file})" +
//		".then(r => r.text()).then(url => evt.detail.insert('![' + evt.detail.file.name + '](' + url + ')'))")
func WithUploadAction(action string) MarkdownEditorOption {
	return func(c *mdConfig) { c.upload = action }
}

// WithHint replaces the footer hint, by default "Markdown supported" linking to the CommonMark
// reference (https://commonmark.org/help/). The hint describes the textarea (aria-describedby).
// An empty text removes it; an empty href shows it without a link.
func WithHint(text, href string) MarkdownEditorOption {
	return func(c *mdConfig) { c.hint, c.hintHref = text, href }
}

// MarkdownEditorElement is the editor MarkdownEditor returns: a <div class="ui-mdeditor">
// framing a toolbar, the textarea, an optional preview and a footer. Like SelectElement, Attrs,
// SetAttr and SetAttrUnsafe target the textarea (so .Attrs(datastar.On(...)) lands on the
// control), while Id, Class and AddClass are the frame's. Field labels the textarea through
// Control.
type MarkdownEditorElement struct {
	*html.DivElement
	control   *mdTextarea
	previewID string
}

// mdTextarea is the editor's textarea as Field sees it: aria-describedby keeps the editor's own
// hint and counter, and a new id is mirrored to the toolbar's aria-controls.
type mdTextarea struct {
	*html.TextareaElement
	describedBy string
	toolbar     *html.DivElement
}

func (t *mdTextarea) SetAttr(name, value string) {
	switch name {
	case "aria-describedby":
		value = strings.TrimSpace(value + " " + t.describedBy)
	case "id":
		if t.toolbar != nil {
			t.toolbar.SetAttr("aria-controls", value)
		}
	}
	t.TextareaElement.SetAttr(name, value)
}

// Control returns the textarea, for Field (see CompositeControl).
func (e *MarkdownEditorElement) Control() FormControl { return e.control }

// PreviewID returns the id of the preview element, which the server patches; "" without
// WithPreview.
func (e *MarkdownEditorElement) PreviewID() string { return e.previewID }

// Attrs applies attributes to the textarea.
func (e *MarkdownEditorElement) Attrs(attrs ...htemel.Attribute) *MarkdownEditorElement {
	htemel.Apply(e.control, attrs...)
	return e
}

// SetAttr sets an attribute on the textarea.
func (e *MarkdownEditorElement) SetAttr(name, value string) { e.control.SetAttr(name, value) }

// SetAttrUnsafe sets an attribute on the textarea without escaping the value.
func (e *MarkdownEditorElement) SetAttrUnsafe(name, value string) {
	e.control.SetAttrUnsafe(name, value)
}

const (
	mdFrame = "ui-mdeditor @container min-w-0 rounded-control bg-white outline-1 -outline-offset-1 outline-field-border " +
		"not-focus-within:hover:outline-field-border-hover focus-within:outline-2 focus-within:-outline-offset-2 focus-within:outline-primary-600 " +
		"dark:bg-white/5 dark:focus-within:outline-primary-500"
	mdHeader   = "flex flex-wrap items-center gap-x-4 gap-y-1 border-b border-neutral-200 px-2 dark:border-white/10"
	mdTextArea = "ui-control block w-full min-w-0 resize-y border-0 bg-transparent px-3 py-2 text-base text-neutral-900 " +
		"placeholder:text-neutral-500 focus:outline-none sm:text-sm/6 dark:text-white dark:placeholder:text-neutral-400"
	mdToolbar  = "ui-mdeditor-toolbar flex flex-wrap items-center gap-0.5 py-1"
	mdTool     = "hover:text-neutral-900 dark:hover:text-white"
	mdSep      = "mx-1 h-5 w-px shrink-0 bg-neutral-200 dark:bg-white/10"
	mdFooter   = "flex flex-wrap items-center justify-between gap-x-4 gap-y-1 border-t border-neutral-200 px-3 py-1.5 text-xs text-neutral-600 dark:border-white/10 dark:text-neutral-400"
	mdHintLink = "font-medium text-primary-600 underline underline-offset-2 hover:text-primary-700 focus-visible:rounded-xs focus-visible:outline-2 " +
		"focus-visible:outline-offset-2 focus-visible:outline-primary-600 dark:text-primary-400 dark:hover:text-primary-300 dark:focus-visible:outline-primary-500"
	mdPreview      = "min-h-40 min-w-0 px-4 py-3"
	mdPreviewEmpty = "text-sm text-neutral-600 italic dark:text-neutral-400"
	mdCountNormal  = "text-neutral-600 dark:text-neutral-400"
	mdCountNear    = "text-warning-800 dark:text-warning-300"
	mdCountOver    = "font-semibold text-danger-600 dark:text-danger-400"
)

// mdDebounce is how long the split preview waits after the last keystroke.
const mdDebounce = 400 * time.Millisecond

// MarkdownEditor returns a Markdown editor for the form field name: a toolbar, a textarea bound
// to a signal (the name, or WithSignal), an optional server-rendered preview (WithPreview) and a
// footer with a "Markdown supported" hint and a word and character count. It is a
// CompositeControl, so Field labels the textarea:
//
//	ed := ui.MarkdownEditor("body", ui.WithPreview(datastar.Post("/preview")), ui.WithValue(post.Body))
//	ui.Field("body", "Entry", ed)
//
// and the preview handler renders the signal with the markdown package:
//
//	var in struct{ Body string `json:"body"` }
//	_ = datastar.ReadSignals(r, &in)
//	sse.PatchElements(render(ui.MarkdownPreview(ed.PreviewID(), markdown.Render(in.Body))))
//
// The root package does not render Markdown (it has no parser dependency): the preview always
// comes from the server, as an SSE patch.
//
// # Toolbar and keyboard
//
// The toolbar buttons only carry a data-ui-md-action attribute: the <ui-mdeditor> custom element
// of ui.js (ScriptTag) does the editing. A button wraps the selection in Markdown (bold, italic,
// strikethrough, inline code or a fenced block for a multi-line selection, a link or image with
// its URL selected) or toggles a line prefix on every selected line (heading, quote, bulleted,
// numbered and task lists), clicking again undoes it, and the edit goes through the browser's
// insertText command, so Ctrl+Z undoes it too and an input event updates the bound signal.
// Ctrl/Cmd+B, I, E, K and Shift+X are shortcuts for bold, italic, code, link and
// strikethrough, and Enter in a list item continues the list (Enter on an empty item ends it).
// Tab is left alone, so keyboard users can always leave the textarea. Without ui.js the toolbar
// stays hidden and the textarea, the counts and the preview still work.
//
// # Accessibility
//
// The toolbar is role="toolbar" (aria-label "Formatting", aria-controls the textarea) with a
// roving tabindex: Tab enters it once, and ArrowLeft / ArrowRight, Home and End move between
// buttons. Each button has an aria-label and a tooltip naming its shortcut (and
// aria-keyshortcuts). The hint and the counter describe the textarea, alongside Field's help and
// error text. The Write / Preview tabs follow Tabs (role="tablist", roving focus, arrow keys);
// the preview is a role="region" labelled "Preview" (WithPreviewLabel) and a polite live region. The whole frame
// shows the focus ring while the textarea or a button has focus, and the preview tab panel fades
// in like Tabs (no motion under prefers-reduced-motion).
func MarkdownEditor(name string, opts ...MarkdownEditorOption) *MarkdownEditorElement {
	c := mdConfig{rows: 8, tools: DefaultMarkdownToolbar, hint: "Markdown supported", hintHref: "https://commonmark.org/help/", previewLabel: "Preview"}
	for _, o := range opts {
		if o != nil {
			o(&c)
		}
	}
	if c.signal == "" {
		c.signal = name
	}
	if c.id == "" {
		c.id = domID(name)
	}
	if c.preview == "" {
		c.split = false
	}
	id := c.id
	sig := "$" + c.signal

	ta := html.Textarea().Name(name).Rows(c.rows).Id(id).
		Class(Classes(mdTextArea, cond(c.autosize, "field-sizing-content min-h-40 max-h-[70vh]"), cond(c.mono, "font-mono"))).
		Attrs(datastar.Bind(c.signal)).
		Text(c.value)
	if c.placeholder != "" {
		ta.Placeholder(c.placeholder)
	}
	if c.split {
		ta.Attrs(datastar.OnInput(c.preview, datastar.Debounce(mdDebounce)))
	}
	control := &mdTextarea{TextareaElement: ta}

	var described []string
	footer := html.Div().Class(Classes(mdFooter))
	if c.hint != "" {
		described = append(described, id+"-hint")
		hint := html.P().Id(id + "-hint")
		if c.hintHref != "" {
			hint.Children(html.A().Href(c.hintHref).Class(Classes(mdHintLink)).Text(c.hint))
		} else {
			hint.Text(c.hint)
		}
		footer.Children(hint)
	}
	described = append(described, id+"-count")
	footer.Children(counter(id+"-count", sig, c.value, c.maxLength))
	control.describedBy = strings.Join(described, " ")
	control.SetAttr("aria-describedby", "")

	header := html.Div().Class(Classes(mdHeader))
	var toolbar *html.DivElement
	if len(c.tools) > 0 {
		toolbar = markdownToolbar(c.tools).Aria("controls", id)
		control.toolbar = toolbar
	}

	var host []htemel.FlowContent
	root := html.Div().Class(Classes(mdFrame))
	if c.upload != "" {
		root.Attrs(datastar.On("ui:md-upload", c.upload))
	}

	var body htemel.FlowContent
	switch {
	case c.preview == "":
		body = html.Div(ta)
	case c.split:
		preview := previewRegion(id, c.previewLabel).Attrs(datastar.Init(c.preview))
		preview.AddClass(Classes("border-t border-neutral-200 @xl:border-t-0 @xl:border-l dark:border-white/10"))
		body = html.Div(html.Div(ta).Class("min-w-0"), preview).Class(Classes("grid grid-cols-1 @xl:grid-cols-2"))
	default:
		mode := "_" + signalID(id) + "Mode"
		modeSig := "$" + mode
		root.Attrs(initially(mode, "'write'"))
		list := html.Div().Role("tablist").Aria("label", "Editor mode").
			Attrs(datastar.OnKeyDown(tabKeys)).Class(Classes("-mb-px flex gap-x-4"))
		header.Children(list)
		active, inactive := Classes(tabActive), Classes(tabInactive)
		var panels []htemel.FlowContent
		for i, t := range []struct{ key, label, click string }{
			{"write", "Write", modeSig + " = 'write'"},
			{"preview", "Preview", modeSig + " = 'preview'; " + c.preview},
		} {
			tid := id + "-" + t.key
			selected := modeSig + " === '" + t.key + "'"
			state := map[bool]string{true: active, false: inactive}[i == 0]
			list.Children(html.Button().Type(html.ButtonTypeButton).Role("tab").Id(tid+"-tab").
				Aria("controls", tid+"-panel").
				Aria("selected", strconv.FormatBool(i == 0)).
				Tabindex(map[bool]int{true: 0, false: -1}[i == 0]).
				Class(Classes(tabBase), state).
				Attrs(
					datastar.Classes("{"+datastar.Quote(active)+": "+selected+", "+datastar.Quote(inactive)+": !("+selected+")}"),
					datastar.Attr("aria-selected", "("+selected+") ? 'true' : 'false'"),
					datastar.Attr("tabindex", "("+selected+") ? '0' : '-1'"),
					datastar.OnClick(t.click),
				).Text(t.label))
			panel := html.Div().Role("tabpanel").Id(tid+"-panel").Aria("labelledby", tid+"-tab").
				Class(Classes("transition-opacity duration-200 ease-out starting:opacity-0 motion-reduce:transition-none")).
				Attrs(datastar.Show(selected))
			if i == 0 {
				panel.Children(ta)
			} else {
				panel.Tabindex(0).Style(hiddenUntilDatastar).
					AddClass(Classes("focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-primary-600 dark:focus-visible:outline-primary-500")).
					Children(previewRegion(id, c.previewLabel))
			}
			panels = append(panels, panel)
		}
		if toolbar != nil {
			toolbar.Attrs(datastar.Show(modeSig + " === 'write'"))
		}
		body = html.Div(panels...)
	}
	if toolbar != nil {
		toolbar.AddClass("ml-auto")
		header.Children(toolbar)
	}
	if toolbar != nil || c.preview != "" && !c.split {
		host = append(host, header)
	}
	host = append(host, body, footer)
	root.Children(mdHost{upload: c.upload != "", children: host})

	e := &MarkdownEditorElement{DivElement: root, control: control}
	if c.preview != "" {
		e.previewID = MarkdownPreviewID(id)
	}
	return e
}

// MarkdownPreviewID returns the id of the preview of the MarkdownEditor whose id (WithEditorID,
// by default the name with anything but letters and digits turned into hyphens) is editorID:
// editorID + "-preview". A preview handler can use it without building the editor.
func MarkdownPreviewID(editorID string) string { return domID(editorID) + "-preview" }

// MarkdownPreview returns the patchable part of a MarkdownEditor preview: a <div> with the given
// id (MarkdownEditorElement.PreviewID, MarkdownPreviewID) holding content, usually
// markdown.Render(src). The editor renders it empty, showing "Nothing to preview yet."; the
// server answers a preview request by patching it:
//
//	sse.PatchElements(render(ui.MarkdownPreview(ed.PreviewID(), markdown.Render(src))))
//
// The editor wraps it in the preview region (role="region", aria-label from WithPreviewLabel,
// aria-live="polite"), which the patch never replaces, so the live region and its label stay put.
func MarkdownPreview(id string, content ...htemel.FlowContent) *html.DivElement {
	el := html.Div(content...).Id(id).Class("min-w-0")
	if len(content) == 0 {
		el.Children(html.P().Class(Classes(mdPreviewEmpty)).Text("Nothing to preview yet."))
	}
	return el
}

// previewRegion is the live region around the patchable MarkdownPreview of editor id.
func previewRegion(id, label string) *html.DivElement {
	return html.Div(MarkdownPreview(MarkdownPreviewID(id))).Role("region").Aria("label", label).Aria("live", "polite").
		Class(Classes(mdPreview))
}

// markdownToolbar renders the tool buttons: the first is in the tab order, the others are
// reached with the arrow keys (the roving tabindex of <ui-mdeditor>).
func markdownToolbar(tools []MarkdownTool) *html.DivElement {
	bar := html.Div().Role("toolbar").Aria("label", "Formatting").Class(Classes(mdToolbar))
	first := true
	for _, t := range tools {
		if t == MarkdownSeparator {
			bar.Children(html.Div().Role("separator").Aria("orientation", "vertical").Class(Classes(mdSep)))
			continue
		}
		info, ok := markdownTools[t]
		if !ok {
			continue
		}
		title := info.label
		if info.keys != "" {
			title += " (" + info.keys + ")"
		}
		b := IconButton(Ghost, info.icon().Class("size-4 shrink-0"), info.label, WithSize(Sm), withClass(mdTool)).
			Title(title).Tabindex(map[bool]int{true: 0, false: -1}[first]).Data("ui-md-action", string(t))
		if info.aria != "" {
			b.Aria("keyshortcuts", info.aria)
		}
		bar.Children(b)
		first = false
	}
	return bar
}

// counter is the footer's word and character count, server-rendered for the initial text and
// then kept current by Datastar. With max, it shows "used / max" and turns to the warning color
// at 90% and the danger color past max.
func counter(id, sig, value string, maxLen int) *html.PElement {
	chars := len(utf16.Encode([]rune(value))) // what String.length counts in the browser
	words := len(strings.Fields(value))
	v := "(" + sig + " ?? '')"
	wordsExpr := "((" + v + ".match(/\\S+/g) || []).length)"
	text := plural(words, "word") + " · " + strconv.Itoa(chars)
	expr := "(n => n + (n === 1 ? ' word' : ' words'))" + wordsExpr + " + ' · ' + " + v + ".length"
	p := html.P().Id(id).Class(Classes("tabular-nums"))
	if maxLen > 0 {
		m := strconv.Itoa(maxLen)
		text += " / " + m + " characters"
		expr += " + ' / " + m + " characters'"
		nearAt := strconv.Itoa((maxLen*9 + 9) / 10)
		near := v + ".length >= " + nearAt + " && " + v + ".length <= " + m
		over := v + ".length > " + m
		state := Classes(mdCountNormal)
		switch {
		case chars > maxLen:
			state = Classes(mdCountOver)
		case chars*10 >= maxLen*9:
			state = Classes(mdCountNear)
		}
		p.AddClass(state)
		p.Attrs(datastar.Classes("{" + datastar.Quote(Classes(mdCountNormal)) + ": !(" + near + ") && !(" + over + "), " +
			datastar.Quote(Classes(mdCountNear)) + ": " + near + ", " + datastar.Quote(Classes(mdCountOver)) + ": " + over + "}"))
	} else {
		text += " characters"
		expr += " + ' characters'"
	}
	return p.Attrs(datastar.Text(expr)).Text(text)
}

func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	return strconv.Itoa(n) + " " + word + "s"
}

// signalID turns an id into a camelCase signal name fragment: "post-body" -> "postBody".
func signalID(s string) string {
	var b strings.Builder
	up := false
	for _, r := range s {
		switch {
		case 'a' <= r && r <= 'z', 'A' <= r && r <= 'Z', '0' <= r && r <= '9' && b.Len() > 0:
			if up && b.Len() > 0 {
				r = []rune(strings.ToUpper(string(r)))[0]
			}
			b.WriteRune(r)
			up = false
		default:
			up = true
		}
	}
	if b.Len() == 0 {
		return "md"
	}
	return b.String()
}

func cond(ok bool, class string) string {
	if ok {
		return class
	}
	return ""
}

// mdHost is the <ui-mdeditor> custom element inside the frame. It generates no box
// (ui-mdeditor { display: contents } in ui.css). data-ui-upload asks it to intercept pasted and
// dropped files (WithUploadAction).
type mdHost struct {
	htemel.PhrasingMarker
	htemel.PalpableMarker
	upload   bool
	children []htemel.FlowContent
}

// Render implements htemel.Node.
func (h mdHost) Render(w io.Writer) error {
	open := "<ui-mdeditor>"
	if h.upload {
		open = "<ui-mdeditor data-ui-upload>"
	}
	if _, err := io.WriteString(w, open); err != nil {
		return err
	}
	for _, c := range h.children {
		if err := c.Render(w); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, "</ui-mdeditor>")
	return err
}

var _ CompositeControl = (*MarkdownEditorElement)(nil)
