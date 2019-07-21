package bytespool

import (
	"math/rand"
	"runtime/debug"
	"testing"
	"time"
)

func TestSpecialMalloc(t *testing.T) {
	pool := GetPool()
	for _, size := range []int{0, 1, LittleCapacityUpper, LittleCapacityUpper + 1, LargeCapacityUpper, LargeCapacityUpper + 1, 1024 * 1024 * 1024} {
		bytes := pool.Get(size)
		pool.Put(bytes)
	}
	PutPool(pool)
}

func TestBatchMalloc(t *testing.T) {
	rand.Seed(time.Now().Unix())

	pool := GetPool()
	size := 0

	defer func() {
		r := recover()
		if r != nil {
			t.Fatalf("panic at size = %d, recover: %v\n%s", size, r, string(debug.Stack()))
		}
	}()

	for size < 1024*1025 {
		bytes := pool.Get(size)
		if len(bytes) != 0 {
			t.Fatalf("size: %d, want: 0, have: %d", size, len(bytes))
			return
		}
		if cap(bytes) < size {
			t.Fatalf("size: %d, want >= %d, have: %d", size, size, cap(bytes))
			return
		}

		if size < 1024 {
			size++
		} else {
			size += rand.Intn(100) + 10
		}
	}
	PutPool(pool)
}
