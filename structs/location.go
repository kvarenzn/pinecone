package structs

type Location struct {
	Column int
	Row    int
}

func (loc Location) IsInvalid() bool {
	return loc.Column == -1 && loc.Row == -1
}

