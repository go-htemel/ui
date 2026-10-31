package ui

import (
	"strings"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// PageHeaderOption configures PageHeader.
type PageHeaderOption func(*pageHeaderConfig)

type pageHeaderConfig struct {
	crumbs            htemel.FlowContent
	description       string
	actions, meta     []htemel.FlowContent
	backHref, backTxt string
	level             int
}

// WithBreadcrumbs puts a trail above the title, usually Breadcrumbs(...).
func WithBreadcrumbs(node htemel.FlowContent) PageHeaderOption {
	return func(c *pageHeaderConfig) { c.crumbs = node }
}

// WithDescription adds a line of text under the title.
func WithDescription(text string) PageHeaderOption {
	return func(c *pageHeaderConfig) { c.description = text }
}

// WithActions puts buttons beside the title (below it in a narrow container), primary last.
func WithActions(nodes ...htemel.FlowContent) PageHeaderOption {
	return func(c *pageHeaderConfig) { c.actions = append(c.actions, nodes...) }
}

// WithMeta adds a row of small facts under the title: badges, dates, an author.
func WithMeta(nodes ...htemel.FlowContent) PageHeaderOption {
	return func(c *pageHeaderConfig) { c.meta = append(c.meta, nodes...) }
}

// WithBackLink puts a "← label" link to href above the title, for a detail page's parent.
func WithBackLink(href, label string) PageHeaderOption {
	return func(c *pageHeaderConfig) { c.backHref, c.backTxt = href, label }
}

// WithHeadingLevel sets the title's heading level, 1 (the default, the page's <h1>) to 6: use 2
// for a header inside a page that already has its <h1>.
func WithHeadingLevel(level int) PageHeaderOption {
	return func(c *pageHeaderConfig) { c.level = level }
}

const (
	pageBackClass = "inline-flex items-center gap-x-1.5 rounded-control text-sm font-medium text-neutral-600 transition-colors duration-150 " +
		"hover:text-neutral-900 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600 " +
		"dark:text-neutral-400 dark:hover:text-white dark:focus-visible:outline-primary-500"
	pageTitleClass = "text-2xl/8 font-semibold tracking-tight text-neutral-900 @xl:truncate dark:text-white"
	pageDescClass  = "mt-1 text-sm text-neutral-600 dark:text-neutral-400"
	pageMetaClass  = "mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-neutral-600 dark:text-neutral-400"
)

// PageHeader returns the header of a page: optional back link and breadcrumbs, the title, a
// description, a meta row and actions. It is a <header> (inside <main> it is not a banner
// landmark) and a CSS container: from the @xl container width (36rem) the actions sit to the
// right of the title, which then truncates on one line; below it they wrap under the title.
//
//	ui.PageHeader("Edit post",
//		ui.WithBackLink("/posts", "Posts"),
//		ui.WithMeta(ui.Badge(ui.Warning, "Draft"), htemel.Text("Saved 2 minutes ago")),
//		ui.WithActions(ui.Button(ui.Secondary, "Preview"), ui.Button(ui.Primary, "Publish")))
func PageHeader(title string, opts ...PageHeaderOption) *html.HeaderElement {
	c := pageHeaderConfig{level: 1}
	for _, o := range opts {
		if o != nil {
			o(&c)
		}
	}
	h := html.Header().Class(Classes("@container"))
	if c.backHref != "" {
		h.Children(html.A(IconArrowLeft().Class("size-4 shrink-0")).Href(c.backHref).Class(Classes(pageBackClass, "mb-3")).Text(c.backTxt))
	}
	if c.crumbs != nil {
		h.Children(html.Div(c.crumbs).Class("mb-3"))
	}
	head := html.Div(heading(c.level, title, pageTitleClass)).Class("min-w-0 flex-1")
	if c.description != "" {
		head.Children(html.P().Class(Classes(pageDescClass)).Text(c.description))
	}
	if len(c.meta) > 0 {
		// Each item is its own flex item, so the gap separates bare text nodes too.
		meta := html.Div().Class(Classes(pageMetaClass))
		for _, m := range c.meta {
			meta.Children(html.Div(m).Class("flex items-center"))
		}
		head.Children(meta)
	}
	row := html.Div(head).Class(Classes("flex flex-col gap-4 @xl:flex-row @xl:items-start @xl:justify-between"))
	if len(c.actions) > 0 {
		row.Children(html.Div(c.actions...).Class(Classes("flex shrink-0 flex-wrap items-center gap-3")))
	}
	return h.Children(row)
}

// heading returns an <h1>..<h6> (level clamped to 1..6) with text and class.
func heading(level int, text, class string) htemel.FlowContent {
	cls := Classes(class)
	switch level {
	case 2:
		return html.H2().Class(cls).Text(text)
	case 3:
		return html.H3().Class(cls).Text(text)
	case 4:
		return html.H4().Class(cls).Text(text)
	case 5:
		return html.H5().Class(cls).Text(text)
	case 6:
		return html.H6().Class(cls).Text(text)
	}
	return html.H1().Class(cls).Text(text)
}

// SectionHeader returns the heading row of a section within a page: an <h2> and actions on the
// right, over a bottom border.
func SectionHeader(title string, actions ...htemel.FlowContent) *html.DivElement {
	row := html.Div(
		html.H2().Class(Classes("text-base/7 font-semibold text-neutral-900 dark:text-white")).Text(title),
	).Class(Classes("flex flex-wrap items-center justify-between gap-3 border-b border-neutral-200 pb-3 dark:border-white/10"))
	if len(actions) > 0 {
		row.Children(html.Div(actions...).Class(Classes("flex items-center gap-2")))
	}
	return row
}

// tabLinkClass is a link drawn as a tab (the Tabs look) whose selected state is
// aria-current="page": the tabs of StackedTabs, and SettingsLayout's nav in a narrow container.
const tabLinkClass = "block border-b-2 border-transparent px-1 py-3 text-sm font-medium whitespace-nowrap text-neutral-500 transition-colors duration-150 " +
	"hover:border-neutral-300 hover:text-neutral-700 focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-primary-600 " +
	"aria-[current=page]:border-primary-600 aria-[current=page]:text-primary-600 " +
	"dark:text-neutral-400 dark:hover:border-white/20 dark:hover:text-neutral-200 dark:focus-visible:outline-primary-500 " +
	"dark:aria-[current=page]:border-primary-400 dark:aria-[current=page]:text-primary-400"

// StackedTabsElement is the row of link tabs built by StackedTabs. StackedLayout moves it into
// its header band.
type StackedTabsElement struct {
	*html.NavElement
}

// StackedTabs returns a row of links drawn as tabs, for the sections of a record (Overview,
// Comments, Revisions): unlike Tabs each one is a page, so the current one is the NavLink marked
// active (aria-current="page"). label names the <nav> landmark. The row scrolls sideways when it
// does not fit. Pass it among StackedLayout's content and it is drawn under the header.
func StackedTabs(label string, items ...NavItem) *StackedTabsElement {
	ul := html.Ul().Class(Classes("-mb-px flex gap-x-6 overflow-x-auto"))
	for _, it := range items {
		ul.Children(html.Li(it.linkClass(tabLinkClass)).Class("shrink-0"))
	}
	return &StackedTabsElement{html.Nav(ul).Aria("label", navLabel([]string{label}, "Sections"))}
}

// stackedPad is the side padding of StackedLayout's bands.
const stackedPad = "px-4 @xl:px-6 @5xl:px-8"

// StackedLayout returns a page body: a header band (white, a bottom border) holding header,
// usually a PageHeader, then the content. A StackedTabs among content moves into the header band,
// under the header. The root is a CSS container: the side padding grows from 1rem to 1.5rem at
// @xl and 2rem at @5xl. It sets no max-width; wrap the content in Container for one.
func StackedLayout(header htemel.FlowContent, content ...htemel.FlowContent) *html.DivElement {
	var tabs *StackedTabsElement
	body := html.Div().Class(Classes("flex-1 py-6", stackedPad))
	for _, c := range content {
		if t, ok := c.(*StackedTabsElement); ok && tabs == nil {
			tabs = t
			continue
		}
		body.Children(c)
	}
	band := html.Div(html.Div(header).Class(Classes("pt-6", stackedPad, cond(tabs == nil, "pb-6"), cond(tabs != nil, "pb-2")))).
		Class(Classes("border-b border-neutral-200 bg-white dark:border-white/10 dark:bg-neutral-900"))
	if tabs != nil {
		band.Children(html.Div(tabs).Class(Classes(stackedPad)))
	}
	return html.Div(band, body).Class(Classes("@container flex min-h-full flex-col"))
}

// settingsLinkClass is the vertical link of SettingsLayout's nav, shown from @xl.
const settingsLinkClass = "block rounded-control px-3 py-2 text-sm font-medium text-neutral-600 transition-colors duration-150 " +
	"hover:bg-neutral-100 hover:text-neutral-900 focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-primary-600 " +
	"aria-[current=page]:bg-neutral-100 aria-[current=page]:text-primary-700 " +
	"dark:text-neutral-400 dark:hover:bg-white/5 dark:hover:text-white dark:focus-visible:outline-primary-500 " +
	"dark:aria-[current=page]:bg-white/10 dark:aria-[current=page]:text-white"

// SettingsLayout returns a settings page: a secondary navigation beside the content. The root is
// a CSS container: from the @xl container width (36rem) the nav is a vertical list of links
// (13rem) left of the content; below it the nav is a labelled Select of the sections above the
// content (a row of tabs would clip its last items on a phone and show a scrollbar), which
// navigates to the chosen link's href on change. The current page is the NavLink marked active
// (aria-current="page", and the selected option); the links keep their icons and badges. The nav
// is a <nav aria-label="Settings">.
//
//	ui.SettingsLayout([]ui.NavItem{
//		ui.NavLink("Profile", "/settings/profile", true),
//		ui.NavLink("Notifications", "/settings/notifications", false),
//	}, profileForm)
func SettingsLayout(nav []NavItem, content ...htemel.FlowContent) *html.DivElement {
	return SettingsLayoutLabel("Settings", nav, content...)
}

// SettingsLayoutLabel is SettingsLayout with label as the name of its <nav> landmark instead of
// "Settings", for a page with two of them (or one inside another settings area).
func SettingsLayoutLabel(label string, nav []NavItem, content ...htemel.FlowContent) *html.DivElement {
	name := navLabel([]string{label}, "Settings")
	ul := html.Ul().Class(Classes("hidden flex-col gap-y-1 @xl:flex"))
	options := make([]html.SelectChild, 0, len(nav))
	for _, it := range nav {
		ul.Children(html.Li(it.linkClass(settingsLinkClass)))
		options = append(options, Choice(it.href, it.label).Selected(it.active))
	}
	selectID := strings.ToLower(domID(name)) + "-section"
	picker := Select("section", options...).
		Attrs(datastar.On("change", "window.location.href = evt.target.value"))
	picker.SetAttr("id", selectID)
	return html.Div(
		html.Div(
			html.Nav(
				html.Div(
					html.Label().For(selectID).Class("sr-only").Text("Section"),
					picker,
				).Class("@xl:hidden"),
				ul,
			).Aria("label", name).Class("min-w-0"),
			html.Div(content...).Class("min-w-0"),
		).Class(Classes("grid grid-cols-1 gap-6 @xl:grid-cols-[13rem_minmax(0,1fr)] @xl:gap-10")),
	).Class(Classes("@container"))
}
