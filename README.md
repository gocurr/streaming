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
	"time"
)

func main() {
	file, err := ioutil.ReadFile("Isaac.Newton-Opticks.txt")
	if err != nil {
		return
	}

	since := time.Now()
	reduce := streaming.
		Of(streaming.Strings{string(file)}).
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
	fmt.Printf("%v, took %v\n", reduce, time.Since(since))
}

```