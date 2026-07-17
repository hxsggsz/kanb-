package update

import "testing"

func TestExtractErrorReason(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			name:   "single error line",
			output: "Installing kanba v1.2.0 (linux/amd64)...\nError: release asset not found: https://...\n",
			want:   "release asset not found: https://...",
		},
		{
			name:   "picks last error line among several",
			output: "Error: unsupported OS\nsome other output\nError: checksum verification failed. Aborting install.",
			want:   "checksum verification failed. Aborting install.",
		},
		{
			name:   "no error line falls back to generic message",
			output: "curl: command not found",
			want:   "install script exited with an error",
		},
		{
			name:   "empty output falls back to generic message",
			output: "",
			want:   "install script exited with an error",
		},
	}

	for _, tt := range tests {
		got := extractErrorReason(tt.output)
		if got != tt.want {
			t.Errorf("%s: extractErrorReason() = %q, want %q", tt.name, got, tt.want)
		}
	}
}
