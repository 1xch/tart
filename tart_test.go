package tart

import (
	"strings"
	"testing"
	"time"
)

func TestTart(t *testing.T) {
	tartInstance := New(HolidaysUS)
	testTimeExact := time.Date(2019, time.July, 4, 12, 0, 0, 0, time.Local)
	tartInstance.SetTime(testTimeExact)
	currYear := time.Now().Year()
	//lj4 := func() time.Time {
	//	var ret time.Time
	//	j4 := time.Date(currYear, time.July, 4, 0, 0, 0, 0, time.Local)
	//	n := time.Now()
	//	switch {
	//	case n.Before(j4):
	//		ret = time.Date(currYear-1, time.July, 4, 0, 0, 0, 0, time.Local)
	//	case n.After(j4):
	//		ret = j4
	//	}
	//	return ret
	//}
	//nj4 := func() time.Time {
	//	var ret time.Time
	//	j4 := time.Date(currYear, time.July, 4, 0, 0, 0, 0, time.Local)
	//	n := time.Now()
	//	switch {
	//	case n.Before(j4):
	//		ret = j4
	//	case n.After(j4):
	//		ret = time.Date(currYear+1, time.July, 4, 0, 0, 0, 0, time.Local)
	//	}
	//	return ret
	//}
	// base
	var testT = []struct {
		req string
		exp time.Time
	}{
		// manipulation of...
		// minute

		// hour
		//{"shift!-1h", time.Date(2019, time.July, 4, 11, 0, 0, 0, time.Local)},
		//{"shift!+1h", time.Date(2019, time.July, 4, 13, 0, 0, 0, time.Local)},
		//{"shiftFrom!eod,-1min", time.Date(2019, time.July, 4, 23, 58, 59, 0, time.Local)},
		//{"shiftFrom!eod,-1h", time.Date(2019, time.July, 4, 22, 59, 59, 0, time.Local)},
		//{"shiftFrom!soy,+1h", time.Date(2020, time.January, 1, 1, 0, 0, 0, time.Local)},
		// day
		{"July 4th 2019 at 11:00PM", time.Date(2019, time.July, 4, 23, 0, 0, 0, time.Local)},
		{"july 4th 2099", time.Date(2099, time.July, 4, 0, 0, 0, 0, time.Local)},
		{"july 4", time.Date(currYear, time.July, 4, 0, 0, 0, 0, time.Local)},
		//{"4th of july", time.Date(2019, time.July, 4, 0, 0, 0, 0, time.Local)},
		{"yesterday", time.Date(2019, time.July, 3, 0, 0, 0, 0, time.Local)},
		{"tomorrow", time.Date(2019, time.July, 5, 0, 0, 0, 0, time.Local)},
		{"sod", time.Date(2019, time.July, 5, 0, 0, 0, 0, time.Local)},
		{"eod", time.Date(2019, time.July, 4, 23, 59, 59, 0, time.Local)},
		{"tuesday", time.Date(2019, time.July, 9, 0, 0, 0, 0, time.Local)},
		//{"shift!-1d", time.Date(2019, time.July, 3, 12, 0, 0, 0, time.Local)},
		//{"shift!+1d", time.Date(2019, time.July, 5, 12, 0, 0, 0, time.Local)},
		//{"shiftFrom!tuesday,+1d", time.Date(2019, time.July, 10, 0, 0, 0, 0, time.Local)},
		//{"next!tuesday,0w", time.Date(2019, time.July, 9, 0, 0, 0, 0, time.Local)},
		//{"last!tuesday,1w", time.Date(2019, time.July, 2, 0, 0, 0, 0, time.Local)},
		//{"next!tuesday,1w", time.Date(2019, time.July, 16, 0, 0, 0, 0, time.Local)},
		//{"next!friday,0w", time.Date(2019, time.July, 5, 0, 0, 0, 0, time.Local)},
		//{"last!friday,1w", time.Date(2019, time.June, 28, 0, 0, 0, 0, time.Local)},
		//{"next!friday,1w", time.Date(2019, time.July, 12, 0, 0, 0, 0, time.Local)},
		// week
		{"sow", time.Date(2019, time.July, 7, 0, 0, 0, 0, time.Local)},
		{"sunday", time.Date(2019, time.July, 7, 0, 0, 0, 0, time.Local)},
		{"socw", time.Date(2019, time.June, 30, 0, 0, 0, 0, time.Local)},
		{"eow", time.Date(2019, time.July, 6, 0, 0, 0, 0, time.Local)},
		{"eocw", time.Date(2019, time.July, 6, 0, 0, 0, 0, time.Local)},
		{"soww", time.Date(2019, time.July, 8, 0, 0, 0, 0, time.Local)},
		{"eoww", time.Date(2019, time.July, 5, 23, 59, 59, 0, time.Local)},
		//shift
		//shiftFrom
		//{"next!week,1w",time.Date()},
		//{"last!week,1w",time.Date()},
		// month
		{"socm", time.Date(2019, time.July, 1, 0, 0, 0, 0, time.Local)},
		{"som", time.Date(2019, time.August, 1, 0, 0, 0, 0, time.Local)},
		{"eom", time.Date(2019, time.July, 31, 23, 59, 59, 0, time.Local)},
		{"eocm", time.Date(2019, time.July, 31, 23, 59, 59, 0, time.Local)},
		{"december", time.Date(2019, time.December, 1, 0, 0, 0, 0, time.Local)},
		{"june", time.Date(2020, time.June, 1, 0, 0, 0, 0, time.Local)},
		//{"shift!-1m", time.Date(2019, time.June, 4, 12, 0, 0, 0, time.Local)},
		//{"shift!+1m", time.Date(2019, time.August, 4, 12, 0, 0, 0, time.Local)},
		// shiftFrom
		// {"next!january,1m", time.Date(2020, time.January, 1, 0, 0, 0, 0, time.Local)},
		// {"last!january,1m", time.Date(2019, time.January, 1, 0, 0, 0, 0, time.Local)},
		// multimonth
		{"soq", time.Date(2019, time.October, 1, 0, 0, 0, 0, time.Local)},
		{"eoq", time.Date(2019, time.September, 30, 23, 59, 59, 59, time.Local)},
		// year
		{"soy", time.Date(2020, time.January, 1, 0, 0, 0, 0, time.Local)},
		{"eoy", time.Date(2019, time.December, 31, 0, 0, 0, 0, time.Local)},
		//{"shiftFrom!soy,-1y", time.Date(2019, time.January, 1, 0, 0, 0, 0, time.Local)},
		//{"next!july 4,0y", nj4()},
		//{"last!july 4,0y", lj4()},
		// {"last!july 4,1y", time.Date(currYear-1, time.July, 4, 0, 0, 0, 0, time.Local)},
		// {"next!july 4,1y", time.Date(currYear+1, time.July, 4, 0, 0, 0, 0, time.Local)},
		// {"last!christmas,1y", time.Date(2018, time.December, 25, 0, 0, 0, 0, time.Local)},
		// {"next!christmas,1y", time.Date(2019, time.December, 25, 0, 0, 0, 0, time.Local)},
		// misc
		{"someday", time.Date(2077, time.April, 27, 14, 37, 0, 0, time.Local)},
		//{"next!july 4,weekly", time.Date(currYear, time.July, 11, 0, 0, 0, 0, time.Local)},

	}
	for _, v := range testT {
		cmp := tartInstance.TimeOf(v.req)
		if !cmp.Equal(v.exp) {
			t.Errorf("%s expected %v, but got %v", strings.ToUpper(v.req), v.exp, cmp)
		}
	}

	// Association
	tartInstance.SetDirect("after", time.Date(2019, time.July, 5, 12, 0, 0, 0, time.Local))
	tartInstance.SetDirect("before", time.Date(2019, time.July, 3, 0, 0, 0, 0, time.Local))
	tartInstance.SetRelative("end", wrapRelativeFunc(time.Date(2019, time.July, 4, 23, 59, 59, 0, time.Local)))
	tartInstance.SetParsed("start", "july 2 2019 12:01 PM")
	var testAS = []struct {
		k, v string
		exp  time.Time
	}{
		{"as0", "start+1d", time.Date(2019, time.July, 3, 12, 1, 0, 0, time.Local)},
		{"as1", "before+12h3m", time.Date(2019, time.July, 3, 12, 3, 0, 0, time.Local)},
		{"as2", "end-59m", time.Date(2019, time.July, 4, 23, 0, 59, 0, time.Local)},
		{"as3", "after-13h", time.Date(2019, time.July, 4, 23, 0, 0, 0, time.Local)},
		{"as4", "now+1w", time.Date(2019, time.July, 11, 12, 0, 0, 0, time.Local)},
	}
	for _, v := range testAS {
		cmp, err := tartInstance.Associate(v.k, v.v)
		if err != nil {
			t.Errorf("associate: %s,%s got error -- %s", v.k, v.v, err.Error())
		}
		if !cmp.Equal(v.exp) {
			t.Errorf("associate: %s,%s expected %v, but got %v", v.k, v.v, v.exp, cmp)
		}
	}
	for _, v := range testAS {
		cmp := tartInstance.TimeOf(v.k)
		if !cmp.Equal(v.exp) {
			t.Errorf("TimeOf associated: %s expected %v, but got %v", strings.ToUpper(v.v), v.exp, cmp)
		}
	}

	// Duration
	testD := []struct {
		id       string
		pd       string
		when     time.Time
		instance bool
	}{
		{"7d+7d+7d+1h", "505h", time.Now(), true},
		{"hourly", "1h", time.Now(), true},
		{"weekly", "168h", time.Now(), true},
		{"1m", "744h", time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local), false},
		{"1m", "672h", time.Date(2019, 2, 1, 0, 0, 0, 0, time.Local), false},
		{"monthly", "744h", time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local), false},
		{"monthly", "672h", time.Date(2019, 2, 1, 0, 0, 0, 0, time.Local), false},
		//{"","",time.Date(2019, 2, 1, 0, 0, 0, 0, time.Local),true}
	}
	du := defaultUnits()
	dr := defaultReplace()
	for _, v := range testD {
		var id time.Duration
		var iErr error
		switch {
		case v.instance:
			id = tartInstance.DurationOf(v.id)
		case !v.instance:
			id, iErr = isDuration(v.id, v.when, du, dr)
		}
		if iErr != nil {
			t.Error(iErr.Error())
		}
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
