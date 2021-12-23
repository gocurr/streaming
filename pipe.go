package streaming

// returns previous pipe
func (s *Stream) prevPipe() chan interface{} {
	var ch chan interface{}
	if len(s.pipeline) == 0 {
		ch = s.curPipe()
		if s.slice.Len() > 0 {
			ch <- s.slice.Index(0)
		}
		go func() {
			for i := 1; i < s.slice.Len(); i++ {
				ch <- s.slice.Index(i)
			}
			close(ch)
		}()
	} else {
		ch = s.pipeline[len(s.pipeline)-1]
	}
	return ch
}

// returns current pipe
func (s *Stream) curPipe() chan interface{} {
	cur := make(chan interface{}, defaultChanBufSize)
	s.pipeline = append(s.pipeline, cur)
	return cur
}
