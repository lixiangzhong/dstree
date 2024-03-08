package dstree

import "sync"

type Locker interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

type noneLocker struct{}

func (noneLocker) Lock()    {}
func (noneLocker) Unlock()  {}
func (noneLocker) RLock()   {}
func (noneLocker) RUnlock() {}

type Option[T any] func(*node[T])

func WithLocker[T any]() Option[T] {
	return func(n *node[T]) {
		n.locker = new(sync.RWMutex)
	}
}
