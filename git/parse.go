package git

import (
	"fmt"
	"strconv"
	"strings"
)

type parseState int

const (
	stateHeader    parseState = iota
	stateOldPath
	stateNewPath
	stateHunk
	stateLine
)

const (
	diffLineParts    = 4 // "diff --git a/old b/new" splits into [diff, --git, a/old, b/new]
	oldPathIdx       = 2
	newPathIdx       = 3
	hunkParts        = 2 // "@@ coords @@ desc" splits into [coords, desc]
	lineNumBase      = 1 // unified diff line numbers are 1-based
	lineNumNone      = 0 // sentinel when a line has no old or new number
	defaultHunkCount = 1 // implied count when hunk header omits it (e.g. @@ -1 +1 @@)
)

func parseRawDiff(lines []string) ([]FileDiff, error) {
	var files []FileDiff
	var cur *FileDiff
	var state parseState

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "diff --git"):
			if cur != nil {
				files = append(files, *cur)
			}
			cur = &FileDiff{}
			parts := strings.SplitN(line, " ", diffLineParts)
			if len(parts) >= diffLineParts {
				cur.OldPath = strings.TrimPrefix(parts[oldPathIdx], "a/")
				cur.NewPath = strings.TrimPrefix(parts[newPathIdx], "b/")
			}
			state = stateOldPath

		case strings.HasPrefix(line, "--- "):
			state = stateNewPath

		case strings.HasPrefix(line, "+++ "):
			state = stateHunk

		case strings.HasPrefix(line, "@@"):
			h, err := parseHunkHeader(line)
			if err != nil {
				return nil, fmt.Errorf("parse hunk header: %w", err)
			}
			cur.Hunks = append(cur.Hunks, h)
			state = stateLine

		case strings.HasPrefix(line, "Binary files"):
			cur.IsBinary = true

		case strings.HasPrefix(line, "new file mode"):
			cur.IsNew = true
			cur.Status = "A"

		case strings.HasPrefix(line, "deleted file mode"):
			cur.IsDelete = true
			cur.Status = "D"

		case strings.HasPrefix(line, "rename from"):
			cur.IsRename = true
			cur.Status = "R"

		case strings.HasPrefix(line, "old mode") || strings.HasPrefix(line, "new mode"):

		case strings.HasPrefix(line, "similarity index") || strings.HasPrefix(line, "copy from") || strings.HasPrefix(line, "copy to"):

		case state == stateLine && len(line) > 0 && cur != nil && len(cur.Hunks) > 0:
			h := &cur.Hunks[len(cur.Hunks)-1]
			prefix := line[0]
			content := line[1:]

			oldNum, newNum := lineNumNone, lineNumNone
			if len(h.Lines) > 0 {
				last := h.Lines[len(h.Lines)-1]
				oldNum = last.OldLineNum
				newNum = last.NewLineNum
			} else {
				oldNum = h.OldStart - lineNumBase
				newNum = h.NewStart - lineNumBase
			}

			switch prefix {
			case ' ':
				oldNum++
				newNum++
				h.Lines = append(h.Lines, Line{Type: LineContext, OldLineNum: oldNum, NewLineNum: newNum, Content: content})
			case '+':
				newNum++
				h.Lines = append(h.Lines, Line{Type: LineAdded, OldLineNum: lineNumNone, NewLineNum: newNum, Content: content})
			case '-':
				oldNum++
				h.Lines = append(h.Lines, Line{Type: LineDeleted, OldLineNum: oldNum, NewLineNum: lineNumNone, Content: content})
			case '\\':
				if len(h.Lines) > 0 {
					last := &h.Lines[len(h.Lines)-1]
					last.Content += " " + strings.TrimPrefix(line, "\\ ")
				}
			}

			recalcLineNums(h)
		}
	}

	if cur != nil {
		files = append(files, *cur)
	}

	for i := range files {
		if files[i].Status == "" {
			files[i].Status = "M"
		}
	}

	return files, nil
}

func parseHunkHeader(line string) (Hunk, error) {
	var h Hunk
	line = strings.TrimPrefix(line, "@@ ")
	parts := strings.SplitN(line, " @@", hunkParts)
	if len(parts) < 1 {
		return h, fmt.Errorf("invalid hunk header: %s", line)
	}
	spaces := strings.Fields(parts[0])
	if len(spaces) < 2 {
		return h, fmt.Errorf("invalid hunk header: %s", line)
	}

	parsePart := func(s string) (start, count int, err error) {
		s = strings.TrimPrefix(s, "-")
		s = strings.TrimPrefix(s, "+")
		comma := strings.Index(s, ",")
		if comma < 0 {
			start, err = strconv.Atoi(s)
			count = defaultHunkCount
		} else {
			start, err = strconv.Atoi(s[:comma])
			if err == nil {
				count, err = strconv.Atoi(s[comma+1:])
			}
		}
		return
	}

	var err error
	h.OldStart, h.OldCount, err = parsePart(spaces[0])
	if err != nil {
		return h, fmt.Errorf("parse old position: %w", err)
	}
	h.NewStart, h.NewCount, err = parsePart(spaces[1])
	if err != nil {
		return h, fmt.Errorf("parse new position: %w", err)
	}

	h.Header = line
	return h, nil
}

func recalcLineNums(h *Hunk) {
	oldNum := h.OldStart - lineNumBase
	newNum := h.NewStart - lineNumBase
	for i := range h.Lines {
		ln := &h.Lines[i]
		switch ln.Type {
		case LineContext:
			oldNum++
			newNum++
			ln.OldLineNum = oldNum
			ln.NewLineNum = newNum
		case LineAdded:
			newNum++
			ln.OldLineNum = lineNumNone
			ln.NewLineNum = newNum
		case LineDeleted:
			oldNum++
			ln.OldLineNum = oldNum
			ln.NewLineNum = lineNumNone
		}
	}
}
