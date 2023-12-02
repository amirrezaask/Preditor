package preditor

type ListComponent[T any] struct {
	VisibleStart   int
	VisibleEnd     int
	MaxLineVisible int
	Items          []T
	Selection      int
}

func NewListComponent[T any]() *ListComponent[T] {
	return &ListComponent[T]{}
}

func (l *ListComponent[T]) NextItem() {}
func (l *ListComponent[T]) PrevItem() {}
func (l *ListComponent[T]) Scroll(n int) {
	l.VisibleStart += n
	l.VisibleEnd += n

	if int(l.VisibleEnd) >= len(l.Items) {
		l.VisibleEnd = len(l.Items) - 1
		l.VisibleStart = l.VisibleEnd - l.MaxLineVisible
	}

	if l.VisibleStart < 0 {
		l.VisibleStart = 0
		l.VisibleEnd = l.MaxLineVisible
	}
	if l.VisibleEnd < 0 {
		l.VisibleStart = 0
		l.VisibleEnd = l.MaxLineVisible
	}

}
func (l *ListComponent[T]) VisibleView() []T {
	return l.Items[l.VisibleStart:l.VisibleEnd]
}
