package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsAny(t *testing.T) {
	assert.True(t, ContainsAny([]int{1, 2, 3}, 3, 4))
	assert.True(t, ContainsAny([]int{1, 2, 3}, 4, 1))
	assert.False(t, ContainsAny([]int{1, 2, 3}, 4, 5))
	assert.False(t, ContainsAny([]int{1, 2, 3}))
	assert.False(t, ContainsAny([]string{"a"}, "b"))
	assert.True(t, ContainsAny([]string{"a", "b"}, "b"))
}

func TestContainsAll(t *testing.T) {
	assert.True(t, ContainsAll([]int{1, 2, 3}, 1, 3))
	assert.True(t, ContainsAll([]int{1, 2, 3}, 3, 2, 1))
	assert.False(t, ContainsAll([]int{1, 2, 3}, 1, 4))
	assert.True(t, ContainsAll([]int{1, 2, 3}))
	assert.False(t, ContainsAll([]int(nil), 1))
	assert.True(t, ContainsAll([]int(nil)))
}

func TestContainsAnyAllEmptyCollection(t *testing.T) {
	// 空的接收切片：ContainsAny 恒为 false，ContainsAll 仅当 vals 也为空时为 true。
	var empty []int
	assert.False(t, ContainsAny(empty, 1))
	assert.False(t, ContainsAny(empty, 1, 2))
	assert.False(t, ContainsAll(empty, 1))
	assert.True(t, ContainsAll(empty))

	// nil 与空切片等价。
	assert.False(t, ContainsAny(nil, "x"))
	assert.True(t, ContainsAll[string](nil))
}

func TestContainsAnyAllDuplicateValues(t *testing.T) {
	// vals 中的重复元素不影响布尔结果，也不会被要求出现多次。
	assert.True(t, ContainsAny([]int{1, 2}, 2, 2))
	assert.True(t, ContainsAll([]int{1, 2}, 2, 2))
	assert.False(t, ContainsAll([]int{1, 2}, 2, 3, 3))
	assert.False(t, ContainsAny([]int{1, 2}, 3, 3))
}
