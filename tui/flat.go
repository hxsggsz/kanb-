package tui

import "kanba/git"

type flatLine struct {
	isHeader bool
	fileIdx  int
	hunkIdx  int
	lineIdx  int
}

type fileStat struct {
	added   int
	deleted int
}

func buildFlatLines(diffs []git.SideBySideDiff) []flatLine {
	var lines []flatLine
	for fi := range diffs {
		lines = append(lines, flatLine{isHeader: true, fileIdx: fi})
		for hi, h := range diffs[fi].Hunks {
			for li := range h.Lines {
				lines = append(lines, flatLine{fileIdx: fi, hunkIdx: hi, lineIdx: li})
			}
		}
	}
	return lines
}

func computeFileStats(diffs []git.SideBySideDiff) []fileStat {
	stats := make([]fileStat, len(diffs))
	for fi, f := range diffs {
		for _, h := range f.Hunks {
			for _, ln := range h.Lines {
				switch ln.Kind {
				case git.KindAdded:
					stats[fi].added++
				case git.KindDeleted:
					stats[fi].deleted++
				case git.KindModified:
					stats[fi].added++
					stats[fi].deleted++
				}
			}
		}
	}
	return stats
}
