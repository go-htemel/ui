package markdown_test

import (
	"fmt"

	"github.com/go-htemel/htemel"

	"github.com/go-htemel/ui/markdown"
)

func ExampleRender() {
	out, _ := htemel.RenderString(markdown.Render("# Hello\n\nSome **bold** text. <script>alert(1)</script>",
		markdown.WithHeadingOffset(1)))
	fmt.Println(out)
	// Output:
	// <div class="ui-prose"><h2 id="hello">Hello</h2>
	// <p>Some <strong>bold</strong> text. alert(1)</p>
	// </div>
}

func ExampleExcerpt() {
	fmt.Println(markdown.Excerpt("## Frost\n\nWoke early; the **first frost** of the year.", 30))
	// Output: Frost Woke early; the first…
}
