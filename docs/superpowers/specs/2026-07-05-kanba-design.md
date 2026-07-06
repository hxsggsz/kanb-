# Kanba — TUI Git Diff Viewer

## Overview

Kanba is a multi-file terminal diff viewer for git, built with Go and Bubble Tea v2. It mirrors the review-first workflow of [hunk](https://github.com/modem-dev/hunk) with a sidebar + split diff layout, rendered entirely in the terminal.

## Architecture

```
main.go → tea.NewProgram(model)
            │
            │ Init()
            │  └─ returns gitDiffCmd (async)
            │       └─ goroutine → exec.Command("git", "diff", ...)
            │            └─ parseRawDiff() → diffMsg
            │
            │ Update()
            │  ├─ diffMsg   → store diffs, clear loading
            │  ├─ KeyPressMsg → navigate files/hunks, scroll, toggle help
            │  └─ WindowSizeMsg → recalculate layout
            │
            │ View()
            │  ├─ loading    → centered spinner
            │  └─ ready      → sidebar + diff panel + status bar
```

## Packages

### `git/` — zero dependencies, pure Go

**`git/models.go`** — type definitions:

```go
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
    OldStart, OldCount int
    NewStart, NewCount int
    Header             string
    Lines              []Line
}

type FileDiff struct {
    OldPath string
    NewPath string
    Status  string // M, A, D, R
    Hunks   []Hunk
    IsBinary bool
    IsNew    bool
    IsDelete bool
    IsRename bool
}
```

**`git/diff.go`** — runs git and returns parsed diffs:

```go
func Diff(repoPath string, args []string) ([]FileDiff, error)
```

Shells out to `git diff --no-color --unified=3 [args...]` via `os/exec`. Captures stdout, returns parser results. Error on non-zero exit or parse failure.

**`git/parse.go`** — unified diff state machine parser.

State machine with transitions: file_header → old_path → new_path → hunk_header → diff_lines, handling edge cases for binary files, new/deleted/renamed files, permission-only changes, and the `\ No newline at end of file` marker.

### `tui/` — Bubble Tea application

**`tui/model.go`** — application state:

```go
type model struct {
    diffs    []git.FileDiff
    fileIdx  int         // current file index
    scroll   int         // vertical scroll offset
    screen   screen      // diff, help
    loading  bool
    err      error
    width    int
    height   int
}
```

**`tui/cmd.go`** — async commands:

```go
func gitDiffCmd(repoPath string, args []string) tea.Cmd {
    return func() tea.Msg {
        diffs, err := git.Diff(repoPath, args)
        return diffMsg{diffs, err}
    }
}
```

**`tui/update.go`** — message routing:

| Message | Action |
|---|---|
| `diffMsg` | Store diffs, set loading=false |
| `tea.KeyPressMsg` | Route to per-screen handler |
| `tea.WindowSizeMsg` | Save dimensions |

**`tui/view.go`** — renders three regions:

1. **Sidebar** (left, ~30 cols): file list with status colors. Current file highlighted.
2. **Diff panel** (right): line numbers | colored content lines. Green for additions, red for deletions.
3. **Status bar** (bottom, 1 line): "file N/M • ↑↓ scroll • n/p file • ? help • q quit"

**`tui/keys.go`** — key constants:

| Key | Action |
|---|---|
| `q`, `ctrl+c` | Quit |
| `↑`, `k` | Scroll up (diff panel) |
| `↓`, `j` | Scroll down (diff panel) |
| `n` | Next file |
| `p` | Previous file |
| `g` | Scroll to top |
| `G` | Scroll to bottom |
| `?` | Toggle help |

**`tui/styles.go`** — Lip Gloss styles for:
- Added lines: green foreground/background
- Deleted lines: red foreground/background
- Context lines: default
- File list items: normal + highlighted
- Status bar: dimmed
- Error message: bold red

**`tui/messages.go`** — custom message types:

```go
type diffMsg struct {
    diffs []git.FileDiff
    err   error
}
```

### `main.go` — entry point

Parses CLI args to determine git diff mode (unstaged, staged, commit), sets up debug logging via `DEBUG` env var, resolves git repo root, creates and runs the Bubble Tea program.

## CLI Interface

```
kanba              # equivalent to `kanba diff`
kanba diff         # unstaged changes
kanba diff --staged  # staged changes
kanba show [ref]   # show a commit
kanba --help       # usage text
```

## Error Handling

1. **Not a git repo**: `git diff` fails → show error message centered, exit on any keypress
2. **No changes to show**: empty `[]FileDiff` → show "No changes" message
3. **Binary files**: skip binaries, show warning in sidebar
4. **git not installed**: exec error → show "git not found" error

## Testing

| Package | Focus | Method |
|---|---|---|
| `git/parse_test.go` | Parser correctness | Raw diff strings as test inputs |
| `git/diff_test.go` | Git integration | `t.TempDir()` + `git init` + sample commits |
| `tui/*_test.go` | Model/View logic | Direct struct tests |

## Future (post-v1)

- Split/stack layout toggle
- Syntax highlighting
- Watch mode (auto-reload on file change)
- Agent annotation support
