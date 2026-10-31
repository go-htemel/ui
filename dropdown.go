package ui

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// Dropdown returns a trigger button and a menu shown with the Popover API: the menu is a
// popover="auto" element and the trigger its popovertarget, so the browser toggles it, closes it
// on a click outside and on Escape, and draws it in the top layer, even before Datastar loads.
// Choosing an item closes it too. Items are usually MenuItem, MenuLink and MenuDivider:
//
//	ui.Dropdown("rowMenu", ui.Button(ui.Secondary, "Options", ui.WithTrailingIcon(ui.IconChevronDown())),
//		ui.MenuItem("Edit", datastar.OnClick(datastar.Get("/rows/1/edit"))),
//		ui.MenuDivider(),
//		ui.MenuItem("Delete", datastar.OnClick(datastar.Delete("/rows/1"))),
//	)
//
// signal mirrors the open state (it starts false unless the page defines it): the menu's toggle
// event keeps it current, and setting it with ui.Open or ui.Close shows or hides the menu.
// Nothing needs it to open the menu.
//
// The menu is placed below the trigger, right edges aligned, with CSS anchor positioning
// (inline anchor-name on the trigger, position-anchor and position-area on the menu, so no
// per-menu class reaches the Tailwind safelist), flipping above
// the trigger when there is no room below. Browsers without anchor positioning get the same
// placement from the trigger's bounding box, computed when the menu opens; there it does not
// follow the trigger while the page scrolls or flip at the viewport edge.
//
// # Accessibility
//
// The trigger gets aria-haspopup="menu", aria-controls and a bound aria-expanded; the menu is
// role="menu" labelled by the trigger and items are role="menuitem" with tabindex="-1" (roving
// focus, so they are not in the Tab order).
//
//   - Native (popover="auto"): Escape closes the menu and returns focus to the element focused
//     when it opened; a click outside closes it; opening another auto popover closes it.
//   - ArrowDown / ArrowUp on the trigger open the menu and focus the first / last item.
//   - ArrowDown / ArrowUp / Home / End move focus among the items (wrapping).
//   - Choosing an item closes the menu and focuses the trigger.
//   - Tab closes the menu and lets focus move on.
//
// Not implemented: type-ahead (jumping to an item by its first letter), which needs a
// character buffer with a timeout and so state beyond what an expression can sensibly hold;
// submenus; and focusing the first item when the menu is opened with the mouse.
func Dropdown(signal string, trigger *html.ButtonElement, items ...htemel.FlowContent) *html.DivElement {
	triggerID := domID(signal) + "-trigger"
	menuID := domID(signal) + "-menu"
	anchor := "--" + domID(signal) + "-anchor"
	trigger.Id(triggerID).
		Popovertarget(menuID).
		Popovertargetaction(html.PopovertargetactionToggle).
		Aria("haspopup", "menu").Aria("controls", menuID).
		Style("anchor-name: " + anchor).
		Attrs(expanded("$" + signal)...)
	menu := html.Div(items...).
		Id(menuID).
		Popover(html.PopoverAuto).
		Role("menu").
		Aria("labelledby", triggerID).
		Style("position-anchor: "+anchor+"; position-area: bottom span-left; position-try-fallbacks: flip-block").
		Class(Classes("mt-2 w-56 rounded-surface bg-white p-1 shadow-lg ring-1 ring-neutral-950/5",
			"dark:bg-neutral-800 dark:ring-white/10", menuMotion)).
		Attrs(
			datastar.On("beforetoggle", menuPlace(triggerID)),
			datastar.On("toggle", "$"+signal+" = evt.newState === 'open'"),
			datastar.Effect(menuSync(signal)),
			datastar.OnClick("el.hidePopover(); el.previousElementSibling.focus()"),
		)
	return html.Div(trigger, menu).
		Class(Classes("relative inline-block text-left")).
		Attrs(
			initially(signal, "false"),
			datastar.OnKeyDown(menuKeys()),
		)
}

// menuMotion fades the menu in while it moves down by 1px, and back out on close, in 200ms.
// starting:open: (@starting-style) gives the first frame after showPopover() the closed values,
// and transition-discrete keeps the popover displayed and in the top layer until it has faded.
const menuMotion = "opacity-0 -translate-y-px open:opacity-100 open:translate-y-0 starting:open:opacity-0 starting:open:-translate-y-px " +
	"transition-[opacity,translate,display,overlay] transition-discrete duration-200 ease-out motion-reduce:transition-none"

// menuSync shows or hides the menu when the signal is set from outside (ui.Open, ui.Close or the
// server). The toggle event sets the signal to the state the menu is already in, so it does not
// loop.
func menuSync(signal string) string {
	return "if ($" + signal + " !== el.matches(\":popover-open\")) $" + signal + " ? el.showPopover() : el.hidePopover()"
}

// menuPlace positions the menu under the trigger in browsers without CSS anchor positioning. It
// runs on beforetoggle, which fires synchronously before the menu is shown, so it never paints
// at the popover's default place in the middle of the viewport.
func menuPlace(triggerID string) string {
	return "if (evt.newState !== 'open' || CSS.supports('position-anchor: --a')) return; " +
		"const r = document.getElementById('" + triggerID + "').getBoundingClientRect(); " +
		"el.style.inset = 'auto'; el.style.top = r.bottom + 'px'; " +
		"el.style.right = (document.documentElement.clientWidth - r.right) + 'px'"
}

// menuKeys opens the menu from the trigger with the arrow keys and moves focus among the items.
// The root's first child is the trigger and the second the menu.
func menuKeys() string {
	return "const t = el.firstElementChild, m = t.nextElementSibling, a = document.activeElement, k = evt.key, " +
		"open = m.matches(\":popover-open\"); " +
		"if (k === 'Tab') { if (open) m.hidePopover(); return } " +
		"if (!['ArrowDown', 'ArrowUp', 'Home', 'End'].includes(k)) return; " +
		"if (!open && (a !== t || k === 'Home' || k === 'End')) return; " +
		"const items = [...m.querySelectorAll(\"[role='menuitem']:not([disabled])\")], i = items.indexOf(a); " +
		"if (!items.length) return; " +
		"evt.preventDefault(); " +
		"if (!open) m.showPopover(); " +
		"const n = k === 'Home' ? 0 : k === 'End' ? -1 : k === 'ArrowDown' ? i + 1 : i < 0 ? -1 : i - 1; " +
		"items[(n + items.length) % items.length].focus()"
}

const menuItemClass = "flex w-full cursor-pointer items-center gap-x-2 rounded-control px-3 py-2 text-left text-sm text-neutral-700 transition-colors duration-150 " +
	"hover:bg-neutral-100 hover:text-neutral-900 focus-visible:bg-neutral-100 focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-primary-600 " +
	"dark:focus-visible:outline-primary-400 dark:text-neutral-300 dark:hover:bg-white/5 dark:hover:text-white dark:focus-visible:bg-white/5"

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
	return html.Hr().Class(Classes("-mx-1 my-1 border-neutral-100 dark:border-white/10"))
}
