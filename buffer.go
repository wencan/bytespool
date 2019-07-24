package bytespool

import (
	"io"
	"sync"
)

const (
	minGrowLength = 64
)

var (
	bufferPool = sync.Pool{
		New: func() interface{} {
			return new(Buffer)
		},
	}

	defaultBytesPool = DefaultPool
)

// BytesPool acquire and release bytes.
// Interface of bytes pool. default is DefaultPool.
type BytesPool interface {
	Get(size int) []byte
	Put(bytes []byte)
}

// GetBuffer acquire a buffer base on default pool.
func GetBuffer() *Buffer {
	return bufferPool.Get().(*Buffer)
}

// PutBuffer reset and release buffer.
func PutBuffer(buffer *Buffer) {
	buffer.Reset()
	bufferPool.Put(buffer)
}

// Buffer get bytes from pool and put idle bytes to pool.
type Buffer struct {
	bytes []byte

	readOffset int

	pool BytesPool
}

// SetAllocator change bytes pool.
func (buffer *Buffer) SetAllocator(pool BytesPool) {
	buffer.pool = pool
}

// Bytes returns bytes of buffer.
func (buffer *Buffer) Bytes() []byte {
	return buffer.bytes
}

// Cap returns the capacity of the bytes of buffer.
func (buffer *Buffer) Cap() int {
	if buffer.bytes == nil {
		return 0
	}
	return cap(buffer.bytes)
}

// Len returns the size of the bytes of buffer.
func (buffer *Buffer) Len() int {
	if buffer.bytes == nil {
		return 0
	}
	return len(buffer.bytes)
}

// unreadLength returns the number of bytes of the unread portion of the buffer.
func (buffer *Buffer) unreadLength() int {
	if buffer.bytes == nil {
		return 0
	}
	return len(buffer.bytes) - buffer.readOffset
}

// Read reads the next len(p) bytes from the buffer or until the buffer is drained.
// The return first value is the number of bytes read.
// If the buffer has no data to return, return 0, io.EOF.
func (buffer *Buffer) Read(p []byte) (int, error) {
	if buffer.unreadLength() == 0 {
		return 0, io.EOF
	}

	nRead := copy(p, buffer.bytes[buffer.readOffset:])
	buffer.readOffset += nRead
	return nRead, nil
}

func (buffer *Buffer) writeableLen() int {
	return buffer.Cap() - buffer.Len()
}

// ReadFrom reads data from r until EOF and appends it to the buffer, growing the buffer as needed.
func (buffer *Buffer) ReadFrom(r io.Reader) (int64, error) {
	var nRead int64
	for {
		if buffer.writeableLen() < minGrowLength {
			buffer.grow(minGrowLength)
		}

		buff := buffer.bytes[len(buffer.bytes):cap(buffer.bytes)]
		n, err := r.Read(buff)
		nRead += int64(n)
		buffer.bytes = buffer.bytes[:len(buffer.bytes)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return nRead, err
		}
	}
}

// Write appends the contents of p to the buffer, growing the buffer as needed.
func (buffer *Buffer) Write(p []byte) (int, error) {
	if p == nil || len(p) == 0 {
		return 0, nil
	}

	if buffer.writeableLen() < len(p) {
		buffer.grow(len(p))
	}

	buff := buffer.bytes[len(buffer.bytes):cap(buffer.bytes)]
	nWrote := copy(buff, p)
	buffer.bytes = buffer.bytes[:len(buffer.bytes)+nWrote]

	return nWrote, nil
}

// WriteTo writes data to w until the buffer is drained or an error occurs.
func (buffer *Buffer) WriteTo(w io.Writer) (int64, error) {
	if buffer.unreadLength() == 0 {
		return 0, nil
	}
	buff := buffer.bytes[buffer.readOffset:buffer.Len()]
	nWrote, err := w.Write(buff)
	buffer.readOffset += nWrote
	return int64(nWrote), err
}

func (buffer *Buffer) grow(n int) {
	if buffer.pool == nil {
		buffer.pool = defaultBytesPool
	}

	bytes := buffer.pool.Get(buffer.Len() - buffer.readOffset + n)

	if buffer.bytes != nil {
		nCopy := copy(bytes, buffer.bytes[buffer.readOffset:])
		bytes = bytes[:nCopy]
		buffer.pool.Put(buffer.bytes)
	} else {
		bytes = bytes[:0]
	}
	buffer.bytes = bytes
	buffer.readOffset = 0
}

// Grow grows the buffer's capacity.
// After Grow(n), at least n bytes can be written to the buffer without another allocation.
func (buffer *Buffer) Grow(n int) {
	if buffer.writeableLen() >= n {
		return
	}

	buffer.grow(n)
}

// Reset release bytes and set pool to defaultallocator.
func (buffer *Buffer) Reset() {
	if buffer.bytes != nil {
		if buffer.pool == nil {
			buffer.pool = defaultBytesPool
		}

		buffer.pool.Put(buffer.bytes)
		buffer.bytes = nil
	}

	buffer.pool = nil
	buffer.readOffset = 0
}
