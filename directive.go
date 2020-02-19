package tart

import (
	"bytes"
	"math"
	"strconv"
	"time"
	"unicode"
)

type directive struct {
	origin string
	shift  []*shiftFrag
	phrase string
}

func (d *directive) calcShifts() {
	for _, v := range d.shift {
		v.calcDur()
	}
}

func (d *directive) Shift() []*shifter {
	if len(d.shift) > 0 {
		var ret []*shifter
		for _, v := range d.shift {
			ret = append(ret, v.res...)
		}
		return ret
	}
	return []*shifter{}
}

const (
	tIterPlus   byte = '+'
	tIterMinus  byte = '-'
	tShiftLeft  byte = '<'
	tShiftRight byte = '>'
	tPoint      byte = '!'
)

func isPhrase(b byte) bool {
	var ret bool
	switch {
	case b == tIterPlus, b == tIterMinus, b == tShiftLeft, b == tShiftRight, b == tPoint:
		ret = false
	default:
		ret = true
	}
	return ret
}

func isToken(b byte, t ...byte) bool {
	for _, v := range t {
		if b == v {
			return true
		}
	}
	return false
}

type prs struct {
	currShift  *shiftFrag
	shifts     []*shiftFrag
	inPhrase   bool
	currPhrase *phraseFrag
}

func newPrs() *prs {
	return &prs{
		currShift:  &shiftFrag{0, make([]byte, 0), nil},
		shifts:     make([]*shiftFrag, 0),
		inPhrase:   true,
		currPhrase: &phraseFrag{make([]byte, 0)},
	}
}

type shiftFrag struct {
	count int
	sD    []byte
	res   []*shifter
}

func (s *shiftFrag) append(b byte) {
	if isPhrase(b) {
		s.sD = append(s.sD, b)
	}
}

func durString(s *shiftFrag) string {
	ds := string(s.sD) //
	if ds == "" {
		ds = "0s"
	}
	return ds
}

func (s *shiftFrag) calcDur() {
	res := make([]*shifter, 0)
	cabs := int(math.Abs(float64(s.count)))
	switch {
	case cabs > 0:
		for i := 0; i < cabs; i++ {
			switch {
			case s.count > 0:
				res = append(res, newShifter(durString(s), 1))
			case s.count < 0:
				res = append(res, newShifter(durString(s), -1))
			}
		}
	default:
		res = append(res, newShifter("0s", 1))
	}
	s.res = res
}

type shifter struct {
	origin  string
	y, m, d int
	dur     time.Duration
	err     error
}

func newShifter(in string, dir int) *shifter {
	// alternating numbers and strings
	var y, m, d int
	var accum int     // accumulates digits
	var unit []byte   // accumulates units
	var unproc []byte // accumulate unprocessed durations to return

	unitComplete := func() {
		// NOTE: compare byte slices because some units, i.e. ms, are multi-rune
		if bytes.Equal(unit, []byte{'d'}) ||
			bytes.Equal(unit, []byte{'d', 'a', 'y'}) ||
			bytes.Equal(unit, []byte{'d', 'a', 'y', 's'}) {
			d += accum
		} else if bytes.Equal(unit, []byte{'w'}) ||
			bytes.Equal(unit, []byte{'w', 'e', 'e', 'k'}) ||
			bytes.Equal(unit, []byte{'w', 'e', 'e', 'k', 's'}) {
			d += 7 * accum
		} else if bytes.Equal(unit, []byte{'m', 'o'}) ||
			bytes.Equal(unit, []byte{'m', 'o', 'n'}) ||
			bytes.Equal(unit, []byte{'m', 'o', 'n', 't', 'h'}) ||
			bytes.Equal(unit, []byte{'m', 'o', 'n', 't', 'h', 's'}) ||
			bytes.Equal(unit, []byte{'m', 't', 'h'}) || bytes.Equal(unit, []byte{'m', 'n'}) {
			m += accum
		} else if bytes.Equal(unit, []byte{'y'}) ||
			bytes.Equal(unit, []byte{'y', 'e', 'a', 'r'}) ||
			bytes.Equal(unit, []byte{'y', 'e', 'a', 'r', 's'}) {
			y += accum
		} else {
			unproc = append(append(unproc, strconv.Itoa(accum)...), unit...)
		}
	}

	expectDigit := true
	for _, rune := range in {
		if unicode.IsDigit(rune) {
			if expectDigit {
				accum = accum*10 + int(rune-'0')
			} else {
				unitComplete()
				unit = unit[:0]
				accum = int(rune - '0')
			}
			continue
		}
		unit = append(unit, string(rune)...)
		expectDigit = false
	}
	if len(unit) > 0 {
		unitComplete()
		accum = 0
		unit = unit[:0]
	}

	remaining, err := time.ParseDuration(string(unproc))

	if dir < 0 {
		y = -y
		m = -m
		d = -d
		remaining = -remaining
	}

	return &shifter{in, y, m, d, remaining, err}
}

