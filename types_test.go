package bluegenes

import (
	"fmt"
	"sort"
	"testing"
)

func TestSets(t *testing.T) {
	t.Run("Newlencontainsaddremove", func(t *testing.T) {
		t.Parallel()
		var intSet set[int]
		var floatSet set[float32]
		var strSet set[string]
		for i := 1; i < 444; i++ {
			ints := make([]int, i)
			floats := make([]float32, i)
			strs := make([]string, i)
			for k := 0; k < i; k++ {
				ints[k] = k
				floats[k] = float32(k)
				strs[k] = fmt.Sprint(k)
			}

			intSet = newSet(ints...)
			floatSet = newSet(floats...)
			strSet = newSet(strs...)

			if intSet.len() != i {
				t.Errorf("Set[int].len=%v, expected %v", intSet.len(), i)
			}
			if floatSet.len() != i {
				t.Errorf("Set[float32].len=%v, expected %v", floatSet.len(), i)
			}
			if strSet.len() != i {
				t.Errorf("Set[string].len=%v, expected %v", strSet.len(), i)
			}

			if !intSet.contains(i - 1) {
				t.Errorf("!Set[int].contains(%v) but should have", i-1)
			}
			if !floatSet.contains(float32(i - 1)) {
				t.Errorf("!Set[float32].contains(%v) but should have", i-1)
			}
			if !strSet.contains(fmt.Sprint(i - 1)) {
				t.Errorf("!Set[string].contains(%v) but should have", i-1)
			}

			l1 := intSet.len()
			intSet.add(i - 1)
			if l1 != intSet.len() {
				t.Error("Set[int] added duplicate item")
			}

			intSet.remove(i - 1)
			if l1 == intSet.len() {
				t.Error("Set[int] failed to remove item")
			}
		}
	})

	t.Run("equalunionintersectionsubsetdifference", func(t *testing.T) {
		t.Parallel()
		intSet := newSet(1, 2, 3)
		copied := newSet(1, 2, 3)
		intSet2 := newSet(1, 4, 5)

		if !intSet.equal(copied) {
			t.Error("Set[int].equal produced invalid result")
		} else if intSet.equal(intSet2) {
			t.Error("Set[int].equal produced invalid result")
		}

		expected := newSet(1, 2, 3, 4, 5)
		observed := intSet.union(intSet2)
		if !expected.equal(observed) {
			t.Error("Set[int].union produced invalid result")
		}

		expected = newSet(1)
		observed = intSet.intersection(intSet2)
		if !expected.equal(observed) {
			t.Error("Set[int].union produced invalid result")
		}

		intSet = newSet(1, 2, 3, 4, 5, 6, 7, 8)
		expected = newSet(2, 4, 6, 8)
		observed = intSet.subset(func(i int) bool {
			return i%2 == 0
		})
		if !expected.equal(observed) {
			t.Error("Set[int].subset produced invalid result")
		}

		smaller := newSet(1, 2, 3)
		expected = newSet(4, 5, 6, 7, 8)
		observed = intSet.difference(smaller)
		if !expected.equal(observed) {
			t.Error("Set[int].difference produced invalid result")
		}
	})

	t.Run("reduce", func(t *testing.T) {
		t.Parallel()
		for i := 3; i < 1_111; i++ {
			intSet := newSet[int]()
			expected := 0
			for k := 1; k < i; k++ {
				intSet.add(k)
				expected += k
			}
			observed := intSet.reduce(func(i1, i2 int) int { return i1 + i2 })
			if expected != observed {
				t.Errorf("Set[int].reduce produced invalid result; got %d, expected %d", observed, expected)
			}
		}
	})

	t.Run("toSlice", func(t *testing.T) {
		t.Parallel()
		for i := 3; i < 111; i++ {
			intSet := newSet[int]()
			expected := []int{}
			for k := 1; k < i; k++ {
				expected = append(expected, k)
			}
			intSet.fill(expected...)
			observed := intSet.toSlice()
			sort.Ints(observed)

			if !equal(expected, observed) {
				t.Fatalf("Set[int].toSlice failed: expected %v, observed %v", expected, observed)
			}
		}
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("min", func(t *testing.T) {
		t.Parallel()
		for i := 11; i < 1111; i++ {
			ints := []int{}
			for k := 1; k < i; k++ {
				ints = append(ints, k)
			}

			observed, err := min(ints...)
			if err != nil {
				t.Errorf("min failed with the following error: %v", err.Error())
			}
			if observed != 1 {
				t.Errorf("min failed; got %d, expected 1", observed)
			}
		}
	})
	t.Run("max", func(t *testing.T) {
		t.Parallel()
		for i := 11; i < 1111; i++ {
			ints := []int{}
			for k := i; k > 0; k-- {
				ints = append(ints, k)
			}

			observed, err := max(ints...)
			if err != nil {
				t.Errorf("max failed with the following error: %v", err.Error())
			}
			if observed != i {
				t.Errorf("max failed; got %d, expected %d", observed, i)
			}
		}
	})
	t.Run("contains", func(t *testing.T) {
		t.Parallel()
		for i := 11; i < 1111; i++ {
			ints := []int{}
			for k := i; k > 0; k-- {
				ints = append(ints, k)
			}
			if !contains(ints, i) {
				t.Error("contains[int] failed to detect item")
			}
		}

	})
	t.Run("equal", func(t *testing.T) {
		t.Parallel()
		for i := 1; i < 1111; i++ {
			first := []int{}
			second := []int{}
			for k := 0; k < i; k++ {
				first = append(first, k)
				second = append(second, k)
			}

			if !equal(first, second) {
				t.Fatalf("equal failed on [%v], [%v]", first, second)
			}
		}
	})

	t.Run("Slicecontains", func(t *testing.T) {
		t.Parallel()
		slices := make([][]int, 5)
		for i := 0; i < 5; i++ {
			slice := []int{i, i + 1, i + 2}
			slices = append(slices, slice)
		}

		if !containsSlice(slices, []int{0, 1, 2}) {
			t.Fatalf("containsSlice failed to find slice [0 1 2]; slices = %v", slices)
		}
	})
}
