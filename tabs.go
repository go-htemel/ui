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
	tabBase     = "-mb-px cursor-pointer border-b-2 px-1 py-3 text-sm font-medium whitespace-nowrap focus-visible:outline-2 focus-visible:outline-sky-600"
	tabActive   = "border-sky-600 text-sky-600 dark:border-sky-400 dark:text-sky-400"
	tabInactive = "border-transparent text-zinc-500 hover:border-zinc-300 hover:text-zinc-700 dark:text-zinc-400 dark:hover:border-white/20 dark:hover:text-zinc-200"
)

// Tabs returns a tab list and its panels, switched entirely in the browser: signal holds the key
// of the selected tab and starts at the first tab unless the page already defines it. Set the
// signal from anywhere (a link, a server patch) to switch tabs.
//
// Arrow-key navigation between tabs is not implemented; every tab is in the Tab order.
func Tabs(signal string, tabs ...TabItem) *html.DivElement {
	if len(tabs) == 0 {
		return html.Div()
	}
	sig := "$" + signal
	active, inactive := Classes(tabActive), Classes(tabInactive)
	list := html.Div().Role("tablist").Class(Classes("flex gap-x-6 overflow-x-auto border-b border-zinc-200 dark:border-white/10"))
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
			Class(Classes(tabBase), state).
			Attrs(
				datastar.Classes("{"+datastar.Quote(active)+": "+selected+", "+datastar.Quote(inactive)+": !("+selected+")}"),
				datastar.Attr("aria-selected", "("+selected+") ? 'true' : 'false'"),
				datastar.OnClick(sig+" = "+datastar.Quote(t.key)),
			).
			Text(t.label))
		panel := html.Div(t.panel...).
			Role("tabpanel").
			Id(id+"-panel").
			Aria("labelledby", id+"-tab").
			Class("pt-4").
			Attrs(datastar.Show(selected))
		if i > 0 {
			panel.Style(hiddenUntilDatastar)
		}
		root.Children(panel)
	}
	return root
}
