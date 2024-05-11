package types

import "strings"

type unionType struct {
	BaseType

	members []Type
}

func (ut unionType) Kind() TypeKind {
	return UnionKind
}

func (ut unionType) String() string {
	types := []string{}
	for _, t := range ut.members {
		types = append(types, t.String())
	}
	return strings.Join(types, " | ")
}

func UnionOf(types []Type) Type {
	switch len(types) {
	case 0:
		panic("union type must be consisted of two or more types")
	case 1:
		return types[0]
	}

	result := unionType{
		members: []Type{},
	}

	appendType := func(t Type) {
		for _, tt := range result.members {
			if Equal(tt, t) {
				return
			}
		}

		result.members = append(result.members, t)
	}

	for _, t := range types {
		if t.Kind() == UnionKind {
			for _, m := range t.Members() {
				appendType(m)
			}
		} else {
			appendType(t)
		}
	}

	return result
}

func Union(types ...Type) Type {
	return UnionOf(types)
}

func (bt BaseType) Members() []Type {
	panic("not applicable")
}

func (ut unionType) Members() []Type {
	return ut.members
}
