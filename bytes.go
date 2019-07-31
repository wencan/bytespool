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
	// DefaultBytesPool is the default instance of BytesPool.
	DefaultBytesPool = GetBytesPool()

	// EmptyBytes represents empty bytes
	EmptyBytes = make([]byte, 0)

	// BytesPoolPool is a pool for BytesPool instance.
	BytesPoolPool Pool = &sync.Pool{
		New: func() interface{} { return new(BytesPool) },
	}

	// DefaultSizedBytesPoolFactory is a default factory for producing SizedBytesPool instance.
	DefaultSizedBytesPoolFactory = func(size int) Pool {
		return &sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, size)
			},
		}
	}
)

// GetBytesPool acquire a bytes pool
func GetBytesPool() *BytesPool {
	return BytesPoolPool.Get().(*BytesPool)
}

// PutBytesPool reset and release a bytes pool
func PutBytesPool(pool *BytesPool) {
	pool.Reset()
	BytesPoolPool.Put(pool)
}

// GetBytes is a quick method for DefaultBytesPool.Get.
func GetBytes(length int) []byte {
	return DefaultBytesPool.Get(length)
}

// PutBytes is a quick method for DefaultBytesPool.Put.
func PutBytes(bytes []byte) {
	DefaultBytesPool.Put(bytes)
}

// SizedBytesPoolFactory creates the pool for the bytes that length is fixed.
type SizedBytesPoolFactory func(length int) Pool

// BytesPool represents bytes pool
type BytesPool struct {
	// SizedPoolFactory is a factory for producing SizedBytesPool instance.
	// default is DefaultSizedBytesPoolFactory.
	SizedPoolFactory SizedBytesPoolFactory

	pools       [indexLength]Pool
	capacities  [indexLength]int
	newPoolMutx sync.Mutex
}

func (pool *BytesPool) getCapacity(idx int) int {
	if idx <= littleIndexUpper {
		return int(math.Pow(2, float64(idx)))
	} else if idx < indexLength {
		return 1024 * (idx - littleIndexUpper + 1)
	}
	panic(fmt.Sprintf("index must be smaller than %d", indexLength))
}

func (pool *BytesPool) getIndex(length int) int {
	if length <= littleCapacityUpper {
		return int(math.Ceil(math.Log2(float64(length))))
	} else if length <= largeCapacityUpper {
		return int(math.Ceil(float64(length)/1024)) + littleIndexUpper - 1
	}
	panic(fmt.Sprintf("length must be smaller or equal than %d", largeCapacityUpper))
}

func (pool *BytesPool) findIndex(capacity int) (int, bool) {
	if capacity > largeCapacityUpper {
		return 0, false
	} else if capacity <= 0 {
		return 0, false
	} else if capacity <= littleCapacityUpper {
		log2 := math.Log2(float64(capacity))
		idx := int(log2)
		if log2 != float64(idx) {
			return 0, false
		}
		return idx, true
	} else { // littleCapacityUpper < capacity <= largeCapacityUpper
		log := float64(capacity) / 1024
		idx := int(log)
		if log != float64(idx) {
			return 0, false
		}
		return idx + littleIndexUpper - 1, true
	}
}

func (pool *BytesPool) getPool(idx int) Pool {
	p := pool.pools[idx]
	if p == nil {
		pool.newPoolMutx.Lock()
		defer pool.newPoolMutx.Unlock()

		p = pool.pools[idx]
		if p == nil {
			capacity := pool.getCapacity(idx)

			if pool.SizedPoolFactory == nil {
				pool.SizedPoolFactory = DefaultSizedBytesPoolFactory
			}
			p = pool.SizedPoolFactory(capacity)

			pool.pools[idx] = p
			pool.capacities[idx] = capacity
		}
	}
	return p
}

func (pool *BytesPool) acquireBytes(length int) []byte {
	idx := pool.getIndex(length)
	p := pool.getPool(idx)
	return p.Get().([]byte)
}

// Get acquire a slice with a len of length and a capacity of at least length.
func (pool *BytesPool) Get(length int) []byte {
	var bytes []byte
	if length > largeCapacityUpper {
		bytes = make([]byte, 0, length)
	} else if length == 0 {
		bytes = EmptyBytes
	} else if length < 0 {
		panic("length must be greater than 0")
	} else {
		bytes = pool.acquireBytes(length)
	}

	return bytes[:length]
}

// Put reset and release a bytes slice.
func (pool *BytesPool) Put(bytes []byte) {
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

func (pool *BytesPool) Reset() {
	pool.SizedPoolFactory = nil
}
