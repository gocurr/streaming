package streaming

import (
	"container/heap"
	"sort"
)

// empty stream
var empty = &Stream{slice: make(Slice, 0)}

// Stream Slice holder
type Stream struct {
	slice Slicer
}

// newStream Stream constructor
func newStream(slice Slicer) *Stream {
	if slice == nil {
		return empty
	}
	return &Stream{slice: slice}
}

// Of wraps Slicer into Stream
//
// Returns empty when raw is nil
// Or is NOT a slice or an array
func Of(slicer Slicer) *Stream {
	return newStream(slicer)
}

// ForEach performs an action for each element of this stream.
func (s *Stream) ForEach(act func(interface{})) {
	for i := 0; i < s.slice.Len(); i++ {
		v := s.slice.Index(i)
		act(v)
	}
}

// Peek returns the same stream,
// additionally performing the provided action on each element
// as elements are consumed from the resulting stream
func (s *Stream) Peek(act func(interface{})) *Stream {
	for i := 0; i < s.slice.Len(); i++ {
		v := s.slice.Index(i)
		act(v)
	}
	return s
}

// Limit returns a stream consisting of the elements of this stream,
// truncated to be no longer than max-size in length.
func (s *Stream) Limit(n int) *Stream {
	if n < 1 {
		return empty
	}

	if n > s.slice.Len() {
		return s
	}

	var slice = s.slice.Sub(0, n)
	return newStream(slice)
}

// Skip returns a stream consisting of the remaining elements
// of this stream after discarding the first n elements
// of the stream. If the stream contains fewer than n elements then
// an empty stream will be returned.
func (s *Stream) Skip(n int) *Stream {
	if n < 0 {
		return s
	}

	if n > s.slice.Len() {
		return empty
	}

	var slice = s.slice.Sub(n, s.slice.Len())
	return newStream(slice)
}

// MapSame returns the same stream whose elements
// are applied by the given function.
//
// The apply function must return the same type,
// or else it will PANIC
func (s *Stream) MapSame(apply func(interface{}) interface{}) *Stream {
	for i := 0; i < s.slice.Len(); i++ {
		v := s.slice.Index(i)
		s.slice.Set(i, apply(v))
	}

	return s
}

// Map returns a stream consisting of the results (any type)
// of applying the given function to the elements of this stream.
func (s *Stream) Map(apply func(interface{}) interface{}) *Stream {
	var slice Slice
	for i := 0; i < s.slice.Len(); i++ {
		v := s.slice.Index(i)
		slice = append(slice, apply(v))
	}

	return newStream(slice)
}

// FlatMap returns a stream consisting of the results
// of replacing each element of this stream
func (s *Stream) FlatMap(apply func(interface{}) Slicer) *Stream {
	var slice Slice
	for i := 0; i < s.slice.Len(); i++ {
		v := apply(s.slice.Index(i))
		if v == nil {
			continue
		}

		for i := 0; i < v.Len(); i++ {
			ele := v.Index(i)
			slice = append(slice, ele)
		}
	}

	return newStream(slice)
}

// Reduce performs a reduction on the elements of this stream,
// using the provided comparing function
//
// When steam is empty, Reduce returns nil
func (s *Stream) Reduce(compare func(a, b interface{}) bool) interface{} {
	if s.slice.Len() < 1 {
		return nil
	}

	t := s.slice.Index(0)
	for j := 1; j < s.slice.Len(); j++ {
		v := s.slice.Index(j)
		if compare(v, t) {
			t = v
		}
	}

	return t
}

// Filter returns a stream consisting of the elements of this stream
// that match the given predicate.
func (s *Stream) Filter(predicate func(interface{}) bool) *Stream {
	var slice Slice
	for i := 0; i < s.slice.Len(); i++ {
		v := s.slice.Index(i)
		if predicate(v) {
			slice = append(slice, v)
		}
	}
	return newStream(slice)
}

