package models

import (
	"kanba/git"
	"sort"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/styles"
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

var Themes = map[string]Theme{}

func init() {
	for k, v := range handcrafted {
		Themes[k] = v
	}
	for _, name := range styles.Names() {
		if name == "swapoff" {
			continue
		}
		if _, ok := handcrafted[name]; ok {
			continue
		}
		Themes[name] = themeFromChromaStyle(name)
	}
}

var handcrafted = map[string]Theme{
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

func chromaStr(c chroma.Colour) string {
	if c.IsSet() {
		return c.String()
	}
	return ""
}

func accentFrom(s *chroma.Style, types ...chroma.TokenType) chroma.Colour {
	for _, t := range types {
		e := s.Get(t)
		if e.Colour.IsSet() {
			return e.Colour
		}
	}
	return chroma.Colour(0)
}

func blend(c1, c2 chroma.Colour, t float64) chroma.Colour {
	return chroma.NewColour(
		uint8(float64(c1.Red())*(1-t)+float64(c2.Red())*t),
		uint8(float64(c1.Green())*(1-t)+float64(c2.Green())*t),
		uint8(float64(c1.Blue())*(1-t)+float64(c2.Blue())*t),
	)
}

func luma(c chroma.Colour) float64 {
	return (0.299*float64(c.Red()) + 0.587*float64(c.Green()) + 0.114*float64(c.Blue())) / 255
}

func isDark(c chroma.Colour) bool {
	return luma(c) < 0.5
}

func themeFromChromaStyle(name string) Theme {
	s := styles.Get(name)

	bg := s.Get(chroma.Background).Background
	fg := s.Get(chroma.Text).Colour

	if !bg.IsSet() {
		bg = chroma.ParseColour("#1a1a1a")
	}
	if !fg.IsSet() {
		fg = chroma.ParseColour("#e0e0e0")
	}

	dark := isDark(bg)
	appearance := "dark"
	if !dark {
		appearance = "light"
	}

	headerBg := bg.BrightenOrDarken(1.08)
	lineNumBg := headerBg
	cursorBg := blend(bg, fg, 0.2)
	sidebarBg := bg
	surfaceBg := headerBg

	addedSign := accentFrom(s, chroma.LiteralString, chroma.LiteralStringAffix)
	if !addedSign.IsSet() {
		if dark {
			addedSign = chroma.ParseColour("#7ec699")
		} else {
			addedSign = chroma.ParseColour("#388e3c")
		}
	}

	removedSign := accentFrom(s, chroma.LiteralNumber, chroma.NameException)
	if !removedSign.IsSet() {
		if dark {
			removedSign = chroma.ParseColour("#cc6666")
		} else {
			removedSign = chroma.ParseColour("#d32f2f")
		}
	}

	sidebarSelected := accentFrom(s, chroma.Keyword, chroma.KeywordType, chroma.NameFunction)
	if !sidebarSelected.IsSet() {
		if dark {
			sidebarSelected = chroma.ParseColour("#c678dd")
		} else {
			sidebarSelected = chroma.ParseColour("#7b1fa2")
		}
	}

	loadingFg := sidebarSelected

	sidebarAdded := addedSign
	sidebarDeleted := removedSign

	sidebarModified := accentFrom(s, chroma.NameFunction, chroma.NameAttribute)
	if !sidebarModified.IsSet() {
		if dark {
			sidebarModified = chroma.ParseColour("#e5c07b")
		} else {
			sidebarModified = chroma.ParseColour("#ef6c00")
		}
	}

	sidebarDir := blend(fg, bg, 0.5)
	if !isDark(sidebarDir) && dark {
		sidebarDir = sidebarDir.BrightenOrDarken(0.7)
	} else if isDark(sidebarDir) && !dark {
		sidebarDir = sidebarDir.BrightenOrDarken(1.3)
	}

	contextFg := fg

	greenMix := chroma.ParseColour("#00ff00")
	redMix := chroma.ParseColour("#ff0000")

	addWeight := 0.08
	if !dark {
		addWeight = 0.12
	}
	addedBg := blend(bg, greenMix, addWeight)
	removedBg := blend(bg, redMix, addWeight)
	modifiedAddedBg := blend(bg, greenMix, addWeight*0.7)
	modifiedRemovedBg := blend(bg, redMix, addWeight*0.7)

	errorFg := removedSign
	statusBarFg := fg

	return Theme{
		Name:      name,
		ChromaKey: name,
		Appearance: appearance,

		AddedBg:           chromaStr(addedBg),
		RemovedBg:         chromaStr(removedBg),
		ModifiedAddedBg:   chromaStr(modifiedAddedBg),
		ModifiedRemovedBg: chromaStr(modifiedRemovedBg),

		PanelBg:       chromaStr(bg),
		PanelHeaderBg: chromaStr(headerBg),
		LineNumberBg:  chromaStr(lineNumBg),

		AddedSign:   chromaStr(addedSign),
		RemovedSign: chromaStr(removedSign),
		ContextFg:   chromaStr(contextFg),

		CursorBg: chromaStr(cursorBg),

		SidebarBg:       chromaStr(sidebarBg),
		SidebarSelected: chromaStr(sidebarSelected),
		SidebarAdded:    chromaStr(sidebarAdded),
		SidebarDeleted:  chromaStr(sidebarDeleted),
		SidebarModified: chromaStr(sidebarModified),
		SidebarDir:      chromaStr(sidebarDir),

		SurfaceBg:  chromaStr(surfaceBg),
		StatusBarFg: chromaStr(statusBarFg),

		ErrorFg:   chromaStr(errorFg),
		LoadingFg: chromaStr(loadingFg),
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
