package models

import (
	"fmt"
	"kanba/git"
	"sort"
	"strings"
)

type Theme struct {
	Name      string
	ChromaKey string
	Appearance string

	AddedBg           string
	RemovedBg         string
	ModifiedAddedBg   string
	ModifiedRemovedBg string

	PanelBg      string
	LineNumberBg string

	AddedSign   string
	RemovedSign string
	ContextFg   string

	CursorBg string

	SidebarSelected string
	SidebarAdded    string
	SidebarDeleted  string
	SidebarModified string
	SidebarDir      string

	StatusBarBg string
	StatusBarFg string

	ErrorFg     string
	LoadingFg   string
	BorderColor string
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
		PanelBg:      "#1f1d2e",
		LineNumberBg: "#26233a",
		AddedSign:    "#9ccfd8",
		RemovedSign:  "#eb6f92",
		ContextFg:    "#6e6a86",
		CursorBg:     "#403d52",
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
	},
	"rose-pine-moon": {
		Name:      "Rosé Pine Moon",
		ChromaKey: "rose-pine-moon",
		Appearance: "dark",
		AddedBg:           "#3b4456",
		RemovedBg:         "#4b3148",
		ModifiedAddedBg:   "#3b4456",
		ModifiedRemovedBg: "#4b3148",
		PanelBg:      "#2a273f",
		LineNumberBg: "#393552",
		AddedSign:    "#9ccfd8",
		RemovedSign:  "#eb6f92",
		ContextFg:    "#6e6a86",
		CursorBg:     "#44415a",
		SidebarSelected: "#c4a7e7",
		SidebarAdded:    "#9ccfd8",
		SidebarDeleted:  "#eb6f92",
		SidebarModified: "#f6c177",
		SidebarDir:      "#6e6a86",
		StatusBarBg: "#393552",
		StatusBarFg: "#e0def4",
		ErrorFg:     "#eb6f92",
		LoadingFg:   "#c4a7e7",
		BorderColor: "#e0def4",
	},
	"rose-pine-dawn": {
		Name:      "Rosé Pine Dawn",
		ChromaKey: "rose-pine-dawn",
		Appearance: "light",
		AddedBg:           "#e6e8e4",
		RemovedBg:         "#f2e3df",
		ModifiedAddedBg:   "#e6e8e4",
		ModifiedRemovedBg: "#f2e3df",
		PanelBg:      "#fffaf3",
		LineNumberBg: "#f2e9de",
		AddedSign:    "#56949f",
		RemovedSign:  "#b4637a",
		ContextFg:    "#9893a5",
		CursorBg:     "#dfdad9",
		SidebarSelected: "#907aa9",
		SidebarAdded:    "#56949f",
		SidebarDeleted:  "#b4637a",
		SidebarModified: "#ea9d34",
		SidebarDir:      "#9893a5",
		StatusBarBg: "#f2e9de",
		StatusBarFg: "#575279",
		ErrorFg:     "#b4637a",
		LoadingFg:   "#907aa9",
		BorderColor: "#575279",
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

func (t Theme) CursorBgFor(bg string) string {
	if bg == "" {
		return t.CursorBg
	}
	return blendHex(t.CursorBg, bg, 0.75)
}

func blendHex(fg, bg string, ratio float64) string {
	fr, fgC, fb := parseHex(fg)
	br, bgC, bb := parseHex(bg)
	r := int(float64(br) + (float64(fr)-float64(br))*ratio + 0.5)
	g := int(float64(bgC) + (float64(fgC)-float64(bgC))*ratio + 0.5)
	b := int(float64(bb) + (float64(fb)-float64(bb))*ratio + 0.5)
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func parseHex(hex string) (r, g, b int) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	r = intHex(hex[0:2])
	g = intHex(hex[2:4])
	b = intHex(hex[4:6])
	return
}

func intHex(s string) int {
	var v int
	fmt.Sscanf(s, "%x", &v)
	return v
}
