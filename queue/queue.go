package queue

import "sync"

var (
	ChanLen           = 32
	InitialBufferSize = 8
)

type Queue interface {
	Add(items ...interface{})
	Next() <-chan interface{}
	Len() int
	Cap() int
	Close()
}

func New() Queue {
	q := &queue{
		channel: make(chan interface{}, ChanLen),
		buffer: ringBuffer{
			items: make([]interface{}, InitialBufferSize),
		},
	}
	q.bufferCond = sync.NewCond(&q.mutex)
	go q.channelPump()
	return q
}

type queue struct {
	mutex      sync.RWMutex
	bufferCond *sync.Cond
	channel    chan interface{}
	buffer     ringBuffer
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

func (q *queue) Len() int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	return len(q.channel) + q.buffer.count
}

func (q *queue) Cap() int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	return cap(q.channel) + len(q.buffer.items)
}

func (q *queue) Close() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.closed = true
	q.bufferCond.Signal()
	close(q.channel)
}
