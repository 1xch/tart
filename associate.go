package tart

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type association struct {
	t  *Tart
	ss map[string]string
	tt map[string]time.Time
}

func defaultAssociation(t *Tart) *association {
	a := &association{}
	a.reset(t)
	return a
}

func defaultTimeAssociation(t *Tart) map[string]time.Time {
	ret := map[string]time.Time{
		"now": t.Time,
	}
	return ret
}

func (a *association) reset(t *Tart) {
	a.t = t
	a.ss = make(map[string]string)
	a.tt = defaultTimeAssociation(t)
}

//
func (a *association) AddAssociationString(k, v string) {
	a.ss[k] = v
	a.updateAssociationStrings()
}

var timeFmt = time.RFC3339

func (a *association) updateAssociationStrings() {
	var pErr error
	for ik, iv := range a.ss {
		a.tt[ik], pErr = parseWithMap(timeFmt, iv, a.tt)
		if pErr != nil {
			a.tt[ik] = a.t.TimeOf(iv)
		}
	}
}

func parseWithMap(layout, value string, dict map[string]time.Time) (time.Time, error) {
	if epoch, err := strconv.ParseFloat(value, 64); err == nil && epoch >= 0 {
		trunc := math.Trunc(epoch)
		nanos := fractionToNanos(epoch - trunc)
		return time.Unix(int64(trunc), int64(nanos)), nil
	}
	var base time.Time
	var y, m, d int
	var duration time.Duration
	var direction = 1
	var err error

	for k, v := range dict {
		if strings.HasPrefix(value, k) {
			base = v
			if len(value) > len(k) {
				// maybe has +, -
				switch dir := value[len(k)]; dir {
				case '+':
					// no-op
				case '-':
					direction = -1
				default:
					return base, fmt.Errorf("expected '+' or '-': %q", dir)
				}
				var nv string
				y, m, d, nv = ymd(value[len(k)+1:])
				if len(nv) > 0 {
					duration, err = time.ParseDuration(nv)
					if err != nil {
						return base, err
					}
				}
			}
			if direction < 0 {
				y = -y
				m = -m
				d = -d
			}
			return base.Add(time.Duration(int(duration)*direction)).AddDate(y, m, d), nil
		}
	}
	return time.Parse(layout, value)
}

func fractionToNanos(fraction float64) int64 {
	return int64(fraction * float64(time.Second/time.Nanosecond))
}

func ymd(value string) (int, int, int, string) {
	// alternating numbers and strings
	var y, m, d int
	var accum int     // accumulates digits
	var unit []byte   // accumulates units
	var unproc []byte // accumulate unprocessed durations to return

	unitComplete := func() {
		// NOTE: compare byte slices because some units, i.e. ms, are multi-rune
		if bytes.Equal(unit, []byte{'d'}) || bytes.Equal(unit, []byte{'d', 'a', 'y'}) || bytes.Equal(unit, []byte{'d', 'a', 'y', 's'}) {
			d += accum
		} else if bytes.Equal(unit, []byte{'w'}) || bytes.Equal(unit, []byte{'w', 'e', 'e', 'k'}) || bytes.Equal(unit, []byte{'w', 'e', 'e', 'k', 's'}) {
			d += 7 * accum
		} else if bytes.Equal(unit, []byte{'m', 'o'}) || bytes.Equal(unit, []byte{'m', 'o', 'n'}) || bytes.Equal(unit, []byte{'m', 'o', 'n', 't', 'h'}) || bytes.Equal(unit, []byte{'m', 'o', 'n', 't', 'h', 's'}) || bytes.Equal(unit, []byte{'m', 't', 'h'}) || bytes.Equal(unit, []byte{'m', 'n'}) {
			m += accum
		} else if bytes.Equal(unit, []byte{'y'}) || bytes.Equal(unit, []byte{'y', 'e', 'a', 'r'}) || bytes.Equal(unit, []byte{'y', 'e', 'a', 'r', 's'}) {
			y += accum
		} else {
			unproc = append(append(unproc, strconv.Itoa(accum)...), unit...)
		}
	}

	expectDigit := true
	for _, rune := range value {
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
	// log.Printf("y: %d; m: %d; d: %d; nv: %q", y, m, d, unproc)
	return y, m, d, string(unproc)
}

//
func (a *association) GetAssociationTime(k string) time.Time {
	if t, ok := a.tt[k]; ok {
		return t
	}
	return time.Time{}
}

//
func (a *association) SetAssociationTime(k string, t time.Time) {
	a.tt[k] = t
}

//
func (a *association) Associate(k, v string) time.Time {
	a.AddAssociationString(k, v)
	var ret = time.Time{}
	if rt, ok := a.tt[k]; ok {
		ret = rt
	}
	delete(a.ss, k)
	delete(a.tt, k)
	return ret
}

// Associate Tart instance around the provided string key to the provided
// map[string]string as a map relaying string key to time.Time value and error.
//func (t *Tart) Associate(key string, in map[string]string) (map[string]time.Time, error) {
//	var ret = make(map[string]time.Time)
//	var to time.Time
//	kv, ok := in[key]
//	switch {
//	case ok:
//		to = t.TimeOf(kv)
//		ret["now"] = t.Time
//	default:
//		to = t.Time
//	}
//	ret[key] = to
//	delete(in, key)
//	var pErr error
//	for ik, iv := range in {
//		ret[ik], pErr = parseWithMap(time.RFC3339, iv, ret)
//		if pErr != nil {
//			ret[ik] = t.TimeOf(iv)
//		}
//	}
//	return ret, nil
//}

// Associate, returning map[string]string (time values formatted to provided format value)
//func (t *Tart) AssociateFmtString(key, fkey string, in map[string]string) (map[string]string, error) {
//	ret := make(map[string]string)
//	assc, err := t.Associate(key, in)
//	if err != nil {
//		return nil, err
//	}
//	for k, v := range assc {
//		ret[k] = v.Format(fkey)
//	}
//	return ret, nil
//}

// Associate, returning map[string]int64 (unix time)
//func (t *Tart) AssociateInt64(key string, in map[string]string) (map[string]int64, error) {
//	ret := make(map[string]int64)
//	assc, err := t.Associate(key, in)
//	if err != nil {
//		return nil, err
//	}
//	for k, v := range assc {
//		ret[k] = v.UnixNano()
//	}
//	return ret, nil
//}

// Associate, returning map[string]string derived from int64(unix time) values
//func (t *Tart) AssociateInt64String(key string, in map[string]string) (map[string]string, error) {
//	ret := make(map[string]string)
//	assc, err := t.AssociateInt64(key, in)
//	if err != nil {
//		return nil, err
//	}
//	for k, v := range assc {
//		ret[k] = strconv.FormatInt(v, 10)
//	}
//	return ret, nil
//}

//var associateKeyErr = xrr.Xrror("key '%s' not present").Out
