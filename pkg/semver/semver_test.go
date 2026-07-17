package semver

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		in      string
		want    Version
		wantErr bool
	}{
		{"v1.2.3", Version{1, 2, 3}, false},
		{"1.2.3", Version{1, 2, 3}, false},
		{"v0.0.1", Version{0, 0, 1}, false},
		{"v1.2", Version{}, true},
		{"v1.2.x", Version{}, true},
		{"", Version{}, true},
		{"vx", Version{}, true},
	}

	for _, tt := range tests {
		got, err := Parse(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Errorf("Parse(%q): expected error, got none", tt.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("Parse(%q): unexpected error: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("Parse(%q) = %+v, want %+v", tt.in, got, tt.want)
		}
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		a, b    string
		want    int
		wantErr bool
	}{
		{"v1.0.0", "v1.0.0", 0, false},
		{"v1.0.0", "v1.0.1", -1, false},
		{"v1.0.1", "v1.0.0", 1, false},
		{"v1.1.0", "v1.0.9", 1, false},
		{"v2.0.0", "v1.9.9", 1, false},
		{"v1.0.0", "v2.0.0", -1, false},
		{"v1.2.3", "1.2.3", 0, false},
		{"bogus", "v1.0.0", 0, true},
		{"v1.0.0", "bogus", 0, true},
	}

	for _, tt := range tests {
		got, err := Compare(tt.a, tt.b)
		if tt.wantErr {
			if err == nil {
				t.Errorf("Compare(%q, %q): expected error, got none", tt.a, tt.b)
			}
			continue
		}
		if err != nil {
			t.Errorf("Compare(%q, %q): unexpected error: %v", tt.a, tt.b, err)
			continue
		}
		if got != tt.want {
			t.Errorf("Compare(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}
