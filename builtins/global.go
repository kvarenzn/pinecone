package builtins

import (
	"fmt"

	"github.com/kvarenzn/pinecone/base"
	"github.com/kvarenzn/pinecone/types"
)

var GlobalNamespace = base.Namespace{
	Types: map[string]types.TypeOrCtor{
		"bool":     types.NewTocType(types.Bool),
		"int":      types.NewTocType(types.Int),
		"float":    types.NewTocType(types.Float),
		"string":   types.NewTocType(types.String),
		"box":      types.NewTocType(types.Box),
		"color":    types.NewTocType(types.Color),
		"label":    types.NewTocType(types.Label),
		"line":     types.NewTocType(types.Line),
		"linefill": types.NewTocType(types.LineFill),
		"polyline": types.NewTocType(types.PolyLine),
		"table":    types.NewTocType(types.Table),
		"array": types.NewTocCtor(func(args []types.Type) (types.Type, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("'array' type need one type argument, for item type")
			}

			return types.ArrayOf(args[0]), nil
		}),
		"map": types.NewTocCtor(func(args []types.Type) (types.Type, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("'array' type need two type argument, for key type and value type")
			}

			return types.MapOf(args[0], args[1]), nil
		}),
		"matrix": types.NewTocCtor(func(args []types.Type) (types.Type, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("'matrix' type need one type argument, for item type")
			}

			return types.MatrixOf(args[0]), nil
		}),
	},
	SubNamespace: map[string]base.Namespace{
		"chart": Chart,
	},
}
