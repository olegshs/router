package helpers

import (
	"sort"

	"golang.org/x/exp/constraints"
)

type Map[K constraints.Ordered, V any] map[K]V

func (m Map[K, V]) Keys() []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

func (m Map[K, V]) SortedKeys() []K {
	keys := m.Keys()
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	return keys
}
