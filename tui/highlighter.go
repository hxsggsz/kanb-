package tui

import (
	"strings"
	"sync"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
)

var syntaxFormatter = formatters.Get("terminal16")
var syntaxStyle = buildTerminalStyle()

// buildTerminalStyle creates a chroma style using hex values that map
// exactly to the TTY8 formatter's lookup table, producing ANSI 30-37
// codes that respect the user's terminal color theme.
func buildTerminalStyle() *chroma.Style {
	b := chroma.NewStyleBuilder("terminal")
	add := func(tt chroma.TokenType, hex string) {
		b.AddEntry(tt, chroma.StyleEntry{Colour: chroma.MustParseColour(hex)})
	}
	add(chroma.Text, "#e5e5e5")
	add(chroma.TextWhitespace, "#e5e5e5")
	add(chroma.Punctuation, "#e5e5e5")
	add(chroma.Keyword, "#ffff00")
	add(chroma.KeywordNamespace, "#00ffff")
	add(chroma.KeywordType, "#0000ff")
	add(chroma.NameFunction, "#00ffff")
	add(chroma.NameBuiltin, "#007f7f")
	add(chroma.LiteralString, "#ff00ff")
	add(chroma.LiteralNumber, "#7f007f")
	add(chroma.Comment, "#00007f")
	add(chroma.CommentSingle, "#00007f")
	add(chroma.CommentPreproc, "#00007f")
	add(chroma.Operator, "#ffff00")
	add(chroma.Keyword, "#ffff00")
	add(chroma.KeywordNamespace, "#00ffff")
	add(chroma.KeywordType, "#0000ff")
	add(chroma.KeywordDeclaration, "#ffff00") // Added for safer variable/func keywords

	add(chroma.Comment, "#00007f")
	add(chroma.CommentSingle, "#00007f")
	add(chroma.CommentMultiline, "#00007f") // Added for block comments

	add(chroma.LiteralString, "#ff00ff")
	add(chroma.LiteralStringDouble, "#ff00ff") // Added
	add(chroma.LiteralStringSingle, "#ff00ff") // Added
	s, err := b.Build()
	if err != nil {
		panic(err)
	}
	return s
}

type SyntaxHighlighter struct {
	mu    sync.RWMutex
	cache map[string]chroma.Lexer
}

func NewSyntaxHighlighter() *SyntaxHighlighter {
	return &SyntaxHighlighter{
		cache: make(map[string]chroma.Lexer),
	}
}

func (sh *SyntaxHighlighter) Highlight(code, filePath string) string {
	if code == "" {
		return ""
	}
	lexer := sh.getLexer(filePath)
	if lexer == nil {
		return code
	}
	tokens, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code
	}
	var buf strings.Builder

	if err := syntaxFormatter.Format(&buf, syntaxStyle, tokens); err != nil {
		return code
	}
	return buf.String()
}

func (sh *SyntaxHighlighter) getLexer(filePath string) chroma.Lexer {
	sh.mu.RLock()
	l, ok := sh.cache[filePath]
	sh.mu.RUnlock()
	if ok {
		return l
	}
	l = lexers.Match(filePath)
	if l != nil {
		l = chroma.Coalesce(l)
	}
	sh.mu.Lock()
	sh.cache[filePath] = l
	sh.mu.Unlock()
	return l
}
