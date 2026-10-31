package code

// Option configures Block. Inline takes no options.
type Option func(*config)

type config struct {
	lineNumbers bool
	start       int
	highlight   [][2]int // inclusive ranges of displayed line numbers
	noCopy      bool
	title       string
	wrap        bool
	id          string
}

func collect(opts []Option) config {
	c := config{start: 1}
	for _, o := range opts {
		if o != nil {
			o(&c)
		}
	}
	return c
}

// WithLineNumbers shows a line-number gutter. The numbers are CSS counters: they are not part of
// the text, so selecting and copying the code never includes them.
func WithLineNumbers() Option { return func(c *config) { c.lineNumbers = true } }

// WithStart sets the number of the first line (default 1), for a snippet taken from the middle
// of a file. It applies to the gutter and to WithHighlight, which names lines by the number the
// gutter shows. Values below 1 are ignored.
func WithStart(n int) Option {
	return func(c *config) {
		if n >= 1 {
			c.start = n
		}
	}
}

// WithHighlight marks lines with a tint and a bar. Lines are named by their displayed number,
// counting from WithStart (1 by default); numbers outside the snippet are ignored.
func WithHighlight(lines ...int) Option {
	return func(c *config) {
		for _, n := range lines {
			c.highlight = append(c.highlight, [2]int{n, n})
		}
	}
}

// WithHighlightRange is WithHighlight for lines from through to, inclusive.
func WithHighlightRange(from, to int) Option {
	return func(c *config) { c.highlight = append(c.highlight, [2]int{from, to}) }
}

func (c config) highlighted(n int) bool {
	for _, r := range c.highlight {
		if r[0] <= n && n <= r[1] {
			return true
		}
	}
	return false
}

// WithoutCopyButton removes the copy button, which Block shows by default.
func WithoutCopyButton() Option { return func(c *config) { c.noCopy = true } }

// WithTitle shows title, typically a file name, in a toolbar above the code, and names the figure
// for assistive technology.
func WithTitle(title string) Option { return func(c *config) { c.title = title } }

// WithWrap wraps long lines instead of scrolling horizontally (the default). Wrapped lines hang
// under the code, not under the line number.
func WithWrap() Option { return func(c *config) { c.wrap = true } }

// WithID sets the string the copy button's signal is named after (letters and digits are kept
// lower-cased; anything else becomes an underscore). By default it is a hash of the language and
// source, which keeps output deterministic but gives two blocks with identical language and
// source the same signal, so copying one shows "Copied" on both. Pass distinct ids for
// duplicates.
func WithID(id string) Option { return func(c *config) { c.id = id } }
