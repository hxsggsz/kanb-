package tui

import "kanba/git"

type Theme struct {
	// Diff backgrounds
	AddedBg           string
	RemovedBg         string
	ModifiedAddedBg   string
	ModifiedRemovedBg string

	// Panel backgrounds
	PanelBg       string
	LineNumberBg  string

	// Diff foregrounds (line numbers, signs)
	AddedSign   string
	RemovedSign string
	ContextFg   string

	// Cursor highlight
	CursorBg string

	// Sidebar
	SidebarSelected string
	SidebarAdded    string
	SidebarDeleted  string
	SidebarModified string
	SidebarDir      string

	// Status bar
	StatusBarBg string
	StatusBarFg string

	// General UI
	ErrorFg     string
	LoadingFg   string
	BorderColor string
}

var RosePine = Theme{
	AddedBg:           "#333c48",
	RemovedBg:         "#312e3f",
	ModifiedAddedBg:   "#333c48",
	ModifiedRemovedBg: "#312e3f",

	PanelBg:      "#1f1d2e",
	LineNumberBg: "#26233a",

	AddedSign:   "#9ccfd8",
	RemovedSign: "#908caa",
	ContextFg:   "#6e6a86",
	CursorBg:    "#403d52",

	SidebarSelected: "#c4a7e7",
	SidebarAdded:    "#9ccfd8",
	SidebarDeleted:  "#eb6f92",
	SidebarModified: "#f6c177",
	SidebarDir:      "#6e6a86",

	StatusBarBg: "#26233a",
	StatusBarFg: "#e0def4",

	ErrorFg:     "#eb6f92",
	LoadingFg:   "#c4a7e7",
	BorderColor: "#e0def4",
}

func (t Theme) BgFor(kind git.LineKind, isLeft bool) string {
	switch kind {
	case git.KindAdded:
		if isLeft {
			return ""
		}
		return t.AddedBg
	case git.KindDeleted:
		if isLeft {
			return t.RemovedBg
		}
		return ""
	case git.KindModified:
		if isLeft {
			return t.ModifiedRemovedBg
		}
		return t.ModifiedAddedBg
	default:
		return ""
	}
}

func (t Theme) LineNumFg(kind git.LineKind, isLeft bool) string {
	switch {
	case kind == git.KindContext:
		return t.ContextFg
	case (kind == git.KindAdded || kind == git.KindModified) && !isLeft:
		return t.AddedSign
	case (kind == git.KindDeleted || kind == git.KindModified) && isLeft:
		return t.RemovedSign
	default:
		return ""
	}
}

func (t Theme) CursorBgFor(bg string) string {
	if bg == "" {
		return t.CursorBg
	}
	return blendHex(t.CursorBg, bg, 0.75)
}
