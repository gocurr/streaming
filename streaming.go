package streaming

import (
	"container/heap"
	"time"
)

const (
	// Default capacity of pipeline.
	defaultPipelineCap = 16

	// Default capacity of each pipe(buffered channel) in pipeline.
	defaultChanBufSize = 1 << 10 // 1 KB
)

var (
	emptySlice = make(Slice, 0)

	emptyStream = &Stream{slice: emptySlice}

	never = time.Duration(0) // never timeout

	defaultOption = &Option{
		ChanBufSize: defaultChanBufSize,
		Timeout:     never,
	}
)

// Stream a sequence of elements supporting aggregate operations.
//
// To perform a computation, stream operations composed into a stream pipeline.
type Stream struct {
	pipeline []chan interface{} // pipeline channels
	slice    Slicer             // which holds the elements to process
	closed   bool               // to report sink-methods invoked

	chanBufSize int       // size of buffered channel
	deadline    time.Time // deadline to exceed

	incorrect bool // reports whether the result is incorrect
}

// Option represents optional conditions.
type Option struct {
	ChanBufSize int           // size of buffered channel
	Timeout     time.Duration // timeout to exceed
}

// newStream returns a new Stream which wraps the given slice.
func newStream(slice Slicer, op *Option) *Stream {
	if slice == nil {
		return emptyStream
	}

	var deadline time.Time
	if op.Timeout > 0 {
		deadline = time.Now().Add(op.Timeout)
	}

	return &Stream{
		chanBufSize: op.ChanBufSize,
		pipeline:    make([]chan interface{}, 0, defaultPipelineCap),
		slice:       slice,
		deadline:    deadline,
	}
}

// Of wraps Slicer into Stream with defaultChanBufSize.
//
// Returns emptyStream when slicer is nil.
func Of(slicer Slicer) *Stream {
	return newStream(slicer, defaultOption)
}

// OfWithOption wraps Slicer into Stream with Option.
//
// Returns emptyStream when slicer is nil.
func OfWithOption(slicer Slicer, op *Option) *Stream {
	if op == nil {
		return Of(slicer)
	}

	if op.ChanBufSize <= 0 {
		op.ChanBufSize = defaultChanBufSize
	}
	return newStream(slicer, op)
}

// exceeded reports the deadline is exceeded.
func (s *Stream) exceeded() bool {
	timeout := !s.deadline.IsZero() && time.Now().After(s.deadline)
	if timeout {
		s.incorrect = true
	}
	return timeout
}

// Validation check for Stream.
//
// It will panic when Stream is closed.
func (s *Stream) valid() {
	if s.closed {
		panic("stream has already been operated upon")
	}
}

// Peek returns the same stream,
// additionally performing the provided action on each element
// as elements are consumed from the resulting stream.
func (s *Stream) Peek(act func(interface{})) *Stream {
	s.valid()

	prev := s.prevPipe()
	cur := s.curPipe()

	go func() {
		defer close(cur)
		for v := range prev {
			if s.exceeded() {
				break
			}
			act(v)
			cur <- v
		}
	}()

	return s
}

// Limit returns the same stream consisting of the elements,
// truncated to be no longer than max-size in length.
func (s *Stream) Limit(n int) *Stream {
	s.valid()

	prev := s.prevPipe()
	cur := s.curPipe()

	go func() {
		defer close(cur)
		var counter int
		for v := range prev {
			if s.exceeded() {
				break
			}
			if counter < n {
				counter++
				cur <- v
				continue
			}
			break
		}
	}()

	return s
}

// Skip returns the same stream consisting of the remaining elements
// after discarding the first n elements of the stream.
// If the stream contains fewer than n elements then
// emptyStream will be returned.
func (s *Stream) Skip(n int) *Stream {
	s.valid()

	prev := s.prevPipe()
	if n <= 0 {
		return s
	}

	cur := s.curPipe()

	go func() {
		defer close(cur)
		var counter int
		for v := range prev {
			if s.exceeded() {
				break
			}
			if counter < n {
				counter++
				continue
			}
			cur <- v
			counter++
		}
	}()

	return s
}

// Map returns the same stream consisting of the results (any type)
// of applying the given function to the elements.
func (s *Stream) Map(apply func(interface{}) interface{}) *Stream {
	s.valid()

	prev := s.prevPipe()
	cur := s.curPipe()

	go func() {
		defer close(cur)
		for v := range prev {
			if s.exceeded() {
				break
			}
			cur <- apply(v)
		}
	}()

	return s
}

// FlatMap returns the same stream consisting of the results
// of replacing each element.
func (s *Stream) FlatMap(apply func(interface{}) Slicer) *Stream {
	s.valid()

	prev := s.prevPipe()
	cur := s.curPipe()

	go func() {
		defer close(cur)
		for v := range prev {
			if s.exceeded() {
				break
			}
			vv := apply(v)
			if vv == nil {
				continue
			}
			for i := 0; i < vv.Len(); i++ {
				if s.exceeded() {
					break
				}
				ele := vv.Index(i)
				cur <- ele
			}
		}
	}()

	return s
}

// Filter returns the same stream consisting of the elements
// that match the given predicate.
func (s *Stream) Filter(predicate func(interface{}) bool) *Stream {
	s.valid()

	prev := s.prevPipe()
	cur := s.curPipe()

	go func() {
		defer close(cur)
		for v := range prev {
			if s.exceeded() {
				break
			}
			if predicate(v) {
				cur <- v
			}
		}
	}()

	return s
}

// discards values in map.
var discard struct{}

// Distinct returns the same stream consisting of the distinct elements
// with original order.
func (s *Stream) Distinct() *Stream {
	s.valid()

	prev := s.prevPipe()
	cur := s.curPipe()

	go func() {
		defer close(cur)
		var distinct = make(map[interface{}]struct{})
		for v := range prev {
			if s.exceeded() {
				break
			}
			if _, ok := distinct[v]; !ok {
				distinct[v] = discard
				cur <- v
			}
		}
	}()

	return s
}

// Top returns the same stream consisting of CountVal pointers
// of the n elements that appear most often.
func (s *Stream) Top(n int) *Stream {
	s.valid()

	prev := s.prevPipe()
	cur := s.curPipe()
	if n < 1 {
		close(cur)
		return s
	}

	go func() {
		defer close(cur)
		var m = make(map[interface{}]int)
		for v := range prev {
			if s.exceeded() {
				break
			}
			if _, ok := m[v]; !ok {
				m[v] = 1
			} else {
				m[v] += 1
			}
		}

		var cvs cvHeap
		for v, count := range m {
			if s.exceeded() {
				break
			}
			cvs = append(cvs, &CountVal{
				Count: count,
				Val:   v,
			})
		}

		h := &cvs
		heap.Init(h)

		var counter int
		for h.Len() != 0 {
			if s.exceeded() {
				break
			}
			cv := heap.Pop(h)
			cur <- cv
			if counter == n-1 {
				break
			}
			counter++
		}
	}()

	return s
}

// Copy returns a new stream that wraps the initial slice.
func (s *Stream) Copy() *Stream {
	return Of(s.slice)
}
