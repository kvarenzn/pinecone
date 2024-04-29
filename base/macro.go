package base

import (
	"fmt"

	"github.com/kvarenzn/pinecone/ast"
)

type MacroArg struct {
	Name     string
	Optional bool
}

type Macro struct {
	Name  string
	Macro func(asts ...ast.Node) (ast.Node, error)
	Args  []MacroArg
}

func (m Macro) Call(args []ast.Node, kwargs map[string]ast.Node) (ast.Node, error) {
	output := []ast.Node{}
	argc := len(args)
	if argc + len(kwargs) > len(m.Args) {
		return nil, fmt.Errorf("macro '%s' only accepts '%d' arguments", m.Name, len(m.Args))
	}

	kws := map[string]ast.Node{}
	for k, v := range kwargs {
		kws[k] = v
	}

	for i, a := range m.Args {
		if i >= argc {
			v, ok := kws[a.Name]
			if !ok {
				if !a.Optional {
					return nil, fmt.Errorf("'%s' of macro '%s' is not optional", a.Name, m.Name)
				}
				output = append(output, nil)
				continue
			}
			delete(kws, a.Name)
			output = append(output, v)
			continue
		}
		output = append(output, args[i])
	}

	if len(kws) > 1 {
		for k := range kws {
			return nil, fmt.Errorf("macro '%s' do not have param '%s'", m.Name, k)
		}
	}
	return m.Macro(output...)
}
