// Accessibility scan of the showcase: every page, light and dark, plus interactive states
// (open dialog, drawer and bottom sheet, open dropdown, switched tab, expanded accordion, open
// mobile nav drawer of the narrow navbar, the open customizable select picker, the color scheme
// toggle switched to the other scheme, the server-rendered Markdown editor preview). Every page
// must also load without console errors (requests to the blocked external CDNs excepted).
//
// Dark mode follows the .dark class on <html> (ui.css's @custom-variant), set by the inline
// ui.ColorSchemeBootstrap script from localStorage or, in system mode, from prefers-color-scheme.
// So each scheme is tested by emulating the OS setting, which the bootstrap picks up, and every
// scan first checks that both the class and the painted background really are that scheme.
//
//   BASE_URL=http://127.0.0.1:8080 node run.mjs [--json out.json]
//
// Fails (exit 1) on any WCAG A/AA violation. Rules tagged only "best-practice" are reported
// separately and never fail the run. Exit 2 means the scan itself could not run properly
// (page unreachable, Tailwind or Datastar did not load): never a vacuous pass.
import { chromium } from "playwright";
import AxeBuilder from "@axe-core/playwright";
import { writeFileSync } from "node:fs";

const BASE = (process.env.BASE_URL || "").replace(/\/$/, "");
if (!BASE) {
  console.error("BASE_URL is required, e.g. BASE_URL=http://127.0.0.1:8080");
  process.exit(2);
}
// A11Y_CDN_DIR: serve the showcase's CDN assets from local files (tailwind.js, datastar.js) instead of
// the network, for sandboxes that block the CDNs. Font Awesome is then stubbed out (the fonts are self-hosted by ui.Assets).
const CDN_DIR = process.env.A11Y_CDN_DIR || "";
const jsonOut = process.argv.includes("--json") ? process.argv[process.argv.indexOf("--json") + 1] : "";

const PAGES = ["/", ...["buttons", "display", "layout", "feedback", "forms", "overlays", "disclosure",
  "drag-and-drop", "data", "navigation", "status", "theming", "code", "markdown", "prose"].map((s) => `/components/${s}`)];
// A11Y_PAGES=forms,overlays limits the run to those pages (by the last path segment, "/" for the
// home page) while iterating; task a11y always runs them all.
if (process.env.A11Y_PAGES) {
  const only = process.env.A11Y_PAGES.split(",").map((p) => p.trim());
  PAGES.splice(0, PAGES.length, ...PAGES.filter((p) => only.includes(p === "/" ? "/" : p.split("/").pop())));
}
const SCHEMES = ["light", "dark"];
const TAGS = ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "best-practice"];
const WCAG = /^wcag\d+a{1,3}$/;

