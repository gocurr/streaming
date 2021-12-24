package streaming

// prevPipe returns the previous pipe(buffered channel) in the pipeline.
//
// If the pipeline is empty, it will make a new channel and push elements
// of slice to the new channel (in a new goroutine). After the slice elements all consumed,
// the new channel will be closed.
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

// curPipe appends a new pipe(buffered channel) to the pipeline
// and returns the new pipe.
func (s *Stream) curPipe() chan interface{} {
	cur := make(chan interface{}, defaultChanBufSize)
	s.pipeline = append(s.pipeline, cur)
	return cur
}
