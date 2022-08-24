package set

import (
	"encoding/json"
	"fmt"
)

// --- Set implemented using map ---

type void struct{}

type Set[T comparable] struct {
	data map[T]void
}

func New[T comparable](data ...T) *Set[T] {
	set := &Set[T]{}
	set.Add(data...)
	return set
}

func (set *Set[T]) init() error {
	if set == nil {
		return fmt.Errorf("set is nil")
	}
	if set.data == nil {
		set.data = map[T]void{}
	}

	return nil
}

func (set *Set[T]) Add(values ...T) {
	if set.init() != nil {
		return
	}

	for _, value := range values {
		set.data[value] = void{}
	}
}

func (set *Set[T]) Remove(values ...T) {
	if set.init() != nil {
		return
	}

	for _, value := range values {
		delete(set.data, value)
	}
}

func (set *Set[T]) Contains(value T) bool {
	if set == nil {
		return false
	}

	_, exists := set.data[value]
	return exists
}

func (set *Set[T]) List() (list []T) {
	if set == nil {
		return
	}

	i := 0
	list = make([]T, len(set.data))
	for value := range set.data {
		list[i] = value
		i++
	}
	return
}

func (set *Set[T]) Length() int {
	if set == nil {
		return 0
	}
	return len(set.data)
}

func (set *Set[T]) String() string {
	if set == nil {
		return "[]"
	}
	return fmt.Sprint(set.List())
}

// Safely marsahal set to JSON
func (set *Set[T]) MarshalJSON() ([]byte, error) {
	if set == nil {
		return nil, fmt.Errorf("set is nil")
	}
	return json.Marshal(set.List())
}

// Safely unmarsahal JSON to set
func (set *Set[T]) UnmarshalJSON(data []byte) error {
	if set == nil {
		return fmt.Errorf("set is nil")
	}

	var values []T
	err := json.Unmarshal(data, &values)
	if err != nil {
		return err
	}

	set.data = nil
	set.Add(values...)

	return nil
}
