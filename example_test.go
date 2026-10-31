package ui_test

import (
	"fmt"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

// profileCard is the README example.
func profileCard() htemel.FlowContent {
	return ui.Card(
		ui.CardHeader("Profile", ui.Badge(ui.Success, "Verified")),
		ui.CardBody(
			ui.Field("name", "Name", ui.Input(html.InputTypeText, "name").Attrs(datastar.Bind("name")),
				ui.ErrorSignal("errors.name")),
		),
		ui.CardFooter(
			ui.Button(ui.Primary, "Save", ui.WithLoading("saving")).
				Attrs(datastar.OnClick(datastar.Post("/profile"))),
		),
	)
}

func Example() {
	out, err := htemel.RenderString(profileCard())
	fmt.Println(err == nil, len(out) > 0)
	// Output: true true
}

func ExampleButton() {
	b := ui.Button(ui.Link, "Docs", ui.WithSize(ui.Sm)).Attrs(datastar.OnClick(datastar.Get("/docs")))
	out, _ := htemel.RenderString(b)
	fmt.Println(out)
	// Output: <button type="button" class="inline-flex items-center justify-center gap-x-1.5 rounded-control font-semibold whitespace-nowrap cursor-pointer transition-colors duration-150 focus-visible:outline-2 focus-visible:outline-offset-2 disabled:cursor-not-allowed disabled:opacity-50 text-primary-700 underline-offset-4 hover:text-primary-800 hover:underline focus-visible:outline-primary-700 dark:text-primary-400 dark:hover:text-primary-300 dark:focus-visible:outline-primary-500 text-sm" data-on:click="@get(&#39;/docs&#39;)">Docs</button>
}

func ExampleClasses() {
	fmt.Println(ui.Classes("  rounded-surface bg-primary-600 ", "", "dark:hover:bg-primary-500"))
	// Output: rounded-surface bg-primary-600 dark:hover:bg-primary-500
}
