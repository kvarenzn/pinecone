package analyzer

import (
	"reflect"

	"github.com/kvarenzn/pinecone/ast"
)

func MarkParent(n ast.Node, p ast.Node, name string, index int) {
	n.SetParent(p)
	n.SetPathAttribute(name)
	n.SetPathIndex(index)
	markParentHelper(n)
}

func markParentHelper(n ast.Node) {
	t := reflect.TypeOf(n)
	kind := t.Kind()
	if kind != reflect.Struct {
		return
	}
	wrapper := reflect.ValueOf(n)
	fields := t.NumField()
	for i := 0; i < fields; i++ {
		field := t.Field(i)
		name := field.Name
		if name[0] < 'A' || name[0] > 'Z' {
			continue
		}
		fv := wrapper.Field(i)
		switch fv.Kind() {
		case reflect.Struct:
			subNode, ok := fv.Interface().(ast.Node)
			if ok {
				MarkParent(subNode, n, name, -1)
			}
		case reflect.Array:
			if fv.Elem().Kind() != reflect.Struct {
				continue
			}
			size := fv.Len()
			for i := 0; i < size; i++ {
				item := fv.Index(i)
				subNode, ok := item.Interface().(ast.Node)
				if ok {
					MarkParent(subNode, n, name, i)
				}
			}
		}
	}
}
