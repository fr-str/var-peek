package peek

import (
	"errors"
	"sort"
	"sync"

	"golang.org/x/exp/constraints"
)

type Iterator[K comparable, V any] <-chan Item[K, V]

type Item[K comparable, V any] struct {
	Key   K
	Value V
}

type Map[K comparable, V any] struct {
	data     map[K]V
	lock     sync.RWMutex
	readonly bool
}

type SortFunction[K constraints.Ordered] func(data []K, i, j int) bool

type SortedMap[K constraints.Ordered, V any] struct {
	Map[K, V]
	lessFunc SortFunction[K]

	sorted     bool
	sortedKeys []K
}

// Create new map with initial data
func NewSortedMap[K constraints.Ordered, V any](data map[K]V, lessFunc SortFunction[K]) *SortedMap[K, V] {
	return &SortedMap[K, V]{
		Map: Map[K, V]{
			data: data,
		},
		lessFunc: lessFunc,
	}
}

func (m *SortedMap[K, V]) ReadOnly() *SortedMap[K, V] {
	m.readonly = true
	return m
}

func (m *SortedMap[K, V]) init() error {
	if m == nil {
		return errors.New("map is nil")
	}

	if m.data == nil {
		m.data = map[K]V{}
	}

	if m.lessFunc == nil {
		m.lessFunc = func(data []K, i, j int) bool {
			return data[i] <= data[j]
		}
	}

	return nil
}

func (m *SortedMap[K, V]) sort() {
	if m == nil {
		return
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	// skip sort if sorted
	if m.sorted {
		return
	}

	m.init()

	// get keys slice
	i := 0
	m.sortedKeys = make([]K, len(m.data))
	for k := range m.data {
		m.sortedKeys[i] = k
		i++
	}

	// sort keys
	sort.Slice(m.sortedKeys, func(i, j int) bool { return m.lessFunc(m.sortedKeys, i, j) })
	m.sorted = true
}

// Safely set value for key
func (m *SortedMap[K, V]) Set(k K, v V) {
	if m.readonly || m.init() != nil {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	m.data[k] = v
	m.sorted = false
}

// Return iterator for safe iterating over map
func (m *SortedMap[K, V]) Iter() Iterator[K, V] {
	if m == nil {
		return nil
	}

	// sort before returning
	m.sort()

	iter := make(chan Item[K, V], len(m.sortedKeys))

	m.lock.RLock()
	go func() {
		defer m.lock.RUnlock()

		for _, k := range m.sortedKeys {
			iter <- Item[K, V]{k, m.data[k]}
		}
		close(iter)
	}()

	return iter
}

// Safely return map length
func (m *SortedMap[K, V]) Len() int {
	if m == nil {
		return 0
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	return len(m.data)
}
