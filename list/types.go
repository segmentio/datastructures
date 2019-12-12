package list

import (
	"fmt"
	"reflect"
	"unsafe"
)

var (
	nodeType = reflect.TypeOf(Node{})
)

type iface struct {
	typ unsafe.Pointer
	ptr unsafe.Pointer
}

type _type struct {
	vtype  reflect.Type
	ptype  reflect.Type
	offset uintptr
}

func (t *_type) known() bool {
	return t.vtype != nil
}

func (t *_type) equal(tt *_type) bool {
	return t.vtype == tt.vtype
}

func (t *_type) match(elem interface{}) bool {
	return t.ptype == reflect.TypeOf(elem)
}

func (t *_type) nodeOf(elem interface{}) *Node {
	ptr := ((*iface)(unsafe.Pointer(&elem))).ptr
	return (*Node)(unsafe.Pointer(uintptr(ptr) + t.offset))
}

func (t *_type) valueOf(node *Node) interface{} {
	ptr := unsafe.Pointer(uintptr(unsafe.Pointer(node)) - t.offset)
	return reflect.NewAt(t.vtype, ptr).Interface()
}

func typeOf(rt reflect.Type) _type {
	t, ok := makeType(rt.Elem())
	if !ok {
		panic(fmt.Errorf("%s: type contains no exported list.Node field and therefore cannot be used as element in an intrusive list", rt))
	}
	t.vtype = rt.Elem()
	t.ptype = rt
	return t
}

func makeType(rt reflect.Type) (_type, bool) {
	n := rt.NumField()

	for i := 0; i < n; i++ {
		f := rt.Field(i)

		if f.PkgPath != "" && f.Name != "_" { // unexported
			continue
		}

		if f.Type == nodeType {
			return _type{offset: f.Offset}, true
		}

		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			if t, ok := makeType(f.Type); ok {
				t.offset += f.Offset
				return t, true
			}
		}
	}

	return _type{}, false
}
