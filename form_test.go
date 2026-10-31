package ui_test

import (
	"strings"
	"testing"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

var (
	_ ui.FormControl         = ui.Input(html.InputTypeText, "x")
	_ ui.FormControl         = ui.Textarea("x", 3)
	_ ui.FormControl         = ui.Select("x")
	_ htemel.PhrasingContent = ui.Checkbox("x")
	_ htemel.PhrasingContent = ui.Radio("n", "v", "x")
	_ htemel.PhrasingContent = ui.Switch("x")
	_ htemel.FlowContent     = ui.RadioGroup("x")
	_ htemel.FlowContent     = ui.Field("id", "x", ui.Input(html.InputTypeText, "x"))
	_ html.SelectChild       = ui.Choice("v", "x")
)

func init() {
	addGolden(
		golden{"input", ui.Input(html.InputTypeEmail, "email").Placeholder("you@example.com").Attrs(datastar.Bind("email"))},
		golden{"textarea", ui.Textarea("notes", 4)},
		golden{"select", ui.Select("plan", ui.Choice("free", "Free"), ui.Choice("pro", "Pro").Selected(true))},
		golden{"field-help", ui.Field("email", "Email", ui.Input(html.InputTypeEmail, "email"), ui.HelpText("We never share it."))},
		golden{"field-error", ui.Field("name", "Name", ui.Input(html.InputTypeText, "name"), ui.ErrorText("Name is required"))},
		golden{"field-error-signal", ui.Field("user", "Username", ui.Input(html.InputTypeText, "user").Attrs(datastar.Bind("user")),
			ui.HelpText("Letters and digits."), ui.ErrorSignal("errors.user"), ui.ErrorText(""))},
		golden{"checkbox", ui.Checkbox("Remember me", datastar.Bind("remember"))},
		golden{"radio-group", ui.RadioGroup("Plan",
			ui.Radio("plan", "free", "Free", datastar.Bind("plan")),
			ui.Radio("plan", "pro", "Pro", datastar.Bind("plan")))},
		golden{"switch", ui.Switch("Email notifications", datastar.Bind("notify"))},
	)
}

func TestFieldDescribedBy(t *testing.T) {
	out, err := htemel.RenderString(ui.Field("e", "E", ui.Input(html.InputTypeText, "e"),
		ui.HelpText("h"), ui.ErrorText("bad"), ui.ErrorSignal("err")))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`<label for="e"`,
		`id="e"`,
		`aria-describedby="e-help e-error e-error-live"`,
		`aria-invalid="true"`,
		`data-attr:aria-invalid="!!$err"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in %s", want, out)
		}
	}
}
