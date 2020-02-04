package tart

import (
	"fmt"
	"strings"
	"time"

	"github.com/araddon/dateparse"
)

// Relation is a struct managing core time relations for a Tart instance.
type Relation struct {
	t  *Tart
	rr map[string]RelativeFunc
	cc map[string]TimeFunc
	rk []string
}

type (
	// RelativeFunc ...
	RelativeFunc func(*Tart) TimeFunc
	// TimeFunc ...
	TimeFunc func() time.Time
)

func newRelation(t *Tart) *Relation {
	r := &Relation{}
	r.reset(t)
	return r
}

func (r *Relation) reset(t *Tart) {
	r.t = t
	r.rr, r.rk = defaultRelativeFuncs(r.t)
	r.cc = make(map[string]TimeFunc)
}

func defaultRelativeFuncs(t *Tart) (map[string]RelativeFunc, []string) {
	r := map[string]RelativeFunc{
		"any":       Any,
		"default":   Any,
		"eocm":      EOM,
		"eocw":      EOW,
		"eod":       EOD,
		"eom":       EOM,
		"eoq":       EOQ,
		"eow":       EOW,
		"eoww":      EOWW,
		"eoy":       EOY,
		"later":     Whenever,
		"now":       Now,
		"shift":     Shift,
		"shiftFrom": ShiftFrom,
		"socm":      SOCM,
		"socw":      SOCW,
		"sod":       Tomorrow,
		"som":       SOM,
		"someday":   Whenever,
		"soq":       SOQ,
		"sow":       SOW,
		"soww":      SOWW,
		"soy":       SOY,
		"today":     Today,
		"tomorrow":  Tomorrow,
		"whenever":  Whenever,
		"yesterday": Yesterday,
	}
	for _, d := range daysOfWeek() {
		r[d] = NominalDay(t, d)
	}
	for _, m := range monthsOfYear() {
		r[m] = NominalMonth(t, m)
	}
	var rk []string
	for k, _ := range r {
		rk = append(rk, k)
	}
	return r, rk
}

func daysOfWeek() []string {
	return []string{
		"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday",
	}
}

func monthsOfYear() []string {
	return []string{
		"january", "february", "march", "april",
		"may", "june", "july", "august",
		"september", "october", "november", "december",
	}
}

// Return the TimeFunc of the provided string, relative to the Tart instance time.
func (r *Relation) popTimeFn(at string) TimeFunc {
	t := r.t

	// directive, phrase := parse(at)

	//if tfn, ok := r.cc[at]; ok {
	//	switch {
	//	case strings.Contains(at, "!"):
	//		if spl := strings.Split(at, "!"); len(spl) == 2 {
	//			t.last = spl[1]
	//		}
	//	default:
	//		t.last = at
	//	}
	//	return tfn
	//}

	var rfn RelativeFunc
	switch {
	case strings.Contains(at, "!"):
		if spl := strings.Split(at, "!"); len(spl) == 2 {
			if fn, ok := r.rr[spl[0]]; ok {
				t.last = spl[1]
				rfn = fn
			}
		}
	default:
		t.last = at
		if fn, ok := r.rr[at]; ok {
			rfn = fn
		} else {
			rfn = r.rr["default"]
		}
	}

	tfn := rfn(t)
	r.cc[at] = tfn

	return tfn
}

func parse(in string) (string, string) {
	return "", ""
}

func reservedKeyError(k string) error {
	return fmt.Errorf("'%s' is a relation reserved key", k)
}

func isReservedKey(rk []string, k string) bool {
	for _, v := range rk {
		if k == v {
			return true
		}
	}
	return false
}

// SetRelative ...
func (r *Relation) SetRelative(k string, v RelativeFunc) error {
	if !isReservedKey(r.rk, k) {
		r.rr[k] = v
		return nil
	}
	return reservedKeyError(k)
}

// SetDirect ...
func (r *Relation) SetDirect(k string, v time.Time) error {
	if !isReservedKey(r.rk, k) {
		r.rr[k] = wrapRelativeFunc(v)
		return nil
	}
	return reservedKeyError(k)
}

func wrapRelativeFunc(t time.Time) RelativeFunc {
	return func(*Tart) TimeFunc {
		return func() time.Time {
			return t
		}
	}
}

