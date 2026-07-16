package diff

const (
	hScrollStep     = 8
	hScrollFastStep = 32
	VScrollStep     = 3
)

type Scroller struct {
	scroll     int
	hScroll    int
	scrollLock bool
}

func NewScroller() *Scroller {
	return &Scroller{}
}

func (s *Scroller) Scroll() int  { return s.scroll }
func (s *Scroller) HScroll() int { return s.hScroll }

func (s *Scroller) MoveDown(total int, vis int) {
	maxScroll := max(total-vis, 0)
	if s.scroll < maxScroll {
		s.scroll = min(s.scroll+VScrollStep, maxScroll)
	}
}

func (s *Scroller) MoveUp() {
	if s.scroll > 0 {
		s.scroll = max(0, s.scroll-VScrollStep)
	}
}

func (s *Scroller) GoToTop() {
	s.scroll = 0
}

func (s *Scroller) GoToBottom(total int, vis int) {
	maxScroll := max(total - vis, 0)
	s.scroll = maxScroll
}

func (s *Scroller) ScrollLeft() {
	s.hScroll = max(0, s.hScroll-hScrollStep)
	s.scrollLock = true
}

func (s *Scroller) ScrollRight() {
	s.hScroll += hScrollStep
	s.scrollLock = true
}

func (s *Scroller) ScrollLeftFast() {
	s.hScroll = max(0, s.hScroll-hScrollFastStep)
	s.scrollLock = true
}

func (s *Scroller) ScrollRightFast() {
	s.hScroll += hScrollFastStep
	s.scrollLock = true
}

func (s *Scroller) ScrollHome() {
	s.hScroll = 0
	s.scrollLock = true
}

func (s *Scroller) ScrollEnd(maxScroll int) {
	s.hScroll = maxScroll
	s.scrollLock = true
}

func (s *Scroller) ClampHScroll(maxScroll int) {
	if s.hScroll > maxScroll {
		s.hScroll = max(0, maxScroll)
	}
}

func (s *Scroller) ScrollViewBy(delta int, total int, vis int) {
	if total <= 0 {
		return
	}
	maxScroll := max(total - vis, 0)
	s.scroll = max(0, min(s.scroll+delta, maxScroll))
	s.scrollLock = true
}

func (s *Scroller) SetScroll(pos, total, vis int) {
	if pos < 0 {
		pos = 0
	}
	maxScroll := max(total - vis, 0)
	if pos > maxScroll {
		pos = maxScroll
	}
	s.scroll = pos
	s.scrollLock = true
}

func (s *Scroller) UpdateScroll(total int, vis int) {
	if s.scrollLock {
		s.scrollLock = false
	}
	if total == 0 || vis <= 0 {
		s.scroll = 0
		return
	}
	maxScroll := max(total - vis, 0)
	if s.scroll > maxScroll {
		s.scroll = maxScroll
	}
	if s.scroll < 0 {
		s.scroll = 0
	}
}
