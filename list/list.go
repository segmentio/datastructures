// Package list contains the implementation of a type-safe, intrusive,
// doubly-linked list.
//
// The standard library provides an implementation of a non-intrusive
// doubly-linked list in the container/list package. Non-intrusive means that
// the list tracks values via an intermediary object, which carries a reference
// to the actual values. This double indirection level often impacts usability
// of the code, and requires programs to maintain more (often circular)
// references, which are error prone and make the code harder to read.
// The indirections also increase the number of objects allocated on the heap,
// and the chances of CPU cache misses by requiring more pointer lookups to
// access the data.
//
// The linked list implementation in this package adopts a different approach to
// enable programs to use lists without the hassle of managing an indirection
// layer. Values inserted in the list must be struct types which contain a field
// of type Node, that the list uses to link the values together without requiring
// an extra object.
//
// The List type also implements a type-checking mechanism to guarantee that all
// values inserted in the list are of the same type. Programs that attempt to
// insert values of different types in the list will receive panics.
//
// To use the list, a program must first declare the type of values it will push
// in:
//
//	type Object struct {
//		Data string
//		_    list.Node
//	}
//
// Lists can be constructed by simple declaration since their zero-value
// represents an empty list, then then program can start inserting values.
// Lists detect and retain the type of the first inserted value, then apply
// type checking on all other inserts.
//
//	l := list.List{}
//	l.PushBack(&Object{Data: "A"})
//	l.PushBack(&Object{Data: "B"})
//	l.PushBack(&Object{Data: "C"})
//
//	for x := l.Front(); x != nil; x = l.Next(x) {
//		e := x.(*Object)
//		...
//	}
//
package list

import (
	"fmt"
	"reflect"
)

// Node values must be embedded as a struct field in the values inserted in a
// list.
//
// Typically, an unnamed field would be used to embed the Node value:
//
//	type Person struct {
//		FirstName string
//		LastName  string
//		// Declaring this field allows values of the Person type to be
//		// inserted in linked lists.
//		_ list.Node
//	}
//
// Note that the Node field does not have to be at a specific position in the
// struct, and may also be part of an embedded struct field.
// In this example, the type T can be inserted in a list because its embedded
// value of type S which has a Node field:
//
//	type S struct {
//		Name string
//		node list.Node
//	}
//
//	type T struct {
//		Time time.Time
//		S
//	}
//
// If multiple fields of type Node are declared in the struct, the first one is
// always used and the other ones are ignored.
type Node struct{ prev, next *Node }

// List values are containers of objects which support insertion and removal at
// the front and back of the list, as well as removal of elements at any
// position in O(1).
//
// The values inserted in the list must be passed as pointers to struct values
// of types that contain a Node field.
//
// The zero-value is a valid, empty, and untyped list.
type List struct {
	typ  _type
	head *Node
	tail *Node
	size int
}

// Len returns the number of elements in the list.
func (list *List) Len() int { return list.size }

// Front returns the element at the front of the list, or nil if the list is
// empty.
func (list *List) Front() interface{} {
	if node := list.head; node != nil {
		return list.valueOf(node)
	}
	return nil
}

// Back returns the element at the back of the list, or nil if the list is
// empty.
func (list *List) Back() interface{} {
	if node := list.tail; node != nil {
		return list.valueOf(node)
	}
	return nil
}

// Prev returns the element right before elem in the list, or nil if the list is
// empty.
//
// Prev can be used to iterate backward through the list:
//
//	for elem := list.Back(); elem != nil; elem = list.Prev(elem) {
//		...
//	}
//
// The method panics the type of elem doesn't match the type of other values in
// the list.
func (list *List) Prev(elem interface{}) interface{} {
	if node := list.nodeOf(elem); node != nil && node.prev != nil {
		return list.valueOf(node.prev)
	}
	return nil
}

// Next returns the element right after elem in the list, or nil if the list is
// empty.
//
// Next can be used to iterate forward through the list:
//
//	for elem := list.Front(); elem != nil; elem = list.Next(elem) {
//		...
//	}
//
// The method panics the type of elem doesn't match the type of other values in
// the list.
func (list *List) Next(elem interface{}) interface{} {
	if node := list.nodeOf(elem); node != nil && node.next != nil {
		return list.valueOf(node.next)
	}
	return nil
}

// PushFront inserts elem at the front of the list.
//
// The method panics if elem is already part of a list, or if its type doesn't
// match the type of other values in the list.
func (list *List) PushFront(elem interface{}) {
	list.pushFront(list.nodeOf(elem))
}

// PushFrontList inserts other at the front of the list. The operation runs in
// constant time.
//
// The method panics if the type of values in other differs from the type of
// values in list.
func (list *List) PushFrontList(other *List) {
	if other != list && other.typ.known() {
		if !list.typ.known() {
			list.typ = other.typ
		}
		if !list.typ.equal(&other.typ) {
			panic(fmt.Errorf("cannot add an list with values of type %s to a list with values of type %s", other.typ.ptype, list.typ.ptype))
		}
		list.pushFrontList(other)
	}
}