// SetParsed ...
func (r *Relation) SetParsed(k, v string) error {
	if !isReservedKey(r.rk, k) {
		now := time.Now()
		t, _ := dateparse.ParseIn(v, r.t.Location())
		if y := t.Year(); y <= 0 {
			t = t.AddDate(now.Year(), 0, 0)
		}
		r.rr[k] = func(*Tart) TimeFunc {
			return func() time.Time {
				return t
			}
		}
		return nil
	}
	return reservedKeyError(k)
}

// SetBatch ...
func (r *Relation) SetBatch(in ...map[string]RelativeFunc) error {
	var err error
	for _, v := range in {
		for kk, vv := range v {
			err = r.SetRelative(kk, vv)
			if err != nil {
				return err
			}
		}
	}
	return err
}

// Now returns TimeFunc for "now" from string "now" where "now" is tart.Time
func Now(t *Tart) TimeFunc {
	return func() time.Time {
		return t.Time
	}
}

// Yesterday returns TimeFunc giving local date for yesterday, with time 00:00:00.
func Yesterday(t *Tart) TimeFunc {
	yt := t.Add(-time.Duration(time.Hour * 24))
	return func() time.Time {
		return time.Date(
			yt.Year(),
			yt.Month(),
			yt.Day(),
			0, 0, 0, 0,
			yt.Location(),
		)
	}
}

// Today returns TimeFunc giving current local date, with time 00:00:00.
func Today(t *Tart) TimeFunc {
	tn := t.Time
	return func() time.Time {
		return time.Date(
			tn.Year(),
			tn.Month(),
			tn.Day(),
			0, 0, 0, 0,
			tn.Location(),
		)
	}
}

// EOD returns TimeFunc for "eod" where end of day is current local date, with time 23:59:59.
func EOD(t *Tart) TimeFunc {
	tn := t.Time
	return func() time.Time {
		return time.Date(
			tn.Year(),
			tn.Month(),
			tn.Day(),
			23, 59, 59, 0,
			tn.Location(),
		)
	}
}

// Tomorrow returns TimeFunc for "tomorrow" as local date for tomorrow, with time 00:00:00. Same as sod(start of day).
func Tomorrow(t *Tart) TimeFunc {
	tt := t.Add(time.Hour * 24)
	return func() time.Time {
		return time.Date(
			tt.Year(),
			tt.Month(),
			tt.Day(),
			0, 0, 0, 0,
			tt.Location(),
		)
	}
}

var days *rn = ring(daysOfWeek())

func weekday(t *Tart) string {
	return strings.ToLower(t.Weekday().String())
}

// NominalDay returns RelativeFunc. The subsequent TimeFunc returned generates
// local date for the specified day(monday, tuesday, etc), after today, with
// time 00:00:00.
func NominalDay(t *Tart, d string) RelativeFunc {
	return func(t *Tart) TimeFunc {
		sd := weekday(t)
		jump := days.jump(sd, d)
		return func() time.Time {
			return time.Date(
				t.Year(),
				t.Month(),
				t.Day()+jump,
				0, 0, 0, 0,
				t.Location(),
			)
		}
	}
}

func weekJump(t *Tart, v string, sub int, timeSub ...int) TimeFunc {
	sd := weekday(t)
	jump := days.jump(sd, v) - sub
	var hour, min, secs int
	l := len(timeSub)
	if l > 0 {
		if l >= 1 {
			hour = timeSub[0]
		}
		if l >= 2 {
			min = timeSub[1]
		}
		if l >= 3 {
			secs = timeSub[2]
		}
	}
	return func() time.Time {
		return time.Date(
			t.Year(),
			t.Month(),
			t.Day()+jump,
			hour, min, secs, 0,
			t.Location(),
		)
	}
}

// SOW returns TimeFunc providing local date for the next Sunday, with time
// 00:00:00.
func SOW(t *Tart) TimeFunc {
	return weekJump(t, "sunday", 0)
}

// SOCW returns TimeFunc providing local date for the last Sunday, with time
// 00:00:00.
func SOCW(t *Tart) TimeFunc {
	return weekJump(t, "sunday", 7)
}

