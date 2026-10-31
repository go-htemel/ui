package ui

import "github.com/go-htemel/htemel/svg"

// Text-formatting icons, for the MarkdownEditor toolbar and your own editors. Same grid and
// stroke as the other inline icons.

// IconBold is a bold B.
func IconBold() *svg.SvgElement {
	return stroked("M7 5h6a3.5 3.5 0 0 1 0 7H7zM7 12h7a3.5 3.5 0 0 1 0 7H7z")
}

// IconItalic is a slanted I.
func IconItalic() *svg.SvgElement { return stroked("M19 4h-9M14 20H5M15 4L9 20") }

// IconStrikethrough is an S struck through.
func IconStrikethrough() *svg.SvgElement {
	return stroked("M4 12h16M16.5 7.5C16 6 14.5 5 12 5c-2.8 0-4.5 1.4-4.5 3.2 0 1.2.7 2.2 2.3 2.8M7.5 16.5C8 18 9.5 19 12 19c2.8 0 4.5-1.4 4.5-3.2 0-.6-.2-1.2-.6-1.8")
}

// IconHeading is an H.
func IconHeading() *svg.SvgElement { return stroked("M6 4v16M18 4v16M6 12h12") }

// IconQuote is a pair of quotation marks.
func IconQuote() *svg.SvgElement {
	return stroked("M10 11H6a1 1 0 0 1-1-1V7a1 1 0 0 1 1-1h3a1 1 0 0 1 1 1v4c0 3-1.5 5-4 6M19 11h-4a1 1 0 0 1-1-1V7a1 1 0 0 1 1-1h3a1 1 0 0 1 1 1v4c0 3-1.5 5-4 6")
}

// IconCode is a pair of angle brackets with a slash.
func IconCode() *svg.SvgElement { return stroked("M8 8l-4 4 4 4M16 8l4 4-4 4M14 5l-4 14") }

// IconLink is two chain links.
func IconLink() *svg.SvgElement {
	return stroked("M10 14a4 4 0 0 0 5.66 0l3-3a4 4 0 0 0-5.66-5.66l-1 1M14 10a4 4 0 0 0-5.66 0l-3 3a4 4 0 0 0 5.66 5.66l1-1")
}

// IconImage is a framed picture with a hill and a sun.
func IconImage() *svg.SvgElement {
	return stroked("M5 4h14a1 1 0 0 1 1 1v14a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1V5a1 1 0 0 1 1-1zM4 16l5-5 4 4 2.5-2.5L20 17M15.5 8.5h.01")
}

// IconListBullet is a bulleted list.
func IconListBullet() *svg.SvgElement {
	return stroked("M9 6h11M9 12h11M9 18h11M4.5 6h.01M4.5 12h.01M4.5 18h.01")
}

// IconListNumbered is a numbered list.
func IconListNumbered() *svg.SvgElement {
	return stroked("M10 6h10M10 12h10M10 18h10M4 4.5l1.5-1V9M4 9h3M4 15a1.5 1.5 0 0 1 3 0c0 1.2-3 2-3 3.5h3")
}

// IconListTask is a checklist.
func IconListTask() *svg.SvgElement {
	return stroked("M11 6h9M11 12h9M11 18h9M3.5 6l1.5 1.5L8 4.5M4 10h4v4H4zM4 16h4v4H4z")
}

// IconTable is a grid of cells.
func IconTable() *svg.SvgElement {
	return stroked("M4 5h16a1 1 0 0 1 1 1v12a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V6a1 1 0 0 1 1-1zM3 10h18M3 15h18M9 5v14M15 5v14")
}
