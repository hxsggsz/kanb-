# Kanba Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a multi-file TUI git diff viewer with sidebar + diff panel layout

**Architecture:** Pure Go project, zero external dependencies for git parsing. Shell out to system git. Bubble Tea v2 for the TUI. Lip Gloss for styling.

**Tech Stack:** Go 1.25+, Bubble Tea v2 (`charm.land/bubbletea/v2`), Lip Gloss v2 (`charm.land/lipgloss/v2`)

---

### Task 1: Git data types

**Files:**
- Create: `git/models.go`

- [ ] **Step 1: Create git/models.go**

```go
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
```

- [ ] **Step 2: Verify it compiles**

Run: `go vet ./git/`
Expected: no output

- [ ] **Step 3: Commit**

```bash
git add git/models.go
git commit -m "feat: add git data models"
```

---

### Task 2: Unified diff parser — state machine

**Files:**
- Create: `git/parse.go`

- [ ] **Step 1: Create git/parse.go with the state machine parser**

The parser reads unified diff lines and produces `[]FileDiff`. It tracks state transitions: file_header → old_path → new_path → hunk_header → diff_lines.

```go
package git

import (
	"fmt"
	"strconv"
	"strings"
)

type parseState int

const (
	stateHeader    parseState = iota
	stateOldPath
	stateNewPath
	stateHunk
	stateLine
)

const (
	diffLineParts = 4  // "diff --git a/old b/new" splits into [diff, --git, a/old, b/new]
	oldPathIdx    = 2
	newPathIdx    = 3
	hunkParts     = 2 // "@@ coords @@ desc" splits into [coords, desc]
	lineNumBase   = 1 // unified diff line numbers are 1-based
	lineNumNone   = 0 // sentinel when a line has no old or new number
	defaultHunkCount = 1 // implied count when hunk header omits it (e.g. @@ -1 +1 @@)
)

func parseRawDiff(lines []string) ([]FileDiff, error) {
	var files []FileDiff
	var cur *FileDiff
	var state parseState

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "diff --git"):
			if cur != nil {
				files = append(files, *cur)
			}
			cur = &FileDiff{}
			parts := strings.SplitN(line, " ", diffLineParts)
			if len(parts) >= diffLineParts {
				cur.OldPath = strings.TrimPrefix(parts[oldPathIdx], "a/")
				cur.NewPath = strings.TrimPrefix(parts[newPathIdx], "b/")
			}
			state = stateOldPath

		case strings.HasPrefix(line, "--- "):
			state = stateNewPath

		case strings.HasPrefix(line, "+++ "):
			state = stateHunk

		case strings.HasPrefix(line, "@@"):
			h, err := parseHunkHeader(line)
			if err != nil {
				return nil, fmt.Errorf("parse hunk header: %w", err)
			}
			cur.Hunks = append(cur.Hunks, h)
			state = stateLine

		case strings.HasPrefix(line, "Binary files"):
			cur.IsBinary = true

		case strings.HasPrefix(line, "new file mode"):
			cur.IsNew = true
			cur.Status = "A"

		case strings.HasPrefix(line, "deleted file mode"):
			cur.IsDelete = true
			cur.Status = "D"

		case strings.HasPrefix(line, "rename from"):
			cur.IsRename = true
			cur.Status = "R"

		case strings.HasPrefix(line, "old mode") || strings.HasPrefix(line, "new mode"):
			// permission change, skip

		case strings.HasPrefix(line, "similarity index") || strings.HasPrefix(line, "copy from") || strings.HasPrefix(line, "copy to"):
			// skip rename/copy metadata

		case state == stateLine && len(line) > 0 && cur != nil && len(cur.Hunks) > 0:
			h := &cur.Hunks[len(cur.Hunks)-1]
			prefix := line[0]
			content := line[1:]

			oldNum, newNum := lineNumNone, lineNumNone
			if len(h.Lines) > 0 {
				last := h.Lines[len(h.Lines)-1]
				oldNum = last.OldLineNum
				newNum = last.NewLineNum
			} else {
				oldNum = h.OldStart - lineNumBase
				newNum = h.NewStart - lineNumBase
			}

			switch prefix {
			case ' ':
				oldNum++
				newNum++
				h.Lines = append(h.Lines, Line{Type: LineContext, OldLineNum: oldNum, NewLineNum: newNum, Content: content})
			case '+':
				newNum++
				h.Lines = append(h.Lines, Line{Type: LineAdded, OldLineNum: lineNumNone, NewLineNum: newNum, Content: content})
			case '-':
				oldNum++
				h.Lines = append(h.Lines, Line{Type: LineDeleted, OldLineNum: oldNum, NewLineNum: lineNumNone, Content: content})
			case '\\':
				// No newline at end of file — attach to last line
				if len(h.Lines) > 0 {
					last := &h.Lines[len(h.Lines)-1]
					last.Content += " " + strings.TrimPrefix(line, "\\ ")
				}
			}

			// Update line numbers correctly
			recalcLineNums(h)
		}
	}

	if cur != nil {
		files = append(files, *cur)
	}

	// Convert empty status to "M" for files with content changes
	for i := range files {
		if files[i].Status == "" {
			files[i].Status = "M"
		}
	}

	return files, nil
}

func parseHunkHeader(line string) (Hunk, error) {
	var h Hunk
	// @@ -start,count +start,count @@ optional
	line = strings.TrimPrefix(line, "@@ ")
	parts := strings.SplitN(line, " @@", hunkParts)
	if len(parts) < 1 {
		return h, fmt.Errorf("invalid hunk header: %s", line)
	}
	spaces := strings.Fields(parts[0])
	if len(spaces) < 2 {
		return h, fmt.Errorf("invalid hunk header: %s", line)
	}

	parsePart := func(s string) (start, count int, err error) {
		s = strings.TrimPrefix(s, "-")
		s = strings.TrimPrefix(s, "+")
		comma := strings.Index(s, ",")
		if comma < 0 {
			start, err = strconv.Atoi(s)
			count = defaultHunkCount
		} else {
			start, err = strconv.Atoi(s[:comma])
			if err == nil {
				count, err = strconv.Atoi(s[comma+1:])
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

	h.Header = line
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
```