// EOW returns TimeFunc for local date for the end of the week, Saturday night,
// with time 00:00:00.
func EOW(t *Tart) TimeFunc {
	return weekJump(t, "saturday", 0)
}

// SOWW returns TimeFunc providing local date for the start of the work week,
// next Monday, with time 00:00:00.
func SOWW(t *Tart) TimeFunc {
	return weekJump(t, "monday", 0)
}

// EOWW returns TimeFunc for local date for the end of the work week, Friday
// night, with time 23:59:59.
func EOWW(t *Tart) TimeFunc {
	return weekJump(t, "friday", 0, 23, 59, 59)
}

//1st, 2nd, ... 	Local date for the next Nth day, with time 00:00:00.
//func OrdinalDay(d string) time.Time {}

var months *rn = ring(monthsOfYear())

func month(t *Tart) string {
	return strings.ToLower(t.Month().String())
}

// NominalMonth returns a RelativeFunc returning a subsequent TimeFunc for
// local date for the specified month(january, february, etc), 1st day, with
// time 00:00:00.
func NominalMonth(t *Tart, m string) RelativeFunc {
	return func(t *Tart) TimeFunc {
		sm := month(t)
		mn := months.jump("january", sm) + 1
		jump := months.jump(sm, m)
		return func() time.Time {
			return time.Date(
				t.Year(),
				time.Month(mn+jump),
				1,
				0, 0, 0, 0,
				t.Location(),
			)
		}
	}
}

// SOCM returns TimeFunc for local date for the 1st day of the current month,
// with time 00:00:00.
func SOCM(t *Tart) TimeFunc {
	return func() time.Time {
		return time.Date(
			t.Year(),
			t.Month(),
			1,
			0, 0, 0, 0,
			t.Location(),
		)
	}
}

// SOM returns TimeFunc providing local date for the 1st day of the next month,
// with time 00:00:00.
func SOM(t *Tart) TimeFunc {
	sm := month(t)
	mn := months.jump("january", sm) + 1
	return func() time.Time {
		return time.Date(
			t.Year(),
			time.Month(mn+1),
			1,
			0, 0, 0, 0,
			t.Location(),
		)
	}
}

// EOM returns TimeFunc providing local date for the last day of the current
// month, with time 23:59:59.
func EOM(t *Tart) TimeFunc {
	sm := month(t)
	mn := months.jump("january", sm) + 1
	return func() time.Time {
		d := time.Date(
			t.Year(),
			time.Month(mn+1),
			1,
			23, 59, 59, 0,
			t.Location(),
		)
		return d.Add(-(24 * time.Hour))
	}
}

func quarters(year int, z *time.Location) map[string]time.Time {
	return map[string]time.Time{
		"Q1":  time.Date(year, time.January, 1, 0, 0, 0, 0, z),
		"Q1x": time.Date(year, time.March, 31, 23, 59, 59, 59, z),
		"Q2":  time.Date(year, time.April, 1, 0, 0, 0, 0, z),
		"Q2x": time.Date(year, time.June, 30, 23, 59, 59, 59, z),
		"Q3":  time.Date(year, time.July, 1, 0, 0, 0, 0, z),
		"Q3x": time.Date(year, time.September, 30, 23, 59, 59, 59, z),
		"Q4":  time.Date(year, time.October, 1, 0, 0, 0, 0, z),
		"Q4x": time.Date(year, time.December, 31, 23, 59, 59, 59, z),
	}
}

// SOQ returns TimeFunc providing local date for the start of the next quarter
// (January, April, July, October), 1st, with time 00:00:00.
func SOQ(t *Tart) TimeFunc {
	q := quarters(t.Year(), t.Location())
	var ret time.Time
	switch {
	case t.After(q["Q1"]) && t.Before(q["Q1x"]):
		ret = q["Q2"]
	case t.After(q["Q2"]) && t.Before(q["Q2x"]):
		ret = q["Q3"]
	case t.After(q["Q3"]) && t.Before(q["Q3x"]):
		ret = q["Q4"]
	case t.After(q["Q4"]) && t.Before(q["Q4x"]):
		ret = q["Q1"]
	}

	return func() time.Time {
		return ret
	}
}

