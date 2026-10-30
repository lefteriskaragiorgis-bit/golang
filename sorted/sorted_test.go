package sorted_test

import (
	"testing"

	"github.com/bitfield/sorted"
)

// func TestIsSorted_IsTrueForSortedSlice(t *testing.T) {
// 	t.Parallel()
// 	input := []int{1, 2, 3}
// 	if !sorted.IsSorted(input) {
// 		t.Errorf("got false for %v", input)
// 	}
// }

// func TestIsSorted_IsFalseForUnsortedSlice(t *testing.T) {
// 	t.Parallel()
// 	input := []int{1, 3, 2}
// 	if sorted.IsSorted(input) {
// 		t.Errorf("got true for %v", input)
// 	}
// }

// func TestIsSorted_IsTrueForSortedSlicesWith(t *testing.T) {
// 	t.Parallel()
// 	inputs := map[string][]int{
// 		"all elements different": {1, 2, 3},
// 		"repeated elements":      {1, 2, 2},
// 		"negative first element": {-1, 2, 3},
// 		"no elements":            {},
// 		"one element":            {1},
// 	}
// 	for name, input := range inputs {
// 		t.Run(name, func(t *testing.T) {
// 			if !sorted.IsSorted(input) {
// 				t.Errorf("got false for %v", input)
// 			}
// 		})
// 	}
// }

func TestIsSorted_IsTrueForSortedSlicesWith(t *testing.T) {
	t.Parallel()

	test := func(description string, input []int, want bool) {
		t.Run(description, func(t *testing.T) {
			got := sorted.IsSorted(input)
			if got != want {
				t.Errorf("IsSorted(%v) = %v, want %v", input, got, want)
			}
		})
	}

	test("all elements different", []int{1, 2, 3}, true)
	test("repeated elements", []int{1, 2, 2}, true)
	test("negative first element", []int{-1, 2, 3}, true)
	test("no elements", []int{}, true)
	test("one element", []int{1}, true)
}

func TestIsSorted_IsFalseForUnsortedSlicesWith(t *testing.T) {
	t.Parallel()

	test := func(description string, input []int, want bool) {
		t.Run(description, func(t *testing.T) {
			got := sorted.IsSorted(input)
			if got != want {
				t.Errorf("IsSorted(%v) = %v, want %v", input, got, want)
			}
		})
	}

	test("all elements different", []int{1, 3, 2}, false)
	test("repeated elements", []int{3, 2, 2}, false)
}