// Interactive states, keyed by page. Each opens something, and returns a function that proves
// the state really changed (so a broken Datastar cannot make the scan pass on the closed state).
const STATES = {
  "/components/overlays": [
    {
      name: "dialog open",
      async open(page) {
        await page.getByRole("button", { name: "Delete project" }).click();
        await page.locator("dialog[open]").waitFor({ timeout: 5000 });
      },
      async close(page) { await page.keyboard.press("Escape"); await page.locator("#confirmDelete-dialog").waitFor({ state: "hidden", timeout: 5000 }); },
    },
    {
      name: "drawer open",
      async open(page) {
        await page.getByRole("button", { name: "Edit profile" }).click();
        await page.locator("#profileDrawer-dialog[open]").waitFor({ timeout: 5000 });
      },
      async close(page) { await page.keyboard.press("Escape"); await page.locator("#profileDrawer-dialog").waitFor({ state: "hidden", timeout: 5000 }); },
    },
    {
      name: "bottom sheet open",
      async open(page) {
        await page.getByRole("button", { name: "Share" }).click();
        await page.locator("#shareSheet-dialog[open]").waitFor({ timeout: 5000 });
      },
      async close(page) { await page.keyboard.press("Escape"); await page.locator("#shareSheet-dialog").waitFor({ state: "hidden", timeout: 5000 }); },
    },
    {
      name: "dropdown open",
      async open(page) {
        await page.getByRole("button", { name: "Options" }).click();
        await page.locator("[popover]:popover-open").waitFor({ timeout: 5000 });
      },
      async close(page) { await page.keyboard.press("Escape"); },
    },
  ],
  "/components/buttons": [
    {
      // Click the opposite scheme in the sidebar toggle: the class, the stored choice, aria-pressed
      // and the painted page must all follow, and the page is scanned in the other scheme. close
      // clicks System, which must restore the OS scheme.
      name: "color scheme toggled",
      async open(page, scheme) {
        const other = scheme === "dark" ? "Light" : "Dark";
        const group = page.getByRole("group", { name: "Color scheme" });
        await group.getByRole("button", { name: other, exact: true }).click();
        await group.getByRole("button", { name: other, exact: true }).and(page.locator("[aria-pressed=true]")).waitFor({ timeout: 5000 });
        await waitForScheme(page, scheme === "dark" ? "light" : "dark", "after clicking " + other);
        const stored = await page.evaluate(() => localStorage.getItem("color-scheme"));
        if (stored !== other.toLowerCase()) throw new Error(`localStorage["color-scheme"] is ${stored}, want ${other.toLowerCase()}`);
      },
      async close(page, scheme) {
        const group = page.getByRole("group", { name: "Color scheme" });
        await group.getByRole("button", { name: "System", exact: true }).click();
        await group.getByRole("button", { name: "System", exact: true }).and(page.locator("[aria-pressed=true]")).waitFor({ timeout: 5000 });
        await waitForScheme(page, scheme, "after clicking System");
        const stored = await page.evaluate(() => localStorage.getItem("color-scheme"));
        if (stored !== null) throw new Error(`System left localStorage["color-scheme"] = ${stored}`);
      },
    },
  ],
  "/components/code": [
    {
      // The copy button writes the block's source to the clipboard (permissions are granted to the
      // context), shows the check mark and announces "Copied" through its live region; the
      // signal resets after 1.5s, which close waits for.
      name: "code copied",
      async open(page) {
        const fig = page.locator("figure.ui-code").first();
        await fig.getByRole("button", { name: "Copy code" }).click();
        await fig.locator("[aria-live=polite]").getByText("Copied").waitFor({ timeout: 5000 });
        const [clip, src] = await page.evaluate(async () => {
          const f = document.querySelector("figure.ui-code");
          return [await navigator.clipboard.readText(), f.querySelector("template").content.textContent];
        });
        if (clip !== src || !clip.startsWith("// Save stores")) throw new Error(`clipboard holds ${JSON.stringify(clip.slice(0, 40))}, want the block source`);
      },
      async close(page) {
        await page.locator("figure.ui-code").first().locator("[aria-live=polite]").getByText("Copied").waitFor({ state: "detached", timeout: 5000 }).catch(() => {});
        await page.waitForFunction(() => document.querySelector("figure.ui-code [aria-live=polite]").textContent === "", null, { timeout: 5000 });
      },
    },
  ],
  "/components/disclosure": [
    {
      name: "tab switched",
      async open(page) {
        await page.getByRole("tab", { name: "Billing" }).click();
        await page.locator('[role=tab][aria-selected=true]', { hasText: "Billing" }).waitFor({ timeout: 5000 });
      },
    },
    {
      name: "disclosure + accordion expanded",
      async open(page) {
        await page.getByRole("button", { name: "Show technical details" }).click();
        await page.getByRole("button", { name: "Show technical details" }).and(page.locator('[aria-expanded=true]')).waitFor({ timeout: 5000 });
        const q = page.getByRole("button", { name: "Do I need to write JavaScript?" });
        await q.click();
        await page.getByRole("button", { name: "Do I need to write JavaScript?" }).and(page.locator('[aria-expanded=true]')).waitFor({ timeout: 5000 });
      },
    },
  ],
  "/components/forms": [
    {
      // appearance: base-select (Chromium 135+) makes the picker DOM in the top layer, so axe can
      // scan its options. Skipped where the browser lacks base-select.
      name: "select picker open",
      async open(page) {
        const supported = await page.evaluate(() => CSS.supports("appearance", "base-select"));
        if (!supported) { console.log("  (base-select unsupported: picker state skipped)"); return; }
        await page.locator("select#plan").click();
        await page.waitForFunction(() => document.querySelector("select#plan").matches(":open"), null, { timeout: 5000 });
      },
      async close(page) { await page.keyboard.press("Escape"); },
    },
  ],
  "/components/markdown": [
    {
      // The Preview tab posts the signals; the server renders the Markdown (with the code package
      // highlighting the fenced Go) and patches the preview, which is then scanned.
      name: "preview rendered",
      async open(page) {
        await page.locator("#journal-preview-tab").click();
        await page.locator("#journal-preview-tab[aria-selected=true]").waitFor({ timeout: 5000 });
        await page.locator("#journal-preview .ui-prose figure.ui-code .tok-keyword").first().waitFor({ timeout: 5000 });
        const toolbar = await page.locator("#journal-write-panel").evaluate((p) => getComputedStyle(p.closest(".ui-mdeditor").querySelector("[role=toolbar]")).display);
        if (toolbar !== "none") throw new Error(`the toolbar is still shown on the Preview tab (${toolbar})`);
      },
      async close(page) {
        await page.locator("#journal-write-tab").click();
        await page.locator("#journal-write-tab[aria-selected=true]").waitFor({ timeout: 5000 });
      },
    },
  ],
  "/components/navigation": [
    {
      // The navbar collapses by container width: the 380px demo frame shows the menu button at
      // any viewport size.
      name: "mobile nav drawer open",
      async open(page) {
        await page.getByRole("button", { name: "Open menu" }).first().click();
        await page.locator('button[aria-label="Open menu"][aria-expanded=true]').first().waitFor({ timeout: 5000 });
        await page.locator("#demoNav-dialog[open]").waitFor({ timeout: 5000 });
      },
      async close(page) { await page.keyboard.press("Escape"); await page.locator("#demoNav-dialog").waitFor({ state: "hidden", timeout: 5000 }); },
    },
  ],
};

