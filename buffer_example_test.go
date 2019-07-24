package bytespool_test

import (
	"fmt"

	"github.com/wencan/bytespool"
)

func ExampleBuffer() {
	buffer := bytespool.GetBuffer()
	defer bytespool.PutBuffer(buffer)

	_, err := buffer.Write([]byte{0, 1, 2, 3})
	_ = err
	fmt.Printf("len: %d, cap: %d\n", buffer.Len(), buffer.Cap())

	// auto growth
	_, err = buffer.Write([]byte{4, 5, 6, 7, 8, 9})
	_ = err
	fmt.Printf("len: %d, cap: %d\n", buffer.Len(), buffer.Cap())

	// read
	buff := bytespool.Get(10)
	defer bytespool.Put(buff)
	nRead, err := buffer.Read(buff)
	_ = err
	buff = buff[:nRead]

	fmt.Printf("data: %X", buff)

	// Output:
	// len: 4, cap: 4
	// len: 10, cap: 16
	// data: 00010203040506070809
}
