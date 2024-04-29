package types

func TypeWithNameEqual(twn1, twn2 TypeWithName) bool {
	return twn1.Name == twn2.Name && Equal(twn1.Type, twn2.Type)
}

func Equal(type1, type2 Type) bool {
	if type1 == nil || type2 == nil {
		return false
	}

	if !type1.Kind().IsValid() || !type2.Kind().IsValid() {
		return false
	}

	if type1.Kind() != type2.Kind() {
		return false
	}

	if type1.Kind().IsPrimitiveType() {
		return true
	}

	switch type1.Kind() {
	case ArrayKind, MatrixKind:
		return Equal(type1.Unit(), type2.Unit())
	case MapKind:
		return Equal(type1.Key(), type2.Key()) && Equal(type1.Value(), type2.Value())
	case StructKind:
		count := type1.Count()
		if count != type2.Count() {
			return false
		}

		for i := 0; i < count; i ++ {
			if !TypeWithNameEqual(type1.Field(i), type2.Field(i)) {
				return false
			}
		}
		return true
	case TupleKind:
		count := type1.Count()
		if count != type2.Count() {
			return false
		}

		for i := 0; i < count; i ++ {
			if !Equal(type1.Item(i), type2.Item(i)) {
				return false
			}
		}

		return true
	case FunctionKind:
		if !Equal(type1.Out(), type2.Out()) {
			return false
		}

		argc := type1.Count()
		if argc != type2.Count() {
			return false
		}

		for i := 0; i < argc; i ++ {
			if !TypeWithNameEqual(type1.In(i), type2.In(i)) {
				return false
			}
		}

		return true
	}

	return true
}
