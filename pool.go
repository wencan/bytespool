package bytespool

import "sync"

var (
	poolPool = sync.Pool{
		New: func() interface{} {
			return new(Pool)
		},
	}
)

// GetPool acquire a bytes pool
func GetPool() *Pool {
	return poolPool.Get().(*Pool)
}

// PutPool release a bytes pool
func PutPool(pool *Pool) {
	poolPool.Put(pool)
}
