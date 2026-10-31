package ui

import "github.com/go-htemel/htemel"

// Option configures a component. Each component documents the options it reads and ignores the
// rest.
type Option func(*options)

type options struct {
	size     Size
	icon     htemel.PhrasingContent
	trailing htemel.PhrasingContent
	loading  string
	rail     bool
}

func collect(opts []Option) options {
	var o options
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	return o
}

// WithSize sets the component size (default Md).
func WithSize(s Size) Option { return func(o *options) { o.size = s } }

// WithIcon puts an icon before a button's label.
func WithIcon(icon htemel.PhrasingContent) Option { return func(o *options) { o.icon = icon } }

// WithTrailingIcon puts an icon after a button's label.
func WithTrailingIcon(icon htemel.PhrasingContent) Option {
	return func(o *options) { o.trailing = icon }
}

// WithLoading wires a button to a Datastar indicator signal: while a request started by the
// button is in flight, the signal is true, the button is disabled and a spinner replaces its
// icon. signal is a camelCase signal name, such as "saving".
func WithLoading(signal string) Option { return func(o *options) { o.loading = signal } }

// WithRailIcon makes a Button or LinkButton collapse to its leading icon in the rail of a
// collapsed AppShell (SidebarCollapsible): the label and the trailing icon become visually hidden
// (the label stays the accessible name), the content centres and the horizontal padding shrinks
// so the button fits the 4rem rail. Outside a collapsed rail it changes nothing. Give the button
// an icon (WithIcon), or nothing is left to show. Use it for the controls of WithSidebarFooter:
//
//	ui.Button(ui.Ghost, "Ada Lovelace", ui.WithIcon(avatar), ui.WithRailIcon())
func WithRailIcon() Option { return func(o *options) { o.rail = true } }
