package ui_test

import (
	"strings"
	"testing"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

var (
	_ htemel.FlowContent     = ui.Progress(1, 2)
	_ htemel.FlowContent     = ui.ProgressSignal("p", 100)
	_ htemel.FlowContent     = ui.Skeleton(2)
	_ htemel.PhrasingContent = ui.Spinner(ui.Md)
	_ htemel.PhrasingContent = ui.Tooltip(ui.Button(ui.Ghost, "x"), "tip")
)

func init() {
	addGolden(
		golden{"progress", ui.Progress(40, 100, ui.ProgressLabel("Upload"))},
		golden{"progress-tone-size", ui.Progress(3, 4, ui.ProgressTone(ui.Success), ui.ProgressSize(ui.Lg))},
		golden{"progress-signal", ui.ProgressSignal("upload", 100, ui.ProgressLabel("Upload"))},
		golden{"skeleton", ui.Skeleton(3)},
		golden{"spinner", ui.Spinner(ui.Lg)},
		golden{"tooltip", ui.Tooltip(ui.IconButton(ui.Ghost, ui.IconTrash(), "Delete"), "Delete this row")},
	)
}

func TestProgressClamps(t *testing.T) {
	for _, tc := range []struct {
		value, max int
		now, width string
	}{
		{150, 100, `aria-valuenow="100"`, "width: 100%"},
		{-5, 100, `aria-valuenow="0"`, "width: 0%"},
		{1, 0, `aria-valuenow="1"`, "width: 1%"},
		{1, 3, `aria-valuenow="1"`, "width: 33.33"},
	} {
		s, err := htemel.RenderString(ui.Progress(tc.value, tc.max))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(s, tc.now) || !strings.Contains(s, tc.width) || !strings.Contains(s, `role="progressbar"`) {
			t.Errorf("Progress(%d, %d) = %s", tc.value, tc.max, s)
		}
	}
}

func TestTooltipDescribesTrigger(t *testing.T) {
	s, _ := htemel.RenderString(ui.Tooltip(html.Button().Text("x"), "tip"))
	i := strings.Index(s, `aria-describedby="`)
	if i < 0 {
		t.Fatalf("trigger lacks aria-describedby: %s", s)
	}
	id := strings.SplitN(s[i+len(`aria-describedby="`):], `"`, 2)[0]
	if !strings.Contains(s, `id="`+id+`" role="tooltip"`) && !strings.Contains(s, `role="tooltip"`) {
		t.Errorf("tooltip id %q not found: %s", id, s)
	}
}
