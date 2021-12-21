package examples

import (
	"fmt"
	"github.com/gocurr/streaming"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

func Test_stream(t *testing.T) {
	file, err := ioutil.ReadFile("Isaac.Newton-Opticks.txt")
	if err != nil {
		return
	}

	since := time.Now()
	being := streaming.Of(streaming.Strings{string(file)}).
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
	fmt.Printf("%v, took %v\n", being, time.Since(since))
}
