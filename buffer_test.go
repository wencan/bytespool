package bytespool

import (
	"bytes"
	"crypto/md5"
	"io"
	"math/rand"
	"testing"
	"time"
)

func TestBufferBasic(t *testing.T) {
	buffer := GetBuffer()
	defer PutBuffer(buffer)

	// make random data
	rand.Seed(time.Now().Unix())
	buff := Get(rand.Intn(1024) + 100)
	defer Put(buff)
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
		buff := Get(rand.Intn(1024*1024) + 1)
		nRead, err := rand.Read(buff)
		if err != nil {
			Put(buff)
			panic(err)
		}
		if nRead == 0 {
			Put(buff)
			t.Fatal("length of random data == 0")
			return
		}
		buff = buff[:nRead]

		// write random data
		nWrite, err := buffer.Write(buff)
		Put(buff)
		if err != nil {
			t.Fatal(err)
			return
		}
		if nWrite != nRead {
			t.Fatalf("write length error, want: %d, have: %d", nRead, nWrite)
			return
		}

		// read data
		buff = Get(rand.Intn(1024 * 1024))
		_, err = buffer.Read(buff)
		Put(buff)
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
		buff := Get(rand.Intn(1024) + 1)

		nRead, err := rand.Read(buff)
		if err != nil {
			Put(buff)
			t.Fatal(err)
			return
		}
		length += nRead
		buff = buff[:nRead]

		_, err = buffer.Write(buff)
		if err != nil {
			Put(buff)
			t.Fatal(err)
			return
		}
		_, err = hash.Write(buff)
		Put(buff)
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
	buff := Get(1024 * 1024)
	nRead, err := rand.Read(buff)
	if err != nil {
		Put(buff)
		panic(err)
	}
	buff = buff[:nRead]

	// sum1
	sum1 := md5.Sum(buff)

	buffer := GetBuffer()
	defer PutBuffer(buffer)

	// write all data
	_, err = buffer.Write(buff)
	Put(buff)
	if err != nil {
		t.Fatal(err)
		return
	}

	// read unit eof
	var length int
	hash := md5.New()
	for {
		buff := Get(rand.Intn(1024) + 1024)
		nRead, err := buffer.Read(buff)
		if err == io.EOF {
			break
		}
		if err != nil {
			Put(buff)
			t.Fatal(err)
			return
		}
		length += nRead
		buff = buff[:nRead]
		_, err = hash.Write(buff)
		Put(buff)
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
	buff := Get(1024*1024 + rand.Intn(1024))
	defer Put(buff)

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
	buff := Get(1024*1024 + rand.Intn(1024))
	defer Put(buff)

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
