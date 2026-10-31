package ui

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// Dropdown returns a trigger button with a menu below it, open while signal is true (the signal
// starts false unless the page defines it). The trigger toggles the menu; a click on an item, a
// click outside and Escape close it. Items are usually MenuItem, MenuLink and MenuDivider:
//
//	ui.Dropdown("rowMenu", ui.Button(ui.Secondary, "Options", ui.WithTrailingIcon(ui.IconChevronDown())),
//		ui.MenuItem("Edit", datastar.OnClick(datastar.Get("/rows/1/edit"))),
//		ui.MenuDivider(),
//		ui.MenuItem("Delete", datastar.OnClick(datastar.Delete("/rows/1"))),
//	)
//
// Arrow-key navigation between items is not implemented; Tab moves through them.
func Dropdown(signal string, trigger *html.ButtonElement, items ...htemel.FlowContent) *html.DivElement {
	menuID := domID(signal) + "-menu"
	trigger.Aria("haspopup", "menu").Aria("controls", menuID).
		Attrs(datastar.OnClick(Toggle(signal))).
		Attrs(expanded("$" + signal)...)
	menu := html.Div(items...).
		Id(menuID).
		Role("menu").
		Class(Classes("absolute right-0 z-20 mt-2 w-56 origin-top-right rounded-md bg-white p-1 shadow-lg ring-1 ring-zinc-950/5",
			"dark:bg-zinc-800 dark:ring-white/10")).
		Style(hiddenUntilDatastar).
		Attrs(datastar.Show("$"+signal), datastar.OnClick(Close(signal)))
	return html.Div(trigger, menu).
		Class(Classes("relative inline-block text-left")).
		Attrs(initially(signal, "false"), datastar.OnClick(Close(signal), datastar.Outside()), closeOnEscape(signal))
}

const menuItemClass = "flex w-full cursor-pointer items-center gap-x-2 rounded-md px-3 py-2 text-left text-sm text-zinc-700 " +
	"hover:bg-zinc-100 hover:text-zinc-900 focus-visible:bg-zinc-100 focus-visible:outline-none " +
	"dark:text-zinc-300 dark:hover:bg-white/5 dark:hover:text-white dark:focus-visible:bg-white/5"

// MenuItem returns a menu entry, <button role="menuitem">. attrs usually include a Datastar click
// handler: MenuItem("Archive", datastar.OnClick(datastar.Post("/archive"))).
func MenuItem(label string, attrs ...htemel.Attribute) *html.ButtonElement {
	return html.Button().Type(html.ButtonTypeButton).Role("menuitem").Class(Classes(menuItemClass)).Attrs(attrs...).Text(label)
}

// MenuLink returns a menu entry that navigates to href, <a role="menuitem">.
func MenuLink(label, href string) *html.AElement {
	return html.A().Href(href).Role("menuitem").Class(Classes(menuItemClass)).Text(label)
}

// MenuDivider returns a rule between groups of menu items.
func MenuDivider() *html.HrElement {
	return html.Hr().Class(Classes("-mx-1 my-1 border-zinc-100 dark:border-white/10"))
}
