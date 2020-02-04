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

// Tart is a a struct encapsulating functionality related to a specific,
// embedded time.Time instance. An acronym for "time and relative time(in
// time)".
type Tart struct {
	time.Time
	*Supplement
	*Relation
	last string
	dir  *directive
	tFmt string
}

// New builds a new Tart instance from the provided Config.
func New(cnf ...Config) *Tart {
	t := &Tart{}
	config := mkConfig(cnf...)
	for _, fn := range config {
		fn(t)
	}
	return t
}

// Config is a function taking a *Tart instance.
type Config func(*Tart)

func mkConfig(cnf ...Config) []Config {
	def := []Config{
		func(t *Tart) { t.Time = time.Now() },
		func(t *Tart) { t.Supplement = defaultSupplement() },
		func(t *Tart) { t.Relation = newRelation(t) },
		func(t *Tart) { t.tFmt = time.RFC3339 },
	}
	def = append(def, cnf...)
	return def
}

// SetTimeFmt ...
func SetTimeFmt(n string) Config {
	return func(t *Tart) {
		t.tFmt = n
	}
}

// SetTime sets the time of the instance to the provided time. This forces a reset to
// align the instance to the new time setting all relative funcs to defaults,
// removing cached time funcs, and erasing any set associations.
func (t *Tart) SetTime(tt time.Time) {
	t.Time = tt
	t.reset()
}

func (t *Tart) reset() {
	t.Relation.reset(t)
}

// TimeOf returns the time of the provided string, relative to the Tart instance time.
func (t *Tart) TimeOf(at string) time.Time {
	fn := t.popTimeFn(at)
	return fn()
}

// Associate ...
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
		r[key] = wrapRelativeFunc(time.Unix(int64(trunc), nanos))
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
	nt, fErr := time.Parse(b.tFmt, value)
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

// DurationOf returns time.Duration of the provided string.
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
	if dur, err := isDuration(dur, t.Time, t.units, t.replace); err == nil {
		return dur
	}
	return zeroD()
}

func zeroD() time.Duration {
	return time.Duration(0)
}

// TODO: reduce cyclomatic complexity
func isDuration(s string, when time.Time, u *units, r *replace) (time.Duration, error) {
	if len(s) == 0 {
		return zeroD(), nil
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
				return zeroD(), fmt.Errorf("cannot parse sign without digits: '+'")
			}
			isNegative = false
			s = s[1:]
		} else if s[0] == '-' {
			if len(s) == 1 {
				return zeroD(), fmt.Errorf("cannot parse sign without digits: '-'")
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
					return zeroD(), fmt.Errorf("invalid floating point number format: two decimal points found")
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
				return zeroD(), fmt.Errorf("unknown unit in duration: %q", unit)
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
