package selection

type PanelSide int

const (
	PanelLeft PanelSide = iota
	PanelRight
)

type Range struct {
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
}

func (r Range) IsEmpty() bool {
	return r.StartLine == r.EndLine && r.StartCol == r.EndCol
}

func (r Range) Normalized() Range {
	startAfterEndLine := r.StartLine > r.EndLine
	sameLineStartAfterEnd := r.StartLine == r.EndLine && r.StartCol > r.EndCol
	isReversed := startAfterEndLine || sameLineStartAfterEnd

	if isReversed {
		return Range{
			StartLine: r.EndLine, StartCol: r.EndCol,
			EndLine: r.StartLine, EndCol: r.StartCol,
		}
	}
	return r
}

type Selection struct {
	Panel PanelSide
	Range Range
}
