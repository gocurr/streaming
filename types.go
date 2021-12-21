package streaming

// Ints integer slice
type Ints []int

func (is Ints) Index(i int) interface{} {
	return is[i]
}

func (is Ints) Len() int {
	return len(is)
}

// Floats float slice
type Floats []float64

func (fs Floats) Index(i int) interface{} {
	return fs[i]
}

func (fs Floats) Len() int {
	return len(fs)
}

// Strings string slice

type Strings []string

func (s Strings) Index(i int) interface{} {
	return s[i]
}

func (s Strings) Len() int {
	return len(s)
}
