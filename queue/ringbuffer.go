package queue

type ringBuffer[T any] struct {
	items []T
	first int
	count int
}

func (b *ringBuffer[T]) isEmpty() bool {
	return b.count == 0
}

func (b *ringBuffer[T]) push(item T) {
	if b.count == len(b.items) {
		// Grow buffer if nothing free
		newBuffer := make([]T, len(b.items)*2)
		copy(newBuffer, b.items[b.first:])
		copy(newBuffer[b.count-b.first:], b.items[:b.first])
		b.items = newBuffer
		b.first = 0
	}

	i := (b.first + b.count) % len(b.items)
	b.items[i] = item
	b.count++
}

func (b *ringBuffer[T]) shift() T {
	if b.count == 0 {
		panic("empty buffer")
	}
	b.count--
	i := b.first
	b.first = (b.first + 1) % len(b.items)
	return b.items[i]
}
