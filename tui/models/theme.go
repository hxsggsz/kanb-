package models

import (
	"fmt"
	"sort"

	"kanba/git"

	"github.com/alecthomas/chroma/v2"
)

type Theme struct {
	Name       string
	Appearance string

	TokenColors map[chroma.TokenType]string

	AddedBg           string
	RemovedBg         string
	ModifiedAddedBg   string
	ModifiedRemovedBg string

	PanelBg       string
	PanelHeaderBg string
	LineNumberBg  string

	AddedSign   string
	RemovedSign string
	ContextFg   string

	CursorBg string

	SidebarBg       string
	SidebarSelected string
	SidebarAdded    string
	SidebarDeleted  string
	SidebarModified string
	SidebarDir      string

	SurfaceBg   string
	StatusBarFg string

	ErrorFg   string
	LoadingFg string
}

var Themes = map[string]Theme{}

func init() {
	for id, chrome := range uiThemes {
		palette, ok := syntaxPalettes[id]
		if !ok {
			panic(fmt.Sprintf("theme %q defined in uiThemes has no matching syntaxPalettes entry", id))
		}
		chrome.TokenColors = resolveSyntaxColors(palette)
		Themes[id] = chrome
	}
	for id := range syntaxPalettes {
		if _, ok := uiThemes[id]; !ok {
			panic(fmt.Sprintf("theme %q defined in syntaxPalettes has no matching uiThemes entry", id))
		}
	}
}

func GetTheme(name string) Theme {
	if t, ok := Themes[name]; ok {
		return t
	}
	return Themes["rose-pine"]
}

func SortedThemeKeys() []string {
	keys := make([]string, 0, len(Themes))
	for k := range Themes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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
