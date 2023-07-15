package bluegenes

type set[T comparable] struct {
	items map[T]bool
}

func newSet[T comparable](items ...T) set[T] {
	var set set[T]
	set.items = make(map[T]bool)
	set.fill(items...)
	return set
}

func (s set[T]) add(item T) {
	s.items[item] = true
}

func (s set[T]) fill(items ...T) {
	for _, i := range items {
		s.add(i)
	}
}

func (s set[T]) equal(other set[T]) bool {
	for item := range other.items {
		if !s.contains(item) {
			return false
		}
	}
	return true
}

func (s set[T]) contains(item T) bool {
	_, ok := s.items[item]
	return ok
}

func (s set[T]) len() int {
	return len(s.items)
}

func (s set[T]) remove(item T) {
	delete(s.items, item)
}

func (s set[T]) union(other set[T]) set[T] {
	set := newSet[T]()
	items := make([]T, 0, len(s.items))
	set.fill(items...)
	items = make([]T, 0, len(other.items))
	set.fill(items...)
	return set
}

func (s set[T]) intersection(other set[T]) set[T] {
	set := newSet[T]()
	items := make([]T, len(s.items))
	set.fill(items...)
	to_remove := make([]T, 0)

	for k := range set.items {
		_, ok := other.items[k]
		if !ok {
			to_remove = append(to_remove, k)
		}
	}

	for _, item := range to_remove {
		set.remove(item)
	}

	return set
}

func (s set[T]) subset(filter func(T) bool) set[T] {
	set := newSet[T]()
	items := make([]T, 0, len(s.items))
	for _, k := range items {
		if filter(k) {
			set.add(k)
		}
	}

	return set
}

func (s set[T]) difference(other set[T]) set[T] {
	set := newSet[T]()
	for item := range s.items {
		if !other.contains(item) {
			set.add(item)
		}
	}
	return set
}

func (s set[T]) reduce(reduce func(T, T) T) T {
	var carry T

	for item := range s.items {
		carry = reduce(item, carry)
	}

	return carry
}

func (s set[T]) toSlice() []T {
	items := []T{}
	for item := range s.items {
		items = append(items, item)
	}
	return items
}
