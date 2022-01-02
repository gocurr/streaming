package streaming_test

import (
	"fmt"
	"github.com/gocurr/streaming"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

func read() []byte {
	file, err := ioutil.ReadFile("testdata/Isaac.Newton-Opticks.txt")
	if err != nil {
		log.Fatalln(err)
	}
	return file
}

func handle(file []byte, op *streaming.Option) {
	lines := strings.Split(string(file), "\n")
	slicer := streaming.Strings(lines)
	stream := streaming.OfWithOption(slicer, op)

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
	handle(read(), nil)
	// Output: to be or not ? true
}

func Example_largeFile() {
	file := read()
	for i := 0; i < 10; i++ {
		// Enlarge the testdata.
		// size = original_size << 10 ≈ 550 MB
		file = append(file, file...)
	}
	// About 1.5 GB memory cost in general.
	handle(file, &streaming.Option{Timeout: 1 * time.Second})
	// Output: ? false
}
