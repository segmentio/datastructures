package tree

// Tree is a balanced binary tree containing elements of type E.
type Tree[E any] struct{ impl Map[E, struct{}] }

// New constructs a new tree using the comparison function passed as argument
// to order the elements.
func New[E any](cmp func(E, E) int) *Tree[E] {
	t := new(Tree[E])
	t.Init(cmp)
	return t
}

// Init initializes the tree with the given comparison function to order the
// elements.
func (t *Tree[E]) Init(cmp func(E, E) int) {
	t.impl.Init(cmp)
}

// Len returns the number of elements in the tree.
func (t *Tree[E]) Len() int { return t.impl.Len() }

// Range calls f for each element in the tree, in the order defined by the
// comprison function. If f returns false, the iteration is stopped.
func (t *Tree[E]) Range(f func(E) bool) {
	t.impl.Range(func(elem E, _ struct{}) bool { return f(elem) })
}

// Insert inserts a new element in the tree. The method panics if the tree
// had not been initialized by a call to New or Init.
func (t *Tree[E]) Insert(elem E) (replaced bool) {
	_, replaced = t.impl.Insert(elem, struct{}{})
	return replaced
}

// Contains returns true if the given element exists in the tree.
func (t *Tree[E]) Contains(elem E) (found bool) {
	_, found = t.impl.Lookup(elem)
	return found
}

// Search returns the largest element less or equal to the one passed as
// argument.
func (t *Tree[E]) Search(elem E) (match E, found bool) {
	match, _, found = t.impl.Search(elem)
	return match, found
}

// Delete removes an element from the tree.
func (t *Tree[E]) Delete(elem E) (deleted bool) {
	_, deleted = t.impl.Delete(elem)
	return deleted
}
