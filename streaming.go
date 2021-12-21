package streaming

import (
	"container/heap"
)

// empty slice
var emptySlice = make(Slice, 0)

// empty stream
var emptyStream = &Stream{slice: emptySlice}

// Stream Slice holder
type Stream struct {
	chans []chan interface{}
	slice Slicer
}

// defaultChans default capacity of chan-slice
const defaultChans = 16

// defaultChanBufSize default capacity of buffered-channel
const defaultChanBufSize = 1024

// newStream Stream constructor
func newStream(slice Slicer) *Stream {
	if slice == nil {
		return emptyStream
	}
	return &Stream{
		chans: make([]chan interface{}, 0, defaultChans),
		slice: slice,
	}
}

// Of wraps Slicer into Stream
//
// Returns emptyStream when slicer is nil
func Of(slicer Slicer) *Stream {
	return newStream(slicer)
}

// prevChan returns previous channel
func (s *Stream) prevChan() chan interface{} {
	var ch chan interface{}
	if len(s.chans) == 0 {
		ch = s.curChan()
		if s.slice.Len() > 0 {
			ch <- s.slice.Index(0)
		}
		go func() {
			for i := 1; i < s.slice.Len(); i++ {
				ch <- s.slice.Index(i)
			}
			close(ch)
		}()
	} else {
		ch = s.chans[len(s.chans)-1]
	}
	return ch
}

// curChan returns current channel
func (s *Stream) curChan() chan interface{} {
	cur := make(chan interface{}, defaultChanBufSize)
	s.chans = append(s.chans, cur)
	return cur
}

// ForEach performs an action for each element of this stream.
func (s *Stream) ForEach(act func(interface{})) {
	prev := s.prevChan()
	for v := range prev {
		act(v)
	}
}

// Peek returns the same stream,
// additionally performing the provided action on each element
// as elements are consumed from the resulting stream
func (s *Stream) Peek(act func(interface{})) *Stream {
	prev := s.prevChan()
	cur := s.curChan()

	go func() {
		for v := range prev {
			act(v)
			cur <- v
		}
		close(cur)
	}()

	return s
}

// Limit returns a stream consisting of the elements of this stream,
// truncated to be no longer than max-size in length.
func (s *Stream) Limit(n int) *Stream {
	prev := s.prevChan()
	cur := s.curChan()

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

// Skip returns a stream consisting of the remaining elements
// of this stream after discarding the first n elements
// of the stream. If the stream contains fewer than n elements then
// emptyStream will be returned.
func (s *Stream) Skip(n int) *Stream {
	prev := s.prevChan()
	if n <= 0 {
		return s
	}

	cur := s.curChan()

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

// Map returns a stream consisting of the results (any type)
// of applying the given function to the elements of this stream.
func (s *Stream) Map(apply func(interface{}) interface{}) *Stream {
	prev := s.prevChan()
	cur := s.curChan()

	go func() {
		for v := range prev {
			cur <- apply(v)
		}
		close(cur)
	}()

	return s
}

// FlatMap returns a stream consisting of the results
// of replacing each element of this stream
func (s *Stream) FlatMap(apply func(interface{}) Slicer) *Stream {
	prev := s.prevChan()
	cur := s.curChan()

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

// Reduce performs a reduction on the elements of this stream,
// using the provided comparing function
//
// When steam is empty, Reduce returns nil
func (s *Stream) Reduce(compare func(a, b interface{}) bool) interface{} {
	prev := s.prevChan()
	if len(prev) == 0 {
		return nil
	}
	t := <-prev
	for v := range prev {
		if compare(v, t) {
			t = v
		}
	}

	return t
}

// Filter returns a stream consisting of the elements of this stream
// that match the given predicate.
func (s *Stream) Filter(predicate func(interface{}) bool) *Stream {
	prev := s.prevChan()
	cur := s.curChan()

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

// FilterCount returns count of the elements of this stream
// that match the given predicate.
func (s *Stream) FilterCount(predicate func(interface{}) bool) int {
	var c int

	prev := s.prevChan()
	for v := range prev {
		if predicate(v) {
			c++
		}
	}
	return c
}

var nothing struct{}

// Distinct returns a stream consisting of the distinct elements
// with original order
func (s *Stream) Distinct() *Stream {
	prev := s.prevChan()
	cur := s.curChan()

	var memory = make(map[interface{}]struct{})
	for v := range prev {
		if _, ok := memory[v]; !ok {
			memory[v] = nothing
			cur <- v
		}
	}
	close(cur)

	return s
}

// Collect returns Slicer of this stream
func (s *Stream) Collect() Slicer {
	var slice Slice

	prev := s.prevChan()
	for v := range prev {
		slice = append(slice, v)
	}

	return slice
}

// Count returns the count of elements in this stream
func (s *Stream) Count() int {
	var counter int
	for range s.prevChan() {
		counter++
	}
	return counter
}

// IsEmpty reports stream is empty
func (s *Stream) IsEmpty() bool {
	return s.Count() == 0
}

// Sum returns the sum of elements in this stream
// using the provided sum function
func (s *Stream) Sum(sum func(interface{}) float64) float64 {
	var r float64

	prev := s.prevChan()
	for v := range prev {
		r += sum(v)
	}

	return r
}

// AnyMatch returns whether any elements of this stream match
// the provided predicate. May not evaluate the predicated
// on all elements if not necessary for determining the result.
// If the stream is empty then false is returned and the predicate is not evaluated.
func (s *Stream) AnyMatch(predicate func(interface{}) bool) bool {
	prev := s.prevChan()
	for v := range prev {
		if predicate(v) {
			return true
		}
	}
	return false
}

// AllMatch returns whether all elements of this stream match
// the provided predicate. May not evaluate the predicated
// on all elements if not necessary for determining the result.
// If the stream is empty then true is returned and the predicate is not evaluated.
func (s *Stream) AllMatch(predicate func(interface{}) bool) bool {
	prev := s.prevChan()
	for v := range prev {
		if predicate(v) {
			continue
		}
		return false
	}

	return true
}

// NonMatch returns whether no elements of this stream match
// the provided predicate. May not evaluate the predicated
// on all elements if not necessary for determining the result.
// If the stream is empty then true is returned and the predicate is not evaluated.
func (s *Stream) NonMatch(predicate func(interface{}) bool) bool {
	return !s.AllMatch(predicate)
}

// FindFirst returns the first element of the stream,
// or nil if the stream is empty
func (s *Stream) FindFirst() interface{} {
	prev := s.prevChan()
	if len(prev) == 0 {
		return nil
	}
	return <-prev
}

// Element returns the element at the specified position in this stream
//
// nil is returned when index is out of range
func (s *Stream) Element(i int) interface{} {
	var counter int

	prev := s.prevChan()
	for v := range prev {
		if counter == i {
			return v
		}
		counter++
	}

	return nil
}

// CountVal Count-Val wrapper
type CountVal struct {
	Count int
	Val   interface{}
}

// An cvHeap is a max-heap of CountVals.
type cvHeap []CountVal

func (h cvHeap) Len() int           { return len(h) }
func (h cvHeap) Less(i, j int) bool { return h[i].Count > h[j].Count }
func (h cvHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *cvHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(CountVal))
}

func (h *cvHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// Top returns a stream consisting of n elements that appear most often
func (s *Stream) Top(n int) *Stream {
	var memory = make(map[interface{}]int)

	prev := s.prevChan()
	cur := s.curChan()

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
			cvs = append(cvs, CountVal{
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
			if counter == n {
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
