package streaming_test

import (
	"fmt"
	"github.com/gocurr/streaming"
	"io/ioutil"
	"log"
	"strings"
)

func read() []byte {
	file, err := ioutil.ReadFile("testdata/Isaac.Newton-Opticks.txt")
	if err != nil {
		log.Fatalln(err)
	}
	return file
}
func handle(file []byte) {
	lines := strings.Split(string(file), "\n")
	stream := streaming.Of(streaming.Strings(lines))

	stream.
		FlatMap(func(i interface{}) streaming.Slicer {
			words := strings.Split(i.(string), " ")
			return streaming.Strings(words)
		}).
		Filter(func(i interface{}) bool {
			s := i.(string)
			return s == "be" || s == "or" || s == "not" || s == "to"
		}).
		Top(5).
		ForEach(func(i interface{}) {
			v := i.(*streaming.CountVal)
			fmt.Printf("%s ", v.Val)
		})
	fmt.Println("?")
}

func Example_stream() {
	handle(read())
	// Output: to be or not ?
}

func _Example_largeFile() {
	file := read()
	for i := 0; i < 10; i++ {
		// enlarge the testdata
		// size = original_size << 10 ≈ 550 MB
		file = append(file, file...)
	}
	// about 1.5 GB memory cost in general
	handle(file)
	// Output: to be or not ?
}
