package tart

import (
	"strings"
	"testing"
	"time"
)

func TestTart(t *testing.T) {
	tt := initialize(t)
	testSet(t, tt)
	testGet(t, tt)
	testDuration(t, tt)
}

type tTart struct {
	*Tart
	timeExact time.Time
	currYear  int
}

func initialize(t *testing.T) *tTart {
	tartInstance, iErr := New(SetTimeFmt(time.RFC3339), HolidaysBase)
	if iErr != nil {
		t.Error(iErr.Error())
	}
	testTimeExact := time.Date(2019, time.July, 4, 12, 0, 0, 0, time.Local)
	tartInstance.Establish(testTimeExact)
	if rErr := tartInstance.SetBatch(holidaysBase(tartInstance)); rErr != nil {
		t.Error(rErr.Error())
	}
	currYear := time.Now().Year()
	return &tTart{tartInstance, testTimeExact, currYear}
}

var nilTime time.Time = time.Time{}

func testSet(t *testing.T, tt *tTart) {
	testSet := []struct {
		k      string
		exp    time.Time
		expErr error
		sfn    func(*Tart) error
		gfn    func(*Tart) time.Time
	}{
		{
			"[SET 'start'=='july 2 2019 12:01 PM']",
			time.Date(2019, time.July, 2, 12, 1, 0, 0, time.Local),
			nil,
			func(x *Tart) error { return x.SetParsedDate("start", "july 2 2019 12:01 PM") },
			func(x *Tart) time.Time { return x.Get("start") },
		},
		{
			"[SETDIRECT  'after'=='time.Date(2019, time.July, 5, 12, 0, 0, 0, time.Local)']",
			time.Date(2019, time.July, 5, 12, 0, 0, 0, time.Local),
			nil,
			func(x *Tart) error {
				return x.SetDirect("after", time.Date(2019, time.July, 5, 12, 0, 0, 0, time.Local))
			},
			func(x *Tart) time.Time { return x.Get("after") },
		},
		{
			"[SETDIRECT 'before'=='time.Date(2019, time.July, 3, 0, 0, 0, 0, time.Local)']",
			time.Date(2019, time.July, 3, 0, 0, 0, 0, time.Local),
			nil,
			func(x *Tart) error {
				return x.SetDirect("before", time.Date(2019, time.July, 3, 0, 0, 0, 0, time.Local))
			},
			func(x *Tart) time.Time { return x.Get("before") },
		},
		{
			"[SETRELATION 'end'=='wrapRelative(time.Date(2019, time.July, 4, 23, 59, 59, 0, time.Local))']",
			time.Date(2019, time.July, 4, 23, 59, 59, 0, time.Local),
			nil,
			func(x *Tart) error {
				return x.SetRelation("end", wrapRelative(time.Date(2019, time.July, 4, 23, 59, 59, 0, time.Local)))
			},
			func(x *Tart) time.Time { return x.Get("end") },
		},
		{
			"[SETPARSEDDATE 'fireworks'=='July 4 2019 at 9PM']",
			time.Date(2019, time.July, 4, 21, 0, 0, 0, time.Local),
			nil,
			func(x *Tart) error { return x.SetParsedDate("fireworks", "July 4, 2019 9:00:00 PM") },
			func(x *Tart) time.Time { return x.Get("fireworks") },
		},
		{
			"[SETUNIX 'lunch'=='1562256000']",
			time.Date(2019, time.July, 4, 12, 0, 0, 0, time.Local),
			nil,
			func(x *Tart) error { return x.SetFloat("lunch", 1562256000) },
			func(x *Tart) time.Time { return x.Get("lunch") },
		},
		{
			"[SET 'as0'=='>1d!start']",
			time.Date(2019, time.July, 3, 12, 1, 0, 0, time.Local),
			nil,
			func(x *Tart) error { return x.Set("as0", ">1d!start") },
			func(x *Tart) time.Time { return x.Get("as0") },
		},
		{
			"[SET 'as1'=='>12h3m!before'",
			time.Date(2019, time.July, 3, 12, 3, 0, 0, time.Local),
			nil,
			func(x *Tart) error { return x.Set("as1", ">12h3m!before") },
			func(x *Tart) time.Time { return x.Get("as1") },
		},
		{
			"[SET 'as2'=='-59m!end']",
			time.Date(2019, time.July, 4, 23, 0, 59, 0, time.Local),
			nil,
			func(x *Tart) error { return x.Set("as2", "-59m!end") },
			func(x *Tart) time.Time { return x.Get("as2") },
		},
		{
			"[SET 'as3'=='<13h!after']",
			time.Date(2019, time.July, 4, 23, 0, 0, 0, time.Local),
			nil,
			func(x *Tart) error { return x.Set("as3", "<13h!after") },
			func(x *Tart) time.Time { return x.Get("as3") },
		},
		{
			"[SET 'as4'=='>>1w",
			time.Date(2019, time.July, 18, 12, 0, 0, 0, time.Local),
			nil,
			func(x *Tart) error { return x.Set("as4", ">>1w") },
			func(x *Tart) time.Time { return x.Get("as4") },
		},
		{
			"[SET 'check SET reservedKeyError']",
			nilTime,
			reservedKeyError("any"),
			func(x *Tart) error { return x.Set("any", "!end") },
			nil,
		},
		{
			"[SETRELATION 'check SETRELATION reservedKeyError']",
			nilTime,
			reservedKeyError("any"),
			func(x *Tart) error { return x.SetRelation("any", newRelation(func(*Tart) TimeFunc { return nil })) },
			nil,
		},
		{
			"[SETDIRECT 'check SETDIRECT reservedKeyError']",
			nilTime,
			reservedKeyError("any"),
			func(x *Tart) error { return x.SetDirect("any", nilTime) },
			nil,
		},
		{
			"[SETPARSEDDATE 'check SETPARSEDDATE reservedKeyError']",
			nilTime,
			reservedKeyError("any"),
			func(x *Tart) error { return x.SetParsedDate("any", "january 1 1927") },
			nil,
		},
		{
			"[SETFLOAT 'check SETFLOAT reservedKeyError']",
			nilTime,
			reservedKeyError("any"),
			func(x *Tart) error { return x.SetFloat("any", 0) },
			nil,
		},
	}
	for _, v := range testSet {
		if v.sfn != nil {
			err := v.sfn(tt.Tart)
			switch {
			case v.expErr != nil:
				if strings.Compare(v.expErr.Error(), err.Error()) != 0 {
					t.Errorf("expected error '%s' but got error '%s'", v.expErr, err)
				}
			case v.expErr == nil:
				if err != nil {
					t.Errorf("%s: resulted in err -- %s", v.k, err.Error())
				}
			}
		}
	}
	for _, v := range testSet {
		if v.gfn != nil && v.exp != nilTime {
			cmp := v.gfn(tt.Tart)
			if !cmp.Equal(v.exp) {
				t.Errorf("%s: expected %v, but got %v", v.k, v.exp, cmp)
			}
		}
	}
}