type phraseFrag struct {
	phrase []byte
}

func (p *phraseFrag) append(b byte) {
	if isPhrase(b) {
		p.phrase = append(p.phrase, b)
	}
}

func (p *phraseFrag) String() string {
	np := string(p.phrase)
	if np == "" {
		np = "now"
	}
	return np
}

func parse(in string) *directive {
	d := &directive{origin: in}
	p := newPrs()
	idx := 0
	for idx <= len(in)-1 {
		var jump int
		for _, idfn := range idFns() {
			add := idfn(idx, in, p)
			jump = jump + add
		}
		switch {
		case jump > 0:
			idx = idx + jump
		default:
			idx++
		}
	}
	for _, fn := range calcFns() {
		fn(d, p)
	}
	return d
}

type idFn func(int, string, *prs) int

func idFns() []idFn {
	return []idFn{
		idShift,
		idPhrase,
	}
}

func idShift(idx int, in string, p *prs) int {
	if isToken(in[idx], tShiftRight, tShiftLeft, tIterPlus, tIterMinus) {
		p.inPhrase = false
		s := idx
		stop := false
		for !stop {
			shiftVal := vShift(in[s])
			switch {
			case shiftVal == 0:
				stopG := false
				for !stopG {
					switch {
					case s > len(in)-1, isToken(in[s], tPoint, tShiftRight, tShiftLeft, tIterPlus, tIterMinus):
						stopG = true
						p.shifts = append(p.shifts, p.currShift)
						p.currShift = &shiftFrag{0, make([]byte, 0), nil}
					default:
						p.currShift.append(in[s])
						s++
					}
				}
				stop = true
			case shiftVal != 0:
				p.currShift.count = p.currShift.count + shiftVal
				s++
			}
		}
		return s - idx
	}
	return 0
}

func vShift(b byte) int {
	var ret int
	switch {
	case b == tShiftRight, b == tIterPlus:
		ret = 1
	case b == tShiftLeft, b == tIterMinus:
		ret = -1
	default:
		ret = 0
	}
	return ret
}

func idPhrase(idx int, in string, p *prs) int {
	b := in[idx]
	if b == tPoint {
		p.inPhrase = true
	}
	if p.inPhrase {
		p.currPhrase.append(b)
	}

	return 0
}

type calcFn func(*directive, *prs)

func calcFns() []calcFn {
	return []calcFn{
		func(d *directive, p *prs) {
			d.phrase = p.currPhrase.String()
		},
		func(d *directive, p *prs) {
			d.shift = p.shifts
			d.calcShifts()
		},
	}
}

type directives struct {
	d    map[string]*directive
	last *directive
}

func newDirectives() *directives {
	d := &directives{}
	d.reset()
	return d
}

func (d *directives) getDirective(k string) *directive {
	if gd, ok := d.d[k]; ok {
		d.last = gd
		return gd
	}
	return nil
}

func (d *directives) setDirective(k string, v *directive) {
	d.d[k] = v
	d.last = v
}

func (d *directives) reset() {
	d.d = make(map[string]*directive)
	d.last = nil
}
