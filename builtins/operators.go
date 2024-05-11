package builtins

import (
	"fmt"

	"github.com/kvarenzn/pinecone/types"
)

type BinaryOperator struct {
	Validate func(types.Type, types.Type) (types.Type, error)
}

type UnaryOperator struct {
	Validate func(types.Type) (types.Type, error)
}

var BinaryOperators = map[string]BinaryOperator{
	"+": {
		Validate: func(left, right types.Type) (types.Type, error) {
			switch left.Kind() {
			case types.IntKind:
				switch right.Kind() {
				case types.IntKind:
					return types.Int, nil
				case types.FloatKind:
					return types.Float, nil
				}
			case types.FloatKind:
				switch right.Kind() {
				case types.IntKind, types.FloatKind:
					return types.Float, nil
				}
			case types.StringKind:
				if right.Kind() == types.StringKind {
					return types.String, nil
				}
			}
			return nil, fmt.Errorf("unsupported '+' operation between '%s' and '%s'", left.String(), right.String())
		},
	},
	"-": {
		Validate: func(left, right types.Type) (types.Type, error) {
			switch left.Kind() {
			case types.IntKind:
				switch right.Kind() {
				case types.IntKind:
					return types.Int, nil
				case types.FloatKind:
					return types.Float, nil
				}
			case types.FloatKind:
				switch right.Kind() {
				case types.IntKind, types.FloatKind:
					return types.Float, nil
				}
			}
			return nil, fmt.Errorf("unsupported '-' operation between '%s' and '%s'", left.String(), right.String())
		},
	},
	"*": {
		Validate: func(left, right types.Type) (types.Type, error) {
			switch left.Kind() {
			case types.IntKind:
				switch right.Kind() {
				case types.IntKind:
					return types.Int, nil
				case types.FloatKind:
					return types.Float, nil
				}
			case types.FloatKind:
				switch right.Kind() {
				case types.IntKind, types.FloatKind:
					return types.Float, nil
				}
			}
			return nil, fmt.Errorf("unsupported '*' operation between '%s' and '%s'", left.String(), right.String())
		},
	},
	"/": {
		Validate: func(left, right types.Type) (types.Type, error) {
			switch left.Kind() {
			case types.IntKind:
				switch right.Kind() {
				case types.IntKind:
					return types.Int, nil
				case types.FloatKind:
					return types.Float, nil
				}
			case types.FloatKind:
				switch right.Kind() {
				case types.IntKind, types.FloatKind:
					return types.Float, nil
				}
			}
			return nil, fmt.Errorf("unsupported '/' operation between '%s' and '%s'", left.String(), right.String())
		},
	},
	"%": {
		Validate: func(left, right types.Type) (types.Type, error) {
			switch left.Kind() {
			case types.IntKind:
				switch right.Kind() {
				case types.IntKind:
					return types.Int, nil
				case types.FloatKind:
					return types.Float, nil
				}
			case types.FloatKind:
				switch right.Kind() {
				case types.IntKind, types.FloatKind:
					return types.Float, nil
				}
			}
			return nil, fmt.Errorf("unsupported '%%' operation between '%s' and '%s'", left.String(), right.String())
		},
	},
}

var UnaryOperators = map[string]UnaryOperator{
	"+": {
		Validate: func(t types.Type) (types.Type, error) {
			switch t.Kind() {
			case types.IntKind:
				return types.Int, nil
			case types.FloatKind:
				return types.Float, nil
			}
			return nil, fmt.Errorf("unsupported '+' operation on '%s'", t.String())
		},
	},
	"-": {
		Validate: func(t types.Type) (types.Type, error) {
			switch t.Kind() {
			case types.IntKind:
				return types.Int, nil
			case types.FloatKind:
				return types.Float, nil
			}
			return nil, fmt.Errorf("unsupported '-' operation on '%s'", t.String())
		},
	},
	"not": {
		Validate: func(t types.Type) (types.Type, error) {
			switch t.Kind() {
			case types.BoolKind:
				return types.Bool, nil
			}
			return nil, fmt.Errorf("unsupported 'not' operation on '%s'", t.String())
		},
	},
}