func testGet(t *testing.T, tt *tTart) {
	testGet := []struct {
		req string
		exp time.Time
	}{
		// dot is now
		{"!", tt.timeExact},
		{"", tt.timeExact},
		// manipulation of...
		// minute
		{">1m", time.Date(2019, time.July, 4, 12, 1, 0, 0, time.Local)},
		{"+1m", time.Date(2019, time.July, 4, 12, 1, 0, 0, time.Local)},
		{"<1m", time.Date(2019, time.July, 4, 11, 59, 0, 0, time.Local)},
		{"-1m", time.Date(2019, time.July, 4, 11, 59, 0, 0, time.Local)},
		{">>>>>1m", time.Date(2019, time.July, 4, 12, 5, 0, 0, time.Local)},
		{"+++++1m", time.Date(2019, time.July, 4, 12, 5, 0, 0, time.Local)},
		{"<<<<<1m", time.Date(2019, time.July, 4, 11, 55, 0, 0, time.Local)},
		{"-----1m", time.Date(2019, time.July, 4, 11, 55, 0, 0, time.Local)},
		{">1m!eod", time.Date(2019, time.July, 5, 0, 0, 59, 0, time.Local)},
		{"<1m!eod", time.Date(2019, time.July, 4, 23, 58, 59, 0, time.Local)},
		{"+1m+1m", time.Date(2019, time.July, 4, 12, 2, 0, 0, time.Local)},
		{"-1m-1m", time.Date(2019, time.July, 4, 11, 58, 0, 0, time.Local)},
		{"+1m-1m", time.Date(2019, time.July, 4, 12, 0, 0, 0, time.Local)},
		{"-1m+1m", time.Date(2019, time.July, 4, 12, 0, 0, 0, time.Local)},
		{">1m<1m", time.Date(2019, time.July, 4, 12, 0, 0, 0, time.Local)},
		{">1m<1m+1m-1m", time.Date(2019, time.July, 4, 12, 0, 0, 0, time.Local)},
		// hour
		{"<1h", time.Date(2019, time.July, 4, 11, 0, 0, 0, time.Local)},
		{">1h", time.Date(2019, time.July, 4, 13, 0, 0, 0, time.Local)},
		{"<1h!eod", time.Date(2019, time.July, 4, 22, 59, 59, 0, time.Local)},
		{">1h!soy", time.Date(2020, time.January, 1, 1, 0, 0, 0, time.Local)},
		// day
		{"!today", time.Date(2019, time.July, 4, 0, 0, 0, 0, time.Local)},
		{"!July 4th 2019 at 11:00PM", time.Date(2019, time.July, 4, 23, 0, 0, 0, time.Local)},
		{"!july 4th 2099", time.Date(2099, time.July, 4, 0, 0, 0, 0, time.Local)},
		{"!july 4", time.Date(tt.currYear, time.July, 4, 0, 0, 0, 0, time.Local)},
		{"!yesterday", time.Date(2019, time.July, 3, 0, 0, 0, 0, time.Local)},
		{"+1d!yesterday", time.Date(2019, time.July, 4, 0, 0, 0, 0, time.Local)},
		{"++1d!yesterday", time.Date(2019, time.July, 5, 0, 0, 0, 0, time.Local)},
		{"!tomorrow", time.Date(2019, time.July, 5, 0, 0, 0, 0, time.Local)},
		{"!sod", time.Date(2019, time.July, 5, 0, 0, 0, 0, time.Local)},
		{"!eod", time.Date(2019, time.July, 4, 23, 59, 59, 0, time.Local)},
		{"!tuesday", time.Date(2019, time.July, 9, 0, 0, 0, 0, time.Local)},
		{"<1d", time.Date(2019, time.July, 3, 12, 0, 0, 0, time.Local)},
		{"+1d", time.Date(2019, time.July, 5, 12, 0, 0, 0, time.Local)},
		{"+1d!tuesday", time.Date(2019, time.July, 10, 0, 0, 0, 0, time.Local)},
		{"!tuesday", time.Date(2019, time.July, 9, 0, 0, 0, 0, time.Local)},
		{"-1w!tuesday", time.Date(2019, time.July, 2, 0, 0, 0, 0, time.Local)},
		{">1w!tuesday", time.Date(2019, time.July, 16, 0, 0, 0, 0, time.Local)},
		{"!christmas", time.Date(2019, time.December, 25, 12, 0, 0, 0, time.Local)},
		// week
		{"!sow", time.Date(2019, time.July, 7, 0, 0, 0, 0, time.Local)},
		{"sunday", time.Date(2019, time.July, 7, 0, 0, 0, 0, time.Local)},
		{"!socw", time.Date(2019, time.June, 30, 0, 0, 0, 0, time.Local)},
		{"eow", time.Date(2019, time.July, 6, 0, 0, 0, 0, time.Local)},
		{"!eocw", time.Date(2019, time.July, 6, 0, 0, 0, 0, time.Local)},
		{"soww", time.Date(2019, time.July, 8, 0, 0, 0, 0, time.Local)},
		{"!eoww", time.Date(2019, time.July, 5, 23, 59, 59, 0, time.Local)},
		// month
		{"!socm", time.Date(2019, time.July, 1, 0, 0, 0, 0, time.Local)},
		{"!som", time.Date(2019, time.August, 1, 0, 0, 0, 0, time.Local)},
		{"eom", time.Date(2019, time.July, 31, 23, 59, 59, 0, time.Local)},
		{"!eocm", time.Date(2019, time.July, 31, 23, 59, 59, 0, time.Local)},
		{"!december", time.Date(2019, time.December, 1, 0, 0, 0, 0, time.Local)},
		{"june", time.Date(2020, time.June, 1, 0, 0, 0, 0, time.Local)},
		{">5months!", time.Date(2019, time.December, 4, 12, 0, 0, 0, time.Local)},
		// multimonth
		{"!soq", time.Date(2019, time.October, 1, 0, 0, 0, 0, time.Local)},
		{">2h!soq", time.Date(2019, time.October, 1, 2, 0, 0, 0, time.Local)},
		{"!eoq", time.Date(2019, time.September, 30, 23, 59, 59, 59, time.Local)},
		{"<2h!eoq", time.Date(2019, time.September, 30, 21, 59, 59, 59, time.Local)},
		// year
		{"!soy", time.Date(2020, time.January, 1, 0, 0, 0, 0, time.Local)},
		{"!eoy", time.Date(2019, time.December, 31, 0, 0, 0, 0, time.Local)},
		{"<1year!", time.Date(2018, time.July, 4, 12, 0, 0, 0, time.Local)},
		// multiple duration value shifts
		{">1m2s", time.Date(2019, time.July, 4, 12, 1, 2, 0, time.Local)},
		{">>>>>1m2s", time.Date(2019, time.July, 4, 12, 5, 10, 0, time.Local)},
		// misc
		{"!someday", time.Date(2077, time.April, 27, 14, 37, 0, 0, time.Local)},
		// purposeful test reduplications
		{">1m", time.Date(2019, time.July, 4, 12, 1, 0, 0, time.Local)},
		{">2h!soq", time.Date(2019, time.October, 1, 2, 0, 0, 0, time.Local)},
		{"!eocm", time.Date(2019, time.July, 31, 23, 59, 59, 0, time.Local)},
	}
	ti := tt.Tart
	for _, v := range testGet {
		cmp := ti.Get(v.req)
		if !cmp.Equal(v.exp) {
			t.Errorf("%s expected %v, but got %v", strings.ToUpper(v.req), v.exp, cmp)
		}
	}
}

func testDuration(t *testing.T, tt *tTart) {
	testDuration := []struct {
		id string
		pd string
	}{
		{">7d>7d>7d+1h", "505h"},
		{"<7d<7d<7d-1h", "505h"},
		{">1h", "1h"},
		{"<1h", "1h"},
		//{">1month", "744h"}, // is current month so requires more thought
	}
	ti := tt.Tart
	for _, v := range testDuration {
		id := ti.Duration(v.id)
		pd, pErr := time.ParseDuration(v.pd)
		if pErr != nil {
			t.Error(pErr.Error())
		}
		if id != pd {
			t.Errorf("unequal durations: %v != %v", id, pd)
		}
	}
}

func TestRing(t *testing.T) {
	ringTest := []string{"one", "two", "three", "four", "five"}
	rt := ring(ringTest)
	var rts = []struct {
		a, b string
		exp  int
	}{
		{"one", "five", 4},
		{"three", "four", 1},
		{"one", "three", 2},
		{"three", "two", 4},
	}
	for _, v := range rts {
		if res := rt.jump(v.a, v.b); res != v.exp {
			t.Errorf("jump %s-%s expected %d, but got %d", v.a, v.b, v.exp, res)
		}
	}
}
