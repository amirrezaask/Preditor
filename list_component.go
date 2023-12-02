package preditor

type ListComponent[T any] struct {
	VisibleStart int
	VisibleEnd   int
	MaxLine      int
	Items        []T
	Selection    int
}

func (l *ListComponent[T]) NextItem() {
	l.Selection++
	if l.Selection >= len(l.Items) {
		l.Selection = len(l.Items) - 1
	}

	if l.Selection > l.VisibleEnd {
		l.VisibleEnd++
		l.VisibleStart++
		if l.VisibleEnd >= len(l.Items)-1 {
			l.VisibleEnd = len(l.Items) - 1
		}
	}
}
func (l *ListComponent[T]) PrevItem() {
	l.Selection--
	if l.Selection < 0 {
		l.Selection = 0
	}

	if l.Selection < l.VisibleStart {
		l.VisibleStart--
		l.VisibleEnd--
		if l.VisibleStart < 0 {
			l.VisibleStart = 0
		}
	}

}
func (l *ListComponent[T]) Scroll(n int) {
	l.VisibleStart += n
	l.VisibleEnd += n

	if int(l.VisibleEnd) >= len(l.Items) {
		l.VisibleEnd = len(l.Items) - 1
		l.VisibleStart = l.VisibleEnd - l.MaxLine
	}

	if l.VisibleStart < 0 {
		l.VisibleStart = 0
		l.VisibleEnd = l.MaxLine
	}
	if l.VisibleEnd < 0 {
		l.VisibleStart = 0
		l.VisibleEnd = l.MaxLine
	}

}
func (l *ListComponent[T]) VisibleView() []T {
	if l.VisibleEnd >= len(l.Items) {
		return nil
	}
	return l.Items[l.VisibleStart:l.VisibleEnd]
}
