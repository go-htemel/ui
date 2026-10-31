package ui

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// FileUploadOption configures FileUpload.
type FileUploadOption func(*fileConfig)

type fileConfig struct {
	multiple       bool
	accept, hint   string
	browse, action string
}

// FileMultiple lets the reader choose several files; dropped files are added to the selection.
func FileMultiple() FileUploadOption { return func(c *fileConfig) { c.multiple = true } }

// FileAccept restricts the choice, as the accept attribute does: "image/*", ".pdf,.md". Dropped
// files that do not match are ignored.
func FileAccept(types string) FileUploadOption { return func(c *fileConfig) { c.accept = types } }

// FileHint adds a line under the drop area, such as "PNG or JPG, up to 10 MB"; it describes the
// file input.
func FileHint(text string) FileUploadOption { return func(c *fileConfig) { c.hint = text } }

// FileBrowseLabel replaces the text of the browse button ("Choose a file", or "Choose files"
// with FileMultiple).
func FileBrowseLabel(text string) FileUploadOption { return func(c *fileConfig) { c.browse = text } }

// FileUploadAction runs action, a Datastar expression, once for every newly chosen or dropped
// file: <ui-fileupload> dispatches a ui:upload event with evt.detail.file (the File) and
// evt.detail.files (the whole selection). Uploading, and its progress, are the application's:
// keep the progress in a signal and show it with ProgressSignal, for example
//
//	ui.FileUploadAction("const x = new XMLHttpRequest(); x.open('POST', '/upload'); " +
//		"x.upload.onprogress = (e) => { $uploadProgress = Math.round(100 * e.loaded / e.total) }; " +
//		"x.send(evt.detail.file)")
//	ui.ProgressSignal("uploadProgress", ...)
//
// The files also stay in the file input, so a form post still sends them.
func FileUploadAction(action string) FileUploadOption {
	return func(c *fileConfig) { c.action = action }
}

// FileUploadElement is the control FileUpload returns: a <div> holding the <ui-fileupload>
// element. Field labels its file input through Control.
type FileUploadElement struct {
	*html.DivElement
	control *describedInput
}

// Control returns the file input, for Field (see CompositeControl).
func (e *FileUploadElement) Control() FormControl { return e.control }

const (
	fileZone = "flex flex-col items-center gap-2 rounded-surface border border-dashed border-neutral-300 bg-white px-6 py-8 text-center " +
		"transition-colors duration-150 dark:border-white/20 dark:bg-white/5 " +
		"data-ui-over:border-primary-600 data-ui-over:bg-primary-500/10 data-ui-over:outline-2 data-ui-over:outline-dashed data-ui-over:outline-primary-600 dark:data-ui-over:outline-primary-500"
	fileBrowse = "relative cursor-pointer rounded-control font-semibold text-primary-600 hover:text-primary-700 " +
		"has-focus-visible:outline-2 has-focus-visible:outline-offset-2 has-focus-visible:outline-primary-600 " +
		"dark:text-primary-400 dark:hover:text-primary-300 dark:has-focus-visible:outline-primary-500"
	fileText    = "text-sm text-neutral-600 dark:text-neutral-400"
	fileHint    = "text-xs text-neutral-600 dark:text-neutral-400"
	fileRow     = "flex items-center gap-3 py-2"
	fileName    = "truncate text-sm font-medium text-neutral-900 dark:text-white"
	fileSize    = "text-xs text-neutral-600 dark:text-neutral-400"
	filePreview = "size-10 shrink-0 rounded-control object-cover ring-1 ring-neutral-950/10 dark:ring-white/10"
	fileIcon    = "flex size-10 shrink-0 items-center justify-center rounded-control bg-neutral-100 text-neutral-500 dark:bg-white/10 dark:text-neutral-400"
)

