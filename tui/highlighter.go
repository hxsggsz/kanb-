package tui

import (
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"charm.land/lipgloss/v2"
)

var syntaxStyle = buildTerminalStyle()

// tty16Colors maps the 16 standard ANSI color hex values to their ANSI
// indices (0-15). These are the same values used by chroma's terminal16
// formatter, producing ANSI 30-37/90-97 codes that respect the terminal
// theme.
var tty16Colors = []struct {
	index int
	color chroma.Colour
}{
	{0, chroma.MustParseColour("#000000")},
	{1, chroma.MustParseColour("#7f0000")},
	{2, chroma.MustParseColour("#007f00")},
	{3, chroma.MustParseColour("#7f7fe0")},
	{4, chroma.MustParseColour("#00007f")},
	{5, chroma.MustParseColour("#7f007f")},
	{6, chroma.MustParseColour("#007f7f")},
	{7, chroma.MustParseColour("#e5e5e5")},
	{8, chroma.MustParseColour("#555555")},
	{9, chroma.MustParseColour("#ff0000")},
	{10, chroma.MustParseColour("#00ff00")},
	{11, chroma.MustParseColour("#ffff00")},
	{12, chroma.MustParseColour("#0000ff")},
	{13, chroma.MustParseColour("#ff00ff")},
	{14, chroma.MustParseColour("#00ffff")},
	{15, chroma.MustParseColour("#ffffff")},
}

func ansiIndexFor(colour chroma.Colour) int {
	best := 0
	bestDist := math.MaxFloat64
	for _, e := range tty16Colors {
		d := colour.Distance(e.color)
		if d < bestDist {
			bestDist = d
			best = e.index
		}
	}
	return best
}

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
	add(chroma.KeywordDeclaration, "#ffff00")
	add(chroma.CommentMultiline, "#00007f")
	add(chroma.LiteralStringDouble, "#ff00ff")
	add(chroma.LiteralStringSingle, "#ff00ff")
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

func (sh *SyntaxHighlighter) HighlightWithStyle(code, filePath string, baseStyle lipgloss.Style) string {
	if code == "" {
		return ""
	}
	lexer := sh.getLexer(filePath)
	if lexer == nil {
		return baseStyle.Render(code)
	}
	tokens, err := chroma.Tokenise(lexer, nil, code)
	if err != nil {
		return baseStyle.Render(code)
	}

	var buf strings.Builder
	for _, token := range tokens {
		entry := syntaxStyle.Get(token.Type)
		tokenStyle := baseStyle
		if entry.Colour.IsSet() {
			ansiIdx := ansiIndexFor(entry.Colour)
			tokenStyle = tokenStyle.Foreground(lipgloss.Color(strconv.Itoa(ansiIdx)))
		}
		buf.WriteString(tokenStyle.Render(token.Value))
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
