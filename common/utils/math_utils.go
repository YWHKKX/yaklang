package utils

import (
	"math/rand"
	"time"
)

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func MinByte(x, y byte) byte {
	return byte(Min(int(x), int(y)))
}

func MaxByte(x, y byte) byte {
	return byte(Max(int(x), int(y)))
}

func MaxInt64(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

func MinInt64(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

// shuffleSlice 使用 Fisher-Yates 算法原地打乱切片。
func shuffleSlice[T any](slice []T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for len(slice) > 0 {
		n := len(slice)
		randIndex := r.Intn(n)
		slice[n-1], slice[randIndex] = slice[randIndex], slice[n-1]
		slice = slice[:n-1]
	}
}

func ShuffleString(slice []string) {
	shuffleSlice(slice)
}

func ShuffleInt(slice []int) {
	shuffleSlice(slice)
}
