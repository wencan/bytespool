package bytespool

import (
	"bytes"
	"crypto/md5"
	"io"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
)

func TestBufferBasic(t *testing.T) {
	buffer := GetBuffer()
	defer PutBuffer(buffer)

	// make random data
	rand.Seed(time.Now().Unix())
	buff := GetBytes(rand.Intn(1024) + 100)
	defer PutBytes(buff)
	nRead, err := rand.Read(buff)
	if err != nil {
		panic(err)
	}
	if nRead == 0 {
		t.Fatal("length of random data == 0")
		return
	}
	buff = buff[:nRead]

	// sum
	sum1 := md5.Sum(buff)

	// write
	nWrote, err := buffer.Write(buff)
	if err != nil {
		t.Fatal(err)
		return
	}
	if nWrote != nRead {
		t.Fatalf("wrote length error, want: %d, have: %d", nRead, nWrote)
		return
	}

	// read
	nRead, err = buffer.Read(buff)
	if err != nil {
		t.Fatal(err)
		return
	}
	if nWrote != nRead {
		t.Fatalf("read length error, want: %d, have: %d", nWrote, nRead)
		return
	}
	buff = buff[:nRead]

	// sum2
	sum2 := md5.Sum(buff)

	if bytes.Compare(sum1[:], sum2[:]) != 0 {
		t.Fatalf("sum missmatch, want: %x, have: %x", sum1, sum2)
	}
}

func TestBufferGrow(t *testing.T) {
	buffer := GetBuffer()
	defer PutBuffer(buffer)

	rand.Seed(time.Now().Unix())
	for n := 0; n < 1024*1024*100; {
		buffer.Grow(n)
		if buffer.Cap()-buffer.Len() < n {
			t.Fatalf("capacity error, want: >= %d, have: %d", n, buffer.Cap()-buffer.Len())
			return
		}

		// random data
		buff := GetBytes(rand.Intn(1024*1024) + 1)
		nRead, err := rand.Read(buff)
		if err != nil {
			PutBytes(buff)
			panic(err)
		}
		if nRead == 0 {
			PutBytes(buff)
			t.Fatal("length of random data == 0")
			return
		}
		buff = buff[:nRead]

		// write random data
		nWrite, err := buffer.Write(buff)
		PutBytes(buff)
		if err != nil {
			t.Fatal(err)
			return
		}
		if nWrite != nRead {
			t.Fatalf("write length error, want: %d, have: %d", nRead, nWrite)
			return
		}

		// read data
		buff = GetBytes(rand.Intn(1024 * 1024))
		_, err = buffer.Read(buff)
		PutBytes(buff)
		if err != nil {
			t.Fatal(err)
			return
		}

		n += rand.Intn(1024*1024*10) + 1
	}
}

func TestBufferManyWrite(t *testing.T) {
	buffer := GetBuffer()
	defer PutBuffer(buffer)

	var length int
	rand.Seed(time.Now().Unix())
	hash := md5.New()
	for i := 0; i < 1024; i++ {
		buff := GetBytes(rand.Intn(1024) + 1)

		nRead, err := rand.Read(buff)
		if err != nil {
			PutBytes(buff)
			t.Fatal(err)
			return
		}
		length += nRead
		buff = buff[:nRead]

		_, err = buffer.Write(buff)
		if err != nil {
			PutBytes(buff)
			t.Fatal(err)
			return
		}
		_, err = hash.Write(buff)
		PutBytes(buff)
		if err != nil {
			panic(err)
		}
	}

	if length != buffer.Len() {
		t.Fatalf("length error, want: %d, have: %d", length, buffer.Len())
		return
	}

	// sum
	sum1 := hash.Sum(nil)
	sum2 := md5.Sum(buffer.Bytes())
	if bytes.Compare(sum1, sum2[:]) != 0 {
		t.Fatalf("sum missmatch, want: %x, have: %x", sum1, sum2)
	}
}

func TestBufferManyRead(t *testing.T) {
	rand.Seed(time.Now().Unix())

	// make random data
	buff := GetBytes(1024 * 1024)
	nRead, err := rand.Read(buff)
	if err != nil {
		PutBytes(buff)
		panic(err)
	}
	buff = buff[:nRead]

	// sum1
	sum1 := md5.Sum(buff)

	buffer := GetBuffer()
	defer PutBuffer(buffer)

	// write all data
	_, err = buffer.Write(buff)
	PutBytes(buff)
	if err != nil {
		t.Fatal(err)
		return
	}

	// read unit eof
	var length int
	hash := md5.New()
	for {
		buff := GetBytes(rand.Intn(1024) + 1024)
		nRead, err := buffer.Read(buff)
		if err == io.EOF {
			break
		}
		if err != nil {
			PutBytes(buff)
			t.Fatal(err)
			return
		}
		length += nRead
		buff = buff[:nRead]
		_, err = hash.Write(buff)
		PutBytes(buff)
		if err != nil {
			panic(err)
		}
	}

	if length != buffer.Len() {
		t.Fatalf("length missmatch, want: %d, have: %d", buffer.Len(), length)
		return
	}

	// sum
	sum2 := hash.Sum(nil)
	if bytes.Compare(sum1[:], sum2[:]) != 0 {
		t.Fatalf("sum missmatch, want: %x, have: %x", sum1, sum2)
	}
}

