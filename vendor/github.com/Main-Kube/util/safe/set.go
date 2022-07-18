package safe

import (
	"fmt"
	"sync"

	"github.com/Main-Kube/util/set"
)

type Set[T comparable] struct {
	super set.Set[T]
	lock  sync.RWMutex
}

func NewSet[T comparable](data ...T) *Set[T] {
	set := &Set[T]{}
	// skip mutex, when initializing
	set.super.Add(data...)
	return set
}

func (set *Set[T]) Add(values ...T) {
	if set == nil {
		return
	}

	set.lock.Lock()
	defer set.lock.Unlock()

	set.super.Add(values...)
}

func (set *Set[T]) Exist(value T) bool {
	if set == nil {
		return false
	}

	set.lock.RLock()
	defer set.lock.RUnlock()

	return set.super.Exist(value)
}

func (set *Set[T]) List() (list []T) {
	if set == nil {
		return
	}

	set.lock.RLock()
	defer set.lock.RUnlock()

	return set.super.List()
}

func (set *Set[T]) Length() int {
	if set == nil {
		return 0
	}

	set.lock.RLock()
	defer set.lock.RUnlock()

	return set.super.Length()
}

func (set *Set[T]) String() string {
	if set == nil {
		return "[]"
	}

	set.lock.RLock()
	defer set.lock.RUnlock()

	return fmt.Sprint(set.super.List())
}

// Safely marsahal set to JSON
func (set *Set[T]) MarshalJSON() ([]byte, error) {
	if set == nil {
		return nil, fmt.Errorf("set is nil")
	}

	set.lock.RLock()
	defer set.lock.RUnlock()
	return set.super.MarshalJSON()
}

// Safely unmarsahal JSON to set
func (set *Set[T]) UnmarshalJSON(data []byte) error {
	if set == nil {
		return fmt.Errorf("set is nil")
	}

	set.lock.Lock()
	defer set.lock.Unlock()
	return set.super.UnmarshalJSON(data)
}
