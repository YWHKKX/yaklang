package utils

import (
	"math/rand"
	"time"
)

// minOrdered 返回 x、y 中较小的一个。
// 本包内已有名为 min 的 time.Duration 辅助函数，会遮蔽内置 min，故这里用独立命名。
func minOrdered[T Ordered](x, y T) T {
	if x < y {
		return x
	}
	return y
}

// maxOrdered 返回 x、y 中较大的一个。
func maxOrdered[T Ordered](x, y T) T {
	if x > y {
		return x
	}
	return y
}

func Min(x, y int) int {
	return minOrdered(x, y)
}

func Max(x, y int) int {
	return maxOrdered(x, y)
}

func MinByte(x, y byte) byte {
	return minOrdered(x, y)
}

func MaxByte(x, y byte) byte {
	return maxOrdered(x, y)
}

func MaxInt64(x, y int64) int64 {
	return maxOrdered(x, y)
}

func MinInt64(x, y int64) int64 {
	return minOrdered(x, y)
}

func ShuffleString(slice []string) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for len(slice) > 0 {
		n := len(slice)
		randIndex := r.Intn(n)
		slice[n-1], slice[randIndex] = slice[randIndex], slice[n-1]
		slice = slice[:n-1]
	}
}

func ShuffleInt(slice []int) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for len(slice) > 0 {
		n := len(slice)
		randIndex := r.Intn(n)
		slice[n-1], slice[randIndex] = slice[randIndex], slice[n-1]
		slice = slice[:n-1]
	}
}
