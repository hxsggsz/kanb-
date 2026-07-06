package git

import (
	"fmt"
	"strconv"
	"strings"
)

type ParseState int

const (
	StateHeader ParseState = iota
	StateOldPath
	StateNewPath
	StateHunk
	StateLine
)

const (
	lineNumBase = 1
	lineNumNone = 0
)

type LineParser interface {
	Match(line string) bool
	Parse(line string, cur *FileDiff, state *ParseState) error
}

type DiffStartParser struct{}

func (p *DiffStartParser) Match(line string) bool {
	return strings.HasPrefix(line, "diff --git")
}

func (p *DiffStartParser) Parse(line string, cur *FileDiff, _ *ParseState) error {
	parts := strings.SplitN(line, " ", 4)
	if len(parts) >= 4 {
		cur.OldPath = strings.TrimPrefix(parts[2], "a/")
		cur.NewPath = strings.TrimPrefix(parts[3], "b/")
	}
	return nil
}

type HunkHeaderParser struct{}

func (p *HunkHeaderParser) Match(line string) bool {
	return strings.HasPrefix(line, "@@")
}

func (p *HunkHeaderParser) Parse(line string, cur *FileDiff, _ *ParseState) error {
	h, err := parseHunkHeader(line)
	if err != nil {
		return fmt.Errorf("parse hunk header: %w", err)
	}
	cur.Hunks = append(cur.Hunks, h)
	return nil
}

type ContentLineParser struct{}

func (p *ContentLineParser) Match(line string) bool {
	return len(line) > 0 && (line[0] == ' ' || line[0] == '+' || line[0] == '-' || line[0] == '\\')
}

func (p *ContentLineParser) Parse(line string, cur *FileDiff, _ *ParseState) error {
	if len(cur.Hunks) == 0 {
		return nil
	}
	h := &cur.Hunks[len(cur.Hunks)-1]
	prefix := line[0]
	content := line[1:]

	switch prefix {
	case ' ':
		h.Lines = append(h.Lines, Line{Type: LineContext, Content: content})
	case '+':
		h.Lines = append(h.Lines, Line{Type: LineAdded, Content: content})
	case '-':
		h.Lines = append(h.Lines, Line{Type: LineDeleted, Content: content})
	case '\\':
		if len(h.Lines) > 0 {
			last := &h.Lines[len(h.Lines)-1]
			last.Content += " " + strings.TrimPrefix(line, "\\ ")
		}
	}
	return nil
}

type MetadataParser struct {
	prefix string
	apply  func(*FileDiff)
}

func NewMetadataParser(prefix string, apply func(*FileDiff)) *MetadataParser {
	return &MetadataParser{prefix: prefix, apply: apply}
}

func (p *MetadataParser) Match(line string) bool {
	return strings.HasPrefix(line, p.prefix)
}

func (p *MetadataParser) Parse(line string, cur *FileDiff, _ *ParseState) error {
	p.apply(cur)
	return nil
}

type ParserBuilder struct {
	parsers []LineParser
}

func NewParserBuilder() *ParserBuilder {
	return &ParserBuilder{}
}

func (b *ParserBuilder) Add(p LineParser) *ParserBuilder {
	b.parsers = append(b.parsers, p)
	return b
}

func (b *ParserBuilder) AddMetadata(prefix string, apply func(*FileDiff)) *ParserBuilder {
	return b.Add(NewMetadataParser(prefix, apply))
}

func (b *ParserBuilder) Build() *UnifiedParser {
	return &UnifiedParser{parsers: b.parsers}
}

type UnifiedParser struct {
	parsers []LineParser
}

