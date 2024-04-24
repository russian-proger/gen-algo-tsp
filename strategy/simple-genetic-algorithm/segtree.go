package simplega

type segtree struct {
	data []int
	size int
}

func (tree *segtree) reset(n int) {
	tree.size = 1
	for tree.size < n {
		tree.size <<= 1
	}

	if cap(tree.data) < (tree.size << 1) {
		tree.data = make([]int, tree.size<<1)
	} else {
		tree.data = tree.data[:tree.size<<1]
	}

	for i := tree.size - 1; i < len(tree.data); i++ {
		tree.data[i] = 1
	}

	for i := tree.size - 2; i >= 0; i-- {
		tree.data[i] = tree.data[(i<<1)+1] + tree.data[(i<<1)+2]
	}
}

func (tree *segtree) getk(_n int) int {
	var n int = _n
	var v int = 0
	for v < tree.size-1 {
		v = (v << 1) + 1
		if tree.data[v] <= n {
			n -= tree.data[v]
			v++
		}
	}
	return v - (tree.size - 1)
}

func (tree *segtree) set(index int, value int) {
	var v = tree.size - 1 + index
	tree.data[v] = value

	for v != 0 {
		v = (v - 1) >> 1
		tree.data[v] = tree.data[(v<<1)+1] + tree.data[(v<<1)+2]
	}
}
