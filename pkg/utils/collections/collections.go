package collections

import (
	"fmt"
	"github.com/tuannh982/simple-workflow-go/pkg/utils/ptr"
)

func MapArray[A any, B any](a []A, f func(A) B) []B {
	result := make([]B, len(a))
	for i, v := range a {
		result[i] = f(v)
	}
	return result
}

func Map[A any, B any](a *A, f func(A) B) *B {
	if a == nil {
		return nil
	}
	return ptr.Ptr(f(*a))
}

func ToMap[K comparable, T any](extractor func(T) K, arr []T) map[K]T {
	result := make(map[K]T)
	for _, value := range arr {
		result[extractor(value)] = value
	}
	return result
}

func GroupBy[K comparable, T any](grouper func(T) K, arr []T) map[K][]T {
	result := make(map[K][]T)
	for _, value := range arr {
		if _, ok := result[grouper(value)]; !ok {
			result[grouper(value)] = make([]T, 0)
		}
		result[grouper(value)] = append(result[grouper(value)], value)
	}
	return result
}

func FirstInArray[T any](m []T) T {
	for _, v := range m {
		return v
	}
	panic(fmt.Sprintf("%v is empty", m))
}

func FirstInMap[K comparable, T any](m map[K]T) (K, T) {
	for k, v := range m {
		return k, v
	}
	panic(fmt.Sprintf("%v is empty", m))
}
