package app

import (
	"fmt"
	"strings"

	"kanba/tui/diff"
	"kanba/tui/models"
	"kanba/tui/widget"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"kanba/git"
)

type RightPanelMode struct{}

func (m *RightPanelMode) Type() ModeType { return ModeRightPanel }

func (m *RightPanelMode) Render(model *Model) string {
	if len(model.flatLines) == 0 {
		return ""
	}

	theme := model.CurrentTheme()
	contentVis := model.VisibleLines()
	panelWidth := max(model.width-panelBorderWidth, panelMinWidth)
	content := m.renderSinglePanel(model, panelWidth, contentVis)

	scroll := model.scroller.Scroll()
	if scroll >= len(model.flatLines) {
		scroll = max(0, len(model.flatLines)-1)
	}
	cursorFileIdx := model.flatLines[scroll].FileIdx
	f := model.diffs[cursorFileIdx]
	statusBar := widget.NewStatusBar(f.NewPath, cursorFileIdx, len(model.diffs), model.width, theme)

	result := fmt.Sprintf("%s\n%s", statusBar.Render(), content)
	result = lipgloss.NewStyle().Background(lipgloss.Color(theme.PanelBg)).Render(result)
	result = model.themeModal.Overlay(result, theme.PanelBg, theme.SidebarSelected, theme.ContextFg, 0, panelWidth)

	if model.helpActive {
		result = model.helpOverlay(result, theme, 0, panelWidth)
	}

	return result
}

func (m *RightPanelMode) HandleInput(model *Model, msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	return model.handleDiffKeys(msg)
}

func (m *RightPanelMode) renderSinglePanel(model *Model, width int, vis int) string {
	total := len(model.flatLines)
	if total == 0 {
		return ""
	}
	if vis <= 0 {
		model.scroller.UpdateScroll(total, vis)
		return ""
	}

	theme := model.CurrentTheme()
	model.scroller.UpdateScroll(total, vis)
	hScroll := model.scroller.HScroll()

	start := model.scroller.Scroll()
	end := min(start+vis, total)

	contentAreaWidth := width - (diff.LineNumColWidth + 3)
	if contentAreaWidth < 0 {
		contentAreaWidth = 0
	}

	end, needsBorder := model.reserveLastPanelBorder(start, end)

	var lines []string
	for gi := start; gi < end; gi++ {
		fl := model.flatLines[gi]

		if fl.IsHeader {
			lines = append(lines, model.renderFileHeader(fl, width))
		} else {
			f := model.diffs[fl.FileIdx]
			h := f.Hunks[fl.HunkIdx]
			ln := h.Lines[fl.LineIdx]

			newNum := ""
			if ln.NewLineNum > 0 {
				newNum = fmt.Sprintf("%d", ln.NewLineNum)
			}

			var prefix, content string
			var kind git.LineKind
			var isLeft bool

			switch ln.Kind {
			case git.KindAdded:
				prefix = "+"
				content = ln.NewContent
				kind = git.KindAdded
				isLeft = false
			case git.KindDeleted:
				prefix = "-"
				content = ln.OldContent
				kind = git.KindDeleted
				isLeft = true
			case git.KindModified:
				prefix = "+"
				content = ln.NewContent
				kind = git.KindAdded
				isLeft = false
			default:
				prefix = " "
				content = ln.NewContent
				kind = git.KindContext
				isLeft = false
			}

			content = ansi.Cut(content, hScroll, hScroll+contentAreaWidth)

			prefixStr := fmt.Sprintf("%*s %s ", diff.LineNumColWidth, newNum, prefix)
			line := renderStyledLine(prefixStr, content, width, kind, isLeft, model.highlighter, f.NewPath, theme)
			lines = append(lines, line)
		}
	}

	if needsBorder {
		borderStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.SidebarDir)).
			Background(lipgloss.Color(theme.PanelBg)).
			Render(strings.Repeat("─", width))
		lines = append(lines, borderStyle)
		marginStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(theme.PanelBg)).
			Render(strings.Repeat(" ", width))
		lines = append(lines, marginStyle)
	}

	return strings.Join(lines, "\n")
}

func renderStyledLine(prefix, content string, width int, kind git.LineKind, isLeft bool, sh *diff.SyntaxHighlighter, filePath string, theme models.Theme) string {
	bgColor := theme.BgFor(kind, isLeft)
	if bgColor == "" {
		bgColor = theme.PanelBg
	}

	numBg := bgColor
	if kind == git.KindContext {
		numBg = theme.LineNumberBg
	}

	baseStyle := lipgloss.NewStyle().Background(lipgloss.Color(bgColor))

	numStyle := lipgloss.NewStyle()
	if fg := theme.LineNumFg(kind, isLeft); fg != "" {
		numStyle = numStyle.Foreground(lipgloss.Color(fg))
	}
	numStyle = numStyle.Background(lipgloss.Color(numBg))

	prefixRendered := numStyle.Render(prefix)

	var contentRendered string
	if sh != nil {
		contentRendered = sh.HighlightWithStyle(content, filePath, baseStyle, theme)
	} else if bgColor != "" {
		contentRendered = baseStyle.Render(content)
	} else {
		contentRendered = content
	}

	styled := prefixRendered + contentRendered
	vis := lipgloss.Width(styled)
	if vis < width {
		padStyle := baseStyle
		styled += padStyle.Render(strings.Repeat(" ", width-vis))
	}

	return styled
}
