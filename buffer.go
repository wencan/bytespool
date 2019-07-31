package bytespool

import (
	"io"
	"sync"
)

var (
	// DefaultBufferMinGrowLength is the default value of Buffer.MinGrowLength.
	DefaultBufferMinGrowLength = 64

	// DefaultBufferReadBuffLength is the default value of Buffer.ReadBuffLength.
	DefaultBufferReadBuffLength = 512

	// DefaultBufferReserveLength is the default value of Buffer..
	DefaultBufferReserveLength = 1024

	// BufferPool is the pool of Buffer instance.
	BufferPool Pool = &sync.Pool{
		New: func() interface{} {
			return new(Buffer)
		},
	}
)

// GetBuffer acquire a buffer at default bytes pool.
func GetBuffer() *Buffer {
	return BufferPool.Get().(*Buffer)
}

// PutBuffer reset and release buffer.
func PutBuffer(buffer *Buffer) {
	buffer.Reset()
	BufferPool.Put(buffer)
}

// SizedBytesPool is a interface that represents a pool of sized bytes.
type SizedBytesPool interface {
	Get(length int) []byte
	Put([]byte)
}

// Buffer get bytes from pool and put idle bytes to pool.
type Buffer struct {
	// BytesPool is a pool of bytes of buffer.
	// default is DefaultBytesPool.
	BytesPool SizedBytesPool

	// MinGrowLength is a minimum length of the buffer growth.
	// default is BufferDefaultMinGrowLength.
	MinGrowLength int

	// ReadBuffLength is a buff length of the Buffer.ReadFrom.
	// default value is BufferDefaultReadBuffLength.
	ReadBuffLength int

	// ReserveLength represents the buffer will reserve the bytes
	// if the capacity of the bytes is equal to less than this value.
	// default is DefaultBufferReserveLength.
	ReserveLength int

	bytes []byte

	readOffset int
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
		if buffer.ReadBuffLength == 0 {
			buffer.ReadBuffLength = DefaultBufferReadBuffLength
		}
		if buffer.writeableLen() < buffer.ReadBuffLength {
			buffer.grow(buffer.ReadBuffLength)
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

// WriteString appends the contents of str to the buffer, growing the buffer as needed.
// The return first value is the length of str.
func (buffer *Buffer) WriteString(str string) (int, error) {
	if len(str) == 0 {
		return 0, nil
	}

	if buffer.writeableLen() < len(str) {
		buffer.grow(len(str))
	}

	buff := buffer.bytes[len(buffer.bytes):cap(buffer.bytes)]
	nWrote := copy(buff, str)
	buffer.bytes = buffer.bytes[:len(buffer.bytes)+nWrote]

	return len(str), nil
}

func (buffer *Buffer) grow(n int) {
	if buffer.BytesPool == nil {
		buffer.BytesPool = DefaultBytesPool
	}

	length := buffer.Len() - buffer.readOffset + n
	if buffer.MinGrowLength == 0 {
		buffer.MinGrowLength = DefaultBufferMinGrowLength
	}
	if length < buffer.MinGrowLength {
		length = buffer.MinGrowLength
	}
	bytes := buffer.BytesPool.Get(length)

	if buffer.bytes != nil {
		nCopy := copy(bytes, buffer.bytes[buffer.readOffset:])
		bytes = bytes[:nCopy]
		buffer.BytesPool.Put(buffer.bytes)
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

// Reset release bytes and reset the buffer status.
func (buffer *Buffer) Reset() {
	if buffer.bytes != nil {
		if buffer.ReserveLength == 0 {
			buffer.ReserveLength = DefaultBufferReserveLength
		}
		if cap(buffer.bytes) > buffer.ReserveLength {
			buffer.BytesPool.Put(buffer.bytes)
			buffer.bytes = nil
		} else {
			buffer.bytes = buffer.bytes[:0]
		}
	}

	buffer.BytesPool = nil
	buffer.readOffset = 0

	buffer.MinGrowLength = 0
	buffer.ReadBuffLength = 0
	buffer.ReserveLength = 0
}