// Behaviour checks beyond axe, keyed by page. Each returns a list of failure messages; any
// failure fails the run (exit 1) like a WCAG violation.
const CHECKS = {
  "/components/code": [
    {
      // Line numbers are generated content: selecting the whole code and reading the selection
      // gives exactly the source, with no numbers (and no doubled newlines).
      name: "line numbers are not selected",
      async run(page) {
        // The "Show source" blocks are in closed <details>, whose content is not rendered: open them.
        await page.evaluate(() => document.querySelectorAll("details.ui-details").forEach((d) => { d.open = true; }));
        await page.waitForTimeout(400);
        const r = await page.evaluate(() => [...document.querySelectorAll("figure.ui-code-lines")].map((f) => {
          const pre = f.querySelector("pre");
          const sel = getSelection();
          sel.removeAllRanges();
          const range = document.createRange();
          range.selectNodeContents(pre.querySelector("code"));
          sel.addRange(range);
          const got = sel.toString().replace(/\r/g, "");
          sel.removeAllRanges();
          const want = f.querySelector("template")?.content.textContent ?? pre.textContent;
          const before = getComputedStyle(pre.querySelector(".line"), "::before");
          return { got, want, select: before.userSelect, counter: before.content };
        }));
        if (!r.length) return ["no figure.ui-code-lines on the page"];
        return r.flatMap((x, i) => [
          ...(x.got === x.want ? [] : [`block ${i}: selection differs from the source: ${JSON.stringify(x.got.slice(0, 60))} vs ${JSON.stringify(x.want.slice(0, 60))}`]),
          ...(/\d/.test(x.got.split("\n")[0].slice(0, 3)) && !/^\d/.test(x.want) ? [`block ${i}: the selection starts with a line number`] : []),
          ...(x.select === "none" ? [] : [`block ${i}: gutter user-select is ${x.select}`]),
        ]);
      },
    },
    {
      // Highlighted lines carry a tint and a bar; scrolling blocks scroll, wrapping blocks do not;
      // the page never scrolls sideways.
      name: "highlight, scroll and wrap",
      async run(page) {
        const r = await page.evaluate(() => {
          const hl = [...document.querySelectorAll(".line[data-hl]")].map((l) => getComputedStyle(l));
          const pres = [...document.querySelectorAll("figure.ui-code pre")].map((p) => ({ wrap: !!p.closest(".ui-code-wrap"), over: p.scrollWidth > p.clientWidth, tab: p.tabIndex }));
          return { hl: hl.map((c) => [c.backgroundColor, c.boxShadow]), pres, page: document.documentElement.scrollWidth - document.documentElement.clientWidth };
        });
        const f = [];
        if (r.hl.length !== 4) f.push(`${r.hl.length} highlighted lines, want 4`);
        for (const [bg, bar] of r.hl) {
          if (/rgba?\(0, 0, 0, 0\)|transparent/.test(bg)) f.push("highlighted line has no tint");
          if (!/inset/.test(bar)) f.push("highlighted line has no bar");
        }
        if (r.pres.some((p) => p.tab !== 0)) f.push("a <pre> is not focusable");
        if (r.pres.some((p) => p.wrap && p.over)) f.push("a wrapping block scrolls horizontally");
        if (!r.pres.some((p) => !p.wrap && p.over)) f.push("no scrolling block overflows");
        if (r.page > 0) f.push(`the page scrolls sideways by ${r.page}px`);
        return f;
      },
    },
    {
      // The copy icon swap fades in 150ms, and the live region exists, empty, before any click.
      name: "copy button fade",
      async run(page) {
        const r = await page.evaluate(() => {
          const b = document.querySelector("figure.ui-code button[aria-label='Copy code']");
          const [a, c] = b.querySelectorAll("svg");
          return { d: [a, c].map((i) => getComputedStyle(i).transitionDuration), live: b.parentElement.querySelector("[aria-live]").textContent, checkOpacity: getComputedStyle(c).opacity };
        });
        const f = [];
        if (r.d.some((d) => d !== "0.15s")) f.push(`icon transition is ${r.d}, want 0.15s`);
        if (r.live !== "") f.push(`live region starts as ${JSON.stringify(r.live)}`);
        if (r.checkOpacity !== "0") f.push(`check icon opacity ${r.checkOpacity} before a click`);
        return f;
      },
    },
  ],
  "/components/markdown": [
    {
      // A toolbar click goes through <ui-mdeditor>: the selection is wrapped, an input event
      // reaches data-bind (the counter, which reads the signal, follows), the word stays
      // selected, and Ctrl+Z undoes the edit (insertText keeps the undo stack).
      name: "toolbar bold updates the textarea and the signal",
      async run(page) {
        const before = await page.evaluate(() => {
          const ta = document.querySelector("#journal");
          window.__inputs = 0;
          ta.addEventListener("input", () => window.__inputs++);
          const i = ta.value.indexOf("Woke");
          ta.focus();
          ta.setSelectionRange(i, i + 4);
          return { i, value: ta.value, count: document.querySelector("#journal-count").textContent };
        });
        await page.locator(".ui-mdeditor:has(#journal) [data-ui-md-action=bold]").click();
        await page.waitForTimeout(150);
        const after = await page.evaluate(() => {
          const ta = document.querySelector("#journal");
          return { value: ta.value, sel: ta.value.slice(ta.selectionStart, ta.selectionEnd), inputs: window.__inputs, count: document.querySelector("#journal-count").textContent };
        });
        const f = [];
        if (after.value !== before.value.slice(0, before.i) + "**Woke**" + before.value.slice(before.i + 4)) f.push(`value after Bold: ${JSON.stringify(after.value.slice(before.i - 2, before.i + 14))}`);
        if (after.sel !== "Woke") f.push(`selection after Bold is ${JSON.stringify(after.sel)}, want "Woke"`);
        if (after.inputs < 1) f.push("no input event after Bold");
        const chars = (s) => +(/(\d+) \/ /.exec(s) || [])[1];
        if (chars(after.count) !== chars(before.count) + 4) f.push(`counter did not follow the signal: ${before.count} -> ${after.count}`);
        await page.locator("#journal").press("Control+z");
        await page.waitForTimeout(100);
        if (await page.locator("#journal").inputValue() !== before.value) f.push("Ctrl+Z did not undo the Bold edit");
        return f;
      },
    },
    {
      // Roving tabindex: one tab stop, arrow keys and Home/End move it.
      name: "toolbar roving tabindex",
      async run(page) {
        const bar = page.locator(".ui-mdeditor:has(#note) [role=toolbar]");
        await bar.locator("[data-ui-md-action]").first().focus();
        await page.keyboard.press("ArrowRight");
        const second = await page.evaluate(() => document.activeElement.getAttribute("data-ui-md-action"));
        await page.keyboard.press("End");
        const last = await page.evaluate(() => document.activeElement.getAttribute("data-ui-md-action"));
        const stops = await bar.locator("[data-ui-md-action][tabindex='0']").count();
        await page.keyboard.press("Home");
        const first = await page.evaluate(() => document.activeElement.getAttribute("data-ui-md-action"));
        const f = [];
        if (second !== "italic" || last !== "task" || first !== "bold") f.push(`arrow keys moved focus to ${second}, ${last}, ${first}; want italic, task, bold`);
        if (stops !== 1) f.push(`${stops} toolbar buttons are tab stops, want 1`);
        return f;
      },
    },
    {
      // Enter in a list item continues the list; Enter on the empty item ends it. Ctrl+B with
      // nothing selected inserts a placeholder.
      name: "list continuation and shortcuts",
      async run(page) {
        const ta = page.locator("#note");
        await ta.fill("");
        await ta.focus();
        await page.keyboard.type("- [x] milk");
        await page.keyboard.press("Enter");
        const continued = await ta.inputValue();
        await page.keyboard.press("Enter");
        const ended = await ta.inputValue();
        await page.keyboard.press("Control+b");
        const bolded = await ta.inputValue();
        const f = [];
        if (continued !== "- [x] milk\n- [ ] ") f.push(`after Enter: ${JSON.stringify(continued)}`);
        if (ended !== "- [x] milk\n") f.push(`after a second Enter: ${JSON.stringify(ended)}`);
        if (bolded !== "- [x] milk\n**bold text**") f.push(`after Ctrl+B: ${JSON.stringify(bolded)}`);
        return f;
      },
    },
    {
      // The split editor fetches its preview on load (data-init) and lays out side by side.
      name: "split preview",
      async run(page) {
        await page.locator("#post-preview .ui-prose table").waitFor({ timeout: 5000 }).catch(() => {});
        const r = await page.evaluate(() => {
          const p = document.querySelector("#post-preview").closest("[role=region]"), ta = document.querySelector("#post");
          return { prose: !!p.querySelector(".ui-prose table"), side: Math.abs(p.getBoundingClientRect().top - ta.getBoundingClientRect().top) < 2 && p.getBoundingClientRect().left > ta.getBoundingClientRect().right - 2 };
        });
        return [...(r.prose ? [] : ["split preview did not render"]), ...(r.side ? [] : ["split preview is not beside the textarea"])];
      },
    },
  ],
  "/components/prose": [
    {
      // Fenced code is highlighted by the code package; task lists lose their bullets and draw
      // the checkbox; the page never scrolls sideways.
      name: "code, task list and width",
      async run(page) {
        const r = await page.evaluate(() => {
          const kw = document.querySelector(".ui-prose figure.ui-code .tok-keyword");
          const list = document.querySelector(".ui-prose ul.contains-task-list");
          const box = list && list.querySelector("input[type=checkbox]");
          return {
            kw: kw && getComputedStyle(kw).color, body: getComputedStyle(document.querySelector(".ui-prose p")).color,
            bullet: list && getComputedStyle(list).listStyleType, appearance: box && getComputedStyle(box).appearance,
            page: document.documentElement.scrollWidth - document.documentElement.clientWidth,
          };
        });
        const f = [];
        if (!r.kw || r.kw === r.body) f.push(`fenced code is not highlighted (${r.kw})`);
        if (r.bullet !== "none") f.push(`task list bullet is ${r.bullet}`);
        if (r.appearance !== "none") f.push(`task checkbox appearance is ${r.appearance}`);
        if (r.page > 0) f.push(`the page scrolls sideways by ${r.page}px`);
        return f;
      },
    },
  ],
  "/components/navigation": [
    {
      // Container queries: the narrow frame's navbar shows the menu button and hides the links;
      // the full-width one shows the links and hides the button.
      name: "navbar follows its container width",
      async run(page) {
        const r = await page.evaluate(() => [...document.querySelectorAll("nav[aria-label^='Demo']")].map((n) => ({
          label: n.getAttribute("aria-label"), width: Math.round(n.getBoundingClientRect().width),
          button: getComputedStyle(n.querySelector("button[aria-label='Open menu']")).display,
          links: getComputedStyle(n.querySelector("ul")).display,
        })));
        const f = [];
        const narrow = r.find((x) => x.label.includes("narrow")), wide = r.find((x) => x.label.includes("wide"));
        if (!narrow || !wide) return [`demo navbars not found: ${JSON.stringify(r)}`];
        if (narrow.button === "none" || narrow.links !== "none") f.push(`narrow navbar (${narrow.width}px): button ${narrow.button}, links ${narrow.links}`);
        if (wide.button !== "none" || wide.links === "none") f.push(`wide navbar (${wide.width}px): button ${wide.button}, links ${wide.links}`);
        return f;
      },
    },
  ],
  "/components/drag-and-drop": [
    {
      // A drop (mouse or the touch ui:drop event) must not show the Move here buttons, which
      // only belong to the pick-up flow, not even for one frame.
      name: "no Move here flash on drop",
      async run(page) {
        const seen = await page.evaluate(async () => {
          const zone = document.querySelector("[data-ui-dropzone]");
          const moves = () => [...zone.closest("#kanban").querySelectorAll("button")].filter((b) => /Move here/.test(b.textContent));
          const out = [];
          const rec = () => out.push(moves().filter((m) => getComputedStyle(m).display !== "none").length);
          // The payload names no card, so the server move is a no-op and the board is unchanged.
          zone.dispatchEvent(new CustomEvent("ui:drop", { bubbles: true, detail: { payload: "none" } }));
          rec();
          await new Promise((r) => requestAnimationFrame(() => { rec(); r(); }));
          await new Promise((r) => requestAnimationFrame(() => { rec(); r(); }));
          return out;
        });
        return seen.some((n) => n > 0) ? [`Move here buttons visible after a drop: ${JSON.stringify(seen)}`] : [];
      },
    },
  ],
  "/components/disclosure": [
    {
      // The tablist scrolls horizontally (overflow-x: auto, which makes overflow-y auto too), so
      // anything taller than the list, such as a tab with a negative bottom margin, shows a
      // vertical scrollbar.
      name: "tablist has no vertical overflow",
      async run(page) {
        const lists = await page.evaluate(() => [...document.querySelectorAll("[role=tablist]")].map((l) => ({
          id: l.querySelector("[role=tab]")?.id, scrollHeight: l.scrollHeight, clientHeight: l.clientHeight,
          underline: getComputedStyle(l.querySelector("[aria-selected=true]")).borderBottomWidth,
        })));
        if (!lists.length) return ["no [role=tablist] on the page"];
        return lists.flatMap((l) => [
          ...(l.scrollHeight === l.clientHeight ? [] : [`tablist of ${l.id}: scrollHeight ${l.scrollHeight} != clientHeight ${l.clientHeight}`]),
          ...(l.underline === "2px" ? [] : [`tablist of ${l.id}: active underline is ${l.underline}, want 2px`]),
        ]);
      },
    },
  ],
  "/components/overlays": [
    {
      // Initial focus, opened from the keyboard (so :focus-visible heuristics are at their
      // strictest): the drawer with a form focuses its first field; the Delete dialog its
      // autofocus Cancel; the share sheet (buttons only, no form control) the <dialog> itself,
      // with no outline painted on it or on the X button. Tab then reaches the X first.
      name: "dialog initial focus",
      async run(page) {
        const open = async (trigger, id) => {
          await page.getByRole("button", { name: trigger, exact: true }).focus();
          await page.keyboard.press("Enter");
          await page.locator(`#${id}-dialog[open]`).waitFor({ timeout: 5000 });
          await page.waitForTimeout(50);
          return page.evaluate((id) => {
            const a = document.activeElement, d = document.getElementById(id + "-dialog");
            const x = d.querySelector("button[aria-label=Close]");
            const cs = getComputedStyle(a);
            return {
              tag: a.tagName.toLowerCase(), id: a.id, text: a.textContent.trim().slice(0, 20), isDialog: a === d,
              outline: cs.outlineStyle === "none" || parseFloat(cs.outlineWidth) === 0 ? "none" : cs.outlineStyle + " " + cs.outlineWidth,
              xFocusVisible: x.matches(":focus-visible"),
            };
          }, id);
        };
        const close = async (id) => { await page.keyboard.press("Escape"); await page.locator(`#${id}-dialog`).waitFor({ state: "hidden", timeout: 5000 }); };
        const f = [];
        const drawer = await open("Edit profile", "profileDrawer");
        if (drawer.id !== "drawerName") f.push(`drawer with a form: focus on ${drawer.tag}#${drawer.id}, want the Name input`);
        await close("profileDrawer");
        const del = await open("Delete project", "confirmDelete");
        if (del.text !== "Cancel") f.push(`Delete dialog: focus on ${del.tag} "${del.text}", want Cancel`);
        await close("confirmDelete");
        const sheet = await open("Share", "shareSheet");
        if (!sheet.isDialog) f.push(`share sheet: focus on ${sheet.tag}#${sheet.id} "${sheet.text}", want the dialog`);
        if (sheet.outline !== "none") f.push(`share sheet: the focused dialog paints an outline (${sheet.outline})`);
        if (sheet.xFocusVisible) f.push("share sheet: the X button matches :focus-visible on open");
        await page.keyboard.press("Tab");
        const next = await page.evaluate(() => document.activeElement.getAttribute("aria-label"));
        if (next !== "Close") f.push(`Tab from the focused sheet went to ${next}, want the Close button`);
        await close("shareSheet");
        return f;
      },
    },
    {
      // The dialog fades in and out (opacity via @starting-style and transition-discrete) in 200ms.
      name: "dialog open/close transitions",
      async run(page) {
        const r = await page.evaluate(async () => {
          const d = document.querySelector("dialog");
          const op = () => +getComputedStyle(d).opacity;
          const frame = () => new Promise((res) => requestAnimationFrame(() => res()));
          const wait = (ms) => new Promise((res) => setTimeout(res, ms));
          d.showModal(); await frame();
          const opened0 = op(); await wait(250);
          const opened = op();
          d.close(); await frame(); await wait(40);
          const closing = [op(), getComputedStyle(d).display]; await wait(250);
          return { opened0, opened, closing, closed: getComputedStyle(d).display };
        });
        const f = [];
        if (!(r.opened0 < 1)) f.push(`opacity right after showModal() is ${r.opened0}, want < 1`);
        if (r.opened !== 1) f.push(`opacity 250ms after showModal() is ${r.opened}, want 1`);
        if (!(r.closing[0] < 1 && r.closing[0] > 0 && r.closing[1] !== "none")) f.push(`no fade-out after close(): ${JSON.stringify(r.closing)}`);
        if (r.closed !== "none") f.push(`dialog still displayed 290ms after close(): ${r.closed}`);
        return f;
      },
    },
  ],
};

