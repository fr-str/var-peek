package safe

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/Main-Kube/util"
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

// Create new map with initial data
func NewMap[K comparable, V any](data map[K]V) *Map[K, V] {
	return &Map[K, V]{data: data}
}

func (m *Map[K, V]) ReadOnly() *Map[K, V] {
	m.readonly = true
	return m
}

func (m *Map[K, V]) init() error {
	if m == nil {
		return errors.New("map is nil")
	}
	if m.data == nil {
		m.data = map[K]V{}
	}

	return nil
}

// Safely return key existance
func (m *Map[K, V]) Exists(k K) bool {
	if m == nil {
		return false
	}

	m.lock.RLock()
	defer m.lock.RUnlock()
	_, exists := m.data[k]
	return exists
}

// Safely return value for key
func (m *Map[K, V]) Get(k K) V {
	if m == nil {
		return *new(V)
	}

	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.data[k]
}

// Safely return value and existance of key
func (m *Map[K, V]) GetFull(k K) (obj V, exists bool) {
	if m == nil {
		return
	}

	m.lock.RLock()
	defer m.lock.RUnlock()
	obj, exists = m.data[k]
	return
}

// Safely set value for key
func (m *Map[K, V]) Set(k K, v V) {
	if m.readonly || m.init() != nil {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	m.data[k] = v
}

// Safely delete key from map
func (m *Map[K, V]) Delete(k K) {
	if m == nil || m.readonly {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.data, k)
}

// Safely run function with direct access to map
func (m *Map[K, V]) Commit(fn func(data map[K]V)) {
	if m.readonly || m.init() != nil {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	fn(m.data)
}

// Return iterator for safe iterating over map
func (m *Map[K, V]) Iter() Iterator[K, V] {
	if m == nil {
		return nil
	}

	iter := make(chan Item[K, V], len(m.data))

	m.lock.RLock()
	go func() {
		defer m.lock.RUnlock()

		for k, v := range m.data {
			iter <- Item[K, V]{k, v}
		}
		close(iter)
	}()

	return iter
}

// Safely range over map
func (m *Map[K, V]) ForEach(fn func(k K, v V)) {
	if m == nil {
		return
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	for k, v := range m.data {
		fn(k, v)
	}
}

// Safely return all map keys
func (m *Map[K, V]) Keys() (keys []K) {
	if m == nil {
		return
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	i := 0
	keys = make([]K, len(m.data))

	for k := range m.data {
		keys[i] = k
		i++
	}

	return
}

// Safely return all map values
func (m *Map[K, V]) Values() (values []V) {
	if m == nil {
		return
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	i := 0
	values = make([]V, len(m.data))

	for _, v := range m.data {
		values[i] = v
		i++
	}

	return
}

// Safely return map length
func (m *Map[K, V]) Len() int {
	if m == nil {
		return 0
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	return len(m.data)
}

func (m *Map[K, V]) Copy() *Map[K, V] {
	if m == nil {
		return nil
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	copy := util.DeepCopy(m.data)
	if copy == nil {
		return nil
	}
	return NewMap(*copy)
}

// Safely marsahal map to JSON
func (m *Map[K, V]) MarshalJSON() ([]byte, error) {
	// init empty map
	if err := m.init(); err != nil {
		return nil, err
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	return json.Marshal(m.data)
}

// Safely unmarsahal JSON to map
func (m *Map[K, V]) UnmarshalJSON(data []byte) error {
	if m.readonly {
		return errors.New("map is readonly")
	}
	// init empty map
	if err := m.init(); err != nil {
		return err
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	return json.Unmarshal(data, &m.data)
}
