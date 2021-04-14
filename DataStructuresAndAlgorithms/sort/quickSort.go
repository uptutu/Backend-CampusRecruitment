package main

import (
	"fmt"
	"math/rand"
	"time"
)

// quickSort 使用快速排序算法，排序整型数组
func quickSort(arr *[]int, start, end int) {
	if end-start < 1 {
		return
	}

	if end-start == 1 {
		if (*arr)[start] > (*arr)[end] {
			(*arr)[start], (*arr)[end] = (*arr)[end], (*arr)[start]
		}
		return
	}

	// 使用随机数优化快速排序
	rand.Seed(time.Now().Unix())
	pivotIndex := rand.Intn(end-start) + start
	pivot := (*arr)[pivotIndex]

	// 将选取的枢轴放置数组最后一位，然后对数组遍历将所有小于枢轴值的值前置
	(*arr)[end], (*arr)[pivotIndex] = (*arr)[pivotIndex], (*arr)[end]

	// j 记录 pivot 小的数字个数说形成的index，即用于将所有数组中小于 pivot 的数前置
	j := start
	for i := start; i < end; i++ {
		if (*arr)[i] < pivot && start != i {
			(*arr)[i], (*arr)[j] = (*arr)[j], (*arr)[i]
			j++
		}
	}

	// 遍历完成后的 j 值即 pivot 值应该在的地方
	(*arr)[j], (*arr)[end] = (*arr)[end], (*arr)[j]

	quickSort(arr, start, j-1)
	quickSort(arr, j+1, end)

}

func main() {
	// 测试代码
	arr := []int{9, 8, 7, 6, 5, 1, 2, 3, 4, 0}
	fmt.Println(arr)
	quickSort(&arr, 0, len(arr)-1)
	fmt.Println(arr)
}
