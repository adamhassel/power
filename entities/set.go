package entities

type IntSet map[int]struct{}

func NewIntSet(v ...int) IntSet {
	rv := make(IntSet)
	rv.Add(v...)
	return rv
}

func (s IntSet) Add(v ...int) {
	for _, i := range v {
		s[i] = struct{}{}
	}
}

func (s IntSet) Contains(v int) bool {
	_, ok := s[v]
	return ok
}

func (s IntSet) Slice() []int {
	if len(s) == 0 {
		return nil
	}
	rv := make([]int, 0, len(s))
	for k := range s {
		rv = append(rv, k)
	}
	return rv
}