// axe-core 4.11 cannot parse the missing components Chromium serializes for achromatic Tailwind v4
// colors ("oklch(0.985 0 none)") and silently skips the whole color-contrast rule on such a page.
// Report "none" as 0 (which is what it means for rendering) so contrast is really measured.
const FIX_NONE = () => {
  const fix = (v) => (typeof v === "string" && /^(oklch|oklab|lch|lab|hsl|hwb)\(/.test(v) ? v.replace(/\bnone\b/g, "0") : v);
  const gpv = CSSStyleDeclaration.prototype.getPropertyValue;
  CSSStyleDeclaration.prototype.getPropertyValue = function (name) { return fix(gpv.call(this, name)); };
  for (const prop of ["color", "backgroundColor", "borderTopColor", "borderRightColor", "borderBottomColor", "borderLeftColor"]) {
    const d = Object.getOwnPropertyDescriptor(CSSStyleDeclaration.prototype, prop);
    if (d && d.get) Object.defineProperty(CSSStyleDeclaration.prototype, prop, { ...d, get() { return fix(d.get.call(this)); } });
  }
};

async function routeCDN(page) {
  const files = [[/@tailwindcss\/browser/, "tailwind.js"], [/datastar/, "datastar.js"]];
  await page.route(/^https:\/\/(cdn\.jsdelivr\.net|cdnjs\.cloudflare\.com)\//, (route) => {
    const u = route.request().url();
    const hit = files.find(([re]) => re.test(u));
    if (hit) return route.fulfill({ path: `${CDN_DIR}/${hit[1]}`, contentType: "application/javascript" });
    return route.fulfill({ status: 200, contentType: u.endsWith(".css") || u.includes("css") ? "text/css" : "text/plain", body: "" });
  });
}

