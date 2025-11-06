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

func (q *Queue[T]) Get(index int) T {
	var zero T
	if index < 0 || index >= len(q.items) {
		return zero
	}
	return q.items[index]
}

func (q *Queue[T]) Set(index int, value T) {
	if index >= 0 && index < len(q.items) {
		q.items[index] = value
	}
}

func (q *Queue[T]) Push(item T) {
	q.items = append(q.items, item)
}

func (q *Queue[T]) IncrementNumber(index int) {
	if index >= 0 && index < len(q.items) {
		// This only works for numeric types - for int specifically
		// We'll need to use type assertion
		if v, ok := any(q.items[index]).(int); ok {
			q.items[index] = any(v + 1).(T)
		}
	}
}

func (q *Queue[T]) IncrementLast() {
	if len(q.items) > 0 {
		q.IncrementNumber(len(q.items) - 1)
	}
}
