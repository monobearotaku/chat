package slices

import "reflect"

func Contains[T comparable](slice []T, value T) bool {
	for _, item := range slice {
		if item == value || reflect.DeepEqual(item, value) {
			return true
		}
	}

	return false
}

func RemoveFunc[T any](slice []T, f func(i int) bool) []T {
	newSlice := make([]T, 0, len(slice))

	for i, item := range slice {
		if !f(i) {
			newSlice = append(newSlice, item)
		}
	}

	return newSlice
}

func FromElenent[T any](item T) []T {
	return []T{item}
}
