package inmemory

type intersect interface {
	Intersect(in1 []string, in ...[]string) (res []string)
	IntersectInt(in1, in2 []int) (res []int)
}

// Intersect ...
type Intersect struct {
}

// Intersect ...
func (s *Intersect) Intersect(in1 []string, in ...[]string) (res []string) {
	t := make(map[string]struct{}, len(in1))
	for _, v := range in1 {
		t[v] = struct{}{}
	}
	tm := 0
	for _, i := range in {
		if tm > 0 {
			t = make(map[string]struct{})
			for _, v := range res {
				t[v] = struct{}{}
			}
			res = make([]string, 0)
		}
		for _, d := range i {
			if _, ok := t[d]; ok {
				res = append(res, d)
			}
		}
		tm++
	}
	return
}

// IntersectInt ...
func (s *Intersect) IntersectInt(in1, in2 []int) (res []int) {
	one := in1
	two := in2
	if len(one) < len(two) {
		one, two = two, one
	}
	t := make(map[int]struct{}, len(one))
	for _, v := range one {
		t[v] = struct{}{}
	}
	res = make([]int, 0, len(two))
	for _, d := range two {
		if _, ok := t[d]; ok {
			res = append(res, d)
		}
	}
	return
}

// Union ...
func (s *Intersect) Union(in1, in2 []string) (res []string) {
	t := make(map[string]struct{}, len(in1))
	for _, d := range in1 {
		t[d] = struct{}{}
	}
	for _, d := range in2 {
		t[d] = struct{}{}
	}
	res = make([]string, 0, len(t))
	for d := range t {
		res = append(res, d)
	}
	return
}

// UnionInt ...
func (s *Intersect) UnionInt(in1, in2 []int) (res []int) {
	t := make(map[int]struct{}, len(in1))
	for _, d := range in1 {
		t[d] = struct{}{}
	}
	for _, d := range in2 {
		t[d] = struct{}{}
	}
	res = make([]int, 0, len(t))
	for d := range t {
		res = append(res, d)
	}
	return
}

// NewIntersect ...
func NewIntersect() *Intersect {
	return &Intersect{}
}
