package genetics

type Set[T comparable] struct {
	items map[T]bool
}

func NewSet[T comparable](items ...T) Set[T] {
	var set Set[T]
	set.items = make(map[T]bool)
	set.Fill(items...)
	return set
}

func (s Set[T]) Add(item T) {
	s.items[item] = true
}

func (s Set[T]) Fill(items ...T) {
	for _, i := range items {
		s.Add(i)
	}
}

func (s Set[T]) Equal(other Set[T]) bool {
	for item := range other.items {
		if !s.Contains(item) {
			return false
		}
	}
	return true
}

func (s Set[T]) Contains(item T) bool {
	_, ok := s.items[item]
	return ok
}

func (s Set[T]) Len() int {
	return len(s.items)
}

func (s Set[T]) Remove(item T) {
	delete(s.items, item)
}

func (s Set[T]) Union(other Set[T]) Set[T] {
	set := NewSet[T]()
	items := make([]T, 0, len(s.items))
	set.Fill(items...)
	items = make([]T, 0, len(other.items))
	set.Fill(items...)
	return set
}

func (s Set[T]) Intersection(other Set[T]) Set[T] {
	set := NewSet[T]()
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

func (s Set[T]) Subset(filter func(T) bool) Set[T] {
	set := NewSet[T]()
	items := make([]T, 0, len(s.items))
	for _, k := range items {
		if filter(k) {
			set.Add(k)
		}
	}

	return set
}

func (s Set[T]) Difference(other Set[T]) Set[T] {
	set := NewSet[T]()
	for item := range s.items {
		if !other.Contains(item) {
			set.Add(item)
		}
	}
	return set
}

func (s Set[T]) Reduce(reduce func(T, T) T) T {
	var carry T

	for item := range s.items {
		carry = reduce(item, carry)
	}

	return carry
}

func (s Set[T]) ToSlice() []T {
	items := []T{}
	for item := range s.items {
		items = append(items, item)
	}
	return items
}
