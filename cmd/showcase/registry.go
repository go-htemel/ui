package main

import (
	"embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"reflect"
	"runtime"
	"strings"

	"github.com/go-htemel/htemel"
)

// The example sources are embedded and parsed at startup, so the code shown next to each example
// is always the code that rendered it: there is no generated file or string copy to keep in sync.
//
//go:embed examples_*.go
var exampleFiles embed.FS

// Example is one live component demo with the Go source that builds it.
type Example struct {
	Title  string
	Source string
	// Snippets are more code shown under Source, such as the CSS behind an example.
	Snippets []Snippet
	// Node builds the example. It is called on every request, so examples can show live state.
	Node func() htemel.FlowContent
}

// Snippet is a titled block of code in some language.
type Snippet struct{ Lang, Title, Src string }

// Group is one showcase page.
type Group struct {
	Slug, Title, Intro string
	Examples           []Example
}

// sources maps the name of each function to the source of that function in examples_*.go.
var sources = mustParseSources(exampleFiles)

// ex builds an Example from an example function, looking up its source by name.
func ex(title string, fn func() htemel.FlowContent) Example {
	name := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	name = name[strings.LastIndexByte(name, '.')+1:] // "main.f", or the import path in tests
	src, ok := sources[name]
	if !ok {
		panic(fmt.Sprintf("showcase: no source for example %s; example functions must be top-level functions in examples_*.go", name))
	}
	return Example{Title: title, Source: src, Node: fn}
}

func mustParseSources(fsys fs.FS) map[string]string {
	out, err := parseSources(fsys)
	if err != nil {
		panic(err)
	}
	return out
}

func parseSources(fsys fs.FS) (map[string]string, error) {
	files, err := fs.Glob(fsys, "examples_*.go")
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, name := range files {
		src, err := fs.ReadFile(fsys, name)
		if err != nil {
			return nil, err
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, name, src, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			return nil, err
		}
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Recv != nil {
				continue
			}
			// Slice the original text rather than printing the AST, so comments inside the body stay.
			start, end := fset.Position(fd.Pos()).Offset, fset.Position(fd.End()).Offset
			out[fd.Name.Name] = strings.ReplaceAll(string(src[start:end]), "\t", "    ")
		}
	}
	return out, nil
}
