package streaming

import (
	"container/heap"
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
)

// Stream a sequence of elements supporting aggregate operations.
//
// To perform a computation, stream operations composed into a stream pipeline.
type Stream struct {
	chanBufSize int                // size of buffered channel
	pipeline    []chan interface{} // pipeline channels
	slice       Slicer             // which holds the elements to process
	closed      bool               // to report sink-methods invoked
}

// newStream returns a new Stream which wraps the given slice.
func newStream(slice Slicer, chanBufSize int) *Stream {
	if slice == nil {
		return emptyStream
	}
	return &Stream{
		chanBufSize: chanBufSize,
		pipeline:    make([]chan interface{}, 0, defaultPipelineCap),
		slice:       slice,
	}
}

// Of wraps Slicer into Stream with defaultChanBufSize.
//
// Returns emptyStream when slicer is nil.
func Of(slicer Slicer) *Stream {
	return newStream(slicer, defaultChanBufSize)
}

// OfWithChanBufSize wraps Slicer into Stream with the given chanBufSize.
//
// Returns emptyStream when slicer is nil.
func OfWithChanBufSize(slicer Slicer, chanBufSize int) *Stream {
	if chanBufSize <= 0 {
		panic("OfWithChanBufSize: negative or 0 chanBufSize")
	}
	return newStream(slicer, chanBufSize)
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
			vv := apply(v)
			if vv == nil {
				continue
			}
			for i := 0; i < vv.Len(); i++ {
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
			if _, ok := m[v]; !ok {
				m[v] = 1
			} else {
				m[v] += 1
			}
		}

		var cvs cvHeap
		for v, count := range m {
			cvs = append(cvs, &CountVal{
				Count: count,
				Val:   v,
			})
		}

		h := &cvs
		heap.Init(h)

		var counter int
		for h.Len() != 0 {
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
	return newStream(s.slice, s.chanBufSize)
}
