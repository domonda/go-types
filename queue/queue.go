package queue

import "sync"

const (
	// DefaultChanLen is the consumer-channel capacity used by New
	// when no ChanLen option is passed.
	DefaultChanLen = 32
	// DefaultInitialBufferSize is the initial ring-buffer capacity used
	// by New when no InitialBufferSize option is passed.
	DefaultInitialBufferSize = 32
)

// Option is an optional configuration value for New.
// The available options are ChanLen and InitialBufferSize.
type Option interface {
	apply(*config)
}

type config struct {
	chanLen           int
	initialBufferSize int
}

// ChanLen is an Option that sets the capacity of the buffered channel a
// Queue exposes to consumers. It is a pure throughput knob: a larger value
// lets more items sit ready to be consumed and reduces how often the internal
// pump goroutine has to run. It must be >= 0; New panics otherwise.
type ChanLen int

func (o ChanLen) apply(c *config) { c.chanLen = int(o) }

// InitialBufferSize is an Option that sets the initial capacity of the
// internal ring-buffer that holds items the consumer channel cannot accept
// yet. The buffer grows automatically as needed. It must be >= 1; New panics
// otherwise.
type InitialBufferSize int

func (o InitialBufferSize) apply(c *config) { c.initialBufferSize = int(o) }

// Queue is an unbounded, goroutine-safe FIFO queue of items of type T that
// exposes a channel for consuming them. Items are added via Add and received
// from the channel returned by Next. An internal ring-buffer holds items
// until the consumer channel can accept them, so Add never blocks. Call Close
// when the queue is no longer needed; the consumer channel is closed, every
// item still queued is discarded, and the background pump goroutine exits.
type Queue[T any] interface {
	Add(items ...T)
	Next() <-chan T
	Len() int
	Cap() int
	Close()
}

// New returns a new Queue of items of type T, configured by the optional
// Options. Without options the consumer channel has capacity DefaultChanLen
// and the internal ring-buffer starts at DefaultInitialBufferSize; the
// ring-buffer grows without bound, so Add never blocks regardless of the
// configuration. A background goroutine pumps items from the ring-buffer into
// the channel; call Close to stop it and release resources.
//
// New panics if a ChanLen below 0 or an InitialBufferSize below 1 is configured.
func New[T any](options ...Option) Queue[T] {
	cfg := config{
		chanLen:           DefaultChanLen,
		initialBufferSize: DefaultInitialBufferSize,
	}
	for _, opt := range options {
		opt.apply(&cfg)
	}
	if cfg.chanLen < 0 {
		panic("queue.New: ChanLen must be >= 0")
	}
	if cfg.initialBufferSize < 1 {
		panic("queue.New: InitialBufferSize must be >= 1")
	}

	q := &queue[T]{
		channel:    make(chan T, cfg.chanLen),
		done:       make(chan struct{}),
		pumpExited: make(chan struct{}),
		buffer:     ringBuffer[T]{items: make([]T, cfg.initialBufferSize)},
	}
	q.bufferCond = sync.NewCond(&q.mutex)
	go q.channelPump()
	return q
}

type queue[T any] struct {
	mutex      sync.RWMutex
	bufferCond *sync.Cond
	channel    chan T
	done       chan struct{}
	pumpExited chan struct{}
	buffer     ringBuffer[T]
	closed     bool
}

// channelPump moves items from the ring-buffer into the consumer channel one
// at a time. The blocking send is done without holding the mutex, so a full
// channel parks this goroutine in the scheduler instead of busy-spinning,
// while Add stays free to append to the unbounded ring-buffer. The pump is
// the sole sender to and closer of the channel, which keeps delivery strictly
// FIFO and free of send-on-closed races. On Close it discards whatever is
// still staged in the channel before closing it. On return it closes
// pumpExited, which is what makes Close synchronous.
func (q *queue[T]) channelPump() {
	defer close(q.pumpExited)
	for {
		q.mutex.Lock()
		for !q.closed && q.buffer.isEmpty() {
			q.bufferCond.Wait()
		}
		if q.closed {
			q.mutex.Unlock()
			q.drainAndClose()
			return
		}
		item := q.buffer.shift()
		q.mutex.Unlock()

		select {
		case q.channel <- item:
		case <-q.done:
			q.drainAndClose()
			return
		}
	}
}

// drainAndClose discards every item still buffered in the consumer channel
// and then closes it. It is called exactly once, by channelPump on shutdown.
func (q *queue[T]) drainAndClose() {
	for {
		select {
		case <-q.channel:
		default:
			close(q.channel)
			return
		}
	}
}

// Add appends items to the queue. It never blocks: items are pushed onto the
// unbounded ring-buffer and the pump goroutine delivers them to the consumer
// channel. Add is a no-op after Close.
func (q *queue[T]) Add(items ...T) {
	if len(items) == 0 {
		return
	}

	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.closed {
		return
	}
	for _, item := range items {
		q.buffer.push(item)
	}
	q.bufferCond.Signal()
}

func (q *queue[T]) Next() <-chan T {
	return q.channel
}

func (q *queue[T]) Len() int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	return len(q.channel) + q.buffer.count
}

func (q *queue[T]) Cap() int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	return cap(q.channel) + len(q.buffer.items)
}

// Close shuts the queue down: the ring-buffer is dropped and the pump
// goroutine discards whatever is still staged in the consumer channel, then
// closes it so ranging consumers exit. Close blocks until the pump has exited
// and the channel is closed, so the queue is fully shut down once Close
// returns.
//
// A consumer reading concurrently with Close may still win some of the items
// staged in the channel (up to ChanLen of them) before the queue discards
// them; the discard guarantee covers the queue's own backlog, not an actively
// racing consumer. Add is a no-op afterwards. Close is idempotent.
func (q *queue[T]) Close() {
	q.mutex.Lock()
	if q.closed {
		q.mutex.Unlock()
		<-q.pumpExited
		return
	}
	q.closed = true
	close(q.done)
	q.buffer = ringBuffer[T]{}
	q.bufferCond.Signal()
	q.mutex.Unlock()

	<-q.pumpExited
}
