package tui

import (
	"strings"
	"testing"
)

func round(f float64) float64 {
	return float64(int(f*100)) / 100
}

func TestParseHex(t *testing.T) {
	c, err := ParseColor("#FF0000")
	if err != nil {
		t.Fatal(err)
	}
	if round(c.H) != 0 || round(c.S) != 100 || round(c.L) != 50 {
		t.Fatalf("expected H=0 S=100 L=50, got H=%.0f S=%.0f L=%.0f", c.H, c.S, c.L)
	}
}

func TestParseShortHex(t *testing.T) {
	c, err := ParseColor("#F00")
	if err != nil {
		t.Fatal(err)
	}
	if round(c.H) != 0 || round(c.S) != 100 || round(c.L) != 50 {
		t.Fatalf("expected H=0 S=100 L=50, got H=%.0f S=%.0f L=%.0f", c.H, c.S, c.L)
	}
}

func TestParseRGB(t *testing.T) {
	c, err := ParseColor("rgb(255, 0, 0)")
	if err != nil {
		t.Fatal(err)
	}
	if round(c.H) != 0 || round(c.S) != 100 || round(c.L) != 50 {
		t.Fatalf("expected H=0 S=100 L=50, got H=%.0f S=%.0f L=%.0f", c.H, c.S, c.L)
	}
}

func TestParseHSL(t *testing.T) {
	c, err := ParseColor("hsl(120, 100%, 50%)")
	if err != nil {
		t.Fatal(err)
	}
	if round(c.H) != 120 || round(c.S) != 100 || round(c.L) != 50 {
		t.Fatalf("expected H=120 S=100 L=50, got H=%.0f S=%.0f L=%.0f", c.H, c.S, c.L)
	}
}

func TestHexRoundTrip(t *testing.T) {
	c, _ := ParseColor("#3366CC")
	hex := c.Hex()
	if !strings.EqualFold(hex, "#3366CC") {
		t.Fatalf("expected #3366CC, got %s", hex)
	}
}

func TestDarken(t *testing.T) {
	c, _ := ParseColor("#00FF00")
	dark := c.Darken(0.2)
	if dark.L >= c.L {
		t.Fatalf("expected darkened L < original L (%.0f < %.0f)", dark.L, c.L)
	}
	if dark.H != c.H {
		t.Fatalf("expected hue preserved (%.0f == %.0f)", dark.H, c.H)
	}
}

func TestBrighten(t *testing.T) {
	c, _ := ParseColor("#FF0000")
	bright := c.Brighten(0.7)
	if bright.L <= c.L {
		t.Fatalf("expected brightened L > original L (%.0f > %.0f)", bright.L, c.L)
	}
	if bright.H != c.H {
		t.Fatalf("expected hue preserved (%.0f == %.0f)", bright.H, c.H)
	}
}

func TestDarkenPreservesHue(t *testing.T) {
	c, _ := ParseColor("hsl(200, 80%, 50%)")
	dark := c.Darken(0.2)
	if round(dark.H) != 200 {
		t.Fatalf("expected H=200, got H=%.0f", dark.H)
	}
}

func TestBrightenPreservesHue(t *testing.T) {
	c, _ := ParseColor("hsl(300, 60%, 40%)")
	bright := c.Brighten(0.6)
	if round(bright.H) != 300 {
		t.Fatalf("expected H=300, got H=%.0f", bright.H)
	}
}
