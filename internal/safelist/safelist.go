// Package safelist lists the Tailwind classes the ui components use, so that a Tailwind build
// that does not scan the ui sources (they live in the Go module cache) still generates them.
package safelist

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// classPattern matches a Tailwind class: optional variant prefixes (dark:, hover:, *:,
// has-checked:, ...), an optional ! or -, then a lowercase utility made of hyphenated segments,
// which may carry an opacity (/10), a fraction or an arbitrary value ([0.625rem]).
var classPattern = regexp.MustCompile(`^(?:[a-z0-9*&>\[\]=_-]+:)*!?-?[a-z][a-z0-9]*(?:-[a-z0-9.\[\]%#(),]+)*(?:/[0-9.]+)?$`)

// classCalls are the methods and functions whose string arguments are class lists. A call
// through the datastar package is handled separately: only its name argument is a class.
var classCalls = map[string]bool{"Class": true, "AddClass": true, "Classes": true, "ToggleClass": true}

// classFuncs are the functions whose return values are class lists (BodyClasses).
var classFuncs = map[string]bool{"BodyClasses": true}

// quotedKey matches a quoted key of a JavaScript object literal, as in {'rotate-180': open}.
var quotedKey = regexp.MustCompile(`['"]([^'"]+)['"]\s*:`)

// Classes returns the sorted, de-duplicated classes the non-test Go files of fsys's root
// directory can emit. A class is a class-shaped token of a string that
// reaches a class-bearing call: an argument of .Class, .AddClass, .ToggleClass, Classes, or the
// name of datastar.Class (and the keys of datastar.Classes). Identifiers in those arguments are
// followed to the constants, variables, map and slice literals, assignments and call-site
// arguments that supply them, so strings that never reach such a call (JavaScript expressions,
// HTTP headers, SVG paths) contribute nothing. The return values of BodyClasses count too.
func Classes(fsys fs.FS) ([]string, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, err
	}
	var files []*ast.File
	fset := token.NewFileSet()
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := fs.ReadFile(fsys, name)
		if err != nil {
			return nil, err
		}
		f, err := parser.ParseFile(fset, name, src, parser.SkipObjectResolution)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	c := newCollector(files)
	c.run(files)

	seen := map[string]bool{}
	for s := range c.strs {
		for _, tok := range strings.Fields(s) {
			if classPattern.MatchString(tok) {
				seen[tok] = true
			}
		}
	}
	out := make([]string, 0, len(seen))
	for c := range seen {
		out = append(out, c)
	}
	slices.Sort(out)
	return out, nil
}

// collector gathers the strings that reach class-bearing calls. Identifiers are resolved by
// name, ignoring scope, which over-approximates but never misses a flow.
type collector struct {
	decls   map[string][]ast.Expr // identifier -> expressions assigned to it
	funcs   map[string]*ast.FuncDecl
	visited map[string]bool
	strs    map[string]bool
}

func newCollector(files []*ast.File) *collector {
	c := &collector{
		decls:   map[string][]ast.Expr{},
		funcs:   map[string]*ast.FuncDecl{},
		visited: map[string]bool{},
		strs:    map[string]bool{},
	}
	for _, f := range files {
		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok && fd.Body != nil {
				c.funcs[fd.Name.Name] = fd
			}
		}
		ast.Inspect(f, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.ValueSpec:
				for i, id := range n.Names {
					if i < len(n.Values) {
						c.decls[id.Name] = append(c.decls[id.Name], n.Values[i])
					}
				}
			case *ast.AssignStmt:
				for i, l := range n.Lhs {
					if id, ok := l.(*ast.Ident); ok && len(n.Lhs) == len(n.Rhs) {
						c.decls[id.Name] = append(c.decls[id.Name], n.Rhs[i])
					}
				}
			}
			return true
		})
	}
	// Call-site arguments flow into the parameters of the functions they call.
	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			fd := c.funcs[calleeName(call.Fun)]
			if fd == nil {
				return true
			}
			i := 0
			for _, field := range fd.Type.Params.List {
				for _, id := range field.Names {
					if _, variadic := field.Type.(*ast.Ellipsis); !variadic && i < len(call.Args) {
						c.decls[id.Name] = append(c.decls[id.Name], call.Args[i])
					}
					i++
				}
			}
			return true
		})
	}
	return c
}

