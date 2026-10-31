package ui

import (
	"strings"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
)

// Row selection for lists and tables with checkboxes and a select-all box, bulk actions and a
// count. The state is one object signal of booleans, a property per row:
//
//	$sel = {p1: true, p2: false, ...}
//
// Declare it where the rows are (datastar.Signals(map[string]any{"sel": map[string]bool{}})); a
// row appears in it when its box is first ticked. The key of a row is a JavaScript identifier
// of your choosing, such as "p12" for post 12: letters, digits and underscores, not starting
// with a digit. The server reads the object back with the other signals and can patch
// {"sel": {"p12": false}} to untick a row.
//
//	rows:      ui.Checkbox("", ui.SelectRow("sel", "p12"), htemel.Attr("aria-label", "Select “Hello”"))
//	select all: ui.Checkbox("", ui.SelectAll("sel", []string{"p12", "p13"}), htemel.Attr("aria-label", "Select all"))
//	count:     datastar.Computed("selCount", ui.SelectedCount("sel"))
//	clear:     datastar.OnClick(ui.SelectClear("sel"))

// SelectRow returns the attribute that binds a row's Checkbox to the property key of the object
// signal signal ($signal.key).
func SelectRow(signal, key string) htemel.Attribute {
	return datastar.Bind(signal + "." + key)
}

// SelectAll returns the attributes of the select-all Checkbox for the rows keys of the page
// shown: ticking it ticks (or unticks) all of them; it follows the rows, checked while all of
// them are ticked, indeterminate (the dash) while only some are, and unchecked with none. The
// keys are in the expression, so when the page changes, patch the box (or its cell) with the
// new list. The box has no text: add an aria-label to it ("Select all posts on this page").
func SelectAll(signal string, keys []string) []htemel.Attribute {
	quoted := make([]string, len(keys))
	for i, k := range keys {
		quoted[i] = datastar.Quote(k)
	}
	list := "[" + strings.Join(quoted, ", ") + "]"
	sel := "$" + signal
	return []htemel.Attribute{
		datastar.Effect("const n = " + list + ".filter((k) => " + sel + "[k]).length; " +
			"el.checked = n > 0 && n === " + list + ".length; el.indeterminate = n > 0 && n < " + list + ".length"),
		datastar.OnChange("const on = el.checked; " + list + ".forEach((k) => " + sel + "[k] = on)"),
	}
}

// SelectedCount returns the expression that counts the ticked rows of signal. Use it in a
// computed signal, datastar.Computed("selCount", ui.SelectedCount("sel")), then show a bulk
// action bar with datastar.Show("$selCount > 0").
func SelectedCount(signal string) string {
	return "Object.values($" + signal + ").filter(Boolean).length"
}

// SelectClear returns the expression that unticks every row of signal.
func SelectClear(signal string) string {
	return "Object.keys($" + signal + ").forEach((k) => $" + signal + "[k] = false)"
}
