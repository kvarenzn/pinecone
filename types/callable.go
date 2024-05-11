package types

import "fmt"

type Callable interface {
	Call(args []any) (any, error)
	Dispatch(args []Type, kwargs map[string]Type) (Type, error)
	IsMethod() bool
	FirstArgType() Type
}

type BuiltinFunction struct {
	Name     string
	Function func(args ...any) (any, error)
	Types    []Type
	OutType  func(args []Type, kwargs map[string]Type) (Type, error)
	SelfType Type
	Method   bool
}

func (bf BuiltinFunction) Call(args []any) (any, error) {
	return bf.Function(args...)
}

func matchArgumentType(argTypes []TypeWithName, args []Type, kwargs map[string]Type) bool {
	argc := len(argTypes)
	if argc < len(args)+len(kwargs) {
		return false
	}

	remainIndex := 0
	for i, a := range args {
		if !Equal(argTypes[i].Type, a) && !CanDoImplicitConversion(a, argTypes[i].Type) {
			return false
		}
		remainIndex++
	}

	remains := map[string]TypeWithName{}

	for i := remainIndex; i < argc; i++ {
		remains[argTypes[i].Name] = argTypes[i]
	}

	for k, v := range kwargs {
		req, ok := remains[k]
		if !ok {
			return false
		}
		if !Equal(req.Type, v) && !CanDoImplicitConversion(v, req.Type) {
			return false
		}

		delete(remains, k)
	}

	for _, v := range remains {
		if v.Optional != true {
			return false
		}
	}

	return true
}

func (bf BuiltinFunction) Dispatch(args []Type, kwargs map[string]Type) (Type, error) {
	if bf.OutType != nil {
		return bf.OutType(args, kwargs)
	}

	if len(bf.Types) == 0 {
		return nil, fmt.Errorf("function %s cannot be called", bf.Name)
	}

	for _, a := range bf.Types {
		allIn := a.AllIn()
		if matchArgumentType(allIn, args, kwargs) {
			return a.Out(), nil
		}
	}

	return nil, fmt.Errorf("mismatch argument type %v, %v", args, kwargs)
}

func (bf BuiltinFunction) IsMethod() bool {
	return bf.Method == true
}

func (bf BuiltinFunction) FirstArgType() Type {
	if !bf.IsMethod() {
		return nil
	}
	if bf.SelfType != nil {
		return bf.SelfType
	}

	var selfType Type = nil
	for _, fnt := range bf.Types {
		if fnt.Count() < 1 {
			return nil
		}

		if selfType == nil {
			selfType = fnt.In(0).Type
			continue
		}

		if !Equal(selfType, fnt.In(0).Type) {
			return nil
		}
	}
	return selfType
}


func CallableTypeWrap(callable Callable) CallableType {
	return CallableType{
		Callable: callable,
	}
}

type CallableType struct {
	BaseType
	Callable Callable
}

func (ct CallableType) Kind() TypeKind {
	return CallableKind
}

func (ct CallableType) String() string {
	return "callable"
}