func TestBufferReadFrom(t *testing.T) {
	rand.Seed(time.Now().Unix())
	buff := GetBytes(1024*1024 + rand.Intn(1024))
	defer PutBytes(buff)

	// make random data
	rand.Seed(time.Now().Unix())
	_, err := rand.Read(buff)
	if err != nil {
		panic(err)
	}

	reader := bytes.NewReader(buff)

	buffer := GetBuffer()
	defer PutBuffer(buffer)
	nRead, err := buffer.ReadFrom(reader)
	if err != nil {
		t.Fatal(err)
		return
	}
	if int(nRead) != len(buff) {
		t.Fatalf("read length missmatch, want: %d, have: %d", len(buff), nRead)
		return
	}
	if int(nRead) != buffer.Len() {
		t.Fatalf("buffer length missmatch, want: %d, have: %d", nRead, buffer.Len())
		return
	}
	if bytes.Compare(buff, buffer.Bytes()) != 0 {
		t.Fatal("buff data missmatch")
		return
	}
}

func TestBufferWriteTo(t *testing.T) {
	rand.Seed(time.Now().Unix())
	buff := GetBytes(1024*1024 + rand.Intn(1024))
	defer PutBytes(buff)

	// make random data
	rand.Seed(time.Now().Unix())
	_, err := rand.Read(buff)
	if err != nil {
		panic(err)
	}
	buffer := GetBuffer()
	defer PutBuffer(buffer)
	_, err = buffer.Write(buff)
	if err != nil {
		t.Fatal(err)
		return
	}

	writer := bytes.NewBuffer(nil)

	// write data
	nWrote, err := buffer.WriteTo(writer)
	if err != nil {
		t.Fatal(err)
		return
	}
	if int(nWrote) != len(buff) {
		t.Fatalf("write length missmatch, want: %d, have: %d", len(buff), nWrote)
		return
	}
	if writer.Len() != len(buff) {
		t.Fatalf("writer length missmatch, want: %d, have: %d", len(buff), writer.Len())
		return
	}
	if bytes.Compare(buff, writer.Bytes()) != 0 {
		t.Fatal("buff data missmatch")
		return
	}
}

func TestBufferWriteString(t *testing.T) {
	buffer := GetBuffer()
	defer PutBuffer(buffer)

	str := time.Now().String()
	nWrote, err := buffer.WriteString(str)
	if err != nil {
		t.Fatal(err)
		return
	}
	if nWrote != len(str) {
		t.Fatalf("wrote length error, want: %d, have: %d", len(str), nWrote)
		return
	}

	haveStr := string(buffer.Bytes())
	if haveStr != str {
		t.Fatalf("wrote str error, want: %s, have: %s", str, haveStr)
		return
	}
}

type sampleBytesPool struct {
	bytess [][]byte

	calloc func(size int) []byte
}

func (p *sampleBytesPool) Get(size int) []byte {
	for idx, bytes := range p.bytess {
		if cap(bytes) == size {
			p.bytess = append(p.bytess[:idx], p.bytess[idx+1:]...)
			return bytes
		}
	}
	if p.calloc != nil {
		return p.calloc(size)
	}
	return make([]byte, 0, size)
}

func (p *sampleBytesPool) Put(bytes []byte) {
	p.bytess = append(p.bytess, bytes)
}

func TestBufferReuseBytes(t *testing.T) {
	var counter uint64
	bytesPool := &sampleBytesPool{
		calloc: func(size int) []byte {
			atomic.AddUint64(&counter, 1)
			return make([]byte, 0, size)
		},
	}

	rand.Seed(time.Now().Unix())
	length := rand.Intn(littleCapacityUpper) + 1

	buffer1 := GetBuffer()
	buffer1.BytesPool = bytesPool
	buffer1.Grow(length)
	PutBuffer(buffer1)

	buffer2 := GetBuffer()
	buffer2.BytesPool = bytesPool
	buffer2.Grow(length)
	PutBuffer(buffer2)

	if counter != 1 {
		t.Fatalf("counter error, want: %d, have: %d", 1, counter)
		return
	}
}

func BenchmarkBytesBuffer(b *testing.B) {
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

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buffer := GetBuffer()
			buffer.MinGrowLength = 1024

			for _, s := range str {
				buffer.WriteString(s)
			}

			PutBuffer(buffer)
		}
	})
}