// PushBack inserts elem at the back of the list.
//
// The method panics if elem is already part of a list, or if its type doesn't
// match the type of other values in the list.
func (list *List) PushBack(elem interface{}) {
	list.pushBack(list.nodeOf(elem))
}

// PushBackList inserts other at the front of the list. The operation runs in
// constant time.
//
// The method panics if the type of values in other differs from the type of
// values in list.
func (list *List) PushBackList(other *List) {
	if other != list && other.typ.known() {
		if !list.typ.known() {
			list.typ = other.typ
		}
		if !list.typ.equal(&other.typ) {
			panic(fmt.Errorf("cannot add an list with values of type %s to a list with values of type %s", other.typ.ptype, list.typ.ptype))
		}
		list.pushBackList(other)
	}
}

// MoveToFront moves elem at the front of the list.
//
// The operation is idempotent, it does nothing if elem is already at the front
// of the list. If elem is not part of the list, it is simply inserted at the
// front.
//
// The method panics the type of elem doesn't match the type of other values in
// the list.
func (list *List) MoveToFront(elem interface{}) {
	list.moveToFront(list.nodeOf(elem))
}

// MoveToBack moves elem at the back of the list.
//
// The operation is idempotent, it does nothing if elem is already at the back
// of the list. If elem is not part of the list, it is simply inserted at the
// back.
//
// The method panics the type of elem doesn't match the type of other values in
// the list.
func (list *List) MoveToBack(elem interface{}) {
	list.moveToBack(list.nodeOf(elem))
}

// RemoveFront removes the element at the front of the list and returns it, or
// returns nil if the list was empty.
//
// This method is a more efficient equivalent to:
//
//	list.Remove(list.Front())
//
func (list *List) RemoveFront() interface{} {
	if node := list.removeFront(); node != nil {
		return list.valueOf(node)
	}
	return nil
}

// RemoveBack removes the element at the back of the list and returns it, or
// returns nil if the list was empty.
//
// This method is a more efficient equivalent to:
//
//	list.Remove(list.Back())
//
func (list *List) RemoveBack() interface{} {
	if node := list.removeBack(); node != nil {
		return list.valueOf(node)
	}
	return nil
}

// Remove removes elem from the list.
//
// If elem is nil, the method does nothing.
//
// The method panics the type of elem doesn't match the type of other values in
// the list.
func (list *List) Remove(elem interface{}) {
	if elem != nil {
		list.remove(list.nodeOf(elem))
	}
}

// Removeall removes all elements from the list. The operation runs in constant
// time.
func (list *List) RemoveAll() {
	list.reset()
}

func (list *List) pushFront(node *Node) {
	if list.head == nil {
		list.tail = node
	} else {
		node.next = list.head
		list.head.prev = node
	}
	list.head = node
	list.size++
}

func (list *List) pushFrontList(other *List) {
	if list.head == nil {
		list.head = other.head
		list.tail = other.tail
		list.size = other.size
	} else {
		other.tail.next = list.head
		list.head.prev = other.tail
		list.head = other.head
		list.size += other.size
	}
	other.reset()
}

func (list *List) pushBack(node *Node) {
	if list.tail == nil {
		list.head = node
	} else {
		node.prev = list.tail
		list.tail.next = node
	}
	list.tail = node
	list.size++
}

func (list *List) pushBackList(other *List) {
	if list.head == nil {
		list.head = other.head
		list.tail = other.tail
		list.size = other.size
	} else {
		other.head.prev = list.tail
		list.tail.next = other.head
		list.tail = other.tail
		list.size += other.size
	}
	other.reset()
}

func (list *List) moveToFront(node *Node) {
	if node != list.head {
		list.remove(node)
		list.pushFront(node)
	}
}

func (list *List) moveToBack(node *Node) {
	if node != list.tail {
		list.remove(node)
		list.pushBack(node)
	}
}

func (list *List) removeFront() *Node {
	node := list.head
	list.remove(node)
	return node
}

func (list *List) removeBack() *Node {
	node := list.tail
	list.remove(node)
	return node
}

func (list *List) remove(node *Node) {
	if node != nil {
		prev := node.prev
		next := node.next

		node.prev = nil
		node.next = nil

		if prev != nil {
			prev.next = next
		}

		if next != nil {
			next.prev = prev
		}

		if node == list.head {
			list.head = next
		}

		if node == list.tail {
			list.tail = prev
		}

		list.size--
	}
}

func (list *List) reset() {
	list.head = nil
	list.tail = nil
	list.size = 0
}

func (list *List) nodeOf(elem interface{}) *Node {
	// Set the list type to guarantee that all nodes in the list will be using
	// the same type.
	if !list.typ.known() {
		list.typ = typeOf(reflect.TypeOf(elem))
	}
	if !list.typ.match(elem) {
		panic(fmt.Errorf("cannot add an element of type %T to a list with values of type %s", elem, list.typ.ptype))
	}
	return list.typ.nodeOf(elem)
}

func (list *List) valueOf(node *Node) interface{} {
	return list.typ.valueOf(node)
}
