package builtins

import (
	"github.com/kvarenzn/pinecone/base"
	"github.com/kvarenzn/pinecone/types"
)

var Chart = base.Namespace{
	Types: map[string]types.TypeOrCtor{
		"point": types.NewTocType(types.Point),
	},
}
