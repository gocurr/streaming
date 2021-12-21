package streaming

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"testing"
)

type Value struct {
	val float64
}

type Values []*Value

func (vs Values) Set(i int, v interface{}) {
	vs[i] = v.(*Value)
}

func (vs Values) Index(i int) interface{} {
	return vs[i]
}

func (vs Values) Len() int {
	return len(vs)
}

func (vs Values) Sub(i, j int) Slicer {
	return vs[i:j]
}

func Test_Values(t *testing.T) {
	var vs Slicer = Values{&Value{val: 1}, &Value{val: 2}, &Value{val: 3}}
	stream := Of(vs)
	stream.Limit(1).ForEach(func(i interface{}) {
		fmt.Printf("foreach: %v\n", *i.(*Value))
	})

	fmt.Println()
	stream = Of(vs)
	stream.Map(func(i interface{}) interface{} {
		return (*i.(*Value)).val * 100
	}).ForEach(func(i interface{}) {
		fmt.Printf("foreach: %v\n", i)
	})
	fmt.Printf("\ncount: %v\n\n", Of(vs).Count())
	fmt.Printf("filter: %v\n\n", Of(vs).Filter(func(i interface{}) bool {
		return i.(*Value).val > 1
	}).Count())

	Of(vs).Limit(3).ForEach(func(i interface{}) {
		fmt.Printf("foreach: %v\n", i.(*Value).val)
	})
}

var ints = Ints{1, 5, 7, 2, 8, 6, 9, 3, 9, 1}

func Test_Of(t *testing.T) {
	s := Of(ints)
	fmt.Printf("%v\n", s)
}

func Test_Filter(t *testing.T) {
	s := Of(ints)
	filter := s.Filter(func(i interface{}) bool {
		return i.(int) > 5
	})
	filter.ForEach(func(i interface{}) {
		fmt.Println(i)
	})
}

func Test_Collect(t *testing.T) {
	s := Of(ints)
	collect := s.Filter(func(i interface{}) bool {
		return i.(int) > 2
	}).Collect()
	fmt.Printf("%v\n", collect)
}

func Test_ForEach(t *testing.T) {
	s := Of(ints)
	s.ForEach(func(i interface{}) {
		fmt.Printf("%v\n", i)
	})
}

func Test_Limit(t *testing.T) {
	s := Of(ints)
	s.Limit(3).ForEach(func(i interface{}) {
		fmt.Printf("%v\n", i)
	})
}

func Test_Array(t *testing.T) {
	var arr = ints
	s := Of(arr)
	s.Map(func(i interface{}) interface{} {
		return math.Pow(float64(i.(int)), 2)
	}).ForEach(func(i interface{}) {
		fmt.Printf("%v\n", i)
	})
	fmt.Println()
	println(s.Count())
}

func TestStream_Reduce(t *testing.T) {
	s := Of(ints)

	reduce := s.Reduce(func(a, b interface{}) bool {
		return a.(int) < b.(int)
	})
	fmt.Println(reduce)
}

func Test_Distinct(t *testing.T) {
	stream := Of(ints)
	stream.Distinct().ForEach(func(i interface{}) {
		fmt.Printf("%v\n", i)
	})
}

func Test_Sum(t *testing.T) {
	stream := Of(ints)
	sum := stream.Sum(func(i interface{}) float64 {
		return float64(i.(int))
	})
	fmt.Printf("%v\n", sum)
}

func Test_Match(t *testing.T) {
	println(Of(ints).AnyMatch(func(i interface{}) bool {
		return i.(int) > 2
	}))

	println(Of(ints).AllMatch(func(i interface{}) bool {
		return i.(int) > 2
	}))

	println(Of(ints).NonMatch(func(i interface{}) bool {
		return i.(int) > 2
	}))
}

func Test_IsEmpty(t *testing.T) {
	println(Of(nil).IsEmpty())
	println(Of(ints).IsEmpty())
}

func Test_FlatMap(t *testing.T) {
	stream := Of(words)
	flatMap := stream.FlatMap(func(i interface{}) Slicer {
		split := strings.Split(i.(string), " ")
		return Strings(split)
	})
	flatMap.ForEach(func(i interface{}) {
		fmt.Printf("%v\n", i)
	})
}

var words = Strings{"one", "two", "three good"}

func Test_Peek(t *testing.T) {
	stream := Of(words)
	collect := stream.Peek(func(i interface{}) {
		fmt.Printf("%v is consumed\n", i)
	}).Collect()
	fmt.Printf("%v\n", collect)
}

func Test_Skip(t *testing.T) {
	stream := Of(ints)
	collect := stream.Skip(1).Collect()
	fmt.Printf("%v\n", collect)
}

func Test_FilterCount(t *testing.T) {
	stream := Of(ints)
	println(stream.FilterCount(func(i interface{}) bool {
		return i.(int) > 1
	}))
}

func Test_FindFirst(t *testing.T) {
	stream := Of(ints)
	first := stream.FindFirst()
	fmt.Printf("%v\n", first)
}

func Test_std_Sort(t *testing.T) {
	s := Of(ints)
	c := s.Collect()
	sort.Slice(c, func(i, j int) bool {
		return c.Index(i).(int) > c.Index(j).(int)
	})
	fmt.Printf("%v\n", c)

	sort.Slice(ints, func(i, j int) bool {
		return ints[i] < ints[j]
	})
	fmt.Printf("%v\n", ints)
}

func Test_Element(t *testing.T) {
	s := Of(ints)
	fmt.Printf("%v\n", s.Element(1))
}

var floats = Floats{1, 2, 2, 3, 3, 3, 3, 3, 3, 4, 4, 4, 4, 4, 44, 5, 5, 5}

func Test_TopN(t *testing.T) {
	top := Of(floats).Top(1).Collect()
	fmt.Printf("%v\n", top)
}

func TestCopy(t *testing.T) {
	of := Of(ints)
	cof := of.Copy()

	of.Limit(2).ForEach(func(i interface{}) {
		fmt.Println(i)
	})

	fmt.Println()

	cof.Limit(3).ForEach(func(i interface{}) {
		fmt.Println(i)
	})
}
