package streaming_test

import (
	"fmt"
	"github.com/gocurr/streaming"
	"io/ioutil"
	"strings"
)

func Example_stream() {
	file, err := ioutil.ReadFile("testdata/Isaac.Newton-Opticks.txt")
	if err != nil {
		return
	}

	be := streaming.Of(streaming.Strings{string(file)}).
		FlatMap(func(i interface{}) streaming.Slicer {
			return streaming.Strings(strings.Split(i.(string), "\n"))
		}).
		Filter(func(i interface{}) bool {
			return i.(string) != ""
		}).
		FlatMap(func(i interface{}) streaming.Slicer {
			return streaming.Strings(strings.Split(i.(string), " "))
		}).
		Top(100).Element(9)
	fmt.Println(be.(streaming.CountVal).Val)

	// Output: be
}
