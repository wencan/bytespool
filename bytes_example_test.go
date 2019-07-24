package bytespool_test

import (
	"fmt"

	"github.com/wencan/bytespool"
)

func ExampleGet() {
	bytes := bytespool.Get(100)
	fmt.Printf("len: %d, cap: %d", len(bytes), cap(bytes))
	bytespool.Put(bytes)

	// Output:
	// len: 100, cap: 128
}
