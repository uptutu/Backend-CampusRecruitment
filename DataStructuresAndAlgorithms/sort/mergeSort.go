package main

import (
	"fmt"
)

func mergeSort(arr []int) []int {
	if len(arr) < 2 {
		return arr
	}

	left := mergeSort(arr[:len(arr)/2])
	right := mergeSort(arr[len(arr)/2:])
	return merge(left, right)
}

func merge(left, right []int) []int {
	result := make([]int, len(left)+len(right))
	r := 0
	for n, m := 0, 0; n < len(left) || m < len(right); {
		if n == len(left) {
			for ; m < len(right); m++ {
				result[r] = right[m]
				r++
			}
			break
		}
		if m == len(right) {
			for ; n < len(left); n++ {
				result[r] = left[n]
				r++
			}
			break
		}
		if left[n] < right[m] {
			result[r] = left[n]
			r++
			n++
		} else {
			result[r] = right[m]
			r++
			m++
		}
	}

	return result
}

func main() {
	// 测试代码
	arr := []int{9, 8, 7, 6, 5, 1, 2, 3, 4, 0}
	fmt.Println(arr)
	fmt.Println(mergeSort(arr))
}