// FileUpload returns a file picker for the form field name: a drop area with a browse button and
// a list of the chosen files (name, size, an image preview, a remove button). It is a
// CompositeControl, so Field labels the file input:
//
//	ui.Field("cover", "Cover image", ui.FileUpload("cover", ui.FileAccept("image/*"), ui.FileHint("PNG or JPG, up to 5 MB")))
//
// The browse button is a <label> around a visually hidden <input type="file">, so choosing a
// file needs no script, the input is in the Tab order (the label shows its focus ring) and a
// form post sends the files. The <ui-fileupload> element of ui.js (ScriptTag) adds the rest: it
// accepts files dropped on the area (styled like DropZone with data-ui-over while a drag of files
// is above it), lists the selection with sizes and previews of images (URL.createObjectURL,
// revoked when the list changes), removes a file from the selection (the input's FileList is
// rebuilt with a DataTransfer) and runs FileUploadAction. The list is a polite live region that
// announces "2 files selected" and removals; after a removal focus returns to the input.
func FileUpload(name string, opts ...FileUploadOption) *FileUploadElement {
	var c fileConfig
	for _, o := range opts {
		if o != nil {
			o(&c)
		}
	}
	if c.browse == "" {
		c.browse = "Choose a file"
		if c.multiple {
			c.browse = "Choose files"
		}
	}
	id := domID(name)
	in := html.Input().Type(html.InputTypeFile).Name(name).Class("sr-only")
	if c.multiple {
		in.Multiple(true)
	}
	if c.accept != "" {
		in.Accept(c.accept)
	}
	control := &describedInput{InputElement: in}
	drag := " or drag it here"
	if c.multiple {
		drag = " or drag them here"
	}
	zone := html.Div(
		IconUpload().SetClass(Classes("size-8 text-neutral-400 dark:text-neutral-500")),
		html.P(
			html.Label(html.Span().Text(c.browse), in).Class(Classes(fileBrowse)),
			htemel.Text(drag),
		).Class(Classes(fileText)),
	).Class(Classes(fileZone)).Attrs(htemel.Attr("data-ui-dropzone", ""))
	if c.hint != "" {
		control.describedBy = id + "-file-hint"
		zone.Children(html.P().Id(id + "-file-hint").Class(Classes(fileHint)).Text(c.hint))
	}
	control.SetAttr("aria-describedby", "")
	row := html.Li(
		html.Img().Alt("").Class(Classes(filePreview)).Attrs(htemel.Attr("data-ui-file-preview", ""), htemel.Attr("hidden", "")),
		html.Span(IconFile()).Class(Classes(fileIcon)).Attrs(htemel.Attr("data-ui-file-icon", "")),
		html.Div(
			html.P().Class(Classes(fileName)).Attrs(htemel.Attr("data-ui-file-name", "")),
			html.P().Class(Classes(fileSize)).Attrs(htemel.Attr("data-ui-file-size", "")),
		).Class("min-w-0 flex-1"),
		IconButton(Ghost, IconX().SetClass("size-4 shrink-0"), "Remove", WithSize(Xs)).Attrs(htemel.Attr("data-ui-file-remove", "")),
	).Class(Classes(fileRow))
	attrs := map[string]any{"class": "block"}
	if c.action != "" {
		attrs["data-ui-upload"] = true
	}
	host := htemel.Generic("ui-fileupload", attrs,
		zone,
		html.Ul().Role("list").Aria("label", "Selected files").
			Class(Classes("mt-3 divide-y divide-neutral-200 empty:hidden dark:divide-white/10")).Attrs(htemel.Attr("data-ui-files", "")),
		html.Template(row).Attrs(htemel.Attr("data-ui-file-template", "")),
		html.Span().Class("sr-only").Aria("live", "polite").Attrs(htemel.Attr("data-ui-file-status", "")),
	)
	root := html.Div(host)
	if c.action != "" {
		root.Attrs(datastar.On("ui:upload", c.action))
	}
	return &FileUploadElement{DivElement: root, control: control}
}

var _ CompositeControl = (*FileUploadElement)(nil)
