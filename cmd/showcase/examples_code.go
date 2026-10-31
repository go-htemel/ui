package main

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
	"github.com/go-htemel/ui/code"
)

func codeGo() htemel.FlowContent {
	// Line numbers count from WithStart, and so do the highlighted lines: 16, and 19 to 21.
	return code.Block("go", `// Save stores the form and tells the browser what to do next.
func (h *Handler) Save(w http.ResponseWriter, r *http.Request) {
	var form struct {
		Name  string `+"`json:\"name\"`"+`
		Email string `+"`json:\"email\"`"+`
	}
	if err := datastar.ReadSignals(r, &form); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sse := datastar.NewSSE(w, r)
	if err := h.store.Save(r.Context(), form.Name, form.Email); err != nil {
		sse.PatchElements(ui.Toast(ui.Danger, "Could not save", err.Error()))
		return
	}
	sse.PatchElements(ui.Toast(ui.Success, "Saved", "Your changes are stored."))
}`,
		code.WithTitle("handler.go"),
		code.WithLineNumbers(),
		code.WithStart(12),
		code.WithHighlight(16),
		code.WithHighlightRange(19, 21),
	)
}

func codeJSON() htemel.FlowContent {
	return code.Block("json", `{
  "name": "go-htemel/ui",
  "stable": false,
  "components": 41,
  "themes": ["default", "sky", "indigo", "emerald"],
  "tokens": {
    "primary": "indigo",
    "radius": null
  }
}`, code.WithTitle("ui.config.json"))
}

func codeBash() htemel.FlowContent {
	return code.Block("bash", `# Add the library and build the Tailwind safelist
go get github.com/go-htemel/ui@latest
go run github.com/go-htemel/ui/cmd/safelist > tailwind.safelist.txt

for page in buttons forms overlays; do
  curl -fsS "http://localhost:8080/components/$page" > /dev/null && echo "$page ok"
done`, code.WithLineNumbers())
}

func codeInline() htemel.FlowContent {
	return html.P(
		htemel.Text("Call "),
		code.Inline("go", `ui.Button(ui.Primary, "Save")`),
		htemel.Text(" for a button, style it with "),
		code.Inline("css", `--color-primary-600: #4f46e5;`),
		htemel.Text(", and run "),
		code.Inline("bash", `task a11y`),
		htemel.Text(" before you ship. A language the highlighter does not know, like "),
		code.Inline("nope", `<plain & simple>`),
		htemel.Text(", is shown as plain text."),
	).Class("max-w-prose text-sm/7 text-neutral-700 dark:text-neutral-300")
}

func codeWrap() htemel.FlowContent {
	const src = `curl -fsS -X POST "https://api.example.com/v1/projects/42/deployments" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"ref": "main", "environment": "production", "wait": true}'`
	// The same source twice, so each block gets its own id for the copy button's signal.
	return ui.Stack(ui.Md,
		html.P().Class("text-xs font-medium text-neutral-600 dark:text-neutral-400").Text("Scroll (default)"),
		code.Block("bash", src, code.WithID("scroll")),
		html.P().Class("text-xs font-medium text-neutral-600 dark:text-neutral-400").Text("WithWrap()"),
		code.Block("bash", src, code.WithWrap(), code.WithLineNumbers(), code.WithID("wrap")),
	)
}

func codeNoCopy() htemel.FlowContent {
	return ui.Stack(ui.Md,
		code.Block("sql", `SELECT name, count(*) AS orders
FROM customers
JOIN orders ON orders.customer_id = customers.id
WHERE orders.created_at >= '2025-01-01'
GROUP BY name
ORDER BY orders DESC;`, code.WithoutCopyButton()),
		code.Block("yaml", `# No toolbar at all without a title or a copy button
name: ci
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: go test -race ./...`, code.WithoutCopyButton()),
	)
}

func codeMaxHeight() htemel.FlowContent {
	// A capped height scrolls the code, not the toolbar: add the CSS variable with AddClass.
	return code.Block("html", `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>Dashboard</title>
    <link rel="stylesheet" href="/assets/app.css">
    <script type="module" src="/assets/datastar.js"></script>
  </head>
  <body class="font-sans">
    <main id="app" data-signals:count="0">
      <button data-on:click="$count++">Clicked <span data-text="$count"></span> times</button>
    </main>
  </body>
</html>`, code.WithLineNumbers(), code.WithTitle("index.html")).AddClass("[--ui-code-max-height:12rem]")
}