// FilterCount returns count of the elements of this stream
// that match the given predicate.
func (s *Stream) FilterCount(predicate func(interface{}) bool) int {
	var c int
	for i := 0; i < s.slice.Len(); i++ {
		v := s.slice.Index(i)
		if predicate(v) {
			c++
		}
	}
	return c
}

// Distinct returns a stream consisting of the distinct elements
// with original order
func (s *Stream) Distinct() *Stream {
	if s.slice.Len() < 1 {
		return empty
	}

	var slice Slice
	var memory = make(map[interface{}]int)
	for i := 0; i < s.slice.Len(); i++ {
		v := s.slice.Index(i)
		if _, ok := memory[v]; !ok {
			memory[v] = i
			slice = append(slice, v)
		}
	}

	return newStream(slice)
}

// Collect returns data load of this stream
func (s *Stream) Collect() Slicer {
	if s.slice.Len() < 1 {
		return nil
	}
	return s.slice.Sub(0, s.slice.Len())
}

// Count returns the count of elements in this stream
func (s *Stream) Count() int {
	return s.slice.Len()
}

// IsEmpty reports stream is empty
func (s *Stream) IsEmpty() bool {
	return s.slice.Len() < 1
}

// Sum returns the sum of elements in this stream
// using the provided sum function
func (s *Stream) Sum(sum func(interface{}) float64) float64 {
	var r float64
	for i := 0; i < s.slice.Len(); i++ {
		v := s.slice.Index(i)
		r += sum(v)
	}
	return r
}

// AnyMatch returns whether any elements of this stream match
// the provided predicate. May not evaluate the predicated
// on all elements if not necessary for determining the result.
// If the stream is empty then false is returned and the predicate is not evaluated.
func (s *Stream) AnyMatch(predicate func(interface{}) bool) bool {
	for i := 0; i < s.slice.Len(); i++ {
		v := s.slice.Index(i)
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
	for i := 0; i < s.slice.Len(); i++ {
		v := s.slice.Index(i)
		if predicate(v) {
			continue
		}
		return false
	}
	return true
}

// NoneMatch returns whether no elements of this stream match
// the provided predicate. May not evaluate the predicated
// on all elements if not necessary for determining the result.
// If the stream is empty then true is returned and the predicate is not evaluated.
func (s *Stream) NoneMatch(predicate func(interface{}) bool) bool {
	for i := 0; i < s.slice.Len(); i++ {
		v := s.slice.Index(i)
		if !predicate(v) {
			continue
		}
		return false
	}
	return true
}

// FindFirst returns the first element of the stream,
// or nil if the stream is empty
func (s *Stream) FindFirst() interface{} {
	if s.slice.Len() < 1 {
		return nil
	}
	return s.slice.Index(0)
}

// Element returns the element at the specified position in this stream
//
// nil is returned when index is out of range
func (s *Stream) Element(i int) interface{} {
	if i < 0 || i >= s.slice.Len() {
		return nil
	}
	return s.slice.Index(i)
}

// Copy returns a new stream containing the elements,
// the new stream holds a copied slice
func (s *Stream) Copy() *Stream {
	slice := make(Slice, s.slice.Len())
	for i := 0; i < s.slice.Len(); i++ {
		slice[i] = s.slice.Index(i)
	}
	return newStream(slice)
}

// Sorted returns a sorted stream consisting of the elements of this stream
// sorted according to the provided less.
//
// Sorted reorders inside slice
// For keeping the order relation of original slice, use Copy first
func (s *Stream) Sorted(less func(i, j int) bool) *Stream {
	sort.Slice(s.slice, less)
	return s
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
	if n < 1 || s.slice.Len() < 1 {
		return empty
	}

	var memory = make(map[interface{}]int)
	for i := 0; i < s.slice.Len(); i++ {
		v := s.slice.Index(i)
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

	var slice Slice
	for h.Len() != 0 {
		cv := heap.Pop(h)
		slice = append(slice, cv)
		if len(slice) == n {
			break
		}
	}

	return newStream(slice)
}
