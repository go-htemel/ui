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
	_ ui.CompositeControl    = ui.InputGroup(nil, ui.Input(html.InputTypeText, "x"), nil)
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
		golden{"field-input-group", ui.Field("site", "Website",
			ui.InputGroup(htemel.Text("https://"), ui.Input(html.InputTypeText, "site"), nil), ui.HelpText("Your homepage."))},
		golden{"checkbox", ui.Checkbox("Remember me", datastar.Bind("remember"))},
		golden{"radio-group", ui.RadioGroup("Plan",
			ui.Radio("plan", "free", "Free", datastar.Bind("plan")),
			ui.Radio("plan", "pro", "Pro", datastar.Bind("plan")))},
		golden{"switch", ui.Switch("Email notifications", datastar.Bind("notify"))},
		golden{"checkbox-disabled-checked", ui.Checkbox("Locked", htemel.Attr("disabled", ""), htemel.Attr("checked", ""))},
		golden{"input-file", ui.Input(html.InputTypeFile, "avatar").Accept("image/*")},
		golden{"input-color", ui.Input(html.InputTypeColor, "accent").Value("#4f46e5")},
		golden{"input-range", ui.Input(html.InputTypeRange, "volume")},
		golden{"input-number", ui.Input(html.InputTypeNumber, "qty")},
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
		`data-attr:aria-invalid="$err ? &#39;true&#39; : false"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in %s", want, out)
		}
	}
}

// TestFieldLabelsInputGroupControl checks that Field labels the input inside an InputGroup, not
// the group's wrapper <div> (a <label for> pointing at a div names nothing).
func TestFieldLabelsInputGroupControl(t *testing.T) {
	out := renderStr(t, ui.Field("site", "Website",
		ui.InputGroup(htemel.Text("https://"), ui.Input(html.InputTypeText, "site"), nil),
		ui.HelpText("h"), ui.ErrorText("bad"), ui.ErrorSignal("err")))
	_, input, ok := strings.Cut(out, "<input")
	input, _, _ = strings.Cut(input, ">")
	if !ok {
		t.Fatalf("no input in %s", out)
	}
	for _, want := range []string{`id="site"`, `aria-describedby="site-help site-error site-error-live"`, `aria-invalid="true"`, `data-attr:aria-invalid=`} {
		if !strings.Contains(input, want) {
			t.Errorf("input lacks %s: %s", want, input)
		}
	}
	if n := strings.Count(out, `id="site"`); n != 1 {
		t.Errorf("want one id=site, got %d: %s", n, out)
	}
	if strings.Contains(out, `<div id="site"`) || strings.Contains(out, `<div aria-describedby`) {
		t.Errorf("the wrapper got the control's attributes: %s", out)
	}
}
