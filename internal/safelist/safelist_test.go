package safelist

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

const root = "../.."

// TestCommittedSafelistIsCurrent fails when a class was added to or removed from the components
// without regenerating tailwind.safelist.txt.
func TestCommittedSafelistIsCurrent(t *testing.T) {
	classes, err := Classes(os.DirFS(root))
	if err != nil {
		t.Fatal(err)
	}
	committed, err := os.ReadFile(filepath.Join(root, "tailwind.safelist.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(committed) != File(classes) {
		t.Error("tailwind.safelist.txt is out of date; run task tailwind-safelist")
	}
}

var (
	classAttr = regexp.MustCompile(` class="([^"]*)"`)
	classKey  = regexp.MustCompile(`&#39;([^&]*)&#39;:`) // keys of data-class object literals
)

// TestGoldenClassesAreListed checks that every class in the rendered golden output, including
// the class names inside data-class expressions, is in the safelist.
func TestGoldenClassesAreListed(t *testing.T) {
	classes, err := Classes(os.DirFS(root))
	if err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(root, "testdata", "*.golden.html"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no golden files: %v", err)
	}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var found []string
		for _, m := range classAttr.FindAllStringSubmatch(string(b), -1) {
			found = append(found, strings.Fields(m[1])...)
		}
		for _, m := range classKey.FindAllStringSubmatch(string(b), -1) {
			found = append(found, strings.Fields(m[1])...)
		}
		for _, c := range found {
			if strings.HasPrefix(c, "fa-") {
				continue // Font Awesome classes are the caller's, passed to ui.Icon
			}
			if _, ok := slices.BinarySearch(classes, c); !ok {
				t.Errorf("%s: class %q is not in the safelist", filepath.Base(f), c)
			}
		}
	}
}

// TestTokenSafelist checks that the list is written in design tokens: the semantic color and
// radius utilities are there (including variant and side forms, and the BodyClasses classes),
// and no raw palette or radius class is.
func TestTokenSafelist(t *testing.T) {
	classes, err := Classes(os.DirFS(root))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"bg-primary-600", "rounded-control", "rounded-surface", "*:first:rounded-l-control",
		"font-sans", "dark:bg-neutral-950", "has-[dialog:modal]:overflow-hidden", "bg-danger-600"} {
		if _, ok := slices.BinarySearch(classes, want); !ok {
			t.Errorf("safelist lacks %q", want)
		}
	}
	raw := regexp.MustCompile(`-(sky|zinc|green|red|amber|blue)-\d|rounded(-[a-z]+)?-(sm|md|lg|xl)$`)
	for _, c := range classes {
		if raw.MatchString(c) {
			t.Errorf("safelist has raw class %q", c)
		}
	}
}

// TestOnlyClassBearingStrings checks that strings which never reach a class-bearing call (JS
// expressions, HTTP header values) contribute nothing, and that literals, referenced constants,
// map values and datastar class names do.
func TestOnlyClassBearingStrings(t *testing.T) {
	const src = `package x

import "github.com/starfederation/datastar-go/datastar"

const base = "flex gap-2"
const unused = "unused-class"

var sizes = map[int]string{1: "text-sm", 2: "text-lg"}

func a(w http.ResponseWriter, el El) {
	w.Header().Set("Cache-Control", "immutable")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	el.Attrs(datastar.On("toggle", "evt.newState === 'open' && close"))
	el.Class("px-2 py-1").AddClass(Classes(base, pick(sizes, 1, 2)))
	el.Attrs(datastar.Class("rotate-180", "open"))
}

const body = "font-sans antialiased"

func BodyClasses() string { return body }
`
	got, err := Classes(fstest.MapFS{"x.go": {Data: []byte(src)}})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"antialiased", "flex", "font-sans", "gap-2", "px-2", "py-1", "rotate-180", "text-lg", "text-sm"}
	if !slices.Equal(got, want) {
		t.Errorf("Classes = %q, want %q", got, want)
	}
}

// TestSubpackagesAreScanned: classes written in the code package reach the safelist, and a
// module without that directory still works.
func TestSubpackagesAreScanned(t *testing.T) {
	fsys := fstest.MapFS{
		"x.go":           {Data: []byte("package x\nfunc a(e El) { e.Class(\"flex\") }\n")},
		"code/y.go":      {Data: []byte("package code\nfunc b(e El) { e.Class(\"sr-only\") }\n")},
		"code/y_test.go": {Data: []byte("package code\nfunc c(e El) { e.Class(\"test-only\") }\n")},
	}
	got, err := Classes(fsys)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"flex", "sr-only"}; !slices.Equal(got, want) {
		t.Errorf("Classes = %q, want %q", got, want)
	}
	if _, err := Classes(fstest.MapFS{"x.go": fsys["x.go"]}); err != nil {
		t.Errorf("no code directory: %v", err)
	}
}

// TestCodeGoldenClassesAreListed is TestGoldenClassesAreListed for the code package's output.
func TestCodeGoldenClassesAreListed(t *testing.T) {
	classes, err := Classes(os.DirFS(root))
	if err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(root, "code", "testdata", "*.golden.html"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no code golden files: %v", err)
	}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var found []string
		for _, m := range classAttr.FindAllStringSubmatch(string(b), -1) {
			found = append(found, strings.Fields(m[1])...)
		}
		for _, m := range regexp.MustCompile(`data-class:([^=]+)=`).FindAllStringSubmatch(string(b), -1) {
			found = append(found, m[1])
		}
		for _, c := range found {
			// The tok-*, line and language-* classes are plain CSS classes from ui.css, not utilities.
			if strings.HasPrefix(c, "tok-") || strings.HasPrefix(c, "language-") || c == "line" {
				continue
			}
			if _, ok := slices.BinarySearch(classes, c); !ok {
				t.Errorf("%s: class %q is not in the safelist", filepath.Base(f), c)
			}
		}
	}
}
