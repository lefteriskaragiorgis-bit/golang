package sorted

import "slices"

// func IsSorted(input []int) bool {
// 	if len(input) < 2 {
// 		return true
// 	}

// 	prev := input[0]
// 	for _, v := range input[1:] {
// 		if v < prev {
// 			return false
// 		}
// 		prev = v
// 	}
// 	return true
// }

func IsSorted(input []int) bool {
	return slices.IsSorted(input)
}
