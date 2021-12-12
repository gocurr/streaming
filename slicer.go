package streaming

// Slicer interface type for streaming
type Slicer interface {
	Set(i int, v interface{})
	Index(i int) interface{}
	Len() int
	Sub(i, j int) Slicer
}

// Slice alias of interface slice
type Slice []interface{}

// Set element in the specific position
func (s Slice) Set(i int, v interface{}) {
	s[i] = v
}

// Index returns element in the specific position
func (s Slice) Index(i int) interface{} {
	return s[i]
}

// Len returns length
func (s Slice) Len() int {
	return len(s)
}

// Sub returns sub Slicer with range from i to j
func (s Slice) Sub(i, j int) Slicer {
	return s[i:j]
}
