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
// Brand, Info, Success, Warning, Danger) are typed; Brand is the primary color's tints. Danger is untyped so it serves as both a Variant and
// a Tone. Options such as WithSize and WithIcon configure the components that read them.
//
// # Interactivity
//
// Open/close state lives in Datastar signals named by the caller: Dialog, Drawer, Dropdown,
// Disclosure, Tabs and Accordion take the signal name as their first argument and declare it with
// __ifmissing so the server can preset it. Tabs are driven by data-show, data-class and data-attr;
// Disclosure and Accordion bind aria-expanded and their panels follow it in CSS. Dialog, Drawer and Dropdown are native overlays, modal <dialog>s and a
// popover: the browser shows and hides them, and the signal mirrors their state both ways (an
// effect calls showModal, close, showPopover or hidePopover, and the close and toggle events
// write the signal back). Open, Close and Toggle return the matching expressions, and Opens is
// the click handler for a trigger outside the component. Toasts are appended to ToastRegion by the server (for example
// with datastar-go) and hide themselves after a timeout, which pauses while the pointer or focus
// is on the toast.
//
// # Long-form content
//
// Prose (ProseSm, ProseLg) styles HTML that has no classes of its own, such as rendered Markdown,
// through .ui-prose descendant rules in the design tokens. MarkdownEditor is a textarea with a
// formatting toolbar, Write / Preview tabs or a split view and a word count; its preview is
// rendered by the application (the markdown subpackage, github.com/go-htemel/ui/markdown, turns
// Markdown into sanitised HTML in Prose) and patched by id, so this package has no Markdown
// dependency.
//
// # Layouts
//
// AppShell (sidebar, top bar and <main>, with a SkipLink; SidebarFixed, SidebarCollapsible,
// SidebarDark, TopbarOnly), SplitLayout (master/detail, SplitSelect, SplitDetailHeader),
// SettingsLayout, StackedLayout with StackedTabs, PageHeader and SectionHeader frame admin and
// content pages. Each layout's root is a CSS container and its breakpoints are container queries
// (@3xl:, @2xl:, @xl:), so a layout switches on its own width, not the viewport's: it takes its
// narrow form in a narrow column or preview frame, and layouts nest. NavLink items take an Icon,
// a Badge and an Ancestor state.
//
// # Content inputs
//
// TagInput (chips that submit as repeated form values), FileUpload (a label-wrapped file input
// with a drop area and a file list), DateInput, DateTimeInput and TimeInput, SlugField (follows a
// title signal until edited) and FilterBar with FilterSearch and FilterReset are form controls
// that Field labels.
//
// # Assets
//
// One optional embedded script, served by Assets and loaded with ScriptTag, defines four custom
// elements: <ui-sortable>, which Sortable renders, adds touch dragging to Draggable and DropZone;
// <ui-mdeditor>, inside every MarkdownEditor, runs the toolbar buttons and the editing
// shortcuts; <ui-taginput> and <ui-fileupload> edit the tags of TagInput and the selection of
// FileUpload. Every component works without it (the PickUp and MoveHere buttons are the touch
// fallback; the editor hides its toolbar and keeps the textarea; TagInput submits its
// server-rendered tags and FileUpload is a plain file input), and it needs no inline script, so
// a strict Content-Security-Policy allows it.
//
// # Light and dark mode
//
// The dark: variants follow a .dark class on <html> (see ThemeCSS). ColorSchemeBootstrap, an inline
// script for <head> before the stylesheet, applies the visitor's stored choice before the first
// paint, and ColorSchemeToggle is the System / Light / Dark control that changes it. Apps that
// want prefers-color-scheme only drop the @custom-variant line from the tokens.
//
// # Accessibility
//
// Dialog, Dropdown, Tabs, Disclosure, Accordion, Draggable and Toast carry their ARIA roles and
// states and handle the keys of the matching WAI-ARIA pattern. For Dialog and Dropdown the
// browser does most of it (inertness, Escape, light dismiss, focus return, the top layer); the
// others use Datastar expressions. Each component's documentation says what is and is not
// covered (for example, a modal Dialog does not lock page scrolling).
//
// # Styling
//
// Classes are Tailwind v4 utilities with dark: variants, written against a small set of design
// tokens instead of fixed palette colors: the color families primary (brand and interactive),
// neutral (surfaces, text, borders), success, warning, danger and info (bg-primary-600,
// text-neutral-500, ring-danger-600/20), the radii rounded-control and rounded-surface, and
// Tailwind's own font-sans (FontCSS, opt-in, sets it to the self-hosted Geist). ThemeCSS (also
// cmd/themecss, and /ui.css from Assets) holds the default tokens as a Tailwind @theme block: an indigo primary,
// zinc, green, amber, red and sky aliases, and md corners, plus the rules for native control
// parts (autofill, spin buttons, the customizable select picker) that Input, Textarea and
// Select opt into with the ui-control and ui-select classes. Applications retheme by overriding those CSS variables, globally in their own
// @theme or per scope under a selector such as [data-theme="indigo"]; no Go code changes and
// nothing is rewritten at render time. Classes joins and normalizes class lists for your own
// components. Tailwind only generates classes it finds in your sources, so add the safelist
// (tailwind.safelist.txt, or the output of cmd/safelist) to your Tailwind build; see the README.
package ui
