package streaming

// Ints integer slice
type Ints []int

func (is Ints) Set(i int, v interface{}) {
	is[i] = v.(int)
}

func (is Ints) Index(i int) interface{} {
	return is[i]
}

func (is Ints) Len() int {
	return len(is)
}

func (is Ints) Sub(i, j int) Slicer {
	return is[i:j]
}

// Strings string slice

type Strings []string

func (s Strings) Set(i int, v interface{}) {
	s[i] = v.(string)
}

func (s Strings) Index(i int) interface{} {
	return s[i]
}

func (s Strings) Len() int {
	return len(s)
}

func (s Strings) Sub(i, j int) Slicer {
	return s[i:j]
}
