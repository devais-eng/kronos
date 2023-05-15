package util

type Set map[interface{}]struct{}

func NewSet(values ...interface{}) Set {
	s := make(map[interface{}]struct{})
	for _, v := range values {
		s[v] = struct{}{}
	}
	return s
}

func (s Set) Has(v interface{}) bool {
	_, ok := s[v]
	return ok
}

func (s Set) Add(v interface{}) {
	s[v] = struct{}{}
}

func (s Set) AddMulti(values ...interface{}) {
	for _, v := range values {
		s.Add(v)
	}
}

func (s Set) Remove(v interface{}) {
	delete(s, v)
}

func (s Set) Clear() {
	for k := range s {
		delete(s, k)
	}
}

func (s Set) Size() int {
	return len(s)
}

type SetFilterFunc func(v interface{}) bool

func (s Set) Filter(f SetFilterFunc) Set {
	res := NewSet()
	for v := range s {
		if f(v) {
			res.Add(v)
		}
	}
	return res
}

func (s Set) Union(s2 Set) Set {
	res := NewSet()
	for v := range s {
		res.Add(v)
	}
	for v := range s2 {
		res.Add(v)
	}
	return res
}

func (s Set) Intersect(s2 Set) Set {
	res := NewSet()
	for v := range s {
		if s2.Has(v) {
			res.Add(v)
		}
	}
	return res
}

func (s Set) Difference(s2 Set) Set {
	res := NewSet()
	for v := range s {
		if !s2.Has(v) {
			res.Add(v)
		}
	}
	return res
}

func (s Set) Equal(s2 Set) bool {
	if s.Size() != s2.Size() {
		return false
	}
	for k := range s {
		if !s2.Has(k) {
			return false
		}
	}
	return true
}
