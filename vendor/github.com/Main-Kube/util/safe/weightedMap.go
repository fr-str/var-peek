package safe

import (
	"encoding/json"
	"errors"
	"sort"

	"github.com/Main-Kube/util"
	"github.com/Main-Kube/util/set"
)

type Weighted[V any] struct {
	Value  V
	Weight uint8
}

type WeightedMap[K comparable, V any] struct {
	Map[K, Weighted[V]]

	sorted     bool
	sortedKeys []K

	weightCount int
}

// Create new map with initial data
func NewWeightedMap[K comparable, V any](data map[K]Weighted[V]) *WeightedMap[K, V] {
	return &WeightedMap[K, V]{
		Map: Map[K, Weighted[V]]{
			data: data,
		},
	}
}

func (m *WeightedMap[K, V]) ReadOnly() *WeightedMap[K, V] {
	m.readonly = true
	return m
}

func (m *WeightedMap[K, V]) init() error {
	if m == nil {
		return errors.New("map is nil")
	}

	if m.data == nil {
		m.data = map[K]Weighted[V]{}
	}

	return nil
}

// sort keys
func (m *WeightedMap[K, V]) sort() {
	if m == nil {
		return
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	// skip sort if sorted
	if m.sorted {
		return
	}

	weights := set.New[uint8]()

	// get keys slice
	i := 0
	m.sortedKeys = make([]K, len(m.data))
	for k, v := range m.data {
		weights.Add(v.Weight)
		m.sortedKeys[i] = k
		i++
	}

	m.weightCount = weights.Length()

	// sort keys
	sort.Slice(m.sortedKeys, func(i, j int) bool {
		a := m.data[m.sortedKeys[i]]
		b := m.data[m.sortedKeys[j]]
		return a.Weight > b.Weight
	})

	m.sorted = true
}

// Safely return value for key
func (m *WeightedMap[K, V]) Get(k K) V {
	if m == nil {
		return *new(V)
	}

	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.data[k].Value
}

// Safely return value and existance of key
func (m *WeightedMap[K, V]) GetFull(k K) (obj V, exists bool) {
	if m == nil {
		return
	}

	m.lock.RLock()
	defer m.lock.RUnlock()
	weight, exists := m.data[k]
	return weight.Value, exists
}

// Safely set value for key
func (m *WeightedMap[K, V]) Set(k K, v Weighted[V]) {
	if m.readonly || m.init() != nil {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	m.data[k] = v
	m.sorted = false
}

// Safely delete key from map
func (m *WeightedMap[K, V]) Delete(k K) {
	if m == nil || m.readonly {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.data, k)
	m.sorted = false
}

// Safely run function with direct access to map
func (m *WeightedMap[K, V]) Commit(fn func(data map[K]Weighted[V])) {
	if m.readonly || m.init() != nil {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	fn(m.data)
	m.sorted = false
}

// Return safe sorted iterator
func (m *WeightedMap[K, V]) Iter() Iterator[K, V] {
	if m == nil {
		return nil
	}

	// sort before returning
	m.sort()

	iter := make(chan Item[K, V], len(m.data))

	m.lock.RLock()
	go func() {
		defer m.lock.RUnlock()

		for _, k := range m.sortedKeys {
			iter <- Item[K, V]{k, m.data[k].Value}
		}
		close(iter)
	}()

	return iter
}

func (m *WeightedMap[K, V]) WeightIter() <-chan Iterator[K, V] {
	if m == nil {
		return nil
	}

	// sort before returning
	m.sort()

	weightChan := make(chan Iterator[K, V], m.weightCount)

	m.lock.RLock()
	go func() {
		defer m.lock.RUnlock()

		lastWeight := -1
		var iter chan Item[K, V]

		for _, k := range m.sortedKeys {
			v := m.data[k]
			if lastWeight != int(v.Weight) {
				// Close previous
				if iter != nil {
					close(iter)
				}

				// Optimisation: set correct length
				iter = make(chan Item[K, V], len(m.data))
				weightChan <- iter
				lastWeight = int(v.Weight)
			}

			iter <- Item[K, V]{k, v.Value}
		}
		close(iter)
		close(weightChan)
	}()

	return weightChan
}

// Safely range over sorted map
func (m *WeightedMap[K, V]) ForEach(fn func(k K, v V)) {
	if m == nil {
		return
	}

	// sort before returning
	m.sort()

	m.lock.RLock()
	defer m.lock.RUnlock()

	for _, k := range m.sortedKeys {
		fn(k, m.data[k].Value)
	}
}

// Safely return all map keys
func (m *WeightedMap[K, V]) Keys() (keys []K) {
	if m == nil {
		return
	}

	// sort before returning
	m.sort()

	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.sortedKeys
}

// Safely return all sorted map values
func (m *WeightedMap[K, V]) Values() (values []V) {
	if m == nil {
		return
	}

	// sort before returning
	m.sort()

	m.lock.RLock()
	defer m.lock.RUnlock()

	i := 0
	values = make([]V, len(m.data))

	for _, k := range m.sortedKeys {
		values[i] = m.data[k].Value
		i++
	}

	return
}

func (m *WeightedMap[K, V]) Copy() *WeightedMap[K, V] {
	if m == nil {
		return nil
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	copy := util.DeepCopy(m.data)
	if copy == nil {
		return nil
	}
	return NewWeightedMap(*copy)
}

// Safely marsahal map to JSON
func (m *WeightedMap[K, V]) MarshalJSON() ([]byte, error) {
	// init empty map
	if err := m.init(); err != nil {
		return nil, err
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	simpleData := map[K]V{}

	for k, v := range m.data {
		simpleData[k] = v.Value
	}

	return json.Marshal(simpleData)
}

// Safely unmarsahal JSON to map
func (m *WeightedMap[K, V]) UnmarshalJSON(data []byte) error {
	if m.readonly {
		return errors.New("map is readonly")
	}
	// init empty map
	if err := m.init(); err != nil {
		return err
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	// TODO: try unmarshal weighted values
	// var finalMap map[K]WeightedValue[V]
	// err := json.Unmarshal(data, &finalMap)
	// if err == nil {
	// 	// OK
	// 	m.data = finalMap
	// 	m.sort()
	// 	return nil
	// }

	// unmarshal weightless values
	var simpleMap map[K]V
	err := json.Unmarshal(data, &simpleMap)
	if err != nil {
		return err
	}

	// assign existing weights
	finalMap := map[K]Weighted[V]{}
	for k, v := range simpleMap {
		var weight uint8

		if x, exists := m.data[k]; exists {
			weight = x.Weight
		}

		finalMap[k] = Weighted[V]{v, weight}
	}

	// save final map
	m.data = finalMap

	return nil
}
