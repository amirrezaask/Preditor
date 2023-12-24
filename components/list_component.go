package components

type ListComponent[T any] struct {
	VisibleStart int
	Items        []T
	Selection    int
}

func (l *ListComponent[T]) NextItem() {
	l.Selection++
	if l.Selection >= len(l.Items) {
		l.Selection = len(l.Items) - 1
	}

}

func (l *ListComponent[T]) PrevItem() {
	l.Selection--
	if l.Selection < 0 {
		l.Selection = 0
	}

	if l.Selection < l.VisibleStart {
		l.VisibleStart--
		if l.VisibleStart < 0 {
			l.VisibleStart = 0
		}
	}

}
func (l *ListComponent[T]) Scroll(n int) {
	l.VisibleStart += n

	if l.VisibleStart < 0 {
		l.VisibleStart = 0
	}

}
func (l *ListComponent[T]) VisibleView(maxLine int) []T {
	if l.Selection < l.VisibleStart {
		l.VisibleStart -= maxLine / 3
		if l.VisibleStart < 0 {
			l.VisibleStart = 0
		}
	}

	if l.Selection >= l.VisibleStart+maxLine {
		l.VisibleStart += maxLine / 3
		if l.VisibleStart >= len(l.Items) {
			l.VisibleStart = len(l.Items)
		}
	}

	if len(l.Items) > l.VisibleStart+maxLine {
		return l.Items[l.VisibleStart : l.VisibleStart+maxLine]
	} else {
		return l.Items[l.VisibleStart:len(l.Items)]
	}
}
