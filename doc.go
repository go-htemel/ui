// Package ui is a component library for htemel (github.com/go-htemel/htemel): Go functions that
// render Tailwind-styled HTML, with Datastar (https://data-star.dev) attributes for every
// interaction and no custom JavaScript.
//
// # Components return elements
//
// A component is a plain function that returns the concrete htemel element, so the caller keeps
// chaining element methods and Datastar attributes onto it:
//
//	ui.Button(ui.Primary, "Save", ui.WithLoading("saving")).
//		Attrs(datastar.OnClick(datastar.Post("/save")))
//
// Components are typed by the HTML content model like the elements they return: Button and Badge
// are phrasing content (they fit in a paragraph), Card and Alert are flow content, and the
// compiler rejects a Card inside an html.P. Parts that only make sense inside one component are
// their own types and are not content at all: CardHeader, CardBody and CardFooter return a
// CardSlot that only Card accepts, Tab returns a TabItem for Tabs, and AccordionItem returns an
// AccordionPanel for Accordion.
//
// # Variants, sizes and tones
//
// Variant (Primary, Secondary, Outline, Ghost, Danger, Link), Size (Xs to Xl) and Tone (Neutral,
// Info, Success, Warning, Danger) are typed. Danger is untyped so it serves as both a Variant and
// a Tone. Options such as WithSize and WithIcon configure the components that read them.
//
// # Interactivity
//
// Open/close state lives in Datastar signals named by the caller: Dialog, Dropdown, Disclosure,
// Tabs and Accordion take the signal name as their first argument, declare it with
// __ifmissing so the server can preset it, and are driven by data-show, data-class and data-attr.
// Open, Close and Toggle return the matching expressions, and Opens is the click handler for a
// trigger outside the component. Toasts are appended to ToastRegion by the server (for example
// with datastar-go) and hide themselves after a timeout, which pauses while the pointer or focus
// is on the toast.
//
// # Accessibility
//
// Dialog, Dropdown, Tabs, Disclosure, Accordion, Draggable and Toast carry their ARIA roles and
// states and handle the keys of the matching WAI-ARIA pattern with Datastar expressions only;
// each component's documentation says what that does and does not cover (for example, a Dialog
// traps Tab but cannot make the rest of the page inert).
//
// # Styling
//
// Classes are Tailwind v4 utilities on the zinc palette with dark: variants, written for
// DefaultTheme. SetTheme changes the primary color family, the corner radius and the body font;
// Classes applies the current theme to your own class strings. Tailwind only generates classes it
// finds in your sources, so add the safelist (tailwind.safelist.txt, or the output of
// cmd/safelist for a custom theme) to your Tailwind build; see the README.
package ui
