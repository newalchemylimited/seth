package tevm

import (
	"bytes"
	"math/rand"
	"testing"
)

func mustContain(t *testing.T, tree *Tree, key []byte) {
	t.Helper()
	if tree.Get(key) == nil {
		t.Fatalf("key %q not found", key)
	}
}

func mustNotContain(t *testing.T, tree *Tree, key []byte) {
	t.Helper()
	if tree.Get(key) != nil {
		t.Fatalf("shouldn't have found key %q", key)
	}
}

func TestTree(t *testing.T) {
	t.Parallel()
	var tree Tree

	mustNotContain(t, &tree, []byte("foo"))
	mustNotContain(t, &tree, []byte("bar"))

	if tree.Snapshot() != -1 {
		t.Fatal("empty snapshot should be -1")
	}

	const setsize = 500
	set := make([][2][]byte, setsize)
	for i := range set {
		set[i][0] = make([]byte, 32)
		rand.Read(set[i][0])
		set[i][1] = make([]byte, 32)
		rand.Read(set[i][1])
	}
	for i := range set {
		tree.Insert(set[i][0], set[i][1])
		if tree.Get(set[i][0]) == nil {
			t.Fatalf("get after insert %d failed", i)
		}
	}

	for i := range set {
		v := tree.Get(set[i][0])
		if v == nil {
			t.Fatalf("get on entry %d failed", i)
		}
		if !bytes.Equal(v, set[i][1]) {
			t.Fatalf("entry %d value was corrupted", i)
		}
	}

	// inserts should have 'perfect' efficiency; there
	// should be no wasted CoW nodes
	if setsize != len(tree.All) {
		t.Errorf("expected %d total nodes; found %d", setsize, len(tree.All))
	}

	var prev []byte
	iters := 0
	tree.Iterate(func(key, value []byte) bool {
		if prev != nil && bytes.Compare(key, prev) <= 0 {
			t.Fatal("iterate out of order")
		}
		iters++
		return true
	})

	if iters != setsize {
		t.Fatalf("only iterated %d entries", iters)
	}

	snap0 := tree.Snapshot()

	// select a random subset of key/value pairs to remove
	perm := rand.Perm(setsize)[:setsize/2]
	deleted := make([]bool, setsize)
	for _, idx := range perm {
		if !tree.Delete(set[idx][0]) {
			t.Fatalf("couldn't delete entry %d", idx)
		}
		if tree.Get(set[idx][0]) != nil {
			t.Fatalf("delete entry %d didn't work?", idx)
		}
		if tree.Delete(set[idx][0]) {
			t.Fatalf("duplicate delete of entry %d worked?", idx)
		}
		deleted[idx] = true
	}

	for i := range deleted {
		if deleted[i] {
			if tree.Get(set[i][0]) != nil {
				t.Fatalf("found entry %d again?", i)
			}
		} else {
			if tree.Get(set[i][0]) == nil {
				t.Fatalf("entry %d disappeared...?", i)
			}
		}
	}

	t.Logf("after deleting half, %d nodes", len(tree.All))

	// revert to the old state and see that everything is there
	tree.Rollback(snap0)

	iters = 0
	tree.Iterate(func(key, value []byte) bool {
		iters++
		return true
	})
	if iters != setsize {
		t.Fatalf("snap0 only has %d entries", iters)
	}

	for i := range set {
		mustContain(t, &tree, set[i][0])
	}

	// now delete the other half and
	// see that everything works again
	for i := range deleted {
		deleted[i] = !deleted[i]
	}
	for i := range deleted {
		if deleted[i] {
			if !tree.Delete(set[i][0]) {
				t.Fatalf("couldn't delete %d", i)
			}
			if tree.Get(set[i][0]) != nil {
				t.Fatalf("delete of %d failed", i)
			}
			if tree.Delete(set[i][0]) {
				t.Fatalf("second delete of %d succeeded?", i)
			}
		} else {
			if tree.Get(set[i][0]) == nil {
				t.Fatalf("element %d disappeared", i)
			}
		}
	}
	iters = 0
	tree.Iterate(func(key, value []byte) bool {
		iters++
		return true
	})
	if iters != setsize/2 {
		t.Fatalf("expected %d items; found %d", setsize/2, iters)
	}
	t.Logf("after deleting half again, %d nodes", len(tree.All))
}

func BenchmarkTreeInsert500(b *testing.B) {
	var tree Tree
	const setsize = 500
	set := make([][2][]byte, setsize)
	for i := range set {
		set[i][0] = make([]byte, 32)
		rand.Read(set[i][0])
		set[i][1] = make([]byte, 32)
		rand.Read(set[i][1])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Rollback(-1)
		for j := range set {
			tree.Insert(set[j][0], set[j][1])
		}
	}
}

func BenchmarkTreeGet500(b *testing.B) {
	var tree Tree
	const setsize = 500
	set := make([][2][]byte, setsize)
	for i := range set {
		set[i][0] = make([]byte, 32)
		rand.Read(set[i][0])
		set[i][1] = make([]byte, 32)
		rand.Read(set[i][1])
	}

	for j := range set {
		tree.Insert(set[j][0], set[j][1])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := range set {
			tree.Get(set[j][0])
		}
	}
}
