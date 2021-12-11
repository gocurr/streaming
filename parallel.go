package streaming

import (
	"github.com/gocurr/partition"
	"runtime"
	"sort"
	"sync"
)

// cpu number
var cpu = runtime.NumCPU()

// empty parallel
var parallelEmpty = &ParallelStream{}

// ParallelStream Stream holder in a parallel way
type ParallelStream struct {
	*Stream
	ranges Ranges
	wg     sync.WaitGroup
	mu     sync.Mutex
}

// ParallelOf wraps input into *ParallelStream
//
// Returns parallelEmpty when raw is nil
// Or is NOT a slice or an array
func ParallelOf(raw Slicer) *ParallelStream {
	stream := Of(raw)
	slice := stream.slice

	var size int
	if slice != nil {
		size = slice.Len()
	}
	return &ParallelStream{
		Stream: stream,
		ranges: split(size),
		wg:     sync.WaitGroup{},
		mu:     sync.Mutex{},
	}
}

type Ranges []partition.Range

func split(size int) Ranges {
	if size < 1 {
		return nil
	}
	return partition.RangesN(size, cpu)
}

// ForEach performs an action for each element of this stream
// in a Parallel way
func (s *ParallelStream) ForEach(act func(interface{})) {
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

// MapSame returns a stream consisting of the results of applying the given
// function to the elements of this stream in a Parallel way
func (s *ParallelStream) MapSame(apply func(interface{}) interface{}) *ParallelStream {
	if s.slice == nil || s.slice.Len() == 0 {
		return parallelEmpty
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
	if s.slice == nil || s.slice.Len() == 0 {
		return parallelEmpty
	}

	var mapSlice = make(map[int]Slice, cpu)
	for i, r := range s.ranges {
		s.wg.Add(1)
		r := r
		go func(i int) {
			var slice Slice
			for i := r.From; i < r.To; i++ {
				v := s.slice.Index(i)
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

	var sliceLen int
	if slice != nil {
		sliceLen = slice.Len()
	}
	return &ParallelStream{
		Stream: &Stream{slice: slice},
		ranges: split(sliceLen),
	}
}

// FlatMap returns a stream consisting of the results
// of replacing each element of this stream
func (s *ParallelStream) FlatMap(apply func(interface{}) Slicer) *ParallelStream {
	if s.slice == nil || s.slice.Len() == 0 {
		return parallelEmpty
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

	var sliceLen int
	if slice != nil {
		sliceLen = slice.Len()
	}
	return &ParallelStream{
		Stream: &Stream{slice: slice},
		ranges: split(sliceLen),
	}
}

// Reduce performs a reduction on the elements of this stream,
// using the provided comparing function in a Parallel way
//
// When steam is empty, Reduce returns nil, -1
func (s *ParallelStream) Reduce(compare func(a, b interface{}) bool) interface{} {
	if s.slice == nil || s.slice.Len() == 0 {
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

// Sum returns the sum of elements in this stream
// using the provided sum function in a Parallel way
func (s *ParallelStream) Sum(sum func(interface{}) float64) float64 {
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

// Copy returns a new stream containing the elements,
// the new stream holds a copied slice
func (s *ParallelStream) Copy() *ParallelStream {
	if s.slice == nil || s.slice.Len() == 0 {
		return parallelEmpty
	}
	slice := make(Slice, s.slice.Len())
	for i := 0; i < s.slice.Len(); i++ {
		slice[i] = s.slice.Index(i)
	}
	return &ParallelStream{
		Stream: &Stream{slice: slice},
		ranges: split(slice.Len()),
	}
}

// Sorted returns a sorted stream consisting of the elements of this stream
// sorted according to the provided less.
//
// Sorted reorders inside slice
// For keeping the order relation of original slice, use Copy first
func (s *ParallelStream) Sorted(less func(i, j int) bool) *ParallelStream {
	if s.slice == nil || s.slice.Len() == 0 {
		return parallelEmpty
	}
	sort.Slice(s.slice, less)
	s.ranges = split(s.slice.Len())
	return s
}
