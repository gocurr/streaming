package streaming_test

import (
	"fmt"
	"github.com/gocurr/streaming"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

func slicer(shift int) streaming.Slicer {
	file, err := ioutil.ReadFile("testdata/Isaac.Newton-Opticks.txt")
	if err != nil {
		log.Fatalln(err)
	}
	for i := 0; i < shift; i++ {
		// Enlarge the testdata.
		file = append(file, file...)
	}
	lines := strings.Split(string(file), "\n")
	return streaming.Strings(lines)
}

func handle(stream *streaming.Stream) {
	top := stream.
		FlatMap(func(i interface{}) streaming.Slicer {
			words := strings.Split(i.(string), " ")
			return streaming.Strings(words)
		}).
		Filter(func(i interface{}) bool {
			s := i.(string)
			return s == "be" || s == "or" || s == "not" || s == "to"
		}).
		Top(5)
	top.ForEach(func(i interface{}) {
		v := i.(*streaming.CountVal)
		fmt.Printf("%s ", v.Val)
	})

	fmt.Printf("? %v", stream.Correct())
}

func Example_stream() {
	stream := streaming.Of(slicer(0))
	handle(stream)
	// Output: to be or not ? true
}

func Example_largeFile_Correct() {
	stream := streaming.Of(slicer(10))
	handle(stream)
	// Output: to be or not ? true
}

func Example_largeFile_Incorrect() {
	stream := streaming.OfWithOption(slicer(10), &streaming.Option{
		ChanBufSize: 1 << 10,
		Timeout:     1 * time.Second},
	)
	handle(stream)
	// Output: ? false
}
