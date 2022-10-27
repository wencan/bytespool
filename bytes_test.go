package bytespool

import (
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
)

func TestSpecialMalloc(t *testing.T) {
	var pool BytesPool
	for _, length := range []int{0, 1, littleCapacityUpper, littleCapacityUpper + 1, largeCapacityUpper, largeCapacityUpper + 1, 1024 * 1024 * 1024} {
		bytes := pool.Get(length)
		if len(bytes) != length {
			t.Fatalf("bytes length error, want: %d, have: %d", length, len(bytes))
		}
		pool.Put(bytes)
	}
}

func TestBatchMalloc(t *testing.T) {
	rand.Seed(time.Now().Unix())

	var pool BytesPool
	length := 0

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
}

type samplePool struct {
	bytess [][]byte

	calloc func(size int) []byte

	size int
}

func (p *samplePool) Get() interface{} {
	if len(p.bytess) > 0 {
		bytes := p.bytess[0]
		p.bytess = p.bytess[1:]
		return bytes
	}
	if p.calloc != nil {
		return p.calloc(p.size)
	}
	return make([]byte, 0, p.size)
}

func (p *samplePool) Put(x interface{}) {
	bytes := x.([]byte)
	p.bytess = append(p.bytess, bytes)
}

func TestBytesReuse(t *testing.T) {
	var counter uint64
	var pool BytesPool
	pool.SizedPoolFactory = func(size int) Pool {
		return &samplePool{
			calloc: func(size int) []byte {
				atomic.AddUint64(&counter, 1)
				return make([]byte, 0, size)
			},
			bytess: make([][]byte, 0),
			size:   size,
		}
	}

	rand.Seed(time.Now().Unix())
	length := rand.Intn(littleCapacityUpper) + 1

	bytes1 := pool.Get(length)
	pool.Put(bytes1)

	bytes2 := pool.Get(length)
	pool.Put(bytes2)

	if counter != 1 {
		t.Fatalf("counter error, want: %d, have: %d", 1, counter)
		return
	}
}
