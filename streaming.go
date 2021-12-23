package streaming

import (
	"container/heap"
)

const (
	// default capacity of pipeline
	defaultPipelineCap = 16

	// default capacity of each pipe(buffered channel) in pipeline
	defaultChanBufSize = 1 << 10
)

var (
	// empty slice
	emptySlice = make(Slice, 0)

	// empty stream
	emptyStream = &Stream{slice: emptySlice}
)

// Stream a sequence of elements supporting aggregate operations
//
// To perform a computation, stream operations composed into a stream pipeline
type Stream struct {
	pipeline []chan interface{} // pipeline channels
	slice    Slicer
}

// newStream returns a new Stream
func newStream(slice Slicer) *Stream {
	if slice == nil {
		return emptyStream
	}
	return &Stream{
		pipeline: make([]chan interface{}, 0, defaultPipelineCap),
		slice:    slice,
	}
}

// Of wraps Slicer into Stream
//
// Returns emptyStream when slicer is nil
func Of(slicer Slicer) *Stream {
	return newStream(slicer)
}

// Peek returns the same stream,
// additionally performing the provided action on each element
// as elements are consumed from the resulting stream
func (s *Stream) Peek(act func(interface{})) *Stream {
	prev := s.prevPipe()
	cur := s.curPipe()

	go func() {
		for v := range prev {
			act(v)
			cur <- v
		}
		close(cur)
	}()

	return s
}

// Limit returns the same stream consisting of the elements,
// truncated to be no longer than max-size in length.
func (s *Stream) Limit(n int) *Stream {
	prev := s.prevPipe()
	cur := s.curPipe()

	go func() {
		var counter int
		for v := range prev {
			if counter < n {
				counter++
				cur <- v
				continue
			}
			break
		}
		close(cur)
	}()

	return s
}

// Skip returns the same stream consisting of the remaining elements
// after discarding the first n elements of the stream.
// If the stream contains fewer than n elements then
// emptyStream will be returned.
func (s *Stream) Skip(n int) *Stream {
	prev := s.prevPipe()
	if n <= 0 {
		return s
	}

	cur := s.curPipe()

	go func() {
		var counter int
		for v := range prev {
			if counter < n {
				counter++
				continue
			}
			cur <- v
			counter++
		}
		close(cur)
	}()

	return s
}

// Map returns the same stream consisting of the results (any type)
// of applying the given function to the elements.
func (s *Stream) Map(apply func(interface{}) interface{}) *Stream {
	prev := s.prevPipe()
	cur := s.curPipe()

	go func() {
		for v := range prev {
			cur <- apply(v)
		}
		close(cur)
	}()

	return s
}

// FlatMap returns the same stream consisting of the results
// of replacing each element
func (s *Stream) FlatMap(apply func(interface{}) Slicer) *Stream {
	prev := s.prevPipe()
	cur := s.curPipe()

	go func() {
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
		close(cur)
	}()

	return s
}

// Filter returns the same stream consisting of the elements
// that match the given predicate.
func (s *Stream) Filter(predicate func(interface{}) bool) *Stream {
	prev := s.prevPipe()
	cur := s.curPipe()

	go func() {
		for v := range prev {
			if predicate(v) {
				cur <- v
			}
		}
		close(cur)
	}()

	return s
}

// discards values in map
var discard struct{}

// Distinct returns the same stream consisting of the distinct elements
// with original order
func (s *Stream) Distinct() *Stream {
	prev := s.prevPipe()
	cur := s.curPipe()

	go func() {
		var distinct = make(map[interface{}]struct{})
		for v := range prev {
			if _, ok := distinct[v]; !ok {
				distinct[v] = discard
				cur <- v
			}
		}
		close(cur)
	}()

	return s
}

// Top returns the same stream consisting of n elements
// that appear most often
func (s *Stream) Top(n int) *Stream {
	prev := s.prevPipe()
	cur := s.curPipe()
	if n < 1 {
		close(cur)
		return s
	}

	var memory = make(map[interface{}]int)

	go func() {
		for v := range prev {
			if _, ok := memory[v]; !ok {
				memory[v] = 1
			} else {
				memory[v] += 1
			}
		}

		var cvs cvHeap
		for v, count := range memory {
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
		close(cur)
	}()

	return s
}

// Copy returns a new stream
func (s *Stream) Copy() *Stream {
	return newStream(s.slice)
}
