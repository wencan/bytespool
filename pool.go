package bytespool

import (
	"fmt"
	"math"
	"sync"
)

const (
	// IndexLength maximum length of the index 1 2 4 8 16 32 64 128 256 512 1024
	IndexLength = 1034

	// LittleCapacityUpper upper limit of the capacity of little bytes
	LittleCapacityUpper = 1024

	// LargeCapacityUpper upper limit of the capacity of large bytes
	LargeCapacityUpper = 1024 * 1024

	// LittleIndexUpper maximum length of the index of little bytes
	LittleIndexUpper = 10
)

var (
	// DefaultPool default bytes pool
	DefaultPool = GetPool()

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

// Get acquire a bytes slice with a capacity of at least size from default pool.
func Get(size int) []byte {
	return DefaultPool.Get(size)
}

// Put release a bytes slice.
func Put(bytes []byte) {
	DefaultPool.Put(bytes)
}

// Pool represents bytes pool
type Pool struct {
	pools       [IndexLength]*sync.Pool
	capacities  [IndexLength]int
	newPoolMutx sync.Mutex
}

func (pool *Pool) getCapacity(idx int) int {
	if idx <= LittleIndexUpper {
		return int(math.Pow(2, float64(idx)))
	} else if idx < IndexLength {
		return 1024 * (idx - LittleIndexUpper + 1)
	}
	panic(fmt.Sprintf("index must be smaller than %d", IndexLength))
}

func (pool *Pool) getIndex(size int) int {
	if size <= LittleCapacityUpper {
		return int(math.Ceil(math.Log2(float64(size))))
	} else if size <= LargeCapacityUpper {
		return int(math.Ceil(float64(size)/1024)) + LittleIndexUpper - 1
	}
	panic(fmt.Sprintf("size must be smaller or equal than %d", LargeCapacityUpper))
}

func (pool *Pool) findIndex(capacity int) (int, bool) {
	if capacity > LargeCapacityUpper {
		return 0, false
	} else if capacity <= 0 {
		return 0, false
	} else if capacity <= LittleCapacityUpper {
		idx := math.Log2(float64(capacity))
		if math.Mod(idx, 1) > 0 {
			return 0, false
		}
		return int(idx), true
	} else { // LittleCapacityUpper < capacity <= LargeCapacityUpper
		idx := float64(capacity) / 1024
		if math.Mod(idx, 1) > 1 {
			return 0, false
		}
		return int(math.Ceil(idx)) + LittleIndexUpper - 1, true
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

// Get acquire a bytes slice with a capacity of at least size.
func (pool *Pool) Get(size int) []byte {
	if size > LargeCapacityUpper {
		return make([]byte, 0, size)
	} else if size == 0 {
		return []byte{}
	} else if size < 0 {
		panic("size must be greater than 0")
	}

	idx := pool.getIndex(size)
	p := pool.getPool(idx)
	return p.Get().([]byte)
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
