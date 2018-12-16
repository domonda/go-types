package queue

import "sync"

type Queue interface {
	Add(items ...interface{})
	Next() <-chan interface{}
	Close()
}

func New() Queue {
	q := &queue{
		buffer: buffer{
			items: make([]interface{}, 16),
		},
		channel: make(chan interface{}, 16),
	}
	q.bufferCond = sync.NewCond(&q.mutex)
	go q.channelPump()
	return q
}

type queue struct {
	mutex      sync.Mutex
	bufferCond *sync.Cond
	channel    chan interface{}
	buffer     buffer
	closed     bool
}

func (q *queue) channelPump() {
	for {
		q.bufferCond.Wait()
		if q.closed {
			return
		}

		q.mutex.Lock()
		q.fillChanFromBuffer()
		q.mutex.Unlock()
	}
}

func (q *queue) fillChanFromBuffer() {
	for !q.buffer.isEmpty() && len(q.channel) < cap(q.channel) {
		q.channel <- q.buffer.shift()
	}
}

func (q *queue) Add(items ...interface{}) {
	if len(items) == 0 {
		return
	}

	q.mutex.Lock()
	defer q.mutex.Unlock()

	// While locked fill the channel so channelPump() may not have to wake up
	q.fillChanFromBuffer()

	freeInChannel := cap(q.channel) - len(q.channel)
	if q.buffer.isEmpty() && freeInChannel > 0 {
		// If buffer is empty and still place in channel,
		// take shortcut and send directly to channel
		if freeInChannel >= len(items) {
			// Send all to channel and be done with it
			for _, item := range items {
				q.channel <- item
			}
			return
		}

		// Send only as many items as are fitting into channel
		for _, item := range items[:freeInChannel] {
			q.channel <- item
		}
		// Buffer the rest
		items = items[freeInChannel:]
	}

	// Push items on buffer
	for _, item := range items {
		q.buffer.push(item)
	}
	// and signal for channelPump()
	q.bufferCond.Signal()
}

func (q *queue) Next() <-chan interface{} {
	return q.channel
}

func (q *queue) Close() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.closed = true
	q.bufferCond.Signal()
	close(q.channel)
}
