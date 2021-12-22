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

	streaming.
		Of(streaming.Strings{string(file)}).
		FlatMap(func(i interface{}) streaming.Slicer {
			return streaming.Strings(strings.Split(i.(string), "\n"))
		}).
		FlatMap(func(i interface{}) streaming.Slicer {
			return streaming.Strings(strings.Split(i.(string), " "))
		}).
		Filter(func(i interface{}) bool {
			s := i.(string)
			return s == "be" || s == "or" || s == "not" || s == "to"
		}).
		Top(5).
		ForEach(func(i interface{}) {
			v := i.(streaming.CountVal)
			fmt.Printf("%s ", v.Val)
		})
	fmt.Println("?")
}
```