func NewUnifiedParser() *UnifiedParser {
	return NewParserBuilder().
		Add(&DiffStartParser{}).
		AddMetadata("--- ", func(f *FileDiff) {}).
		AddMetadata("+++ ", func(f *FileDiff) {}).
		Add(&HunkHeaderParser{}).
		AddMetadata("Binary files", func(f *FileDiff) { f.IsBinary = true }).
		AddMetadata("new file mode", func(f *FileDiff) { f.IsNew = true; f.Status = "A" }).
		AddMetadata("deleted file mode", func(f *FileDiff) { f.IsDelete = true; f.Status = "D" }).
		AddMetadata("rename from", func(f *FileDiff) { f.IsRename = true; f.Status = "R" }).
		AddMetadata("rename to", func(f *FileDiff) {}).
		AddMetadata("old mode", func(f *FileDiff) {}).
		AddMetadata("new mode", func(f *FileDiff) {}).
		AddMetadata("similarity index", func(f *FileDiff) {}).
		AddMetadata("copy from", func(f *FileDiff) {}).
		AddMetadata("copy to", func(f *FileDiff) {}).
		AddMetadata("index ", func(f *FileDiff) {}).
		Add(&ContentLineParser{}).
		Build()
}

type parseEngine struct {
	parsers []LineParser
	files   []FileDiff
	cur     *FileDiff
	state   ParseState
	err     error
}

func (e *parseEngine) feed(line string) {
	p := e.match(line)
	if p == nil {
		if e.cur != nil && e.state == StateLine {
			e.cur.IsBinary = true
		}
		return
	}
	if _, ok := p.(*DiffStartParser); ok {
		if e.cur != nil {
			e.files = append(e.files, *e.cur)
		}
		e.cur = &FileDiff{}
		e.state = StateHeader
	}
	if e.cur == nil {
		return
	}
	e.err = p.Parse(line, e.cur, &e.state)
}

func (e *parseEngine) match(line string) LineParser {
	for _, p := range e.parsers {
		if p.Match(line) {
			return p
		}
	}
	return nil
}

func (p *UnifiedParser) Parse(raw string) ([]FileDiff, error) {
	eng := &parseEngine{parsers: p.parsers}
	for _, line := range strings.Split(raw, "\n") {
		eng.feed(line)
		if eng.err != nil {
			return nil, eng.err
		}
	}
	if eng.cur != nil {
		eng.files = append(eng.files, *eng.cur)
	}
	for i := range eng.files {
		if eng.files[i].Status == "" {
			eng.files[i].Status = "M"
		}
		for j := range eng.files[i].Hunks {
			recalcLineNums(&eng.files[i].Hunks[j])
		}
	}
	return eng.files, nil
}

func parseHunkHeader(rawLine string) (Hunk, error) {
	var h Hunk
	h.Header = rawLine
	line := strings.TrimPrefix(rawLine, "@@ ")
	parts := strings.SplitN(line, " @@", 2)
	if len(parts) < 2 {
		return h, fmt.Errorf("invalid hunk header: %s", line)
	}
	spaces := strings.Fields(parts[0])
	if len(spaces) < 2 {
		return h, fmt.Errorf("invalid hunk header: %s", line)
	}

	parsePart := func(s string) (start, count int, err error) {
		s = strings.TrimPrefix(s, "-")
		s = strings.TrimPrefix(s, "+")
		before, after, ok := strings.Cut(s, ",")
		if !ok {
			start, err = strconv.Atoi(s)
			count = 1
		} else {
			start, err = strconv.Atoi(before)
			if err == nil {
				count, err = strconv.Atoi(after)
			}
		}
		return
	}

	var err error
	h.OldStart, h.OldCount, err = parsePart(spaces[0])
	if err != nil {
		return h, fmt.Errorf("parse old position: %w", err)
	}
	h.NewStart, h.NewCount, err = parsePart(spaces[1])
	if err != nil {
		return h, fmt.Errorf("parse new position: %w", err)
	}
	return h, nil
}

func recalcLineNums(h *Hunk) {
	oldNum := h.OldStart - lineNumBase
	newNum := h.NewStart - lineNumBase
	for i := range h.Lines {
		ln := &h.Lines[i]
		switch ln.Type {
		case LineContext:
			oldNum++
			newNum++
			ln.OldLineNum = oldNum
			ln.NewLineNum = newNum
		case LineAdded:
			newNum++
			ln.OldLineNum = lineNumNone
			ln.NewLineNum = newNum
		case LineDeleted:
			oldNum++
			ln.OldLineNum = oldNum
			ln.NewLineNum = lineNumNone
		}
	}
}
