package models

import (
	"kanba/git"
	"sort"
)

type Theme struct {
	Name      string
	ChromaKey string
	Appearance string

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

	SurfaceBg    string
	StatusBarFg string

	ErrorFg      string
	LoadingFg    string


}

var Themes = map[string]Theme{
	"rose-pine": {
		Name:      "Rosé Pine",
		ChromaKey: "rose-pine",
		Appearance: "dark",
		AddedBg:           "#333c48",
		RemovedBg:         "#43293a",
		ModifiedAddedBg:   "#333c48",
		ModifiedRemovedBg: "#43293a",
		PanelBg:       "#1f1d2e",
		PanelHeaderBg: "#26233a",
		LineNumberBg:  "#26233a",
		AddedSign:    "#9ccfd8",
		RemovedSign:  "#eb6f92",
		ContextFg:    "#e0def4",
		CursorBg:     "#403d52",
		SidebarBg:       "#191724",
		SidebarSelected: "#c4a7e7",
		SidebarAdded:    "#9ccfd8",
		SidebarDeleted:  "#eb6f92",
		SidebarModified: "#f6c177",
		SidebarDir:      "#6e6a86",
		SurfaceBg: "#26233a",
		StatusBarFg: "#e0def4",
		ErrorFg:      "#eb6f92",
		LoadingFg:    "#c4a7e7",
	},
	"rose-pine-moon": {
		Name:      "Rosé Pine Moon",
		ChromaKey: "rose-pine-moon",
		Appearance: "dark",
		AddedBg:           "#3b4456",
		RemovedBg:         "#4b3148",
		ModifiedAddedBg:   "#3b4456",
		ModifiedRemovedBg: "#4b3148",
		PanelBg:       "#2a273f",
		PanelHeaderBg: "#393552",
		LineNumberBg:  "#393552",
		AddedSign:    "#9ccfd8",
		RemovedSign:  "#eb6f92",
		ContextFg:    "#e0def4",
		CursorBg:     "#44415a",
		SidebarBg:       "#232136",
		SidebarSelected: "#c4a7e7",
		SidebarAdded:    "#9ccfd8",
		SidebarDeleted:  "#eb6f92",
		SidebarModified: "#f6c177",
		SidebarDir:      "#6e6a86",
		SurfaceBg: "#393552",
		StatusBarFg: "#e0def4",
		ErrorFg:      "#eb6f92",
		LoadingFg:    "#c4a7e7",
	},
	"rose-pine-dawn": {
		Name:      "Rosé Pine Dawn",
		ChromaKey: "rose-pine-dawn",
		Appearance: "light",
		AddedBg:           "#e6e8e4",
		RemovedBg:         "#f2e3df",
		ModifiedAddedBg:   "#e6e8e4",
		ModifiedRemovedBg: "#f2e3df",
		PanelBg:       "#fffaf3",
		PanelHeaderBg: "#f2e9de",
		LineNumberBg:  "#f2e9de",
		AddedSign:    "#56949f",
		RemovedSign:  "#b4637a",
		ContextFg:    "#575279",
		CursorBg:     "#dfdad9",
		SidebarBg:       "#faf4ed",
		SidebarSelected: "#907aa9",
		SidebarAdded:    "#56949f",
		SidebarDeleted:  "#b4637a",
		SidebarModified: "#ea9d34",
		SidebarDir:      "#6e6a86",
		SurfaceBg: "#f2e9de",
		StatusBarFg: "#575279",
		ErrorFg:      "#b4637a",
		LoadingFg:    "#907aa9",
	},
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