// Poll until Tailwind has generated and applied styles. Fails loudly if it never does.
async function waitForStyles(page, scheme, url) {
  const ok = await page
    .waitForFunction(
      () => {
        const bg = getComputedStyle(document.body).backgroundColor;
        const main = document.querySelector("main");
        return bg && bg !== "rgba(0, 0, 0, 0)" && bg !== "transparent" && main &&
          getComputedStyle(main).flexGrow === "1" && document.styleSheets.length > 0;
      },
      null,
      { timeout: 20000 },
    )
    .then(() => true, () => false);
  if (!ok) {
    throw new Error(`Tailwind did not apply on ${url} (${scheme}): body background is transparent. ` +
      `Is the Tailwind CDN (cdn.jsdelivr.net) reachable? Refusing to scan unstyled pages.`);
  }
  const err = await schemeProblem(page, scheme);
  if (err) throw new Error(`${url}: ${err}`);
  await page.evaluate(() => document.fonts.ready);
}

// Chromium reports oklch(); a canvas pixel gives plain sRGB to take the lightness from. Returns a
// description of what is wrong when the page is not in `scheme`, else "". Both the .dark class
// (what the dark: variants follow) and the painted background must agree with the scheme.
async function schemeProblem(page, scheme) {
  const [bg, lum, dark] = await page.evaluate(() => {
    const css = getComputedStyle(document.body).backgroundColor;
    const c = document.createElement("canvas").getContext("2d", { willReadFrequently: true });
    c.fillStyle = css;
    c.fillRect(0, 0, 1, 1);
    const [r, g, b] = c.getImageData(0, 0, 1, 1).data;
    return [css, (r + g + b) / 3, document.documentElement.classList.contains("dark")];
  });
  if (dark !== (scheme === "dark")) return `${scheme} scheme not applied: <html> ${dark ? "has" : "lacks"} the .dark class.`;
  if (scheme === "dark" ? lum > 100 : lum < 150) return `${scheme} scheme not applied (body background ${bg}).`;
  return "";
}