- [ ] **Step 2: Verify it compiles**

Run: `go vet ./git/`
Expected: no output

- [ ] **Step 3: Commit**

```bash
git add git/parse.go
git commit -m "feat: add unified diff parser"
```

---

### Task 3: Parser tests

**Files:**
- Create: `git/parse_test.go`

- [ ] **Step 1: Create git/parse_test.go**

```go
package git

import (
	"testing"
)

func TestParseSimpleDiff(t *testing.T) {
	input := `diff --git a/hello.go b/hello.go
index abc123..def456 100644
--- a/hello.go
+++ b/hello.go
@@ -1,4 +1,5 @@
 package main
 
-func hello() {
-       fmt.Println("hello")
+func hello(name string) {
+       fmt.Println("hello", name)
 }
`

	files, err := parseRawDiff(splitLines(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	f := files[0]
	if f.NewPath != "hello.go" {
		t.Errorf("expected hello.go, got %s", f.NewPath)
	}
	if f.Status != "M" {
		t.Errorf("expected M, got %s", f.Status)
	}
	if len(f.Hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(f.Hunks))
	}
	h := f.Hunks[0]
	if h.OldStart != 1 || h.OldCount != 4 {
		t.Errorf("old start/count: %d,%d", h.OldStart, h.OldCount)
	}
	if h.NewStart != 1 || h.NewCount != 5 {
		t.Errorf("new start/count: %d,%d", h.NewStart, h.NewCount)
	}
	if len(h.Lines) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(h.Lines))
	}
	expected := []struct {
		typ LineType
		old int
		new int
	}{
		{LineContext, 1, 1},
		{LineContext, 2, 2},
		{LineDeleted, 3, lineNumNone},
		{LineAdded, lineNumNone, 3},
		{LineContext, 4, 4},
	}
	for i, exp := range expected {
		got := h.Lines[i]
		if got.Type != exp.typ || got.OldLineNum != exp.old || got.NewLineNum != exp.new {
			t.Errorf("line %d: got type=%d old=%d new=%d, want type=%d old=%d new=%d",
				i, got.Type, got.OldLineNum, got.NewLineNum, exp.typ, exp.old, exp.new)
		}
	}
}

func TestParseNewFile(t *testing.T) {
	input := `diff --git a/new.go b/new.go
new file mode 100644
index 0000000..abc1234
--- /dev/null
+++ b/new.go
@@ -0,0 +1,3 @@
+package main
+
+func newFunc() {}
`
	files, err := parseRawDiff(splitLines(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	f := files[0]
	if !f.IsNew {
		t.Error("expected IsNew")
	}
	if f.Status != "A" {
		t.Errorf("expected A, got %s", f.Status)
	}
}

func TestParseDeletedFile(t *testing.T) {
	input := `diff --git a/old.go b/old.go
deleted file mode 100644
index abc1234..0000000
--- a/old.go
+++ /dev/null
@@ -1,5 +0,0 @@
-package main
-
-func oldFunc() {
-       fmt.Println("bye")
-}
`
	files, err := parseRawDiff(splitLines(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	f := files[0]
	if !f.IsDelete {
		t.Error("expected IsDelete")
	}
	if f.Status != "D" {
		t.Errorf("expected D, got %s", f.Status)
	}
	if len(f.Hunks[0].Lines) != 5 {
		t.Errorf("expected 5 lines, got %d", len(f.Hunks[0].Lines))
	}
}

func TestParseBinaryFile(t *testing.T) {
	input := `diff --git a/image.png b/image.png
index abc123..def456 100644
Binary files a/image.png and b/image.png differ
`
	files, err := parseRawDiff(splitLines(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if !files[0].IsBinary {
		t.Error("expected IsBinary")
	}
}

func TestParseEmptyDiff(t *testing.T) {
	files, err := parseRawDiff([]string{})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestParseMultipleFiles(t *testing.T) {
	input := `diff --git a/a.go b/a.go
--- a/a.go
+++ b/a.go
@@ -1 +1 @@
-a
+b
diff --git a/b.go b/b.go
--- a/b.go
+++ b/b.go
@@ -1 +1 @@
-x
+y
`
	files, err := parseRawDiff(splitLines(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if files[0].NewPath != "a.go" || files[1].NewPath != "b.go" {
		t.Errorf("wrong file paths: %s, %s", files[0].NewPath, files[1].NewPath)
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./git/ -v -run TestParse`
Expected: All tests pass

- [ ] **Step 3: Commit**

```bash
git add git/parse_test.go
git commit -m "feat: add parser tests"
```

---

### Task 4: Git diff command runner

**Files:**
- Create: `git/diff.go`

- [ ] **Step 1: Create git/diff.go**

```go
package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func Diff(repoPath string, args []string) ([]FileDiff, error) {
	cmdArgs := []string{"diff", "--no-color", "--unified=3"}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command("git", cmdArgs...)
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git diff failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("git diff: %w", err)
	}

	raw := string(out)
	if raw == "" {
		return []FileDiff{}, nil
	}

	return parseRawDiff(strings.Split(raw, "\n"))
}

func RepoRoot(path string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go vet ./git/`
Expected: no output

- [ ] **Step 3: Commit**

```bash
git add git/diff.go
git commit -m "feat: add git diff command runner"
```

---

### Task 5: Git diff integration tests

**Files:**
- Create: `git/diff_test.go`

- [ ] **Step 1: Create git/diff_test.go**

```go
package git

import (
	"os"
	"os/exec"
	"testing"
)

func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("setup %v: %s", args, out)
		}
	}
	return dir
}

func gitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %s", args, out)
	}
}

func writeFile(t *testing.T, dir, path, content string) {
	t.Helper()
	if err := os.WriteFile(dir+"/"+path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestDiffWithGit(t *testing.T) {
	dir := setupTestRepo(t)
	writeFile(t, dir, "hello.go", "package main\n\nfunc main() {}\n")
	gitCmd(t, dir, "add", ".")
	gitCmd(t, dir, "commit", "-m", "initial")

	writeFile(t, dir, "hello.go", "package main\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n")
	gitCmd(t, dir, "add", ".")
	writeFile(t, dir, "hello.go", "package main\n\nfunc main() {\n\tfmt.Println(\"hello world\")\n}\n")

	diffs, err := Diff(dir, []string{})
	if err != nil {
		t.Fatal(err)
	}
	if len(diffs) == 0 {
		t.Fatal("expected diffs")
	}
	if diffs[0].NewPath != "hello.go" {
		t.Errorf("expected hello.go, got %s", diffs[0].NewPath)
	}
}

func TestDiffStaged(t *testing.T) {
	dir := setupTestRepo(t)
	writeFile(t, dir, "a.go", "package main\n")
	gitCmd(t, dir, "add", ".")
	gitCmd(t, dir, "commit", "-m", "initial")

	writeFile(t, dir, "a.go", "package main\n\nfunc f() {}\n")
	gitCmd(t, dir, "add", ".")

	diffs, err := Diff(dir, []string{"--staged"})
	if err != nil {
		t.Fatal(err)
	}
	if len(diffs) == 0 {
		t.Fatal("expected staged diffs")
	}
}

func TestDiffNoChanges(t *testing.T) {
	dir := setupTestRepo(t)
	writeFile(t, dir, "a.go", "package main\n")
	gitCmd(t, dir, "add", ".")
	gitCmd(t, dir, "commit", "-m", "initial")

	diffs, err := Diff(dir, []string{})
	if err != nil {
		t.Fatal(err)
	}
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs, got %d", len(diffs))
	}
}

func TestRepoRoot(t *testing.T) {
	dir := setupTestRepo(t)
	root, err := RepoRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	if root != dir {
		t.Errorf("expected %s, got %s", dir, root)
	}
}

func TestRepoRootOutside(t *testing.T) {
	dir := t.TempDir()
	_, err := RepoRoot(dir)
	if err == nil {
		t.Error("expected error for non-repo directory")
	}
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./git/ -v`
Expected: All tests pass

- [ ] **Step 3: Commit**

```bash
git add git/diff_test.go
git commit -m "feat: add git diff integration tests"
```

---

### Task 6: TUI messages and async command

**Files:**
- Create: `tui/messages.go`
- Create: `tui/cmd.go`

- [ ] **Step 1: Create tui/messages.go**

```go
package tui

import "kanba/git"

type diffMsg struct {
	diffs []git.FileDiff
	err   error
}
```

- [ ] **Step 2: Create tui/cmd.go**

```go
package tui

import (
	tea "charm.land/bubbletea/v2"
	"kanba/git"
)

func gitDiffCmd(repoPath string, args []string) tea.Cmd {
	return func() tea.Msg {
		diffs, err := git.Diff(repoPath, args)
		return diffMsg{diffs, err}
	}
}
```

- [ ] **Step 3: Verify it compiles**

Run: `go vet ./tui/`
Expected: no output

- [ ] **Step 4: Commit**

```bash
git add tui/messages.go tui/cmd.go
git commit -m "feat: add TUI messages and async git command"
```

---

### Task 7: Key bindings and styles

**Files:**
- Create: `tui/keys.go`
- Create: `tui/styles.go`

- [ ] **Step 1: Create tui/keys.go**

```go
package tui

const (
	keyQuit    = "ctrl+c"
	keyQuitAlt = "q"
	keyUp      = "up"
	keyUpAlt   = "k"
	keyDown    = "down"
	keyDownAlt = "j"
	keyNext    = "n"
	keyPrev    = "p"
	keyTop     = "g"
	keyBottom  = "G"
	keyHelp    = "?"
)
```

- [ ] **Step 2: Create tui/styles.go**

```go
package tui

import "charm.land/lipgloss/v2"

var (
	sidebarStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderRight(true).
			Padding(0, 1).
			Width(30)

	sidebarFile = lipgloss.NewStyle().
			PaddingLeft(1)

	sidebarFileSelected = lipgloss.NewStyle().
				PaddingLeft(0).
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true)

	sidebarStatusAdded = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00"))

	sidebarStatusDeleted = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF0000"))

	sidebarStatusModified = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFA500"))

	lineAddedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Background(lipgloss.Color("#003300"))

	lineDeletedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Background(lipgloss.Color("#330000"))

	lineContextStyle = lipgloss.NewStyle()

	lineNumStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	hunkHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8888FF")).
			Bold(true)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#333333")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Padding(1, 2)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true).
			Padding(2, 4)

	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Padding(2, 4)
)
```

- [ ] **Step 3: Verify it compiles**

Run: `go vet ./tui/`
Expected: no output

- [ ] **Step 4: Commit**

```bash
git add tui/keys.go tui/styles.go
git commit -m "feat: add key bindings and styles"
```

---

### Task 8: TUI model and Init

**Files:**
- Modify: `tui/model.go`

- [ ] **Step 1: Write tui/model.go**

```go
package tui

import (
	tea "charm.land/bubbletea/v2"
	"kanba/git"
)

type screen int

const (
	screenDiff screen = iota
	screenHelp
)

const (
	statusBarHeight = 1
	borderHeight    = 2
	hunkHeaderLines = 2
	lineNumColWidth = 4
)

const (
	sidebarMinWidth    = 15
	sidebarMaxWidth    = 35
	sidebarDefaultWidth = 25
	sidebarDenominator = 4
	panelBorderWidth   = 2
	panelMinWidth      = 10
)

type model struct {
	diffs   []git.FileDiff
	fileIdx int
	scroll  int
	screen  screen
	loading bool
	err     error
	width   int
	height  int

	repoPath string
	gitArgs  []string
}

func New(repoPath string, gitArgs []string) tea.Model {
	return &model{
		repoPath: repoPath,
		gitArgs:  gitArgs,
		loading:  true,
	}
}

func (m *model) Init() tea.Cmd {
	return gitDiffCmd(m.repoPath, m.gitArgs)
}

func (m *model) visibleLines() int {
	return m.height - (statusBarHeight + borderHeight)
}

func (m *model) maxScroll() int {
	if len(m.diffs) == 0 {
		return 0
	}
	f := m.diffs[m.fileIdx]
	totalLines := 0
	for _, h := range f.Hunks {
		totalLines += hunkHeaderLines + len(h.Lines)
	}
	max := totalLines - m.visibleLines()
	if max < 0 {
		return 0
	}
	return max
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go vet ./tui/`
Expected: no output

- [ ] **Step 3: Commit**

```bash
git add tui/model.go
git commit -m "feat: add TUI model with Init"
```

---

### Task 9: TUI update handlers

**Files:**
- Create: `tui/update.go`

- [ ] **Step 1: Create tui/update.go**

```go
package tui

import (
	tea "charm.land/bubbletea/v2"
)

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case diffMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.diffs = msg.diffs
		}
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

func (m *model) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.err != nil {
		return m, tea.Quit
	}

	switch msg.String() {
	case keyQuit, keyQuitAlt:
		return m, tea.Quit
	}

	switch m.screen {
	case screenDiff:
		return m.handleDiffKeys(msg)
	case screenHelp:
		if msg.String() == keyHelp || msg.String() == keyBack {
			m.screen = screenDiff
		}
		return m, nil
	}

	return m, nil
}

func (m *model) handleDiffKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyUp, keyUpAlt:
		if m.scroll > 0 {
			m.scroll--
		}

	case keyDown, keyDownAlt:
		if m.scroll < m.maxScroll() {
			m.scroll++
		}

	case keyNext:
		if m.fileIdx < len(m.diffs)-1 {
			m.fileIdx++
			m.scroll = 0
		}

	case keyPrev:
		if m.fileIdx > 0 {
			m.fileIdx--
			m.scroll = 0
		}

	case keyTop:
		m.scroll = 0

	case keyBottom:
		m.scroll = m.maxScroll()

	case keyHelp:
		m.screen = screenHelp
	}

	return m, nil
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go vet ./tui/`
Expected: no output

- [ ] **Step 3: Commit**

```bash
git add tui/update.go
git commit -m "feat: add TUI update handlers"
```

---

### Task 10: TUI view — sidebar + diff panel + status bar

**Files:**
- Create: `tui/view.go`

- [ ] **Step 1: Create tui/view.go**

```go
package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"kanba/git"
)

func (m *model) View() tea.View {
	v := tea.NewView("")
	v.AltScreen = true
	v.WindowTitle = "kanba"

	if m.loading {
		v.SetContent(m.loadingView())
		return v
	}
	if m.err != nil {
		v.SetContent(m.errorView())
		return v
	}
	if len(m.diffs) == 0 {
		v.SetContent(m.emptyView())
		return v
	}

	switch m.screen {
	case screenDiff:
		v.SetContent(m.diffView())
	case screenHelp:
		v.SetContent(m.helpView())
	}

	return v
}

func (m *model) loadingView() string {
	return loadingStyle.Render(" Loading diffs...")
}

func (m *model) errorView() string {
	return errorStyle.Render(fmt.Sprintf("Error: %v\n\nPress any key to exit.", m.err))
}

func (m *model) emptyView() string {
	return loadingStyle.Render(" No changes to show.")
}

func (m *model) diffView() string {
	sideWidth := sidebarDefaultWidth
	if m.width > 0 {
		sideWidth = m.width / sidebarDenominator
		if sideWidth < sidebarMinWidth {
			sideWidth = sidebarMinWidth
		}
		if sideWidth > sidebarMaxWidth {
			sideWidth = sidebarMaxWidth
		}
	}

	maxFiles := m.height - (statusBarHeight + borderHeight)
	if maxFiles < 1 {
		maxFiles = 1
	}

	var sb strings.Builder
	for i, f := range m.diffs {
		if i >= maxFiles {
			break
		}
		statusColor := statusColorFor(f.Status)
		label := fmt.Sprintf("%s %s", statusColor.Render(f.Status), f.NewPath)
		if i == m.fileIdx {
			sb.WriteString(sidebarFileSelected.Render("▸ " + label) + "\n")
		} else {
			sb.WriteString(sidebarFile.Render("  " + label) + "\n")
		}
	}
	sidebar := sidebarStyle.Width(sideWidth).Render(sb.String())

	panelWidth := m.width - sideWidth - panelBorderWidth
	if panelWidth < panelMinWidth {
		panelWidth = panelMinWidth
	}

	file := m.diffs[m.fileIdx]
	content := m.renderFile(file, panelWidth)

	statusBar := statusBarStyle.Width(m.width).Render(
		fmt.Sprintf(" %d/%d  •  ↑↓ scroll  •  n/p file  •  g/G top/bottom  •  ? help  •  q quit",
			m.fileIdx+1, len(m.diffs)))

	return fmt.Sprintf("%s%s\n%s", sidebar, content, statusBar)
}

func (m *model) renderFile(f git.FileDiff, width int) string {
	var lines []string
	for _, h := range f.Hunks {
		lines = append(lines, hunkHeaderStyle.Render(h.Header))
		for _, ln := range h.Lines {
			lines = append(lines, formatLine(ln, width))
		}
	}

	start := m.scroll
	end := start + m.visibleLines()
	if start >= len(lines) {
		start = 0
	}
	if end > len(lines) {
		end = len(lines)
	}

	if start >= end {
		return ""
	}

	return strings.Join(lines[start:end], "\n")
}

func formatLine(ln git.Line, width int) string {
	oldStr := ""
	newStr := ""
	if ln.OldLineNum > 0 {
		oldStr = strconv.Itoa(ln.OldLineNum)
	}
	if ln.NewLineNum > 0 {
		newStr = strconv.Itoa(ln.NewLineNum)
	}

	lineNumFmt := fmt.Sprintf("%%%ds %%%ds", lineNumColWidth, lineNumColWidth)
	lineNum := fmt.Sprintf(lineNumFmt, oldStr, newStr)
	prefix := " "
	style := lineContextStyle

	switch ln.Type {
	case git.LineAdded:
		prefix = "+"
		style = lineAddedStyle
	case git.LineDeleted:
		prefix = "-"
		style = lineDeletedStyle
	}

	line := fmt.Sprintf("%s %s %s", lineNumStyle.Render(lineNum), prefix, ln.Content)
	if len(line) > width {
		line = line[:width]
	}

	return style.Render(line)
}

func statusColorFor(status string) lipgloss.Style {
	switch status {
	case "A":
		return sidebarStatusAdded
	case "D":
		return sidebarStatusDeleted
	default:
		return sidebarStatusModified
	}
}

func (m *model) helpView() string {
	content := " Keybindings\n\n"
	bindings := []struct{ key, desc string }{
		{"↑/k", "Scroll up"},
		{"↓/j", "Scroll down"},
		{"n", "Next file"},
		{"p", "Previous file"},
		{"g", "Go to top"},
		{"G", "Go to bottom"},
		{"?", "Toggle help"},
		{"q/ctrl+c", "Quit"},
	}
	for _, b := range bindings {
		content += fmt.Sprintf("  %-12s %s\n", b.key, b.desc)
	}
	return helpStyle.Render(content)
}
```

- [ ] **Step 2: Add keyBack to keys.go**

Edit `tui/keys.go` to add:
```go
const keyBack = "esc"
```

- [ ] **Step 3: Verify it compiles**

Run: `go vet ./tui/`
Expected: no output

- [ ] **Step 4: Commit**

```bash
git add tui/view.go
git commit -m "feat: add TUI view with sidebar and diff panel"
```

---

### Task 11: Main entry point

**Files:**
- Modify: `main.go`

- [ ] **Step 1: Write main.go**

```go
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"kanba/git"
	"kanba/tui"
)

func main() {
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open log file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	args := os.Args[1:]
	gitArgs := []string{}

	if len(args) == 0 || args[0] == "diff" {
		gitArgs = args // includes "diff" or nothing
	} else if args[0] == "show" {
		gitArgs = args // includes "show" and optional ref
	} else if args[0] == "--help" || args[0] == "-h" {
		printUsage()
		return
	} else {
		printUsage()
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting current directory: %v\n", err)
		os.Exit(1)
	}

	repoPath, err := git.RepoRoot(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(tui.New(repoPath, gitArgs))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`kanba — TUI git diff viewer

Usage:
  kanba              Show unstaged changes
  kanba diff         Show unstaged changes
  kanba diff --staged  Show staged changes
  kanba show [ref]   Show a commit
  kanba --help       Show this help`)
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./...`
Expected: no output

- [ ] **Step 3: Run vet**

Run: `go vet ./...`
Expected: no output

- [ ] **Step 4: Commit**

```bash
git add main.go
git commit -m "feat: add main entry point with CLI parsing"
```

---

### Task 12: Verify full build and run all tests

- [ ] **Step 1: Full build**

Run: `go build -o kanba .`
Expected: binary created

- [ ] **Step 2: Run all tests**

Run: `go test ./... -v`
Expected: All tests pass

- [ ] **Step 3: Run vet**

Run: `go vet ./...`
Expected: no output

- [ ] **Step 4: Commit any remaining changes**

```bash
git add -A
git commit -m "chore: final build verification"
```
