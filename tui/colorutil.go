package tui

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Color struct {
	H, S, L float64
}

func ParseColor(s string) (Color, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "#") {
		return parseHex(s)
	}
	if strings.HasPrefix(s, "rgb") {
		return parseRGB(s)
	}
	if strings.HasPrefix(s, "hsl") {
		return parseHSL(s)
	}
	return Color{}, fmt.Errorf("unknown color format: %s", s)
}

func parseHex(s string) (Color, error) {
	s = strings.TrimPrefix(s, "#")
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 {
		return Color{}, fmt.Errorf("invalid hex color: #%s", s)
	}
	r, _ := strconv.ParseUint(s[0:2], 16, 8)
	g, _ := strconv.ParseUint(s[2:4], 16, 8)
	b, _ := strconv.ParseUint(s[4:6], 16, 8)
	return rgbToHSL(float64(r), float64(g), float64(b)), nil
}

func parseRGB(s string) (Color, error) {
	s = strings.TrimPrefix(s, "rgb(")
	s = strings.TrimSuffix(s, ")")
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ' '
	})
	if len(parts) != 3 {
		return Color{}, fmt.Errorf("invalid rgb format: %s", s)
	}
	r, _ := strconv.ParseFloat(parts[0], 64)
	g, _ := strconv.ParseFloat(parts[1], 64)
	b, _ := strconv.ParseFloat(parts[2], 64)
	return rgbToHSL(r, g, b), nil
}

func parseHSL(s string) (Color, error) {
	s = strings.TrimPrefix(s, "hsl(")
	s = strings.TrimSuffix(s, ")")
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ' '
	})
	if len(parts) != 3 {
		return Color{}, fmt.Errorf("invalid hsl format: %s", s)
	}
	h, _ := strconv.ParseFloat(parts[0], 64)
	sp := strings.TrimSuffix(parts[1], "%")
	lp := strings.TrimSuffix(parts[2], "%")
	sat, _ := strconv.ParseFloat(sp, 64)
	lit, _ := strconv.ParseFloat(lp, 64)
	return Color{H: h, S: sat, L: lit}, nil
}

func rgbToHSL(r, g, b float64) Color {
	r /= 255
	g /= 255
	b /= 255
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	l := (max + min) / 2

	if max == min {
		return Color{H: 0, S: 0, L: l * 100}
	}

	var h, s float64
	d := max - min
	if l > 0.5 {
		s = d / (2 - max - min)
	} else {
		s = d / (max + min)
	}
	switch max {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	case b:
		h = (r-g)/d + 4
	}
	h *= 60
	return Color{H: h, S: s * 100, L: l * 100}
}

func hslToRGB(c Color) (r, g, b float64) {
	h := c.H / 360
	s := c.S / 100
	l := c.L / 100

	if s == 0 {
		return l * 255, l * 255, l * 255
	}

	var hue2rgb func(p, q, t float64) float64
	hue2rgb = func(p, q, t float64) float64 {
		if t < 0 {
			t += 1
		}
		if t > 1 {
			t -= 1
		}
		if t < 1.0/6 {
			return p + (q-p)*6*t
		}
		if t < 1.0/2 {
			return q
		}
		if t < 2.0/3 {
			return p + (q-p)*(2.0/3-t)*6
		}
		return p
	}

	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q

	r = hue2rgb(p, q, h+1.0/3) * 255
	g = hue2rgb(p, q, h) * 255
	b = hue2rgb(p, q, h-1.0/3) * 255
	return
}

func (c Color) Hex() string {
	r, g, b := hslToRGB(c)
	return fmt.Sprintf("#%02X%02X%02X", clamp(r), clamp(g), clamp(b))
}

func (c Color) Darken(factor float64) Color {
	return Color{H: c.H, S: c.S, L: c.L * factor}
}

func (c Color) Brighten(factor float64) Color {
	return Color{H: c.H, S: c.S, L: c.L + (100-c.L)*factor}
}

func clamp(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(math.Round(v))
}