// Waits (up to 5s, for the transition-free class flip and style recalculation) until the page is
// in `scheme`; throws with the reason otherwise.
async function waitForScheme(page, scheme, when) {
  let err = "";
  for (let i = 0; i < 50; i++) {
    err = await schemeProblem(page, scheme);
    if (!err) return;
    await page.waitForTimeout(100);
  }
  throw new Error(`${when}: ${err}`);
}

function simplify(v, ctx) {
  return v.nodes.map((n) => {
    const d = (n.any.concat(n.all, n.none).find((c) => c.data && c.data.contrastRatio) || {}).data || {};
    return {
      rule: v.id, impact: v.impact || n.impact, tags: v.tags, help: v.help, helpUrl: v.helpUrl,
      page: ctx.page, scheme: ctx.scheme, state: ctx.state,
      target: n.target.join(" >> "), html: n.html,
      contrast: d.contrastRatio, expected: d.expectedContrastRatio, fg: d.fgColor, bg: d.bgColor, fontSize: d.fontSize,
      summary: (n.failureSummary || "").replace(/\s+/g, " ").trim(),
    };
  });
}

const findings = [];
const checkFailures = [];
let scans = 0;

async function scan(page, ctx) {
  const res = await new AxeBuilder({ page }).withTags(TAGS).analyze();
  scans++;
  // A rule that errored (rather than passing) means the scan proved nothing: never treat it as a pass.
  const broken = res.incomplete.filter((r) => r.error || r.nodes.some((n) => n.none.some((c) => c.id === "error-occurred")));
  if (broken.length) {
    const m = broken[0].nodes[0] && broken[0].nodes[0].none.find((c) => c.data && c.data.message);
    throw new Error(`axe rule ${broken.map((r) => r.id).join(", ")} errored on ${ctx.page} (${ctx.scheme}, ${ctx.state}): ${m ? m.data.message : "unknown"}`);
  }
  for (const v of res.violations) findings.push(...simplify(v, ctx));
  // axe 4.x dropped duplicate-id (WCAG 4.1.1 is obsolete) but duplicate ids still break
  // aria-labelledby/aria-controls/label[for], so check them here.
  const dups = await page.evaluate(() => {
    const seen = new Map();
    for (const el of document.querySelectorAll("[id]")) seen.set(el.id, [...(seen.get(el.id) || []), el]);
    return [...seen].filter(([, els]) => els.length > 1)
      .map(([id, els]) => ({ id, n: els.length, html: els[1].outerHTML.slice(0, 200) }));
  });
  for (const d of dups) {
    findings.push({ rule: "duplicate-id", impact: "serious", tags: ["wcag2a"], help: "id attribute values must be unique",
      page: ctx.page, scheme: ctx.scheme, state: ctx.state, target: `#${d.id} (x${d.n})`, html: d.html, summary: "" });
  }
  // "incomplete" is what axe could not decide (often contrast over gradients); surface the count only.
  return res.incomplete.length;
}

