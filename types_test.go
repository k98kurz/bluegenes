package genetics

import (
	"fmt"
	"sort"
	"testing"
)

func TestSets(t *testing.T) {
	t.Run("NewLenContainsAddRemove", func(t *testing.T) {
		t.Parallel()
		var intSet Set[int]
		var floatSet Set[float32]
		var strSet Set[string]
		for i := 1; i < 444; i++ {
			ints := make([]int, i)
			floats := make([]float32, i)
			strs := make([]string, i)
			for k := 0; k < i; k++ {
				ints[k] = k
				floats[k] = float32(k)
				strs[k] = fmt.Sprint(k)
			}

			intSet = NewSet(ints...)
			floatSet = NewSet(floats...)
			strSet = NewSet(strs...)

			if intSet.Len() != i {
				t.Errorf("Set[int].Len=%v, expected %v", intSet.Len(), i)
			}
			if floatSet.Len() != i {
				t.Errorf("Set[float32].Len=%v, expected %v", floatSet.Len(), i)
			}
			if strSet.Len() != i {
				t.Errorf("Set[string].Len=%v, expected %v", strSet.Len(), i)
			}

			if !intSet.Contains(i - 1) {
				t.Errorf("!Set[int].Contains(%v) but should have", i-1)
			}
			if !floatSet.Contains(float32(i - 1)) {
				t.Errorf("!Set[float32].Contains(%v) but should have", i-1)
			}
			if !strSet.Contains(fmt.Sprint(i - 1)) {
				t.Errorf("!Set[string].Contains(%v) but should have", i-1)
			}

			l1 := intSet.Len()
			intSet.Add(i - 1)
			if l1 != intSet.Len() {
				t.Error("Set[int] added duplicate item")
			}

			intSet.Remove(i - 1)
			if l1 == intSet.Len() {
				t.Error("Set[int] failed to remove item")
			}
		}
	})

	t.Run("EqualUnionIntersectionSubsetDifference", func(t *testing.T) {
		t.Parallel()
		intSet := NewSet(1, 2, 3)
		copied := NewSet(1, 2, 3)
		intSet2 := NewSet(1, 4, 5)

		if !intSet.Equal(copied) {
			t.Error("Set[int].Equal produced invalid result")
		} else if intSet.Equal(intSet2) {
			t.Error("Set[int].Equal produced invalid result")
		}

		expected := NewSet(1, 2, 3, 4, 5)
		observed := intSet.Union(intSet2)
		if !expected.Equal(observed) {
			t.Error("Set[int].Union produced invalid result")
		}

		expected = NewSet(1)
		observed = intSet.Intersection(intSet2)
		if !expected.Equal(observed) {
			t.Error("Set[int].Union produced invalid result")
		}

		intSet = NewSet(1, 2, 3, 4, 5, 6, 7, 8)
		expected = NewSet(2, 4, 6, 8)
		observed = intSet.Subset(func(i int) bool {
			return i%2 == 0
		})
		if !expected.Equal(observed) {
			t.Error("Set[int].Subset produced invalid result")
		}

		smaller := NewSet(1, 2, 3)
		expected = NewSet(4, 5, 6, 7, 8)
		observed = intSet.Difference(smaller)
		if !expected.Equal(observed) {
			t.Error("Set[int].Difference produced invalid result")
		}
	})

	t.Run("Reduce", func(t *testing.T) {
		t.Parallel()
		for i := 3; i < 1_111; i++ {
			intSet := NewSet[int]()
			expected := 0
			for k := 1; k < i; k++ {
				intSet.Add(k)
				expected += k
			}
			observed := intSet.Reduce(func(i1, i2 int) int { return i1 + i2 })
			if expected != observed {
				t.Errorf("Set[int].Reduce produced invalid result; got %d, expected %d", observed, expected)
			}
		}
	})

	t.Run("ToSlice", func(t *testing.T) {
		t.Parallel()
		for i := 3; i < 111; i++ {
			intSet := NewSet[int]()
			expected := []int{}
			for k := 1; k < i; k++ {
				expected = append(expected, k)
			}
			intSet.Fill(expected...)
			observed := intSet.ToSlice()
			sort.Ints(observed)

			if !equal(expected, observed) {
				t.Fatalf("Set[int].ToSlice failed: expected %v, observed %v", expected, observed)
			}
		}
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("Min", func(t *testing.T) {
		t.Parallel()
		for i := 11; i < 1111; i++ {
			ints := []int{}
			for k := 1; k < i; k++ {
				ints = append(ints, k)
			}

			observed, err := Min(ints...)
			if err != nil {
				t.Errorf("Min failed with the following error: %v", err.Error())
			}
			if observed != 1 {
				t.Errorf("Min failed; got %d, expected 1", observed)
			}
		}
	})
	t.Run("Max", func(t *testing.T) {
		t.Parallel()
		for i := 11; i < 1111; i++ {
			ints := []int{}
			for k := i; k > 0; k-- {
				ints = append(ints, k)
			}

			observed, err := Max(ints...)
			if err != nil {
				t.Errorf("Max failed with the following error: %v", err.Error())
			}
			if observed != i {
				t.Errorf("Max failed; got %d, expected %d", observed, i)
			}
		}
	})
	t.Run("Contains", func(t *testing.T) {
		t.Parallel()
		for i := 11; i < 1111; i++ {
			ints := []int{}
			for k := i; k > 0; k-- {
				ints = append(ints, k)
			}
			if !contains(ints, i) {
				t.Error("Contains[int] failed to detect item")
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

	t.Run("SliceContains", func(t *testing.T) {
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
