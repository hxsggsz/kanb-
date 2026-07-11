# Remove Cursor, Scroll-Only Navigation

## Goal

Remove the visual cursor line and cursor-based navigation from all diff panels. Replace with pure scroll-based navigation. Keep `CursorBg` theme color for text selection highlighting only.

## Summary

The current codebase tracks a `cursorLine` index that determines which line is "active". Keyboard navigation (j/k/g/G, arrows) moves this cursor, and the viewport follows. Each render cycle highlights the cursor line with a blended `CursorBg` color. This design removes all of that: no cursor state, no cursor rendering, no cursor-driven scroll. Navigation becomes direct scroll manipulation.

`CursorBg` remains as a theme color used exclusively by `SelectionHighlighter` for text selection background.

## Changes

### 1. Scroller (`tui/diff/scroller.go`)

Remove `cursorLine` field. The struct becomes:

```go
type Scroller struct {
    scroll     int
    hScroll    int
    scrollLock bool
}
```

Methods change to scroll-direct:

- `MoveDown(total int, vis int)`: `s.scroll = min(s.scroll+1, max(0, total-vis))`
- `MoveUp()`: `s.scroll = max(s.scroll-1, 0)`
- `GoToTop()`: `s.scroll = 0`
- `GoToBottom(total int, vis int)`: `s.scroll = max(0, total-vis)`
- `UpdateScroll(total int, vis int)`: clamp `scroll` to `[0, max(0, total-vis)]`, reset `scrollLock`. No cursor clamping, no margin logic.
- Remove `CursorLine()` method.
- `ScrollViewBy`, `ScrollLeft`, `ScrollRight`, `ScrollLeftFast`, `ScrollRightFast`, `ScrollHome`, `ScrollEnd` unchanged.

### 2. Keyboard handling (`tui/app/update.go`)

In `handleDiffKeys()`:

- `KeyUp`/`KeyUpAlt` (`k`/`↑`): call `scroller.MoveUp()` — scroll 1 line up
- `KeyDown`/`KeyDownAlt` (`j`/`↓`): call `scroller.MoveDown(totalLines, vis)` — scroll 1 line down
- `KeyTop` (`g`): call `scroller.GoToTop()`
- `KeyBottom` (`G`): call `scroller.GoToBottom(totalLines, vis)`
- Remove header-skipping loops (no cursor to skip past headers)
- `totalLines` obtained from `len(m.flatLines)`, `vis` from `m.visibleLines`

Mouse click in content area: compute clicked flat line index, set `scroll` directly to position that line at the top (or near top) of the viewport. Remove cursor jump logic.

Mouse click in sidebar: compute first content line of clicked file, set `scroll` to that line.

### 3. Rendering — remove `cursor` boolean everywhere

#### `tui/app/view.go`

`renderContinuous()`:
- Remove `cursorLine := m.scroller.CursorLine()`
- Remove `cursor := gi == cursorLine` from the loop
- Call `m.renderLine(fl, width, hScroll, selHighlighter, gi, theme)` without cursor param

`renderLine()`:
- Remove `cursor bool` parameter
- Pass to `diff.RenderAlignedLine` without cursor
- Pass to `m.renderFileHeader` without cursor

`renderFileHeader()`:
- Remove `cursor bool` parameter
- Remove `if cursor { bgColor = theme.CursorBgFor(bgColor) }`
- Always use `theme.PanelHeaderBg` directly for header background

#### `tui/diff/formatter.go`

`RenderAlignedLine()`:
- Remove `cursor bool` parameter
- Pass to `renderStyledLine` without cursor

`renderStyledLine()`:
- Remove `cursor bool` parameter
- Remove `if cursor { numStyle = numStyle.Background(CursorBgFor(numBg)) }` — always use `numBg`
- Content rendering unchanged (already uses `bgColor` directly)

#### `tui/app/right_panel_mode.go`

`renderSinglePanel()`:
- Remove `cursorLine := model.scroller.CursorLine()`
- Remove `cursor := gi == cursorLine`
- Pass to `renderFileHeader` and `renderStyledLine` without cursor

`renderStyledLine()` (local copy):
- Remove `cursor bool` parameter
- Remove `if cursor` branches for content `baseStyle` and `numStyle` — always use `bgColor`/`numBg`
- Remove `else if cursor || bgColor != ""` branch — always use `bgColor` for content

#### `tui/widget/panel.go`

- Remove `cursorLine := p.scroller.CursorLine()`
- Remove `cursor := gi == cursorLine`
- Pass to `renderAlignedLine` and `renderFileHeader` without cursor

### 4. Theme (`tui/models/theme.go`)

- Remove `CursorBgFor(bg string) string` method
- Remove `blendHex(fg, bg string, ratio float64) string` function
- Keep `CursorBg string` field — used by `SelectionHighlighter`

### 5. Status bar (`tui/widget/statusbar.go`)

- Remove `cursorLine` and `totalLines` from the status format string
- Remove "↑↓ cursor" from help text, replace with "↑↓ scroll" or remove entirely
- Status bar no longer receives cursor position

### 6. Tests

**Remove**:
- `TestCursorAtEndOfFile` (view_test.go)
- `TestViewCursorScrolling` (view_test.go)
- `TestViewLayoutPreservesCursorWithinVisibleRange` (view_test.go)
- `TestCursorStopsAtReturnZero` (scroll_test.go)

**Adapt**:
- `TestScrollForDifferentHeights` — replace cursor walk with scroll walk: call MoveDown/MoveUp, assert scroll stays valid
- `TestScrollStallDetector` — same adaptation
- `TestRenderStyledLineCursorHighlight` — remove assertion on `CursorBgFor` blended color, test that cursor line uses normal `numBg`

**Update**:
- `strategies_test.go` — rename "underscore at cursor" test case name

### 7. What stays unchanged

- `CursorBg` theme field
- `SelectionHighlighter` — uses `CursorBg` directly
- `highlightColumns` / `stripBackgrounds` — selection rendering
- Scroll horizontal keys (h/l, arrows, `_`, `$`)
- Mouse wheel vertical/horizontal scroll
- Text selection (mouse drag)
- `scrollLock` mechanism for mouse wheel

## Files affected

| File | Action |
|------|--------|
| `tui/diff/scroller.go` | Remove `cursorLine`, refactor movement to scroll-direct |
| `tui/app/update.go` | Remove cursor movement, j/k become scroll |
| `tui/app/view.go` | Remove `cursor` boolean from rendering pipeline |
| `tui/app/model.go` | Remove `CursorLine()` calls if any remain |
| `tui/app/default_mode.go` | Remove cursorLine from status bar args |
| `tui/app/diff_only_mode.go` | Remove cursorLine from status bar args |
| `tui/app/right_panel_mode.go` | Remove cursor boolean from local rendering |
| `tui/diff/formatter.go` | Remove cursor boolean from renderStyledLine |
| `tui/widget/panel.go` | Remove cursor boolean |
| `tui/widget/statusbar.go` | Remove cursor display, update help text |
| `tui/models/theme.go` | Remove `CursorBgFor`, `blendHex` |
| `tui/app/view_test.go` | Remove cursor tests |
| `tui/app/scroll_test.go` | Adapt to scroll-direct |
| `tui/diff/background_test.go` | Update cursor highlight test |
