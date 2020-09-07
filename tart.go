package tart

import (
	"time"
)

// Tart is a a struct encapsulating functionality related to a specific,
// embedded time.Time instance. An acronym for "time and relative time(in
// time)".
type Tart struct {
	time.Time
	*relations
	*directives
	tFmt string
}

// New builds a new Tart instance from the provided Config.
func New(cnf ...Config) (*Tart, error) {
	t := &Tart{}
	config := mkConfig(cnf...)
	for _, fn := range config {
		if err := fn(t); err != nil {
			return nil, err
		}
	}
	return t, nil
}

// Config is a function taking a *Tart instance.
type Config func(*Tart) error

func mkConfig(cnf ...Config) []Config {
	def := []Config{
		func(t *Tart) error { t.Time = time.Now(); return nil },
		func(t *Tart) error { t.relations = newRelations(t); return nil },
		func(t *Tart) error { t.directives = newDirectives(); return nil },
		func(t *Tart) error { t.tFmt = time.RFC3339; return nil },
	}
	def = append(def, cnf...)
	return def
}

// SetTimeFmt ...
func SetTimeFmt(n string) Config {
	return func(t *Tart) error {
		t.tFmt = n
		return nil
	}
}

// Establish sets the time of the instance to the provided time. This forces a reset to
// align the instance to the new time setting all relative funcs to defaults,
// removing cached time funcs, and erasing any set associations.
func (t *Tart) Establish(tt time.Time) {
	t.Time = tt
	t.reset()
}

func (t *Tart) reset() {
	t.relations.reset(t)
	t.directives.reset()
}

// Set ...
func (t *Tart) Set(k, v string) error {
	if !isReservedKey(t.relations.rk, k) {
		d := parse(v)
		t.setDirective(v, d)
		nt := pop(t)
		err := t.SetRelation(k, wrapRelative(nt))
		return err
	}
	return reservedKeyError(k)
}

// Get attempts to return the time of the provided directive from the Tart instance.
//
// A directive is a string of the form `modifiers ! point in time` where:
//
// Modifiers are one or more signs followed by duration information:
//	'>' = shift forward
//  '<' = shift backward
//  '+' = iter next
//  '-' = iter last
//
//  'iter' and 'shift' are functionally equivalent currently, use depending on
//   your preference and need
//
// Modifiers stack. Modifiers are collected by type. Duration is applied left wise to
// freestanding modifiers taking duration information.
//	e.g.
//		">>>>>>1h"        = shift forward 6 hours
//      "<1d<2d<<<<3d"    = shift backward 15 days
//      "+++++1h"         = iter forward 5 hours
//      "------1h"        = iter back 6 hours
//      "--3h>3h"         = iter back 6 hours, shifted ahead 3 hours
//
// Point in time is an exclamation point optionally followed by a string. Point may be a
// defined keyword relation or a date construction of some form. When not
// followed by a string, '!' means the tart instance fixed time. When not
// modified, the dot may be omitted, i.e absence of a concrete dot is indicative
// of a statement of point with no modification. A null string is equivalent to
// single point("" == "!")
//
//	e.g.'
//	   "!july 4 1776"     = time of july 4, 1776
//     "!tuesday"         = next tuesday
//     "!eoq"             = end of quarter
//     "!later"           = later
//
// Construction of a directive is dependent on the output you desire. Common use
// is shifting time forward or backward, iteration from a specific point, and
// retrieving a specific point in time from a general specification. Order of
// application is to establish a point in time, iterate, then
// shift.
//
// Examples:
//      `>>>1h.`                       = 3 hours forward from the tart instance time
//      `<<<1h.`                       = 3 hours backward from the tart instance time
//		`>>1h!tuesday`                 = 2 hours forward from next tuesday relative to the tart instance time
//      `<<1h!tuesday`                 = 2 hours backward from next tuesday relative to the tart instance time
//      `->>1h!tuesday`                = 2 hours forward from last tuesday relative to the tart instance time
//      `++<<1h!tuesday`               = 2nd tuesday from now shifted back 2 hours
//      `>>>>>1y!`                     = 5 years from now (where now is tart instance time)
//      `+++++!`                       = 5 years from now (where now is tart instance time)
//      `>>>>>1y!tuesday`              = 5 years from next tuesday
//      `<<<<<1y!oct 31 2025`          = now, if today is oct 31 2020
//		`>.tuesday`, `!tuesday`        = next tuesday, relative to the tart instance time
//      `<!tuesday`                    = last tuesday, relative to the tart instance time
//      `>>>1w!tuesday`, `>>>!tuesday` = 3rd tuesday from tart instance time
//      `<<<!july 4 2006`              = july 4th 2003
//      `+!july 4`, `!july 4`          = the next july 4th
//      `+!christmas`, `!christmas`    = the next christmas (where christmas is defined on the tart instance)
//      `!october 31 1927`             = october 31 1927
//      `!october 31`                  = the next instance of october 31
//      `tomorrow`,`!tomorrow`         = time tomorrow, relative to today
//
// Unique directives are stored by key and reused within the scope of use.
func (t *Tart) Get(in string) time.Time {
	var d *directive
	d = t.getDirective(in)
	if d != nil {
		return pop(t)
	}
	d = parse(in)
	t.setDirective(in, d)
	return pop(t)
}

func pop(t *Tart) time.Time {
	fn := t.popTimeFn()
	return fn()
}

// Duration returns the duration of the modifier of a parsed directive in
// relation to the tart time instance.
// e.g. ">7d>7d>7d+1h!" ==
//      ">7d>7d>7d+1h" ==
//      ">7d>7d>7d+1h!<any relation>" ==
//      "505h" ==
//      (time.Duration) 505h0m0s
func (t *Tart) Duration(in string) time.Duration {
	var d *directive
	d = t.getDirective(in)
	if d != nil {
		return pumpDur(t.Time, d)
	}
	d = parse(in)
	t.setDirective(in, d)
	return pumpDur(t.Time, d)
}

func pumpDur(t time.Time, d *directive) time.Duration {
	var nt time.Time = t
	nt = pumpShift(nt, d)
	nts := nt.Sub(t)
	if nts < 0 {
		nts = -nts
	}
	return nts
}
