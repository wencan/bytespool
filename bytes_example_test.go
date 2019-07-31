package bytespool_test

import (
	"fmt"

	"github.com/wencan/bytespool"
)

func ExampleGet() {
	bytes := bytespool.GetBytes(100)
	fmt.Printf("len: %d, cap: %d", len(bytes), cap(bytes))
	bytespool.PutBytes(bytes)

	// Output:
	// len: 100, cap: 128
}
