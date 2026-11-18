package utils

import "github.com/samber/lo"

// ContainsAny 判断 s 是否至少包含 vals 中的一个元素。
// vals 为空时返回 false。
func ContainsAny[T comparable](s []T, vals ...T) bool {
	return lo.Some(s, vals)
}

// ContainsAll 判断 s 是否包含 vals 中的全部元素。
// vals 为空时返回 true。
func ContainsAll[T comparable](s []T, vals ...T) bool {
	return lo.Every(s, vals)
}
