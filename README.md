# bytespool
[![GoDoc](https://godoc.org/github.com/wencan/bytespool?status.svg)](https://godoc.org/github.com/wencan/bytespool)

## feature

- reuse byte slice based on the range of its length.

- reclaim the original byte slice when the buffer grows.

## benchmark
### plan one
```
go test github.com/wencan/bytespool -bench=. -benchmem
```
```
BenchmarkBufferWriteStrings-16        	100000000	        17.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkBufferWriteRandomTop1K-16    	50000000	        25.1 ns/op	       0 B/op	       0 allocs/op
```

### plan two
```
 go test github.com/wencan/go-benchmark/... -bench=. -benchmem
```
```
BenchmarkGenericBuf-16                         	 2000000	       735 ns/op	    2576 B/op	       4 allocs/op
BenchmarkGenericStackBuf-16                    	 2000000	       768 ns/op	    2576 B/op	       4 allocs/op
BenchmarkAllocBuf-16                           	 2000000	       773 ns/op	    2576 B/op	       4 allocs/op
BenchmarkSyncPoolBuf-16                        	100000000	        15.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkBpoolPoolBuf-16                       	 2000000	       814 ns/op	       0 B/op	       0 allocs/op
BenchmarkByteBufferPoolBuf-16                  	100000000	        28.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkEasyJsonBuffer-16                     	10000000	       180 ns/op	     609 B/op	       4 allocs/op
BenchmarkEasyJsonBuffer_OptimizedConfig-16     	50000000	        29.5 ns/op	      32 B/op	       1 allocs/op
BenchmarkBytesPoolBuffer-16                    	100000000	        15.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkBytesPoolBuffer_OptimizedConfig-16    	100000000	        15.5 ns/op	       0 B/op	       0 allocs/op
```

## usage

### bytes
```go
bytes := bytespool.GetBytes(100)
fmt.Printf("len: %d, cap: %d", len(bytes), cap(bytes))
bytespool.PutBytes(bytes)
```
output:
```
len: 100, cap: 128
```

### buffer
```go
buffer := bytespool.GetBuffer()
defer bytespool.PutBuffer(buffer)

_, err := buffer.Write([]byte{0, 1, 2, 3})
fmt.Printf("len: %d, cap: %d\n", buffer.Len(), buffer.Cap())

// auto growth
_, err = buffer.Write([]byte{4, 5, 6, 7, 8, 9})
fmt.Printf("len: %d, cap: %d\n", buffer.Len(), buffer.Cap())

// read
buff := bytespool.GetBytes(10)
defer bytespool.PutBytes(buff)
nRead, err := buffer.Read(buff)
buff = buff[:nRead]

fmt.Printf("data: %X", buff)
```
Output:
```
len: 4, cap: 4
len: 10, cap: 16
data: 00010203040506070809
```
