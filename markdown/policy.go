package markdown

import (
	"regexp"

	"github.com/microcosm-cc/bluemonday"
)

// Policy returns a new copy of the default sanitising policy: bluemonday's UGCPolicy (the
// common text, list, table, image and link elements; http, https, mailto and relative URLs
// only; rel="nofollow" added to links; no style attributes, scripts, iframes, forms, or event
// handlers) plus exactly what the Markdown renderer itself emits:
//
//   - the code blocks of the code package: <figure class="ui-code" aria-label>, <pre tabindex="0">,
//     <code class="language-*">, <span class="line"> and <span class="tok-*">
//   - GFM task lists: <ul class="contains-task-list">, <li class="task-list-item"> and
//     <input type="checkbox" disabled> (checked or not)
//   - table cell alignment (align="left|center|right" on th and td)
//   - footnotes: <div class="footnotes" role="doc-endnotes">, links with class footnote-ref /
//     footnote-backref and role doc-noteref / doc-backlink, ids on sup and li
//   - target on links (WithLinkTarget)
//
// Each call returns a fresh policy, so callers can extend it and pass it to WithPolicy.
func Policy() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	p.AllowAttrs("class").Matching(codeClasses).OnElements("figure", "span", "code")
	p.AllowAttrs("aria-label").Matching(bluemonday.Paragraph).OnElements("figure")
	p.AllowAttrs("tabindex").Matching(regexp.MustCompile(`^0$`)).OnElements("pre")

	p.AllowAttrs("class").Matching(regexp.MustCompile(`^contains-task-list$`)).OnElements("ul", "ol")
	p.AllowAttrs("class").Matching(regexp.MustCompile(`^task-list-item$`)).OnElements("li")
	p.AllowAttrs("type").Matching(regexp.MustCompile(`^checkbox$`)).OnElements("input")
	p.AllowAttrs("checked", "disabled").Matching(regexp.MustCompile(`^(|checked|disabled)$`)).OnElements("input")

	p.AllowAttrs("align").Matching(regexp.MustCompile(`^(left|center|right)$`)).OnElements("th", "td")

	p.AllowAttrs("class").Matching(regexp.MustCompile(`^footnotes$`)).OnElements("div")
	p.AllowAttrs("role").Matching(regexp.MustCompile(`^doc-endnotes$`)).OnElements("div")
	p.AllowAttrs("class").Matching(regexp.MustCompile(`^footnote-(ref|backref)$`)).OnElements("a")
	p.AllowAttrs("role").Matching(regexp.MustCompile(`^doc-(noteref|backlink)$`)).OnElements("a")

	p.AllowAttrs("target").Matching(regexp.MustCompile(`^(_blank|_self|_parent|_top)$`)).OnElements("a")
	return p
}

// codeClasses are the class lists of the code package's markup.
var codeClasses = regexp.MustCompile(`^(?:ui-code(?: ui-code-[a-z]+)*|line|tok-[a-z]+|language-[a-z0-9-]+)$`)
