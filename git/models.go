package git

type LineType int

const (
	LineContext LineType = iota
	LineAdded
	LineDeleted
)

type Line struct {
	Type       LineType
	OldLineNum int
	NewLineNum int
	Content    string
}

type Hunk struct {
	OldStart int
	OldCount int
	NewStart int
	NewCount int
	Header   string
	Lines    []Line
}

type FileDiff struct {
	OldPath  string
	NewPath  string
	Status   string
	Hunks    []Hunk
	IsBinary bool
	IsNew    bool
	IsDelete bool
	IsRename bool
}