const browser = await chromium.launch(process.env.CHROMIUM_PATH ? { executablePath: process.env.CHROMIUM_PATH } : {});
let incompleteTotal = 0;

// No flash of the wrong scheme: with a choice stored that differs from the OS setting (and with
// none, following the OS), the inline bootstrap has set <html class="dark"> by the time the
// document has been parsed, i.e. before anything could paint. Checked at domcontentloaded, before
// Tailwind's styles are looked at.
async function checkNoFlash() {
  const cases = [
    { os: "light", stored: "dark", dark: true },
    { os: "dark", stored: "light", dark: false },
    { os: "dark", stored: null, dark: true },
    { os: "light", stored: null, dark: false },
  ];
  for (const c of cases) {
    const context = await browser.newContext({ colorScheme: c.os });
    const page = await context.newPage();
    if (CDN_DIR) await routeCDN(page);
    if (c.stored) await page.addInitScript((v) => localStorage.setItem("color-scheme", v), c.stored);
    await page.goto(BASE + "/components/buttons", { waitUntil: "domcontentloaded" });
    const dark = await page.evaluate(() => document.documentElement.classList.contains("dark"));
    const label = `os ${c.os}, stored ${c.stored || "none"}`;
    if (dark !== c.dark) checkFailures.push(`no-flash (${label}): <html> ${dark ? "has" : "lacks"} .dark at domcontentloaded, want ${c.dark ? "dark" : "light"}`);
    process.stdout.write(`checked no-flash (${label}): ${dark === c.dark ? "ok" : "FAIL"}\n`);
    await context.close();
  }
}

