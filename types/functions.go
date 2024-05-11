package types

func (bt BaseType) In(i int) TypeWithName {
	panic("not applicable")
}

func (bt BaseType) InName(name string) *TypeWithName {
	panic("not applicable")
}

func (bt BaseType) AllIn() []TypeWithName {
	panic("not applicable")
}

func (bt BaseType) Out() Type {
	panic("not applicable")
}

func (f functionType) Count() int {
	return len(f.in)
}

func (f functionType) In(i int) TypeWithName {
	return f.in[i]
}

func (f functionType) InName(name string) *TypeWithName {
	for _, a := range f.in {
		if a.Name == name {
			return &a
		}
	}
	return nil
}

func (f functionType) AllIn() []TypeWithName {
	result := []TypeWithName{}
	for _, a := range f.in {
		result = append(result, a)
	}
	return result
}

func (f functionType) Out() Type {
	return f.out
}

func (twq TypeWithQualifier) In(i int) TypeWithName {
	return twq.Type.In(i)
}

func (twq TypeWithQualifier) InName(name string) *TypeWithName {
	return twq.Type.InName(name)
}

func (twq TypeWithQualifier) AllIn() []TypeWithName {
	return twq.Type.AllIn()
}

func (twq TypeWithQualifier) Out() Type {
	return twq.Type.Out()
}
