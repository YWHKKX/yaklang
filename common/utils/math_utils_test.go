package utils

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShufflePreservesElements(t *testing.T) {
	ints := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	ShuffleInt(ints)
	sortedInts := append([]int{}, ints...)
	sort.Ints(sortedInts)
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, sortedInts)

	strs := []string{"a", "b", "c", "d", "e"}
	ShuffleString(strs)
	sortedStrs := append([]string{}, strs...)
	sort.Strings(sortedStrs)
	assert.Equal(t, []string{"a", "b", "c", "d", "e"}, sortedStrs)
}

func TestShuffleHandlesEmptyAndSingle(t *testing.T) {
	var empty []int
	ShuffleInt(empty)
	assert.Empty(t, empty)

	single := []string{"only"}
	ShuffleString(single)
	assert.Equal(t, []string{"only"}, single)
}
