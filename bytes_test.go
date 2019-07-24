package bytespool

import (
	"math/rand"
	"runtime/debug"
	"testing"
	"time"
)

func TestSpecialMalloc(t *testing.T) {
	pool := GetPool()
	for _, length := range []int{0, 1, littleCapacityUpper, littleCapacityUpper + 1, largeCapacityUpper, largeCapacityUpper + 1, 1024 * 1024 * 1024} {
		bytes := pool.Get(length)
		if len(bytes) != length {
			t.Fatalf("bytes length error, want: %d, have: %d", length, len(bytes))
		}
		pool.Put(bytes)
	}
	PutPool(pool)
}

func TestBatchMalloc(t *testing.T) {
	rand.Seed(time.Now().Unix())

	pool := GetPool()
	length := 0

	defer func() {
		r := recover()
		if r != nil {
			t.Fatalf("panic at length = %d, recover: %v\n%s", length, r, string(debug.Stack()))
		}
	}()

	for length < 1024*1025 {
		bytes := pool.Get(length)
		if len(bytes) != length {
			t.Fatalf("length error, want: %d, have: %d", length, len(bytes))
			return
		}

		if length < 1024 {
			length++
		} else {
			length += rand.Intn(1024) + 1024
		}
	}
	PutPool(pool)
}
