package git

type LineKind int

const (
	KindContext LineKind = iota
	KindAdded
	KindDeleted
)

type AlignedLine struct {
	OldLineNum  int
	NewLineNum  int
	OldContent  string
	NewContent  string
	Kind        LineKind
}

type AlignedHunk struct {
	OldStart int
	NewStart int
	Header   string
	Lines    []AlignedLine
}

type SideBySideDiff struct {
	OldPath string
	NewPath string
	Status  string
	Hunks   []AlignedHunk
}

type LineAligner interface {
	Align(hunks []Hunk) []AlignedHunk
}

type UnifiedAligner struct{}

func (a *UnifiedAligner) Align(hunks []Hunk) []AlignedHunk {
	result := make([]AlignedHunk, len(hunks))
	for i, h := range hunks {
		ah := AlignedHunk{OldStart: h.OldStart, NewStart: h.NewStart, Header: h.Header, Lines: make([]AlignedLine, 0, len(h.Lines))}
		for _, ln := range h.Lines {
			al := AlignedLine{OldLineNum: ln.OldLineNum, NewLineNum: ln.NewLineNum}
			switch ln.Type {
			case LineContext:
				al.Kind = KindContext
				al.OldContent = ln.Content
				al.NewContent = ln.Content
			case LineDeleted:
				al.Kind = KindDeleted
				al.OldContent = ln.Content
			case LineAdded:
				al.Kind = KindAdded
				al.NewContent = ln.Content
			}
			ah.Lines = append(ah.Lines, al)
		}
		result[i] = ah
	}
	return result
}
