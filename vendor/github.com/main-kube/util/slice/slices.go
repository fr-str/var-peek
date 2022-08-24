package slice

import (
	"math"

	"golang.org/x/exp/constraints"
)

type numeric interface {
	constraints.Integer | constraints.Float
}

// Avg returns the average of a slice of numeric values
func Avg[T numeric](listOfNumbers []T) T {
	if len(listOfNumbers) == 0 {
		return 0
	}

	var sum T
	for _, v := range listOfNumbers {
		sum += v
	}
	return sum / T(len(listOfNumbers))
}

// Max returns the maximum value of a numeric slice.
// If the slice is empty, Max returns 0
func Max[T numeric](listOfNumbers []T) T {
	if len(listOfNumbers) == 0 {
		return 0
	}

	max := listOfNumbers[0]
	for _, v := range listOfNumbers {
		if v > max {
			max = v
		}
	}
	return max
}

func Equal[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func Contains[T comparable](slice []T, e T) bool {
	for _, n := range slice {
		if e == n {
			return true
		}
	}
	return false
}

// RemoveFromSlice removes item at index from slice,
// but changes the slice, so use only if order doesn't matter!
func RemoveIdx[T comparable](s *[]T, i int) {
	(*s)[i] = (*s)[len((*s))-1]
	(*s) = (*s)[:len((*s))-1]
}

// Remove removes first occurrence of e from slice,
// but changes the slice, so use only if order doesn't matter!
func Remove[T comparable](s *[]T, elem T) {
	for i, v := range *s {
		if v == elem {
			RemoveIdx(s, i)
			return
		}
	}
}

// zliczam ile wartosci z numbers jest powyżej wartości `limit`
func HitLimitRatio[T numeric](numbers []T, limit uint) uint {
	if len(numbers) == 0 {
		return 0
	}

	// limit jest pomniejszony o 10%, ponieważ to jest wystarczająco blisko
	nearLimit := float64(limit) * 0.9

	count := float64(0)
	for _, v := range numbers {
		if float64(v) > nearLimit {
			count++
		}
	}

	return uint(math.Ceil(count / float64(len(numbers)) * 100))
}

func Diff[T comparable](a, b []T) []T {
	if len(a) > len(b) {
		a, b = b, a
	}

	var diff []T
	var m = make(map[T]struct{}, len(a))
	for _, v := range a {
		m[v] = struct{}{}
	}
	for _, v := range b {
		if _, ok := m[v]; !ok {
			diff = append(diff, v)
			delete(m, v)
		}
	}
	for k := range m {
		diff = append(diff, k)
	}
	return diff
}
