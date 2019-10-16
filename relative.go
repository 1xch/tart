package tart

import (
	"fmt"
	"strings"
	"time"

	"github.com/araddon/dateparse"
)

// Add the provided Relative key values to the instance maintained map of relations.
func (t *Tart) AddRelative(r ...Relative) {
	for _, rr := range r {
		for k, v := range rr {
			t.r[k] = v
		}
	}
}

//
type RelativeFunc func(*Tart) TimeFunc

//
type Relative map[string]RelativeFunc

func defaultRelative(t *Tart) map[string]RelativeFunc {
	r := map[string]RelativeFunc{
		"any":            Any,
		"associate":      Associate,
		"associateWhere": AssociateWhere,
		"default":        Any,
		"eocm":           EOM,
		"eocw":           EOW,
		"eod":            EOD,
		"eom":            EOM,
		"eoq":            EOQ,
		"eow":            EOW,
		"eoww":           EOWW,
		"eoy":            EOY,
		"last":           Last,
		"later":          Whenever,
		"next":           Next,
		"now":            Now,
		"shift":          Shift,
		"shiftFrom":      ShiftFrom,
		"socm":           SOCM,
		"socw":           SOCW,
		"sod":            Tomorrow,
		"som":            SOM,
		"someday":        Whenever,
		"soq":            SOQ,
		"sow":            SOW,
		"soww":           SOWW,
		"soy":            SOY,
		"today":          Today,
		"tomorrow":       Tomorrow,
		"whenever":       Whenever,
		"yesterday":      Yesterday,
	}
	for _, d := range daysOfWeek {
		r[d] = NominalDay(t, d)
	}
	for _, m := range monthsOfYear {
		r[m] = NominalMonth(t, m)
	}
	return r
}

var daysOfWeek = []string{
	"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday",
}

var monthsOfYear = []string{
	"january", "february", "march", "april",
	"may", "june", "july", "august",
	"september", "october", "november", "december",
}

// now from string "now" where "now" is tart.Time
func Now(t *Tart) TimeFunc {
	return func() time.Time {
		return t.Time
	}
}

// Local date for yesterday, with time 00:00:00.
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

// Current local date, with time 00:00:00.
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

// End of day is current local date, with time 23:59:59.
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

// Local date for tomorrow, with time 00:00:00. Same as sod(start of day).
func Tomorrow(t *Tart) TimeFunc {
	tt := t.Add(time.Duration(time.Hour * 24))
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

var days *rn = ring(daysOfWeek)

func weekday(t *Tart) string {
	return strings.ToLower(t.Weekday().String())
}

// Local date for the specified day(monday, tuesday, etc), after today, with time 00:00:00.
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

// Local date for the next Sunday, with time 00:00:00.
func SOW(t *Tart) TimeFunc {
	return weekJump(t, "sunday", 0)
}

// Local date for the last Sunday, with time 00:00:00.
func SOCW(t *Tart) TimeFunc {
	return weekJump(t, "sunday", 7)
}

// Local date for the end of the week, Saturday night, with time 00:00:00.
func EOW(t *Tart) TimeFunc {
	return weekJump(t, "saturday", 0)
}

// Local date for the start of the work week, next Monday, with time 00:00:00.
func SOWW(t *Tart) TimeFunc {
	return weekJump(t, "monday", 0)
}

// Local date for the end of the work week, Friday night, with time 23:59:59.
func EOWW(t *Tart) TimeFunc {
	return weekJump(t, "friday", 0, 23, 59, 59)
}

//1st, 2nd, ... 	Local date for the next Nth day, with time 00:00:00.
//func OrdinalDay(d string) time.Time {}

var months *rn = ring(monthsOfYear)

func month(t *Tart) string {
	return strings.ToLower(t.Month().String())
}

// Local date for the specified month(january, february, etc), 1st day, with time 00:00:00.
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

// Local date for the 1st day of the current month, with time 00:00:00.
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

// Local date for the 1st day of the next month, with time 00:00:00.
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

// Local date for the last day of the current month, with time 23:59:59.
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

// Local date for the start of the next quarter (January, April, July, October),
// 1st, with time 00:00:00.
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

// Local date for the end of the current quarter (March, June, September, December),
// last day of the month, with time 23:59:59.
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

// Local date for the next year, January 1st, with time 00:00:00.
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

// Local date for this year, December 31st, with time 00:00:00.
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

// Whenver, later, someday 	Local 2077-04-27, with time 14:37:00.
// A date far away.
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

//
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

//
func Associate(t *Tart) TimeFunc {
	return func() time.Time {
		return t.Associate("req", t.last)
	}
}

//
func AssociateWhere(t *Tart) TimeFunc {
	return func() time.Time {
		if vars := strings.Split(t.last, "where"); len(vars) >= 2 {
			if splw := strings.Split(vars[1], "="); len(splw) == 2 {
				wh := strings.TrimLeft(splw[0], " ")
				t.AddAssociationString(wh, splw[1])
				return t.Associate("req", strings.TrimRight(vars[0], " "))
			}
		}
		return time.Time{}
	}
}

// an arbitrary time shift of now with now being the tart.Time
func Shift(t *Tart) TimeFunc {
	shift := t.DurationOf(t.last)
	return func() time.Time {
		return t.Add(shift)
	}
}

// an arbitrary time shift from a specific point(must be return from TimeOf)
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

// The next iteration of time at interval
// e.g. next!July 4,1y = July 4 2020 (where time now is July 4 2019)
func Next(t *Tart) TimeFunc {
	var date time.Time = time.Time{}
	var dur time.Duration = zeroD
	if spl := strings.Split(t.last, ","); len(spl) >= 2 {
		date = t.TimeOf(spl[0])
		dur = t.DurationOf(spl[1])
	}
	return func() time.Time {
		return date.Add(dur)
	}
}

// The last iteration of time at interval
// e.g. last!July 4,1y = July 4 2018 (where time now is July 4 2019)
func Last(t *Tart) TimeFunc {
	var date time.Time = time.Time{}
	var dur time.Duration = zeroD
	if spl := strings.Split(t.last, ","); len(spl) >= 2 {
		date = t.TimeOf(spl[0])
		dur = t.DurationOf(fmt.Sprintf("-%s", spl[1]))
	}
	return func() time.Time {
		return date.Add(dur)
	}
}
