package streaming

import (
	"sort"
	"sync"
)

// empty parallel stream
var emptyParallel = &ParallelStream{Stream: emptyStream}

// newParallel ParallelStream constructor
func newParallel(slicer Slicer) *ParallelStream {
	if slicer == nil {
		return emptyParallel
	}

	stream := Of(slicer)
	return &ParallelStream{
		Stream: stream,
		ranges: splitSlicer(stream.slice),
	}
}

// ParallelStream Stream holder in a parallel way
type ParallelStream struct {
	*Stream
	ranges Ranges
	wg     sync.WaitGroup
	mu     sync.Mutex
}

// ParallelOf wraps Slicer into ParallelStream
//
// Returns emptyParallel when slicer is nil
func ParallelOf(slicer Slicer) *ParallelStream {
	return newParallel(slicer)
}

// ForEach performs an action for each element of this stream
// in a Parallel way
func (s *ParallelStream) ForEach(act func(interface{})) {
	if s.slice.Len() < 1 {
		return
	}

	for _, r := range s.ranges {
		s.wg.Add(1)
		r := r
		go func() {
			for i := r.From; i < r.To; i++ {
				v := s.slice.Index(i)
				act(v)
			}
			s.wg.Done()
		}()
	}
	s.wg.Wait()
}

// ForEachOrdered performs an action in order for each element of this stream.
func (s *ParallelStream) ForEachOrdered(act func(interface{})) {
	s.Stream.ForEach(act)
}

// MapSame returns the same stream whose elements
// are applied by the given function in a Parallel way.
//
// The apply function must return the same type,
// or else it will PANIC
func (s *ParallelStream) MapSame(apply func(interface{}) interface{}) *ParallelStream {
	if s.slice.Len() < 1 {
		return emptyParallel
	}

	for i, r := range s.ranges {
		s.wg.Add(1)
		r := r
		go func(i int) {
			for i := r.From; i < r.To; i++ {
				v := s.slice.Index(i)
				s.slice.Set(i, apply(v))
			}

			s.wg.Done()
		}(i)
	}
	s.wg.Wait()

	return s
}

// Map returns a stream consisting of the results of applying the given
// function to the elements of this stream in a Parallel way
func (s *ParallelStream) Map(apply func(interface{}) interface{}) *ParallelStream {
	if s.slice.Len() < 1 {
		return emptyParallel
	}

	var mapSlice = make(map[int]Slice, cpu)
	for i, r := range s.ranges {
		s.wg.Add(1)
		r := r
		go func(i int) {
			var slice Slice
			for j := r.From; j < r.To; j++ {
				v := s.slice.Index(j)
				slice = append(slice, apply(v))
			}

			s.mu.Lock()
			mapSlice[i] = slice
			s.mu.Unlock()

			s.wg.Done()
		}(i)
	}
	s.wg.Wait()

	var slice Slice
	for i := 0; i < cpu; i++ {
		for _, e := range mapSlice[i] {
			slice = append(slice, e)
		}
	}

	return newParallel(slice)
}

// FlatMap returns a stream consisting of the results
// of replacing each element of this stream
func (s *ParallelStream) FlatMap(apply func(interface{}) Slicer) *ParallelStream {
	if s.slice.Len() < 1 {
		return emptyParallel
	}

	var mapSlice = make(map[int]Slice, cpu)
	for i, r := range s.ranges {
		s.wg.Add(1)
		r := r
		go func(i int) {
			var slice Slice
			for j := r.From; j < r.To; j++ {
				v := apply(s.slice.Index(j))
				if v == nil {
					continue
				}
				for k := 0; k < v.Len(); k++ {
					ele := v.Index(k)
					slice = append(slice, ele)
				}
			}

			s.mu.Lock()
			mapSlice[i] = slice
			s.mu.Unlock()

			s.wg.Done()
		}(i)
	}
	s.wg.Wait()

	var slice Slice
	for i := 0; i < cpu; i++ {
		for _, e := range mapSlice[i] {
			slice = append(slice, e)
		}
	}

	return newParallel(slice)
}

// Reduce performs a reduction on the elements of this stream,
// using the provided comparing function in a Parallel way
//
// When steam is empty, Reduce returns nil, -1
func (s *ParallelStream) Reduce(compare func(a, b interface{}) bool) interface{} {
	if s.slice.Len() < 1 {
		return nil
	}

	var vs Slice
	for _, r := range s.ranges {
		s.wg.Add(1)
		r := r
		go func() {
			t := s.slice.Index(r.From)
			for i := r.From + 1; i < r.To; i++ {
				v := s.slice.Index(i)
				if compare(v, t) {
					t = v
				}
			}

			s.mu.Lock()
			vs = append(vs, t)
			s.mu.Unlock()

			s.wg.Done()
		}()
	}
	s.wg.Wait()

	t := vs[0]
	for i := 1; i < len(vs); i++ {
		ti := vs[i]
		if compare(ti, t) {
			t = ti
		}
	}

	return t
}

// Filter returns a stream consisting of the elements of this stream
// that match the given predicate.
func (s *ParallelStream) Filter(predicate func(interface{}) bool) *ParallelStream {
	if s.slice.Len() < 1 {
		return emptyParallel
	}

	var mapSlice = make(map[int]Slice, cpu)
	for i, r := range s.ranges {
		s.wg.Add(1)
		r := r
		go func(i int) {
			var slice Slice
			for j := r.From; j < r.To; j++ {
				v := s.slice.Index(j)
				if predicate(v) {
					slice = append(slice, v)
				}
			}

			s.mu.Lock()
			mapSlice[i] = slice
			s.mu.Unlock()

			s.wg.Done()
		}(i)
	}
	s.wg.Wait()

	var slice Slice
	for i := 0; i < cpu; i++ {
		for _, e := range mapSlice[i] {
			slice = append(slice, e)
		}
	}

	return newParallel(slice)
}

// Sum returns the sum of elements in this stream
// using the provided sum function in a Parallel way
func (s *ParallelStream) Sum(sum func(interface{}) float64) float64 {
	if s.slice.Len() < 1 {
		return 0
	}

	var ff float64
	for _, r := range s.ranges {
		s.wg.Add(1)
		r := r
		go func() {
			var f float64
			for i := r.From; i < r.To; i++ {
				v := s.slice.Index(i)
				f += sum(v)
			}

			s.mu.Lock()
			ff += f
			s.mu.Unlock()

			s.wg.Done()
		}()
	}
	s.wg.Wait()
	return ff
}

// Copy returns a new stream that holds a copied slice
//
// Returns emptyParallel when stream is empty
func (s *ParallelStream) Copy() *ParallelStream {
	if s.slice.Len() < 1 {
		return emptyParallel
	}

	slice := make(Slice, s.slice.Len())
	for i := 0; i < s.slice.Len(); i++ {
		slice[i] = s.slice.Index(i)
	}

	return newParallel(slice)
}

// Sorted returns a sorted stream consisting of the elements of this stream
// sorted according to the provided less.
//
// Sorted reorders inside slice
// For keeping the order relation of original slice, use Copy first
func (s *ParallelStream) Sorted(less func(i, j int) bool) *ParallelStream {
	sort.Slice(s.slice, less)
	s.ranges = splitSize(s.slice.Len())
	return s
}
