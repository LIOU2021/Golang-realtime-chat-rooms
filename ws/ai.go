package ws

import (
	"sync"
)

var Ai autoInc

type autoInc struct {
	sync.Mutex     // ensures autoInc is goroutine-safe
	id         int // current max id number
}

var idPool = &sync.Pool{
	New: func() interface{} {
		return -1
	},
}

func (a *autoInc) ID() (id int) {
	a.Lock()
	defer a.Unlock()

	id = idPool.Get().(int)
	if id == -1 {
		a.id = a.id + 1
		id = a.id
	}

	return
}

func (a *autoInc) PutID(id int) {
	a.Lock()
	defer a.Unlock()
	idPool.Put(id)
}