func calleeName(fun ast.Expr) string {
	switch f := fun.(type) {
	case *ast.Ident:
		return f.Name
	case *ast.SelectorExpr:
		return f.Sel.Name
	case *ast.IndexExpr: // generic instantiation
		return calleeName(f.X)
	case *ast.IndexListExpr:
		return calleeName(f.X)
	}
	return ""
}

// run collects the arguments of every class-bearing call in files and the results of the
// classFuncs.
func (c *collector) run(files []*ast.File) {
	for name := range classFuncs {
		if fd := c.funcs[name]; fd != nil {
			c.collectReturns(fd)
		}
	}
	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			name := calleeName(call.Fun)
			if !classCalls[name] {
				return true
			}
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if x, ok := sel.X.(*ast.Ident); ok && x.Name == "datastar" {
					c.datastarCall(name, call)
					return true
				}
			}
			for _, a := range call.Args {
				c.collect(a)
			}
			return true
		})
	}
}

// datastarCall collects the class name of datastar.Class and the keys of datastar.Classes.
func (c *collector) datastarCall(name string, call *ast.CallExpr) {
	if len(call.Args) == 0 {
		return
	}
	if name == "Class" {
		c.collect(call.Args[0])
		return
	}
	sub := &collector{decls: c.decls, funcs: c.funcs, visited: map[string]bool{}, strs: map[string]bool{}}
	for _, a := range call.Args {
		sub.collect(a)
	}
	for s := range sub.strs {
		for _, m := range quotedKey.FindAllStringSubmatch(s, -1) {
			c.strs[m[1]] = true
		}
	}
}

// collect adds the strings that expression e can evaluate to.
func (c *collector) collect(e ast.Expr) {
	switch e := e.(type) {
	case *ast.BasicLit:
		if e.Kind == token.STRING {
			if s, err := strconv.Unquote(e.Value); err == nil {
				c.strs[s] = true
			}
		}
	case *ast.ParenExpr:
		c.collect(e.X)
	case *ast.BinaryExpr:
		c.collect(e.X)
		c.collect(e.Y)
	case *ast.IndexExpr:
		c.collect(e.X)
	case *ast.Ident:
		if c.visited[e.Name] {
			return
		}
		c.visited[e.Name] = true
		for _, v := range c.decls[e.Name] {
			c.collect(v)
		}
		delete(c.visited, e.Name)
	case *ast.CompositeLit:
		for _, elt := range e.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				elt = kv.Value
			}
			c.collect(elt)
		}
	case *ast.CallExpr:
		switch name := calleeName(e.Fun); {
		case classCalls[name]:
			for _, a := range e.Args {
				c.collect(a)
			}
		case name == "pick" && len(e.Args) > 0:
			c.collect(e.Args[0])
		default:
			if fd := c.funcs[name]; fd != nil {
				c.collectReturns(fd)
			}
		}
	}
}

// collectReturns collects the results of fd's return statements.
func (c *collector) collectReturns(fd *ast.FuncDecl) {
	if c.visited["func "+fd.Name.Name] {
		return
	}
	c.visited["func "+fd.Name.Name] = true
	defer delete(c.visited, "func "+fd.Name.Name)
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.FuncLit:
			return false
		case *ast.ReturnStmt:
			for _, r := range n.Results {
				c.collect(r)
			}
		}
		return true
	})
}

// Header is the first line of a generated safelist file.
const Header = "# Tailwind classes used by github.com/go-htemel/ui. Generated by task tailwind-safelist; do not edit."

// File renders classes as a safelist file: Header, then one class per line.
func File(classes []string) string {
	return Header + "\n" + strings.Join(classes, "\n") + "\n"
}
