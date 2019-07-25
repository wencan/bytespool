# bytespool

## feature

- reuse byte slice based on the range of its length.

- reclaim the original byte slice when the buffer grows.

## usage

### bytes
```go
bytes := bytespool.Get(100)
fmt.Printf("len: %d, cap: %d", len(bytes), cap(bytes))
bytespool.Put(bytes)
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
buff := bytespool.Get(10)
defer bytespool.Put(buff)
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
