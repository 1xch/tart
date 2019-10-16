// deeply indebted to:
// https://github.com/karrick/tparse
// https://github.com/araddon/dateparse
// https://github.com/wlbr/feiertage

// TODO
//	- ordinal management (4th, 22nd, 31st, etc)
//  - expressions to flatten api & expand functionality
//  - distance (between 2 dates)
//  - actual implementation and testing or holidays current skeleton
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

// A struct encapsulating functionality related to a specific, embedded time.Time
// instance. A fuzzy acronym for "time and relative time in time".
type Tart struct {
	time.Time
	*Supplement
	*Relation
	last string
}

//
func New(cnf ...Config) *Tart {
	t := &Tart{}
	config := append(defaultConfig, cnf...)
	for _, fn := range config {
		fn(t)
	}
	return t
}

//
type Config func(*Tart)

var defaultConfig = []Config{
	func(t *Tart) { t.Time = time.Now() },
	func(t *Tart) { t.Supplement = defaultSupplement() },
	func(t *Tart) { t.Relation = newRelation(t) },
}

var timeFmt = time.RFC3339

func SetTimeFmt(n string) Config {
	return func(t *Tart) {
		timeFmt = n
	}
}

// Set the time of the instance to the provided time. This forces a reset to
// align the instance to the new time setting all relative funcs to defaults,
// removing cached time funcs, and erasing any set associations.
func (t *Tart) SetTime(tt time.Time) {
	t.Time = tt
	t.reset()
}

func (t *Tart) reset() {
	t.Relation.reset(t)
}

// Return the time of the provided string, relative to the Tart instance time.
func (t *Tart) TimeOf(at string) time.Time {
	fn := t.popTimeFn(at)
	return fn()
}

//
func (t *Tart) Associate(k, v string) (time.Time, error) {
	if err := association(k, v, t.rr, t); err != nil {
		return time.Time{}, err
	}
	return t.TimeOf(k), nil
}

func association(key, value string, r map[string]RelativeFunc, b *Tart) error {
	if epoch, err := strconv.ParseFloat(value, 64); err == nil && epoch >= 0 {
		trunc := math.Trunc(epoch)
		nanos := fractionToNanos(epoch - trunc)
		r[key] = wrapRelativeFunc(time.Unix(int64(trunc), int64(nanos)))
		return nil
	}
	var base RelativeFunc
	var y, m, d int
	var duration time.Duration
	var direction = 1
	var err error

	for k, v := range r {
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
					return fmt.Errorf("expected '+' or '-': %q", dir)
				}
				var nv string
				y, m, d, nv = ymd(value[len(k)+1:])
				if len(nv) > 0 {
					duration, err = time.ParseDuration(nv)
					if err != nil {
						return err
					}
				}
			}
			if direction < 0 {
				y = -y
				m = -m
				d = -d
			}
			tfn := base(b)
			bt := tfn()
			nt := bt.Add(time.Duration(int(duration)*direction)).AddDate(y, m, d)
			r[key] = wrapRelativeFunc(nt)
			return nil
		}
	}
	nt, fErr := time.Parse(timeFmt, value)
	if fErr == nil {
		r[key] = wrapRelativeFunc(nt)
	}
	return fErr
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

// Given a request string duration, gives a duration and error.
// Accepts certain shorthand variations on a duration such as "yearly" or "monthly",
// that convert to durations which of necessity are fuzzy dependent on when they are
// calculated. This can be frustrating, but allows you degrees of freedom to tailor
// calculations to your exact or inexact needs. Use the precision in phrasing you
// require to achieve your goals.
//
// Like `time.ParseDuration`, this accepts multiple fractional scalars, so "now+1.5days-3.21hours"
// is evaluated properly.
//
// The following tokens may be used to specify the respective unit of time:
//
// * Nanosecond: ns
// * Microsecond: us, µs (U+00B5 = micro symbol), μs (U+03BC = Greek letter mu)
// * Millisecond: ms
// * Second: s, sec, second, seconds
// * Minute: min, minute, minutes
// * Hour: h, hr, hour, hours
// * Day: d, day, days
// * Week: w, wk, week, weeks
// * Month: m, mo, mon, month, months
// * Year: y, yr, year, years
func (t *Tart) DurationOf(dur string) time.Duration {
	t.last = dur
	if dur, err := isDuration(dur, t.Time, t.Units, t.Replace); err == nil {
		return dur
	}
	return zeroD
}

var zeroD = time.Duration(0)

func isDuration(s string, when time.Time, u *Units, r *Replace) (time.Duration, error) {
	if len(s) == 0 {
		return zeroD, nil
	}

	// catch some common but not easily parsed durations
	s = r.ReplaceWith(s)

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

		if duration, ok := u.GetUnit(unit); ok {
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

// A convienence function to dateparse.ParseIn(date, t.Location())
//func (t *Tart) Parse(date string) (time.Time, error) {
//	return dateparse.ParseIn(date, t.Location())
//}

// A convienence function to dateparse.ParseAny(date)
//func (t *Tart) ParseAny(date string) (time.Time, error) {
//	return dateparse.ParseAny(date)
//}

// Delegates requested fn(TimeOf, Parse, ParseAny) to Unix time to string.
//func (t *Tart) UnixString(fn, val string) string {
//	var ut time.Time
//	var err error
//	switch fn {
//	case "TimeOf":
//		ut = t.TimeOf(val)
//	case "Parse", "ParseAny":
//		ut, err = t.Parse(val)
//	}
//	if err != nil {
//		return err.Error()
//	}
//	return fmt.Sprintf("%d", ut.Unix())
//}

/*
// Tart to now as a readable age string; not 100% exact, but round and
// apprehendable.
func (t *Tart) ReadableAgeString() string {
	nt := time.Now()
	ra := nt.Sub(t.Time)
	var val float64
	var unit string
	switch {
	case ra < time.Hour:
		val = ra.Minutes()
		switch {
		case val < 2:
			unit = "minute"
		default:
			unit = "minutes"
		}
	case ra < DAY:
		val = ra.Hours()
		switch {
		case val < 2:
			unit = "hour"
		default:
			unit = "hours"
		}
	case ra > DAY && ra < (WEEK*2):
		val = (ra.Hours() / 24)
		switch {
		case val < 2:
			unit = "day"
		default:
			unit = "days"
		}
	case ra > (WEEK*2) && ra < (ROUNDMONTH*2):
		val = (ra.Hours() / (24 * 7))
		switch {
		case val < 2:
			unit = "week"
		default:
			unit = "weeks"
		}
	case ra > (ROUNDMONTH*2) && ra < YEAR:
		val = (ra.Hours() / ((24 * 7) * 30))
		switch {
		case val < 2:
			unit = "month"
		default:
			unit = "months"
		}
	case ra > YEAR:
		val = ra.Hours() / ((24 * 7) * 365)
		switch {
		case val < 2:
			unit = "year"
		default:
			unit = "years"
		}
	}
	return fmt.Sprintf("%.0f %s", val, unit)
}
*/

/*
//
func (t *Tart) Run(x string) interface{} {
	p, cErr := expr.Compile(x, expr.Env(t))
	if cErr != nil {
		return cErr
	}
	out, rErr := expr.Run(p, x)
	if rErr != nil {
		return rErr
	}
	return out
}
*/
