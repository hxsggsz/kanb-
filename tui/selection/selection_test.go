package selection

import "testing"

func TestRange_IsEmpty(t *testing.T) {
	tests := []struct {
		name  string
		r     Range
		empty bool
	}{
		{"both zero", Range{}, true},
		{"same position nonzero", Range{StartLine: 1, StartCol: 1, EndLine: 1, EndCol: 1}, true},
		{"different lines", Range{StartLine: 0, EndLine: 1}, false},
		{"same line different cols", Range{StartLine: 0, StartCol: 0, EndLine: 0, EndCol: 5}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.IsEmpty(); got != tt.empty {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.empty)
			}
		})
	}
}

func TestRange_Normalized(t *testing.T) {
	tests := []struct {
		name string
		r    Range
		want Range
	}{
		{
			"already forward",
			Range{StartLine: 0, StartCol: 0, EndLine: 1, EndCol: 0},
			Range{StartLine: 0, StartCol: 0, EndLine: 1, EndCol: 0},
		},
		{
			"reversed across lines",
			Range{StartLine: 2, StartCol: 0, EndLine: 0, EndCol: 0},
			Range{StartLine: 0, StartCol: 0, EndLine: 2, EndCol: 0},
		},
		{
			"reversed same line",
			Range{StartLine: 0, StartCol: 5, EndLine: 0, EndCol: 0},
			Range{StartLine: 0, StartCol: 0, EndLine: 0, EndCol: 5},
		},
		{
			"same position",
			Range{StartLine: 1, StartCol: 1, EndLine: 1, EndCol: 1},
			Range{StartLine: 1, StartCol: 1, EndLine: 1, EndCol: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.Normalized()
			if got != tt.want {
				t.Errorf("Normalized() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
