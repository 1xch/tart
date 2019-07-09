package tart

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type Units map[string]float64

func defaultUnits() Units {
	return Units{
		"ns":      float64(time.Nanosecond),
		"us":      float64(time.Microsecond),
		"µs":      float64(time.Microsecond), // U+00B5 = micro symbol
		"μs":      float64(time.Microsecond), // U+03BC = Greek letter mu
		"ms":      float64(time.Millisecond),
		"s":       float64(time.Second),
		"sec":     float64(time.Second),
		"second":  float64(time.Second),
		"seconds": float64(time.Second),
		"min":     float64(time.Minute),
		"minute":  float64(time.Minute),
		"minutes": float64(time.Minute),
		"h":       float64(time.Hour),
		"hr":      float64(time.Hour),
		"hour":    float64(time.Hour),
		"hours":   float64(time.Hour),
		"d":       float64(time.Hour * 24),
		"day":     float64(time.Hour * 24),
		"days":    float64(time.Hour * 24),
		"w":       float64(time.Hour * 24 * 7),
		"week":    float64(time.Hour * 24 * 7),
		"weeks":   float64(time.Hour * 24 * 7),
		"wk":      float64(time.Hour * 24 * 7),
	}
}

type Replace map[string]string

func (r Replace) Replace(in string) string {
	if v, ok := r[in]; ok {
		return v
	}
	return in
}

func defaultReplace() Replace {
	return Replace{
		"hourly":       "1h",
		"daily":        "1d",
		"weekly":       "1w",
		"biweekly":     "2w",
		"fortnight":    "2w",
		"monthly":      "1m", // monthly calculation is a fuzzy calculation here,
		"bimonthly":    "2m", // use as absolute with care
		"semiannually": "183d",
		"annually":     "1y",
		"biannually":   "2y",
		"quarterly":    "90d",
		"yearly":       "1y",
		"biyearly":     "2y",
	}
}

type rn struct {
	v       string
	nxt     *rn
	visited bool
	currIdx int
}

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

var daysOfWeek = []string{
	"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday",
}

var monthsOfYear = []string{
	"january", "february", "march", "april",
	"may", "june", "july", "august",
	"september", "october", "november", "december",
}

func weekday(t *Tart) string {
	return strings.ToLower(t.Weekday().String())
}

func month(t *Tart) string {
	return strings.ToLower(t.Month().String())
}

var zeroD = time.Duration(0)

func isDuration(s string, when time.Time, u Units, r Replace) (time.Duration, error) {
	if len(s) == 0 {
		return zeroD, nil
	}

	// catch some common but not easily parsed durations
	s = r.Replace(s)

	var isNegative bool
	var exp, whole, fraction int64
	var number, totalYears, totalMonths, totalDays, totalDuration float64
	var dmy, atd time.Duration

	for s != "" {
		// consume possible sign
		if s[0] == '+' {
			if len(s) == 1 {
				return zeroD, fmt.Errorf("cannot parse sign without digits: '+'")
			}
			isNegative = false
			s = s[1:]
		} else if s[0] == '-' {
			if len(s) == 1 {
				return zeroD, fmt.Errorf("cannot parse sign without digits: '-'")
			}
			isNegative = true
			s = s[1:]
		}
		// consume digits
		var done bool
		for !done {
			c := s[0]
			switch {
			case c >= '0' && c <= '9':
				d := int64(c - '0')
				if exp > 0 {
					exp++
					fraction = 10*fraction + d
				} else {
					whole = 10*whole + d
				}
				s = s[1:]
			case c == '.':
				if exp > 0 {
					return zeroD, fmt.Errorf("invalid floating point number format: two decimal points found")
				}
				exp = 1
				fraction = 0
				s = s[1:]
			default:
				done = true
			}
		}
		// adjust number
		number = float64(whole)
		if exp > 0 {
			number += float64(fraction) * math.Pow(10, float64(1-exp))
		}
		if isNegative {
			number *= -1
		}
		// find end of unit
		var i int
		for ; i < len(s) && s[i] != '+' && s[i] != '-' && (s[i] < '0' || s[i] > '9'); i++ {
			// identifier bytes: no-op
		}
		unit := s[:i]

		//fmt.Printf("number: %f; unit: %q\n", number, unit)

		if duration, ok := u[unit]; ok {
			totalDuration += number * duration
		} else {
			switch unit {
			case "m", "mo", "mon", "month", "months":
				totalMonths += number
			case "y", "yr", "year", "years":
				totalYears += number
			default:
				return zeroD, fmt.Errorf("unknown unit in duration: %q", unit)
			}
		}

		s = s[i:]
		whole = 0
	}
	if totalYears != 0 {
		whole := math.Trunc(totalYears)
		fraction := totalYears - whole
		totalYears = whole
		totalMonths += 12 * fraction
	}
	if totalMonths != 0 {
		whole := math.Trunc(totalMonths)
		fraction := totalMonths - whole
		totalMonths = whole
		totalDays += 30 * fraction
	}
	if totalDays != 0 {
		whole := math.Trunc(totalDays)
		fraction := totalDays - whole
		totalDays = whole
		totalDuration += (fraction * 24.0 * float64(time.Hour))
	}

	var dmyNs, tdNs int64
	if totalYears != 0 || totalMonths != 0 || totalDays != 0 {
		f := when.AddDate(int(totalYears), int(totalMonths), int(totalDays))
		dmy = f.Sub(when)
		dmyNs = dmy.Nanoseconds()
	}
	if totalDuration != 0 {
		atd = time.Duration(totalDuration)
		tdNs = atd.Nanoseconds()
	}
	total := dmyNs + tdNs

	return time.Duration(total), nil
}
