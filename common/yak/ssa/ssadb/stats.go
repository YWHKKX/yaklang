package ssadb

import (
	"sync/atomic"
	"time"
)

var (
	_SSASaveTypeCost   uint64
	_SSAVariableCost   uint64
	_SSASourceCodeCost uint64
)

func GetSSASaveTypeCost() time.Duration {
	return time.Duration(atomic.LoadUint64(&_SSASaveTypeCost))
}

func GetSSAVariableCost() time.Duration {
	return time.Duration(atomic.LoadUint64(&_SSAVariableCost))
}

func GetSSASourceCodeCost() time.Duration {
	return time.Duration(atomic.LoadUint64(&_SSASourceCodeCost))
}
