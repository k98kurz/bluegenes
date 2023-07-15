package genetics

type set[T comparable] struct {
	items map[T]bool
}

func newSset[T comparable](items ...T) set[T] {
	var set set[T]
	set.items = make(map[T]bool)
	set.Fill(items...)
	return set
}

func (s set[T]) Add(item T) {
	s.items[item] = true
}

func (s set[T]) Fill(items ...T) {
	for _, i := range items {
		s.Add(i)
	}
}

func (s set[T]) Equal(other set[T]) bool {
	for item := range other.items {
		if !s.Contains(item) {
			return false
		}
	}
	return true
}

func (s set[T]) Contains(item T) bool {
	_, ok := s.items[item]
	return ok
}

func (s set[T]) Len() int {
	return len(s.items)
}

func (s set[T]) Remove(item T) {
	delete(s.items, item)
}

func (s set[T]) Union(other set[T]) set[T] {
	set := newSset[T]()
	items := make([]T, 0, len(s.items))
	set.Fill(items...)
	items = make([]T, 0, len(other.items))
	set.Fill(items...)
	return set
}

func (s set[T]) Intersection(other set[T]) set[T] {
	set := newSset[T]()
	items := make([]T, len(s.items))
	set.Fill(items...)
	to_remove := make([]T, 0)

	for k := range set.items {
		_, ok := other.items[k]
		if !ok {
			to_remove = append(to_remove, k)
		}
	}

	for _, item := range to_remove {
		set.Remove(item)
	}

	return set
}

func (s set[T]) Subset(filter func(T) bool) set[T] {
	set := newSset[T]()
	items := make([]T, 0, len(s.items))
	for _, k := range items {
		if filter(k) {
			set.Add(k)
		}
	}

	return set
}

func (s set[T]) Difference(other set[T]) set[T] {
	set := newSset[T]()
	for item := range s.items {
		if !other.Contains(item) {
			set.Add(item)
		}
	}
	return set
}

func (s set[T]) Reduce(reduce func(T, T) T) T {
	var carry T

	for item := range s.items {
		carry = reduce(item, carry)
	}

	return carry
}

func (s set[T]) ToSlice() []T {
	items := []T{}
	for item := range s.items {
		items = append(items, item)
	}
	return items
}
