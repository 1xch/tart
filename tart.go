package tart

import (
	"time"

	"github.com/araddon/dateparse"
)

// A struct encapsulating functionality related to a specific, embedded time.Time instance.
// An acronym for "time and relative time in time".
type Tart struct {
	time.Time
	u    Units
	r    Relative
	p    Replace
	last string
}

//
func NewTart(at time.Time, cnf ...Config) *Tart {
	t := &Tart{Time: at}
	config := append(defaultConfig, cnf...)
	for _, fn := range config {
		fn(t)
	}
	return t
}

//
type Config func(*Tart)

var defaultConfig = []Config{
	func(t *Tart) { t.u = defaultUnits() },
	func(t *Tart) { t.r = defaultRelative(t) },
	func(t *Tart) { t.p = defaultReplace() },
}

// Add the provided Units key values to the instance maintained map of units.
func (t *Tart) AddUnits(u ...Units) {
	for _, uu := range u {
		for k, v := range uu {
			t.u[k] = v
		}
	}
}

// Add the provided Relative key values to the instance maintained map of relations.
func (t *Tart) AddRelative(r ...Relative) {
	for _, rr := range r {
		for k, v := range rr {
			t.r[k] = v
		}
	}
}

// Add the provided Replace key values to the instance maintained map of replacements.
func (t *Tart) AddReplace(p ...Replace) {
	for _, pp := range p {
		for k, v := range pp {
			t.p[k] = v
		}
	}
}

// Return the time of the provided string and error, relative to the Tart instance time.
func (t *Tart) TimeOf(at string) (time.Time, error) {
	t.last = at
	if fn, ok := t.r[at]; ok {
		ret := fn(t)
		return ret(), nil
	}
	dfn := t.r["default"]
	ret := dfn(t)
	return ret(), nil
}

// A convienence function to dateparse.ParseIn(date, t.Location())
func (t *Tart) Parse(date string) (time.Time, error) {
	return dateparse.ParseIn(date, t.Location())
}

// A convienence function to dateparse.ParseAny(date)
func (t *Tart) ParseAny(date string) (time.Time, error) {
	return dateparse.ParseAny(date)
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
func (t *Tart) DurationOf(dur string) (time.Duration, error) {
	t.last = dur
	return isDuration(dur, t.Time, t.u, t.p)
}
