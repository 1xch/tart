package tart

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/araddon/dateparse"
)

type (
	// RelativeFunc ...
	RelativeFunc func(*Tart) TimeFunc
	// TimeFunc ...
	TimeFunc func() time.Time
)

// Relation ...
type Relation interface {
	Relative(*Tart) TimeFunc
}

type relation struct {
	rfn RelativeFunc
}

func newRelation(rfn RelativeFunc) *relation {
	return &relation{rfn}
}

func (r *relation) Relative(t *Tart) TimeFunc {
	return r.rfn(t)
}

// relations is a struct managing core time relations for a Tart instance.
type relations struct {
	t              *Tart
	storedRelation map[string]Relation
	storedTfn      map[string]TimeFunc
	rk             []string
}

func newRelations(t *Tart) *relations {
	r := &relations{}
	r.reset(t)
	return r
}

func (r *relations) reset(t *Tart) {
	r.t = t
	r.storedRelation, r.rk = defaultRelativeFuncs(r.t)
	r.storedTfn = make(map[string]TimeFunc)
}

func defaultRelativeFuncs(t *Tart) (map[string]Relation, []string) {
	r := map[string]Relation{
		"any":       newRelation(Any),
		"default":   newRelation(Any),
		"eocm":      newRelation(EOM),
		"eocw":      newRelation(EOW),
		"eod":       newRelation(EOD),
		"eom":       newRelation(EOM),
		"eoq":       newRelation(EOQ),
		"eow":       newRelation(EOW),
		"eoww":      newRelation(EOWW),
		"eoy":       newRelation(EOY),
		"later":     newRelation(Whenever),
		"now":       newRelation(Now),
		"socm":      newRelation(SOCM),
		"socw":      newRelation(SOCW),
		"sod":       newRelation(Tomorrow),
		"som":       newRelation(SOM),
		"someday":   newRelation(Whenever),
		"soq":       newRelation(SOQ),
		"sow":       newRelation(SOW),
		"soww":      newRelation(SOWW),
		"soy":       newRelation(SOY),
		"today":     newRelation(Today),
		"tomorrow":  newRelation(Tomorrow),
		"whenever":  newRelation(Whenever),
		"yesterday": newRelation(Yesterday),
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
func (r *relations) popTimeFn() TimeFunc {
	t := r.t
	d := t.last

	if etfn, ok := r.storedTfn[d.origin]; ok {
		return etfn
	}

	var rfn RelativeFunc

	rl := r.GetRelation(d.phrase)
	if rl != nil {
		rfn = rl.Relative
	} else {
		rfn = r.storedRelation["default"].Relative
	}

	tfn := rfn(t)

	r.storedTfn[d.origin] = tfn

	return tfn
}

// GetRelation ...
func (r *relations) GetRelation(k string) Relation {
	if gr, ok := r.storedRelation[k]; ok {
		return gr
	}
	return nil
}

func reservedKeyError(k string) error {
	return fmt.Errorf("'%s' already exists as a relation and is a reserved key", k)
}

func isReservedKey(rk []string, k string) bool {
	for _, v := range rk {
		if k == v {
			return true
		}
	}
	return false
}

// SetRelation ...
func (r *relations) SetRelation(k string, v Relation) error {
	if !isReservedKey(r.rk, k) {
		r.storedRelation[k] = v
		return nil
	}
	return reservedKeyError(k)
}

// SetDirect ...
func (r *relations) SetDirect(k string, v time.Time) error {
	if !isReservedKey(r.rk, k) {
		r.storedRelation[k] = wrapRelative(v)
		return nil
	}
	return reservedKeyError(k)
}

func wrapRelative(t time.Time) Relation {
	return newRelation(func(ti *Tart) TimeFunc {
		return func() time.Time {
			return pumpShift(t, ti.last)
		}
	})
}

// SetParsedDate ...
func (r *relations) SetParsedDate(k, v string) error {
	if isReservedKey(r.rk, k) {
		return reservedKeyError(k)
	}
	now := time.Now()
	t, pErr := dateparse.ParseIn(v, r.t.Location())
	if pErr != nil {
		return pErr
	}
	if y := t.Year(); y <= 0 {
		t = t.AddDate(now.Year(), 0, 0)
	}
	r.storedRelation[k] = wrapRelative(t)
	return nil
}

// SetFloat ...
func (r *relations) SetFloat(k string, v float64) error {
	trunc := math.Trunc(v)
	nanos := fractionToNanos(v - trunc)
	return r.SetRelation(k, wrapRelative(time.Unix(int64(trunc), nanos)))
}

func fractionToNanos(fraction float64) int64 {
	return int64(fraction * float64(time.Second/time.Nanosecond))
}

// SetBatch ...
func (r *relations) SetBatch(in ...map[string]Relation) error {
	var err error
	for _, v := range in {
		for kk, vv := range v {
			err = r.SetRelation(kk, vv)
			if err != nil {
				return err
			}
		}
	}
	return err
}

// Now returns TimeFunc for "now" from string "now" where "now" is tart.Time
func Now(t *Tart) TimeFunc {
	nt := pumpShift(t.Time, t.last)
	return func() time.Time {
		return nt
	}
}

func pumpShift(t time.Time, d *directive) time.Time {
	if d != nil {
		sh := d.Shift()
		if len(sh) > 0 {
			for _, v := range sh {
				t = t.Add(v.dur).AddDate(v.y, v.m, v.d)
			}
		}
	}
	return t
}

// Yesterday returns TimeFunc giving local date for yesterday, with time 00:00:00.
func Yesterday(t *Tart) TimeFunc {
	yt := t.Add(-(time.Hour * 24))
	yd := time.Date(
		yt.Year(),
		yt.Month(),
		yt.Day(),
		0, 0, 0, 0,
		yt.Location(),
	)
	yd = pumpShift(yd, t.last)
	return func() time.Time {
		return yd
	}
}

// Today returns TimeFunc giving current local date, with time 00:00:00.
func Today(t *Tart) TimeFunc {
	tn := t.Time
	td := time.Date(
		tn.Year(),
		tn.Month(),
		tn.Day(),
		0, 0, 0, 0,
		tn.Location(),
	)
	td = pumpShift(td, t.last)
	return func() time.Time {
		return td
	}
}

// EOD returns TimeFunc for "eod" where end of day is current local date, with time 23:59:59.
func EOD(t *Tart) TimeFunc {
	tn := t.Time
	eod := time.Date(
		tn.Year(),
		tn.Month(),
		tn.Day(),
		23, 59, 59, 0,
		tn.Location(),
	)
	eod = pumpShift(eod, t.last)
	return func() time.Time {
		return eod
	}
}

// Tomorrow returns TimeFunc for "tomorrow" as local date for tomorrow, with time 00:00:00. Same as sod(start of day).
func Tomorrow(t *Tart) TimeFunc {
	tt := t.Add(time.Hour * 24)
	tm := time.Date(
		tt.Year(),
		tt.Month(),
		tt.Day(),
		0, 0, 0, 0,
		tt.Location(),
	)
	tm = pumpShift(tm, t.last)
	return func() time.Time {
		return tm
	}
}

func days() *rn {
	return ring(daysOfWeek())
}

func weekday(t *Tart) string {
	return strings.ToLower(t.Weekday().String())
}

// NominalDay returns a Relation. The subsequent TimeFunc returned generates
// local date for the specified day(monday, tuesday, etc), after today, with
// time 00:00:00.
func NominalDay(t *Tart, d string) Relation {
	return newRelation(func(t *Tart) TimeFunc {
		sd := weekday(t)
		dys := days()
		jump := dys.jump(sd, d)
		nd := time.Date(
			t.Year(),
			t.Month(),
			t.Day()+jump,
			0, 0, 0, 0,
			t.Location(),
		)
		nd = pumpShift(nd, t.last)
		return func() time.Time {
			return nd
		}
	})
}

func weekJump(t *Tart, v string, sub int, timeSub ...int) TimeFunc {
	sd := weekday(t)
	dys := days()
	jump := dys.jump(sd, v) - sub
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
	w := time.Date(
		t.Year(),
		t.Month(),
		t.Day()+jump,
		hour, min, secs, 0,
		t.Location(),
	)
	w = pumpShift(w, t.last)
	return func() time.Time {
		return w
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

func months() *rn {
	return ring(monthsOfYear())
}

func monthString(t *Tart) string {
	return strings.ToLower(t.Month().String())
}

// NominalMonth returns a RelativeFunc returning a subsequent TimeFunc for
// local date for the specified month(january, february, etc), 1st day, with
// time 00:00:00.
func NominalMonth(t *Tart, m string) Relation {
	return newRelation(func(t *Tart) TimeFunc {
		sm := monthString(t)
		mths := months()
		mn := mths.jump("january", sm) + 1
		jump := mths.jump(sm, m)
		nm := time.Date(
			t.Year(),
			time.Month(mn+jump),
			1,
			0, 0, 0, 0,
			t.Location(),
		)
		nm = pumpShift(nm, t.last)
		return func() time.Time {
			return nm
		}
	})
}

// SOCM returns TimeFunc for local date for the 1st day of the current month,
// with time 00:00:00.
func SOCM(t *Tart) TimeFunc {
	sm := time.Date(
		t.Year(),
		t.Month(),
		1,
		0, 0, 0, 0,
		t.Location(),
	)
	sm = pumpShift(sm, t.last)
	return func() time.Time {
		return sm
	}
}

// SOM returns TimeFunc providing local date for the 1st day of the next month,
// with time 00:00:00.
func SOM(t *Tart) TimeFunc {
	sm := monthString(t)
	mths := months()
	mn := mths.jump("january", sm) + 1
	m := time.Date(
		t.Year(),
		time.Month(mn+1),
		1,
		0, 0, 0, 0,
		t.Location(),
	)
	m = pumpShift(m, t.last)
	return func() time.Time {
		return m
	}
}

// EOM returns TimeFunc providing local date for the last day of the current
// month, with time 23:59:59.
func EOM(t *Tart) TimeFunc {
	sm := monthString(t)
	mths := months()
	mn := mths.jump("january", sm) + 1
	d := time.Date(
		t.Year(),
		time.Month(mn+1),
		1,
		23, 59, 59, 0,
		t.Location(),
	)
	d = d.Add(-(24 * time.Hour))
	d = pumpShift(d, t.last)
	return func() time.Time {
		return d
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
	var qt time.Time
	switch {
	case t.After(q["Q1"]) && t.Before(q["Q1x"]):
		qt = q["Q2"]
	case t.After(q["Q2"]) && t.Before(q["Q2x"]):
		qt = q["Q3"]
	case t.After(q["Q3"]) && t.Before(q["Q3x"]):
		qt = q["Q4"]
	case t.After(q["Q4"]) && t.Before(q["Q4x"]):
		qt = q["Q1"]
	}

	qt = pumpShift(qt, t.last)

	return func() time.Time {
		return qt
	}
}

// EOQ returns TimeFunc providing local date for the end of the current quarter
// (March, June, September, December), last day of the month, with time
// 23:59:59.
func EOQ(t *Tart) TimeFunc {
	q := quarters(t.Year(), t.Location())
	var qt time.Time
	switch {
	case t.After(q["Q1"]) && t.Before(q["Q1x"]):
		qt = q["Q1x"]
	case t.After(q["Q2"]) && t.Before(q["Q2x"]):
		qt = q["Q2x"]
	case t.After(q["Q3"]) && t.Before(q["Q3x"]):
		qt = q["Q3x"]
	case t.After(q["Q4"]) && t.Before(q["Q4x"]):
		qt = q["Q4x"]
	}

	qt = pumpShift(qt, t.last)

	return func() time.Time {
		return qt
	}
}

// SOY returns TimeFunc providing local date for the next year, January 1st,
// with time 00:00:00.
func SOY(t *Tart) TimeFunc {
	sy := time.Date(
		t.Year()+1,
		time.January,
		1,
		0, 0, 0, 0,
		t.Location(),
	)
	sy = pumpShift(sy, t.last)

	return func() time.Time {
		return sy
	}
}

// EOY returns TimeFunc providing local date for this year, December 31st, with
// time 00:00:00.
func EOY(t *Tart) TimeFunc {
	ey := time.Date(
		t.Year(),
		time.December,
		31,
		0, 0, 0, 0,
		t.Location(),
	)
	ey = pumpShift(ey, t.last)

	return func() time.Time {
		return ey
	}
}

// Whenever returns TimeFunc for "whenever", "later", "someday" mapped to local
// 2077-04-27, with time 14:37:00. A date far away.
func Whenever(t *Tart) TimeFunc {
	we := time.Date(
		2077,
		time.Month(4),
		27,
		14, 37, 0, 0,
		t.Location(),
	)
	we = pumpShift(we, t.last)

	return func() time.Time {
		return we
	}
}

// Any returns TimeFunc that attempts to parse Tart.last to a valid time.
func Any(t *Tart) TimeFunc {
	now := time.Now()
	d := t.last
	ret, _ := dateparse.ParseIn(d.phrase, t.Location())
	if y := ret.Year(); y <= 0 {
		ret = ret.AddDate(now.Year(), 0, 0)
	}
	ret = pumpShift(ret, t.last)

	return func() time.Time {
		return ret
	}
}
