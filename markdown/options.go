package markdown

import (
	"github.com/microcosm-cc/bluemonday"

	"github.com/go-htemel/ui"
)

// Option configures Render, HTML, Plain and Excerpt (Plain and Excerpt read only
// WithHeadingOffset, which does not change their text, and ignore the rest).
type Option func(*config)

type config struct {
	unsafe        bool
	policy        *bluemonday.Policy
	headingOffset int
	linkTarget    string
	idPrefix      string
	noHeadingIDs  bool
	size          ui.Size
}

func collect(opts []Option) config {
	var c config
	for _, o := range opts {
		if o != nil {
			o(&c)
		}
	}
	return c
}

// WithUnsafeHTML passes raw HTML in the Markdown through and skips sanitising, for content
// written by trusted authors only (the site owner's own posts). Never use it for anything a
// visitor, a customer or an untrusted account can write: it allows <script>, event handler
// attributes and javascript: URLs.
func WithUnsafeHTML() Option { return func(c *config) { c.unsafe = true } }

// WithPolicy sanitises with p instead of the default policy (Policy). Start from Policy() and
// add to it so that code blocks, task lists and footnotes keep working. Ignored with
// WithUnsafeHTML.
func WithPolicy(p *bluemonday.Policy) Option { return func(c *config) { c.policy = p } }

// WithHeadingOffset shifts every heading down by n levels (n = 1 turns # into <h2>), capped at
// <h6>, so that a document's headings sit under the page's own <h1>. Negative values are
// ignored.
func WithHeadingOffset(n int) Option {
	return func(c *config) {
		if n > 0 {
			c.headingOffset = n
		}
	}
}

// WithLinkTarget sets target on every link, such as "_blank" to open links in a new tab; links
// with target="_blank" also get rel="noopener noreferrer". Only _blank, _self, _parent and _top
// are accepted; anything else is ignored.
func WithLinkTarget(target string) Option {
	return func(c *config) {
		switch target {
		case "_blank", "_self", "_parent", "_top":
			c.linkTarget = target
		}
	}
}

// WithIDPrefix prefixes the ids that headings and footnotes get (heading "Install" becomes
// id="<prefix>install", footnote 1 id="<prefix>fn:1"), so that two documents on one page, or a
// document and the page around it, cannot have the same ids. Use it whenever rendered content
// shares a page with other ids, such as a MarkdownEditor preview: a heading named "Body
// preview" would otherwise take the id of the preview element itself.
func WithIDPrefix(prefix string) Option { return func(c *config) { c.idPrefix = prefix } }

// WithoutHeadingIDs stops headings from getting ids derived from their text.
func WithoutHeadingIDs() Option { return func(c *config) { c.noHeadingIDs = true } }

// WithSize renders Render's output with ui.ProseSm (ui.Sm or ui.Xs) or ui.ProseLg (ui.Lg or
// ui.Xl) instead of ui.Prose.
func WithSize(s ui.Size) Option { return func(c *config) { c.size = s } }
