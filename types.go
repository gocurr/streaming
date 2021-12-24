package streaming

// Ints integer slice.
type Ints []int

func (is Ints) Index(i int) interface{} {
	return is[i]
}

func (is Ints) Len() int {
	return len(is)
}

// Floats float64 slice.
type Floats []float64

func (fs Floats) Index(i int) interface{} {
	return fs[i]
}

func (fs Floats) Len() int {
	return len(fs)
}

// Strings string slice.
type Strings []string

func (ss Strings) Index(i int) interface{} {
	return ss[i]
}

func (ss Strings) Len() int {
	return len(ss)
}
