package tart

import (
	"testing"
	"time"
)

func TestTart(t *testing.T) {
	testTimeExact := time.Date(2019, time.July, 4, 12, 0, 0, 0, time.Local)
	tartInstance := NewTart(testTimeExact)
	var testT = []struct {
		req string
		exp time.Time
	}{
		{"yesterday", time.Date(2019, time.July, 3, 0, 0, 0, 0, time.Local)},
		{"tomorrow", time.Date(2019, time.July, 5, 0, 0, 0, 0, time.Local)},
		{"sod", time.Date(2019, time.July, 5, 0, 0, 0, 0, time.Local)},
		{"eod", time.Date(2019, time.July, 4, 23, 59, 59, 0, time.Local)},
		{"tuesday", time.Date(2019, time.July, 9, 0, 0, 0, 0, time.Local)},
		{"december", time.Date(2019, time.December, 1, 0, 0, 0, 0, time.Local)},
		//{},
		//{"sow":       SOW},
		//{"socw":      SOCW},
		//{"eow":       EOW},
		//{"eocw":      EOW},
		//{"soww":      SOWW},
		//{"eoww":      EOWW},
		//{"socm":      SOCM},
		//{"som":       SOM},
		//{"eom":       EOM},
		//{"eocm":      EOM},
		//{"soq":       SOQ},
		//{"eoq":       EOQ},
		//{"soy":       SOY},
		//{"eoy":       EOY},
		{"someday", time.Date(2044, time.October, 16, 13, 37, 0, 0, time.Local)},
	}
	for _, v := range testT {
		cmp, err := tartInstance.TimeOf(v.req)
		if err != nil {
			t.Error(err.Error())
		}
		if !cmp.Equal(v.exp) {
			t.Errorf("expected %v, but got %v", v.exp, cmp)
		}
	}

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
		//{"", }
	}
	du := defaultUnits()
	dr := defaultReplace()
	for _, v := range testD {
		var id time.Duration
		var iErr error
		switch {
		case v.instance:
			id, iErr = tartInstance.DurationOf(v.id)
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
	//	cts := []string{"one", "two", "three", "four", "five"}
	//	c := ring(cts)
	//	//spew.Dump(c)
	//	spew.Dump(c.jump("two", "one"))
	//	spew.Dump(c.jump("one", "five"))
	//  spew.Dump(days.jump("monday", "friday"))
	// spew.Dump(days)
	//spew.Dump(days.jump("monday", "friday"))
	//spew.Dump(days.jump("friday", "wednesday"))
	//spew.Dump((months.jump("january", "december")))
	//spew.Dump((months.jump("december", "january")))
	//spew.Dump((months.jump("july", "march")))
}
