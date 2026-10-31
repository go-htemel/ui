package ui

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"
)

// SelectElement is a native <select> drawn with the library's own chevron instead of the
// browser's arrow (which is heavy and unthemed, especially in dark mode). It renders a grid
// wrapper holding the select and the icon; the wrapper is presentational, so Attrs, SetAttr and
// SetAttrUnsafe all target the <select> itself, and Field labels the <select> through Control.
type SelectElement struct {
	*html.DivElement
	control *html.SelectElement
}

// Control returns the <select>, for Field (see CompositeControl) and for callers that need the
// element itself.
func (s *SelectElement) Control() FormControl { return s.control }

// Attrs applies attributes to the <select>, not the wrapper, so Datastar bindings land on the
// control: ui.Select("plan", ...).Attrs(datastar.Bind("plan")).
func (s *SelectElement) Attrs(attrs ...htemel.Attribute) *SelectElement {
	s.control.Attrs(attrs...)
	return s
}

// SetAttr sets an attribute on the <select>.
func (s *SelectElement) SetAttr(name, value string) { s.control.SetAttr(name, value) }

// SetAttrUnsafe sets an attribute on the <select> without escaping the value.
func (s *SelectElement) SetAttrUnsafe(name, value string) { s.control.SetAttrUnsafe(name, value) }

// Select returns a native <select> with the given name and options (see Choice), styled like
// Input and with the library chevron. Put it in a Field to label it.
func Select(name string, options ...html.SelectChild) *SelectElement {
	control := html.Select(options...).Name(name).
		Class(Classes(controlBase, "col-start-1 row-start-1 appearance-none py-1.5 pr-8 pl-3 dark:*:bg-neutral-800"))
	wrapper := html.Div(
		control,
		IconChevronDown().Class(Classes("pointer-events-none col-start-1 row-start-1 mr-2 size-5 self-center justify-self-end",
			"text-neutral-500 sm:size-4 dark:text-neutral-400")),
	).Class("grid grid-cols-1")
	return &SelectElement{DivElement: wrapper, control: control}
}

var _ CompositeControl = (*SelectElement)(nil)
