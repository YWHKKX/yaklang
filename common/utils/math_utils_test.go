package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMinMax(t *testing.T) {
	assert.Equal(t, 1, Min(1, 2))
	assert.Equal(t, 1, Min(2, 1))
	assert.Equal(t, 2, Min(2, 2))

	assert.Equal(t, 2, Max(1, 2))
	assert.Equal(t, 2, Max(2, 1))
	assert.Equal(t, 2, Max(2, 2))
}

func TestMinMaxInt64(t *testing.T) {
	assert.Equal(t, int64(-5), MinInt64(-5, 3))
	assert.Equal(t, int64(-5), MinInt64(3, -5))
	assert.Equal(t, int64(3), MaxInt64(-5, 3))
	assert.Equal(t, int64(3), MaxInt64(3, -5))
	assert.Equal(t, int64(1<<40), MaxInt64(1<<40, 1))
}

func TestMinMaxByte(t *testing.T) {
	assert.Equal(t, byte(1), MinByte(1, 2))
	assert.Equal(t, byte(1), MinByte(2, 1))
	assert.Equal(t, byte(255), MaxByte(1, 255))
	assert.Equal(t, byte(255), MaxByte(255, 1))
}
