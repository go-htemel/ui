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
// # Accessibility
//
// The trigger gets aria-haspopup="menu", aria-controls and a bound aria-expanded; the menu is
// role="menu" labelled by the trigger and items are role="menuitem" with tabindex="-1" (roving
// focus, so they are not in the Tab order). Keyboard, handled by Datastar expressions:
//
//   - ArrowDown / ArrowUp on the trigger open the menu and focus the first / last item.
//   - ArrowDown / ArrowUp / Home / End move focus among the items (wrapping).
//   - Escape closes and returns focus to the trigger; so does choosing an item.
//   - Tab closes the menu and lets focus move on.
//
// Not implemented: type-ahead (jumping to an item by its first letter), which needs a
// character buffer with a timeout and so state beyond what an expression can sensibly hold;
// submenus; and focusing the first item when the menu is opened with the mouse.
func Dropdown(signal string, trigger *html.ButtonElement, items ...htemel.FlowContent) *html.DivElement {
	menuID := domID(signal) + "-menu"
	trigger.Id(domID(signal)+"-trigger").Aria("haspopup", "menu").Aria("controls", menuID).
		Attrs(datastar.OnClick(Toggle(signal))).
		Attrs(expanded("$" + signal)...)
	menu := html.Div(items...).
		Id(menuID).
		Role("menu").
		Aria("labelledby", domID(signal)+"-trigger").
		Class(Classes("absolute right-0 z-20 mt-2 w-56 origin-top-right rounded-md bg-white p-1 shadow-lg ring-1 ring-zinc-950/5",
			"dark:bg-zinc-800 dark:ring-white/10")).
		Style(hiddenUntilDatastar).
		Attrs(datastar.Show("$"+signal), datastar.OnClick(Close(signal)+"; el.previousElementSibling.focus()"))
	return html.Div(trigger, menu).
		Class(Classes("relative inline-block text-left")).
		Attrs(
			initially(signal, "false"),
			datastar.OnClick(Close(signal), datastar.Outside()),
			datastar.OnKeyDown(menuKeys(signal)),
			datastar.On("keydown", menuEscape(signal), datastar.Window()),
		)
}

// menuKeys moves focus among the menu items. The root's first child is the trigger and the
// second the menu.
func menuKeys(signal string) string {
	return "const t = el.firstElementChild, a = document.activeElement, k = evt.key; " +
		"if (k === 'Tab') { " + Close(signal) + "; return } " +
		"if (!['ArrowDown', 'ArrowUp', 'Home', 'End'].includes(k)) return; " +
		"if (!$" + signal + " && (a !== t || k === 'Home' || k === 'End')) return; " +
		"const items = [...t.nextElementSibling.querySelectorAll(\"[role='menuitem']:not([disabled])\")], i = items.indexOf(a); " +
		"if (!items.length) return; " +
		"evt.preventDefault(); " +
		Open(signal) + "; " +
		"const n = k === 'Home' ? 0 : k === 'End' ? -1 : k === 'ArrowDown' ? i + 1 : i < 0 ? -1 : i - 1; " +
		"setTimeout(() => items[(n + items.length) % items.length].focus())"
}

// menuEscape closes an open menu on Escape, anywhere in the window, and returns focus to the
// trigger when it was inside the menu.
func menuEscape(signal string) string {
	return "if (evt.key !== 'Escape' || !$" + signal + ") return; " + Close(signal) + "; " +
		"if (el.contains(document.activeElement)) el.firstElementChild.focus()"
}

const menuItemClass = "flex w-full cursor-pointer items-center gap-x-2 rounded-md px-3 py-2 text-left text-sm text-zinc-700 " +
	"hover:bg-zinc-100 hover:text-zinc-900 focus-visible:bg-zinc-100 focus-visible:outline-none " +
	"dark:text-zinc-300 dark:hover:bg-white/5 dark:hover:text-white dark:focus-visible:bg-white/5"

// MenuItem returns a menu entry, <button role="menuitem">. attrs usually include a Datastar click
// handler: MenuItem("Archive", datastar.OnClick(datastar.Post("/archive"))).
func MenuItem(label string, attrs ...htemel.Attribute) *html.ButtonElement {
	return html.Button().Type(html.ButtonTypeButton).Role("menuitem").Tabindex(-1).Class(Classes(menuItemClass)).Attrs(attrs...).Text(label)
}

// MenuLink returns a menu entry that navigates to href, <a role="menuitem">.
func MenuLink(label, href string) *html.AElement {
	return html.A().Href(href).Role("menuitem").Tabindex(-1).Class(Classes(menuItemClass)).Text(label)
}

// MenuDivider returns a rule between groups of menu items.
func MenuDivider() *html.HrElement {
	return html.Hr().Class(Classes("-mx-1 my-1 border-zinc-100 dark:border-white/10"))
}
