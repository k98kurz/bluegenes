package bluegenes

import "fmt"

type Integer interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64
}

type Float interface {
	float32 | float64
}

type Ordered interface {
	Integer | Float | ~string
}

type anError struct {
	message string
}

func (e anError) Error() string {
	return e.message
}

type indexError struct{}

func (e indexError) Error() string {
	return "index out of bounds"
}

type missingParameterError struct {
	parameter_name string
}

func (e missingParameterError) Error() string {
	return fmt.Sprintf("missing parameter %s", e.parameter_name)
}

type Option[T any] struct {
	isSet bool
	val   T
}

func (o Option[T]) ok() bool {
	return o.isSet
}

func NewOption[T any](val ...T) Option[T] {
	if len(val) > 0 {
		return Option[T]{
			isSet: true,
			val:   val[0],
		}
	}
	return Option[T]{isSet: false}
}

func min[T Ordered](items ...T) (T, error) {
	if len(items) < 1 {
		var empty T
		return empty, anError{"no items supplied"}
	}
	smallest := items[0]
	for i := 1; i < len(items); i++ {
		if items[i] < smallest {
			smallest = items[i]
		}
	}
	return smallest, nil
}

func max[T Ordered](items ...T) (T, error) {
	if len(items) < 1 {
		var empty T
		return empty, indexError{}
	}
	largest := items[0]
	for i := 1; i < len(items); i++ {
		if items[i] > largest {
			largest = items[i]
		}
	}
	return largest, nil
}

func reduce[T comparable](items []T, reduce func(T, T) T) T {
	var carry T
	for _, item := range items {
		carry = reduce(item, carry)
	}
	return carry
}

func equal[T comparable](slices ...[]T) bool {
	for i, slice := range slices {
		if i < 1 {
			continue
		}
		if len(slice) != len(slices[i-1]) {
			return false
		}
		for k, item := range slice {
			if slices[i-1][k] != item {
				return false
			}
		}
	}
	return true
}

func containsSlice[T comparable](slices [][]T, query []T) bool {
	for _, slice := range slices {
		if equal(slice, query) {
			return true
		}
	}
	return false
}

func contains[T comparable](list []T, query T) bool {
	// contains function adapted from https://stackoverflow.com/a/10485970
	for _, item := range list {
		if item == query {
			return true
		}
	}
	return false
}
