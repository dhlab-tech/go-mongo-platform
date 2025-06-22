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
	t := map[string]int{}
	for _, i := range in {
		for _, _i := range i {
			t[_i]++
		}
	}
	for _, v := range in1 {
		if n, ok := t[v]; ok {
			if n == len(in) {
				res = append(res, v)
			}
		}
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

func (s *Intersect) LeftOutter(in1, in2 []string) (res []string) {
	t := make(map[string]struct{}, len(in1))
	for _, d := range in1 {
		t[d] = struct{}{}
	}
	for _, v := range in2 {
		delete(t, v)
	}
	res = make([]string, 0, len(t))
	for d := range t {
		res = append(res, d)
	}
	return
}

func (s *Intersect) LeftOutterSlice(in1 []any, in2 []any, _in1 func(i any) string, _in2 func(i any) string) (res []any) {
	t := make(map[string]any, len(in1))
	for _, d := range in1 {
		t[_in1(d)] = struct{}{}
	}
	for _, v := range in2 {
		delete(t, _in2(v))
	}
	res = make([]any, 0, len(t))
	for d := range t {
		res = append(res, d)
	}
	return
}

// NewIntersect ...
func NewIntersect() *Intersect {
	return &Intersect{}
}
