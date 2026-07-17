package diff

import (
	"strings"
	"sync"

	models "kanba/tui/models"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

type SyntaxHighlighter struct {
	mu    sync.RWMutex
	cache map[string]chroma.Lexer
}

func NewSyntaxHighlighter() *SyntaxHighlighter {
	return &SyntaxHighlighter{
		cache: make(map[string]chroma.Lexer),
	}
}

func (sh *SyntaxHighlighter) HighlightWithStyle(code, filePath string, baseStyle lipgloss.Style, theme models.Theme) string {
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
		value := strings.TrimSuffix(token.Value, "\n")
		if value == "" {
			continue
		}
		tokenStyle := baseStyle
		if c := theme.TokenColors[token.Type]; c != "" {
			tokenStyle = tokenStyle.Foreground(lipgloss.Color(c))
		}
		buf.WriteString(tokenStyle.Render(value))
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
