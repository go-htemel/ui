# Accessibility scan: component-level violations

Source: `task a11y` (axe-core 4.11, WCAG 2.1 A/AA + best-practice, Chromium, 13 pages x light/dark
plus open dialog, open dropdown, switched tab, expanded disclosure/accordion, open mobile navbar).
Everything below comes from a component in the package root or `assets/ui.css`; the showcase-only
problems found by the same scan were fixed in `cmd/showcase/`. Contrast ratios are as measured by
axe on the rendered page (Tailwind 4.3 browser build, default tokens).

Totals: 44 WCAG nodes (40 `color-contrast`, 4 `aria-prohibited-attr`) and 10 best-practice nodes.

## 1. color-contrast (WCAG 1.4.3, AA)

| # | Light/dark | Ratio (need 4.5) | Colors | Component | Where | Pages |
| --- | --- | --- | --- | --- | --- | --- |
| 1 | light and dark | 4.06 | `#ffffff` on `#0084d1` (`bg-primary-600`, sky-600) | `Button` / `IconButton` Primary (`buttonVariants[Primary]`), all sizes Xs to Lg | `button.go:14-15` (`bg-primary-600 text-white`, `dark:bg-primary-600`); token `--color-primary-600: var(--color-sky-600)` at `assets/ui.css:40` | buttons (10 per scheme), layout, feedback, data, status, theming |
| 2 | light and dark | 4.06 | `#ffffff` on `#0084d1` | `Pagination` current page (`aria-[current=page]:bg-primary-600 ...:text-white`) | `pagination.go:12-17` (`pageItemClass`) | data |
| 3 | dark only | 3.80 | `#ffffff` on `#fb2c36` (`dark:bg-danger-500`, red-500) | `Button` Danger | `button.go:22-23` (`dark:bg-danger-500`); light passes with red-600 | buttons (2), feedback, overlays (default, dialog open, dropdown open), theming |
| 4 | light only | 4.06 | `#0084d1` on `#ffffff` (`text-primary-600`) | `Button` Link variant | `button.go:24` | buttons |
| 5 | light only | 4.06 | `#0084d1` on `#ffffff` | `Tabs` active tab label (`tabActive`) | `tabs.go:23` (`text-primary-600`) | disclosure (default, tab switched, expanded) |
| 6 | light only | 3.24 | `#00a63e` on `#ffffff` (`text-success-600`, green-600) | `Stat` positive delta (`+12.5%`) | `info.go:56` | data |

Snippets (axe truncates class lists):

```html
<button type="button" class="inline-flex items-center ... bg-primary-600 text-white shadow-xs hover:bg-primary-700 ... dark:bg-primary-600 ...">Primary</button>   <!-- #1 -->
<a href="#page-4" class="inline-flex min-w-9 ..." aria-label="Page 24" aria-current="page">24</a>                                                              <!-- #2 -->
<button type="button" class="... bg-danger-600 text-white hover:bg-danger-500 ... dark:bg-danger-500 ...">Delete</button>                                    <!-- #3, dark -->
<button role="tab" id="settingsTab-profile-tab" aria-selected="true" class="... border-primary-600 text-primary-600 ...">Profile</button>                      <!-- #5, light -->
<p class="shrink-0 text-sm font-medium whitespace-nowrap text-success-600 dark:text-success-400">+12.5%</p>                                                    <!-- #6, light -->
```

Notes for the fix: #1, #2, #4 and #5 share one cause, `--color-primary-600` (sky-600, 4.06:1
against white). Moving the default `primary-600` token to a darker stop (sky-700 is about 5.9:1),
or using `primary-700` for filled buttons and text, clears all four at once without touching the
dark-mode Primary hover. #3 needs `dark:bg-danger-600` (or white text swapped for dark text). #6
needs `text-success-700` in light. The Go contrast test from the other agent should cover the same
pairs; this scan confirms them on the rendered pages.

## 2. aria-prohibited-attr (WCAG 4.1.2, serious), 2 per scheme, page data

Component: `Pagination`, disabled previous/next (`pagination.go:57-61`). The disabled item is an
`<a>` with no `href`, so it has no role, and `aria-label` is not allowed on a generic element.

```html
<a class="inline-flex min-w-9 ... pointer-events-none opacity-50" aria-label="Previous page" aria-disabled="true"><svg ...></svg></a>
```

Fix direction: give it `role="link"` with `aria-disabled="true"` (the usual pattern for a
disabled link), or render a `<span>` with the label as visually hidden text.

## 3. landmark-unique (best-practice, moderate; does not fail the task)

Two navs with the same name on one page:

| Page | Landmarks | Component |
| --- | --- | --- |
| navigation (both schemes, also with mobile nav open) | 2x `<nav aria-label="Breadcrumb">`, 2x `<nav aria-label="Progress">` | `Breadcrumbs` (`nav.go:186`), `Steps` (`progress.go:166`) |
| data (both schemes) | `<nav aria-label="Pagination">` twice (links and signal examples, `#people-pager`) | `Pagination`, `PaginationSignal` (`pagination.go:132`) |

The names are hard-coded, so a page with two of the same component cannot distinguish them. It is
only an issue where an app renders two; fix would be an option to override the label. The
showcase shows two of each on purpose, so it stays as advisory output.

## 4. Component API issue found while fixing the showcase

`InputGroup` is documented as "put the group inside a Field to label it" (`info.go:106-108`), but
`Field` sets `id` and `label[for]` on the control it is given (`form.go:81-83`), which is the
group's wrapper `<div>`, not the `<input>`. The `<label for>` then points at a non-labelable
element and the input has no accessible name (axe `label`, critical, 3 per scheme on the status
page). The showcase now works around it with `aria-label` on the inner inputs. Fix direction:
`Field` should detect an `InputGroup` and put the id on the inner control, or `InputGroup` should
accept the id and forward it.

## 5. Not violations, for completeness

- axe `incomplete` ("needs review") count is printed by the task; those are mostly contrast over
  elements axe cannot resolve a background for and are not failures.
- Harness note: axe-core 4.11 cannot parse the `oklch(0.985 0 none)` colors Chromium reports for
  Tailwind v4 neutrals, and silently skipped the whole `color-contrast` rule in the light scheme.
  `run.mjs` normalizes `none` to `0` before axe reads the styles and fails the run if any rule
  errors, so the light scheme is really measured.
