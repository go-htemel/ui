# 0007 — Components adapt to their container, not the viewport

**Status:** Accepted

## Decision
Navbar, AppShell, SplitLayout, SettingsLayout, StackedTabs and PageHeader use Tailwind v4
container queries (`@container`, `@xl:`…) with breakpoints chosen per layout from the
content they hold (documented table in the README).

## Consequences
- Layouts nest and preview correctly at any width; the showcase shows narrow and wide
  states side by side.
- Narrow navigation never scrolls sideways: SettingsLayout and StackedTabs switch to a
  section `Select`, Navbar to a Drawer, Breadcrumbs wrap or collapse.
