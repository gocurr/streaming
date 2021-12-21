package streaming

// Slicer interface type for streaming
type Slicer interface {
	Index(i int) interface{}
	Len() int
}

// Slice alias of interface slice
type Slice []interface{}

// Index returns element in the specific position
func (s Slice) Index(i int) interface{} {
	return s[i]
}

// Len returns length of Slicer
func (s Slice) Len() int {
	return len(s)
}
