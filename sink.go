package streaming

// ForEach performs an action for each element .
func (s *Stream) ForEach(act func(interface{})) {
	prev := s.prevPipe()
	for v := range prev {
		act(v)
	}
}

// Reduce performs a reduction on the elements ,
// using the provided comparing function
//
// When steam is empty, Reduce returns nil
func (s *Stream) Reduce(compare func(a, b interface{}) bool) interface{} {
	prev := s.prevPipe()
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

// FilterCount returns count of the elements
// that match the given predicate.
func (s *Stream) FilterCount(predicate func(interface{}) bool) int {
	var c int

	prev := s.prevPipe()
	for v := range prev {
		if predicate(v) {
			c++
		}
	}
	return c
}

// Collect returns a Slicer consisting of the elements in this stream
func (s *Stream) Collect() Slicer {
	var slice Slice

	prev := s.prevPipe()
	for v := range prev {
		slice = append(slice, v)
	}

	return slice
}

// Count returns the count of elements in this stream
func (s *Stream) Count() int {
	var counter int
	for range s.prevPipe() {
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

	prev := s.prevPipe()
	for v := range prev {
		r += sum(v)
	}

	return r
}

// AnyMatch returns whether any elements  match
// the provided predicate. May not evaluate the predicated
// on all elements if not necessary for determining the result.
// If the stream is empty then false is returned and the predicate is not evaluated.
func (s *Stream) AnyMatch(predicate func(interface{}) bool) bool {
	prev := s.prevPipe()
	for v := range prev {
		if predicate(v) {
			return true
		}
	}
	return false
}

// AllMatch returns whether all elements  match
// the provided predicate. May not evaluate the predicated
// on all elements if not necessary for determining the result.
// If the stream is empty then true is returned and the predicate is not evaluated.
func (s *Stream) AllMatch(predicate func(interface{}) bool) bool {
	prev := s.prevPipe()
	for v := range prev {
		if predicate(v) {
			continue
		}
		return false
	}

	return true
}

// NonMatch returns whether no elements  match
// the provided predicate. May not evaluate the predicated
// on all elements if not necessary for determining the result.
// If the stream is empty then true is returned and the predicate is not evaluated.
func (s *Stream) NonMatch(predicate func(interface{}) bool) bool {
	return !s.AllMatch(predicate)
}

// FindFirst returns the first element of the stream,
// or nil if the stream is empty
func (s *Stream) FindFirst() interface{} {
	prev := s.prevPipe()
	if len(prev) == 0 {
		return nil
	}
	return <-prev
}

// Element returns the element at the specified position in this stream
//
// nil is returned when index is out of range
func (s *Stream) Element(i int) interface{} {
	if i < 0 {
		return nil
	}

	var counter int

	prev := s.prevPipe()
	for v := range prev {
		if counter == i {
			return v
		}
		counter++
	}

	return nil
}
