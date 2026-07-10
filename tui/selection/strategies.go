package selection

import (
	"unicode"
)

// Strategy defines how a selection is created from a click/drag
type Strategy interface {
	Select(content string, startCol, endCol int) Range
}

// CharacterStrategy selects individual characters (default drag)
type CharacterStrategy struct{}

// WordStrategy selects whole words (double-click)
type WordStrategy struct{}

// Select returns a Range with StartCol and EndCol as-is (pass-through).
func (cs CharacterStrategy) Select(content string, startCol, endCol int) Range {
	return Range{
		StartLine: 0,
		StartCol:  startCol,
		EndLine:   0,
		EndCol:    endCol,
	}
}

// Select finds word boundaries and returns Range with word's start/end columns.
func (ws WordStrategy) Select(content string, startCol, endCol int) Range {
	start := startCol
	end := endCol

	if s, e := findWordBoundaries(content, startCol); s != e || startCol == 0 || startCol >= len(content) {
		start = s
	}
	if s, e := findWordBoundaries(content, endCol); s != e || endCol == 0 || endCol >= len(content) {
		end = e
	}

	if start > end {
		start, end = end, start
	}

	return Range{
		StartLine: 0,
		StartCol:  start,
		EndLine:   0,
		EndCol:    end,
	}
}

// IsWordChar returns true if the rune is a word character.
func IsWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// findWordBoundaries finds the word containing the character at position col.
func findWordBoundaries(line string, col int) (start, end int) {
	runes := []rune(line)
	if len(runes) == 0 {
		return 0, 0
	}

	if col < 0 {
		col = 0
	}
	if col >= len(runes) {
		col = len(runes) - 1
	}

	if !IsWordChar(runes[col]) {
		return col, col
	}

	start = col
	for start > 0 && IsWordChar(runes[start-1]) {
		start--
	}

	end = col
	for end < len(runes)-1 && IsWordChar(runes[end+1]) {
		end++
	}

	return start, end
}
