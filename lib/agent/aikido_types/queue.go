package aikido_types

type Queue[T any] struct {
	items   []T
	maxSize int
}

func NewQueue[T any](maxSize int) Queue[T] {
	// Passing 0 as maxSize means no limit on the queue size.
	return Queue[T]{
		items:   []T{},
		maxSize: maxSize,
	}
}

func (q *Queue[T]) Clear() {
	q.items = []T{}
}

func (q *Queue[T]) PushAndGetRemovedItemIfMaxExceeded(item T) *T {
	var oldest *T
	if q.maxSize > 0 && q.Length() >= q.maxSize {
		temp := q.Pop()
		oldest = &temp
	}
	q.items = append(q.items, item)
	return oldest
}

func (q *Queue[T]) Pop() T {
	var zero T
	if len(q.items) == 0 {
		return zero
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item
}

func (q *Queue[T]) IsEmpty() bool {
	return q.Length() == 0
}

func (q *Queue[T]) Length() int {
	return len(q.items)
}
