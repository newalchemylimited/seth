package tevm

import (
	"bytes"
)

// Tree is a tree that stores byte-based key-value pairs
// using a copy-on-write binary tree so that its state can
// be snapshotted in constant time.
//
// The zero value of a Tree is the empty set.
type Tree tree

// This is unexported to work with json.Marshal but not godoc.
type tree struct {
	Epoch   int32
	Root    int32 // index of root node
	All     []treenode
	Snaps   []treesnap
	Touched bool
}

type treesnap struct {
	Root      int32
	Allocated int32
}

type treenode struct {
	Epoch       int32
	Left, Right int32 // index+1 of left and right nodes
	Key, Value  []byte
}

func (t *Tree) new(key, value []byte, left, right int32) int32 {
	t.All = append(t.All, treenode{
		Epoch: t.Epoch,
		Key:   key,
		Value: value,
		Left:  left,
		Right: right,
	})
	return int32(len(t.All))
}

// node numbers are 1-indexed so that the zero
// value of a node contains no children
func (t *Tree) num(i int32) *treenode {
	return &t.All[i-1]
}

func (t *Tree) setLeft(cur, left int32) int32 {
	node := t.num(cur)
	if node.Left == left {
		return cur
	}
	if node.Epoch < t.Epoch {
		return t.new(node.Key, node.Value, left, node.Right)
	}
	node.Left = left
	return cur
}

func (t *Tree) setRight(cur, right int32) int32 {
	node := t.num(cur)
	if node.Right == right {
		return cur
	}
	if node.Epoch < t.Epoch {
		return t.new(node.Key, node.Value, node.Left, right)
	}
	node.Right = right
	return cur
}

func (t *Tree) setPair(cur int32, key, value []byte) int32 {
	node := t.num(cur)
	if node.Epoch < t.Epoch {
		return t.new(key, value, node.Left, node.Right)
	}
	node.Key = key
	node.Value = value
	return cur
}

func (t *Tree) insertat(i int32, key, value []byte) int32 {
	if i == 0 {
		return t.new(key, value, 0, 0)
	}
	node := t.num(i)
	switch bytes.Compare(node.Key, key) {
	case 0:
		return t.setPair(i, key, value)
	case 1:
		return t.setRight(i, t.insertat(node.Right, key, value))
	case -1:
		return t.setLeft(i, t.insertat(node.Left, key, value))
	}
	panic("unreachable")
}

func (t *Tree) leftmost(n *treenode) *treenode {
	for n.Left != 0 {
		n = t.num(n.Left)
	}
	return n
}

func (t *Tree) deleteat(i int32, key []byte) int32 {
	if i == 0 {
		return 0
	}
	node := t.num(i)
	switch bytes.Compare(node.Key, key) {
	case 0:
		t.Touched = true
		if node.Left == 0 {
			return node.Right
		}
		if node.Right == 0 {
			return node.Left
		}
		// we could maybe avoid an allocation of an
		// inner node here, but it is such a pain that
		// it's probably not worth it
		min := t.leftmost(t.num(node.Right))
		return t.new(min.Key, min.Value, node.Left, t.deleteat(node.Right, min.Key))
	case 1:
		return t.setRight(i, t.deleteat(node.Right, key))
	case -1:
		return t.setLeft(i, t.deleteat(node.Left, key))
	}
	panic("unreachable")
}

// Delete removes a value from the tree
func (t *Tree) Delete(key []byte) bool {
	t.Root = t.deleteat(t.Root, key)
	touched := t.Touched
	t.Touched = false
	return touched
}

// Insert inserts a key-value pair into the trie
func (t *Tree) Insert(key, value []byte) {
	t.Root = t.insertat(t.Root, key, value)
}

// Get gets a value from the tree
func (t *Tree) Get(key []byte) []byte {
	i := t.Root
	for i != 0 {
		n := t.num(i)
		switch bytes.Compare(n.Key, key) {
		case 0:
			return n.Value
		case 1:
			i = n.Right
		case -1:
			i = n.Left
		default:
			panic("unreachable")
		}
	}
	return nil
}

// Snapshot returns a snapshot number for the
// current state of the tree
func (t *Tree) Snapshot() int {
	if t.Root == 0 {
		return -1
	}
	s := len(t.Snaps)
	t.Snaps = append(t.Snaps, treesnap{
		Root:      t.Root,
		Allocated: int32(len(t.All)),
	})
	t.Epoch++
	return s
}

// CopyAt returns a logical copy of thre tree
// at a given snapshot. (As an optimization, the
// data itself is not copied.) Updates to the returned
// Tree will not be reflected in t.
//
// A safe copy of the current state of the tree
// can be obtained through code like
//
//     t.CopyAt(t.Snapshot())
//
func (t *Tree) CopyAt(snap int) Tree {
	if snap < 0 {
		return Tree{}
	}
	s := t.Snaps[snap]
	return Tree{
		Snaps: t.Snaps[:snap+1],
		Root:  s.Root,
		// make sure any appends to the node list
		// cause reallocation of the backing data
		All:   t.All[:s.Allocated:s.Allocated],
		Epoch: t.Epoch + 1,
	}
}

// Rollback reverts the tree to an old snapshot state.
// Note that the tree can not be rolled forward; rolling
// back to a prior snapshot is irreversible.
func (t *Tree) Rollback(snap int) {
	if snap < 0 {
		*t = Tree{}
		return
	}
	s := t.Snaps[snap]
	t.Snaps = t.Snaps[:snap+1]
	t.Root = s.Root
	t.All = t.All[:s.Allocated]
	t.Epoch++
}

func (t *Tree) apply(node *treenode, fn func(k, v []byte) bool) bool {
	if node.Left != 0 && !t.apply(t.num(node.Left), fn) {
		return false
	}
	if !fn(node.Key, node.Value) {
		return false
	}
	if node.Right != 0 {
		return t.apply(t.num(node.Right), fn)
	}
	return true
}

// Iterate iterates the key-value space in sorted order
func (t *Tree) Iterate(fn func(k, v []byte) bool) {
	if t.Root != 0 {
		t.apply(t.num(t.Root), fn)
	}
}
