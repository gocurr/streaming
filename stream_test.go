package streaming

import (
	"reflect"
	"strings"
	"testing"
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
	stream = OfWithChanBufSize(vs, 10)
	v(stream, t)

	words := Strings{"one", "two", "two", "three", "good go"}
	stream = Of(words)
	s(stream, t)
	stream = OfWithChanBufSize(words, 10)
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

	var want = []interface{}{
		"one", "two", "three", "good go", // distinct
		"one", "two", "two", "three", "good", "go", // flatmap
		"two", // top 1
	}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
