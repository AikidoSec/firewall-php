package aikido_types

type RateLimitingQueue struct {
	items []int
}

func (q *RateLimitingQueue) Push(item int) {
	q.items = append(q.items, item)
}

func (q *RateLimitingQueue) Pop() int {
	if len(q.items) == 0 {
		return -1
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item
}

func (q *RateLimitingQueue) IsEmpty() bool {
	return q.Length() == 0
}

func (q *RateLimitingQueue) IncrementLast() {
	if q.IsEmpty() {
		return
	}
	q.items[q.Length()-1]++
}

func (q *RateLimitingQueue) Length() int {
	return len(q.items)
}
