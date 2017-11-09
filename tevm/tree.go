package tevm

import (
	"bytes"
)

// Tree is a tree that stores byte-based key-value pairs
// using a copy-on-write binary tree so that its state can
// be snapshotted in constant time.
//
// The zero value of a Tree is the empty set.
type Tree struct {
	epoch   int32
	root    int32 // index of root node
	all     []treenode
	snaps   []treesnap
	touched bool
}

type treesnap struct {
	root      int32
	allocated int32
}

type treenode struct {
	epoch       int32
	left, right int32 // index+1 of left and right nodes
	key, value  []byte
}

func (t *Tree) new(key, value []byte, left, right int32) int32 {
	t.all = append(t.all, treenode{
		epoch: t.epoch,
		key:   key,
		value: value,
		left:  left,
		right: right,
	})
	return int32(len(t.all))
}

// node numbers are 1-indexed so that the zero
// value of a node contains no children
func (t *Tree) num(i int32) *treenode {
	return &t.all[i-1]
}

func (t *Tree) setLeft(cur, left int32) int32 {
	node := t.num(cur)
	if node.left == left {
		return cur
	}
	if node.epoch < t.epoch {
		return t.new(node.key, node.value, left, node.right)
	}
	node.left = left
	return cur
}

func (t *Tree) setRight(cur, right int32) int32 {
	node := t.num(cur)
	if node.right == right {
		return cur
	}
	if node.epoch < t.epoch {
		return t.new(node.key, node.value, node.left, right)
	}
	node.right = right
	return cur
}

func (t *Tree) setPair(cur int32, key, value []byte) int32 {
	node := t.num(cur)
	if node.epoch < t.epoch {
		return t.new(key, value, node.left, node.right)
	}
	node.key = key
	node.value = value
	return cur
}

func (t *Tree) insertat(i int32, key, value []byte) int32 {
	if i == 0 {
		return t.new(key, value, 0, 0)
	}
	node := t.num(i)
	switch bytes.Compare(node.key, key) {
	case 0:
		return t.setPair(i, key, value)
	case 1:
		return t.setRight(i, t.insertat(node.right, key, value))
	case -1:
		return t.setLeft(i, t.insertat(node.left, key, value))
	}
	panic("unreachable")
}

func (t *Tree) leftmost(n *treenode) *treenode {
	for n.left != 0 {
		n = t.num(n.left)
	}
	return n
}

func (t *Tree) deleteat(i int32, key []byte) int32 {
	if i == 0 {
		return 0
	}
	node := t.num(i)
	switch bytes.Compare(node.key, key) {
	case 0:
		t.touched = true
		if node.left == 0 {
			return node.right
		}
		if node.right == 0 {
			return node.left
		}
		// we could maybe avoid an allocation of an
		// inner node here, but it is such a pain that
		// it's probably not worth it
		min := t.leftmost(t.num(node.right))
		return t.new(min.key, min.value, node.left, t.deleteat(node.right, min.key))
	case 1:
		return t.setRight(i, t.deleteat(node.right, key))
	case -1:
		return t.setLeft(i, t.deleteat(node.left, key))
	}
	panic("unreachable")
}

// Delete removes a value from the tree
func (t *Tree) Delete(key []byte) bool {
	t.root = t.deleteat(t.root, key)
	touched := t.touched
	t.touched = false
	return touched
}

// Insert inserts a key-value pair into the trie
func (t *Tree) Insert(key, value []byte) {
	t.root = t.insertat(t.root, key, value)
}

// Get gets a value from the tree
func (t *Tree) Get(key []byte) []byte {
	i := t.root
	for i != 0 {
		n := t.num(i)
		switch bytes.Compare(n.key, key) {
		case 0:
			return n.value
		case 1:
			i = n.right
		case -1:
			i = n.left
		default:
			panic("unreachable")
		}
	}
	return nil
}

// Snapshot returns a snapshot number for the
// current state of the tree
func (t *Tree) Snapshot() int {
	if t.root == 0 {
		return -1
	}
	s := len(t.snaps)
	t.snaps = append(t.snaps, treesnap{
		root:      t.root,
		allocated: int32(len(t.all)),
	})
	t.epoch++
	return s
}

// Rollback reverts the tree to an old snapshot state.
// Note that the tree can not be rolled forward; rolling
// back to a prior snapshot is irreversible.
func (t *Tree) Rollback(snap int) {
	if snap < 0 {
		*t = Tree{}
		return
	}
	s := t.snaps[snap]
	t.snaps = t.snaps[:snap+1]
	t.root = s.root
	t.all = t.all[:s.allocated]
	t.epoch++
}

func (t *Tree) apply(node *treenode, fn func(k, v []byte) bool) bool {
	if node.left != 0 && !t.apply(t.num(node.left), fn) {
		return false
	}
	if !fn(node.key, node.value) {
		return false
	}
	if node.right != 0 {
		return t.apply(t.num(node.right), fn)
	}
	return true
}

// Iterate iterates the key-value space in sorted order
func (t *Tree) Iterate(fn func(k, v []byte) bool) {
	if t.root != 0 {
		t.apply(t.num(t.root), fn)
	}
}
