package queue

type ringBuffer struct {
	items []interface{}
	first int
	count int
}

func (b *ringBuffer) isEmpty() bool {
	return b.count == 0
}

func (b *ringBuffer) push(item interface{}) {
	if b.count == len(b.items) {
		// Grow buffer if nothing free
		newBuffer := make([]interface{}, len(b.items)*2)
		copy(newBuffer, b.items[b.first:])
		copy(newBuffer[b.count-b.first:], b.items[:b.first])
		b.items = newBuffer
		b.first = 0
	}

	i := (b.first + b.count) % len(b.items)
	b.items[i] = item
	b.count++
}

func (b *ringBuffer) shift() interface{} {
	if b.count == 0 {
		panic("empty buffer")
	}
	b.count--
	i := b.first
	b.first = (b.first + 1) % len(b.items)
	return b.items[i]
}
