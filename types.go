package bluegenes

import (
	"fmt"
	"math"
)

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
	IsSet bool
	Val   T
}

func (o Option[T]) Ok() bool {
	return o.IsSet
}

func NewOption[T any](val ...T) Option[T] {
	if len(val) > 0 {
		return Option[T]{
			IsSet: true,
			Val:   val[0],
		}
	}
	return Option[T]{IsSet: false}
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

func flipFloat32(f float32) float32 {
	// flip the bits of the float and return the result
	b := ^math.Float32bits(f)
	return math.Float32frombits(b)
}

func flipFloat64(f float64) float64 {
	// flip the bits of the float and return the result
	b := ^math.Float64bits(f)
	return math.Float64frombits(b)
}

func flipString(s string) string {
	// flip the bits and return the result
	result := []byte{}
	for i := 0; i < len(s); i++ {
		result = append(result, ^s[i])
	}
	return string(result)
}
