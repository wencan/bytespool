package bytespool_test

import (
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wencan/bytespool"
)

func BenchmarkBufferWriteStrings(b *testing.B) {
	str := []string{
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit",
		"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
		`Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris
		nisi ut aliquip ex ea commodo consequat.
		Duis aute irure dolor in reprehenderit in voluptate velit esse cillum
		dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident,
		sunt in culpa qui officia deserunt mollit anim id est laborum`,
		"Sed ut perspiciatis",
		"sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt",
		"Ut enim ad minima veniam, quis nostrum exercitationem ullam corporis suscipit",
		"laboriosam, nisi ut aliquid ex ea commodi consequatur",
		"Quis autem vel eum iure reprehenderit qui in ea voluptate velit esse quam nihil molestiae consequatur",
		"vel illum qui dolorem eum fugiat quo voluptas nulla pariatur",
	}

	// a clean pool
	var pool bytespool.BytesPool

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buffer := bytespool.GetBuffer()
			buffer.BytesPool = &pool
			buffer.MinGrowLength = 1024

			for _, s := range str {
				buffer.WriteString(s)
			}

			bytespool.PutBuffer(buffer)
		}
	})
}

func BenchmarkBufferWriteRandomTop1K(b *testing.B) {
	// a clean pool
	var pool bytespool.BytesPool

	// random data
	rand.Seed(time.Now().Unix())
	var bytess [1024][]byte
	for idx, _ := range bytess {
		length := rand.Intn(1024) + 1
		bytes := bytespool.GetBytes(length)
		nWrote, _ := rand.Read(bytes)
		bytess[idx] = bytes[:nWrote]
	}

	var index uint32

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx := atomic.AddUint32(&index, 1) - 1
			idx = idx % 1024

			buffer := bytespool.GetBuffer()
			buffer.BytesPool = &pool
			buffer.Write(bytess[idx])

			bytespool.PutBuffer(buffer)
		}
	})

	b.StopTimer()
	for _, bytes := range bytess {
		bytespool.PutBytes(bytes)
	}
}
