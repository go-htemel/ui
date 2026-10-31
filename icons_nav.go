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
