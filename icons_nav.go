package ui

import "github.com/go-htemel/htemel/svg"

// IconChevronLeft is a left-pointing chevron, for previous buttons.
func IconChevronLeft() *svg.SvgElement { return stroked("M15 6l-6 6 6 6") }

// IconChevronRight is a right-pointing chevron, for next buttons and breadcrumb separators.
func IconChevronRight() *svg.SvgElement { return stroked("M9 6l6 6-6 6") }

// IconEllipsis is three horizontal dots: an elided range in Pagination, or a "more" menu trigger.
func IconEllipsis() *svg.SvgElement { return stroked("M5 12h.01M12 12h.01M19 12h.01") }

// IconMenu is three horizontal bars, the button that opens a mobile menu.
func IconMenu() *svg.SvgElement { return stroked("M4 6h16M4 12h16M4 18h16") }

// IconHome is a house, for a dashboard or home link.
func IconHome() *svg.SvgElement {
	return stroked("M3 10.5L12 3l9 7.5M5 9.5V20a1 1 0 0 0 1 1h4v-6h4v6h4a1 1 0 0 0 1-1V9.5")
}

// IconDocument is a page with lines of text, for posts, pages and notes.
func IconDocument() *svg.SvgElement {
	return stroked("M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8zM14 3v5h5M9 13h6M9 17h6")
}

// IconFile is a blank page, for a file in an upload list.
func IconFile() *svg.SvgElement {
	return stroked("M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8zM14 3v5h5")
}

// IconFolder is a folder, for collections and categories.
func IconFolder() *svg.SvgElement {
	return stroked("M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z")
}

// IconUsers is two people, for team and member links.
func IconUsers() *svg.SvgElement {
	return stroked("M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2M9 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8zM22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75")
}

// IconSettings is three sliders, for settings and preferences.
func IconSettings() *svg.SvgElement {
	return stroked("M4 6h9M17 6h3M4 12h3M11 12h9M4 18h11M19 18h1M15 4v4M9 10v4M17 16v4")
}

// IconSidebar is a panel with a left column, the button that collapses or expands a sidebar.
func IconSidebar() *svg.SvgElement {
	return stroked("M4 4h16a1 1 0 0 1 1 1v14a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V5a1 1 0 0 1 1-1zM9 4v16")
}

// IconArrowLeft is a left arrow, for back links and buttons.
func IconArrowLeft() *svg.SvgElement { return stroked("M19 12H5M11 6l-6 6 6 6") }

// IconUpload is an arrow rising out of a tray, for upload areas.
func IconUpload() *svg.SvgElement {
	return stroked("M12 15V4M7 9l5-5 5 5M4 15v4a1 1 0 0 0 1 1h14a1 1 0 0 0 1-1v-4")
}
