package base

import (
	"fmt"

	"github.com/kvarenzn/pinecone/types"
)

type ValueWithType struct {
	Type  types.Type
	Value any
}

type Namespace struct {
	Callables    map[string]types.Callable
	Macros       map[string]Macro
	Variables    map[string]ValueWithType
	Types        map[string]types.TypeOrCtor
	SubNamespace map[string]Namespace
}

func (m Namespace) FindType(name string) (*types.TypeOrCtor, error) {
	result, ok := m.Types[name]
	if !ok {
		return nil, fmt.Errorf("type '%s' not found", name)
	}
	return &result, nil
}

func (m Namespace) FindVariableType(name string) (types.Type, error) {
	result, ok := m.Variables[name]
	if !ok {
		return nil, fmt.Errorf("variable '%s' not found", name)
	}
	return result.Type, nil
}

func (m Namespace) FindFunction(name string) (types.Callable, error) {
	result, ok := m.Callables[name]
	if !ok {
		return nil, fmt.Errorf("function '%s' not found", name)
	}

	return result, nil
}

func (m Namespace) FindMethod(name string, selfType types.Type) (types.Callable, error) {
	result, ok := m.Callables[name]
	if ok && result.IsMethod() && types.Equal(selfType, result.FirstArgType()) {
		return result, nil
	}

	// find method in sub namespaces
	for _, sm := range m.SubNamespace {
		result, err := sm.FindMethod(name, selfType)
		if err == nil {
			return result, nil
		}
	}

	return nil, fmt.Errorf("method '%s' on type '%s' not found", name, selfType.String())
}

func (m Namespace) FindNamespace(name string) (*Namespace, error) {
	result, ok := m.SubNamespace[name]
	if !ok {
		return nil, fmt.Errorf("namespace '%s' not found", name)
	}

	return &result, nil
}

func NSTypeWrap(m Namespace) NSType {
	return NSType{
		Namespace: m,
	}
}

type NSType struct {
	types.BaseType
	Namespace Namespace
}

func (m NSType) Kind() types.TypeKind {
	return types.NamespaceKind
}

func (m NSType) String() string {
	return "namespace"
}
