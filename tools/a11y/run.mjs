// Accessibility scan of the showcase: every page, light and dark, plus interactive states
// (open dialog, drawer and bottom sheet, open dropdown, switched tab, expanded accordion, open
// mobile nav drawer of the narrow navbar, the open customizable select picker, the color scheme
// toggle switched to the other scheme). Every page must also load without console errors
// (requests to the blocked external CDNs excepted).
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
// the network, for sandboxes that block the CDNs. Font Awesome and Google Fonts are then stubbed out.
const CDN_DIR = process.env.A11Y_CDN_DIR || "";
const jsonOut = process.argv.includes("--json") ? process.argv[process.argv.indexOf("--json") + 1] : "";

const PAGES = ["/", ...["buttons", "display", "layout", "feedback", "forms", "overlays", "disclosure",
  "drag-and-drop", "data", "navigation", "status", "theming"].map((s) => `/components/${s}`)];
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
  await page.route(/^https:\/\/(cdn\.jsdelivr\.net|cdnjs\.cloudflare\.com|fonts\.googleapis\.com|fonts\.gstatic\.com)\//, (route) => {
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
    const context = await browser.newContext({ colorScheme: scheme, viewport: { width: 1280, height: 900 } });
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
        if (!/^https:\/\/(cdn\.jsdelivr\.net|cdnjs\.cloudflare\.com|fonts\.googleapis\.com|fonts\.gstatic\.com)\//.test(r.url())) consoleErrors.push(`request failed: ${r.url()}`);
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