try {
  await checkNoFlash();
  for (const scheme of SCHEMES) {
    const context = await browser.newContext({ colorScheme: scheme, viewport: { width: 1280, height: 900 }, permissions: ["clipboard-read", "clipboard-write"] });
    for (const path of PAGES) {
      const page = await context.newPage();
      await page.addInitScript(FIX_NONE);
      if (CDN_DIR) await routeCDN(page);
      const url = BASE + path;
      page.on("pageerror", (e) => console.error(`  [pageerror] ${path}: ${e.message}`));
      // Console errors fail the run, except failed loads of the external CDNs (blocked or stubbed
      // in sandboxes), which say nothing about the showcase.
      const consoleErrors = [];
      page.on("console", (m) => { if (m.type() === "error") consoleErrors.push(m.text()); });
      page.on("requestfailed", (r) => {
        if (!/^https:\/\/(cdn\.jsdelivr\.net|cdnjs\.cloudflare\.com)\//.test(r.url())) consoleErrors.push(`request failed: ${r.url()}`);
      });
      const resp = await page.goto(url, { waitUntil: "load" });
      if (!resp || !resp.ok()) throw new Error(`${url}: HTTP ${resp && resp.status()}`);
      await page.emulateMedia({ colorScheme: scheme });
      await waitForStyles(page, scheme, url);
      for (const m of consoleErrors.filter((t) => !/ERR_BLOCKED_BY_CLIENT|ERR_FAILED|ERR_NAME_NOT_RESOLVED|ERR_TUNNEL|ERR_PROXY/.test(t))) {
        checkFailures.push(`${path} (${scheme}) [console]: ${m}`);
      }
      process.stdout.write(`checked ${scheme.padEnd(5)} ${path} [console errors]: ${consoleErrors.length ? consoleErrors.length + " (see failures if any are not external)" : "ok"}\n`);
      incompleteTotal += await scan(page, { page: path, scheme, state: "default" });
      process.stdout.write(`scanned ${scheme.padEnd(5)} ${path}\n`);

      for (const ch of CHECKS[path] || []) {
        const fails = await ch.run(page);
        for (const m of fails) checkFailures.push(`${path} (${scheme}) [${ch.name}]: ${m}`);
        process.stdout.write(`checked ${scheme.padEnd(5)} ${path} [${ch.name}]: ${fails.length ? "FAIL" : "ok"}\n`);
      }

      for (const st of STATES[path] || []) {
        if (st.viewport) await page.setViewportSize(st.viewport);
        try {
          await st.open(page, scheme);
        } catch (e) {
          throw new Error(`${url} (${scheme}): could not reach state "${st.name}": ${e.message.split("\n")[0]}`);
        }
        await page.waitForTimeout(350); // let transitions finish before measuring contrast
        incompleteTotal += await scan(page, { page: path, scheme, state: st.name });
        process.stdout.write(`scanned ${scheme.padEnd(5)} ${path} [${st.name}]\n`);
        if (st.close) await st.close(page, scheme);
        if (st.viewport) await page.setViewportSize({ width: 1280, height: 900 });
      }
      await page.close();
    }
    await context.close();
  }
} catch (e) {
  console.error(`\nSCAN ERROR: ${e.message}`);
  await browser.close();
  process.exit(2);
}
await browser.close();

if (jsonOut) writeFileSync(jsonOut, JSON.stringify(findings, null, 2));

// ---- report ----
const isWcag = (f) => f.tags.some((t) => WCAG.test(t));
const fails = findings.filter(isWcag);
const advice = findings.filter((f) => !isWcag(f));

function group(list) {
  const by = new Map();
  for (const f of list) {
    if (!by.has(f.rule)) by.set(f.rule, []);
    by.get(f.rule).push(f);
  }
  return [...by.entries()].sort((a, b) => b[1].length - a[1].length);
}
function print(title, list) {
  console.log(`\n=== ${title}: ${list.length} node(s) in ${group(list).length} rule(s) ===`);
  for (const [rule, items] of group(list)) {
    const where = new Map();
    for (const f of items) {
      const k = `${f.page} (${f.scheme}${f.state === "default" ? "" : ", " + f.state})`;
      where.set(k, (where.get(k) || 0) + 1);
    }
    console.log(`\n${rule}  [${items[0].impact}]  x${items.length}  ${items[0].help}`);
    const seen = new Set();
    for (const f of items) {
      const k = f.target;
      if (seen.has(k) || seen.size >= 4) continue;
      seen.add(k);
      const c = f.contrast ? `  ratio ${f.contrast} (need ${f.expected}) ${f.fg} on ${f.bg}` : "";
      console.log(`    ${f.target}${c}`);
    }
    if (items.length > seen.size) console.log(`    ... and ${items.length - seen.size} more`);
    console.log(`    pages: ${[...where].map(([k, n]) => `${k} x${n}`).slice(0, 8).join("; ")}${where.size > 8 ? "; ..." : ""}`);
  }
}

console.log(`\n${scans} scans, ${PAGES.length} pages x ${SCHEMES.length} schemes + interactive states. axe "incomplete" (needs manual review): ${incompleteTotal}`);
if (advice.length) print("best-practice (advisory, not failing)", advice);
if (checkFailures.length) {
  console.log(`\n=== behaviour checks: ${checkFailures.length} failure(s) ===`);
  for (const m of checkFailures) console.log(`    ${m}`);
}
if (fails.length) print("WCAG violations", fails);
if (fails.length || checkFailures.length) {
  console.log(`\nFAIL: ${fails.length} WCAG violation node(s), ${checkFailures.length} behaviour check failure(s).`);
  process.exit(1);
}
console.log("\nPASS: no WCAG A/AA violations; behaviour checks passed.");
