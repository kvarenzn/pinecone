package types

func CanDoImplicitConversion(from, to Type) bool {
	if to.Kind() == UnionKind {
		return false
	}
	if from.Kind() == UnionKind {
		for _, m := range from.Members() {
			if !CanDoImplicitConversion(m, to) {
				return false
			}
		}
		return true
	}
	if from.Kind() == UncertainKind && to.Kind() != UncertainKind && to.Kind() != VoidKind {
		// na can be casted to almost any types
		return true
	}
	if to.Kind() == BoolKind {
		if from.Kind() == IntKind || from.Kind() == FloatKind {
			return true
		}
	} else if to.Kind() == FloatKind && from.Kind() == IntKind {
		return true
	}

	return false
}

