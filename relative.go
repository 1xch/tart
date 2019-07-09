package tart

import (
	"time"

	"github.com/araddon/dateparse"
)

type RelativeFunc func(*Tart) TimeFunc

type TimeFunc func() time.Time

type Relative map[string]RelativeFunc

func defaultRelative(t *Tart) map[string]RelativeFunc {
	r := map[string]RelativeFunc{
		"yesterday": Yesterday,
		"today":     Today,
		"eod":       EOD,
		"tomorrow":  Tomorrow,
		"sod":       Tomorrow,
		"sow":       SOW,
		"socw":      SOCW,
		"eow":       EOW,
		"eocw":      EOW,
		"soww":      SOWW,
		"eoww":      EOWW,
		"socm":      SOCM,
		"som":       SOM,
		"eom":       EOM,
		"eocm":      EOM,
		"soq":       SOQ,
		"eoq":       EOQ,
		"soy":       SOY,
		"eoy":       EOY,
		"someday":   Whenever,
		"later":     Whenever,
		"whenever":  Whenever,
		"default":   Any,
	}
	for _, d := range daysOfWeek {
		r[d] = NominalDay(t, d)
	}
	for _, m := range monthsOfYear {
		r[m] = NominalMonth(t, m)
	}
	return r
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

func weekJump(t *Tart, v string, sub int) TimeFunc {
	sd := weekday(t)
	jump := days.jump(sd, v) - sub
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

// Local date for the next Sunday, with time 00:00:00.
func SOW(t *Tart) TimeFunc {
	return weekJump(t, "sunday", 0)
	//sd := weekday(t)
	//jump := days.jump(sd, "sunday")
	//return func() time.Time {
	//	return time.Date(
	//		t.Year(),
	//		t.Month(),
	//		t.Day()+jump,
	//		0, 0, 0, 0,
	//		t.Location(),
	//	)
	//}
}

// Local date for the last Sunday, with time 00:00:00.
func SOCW(t *Tart) TimeFunc {
	return weekJump(t, "sunday", 7)
	//sd := weekday(t)
	//jump := days.jump(sd, "sunday") - 7
	//return func() time.Time {
	//	return time.Date(
	//		t.Year(),
	//		t.Month(),
	//		t.Day()+jump,
	//		0, 0, 0, 0,
	//		t.Location(),
	//	)
	//}
}

// Local date for the end of the week, Saturday night, with time 00:00:00.
func EOW(t *Tart) TimeFunc {
	return weekJump(t, "saturday", 0)
	//sd := weekday(t)
	//jump := days.jump(sd, "saturday")
	//return func() time.Time {
	//	return time.Date(
	//		t.Year(),
	//		t.Month(),
	//		t.Day()+jump,
	//		0, 0, 0, 0,
	//		t.Location(),
	//	)
	//}
}

// Local date for the start of the work week, next Monday, with time 00:00:00.
func SOWW(t *Tart) TimeFunc {
	return weekJump(t, "monday", 0)
	//sd := weekday(t)
	//jump := days.jump(sd, "monday")
	//return func() time.Time {
	//	return time.Date(
	//		t.Year(),
	//		t.Month(),
	//		t.Day()+jump,
	//		0, 0, 0, 0,
	//		t.Location(),
	//	)
	//}
}

// Local date for the end of the work week, Friday night, with time 23:59:59.
func EOWW(t *Tart) TimeFunc {
	return weekJump(t, "friday", 0)
	//sd := weekday(t)
	//jump := days.jump(sd, "friday")
	//return func() time.Time {
	//	return time.Date(
	//		t.Year(),
	//		t.Month(),
	//		t.Day()+jump,
	//		0, 0, 0, 0,
	//		t.Location(),
	//	)
	//}
}

//1st, 2nd, ... 	Local date for the next Nth day, with time 00:00:00.
//func OrdinalDay(d string) time.Time {}

var months *rn = ring(monthsOfYear)

// Local date for the specified month(january, february, etc), 1st day, with time 00:00:00.
func NominalMonth(t *Tart, d string) RelativeFunc {
	return func(t *Tart) TimeFunc {
		sm := month(t)
		var mn int
		for idx, v := range monthsOfYear {
			if sm == v {
				mn = idx + 1
			}
		}
		jump := months.jump(sm, d)
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
	var mn int
	for idx, v := range monthsOfYear {
		if sm == v {
			mn = idx + 1
		}
	}
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
	var mn int
	for idx, v := range monthsOfYear {
		if sm == v {
			mn = idx + 1
		}
	}
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

// Whenver, later, someday 	Local 2044-10-16, with time 13:37:00.
// A date far away.
func Whenever(t *Tart) TimeFunc {
	return func() time.Time {
		return time.Date(
			2044,
			time.Month(10),
			16,
			13, 37, 0, 0,
			t.Location(),
		)
	}
}

func Any(t *Tart) TimeFunc {
	return func() time.Time {
		ret, _ := dateparse.ParseIn(t.last, t.Location())
		return ret
	}
}

//goodfriday 	Local date for the next Good Friday, with time 00:00:00.
//easter 	Local date for the next Easter Sunday, with time 00:00:00.
//eastermonday 	Local date for the next Easter Monday, with time 00:00:00.
//ascension 	Local date for the next Ascension (39 days after Easter Sunday), with time 00:00:00.
//pentecost 	Local date for the next Pentecost (40 days after Easter Sunday), with time 00:00:00.
//midsommar 	Local date for the Saturday after June 20th, with time 00:00:00. Swedish.
//midsommarafton 	Local date for the Friday after June 19th, with time 00:00:00. Swedish.
