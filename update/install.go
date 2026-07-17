package update

import (
	"errors"
	"os/exec"
	"strings"
)

const installScriptURL = "https://raw.githubusercontent.com/hxsggsz/kanba/main/install.sh"

// Run downloads and executes install.sh to update kanba to the latest
// release. On failure it returns a short error built from the script's
// output rather than the raw (possibly multi-line) combined output.
func Run() error {
	cmd := exec.Command("bash", "-c", "curl -fsSL "+installScriptURL+" | bash")
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	return errors.New(extractErrorReason(string(out)))
}

func extractErrorReason(output string) string {
	lines := strings.Split(output, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if reason, ok := strings.CutPrefix(line, "Error:"); ok {
			reason = strings.TrimSpace(reason)
			if reason != "" {
				return reason
			}
		}
	}
	return "install script exited with an error"
}
