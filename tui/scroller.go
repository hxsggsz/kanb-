package tui

const scrollMargin = 8

type Scroller struct {
	scroll     int
	cursorLine int
}

func NewScroller() *Scroller {
	return &Scroller{}
}

func (s *Scroller) CursorLine() int { return s.cursorLine }
func (s *Scroller) Scroll() int     { return s.scroll }

func (s *Scroller) MoveDown(total int) {
	if s.cursorLine < total-1 {
		s.cursorLine++
	}
}

func (s *Scroller) MoveUp() {
	if s.cursorLine > 0 {
		s.cursorLine--
	}
}

func (s *Scroller) GoToTop() {
	s.cursorLine = 0
}

func (s *Scroller) GoToBottom(total int) {
	s.cursorLine = total - 1
}

func (s *Scroller) SetFile() {
	s.cursorLine = 0
	s.scroll = 0
}

func (s *Scroller) UpdateScroll(total int, vis int) {
	if total == 0 || vis <= 0 {
		s.scroll = 0
		return
	}
	if s.cursorLine >= total {
		s.cursorLine = total - 1
	}
	if vis >= total {
		s.scroll = 0
		return
	}
	sm := scrollMargin
	if sm > vis/2 {
		sm = max(1, vis/2)
	}
	maxScroll := total - vis
	if s.cursorLine < s.scroll+sm {
		s.scroll = max(0, s.cursorLine-sm)
	}
	if s.cursorLine >= s.scroll+vis-sm {
		s.scroll = min(s.cursorLine-vis+sm+1, maxScroll)
	}
	if s.scroll > maxScroll {
		s.scroll = maxScroll
	}
}
