package streaming

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

type Value struct {
	val float64
}

type Values []*Value

func (vs Values) Index(i int) interface{} {
	return vs[i]
}

func (vs Values) Len() int {
	return len(vs)
}

func collect(got *[]interface{}, v interface{}) {
	*got = append(*got, v)
}

func Test_Values_Strings(t *testing.T) {
	vs := Values{
		&Value{val: 1},
		&Value{val: 2},
		&Value{val: 3},
	}
	stream := Of(vs)
	v(stream, t)
	stream = OfWithOption(vs, &Option{ChanBufSize: 10})
	v(stream, t)

	words := Strings{"one", "two", "two", "three", "good go"}
	stream = Of(words)
	s(stream, t)
	stream = OfWithOption(words, &Option{ChanBufSize: 10, Timeout: 1 * time.Second})
	s(stream, t)
}

func v(stream *Stream, t *testing.T) {
	var got []interface{}
	stream.Map(func(i interface{}) interface{} {
		return (*i.(*Value)).val * 100
	}).ForEach(func(i interface{}) {
		collect(&got, i)
	})

	stream.Copy().Filter(func(i interface{}) bool {
		return i.(*Value).val > 1
	}).ForEach(func(i interface{}) {
		collect(&got, i.(*Value).val)
	})

	collect(&got, stream.Copy().Count())

	collect(&got, stream.Copy().Sum(func(i interface{}) float64 {
		return i.(*Value).val
	}))

	stream.Copy().Limit(2).ForEach(func(i interface{}) {
		collect(&got, i.(*Value).val)
	})

	stream.Copy().Peek(func(i interface{}) {
		// do nothing
	}).ForEach(func(i interface{}) {
		collect(&got, i.(*Value).val)
	})

	stream.Copy().Skip(2).ForEach(func(i interface{}) {
		collect(&got, i.(*Value).val)
	})

	collect(&got, stream.Copy().Element(1).(*Value).val)

	var want = []interface{}{
		100.0, 200.0, 300.0, // map * 100
		2.0, 3.0, // filter > 1
		3,        // count
		6.0,      // sum
		1.0, 2.0, // limit 2
		1.0, 2.0, 3.0, // peek
		3.0, // skip 2
		2.0, // element 1
	}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func s(stream *Stream, t *testing.T) {
	var got []interface{}
	stream.Distinct().ForEach(func(i interface{}) {
		collect(&got, i)
	})

	stream.Copy().FlatMap(func(i interface{}) Slicer {
		return Strings(strings.Split(i.(string), " "))
	}).ForEach(func(i interface{}) {
		collect(&got, i)
	})

	stream.Copy().Top(1).ForEach(func(i interface{}) {
		collect(&got, i.(*CountVal).Val)
	})

	allMatch := stream.Copy().AllMatch(func(i interface{}) bool {
		w := i.(string)
		return len(w) > 3
	})
	collect(&got, allMatch)

	nonMatch := stream.Copy().NonMatch(func(i interface{}) bool {
		w := i.(string)
		return w == "two"
	})
	collect(&got, nonMatch)

	anyMatch := stream.Copy().AnyMatch(func(i interface{}) bool {
		w := i.(string)
		return w == "two"
	})
	collect(&got, anyMatch)

	var want = []interface{}{
		"one", "two", "three", "good go", // distinct
		"one", "two", "two", "three", "good", "go", // flatmap
		"two", // top 1
		false, // allMatch
		false, // nonMatch
		true,  // anyMatch
	}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
