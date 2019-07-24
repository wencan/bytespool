package bytespool

import (
	"fmt"
	"math"
	"sync"
)

const (
	// indexLength maximum length of the index
	indexLength = 1034

	// littleCapacityUpper upper limit of the capacity of little bytes
	littleCapacityUpper = 1024

	// largeCapacityUpper upper limit of the capacity of large bytes
	largeCapacityUpper = 1024 * 1024

	// littleIndexUpper maximum length of the index of little bytes
	littleIndexUpper = 10
)

var (
	// DefaultPool default bytes pool
	DefaultPool = GetPool()

	// EmptyBytes represents empty bytes
	EmptyBytes = make([]byte, 0)

	poolPool = sync.Pool{
		New: func() interface{} {
			return new(Pool)
		},
	}
)

// GetPool acquire a bytes pool
func GetPool() *Pool {
	return poolPool.Get().(*Pool)
}

// PutPool release a bytes pool
func PutPool(pool *Pool) {
	poolPool.Put(pool)
}

// Get acquire a bytes slice with a capacity of at least length from default pool.
func Get(length int) []byte {
	return DefaultPool.Get(length)
}

// Put release a bytes slice.
func Put(bytes []byte) {
	DefaultPool.Put(bytes)
}

// Pool represents bytes pool
type Pool struct {
	pools       [indexLength]*sync.Pool
	capacities  [indexLength]int
	newPoolMutx sync.Mutex
}

func (pool *Pool) getCapacity(idx int) int {
	if idx <= littleIndexUpper {
		return int(math.Pow(2, float64(idx)))
	} else if idx < indexLength {
		return 1024 * (idx - littleIndexUpper + 1)
	}
	panic(fmt.Sprintf("index must be smaller than %d", indexLength))
}

func (pool *Pool) getIndex(length int) int {
	if length <= littleCapacityUpper {
		return int(math.Ceil(math.Log2(float64(length))))
	} else if length <= largeCapacityUpper {
		return int(math.Ceil(float64(length)/1024)) + littleIndexUpper - 1
	}
	panic(fmt.Sprintf("length must be smaller or equal than %d", largeCapacityUpper))
}

func (pool *Pool) findIndex(capacity int) (int, bool) {
	if capacity > largeCapacityUpper {
		return 0, false
	} else if capacity <= 0 {
		return 0, false
	} else if capacity <= littleCapacityUpper {
		idx := math.Log2(float64(capacity))
		if math.Mod(idx, 1) > 0 {
			return 0, false
		}
		return int(idx), true
	} else { // littleCapacityUpper < capacity <= largeCapacityUpper
		idx := float64(capacity) / 1024
		if math.Mod(idx, 1) > 1 {
			return 0, false
		}
		return int(math.Ceil(idx)) + littleIndexUpper - 1, true
	}
}

func (pool *Pool) getPool(idx int) *sync.Pool {
	p := pool.pools[idx]
	if p == nil {
		pool.newPoolMutx.Lock()
		defer pool.newPoolMutx.Unlock()

		p = pool.pools[idx]
		if p == nil {
			capacity := pool.getCapacity(idx)
			p = &sync.Pool{
				New: func() interface{} {
					return make([]byte, 0, capacity)
				},
			}

			pool.pools[idx] = p
			pool.capacities[idx] = capacity
		}
	}
	return p
}

// Get acquire a slice with a len of length and a capacity of at least length.
func (pool *Pool) Get(length int) []byte {
	var bytes []byte
	if length > largeCapacityUpper {
		bytes = make([]byte, 0, length)
	} else if length == 0 {
		bytes = EmptyBytes
	} else if length < 0 {
		panic("length must be greater than 0")
	} else {
		idx := pool.getIndex(length)
		p := pool.getPool(idx)
		bytes = p.Get().([]byte)
	}

	return bytes[:length]
}

// Put release a bytes slice.
func (pool *Pool) Put(bytes []byte) {
	capacity := cap(bytes)
	idx, found := pool.findIndex(capacity)
	if !found {
		return
	}

	if capacity != pool.capacities[idx] {
		panic(fmt.Sprintf("capacity missmatch, want: %d, have: %d, index: %d", pool.capacities[idx], capacity, idx))
	}

	p := pool.getPool(idx)
	bytes = bytes[:0]
	p.Put(bytes)
}
