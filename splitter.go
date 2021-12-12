package streaming

import (
	"github.com/gocurr/partition"
	"runtime"
)

// cpu number
var cpu = runtime.NumCPU()

// Ranges split parts by cpu
type Ranges []partition.Range

// splitSlicer splits Slicer by cpu
func splitSlicer(slice Slicer) Ranges {
	if slice == nil {
		return nil
	}
	return splitSize(slice.Len())
}

// splits size by cpu
func splitSize(size int) Ranges {
	if size < 1 {
		return nil
	}
	return partition.RangesN(size, cpu)
}
