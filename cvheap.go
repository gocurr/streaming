package streaming

// CountVal wraps Count-Val
type CountVal struct {
	Count int
	Val   interface{}
}

// An cvHeap is a max-heap of CountValues.
type cvHeap []*CountVal

func (h cvHeap) Len() int           { return len(h) }
func (h cvHeap) Less(i, j int) bool { return h[i].Count > h[j].Count }
func (h cvHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *cvHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(*CountVal))
}

func (h *cvHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
