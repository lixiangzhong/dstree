package dstree

import (
	"strings"
)

var _ TreeNode[any] = (*node[any])(nil)
var _ Tree[any] = (*node[any])(nil)

type Tree[T any] interface {
	Add(string, T) TreeNode[T]
	Find(string) TreeNode[T]
	Remove(string)
}

type TreeNode[T any] interface {
	Payload() T
	Path() string
}

func NewTree[T any](opts ...Option[T]) Tree[T] {
	tree := &node[T]{
		locker: noneLocker{},
	}
	for _, opt := range opts {
		opt(tree)
	}
	return tree
}

type node[T any] struct {
	locker   Locker
	name     string
	parent   *node[T]
	children map[string]*node[T]
	payload  T
	leaf     bool
}

func (n *node[T]) addChild(name string) *node[T] {
	if n.children == nil {
		n.children = make(map[string]*node[T])
	}
	child := &node[T]{
		name:     name,
		parent:   n,
		children: make(map[string]*node[T]),
	}
	n.children[child.name] = child
	return child
}

func (n *node[T]) Add(hostname string, payload T) TreeNode[T] {
	n.locker.Lock()
	defer n.locker.Unlock()
	ss := strings.Split(hostname, ".")
	this := n
	for index := len(ss) - 1; index >= 0; index-- {
		child, ok := this.children[ss[index]]
		if !ok {
			child = this.addChild(ss[index])
		}
		this = child
	}
	this.payload = payload
	this.leaf = true
	return this
}

func (n *node[T]) Find(hostname string) TreeNode[T] {
	n.locker.RLock()
	defer n.locker.RUnlock()
	ss := strings.Split(hostname, ".")
	var wildcard *node[T]
	this := n
	for index := len(ss) - 1; index >= 0; index-- {
		child := this.find(ss[index])
		if child == nil {
			if wildcard != nil {
				return wildcard
			}
			return this
		}
		wildcard = this.find("*")
		this = child
	}
	return this
}

func (n *node[T]) find(name string) *node[T] {
	child, ok := n.children[name]
	if ok {
		return child
	}
	child, ok = n.children["*"]
	if ok {
		return child
	}
	return nil
}

func (n *node[T]) Remove(hostname string) {
	n.locker.Lock()
	defer n.locker.Unlock()
	ss := strings.Split(hostname, ".")
	this := n
	for index := len(ss) - 1; index >= 0; index-- {
		child, ok := this.children[ss[index]]
		if !ok {
			return
		}
		this = child
	}
	delete(this.parent.children, this.name)
}

func (n *node[T]) Payload() T {
	return n.payload
}

func (n *node[T]) Path() string {
	names := make([]string, 0)
	this := n
	for this != nil && this.name != "" {
		names = append(names, this.name)
		this = this.parent
	}
	return strings.Join(names, ".")
}
