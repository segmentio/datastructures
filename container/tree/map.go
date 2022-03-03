package tree

/*
	The red-black tree implementation in this file was derived from
	https://github.com/PratikDeoghare/redblack

	A copy of the license is included below:

---
MIT License

Copyright (c) 2022 Pratik Deoghare

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

// Map is a map type associating keys to values in a similar way to the standard
// Go map type, but backed by a balanced binary tree instead of a hashmap, which
// maintains ordering of keys.
//
// The zero-value is a valid empty map which supports lookups and deletes, but
// must be initialized prior to inserting any keys.
type Map[K, V any] struct {
	cmp    func(K, K) int
	len    int
	root   *node[K, V]
	leaf   node[K, V] // This leaf always Black. We don't touch it. Its a sacred leaf.
	bbleaf node[K, V] // This leaf is used for deletion.
}

type color byte

const (
	red    color = 0
	black  color = 1
	bblack color = 2
	nblack color = 3
)

type node[K, V any] struct {
	a     *node[K, V]
	b     *node[K, V]
	key   K
	value V // not the last field so it takes no space when set to struct{}
	color color
}

// NewMap instantiates a new map using the given comparison function to order
// the keys.
func NewMap[K, V any](cmp func(K, K) int) *Map[K, V] {
	m := new(Map[K, V])
	m.Init(cmp)
	return m
}

// Init initializes (or re-initializes) the map. The comparison function passed
// as argument will be used to order the keys.
//
// Init must be called prior to inserting keys in the map, otherwise inserts
// will panic.
//
// Complexity: O(1)
func (m *Map[K, V]) Init(cmp func(K, K) int) {
	m.leaf = node[K, V]{color: black}
	m.leaf.a = &m.leaf
	m.leaf.b = &m.leaf
	m.bbleaf = node[K, V]{color: bblack}
	m.bbleaf.a = &m.leaf
	m.bbleaf.b = &m.leaf
	m.cmp = cmp
	m.len = 0
	m.root = &m.leaf
}

// Len returns the number of entries currently held in the map.
//
// Complexity: O(1)
func (m *Map[K, V]) Len() int { return m.len }

// Range calls f for each entry of the map. The keys and values are presented in
// ascending order according to the comparison function installed on the map.
//
// Complexity: O(N)
func (m *Map[K, V]) Range(f func(K, V) bool) {
	if m.root != nil {
		m.subrange(m.root, f)
	}
}

func (m *Map[K, V]) subrange(n *node[K, V], call func(K, V) bool) bool {
	return n == &m.leaf || (m.subrange(n.a, call) && call(n.key, n.value) && m.subrange(n.b, call))
}

// Insert inserts a new entry in the map, or replaces the value if the key
// already existed. The method returns the previous value associated with the
// key or the zero-value if the key did not exist, and a boolean indicating
// whether the value was replaced.
//
// The map must have been initialized by a call to NewMap or Init or the call
// to Insert will panic.
//
// Complexity: O(log n)
func (m *Map[K, V]) Insert(key K, value V) (previous V, replaced bool) {
	inserted, previous, replaced := m.insert(m.root, key, value)
	m.root = blacken(inserted)
	if !replaced {
		m.len++
	}
	return previous, replaced
}

func (m *Map[K, V]) insert(n *node[K, V], key K, value V) (inserted *node[K, V], previous V, replaced bool) {
	if n == &m.leaf {
		inserted = &node[K, V]{
			a:     &m.leaf,
			b:     &m.leaf,
			key:   key,
			value: value,
			color: red,
		}
	} else {
		switch cmp := m.cmp(key, n.key); {
		case cmp < 0:
			n.a, previous, replaced = m.insert(n.a, key, value)
			inserted = balance(n)
		case cmp > 0:
			n.b, previous, replaced = m.insert(n.b, key, value)
			inserted = balance(n)
		default:
			inserted, previous, replaced = n, n.value, true
			n.value = value
		}
	}
	return inserted, previous, replaced
}

// Min returns the entry with the smallest key in the map.
//
// Complexity: O(log n)
func (m *Map[K, V]) Min() (key K, value V, found bool) {
	if m.root != nil {
		n := min(m.root, &m.leaf)
		key, value, found = n.key, n.value, true
	}
	return key, value, found
}

// Max returns the entry with the largest key in the map.
//
// Complexity: O(log n)
func (m *Map[K, V]) Max() (key K, value V, found bool) {
	if m.root != nil {
		n := max(m.root, &m.leaf)
		key, value, found = n.key, n.value, true
	}
	return key, value, found
}

// Lookup returns the value associated with the given key in the map, and a
// boolean value indicating whether the key was found in the map.
//
// Complexity: O(log n)
func (m *Map[K, V]) Lookup(key K) (value V, found bool) {
	if n := m.root; n != nil {
		for n != &m.leaf {
			switch cmp := m.cmp(key, n.key); {
			case cmp < 0:
				n = n.a
			case cmp > 0:
				n = n.b
			default:
				return n.value, true
			}
		}
	}
	return value, false
}

// Search returns the entry found in the map where the key was less or equal to
// the one passed as argument.
//
// Complexity: O(log n)
func (m *Map[K, V]) Search(key K) (matchKey K, matchValue V, found bool) {
	if n := m.root; n != nil {
		r := (*node[K, V])(nil)

		for n != &m.leaf {
			switch cmp := m.cmp(key, n.key); {
			case cmp < 0:
				n = n.a
			case cmp > 0:
				r = n
				n = n.b
			default:
				return n.key, n.value, true
			}
		}

		if r != nil {
			return r.key, r.value, true
		}
	}
	return matchKey, matchValue, false
}

// Delete deletes the given key from the map. If the key does not exist,
// the map is not modified. The method returns the value removed from the map
// and a boolean indicating whether the key was found.
//
// Complexity: O(log n)
func (m *Map[K, V]) Delete(key K) (value V, deleted bool) {
	if m.root != nil {
		var n *node[K, V]
		n, value, deleted = m.delete(m.root, key)
		if deleted {
			m.root = blacken(n)
		}
	}
	return value, deleted
}

func (m *Map[K, V]) delete(n *node[K, V], key K) (node *node[K, V], value V, deleted bool) {
	if n == &m.leaf {
		return &m.leaf, value, false
	}
	switch cmp := m.cmp(key, n.key); {
	case cmp < 0:
		n.a, value, deleted = m.delete(n.a, key)
		node = m.bubble(n)
	case cmp > 0:
		n.b, value, deleted = m.delete(n.b, key)
		node = m.bubble(n)
	default:
		value, deleted = n.value, true
		node = m.remove(n)
		m.len--
	}
	return node, value, deleted
}

func (m *Map[K, V]) remove(n *node[K, V]) *node[K, V] {
	if n == &m.leaf {
		return &m.leaf
	}
	if n.color == red && n.a == &m.leaf && n.b == &m.leaf {
		return &m.leaf
	}
	if n.color == black && n.a == &m.leaf && n.b == &m.leaf {
		return &m.bbleaf
	}
	if n.color == black && n.a == &m.leaf && n.b != &m.leaf && n.b.color == red {
		n.b.color = black
		return n.b
	}
	if n.color == black && n.b == &m.leaf && n.a != &m.leaf && n.a.color == red {
		n.a.color = black
		return n.a
	}
	// chasing same pointers twice. can optimize by
	// making max return a *node and passing that in to removeMax.
	max := max(n.a, &m.leaf)
	n.key, n.value = max.key, max.value
	n.a = m.removeMax(n.a)
	n = m.bubble(n)
	return n
}

func (m *Map[K, V]) removeMax(n *node[K, V]) *node[K, V] {
	if n.b == &m.leaf {
		return m.remove(n)
	}
	n.b = m.removeMax(n.b)
	return m.bubble(n)
}

func (m *Map[K, V]) bubble(n *node[K, V]) *node[K, V] {
	if n.a.color == bblack || n.b.color == bblack {
		n.color = blacker(n.color)
		n.a = m.redder(n.a)
		n.b = m.redder(n.b)
		return balance(n)
	}
	return balance(n)
}

func (m *Map[K, V]) redder(n *node[K, V]) *node[K, V] {
	if n == &m.bbleaf {
		return &m.leaf
	}
	n.color = redder(n.color)
	return n
}

func blacken[K, V any](n *node[K, V]) *node[K, V] {
	n.color = black
	return n
}

func redden[K, V any](n *node[K, V]) *node[K, V] {
	n.color = red
	return n
}

func colors[K, V any](n1, n2, n3 *node[K, V], c1, c2, c3 color) bool {
	return n1.color == c1 && n2.color == c2 && n3.color == c3
}

func min[K, V any](node, leaf *node[K, V]) *node[K, V] {
	for node.a != leaf {
		node = node.a
	}
	return node
}

func max[K, V any](node, leaf *node[K, V]) *node[K, V] {
	for node.b != leaf {
		node = node.b
	}
	return node
}

func balance[K, V any](n *node[K, V]) *node[K, V] {
	var x, y, z *node[K, V]
	var a, b, c, d *node[K, V]
	okasakiCase := false
	switch {
	case colors(n, n.a, n.a.a, black, red, red):
		x, y, z = n.a.a, n.a, n
		a, b, c, d = x.a, x.b, y.b, z.b
		okasakiCase = true
	case colors(n, n.a, n.a.b, black, red, red):
		x, y, z = n.a, n.a.b, n
		a, b, c, d = x.a, y.a, y.b, z.b
		okasakiCase = true
	case colors(n, n.b, n.b.a, black, red, red):
		x, y, z = n, n.b.a, n.b
		a, b, c, d = x.a, y.a, y.b, z.b
		okasakiCase = true
	case colors(n, n.b, n.b.b, black, red, red):
		x, y, z = n, n.b, n.b.b
		a, b, c, d = x.a, y.a, z.a, z.b
		okasakiCase = true
	}
	if okasakiCase {
		x.a, x.b, z.a, z.b = a, b, c, d
		y.a, y.b = x, z
		x.color, y.color, z.color = black, red, black
		return y
	}
	mightCase := false
	switch {
	case colors(n, n.a, n.a.a, bblack, red, red):
		x, y, z = n.a.a, n.a, n
		a, b, c, d = x.a, x.b, y.b, z.b
		mightCase = true
	case colors(n, n.a, n.a.b, bblack, red, red):
		x, y, z = n.a, n.a.b, n
		a, b, c, d = x.a, y.a, y.b, z.b
		mightCase = true
	case colors(n, n.b, n.b.a, bblack, red, red):
		x, y, z = n, n.b.a, n.b
		a, b, c, d = x.a, y.a, y.b, z.b
		mightCase = true
	case colors(n, n.b, n.b.b, bblack, red, red):
		x, y, z = n, n.b, n.b.b
		a, b, c, d = x.a, y.a, z.a, z.b
		mightCase = true
	default:
		c1, ok := deleteCase1(n)
		if ok {
			return c1
		}
		c2, ok := deleteCase2(n)
		if ok {
			return c2
		}
	}
	if mightCase {
		x.a, x.b, z.a, z.b = a, b, c, d
		y.a, y.b = x, z
		x.color, y.color, z.color = black, black, black
		return y
	}
	return n
}

func deleteCase1[K, V any](n *node[K, V]) (*node[K, V], bool) {
	cond := n.color == bblack && n.b.color == nblack && n.b.a.color == black && n.b.b.color == black
	if !cond {
		return n, false
	}
	x, y, z := n, n.b.a, n.b
	a, b, c, d := x.a, y.a, y.b, z.b
	x.a, x.b = a, b
	z.a, z.b = c, redden(d)
	z.color = black
	y.a, y.b = x, balance(z)
	x.color, y.color, z.color = black, black, black
	return y, true
}

func deleteCase2[K, V any](n *node[K, V]) (*node[K, V], bool) {
	cond := n.color == bblack && n.a.color == nblack && n.a.a.color == black && n.a.b.color == black
	if !cond {
		return n, false
	}
	x, y, z := n.a, n.a.b, n
	a, b, c, d := x.a, y.a, y.b, z.b
	x.a, x.b = redden(a), b
	z.a, z.b = c, d
	x.color = black
	y.a, y.b = balance(x), z
	x.color, y.color, z.color = black, black, black
	return y, true
}

func redder(c color) color {
	switch c {
	case red:
		return nblack
	case black:
		return red
	case bblack:
		return black
	default: // nblack
		panic("cannot be reddened further")
	}
}

func blacker(c color) color {
	switch c {
	case nblack:
		return red
	case red:
		return black
	case black:
		return bblack
	default: // bblack
		panic("cannot be blackened further")
	}
}
