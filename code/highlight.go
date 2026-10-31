package code

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// seg is a run of text of one kind; kind "" is plain text.
type seg struct{ kind, text string }

// highlight splits src (newlines normalised, no trailing newline) into lines of segments. It
// never fails: an unknown language, a lexer error or a lexer whose tokens do not add up to src
// yield plain text.
func highlight(lang, src string) [][]seg {
	n := strings.Count(src, "\n") + 1
	if lex := lexer(lang); lex != nil {
		if lines, ok := tokenize(lex, src, n); ok {
			return lines
		}
	}
	return plain(src)
}

// lexer returns the chroma lexer for a name or alias, or nil for plain text.
func lexer(lang string) chroma.Lexer {
	lang = strings.ToLower(strings.TrimSpace(lang))
	switch lang {
	case "", "text", "txt", "plain", "plaintext":
		return nil
	}
	lex := lexers.Get(lang)
	if lex == nil {
		return nil
	}
	return chroma.Coalesce(lex)
}

func plain(src string) [][]seg {
	parts := strings.Split(src, "\n")
	out := make([][]seg, len(parts))
	for i, p := range parts {
		if p != "" {
			out[i] = []seg{{text: p}}
		}
	}
	return out
}

func tokenize(lex chroma.Lexer, src string, nLines int) ([][]seg, bool) {
	it, err := lex.Tokenise(nil, src)
	if err != nil {
		return nil, false
	}
	tokens := it.Tokens()
	var joined strings.Builder
	for _, t := range tokens {
		joined.WriteString(t.Value)
	}
	// Lexers may add a final newline; the text must otherwise be exactly the source.
	if strings.TrimRight(joined.String(), "\n") != src {
		return nil, false
	}
	name := lex.Config().Name
	keysAreAttributes := name == "JSON" || name == "YAML"
	out := make([][]seg, 0, nLines)
	for _, line := range chroma.SplitTokensIntoLines(tokens) {
		var segs []seg
		for _, t := range line {
			text := strings.TrimSuffix(t.Value, "\n")
			if text == "" {
				continue
			}
			kind := kindOf(t.Type, keysAreAttributes)
			if k := len(segs) - 1; k >= 0 && segs[k].kind == kind {
				segs[k].text += text
				continue
			}
			segs = append(segs, seg{kind, text})
		}
		out = append(out, segs)
	}
	// Drop the empty lines a lexer's final newline would add.
	if len(out) > nLines {
		out = out[:nLines]
	}
	for len(out) < nLines {
		out = append(out, nil)
	}
	return out, true
}

// kindOf maps a chroma token type to one of the semantic classes (tok-KIND), or "" for text that
// needs no colour. keysAreAttributes colours the keys of JSON and YAML (chroma calls them tags)
// like attributes.
func kindOf(t chroma.TokenType, keysAreAttributes bool) string {
	switch t {
	case chroma.KeywordType, chroma.NameClass, chroma.NameException:
		return "type"
	case chroma.KeywordConstant, chroma.NameConstant:
		return "literal"
	case chroma.NameTag:
		if keysAreAttributes {
			return "attribute"
		}
		return "tag"
	case chroma.NameAttribute, chroma.NameDecorator:
		return "attribute"
	case chroma.NameLabel, chroma.NameEntity, chroma.NameNamespace, chroma.NameOther, chroma.NameProperty:
		return "name"
	case chroma.OperatorWord, chroma.GenericHeading, chroma.GenericSubheading:
		return "keyword"
	case chroma.GenericInserted:
		return "string"
	case chroma.Error, chroma.GenericDeleted, chroma.GenericError:
		return "error"
	}
	switch t.SubCategory() {
	case chroma.NameBuiltin:
		return "builtin"
	case chroma.NameVariable:
		return "name"
	case chroma.NameFunction:
		return "function"
	case chroma.LiteralString:
		return "string"
	case chroma.LiteralNumber:
		return "number"
	case chroma.CommentPreproc:
		return "keyword"
	}
	switch t.Category() {
	case chroma.Keyword:
		return "keyword"
	case chroma.Literal:
		return "literal"
	case chroma.Operator:
		return "operator"
	case chroma.Punctuation:
		return "punctuation"
	case chroma.Comment:
		return "comment"
	}
	return ""
}
