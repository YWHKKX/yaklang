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
