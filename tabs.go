package ui

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// TabItem is one tab of Tabs, built by Tab.
type TabItem struct {
	key, label string
	panel      []htemel.FlowContent
}

// Tab returns a tab for Tabs: key identifies it in the signal, label is the tab text and panel is
// shown while the tab is selected. Keys should be short words ("overview", "billing").
func Tab(key, label string, panel ...htemel.FlowContent) TabItem {
	return TabItem{key: key, label: label, panel: panel}
}

const (
	tabBase     = "-mb-px cursor-pointer border-b-2 px-1 py-3 text-sm font-medium whitespace-nowrap focus-visible:outline-2 focus-visible:outline-primary-700 dark:focus-visible:outline-primary-500"
	tabActive   = "border-primary-700 text-primary-700 dark:border-primary-400 dark:text-primary-400"
	tabInactive = "border-transparent text-neutral-500 hover:border-neutral-300 hover:text-neutral-700 dark:text-neutral-400 dark:hover:border-white/20 dark:hover:text-neutral-200"
)

// Tabs returns a tab list and its panels, switched entirely in the browser: signal holds the key
// of the selected tab and starts at the first tab unless the page already defines it. Set the
// signal from anywhere (a link, a server patch) to switch tabs.
//
// # Accessibility
//
// The list is role="tablist", tabs are role="tab" with aria-selected bound to the signal and
// aria-controls pointing at their role="tabpanel", which is aria-labelledby the tab. Focus is
// roving: only the selected tab has tabindex="0" (the others -1, bound with data-attr:tabindex),
// so Tab enters the list once and then the panel (which is focusable). ArrowLeft / ArrowRight
// move to the previous / next tab (wrapping) and Home / End to the first / last, selecting as they
// go (automatic activation). Right-to-left layouts are not mirrored, and vertical tab lists are
// not supported.
func Tabs(signal string, tabs ...TabItem) *html.DivElement {
	if len(tabs) == 0 {
		return html.Div()
	}
	sig := "$" + signal
	active, inactive := Classes(tabActive), Classes(tabInactive)
	list := html.Div().Role("tablist").Attrs(datastar.OnKeyDown(tabKeys)).Class(Classes("flex gap-x-6 overflow-x-auto border-b border-neutral-200 dark:border-white/10"))
	root := html.Div(list).Attrs(initially(signal, datastar.Quote(tabs[0].key)))
	for i, t := range tabs {
		id := domID(signal) + "-" + domID(t.key)
		selected := sig + " === " + datastar.Quote(t.key)
		state := inactive
		if i == 0 {
			state = active
		}
		list.Children(html.Button().
			Type(html.ButtonTypeButton).
			Role("tab").
			Id(id+"-tab").
			Aria("controls", id+"-panel").
			Aria("selected", map[bool]string{true: "true", false: "false"}[i == 0]).
			Tabindex(map[bool]int{true: 0, false: -1}[i == 0]).
			Class(Classes(tabBase), state).
			Attrs(
				datastar.Classes("{"+datastar.Quote(active)+": "+selected+", "+datastar.Quote(inactive)+": !("+selected+")}"),
				datastar.Attr("aria-selected", "("+selected+") ? 'true' : 'false'"),
				datastar.Attr("tabindex", "("+selected+") ? '0' : '-1'"),
				datastar.OnClick(sig+" = "+datastar.Quote(t.key)),
			).
			Text(t.label))
		panel := html.Div(t.panel...).
			Role("tabpanel").
			Id(id+"-panel").
			Aria("labelledby", id+"-tab").
			Tabindex(0).
			Class("pt-4 focus-visible:outline-2 focus-visible:outline-primary-700 dark:focus-visible:outline-primary-500").
			Attrs(datastar.Show(selected))
		if i > 0 {
			panel.Style(hiddenUntilDatastar)
		}
		root.Children(panel)
	}
	return root
}

// tabKeys moves focus among the tabs of a tablist and selects the one it lands on by clicking it,
// so the tab's own handler sets the signal.
const tabKeys = "const tabs = [...el.querySelectorAll(\"[role='tab']\")], i = tabs.indexOf(document.activeElement), k = evt.key; " +
	"if (i < 0) return; " +
	"const n = k === 'ArrowRight' ? (i + 1) % tabs.length : k === 'ArrowLeft' ? (i + tabs.length - 1) % tabs.length " +
	": k === 'Home' ? 0 : k === 'End' ? tabs.length - 1 : -1; " +
	"if (n < 0) return; " +
	"evt.preventDefault(); tabs[n].focus(); tabs[n].click()"
