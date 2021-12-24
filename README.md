# The Stream Api For Golang

`streaming` is a serial of `go` Apis that work like stream in `java`.

## Download and Install

```bash
go get -u github.com/gocurr/streaming
```

## Usage

```go
package main

import (
	"fmt"
	"github.com/gocurr/streaming"
	"io/ioutil"
	"strings"
)

func main() {
	file, err := ioutil.ReadFile("testdata/Isaac.Newton-Opticks.txt")
	if err != nil {
		return
	}

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
```