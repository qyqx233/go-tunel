package pool

import (
	"fmt"
	"sync"
)

type PooledObject interface {
}

type Recent struct {
}

type Pool struct {
	MaxPooled int
	MinPooled int
}

func (p Pool) Get() PooledObject {
	return nil
}

func (p Pool) Put(object PooledObject) {
}

var pool int

func init() {
	pool := sync.Pool{}
	pool.Get()
	fmt.Println(pool.New)
}
