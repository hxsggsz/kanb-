# Syntax Highlighting for Side-by-Side Diff

## Overview

Add file-type-aware syntax highlighting to kanba's side-by-side diff view. The
code content in each diff line is colored by its programming language's syntax
using terminal ANSI foreground codes, while the green/red added/deleted line
backgrounds use terminal ANSI background codes instead of hardcoded hex values.

## Motivation

The current diff view applies a uniform green foreground + green background
to all added lines, and red foreground + red background to all deleted lines.
Code appears monochrome regardless of language. By applying syntax highlighting
and switching to pure ANSI terminal colors, the diff adapts to the user's
terminal theme and makes code easier to read.

## Decisions

### Dependency

Use `github.com/alecthomas/chroma/v2` for lexing (file-type detection) and
terminal-formatted ANSI output. Chroma is the standard Go syntax highlighting
library and its `terminal` formatter outputs ANSI 8-color foreground codes
that naturally adapt to the terminal's theme.

### Library

- chroma v2 for lexing and terminal formatting
- No new framework dependencies beyond chroma
- lipgloss remains for layout (column widths, padding, column separator)

### Highlighter

Per-line tokenization: each line is highlighted individually via chroma's
lexer. This is simpler than full-file tokenization and handles 95% of cases
well. Multiline constructs (strings, comments) may not span across diff lines
but this is acceptable for a terminal diff viewer.

### Background Codes

Replace lipgloss hex-based backgrounds with raw ANSI background codes:

| Context    | ANSI Code | Meaning     |
|------------|-----------|-------------|
| Added      | `\x1b[42m` | Green bg   |
| Deleted    | `\x1b[41m` | Red bg     |
| Context    | (none)    | No bg       |
| Cursor     | `\x1b[7m` | Reverse vid |

These codes work on any ANSI terminal and adapt to the user's theme
(light/dark). The background is injected at every ANSI reset boundary
so it persists across syntax tokens.

## Architecture

### New: `tui/highlighter.go`

`SyntaxHighlighter` struct:
- Caches one `chroma.Lexer` per file path (map keyed by path)
- `Highlight(code string, filePath string) string` — returns code with
  ANSI foreground syntax tokens, or plain code if no lexer matches
- Uses `chroma.Lex(lexer, iterator)` then `chroma.Tokenise()` per line

### New: `tui/background.go`

Background injection utilities:
- `injectBackground(line string, bgCode string) string` — replaces every
  `\x1b[0m` (reset) with `\x1b[0m + bgCode` so the background color
  persists across foreground token changes
- `injectCursor(line string) string` — prepends `\x1b[7m` (reverse video)
  and strips any existing background codes for cursor lines

### Modified: `tui/formatter.go`

- `lineAddedStyle` / `lineDeletedStyle` removed from `styles.go`
- `addedFormatter.RightStyle()` / `deletedFormatter.LeftStyle()` /
  `modifiedFormatter.LeftStyle()` / `modifiedFormatter.RightStyle()`:
  return only `Width(colWidth)` styling (no lipgloss Foreground/Background)
- `contextFormatter.LeftStyle()` / `.RightStyle()` — unchanged (no background)

### Column width padding

Because ANSI background codes are embedded directly in the content string,
lipgloss' `Width()` cannot apply the background to its padding spaces. To
ensure the green/red background fills the full column, padding is handled
manually: compute the visible width of the highlighted line (strip ANSI
codes, count runes), append `bgCode + spaces + reset` to reach `colWidth`.

### Modified: `tui/view.go`

In `renderFile()`, at the `renderAlignedLine` call site:
- Create a `SyntaxHighlighter` once per file (stored on `model` or file-local)
- Before rendering each line, pass the content through the highlighter
- Determine background code based on `ln.Kind` and cursor state
- Apply background injection or cursor reverse video
- Pad the result to `colWidth` with background-colored spaces

### Modified: `tui/styles.go`

Remove `lineAddedStyle` and `lineDeletedStyle` — their functionality is
replaced by ANSI background injection at render time.

## Integration

No changes to:
- `LineFormatter` interface — still 6 methods, same signatures
- `git/` package — no changes to aligned models, parsing, or commands
- `Sidebar`, `StatusBar`, `Scroller` — unchanged
- Keyboard handling, scrolling, cursor movement — unchanged
- Column separator `│` — rendered bare (not syntax-highlighted or reversed)

## Fallback

If `lexers.Match(filePath)` returns nil (unknown extension), the content is
rendered with only the terminal background color — no foreground syntax tokens.
This is clean and readable for any file type.

## Testing

- `SyntaxHighlighter.Highlight` on known file types (Go, JS, Python) produces
  output containing ANSI escape sequences
- ANSI background injection tests: verify `\x1b[42m` / `\x1b[41m` appear per
  token boundary
- Cursor reverse video test: verify `\x1b[7m` appears and no bg codes present
- Existing view tests pass unchanged (no API changes)
