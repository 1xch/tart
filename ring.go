package tart

func ring(from []string) *rn {
	head := &rn{from[0], nil, false, 0}
	var last *rn = head
	for _, v := range from[1:] {
		n := &rn{v, nil, false, 0}
		last.nxt = n
		last = n
	}
	last.nxt = head
	return head
}

type rn struct {
	v       string
	nxt     *rn
	visited bool
	currIdx int
}

func (r *rn) iter(fn func(*rn) bool) {
	if k := fn(r); k {
		return
	}
	r.nxt.iter(fn)
}

func (r *rn) jump(in, out string) int {
	var count bool
	var start bool
	var j int
	var ct int = 0
	r.iter(func(rr *rn) bool {
		if count {
			ct = ct + 1
			rr.currIdx = ct
			rr.visited = true
		}
		if rr.v == in {
			count = true
			start = true
		}
		if rr.v == out && start && rr.visited {
			count = false
			j = rr.currIdx
			return true
		}
		return false
	})
	return j
}
