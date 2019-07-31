package bytespool

// Pool is the interface of the universal pool.
type Pool interface {
	Get() interface{}
	Put(x interface{})
}
