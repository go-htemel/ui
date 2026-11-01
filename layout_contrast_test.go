package ui_test

// layoutRows are the contrast pairs of the layouts (layout_*.go), the NavItem icon, badge and
// ancestor states (nav.go).
func layoutRows() []spec {
	var m []spec
	add := func(component, state, src string, k kind, light, dark side) {
		m = append(m, spec{component: component, state: state, src: src, kind: k, light: light, dark: dark})
	}
	// SkipLink: shown on focus as a white / neutral-800 chip over the page.
	add("SkipLink", "focused text", "layout_shell.go:19,layout_shell.go:21", text, on("focus:text-primary-700", "focus:bg-white"), on("dark:focus:text-white", "dark:focus:bg-neutral-800"))
	add("SkipLink", "focus outline", "layout_shell.go:20,layout_shell.go:21", nonText, on("focus:outline-primary-600", ""), on("dark:focus:outline-primary-500", ""))

	// NavItem.Icon, .Badge, .Ancestor (nav.go:71-80) in a white / neutral-900 sidebar.
	add("Nav link", "icon", "nav.go:111", nonText, on("text-neutral-500", "").over(cardL), on("dark:text-neutral-400", "").over(cardD))
	add("Nav link", "icon, current", "nav.go:110", nonText, on("in-aria-[current=page]:text-primary-700", "aria-[current=page]:bg-neutral-100").over(cardL),
		on("dark:in-aria-[current=page]:text-white", "dark:aria-[current=page]:bg-white/10").over(cardD))
	add("Nav link", "badge", "nav.go:124", text, on("text-neutral-700", "bg-neutral-100").over(cardL), on("dark:text-neutral-300", "dark:bg-white/10").over(cardD))
	add("Nav link", "ancestor", "nav.go:115", text, on("data-ui-ancestor:text-neutral-900", "").over(cardL), on("dark:data-ui-ancestor:text-white", "").over(cardD))

	// AppShell with SidebarDark: the sidebar carries .dark, so its dark: classes apply in BOTH
	// schemes on the neutral-900 surface (layout_shell.go:154-155). Light and dark rows are equal.
	const darkSide = "dark:bg-neutral-900"
	dk := func(fg, bg string) side { return on(fg, bg).over(darkSide) }
	add("AppShell/SidebarDark", "nav link", "nav.go:111,layout_shell.go:155", text, dk("dark:text-neutral-400", ""), dk("dark:text-neutral-400", ""))
	add("AppShell/SidebarDark", "nav link, hover", "nav.go:111", text, dk("dark:hover:text-white", "dark:hover:bg-white/5"), dk("dark:hover:text-white", "dark:hover:bg-white/5"))
	add("AppShell/SidebarDark", "nav link, current", "nav.go:112", text, dk("dark:aria-[current=page]:text-white", "dark:aria-[current=page]:bg-white/10"), dk("dark:aria-[current=page]:text-white", "dark:aria-[current=page]:bg-white/10"))
	add("AppShell/SidebarDark", "nav icon", "nav.go:111", nonText, dk("dark:text-neutral-400", ""), dk("dark:text-neutral-400", ""))
	add("AppShell/SidebarDark", "nav badge", "nav.go:124", text, dk("dark:text-neutral-300", "dark:bg-white/10"), dk("dark:text-neutral-300", "dark:bg-white/10"))
	add("AppShell/SidebarDark", "section heading", "nav.go:272,nav.go:272", text, dk("dark:text-neutral-400", ""), dk("dark:text-neutral-400", ""))
	add("AppShell/SidebarDark", "focus outline", "nav.go:111", nonText, dk("dark:focus-visible:outline-primary-500", ""), dk("dark:focus-visible:outline-primary-500", ""))
	add("AppShell/SidebarDark", "brand / text", "layout_shell.go:155", text, dk("dark:text-white", ""), dk("dark:text-white", ""))
	add("AppShell/SidebarDark", "rail button icon", "button.go:20,button.go:21", nonText, dk("dark:text-neutral-300", ""), dk("dark:text-neutral-300", ""))
	// WithSidebarFooter stays visible in the collapsed rail: a Ghost button (WithRailIcon) holds the
	// avatar and the ColorSchemeToggle stacks; both sit on the sidebar surface (button.go Ghost).
	add("AppShell", "footer button in the rail", "button.go:20,button.go:21", text, on("text-neutral-700", "bg-white"), on("dark:text-neutral-300", "dark:bg-neutral-900"))
	add("AppShell/SidebarDark", "footer button in the rail", "button.go:20,button.go:21", text, dk("dark:text-neutral-300", ""), dk("dark:text-neutral-300", ""))
	add("AppShell", "menu button icon", "layout_shell.go:154,layout_shell.go:154", nonText, on("text-neutral-600", "bg-white"), on("dark:text-neutral-300", "dark:bg-neutral-900"))

	// PageHeader (layout_page.go:54-56) and SectionHeader, on the page and on a card.
	for _, bg := range []struct {
		name string
		l, d string
	}{{"page", pageL, pageD}, {"card", cardL, cardD}} {
		add("PageHeader", "title on the "+bg.name, "layout_page.go:54", largeText, on("text-neutral-900", "").over(bg.l), on("dark:text-white", "").over(bg.d))
		add("PageHeader", "description on the "+bg.name, "layout_page.go:56", text, on("text-neutral-600", "").over(bg.l), on("dark:text-neutral-400", "").over(bg.d))
		add("PageHeader", "meta on the "+bg.name, "layout_page.go:56", text, on("text-neutral-600", "").over(bg.l), on("dark:text-neutral-400", "").over(bg.d))
		add("PageHeader", "back link on the "+bg.name, "layout_page.go:54,layout_page.go:56", text, on("text-neutral-600", "").over(bg.l), on("dark:text-neutral-400", "").over(bg.d))
	}
	add("SectionHeader", "title", "layout_page.go:123", text, on("text-neutral-900", ""), on("dark:text-white", ""))

	// Link tabs (StackedTabs) on the white / neutral-900 header band.
	add("StackedTabs", "tab", "layout_page.go:136,layout_page.go:139", text, on("text-neutral-500", "bg-white"), on("dark:text-neutral-400", "dark:bg-neutral-900"))
	add("StackedTabs", "current tab", "layout_page.go:137,layout_page.go:137", text, on("aria-[current=page]:text-primary-600", "bg-white"), on("dark:aria-[current=page]:text-primary-400", "dark:bg-neutral-900"))
	add("StackedTabs", "current tab underline", "layout_page.go:137,layout_page.go:137", nonText, on("aria-[current=page]:border-primary-600", "bg-white"), on("dark:aria-[current=page]:border-primary-400", "dark:bg-neutral-900"))
	add("SettingsLayout", "link (@xl)", "layout_page.go:188,layout_page.go:188", text, on("text-neutral-600", ""), on("dark:text-neutral-400", ""))
	add("SettingsLayout", "current link (@xl)", "layout_page.go:188,layout_page.go:190", text,
		on("aria-[current=page]:text-primary-700", "aria-[current=page]:bg-neutral-100"), on("dark:aria-[current=page]:text-white", "dark:aria-[current=page]:bg-white/10"))
	add("SplitDetailHeader", "title", "layout_split.go:190,layout_split.go:156", text, on("text-neutral-900", "bg-white"), on("dark:text-white", "dark:bg-neutral-900"))
	add("SplitListHeader", "title", "layout_split.go:206,layout_split.go:156", text, on("text-neutral-900", "bg-white"), on("dark:text-white", "dark:bg-neutral-900"))

	return m
}
