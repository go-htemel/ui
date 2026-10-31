package main

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
			Intro: "Navbar with a mobile menu, a sidebar with collapsible sections, breadcrumbs and steps.",
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
	)
}
