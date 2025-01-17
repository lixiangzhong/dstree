package dstree

import (
	"fmt"
	"sort"
	"strings"
)

var _ TreeNode[any] = (*node[any])(nil)
var _ Tree[any] = (*node[any])(nil)

type Tree[T any] interface {
	Add(string, T) TreeNode[T]
	Find(string) TreeNode[T]
	Remove(string)
	Dump()
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

func (n *node[T]) Add(hostname string, payload T) TreeNode[T] {
	n.locker.Lock()
	defer n.locker.Unlock()

	hostname = strings.TrimSpace(strings.TrimSuffix(hostname, "."))
	if hostname == "" {
		return n
	}

	this := n
	for _, part := range reverse(strings.FieldsFunc(hostname, isDot)) {
		if child, ok := this.children[part]; ok {
			this = child
			continue
		}
		this = this.addChild(part)
	}
	this.payload = payload
	this.leaf = true
	return this
}

func (n *node[T]) Find(hostname string) TreeNode[T] {
	n.locker.RLock()
	defer n.locker.RUnlock()

	this := n
	for _, part := range reverse(strings.Split(hostname, ".")) {
		if child := this.findChild(part); child != nil {
			this = child
			continue
		}
		break
	}
	return this
}

func (n *node[T]) Remove(hostname string) {
	if hostname = strings.TrimSpace(hostname); hostname == "" {
		return
	}

	n.locker.Lock()
	defer n.locker.Unlock()

	this := n
	for _, part := range reverse(strings.Split(hostname, ".")) {
		child, ok := this.children[part]
		if !ok {
			return
		}
		this = child
	}

	if len(this.children) > 0 {
		this.leaf = false
		var zero T
		this.payload = zero
		return
	}

	for current, parent := this, this.parent; parent != nil; {
		delete(parent.children, current.name)
		current.cleanup()
		if len(parent.children) > 0 || parent.leaf {
			break
		}
		current, parent = parent, parent.parent
	}
}

func (n *node[T]) Payload() T {
	n.locker.RLock()
	defer n.locker.RUnlock()
	return n.payload
}

func (n *node[T]) Path() string {
	n.locker.RLock()
	defer n.locker.RUnlock()
	return strings.Join(n.pathParts(), ".")
}

func (n *node[T]) Dump() {
	n.locker.RLock()
	defer n.locker.RUnlock()
	fmt.Println(".")
	n.dumpTree("")
}

// 内部辅助方法
func (n *node[T]) addChild(name string) *node[T] {
	if n.children == nil {
		n.children = make(map[string]*node[T])
	}
	child := &node[T]{
		name:     name,
		parent:   n,
		children: make(map[string]*node[T]),
		locker:   n.locker,
	}
	n.children[name] = child
	return child
}

func (n *node[T]) findChild(name string) *node[T] {
	if child, ok := n.children[name]; ok {
		return child
	}
	return n.children["*"]
}

func (n *node[T]) cleanup() {
	n.parent = nil
	n.children = nil
}

func (n *node[T]) pathParts() []string {
	var parts []string
	for this := n; this != nil && this.name != ""; this = this.parent {
		parts = append(parts, this.name)
	}
	return parts
}

func (n *node[T]) dumpTree(prefix string) {
	children := n.sortedChildren()
	for i, child := range children {
		isLast := i == len(children)-1
		connector := "├──"
		if isLast {
			connector = "└──"
		}

		if child.leaf {
			fmt.Printf("%s%s %s = %v\n", prefix, connector, child.name, child.payload)
		} else {
			fmt.Printf("%s%s %s\n", prefix, connector, child.name)
		}

		newPrefix := prefix
		if isLast {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}
		child.dumpTree(newPrefix)
	}
}

func (n *node[T]) sortedChildren() []*node[T] {
	children := make([]*node[T], 0, len(n.children))
	for _, child := range n.children {
		children = append(children, child)
	}
	sort.Slice(children, func(i, j int) bool {
		return children[i].name < children[j].name
	})
	return children
}

// 工具函数
func isDot(r rune) bool {
	return r == '.'
}

func reverse(ss []string) []string {
	for i, j := 0, len(ss)-1; i < j; i, j = i+1, j-1 {
		ss[i], ss[j] = ss[j], ss[i]
	}
	return ss
}