// EOQ returns TimeFunc providing local date for the end of the current quarter
// (March, June, September, December), last day of the month, with time
// 23:59:59.
func EOQ(t *Tart) TimeFunc {
	q := quarters(t.Year(), t.Location())
	var ret time.Time
	switch {
	case t.After(q["Q1"]) && t.Before(q["Q1x"]):
		ret = q["Q1x"]
	case t.After(q["Q2"]) && t.Before(q["Q2x"]):
		ret = q["Q2x"]
	case t.After(q["Q3"]) && t.Before(q["Q3x"]):
		ret = q["Q3x"]
	case t.After(q["Q4"]) && t.Before(q["Q4x"]):
		ret = q["Q4x"]
	}

	return func() time.Time {
		return ret
	}
}

// SOY returns TimeFunc providing local date for the next year, January 1st,
// with time 00:00:00.
func SOY(t *Tart) TimeFunc {
	return func() time.Time {
		return time.Date(
			t.Year()+1,
			time.January,
			1,
			0, 0, 0, 0,
			t.Location(),
		)
	}
}

// EOY returns TimeFunc providing local date for this year, December 31st, with
// time 00:00:00.
func EOY(t *Tart) TimeFunc {
	return func() time.Time {
		return time.Date(
			t.Year(),
			time.December,
			31,
			0, 0, 0, 0,
			t.Location(),
		)
	}
}

// Whenever returns TimeFunc for "whenever", "later", "someday" mapped to local
// 2077-04-27, with time 14:37:00. A date far away.
func Whenever(t *Tart) TimeFunc {
	return func() time.Time {
		return time.Date(
			2077,
			time.Month(4),
			27,
			14, 37, 0, 0,
			t.Location(),
		)
	}
}

// Any returns TimeFunc that attempts to parse Tart.last to a valid time.
func Any(t *Tart) TimeFunc {
	return func() time.Time {
		now := time.Now()
		ret, _ := dateparse.ParseIn(t.last, t.Location())
		if y := ret.Year(); y <= 0 {
			ret = ret.AddDate(now.Year(), 0, 0)
		}
		return ret
	}
}

// Shift returns TimeFunc for an arbitrary time shift from now derived from
// Tart.last, with now being the tart.Time.
func Shift(t *Tart) TimeFunc {
	shift := t.DurationOf(t.last)
	return func() time.Time {
		return t.Add(shift)
	}
}

// ShiftFrom returns TimeFunc for an arbitrary time shift from a specific
// point(that must be return from TimeOf) derived from Tart.last.
func ShiftFrom(t *Tart) TimeFunc {
	var from = time.Time{}
	var shift = time.Duration(0)

	if vars := strings.Split(t.last, ","); len(vars) >= 2 {
		from = t.TimeOf(vars[0])
		shift = t.DurationOf(vars[1])
	}

	return func() time.Time {
		return from.Add(shift)
	}
}

// Next returns TimeFunc for the next iteration of time within span
// e.g. next!July 4,0y = July 4 2019 (where time now is January 1 2019)
//      next!July 4,0y = July 4 2020 (where time now is October 1 2019)
//func Next(t *Tart) TimeFunc {
//	var date time.Time = time.Time{}
//	var dur time.Duration = zeroD()
//	if spl := strings.Split(t.last, ","); len(spl) >= 2 {
//		date = t.TimeOf(spl[0])
//		dur = t.DurationOf(spl[1])
//	}
//
//	return func() time.Time {
//		return date.Add(dur)
//	}
//}

// Last returns TimeFunc for the last iteration of time at interval
// e.g. last!July 4,0y = July 4 2018 (where time now is January 1 2019)
//		last!July 4,0y = July 4 2019 (where time now is October 1 2019)
//func Last(t *Tart) TimeFunc {
//	var date time.Time = time.Time{}
//	var dur time.Duration = zeroD()
//	if spl := strings.Split(t.last, ","); len(spl) >= 2 {
//		date = t.TimeOf(spl[0])
//		dur = t.DurationOf(fmt.Sprintf("-%s", spl[1]))
//	}
//	return func() time.Time {
//		return date.Add(dur)
//	}
//}
