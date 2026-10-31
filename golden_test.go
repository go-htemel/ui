package ui_test

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-htemel/htemel"
)

var update = flag.Bool("update", false, "rewrite testdata/*.golden.html from the current output")

// golden is one rendered component compared against testdata/<name>.golden.html.
type golden struct {
	name string
	node htemel.Node
}

// goldenCases is filled by the init functions of the *_test.go files, one group per component.
var goldenCases []golden

func addGolden(cases ...golden) { goldenCases = append(goldenCases, cases...) }

func TestGolden(t *testing.T) {
	seen := map[string]bool{}
	for _, tc := range goldenCases {
		if seen[tc.name] {
			t.Fatalf("duplicate golden case %q", tc.name)
		}
		seen[tc.name] = true
		t.Run(tc.name, func(t *testing.T) {
			got, err := htemel.RenderString(tc.node)
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			got += "\n"
			path := filepath.Join("testdata", tc.name+".golden.html")
			if *update {
				if err := os.MkdirAll("testdata", 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("%v (run go test -run TestGolden -update to create it)", err)
			}
			if got != string(want) {
				t.Errorf("output differs from %s (rerun with -update if the change is intended)\n got: %s\nwant: %s",
					path, strings.TrimSpace(got), strings.TrimSpace(string(want)))
			}
		})
	}
	if !*update {
		// Every golden file must belong to a case, so renamed cases do not leave stale files.
		files, _ := filepath.Glob(filepath.Join("testdata", "*.golden.html"))
		for _, f := range files {
			if name := strings.TrimSuffix(filepath.Base(f), ".golden.html"); !seen[name] {
				t.Errorf("stale golden file %s has no case", f)
			}
		}
	}
}
