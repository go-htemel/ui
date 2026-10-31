package main

import (
	"strings"

	"github.com/go-htemel/ui"
)

// withCSS adds a stylesheet to an example's source blocks, titled with label.
func withCSS(e Example, label, css string) Example {
	e.Snippets = append(append([]Snippet(nil), e.Snippets...), Snippet{Lang: "css", Title: label, Src: strings.TrimSpace(css)})
	return e
}

// The data, navigation and status pages are registered here rather than in groups.go, so the
// list in groups.go keeps its original groups and these are appended after them.
func init() {
	groups = append(groups,
		Group{
			Slug: "data", Title: "Data",
			Intro: "Table with sortable headers, pagination, empty states, stats and description lists. The sortable table sorts and pages on the server over SSE.",
			Examples: []Example{
				ex("Sortable, paged table (server over SSE)", sortableTable),
				ex("Striped, dense and sticky header", tableVariants),
				ex("Empty states", emptyStates),
				ex("Pagination (links)", paginationLinks),
				ex("Pagination (signal)", paginationSignal),
				ex("Stats and description list", statsAndLists),
			},
		},
		Group{
			Slug: "navigation", Title: "Navigation",
			Intro: "Navbar that collapses into a drawer by its own width (container queries), a sidebar with collapsible sections, breadcrumbs and steps.",
			Examples: []Example{
				ex("Navbar", navbar),
				ex("Sidebar", sidebarNav),
				ex("Breadcrumbs", breadcrumbs),
				ex("Steps", steps),
			},
		},
		Group{
			Slug: "status", Title: "Progress & status",
			Intro: "Progress bars, skeletons, spinners, CSS-only tooltips and input groups.",
			Examples: []Example{
				ex("Progress", progressBars),
				ex("Progress bound to a signal", progressSignal),
				ex("Skeleton and spinner", skeletonAndSpinner),
				ex("Tooltip", tooltips),
				ex("Input group", inputGroups),
			},
		},
		Group{
			Slug: "theming", Title: "Theming",
			Intro: "Components use design tokens (primary, neutral, success, warning, danger, info; rounded-control, rounded-surface; font-sans) defined as CSS variables. Override them in CSS, globally or per [data-theme] scope, and everything rethemes live with no Go change.",
			Examples: []Example{
				withCSS(ex("Live theme switch", themeSwitcher), "Page CSS: the [data-theme] token overrides", themeDemoCSS),
				withCSS(ex("Default tokens", tokenSwatches), "ui.ThemeCSS(): the default tokens", ui.ThemeCSS()),
			},
		},
		Group{
			Slug: "code", Title: "Code",
			Intro: "Server-side syntax highlighting from the code package (chroma): no highlighter script in the browser. Line numbers are CSS counters, so copying never includes them; the copy button needs Datastar.",
			Examples: []Example{
				ex("Go with a title, line numbers and highlighted lines", codeGo),
				ex("JSON", codeJSON),
				ex("Bash with line numbers", codeBash),
				ex("Inline code in a sentence", codeInline),
				ex("Scroll or soft wrap", codeWrap),
				ex("Without the copy button", codeNoCopy),
				ex("Capped height", codeMaxHeight),
			},
		},
	)
}
