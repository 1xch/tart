package tart

import "time"

type Supplement struct {
	*Units
	*Replace
}

func defaultSupplement() *Supplement {
	return &Supplement{
		defaultUnits(),
		defaultReplace(),
	}
}

type Units struct {
	has map[string]float64
}

func defaultUnits() *Units {
	return &Units{defaultUnitMap()}
}

func defaultUnitMap() map[string]float64 {
	return map[string]float64{
		"ns":      float64(time.Nanosecond),
		"us":      float64(time.Microsecond),
		"µs":      float64(time.Microsecond), // U+00B5 = micro symbol
		"μs":      float64(time.Microsecond), // U+03BC = Greek letter mu
		"ms":      float64(time.Millisecond),
		"s":       float64(time.Second),
		"sec":     float64(time.Second),
		"second":  float64(time.Second),
		"seconds": float64(time.Second),
		"min":     float64(time.Minute),
		"minute":  float64(time.Minute),
		"minutes": float64(time.Minute),
		"h":       float64(time.Hour),
		"hr":      float64(time.Hour),
		"hour":    float64(time.Hour),
		"hours":   float64(time.Hour),
		"d":       float64(time.Hour * 24),
		"day":     float64(time.Hour * 24),
		"days":    float64(time.Hour * 24),
		"w":       float64(time.Hour * 24 * 7),
		"week":    float64(time.Hour * 24 * 7),
		"weeks":   float64(time.Hour * 24 * 7),
		"wk":      float64(time.Hour * 24 * 7),
	}
}

func (u *Units) MergeUnits(in ...*Units) {
	for _, nu := range in {
		for k, v := range nu.has {
			u.has[k] = v
		}
	}
}

func (u *Units) GetUnit(k string) (float64, bool) {
	if r, ok := u.has[k]; ok {
		return r, true
	}
	return 0, false
}

type Replace struct {
	has map[string]string
}

func defaultReplace() *Replace {
	return &Replace{defaultReplacements()}
}

func defaultReplacements() map[string]string {
	return map[string]string{
		"hourly":       "1h",
		"daily":        "1d",
		"week":         "1w",
		"weekly":       "1w",
		"biweekly":     "2w",
		"fortnight":    "2w",
		"monthly":      "1m", // monthly calculation is a fuzzy calculation here,
		"bimonthly":    "2m", // use as absolute with care
		"semiannually": "183d",
		"annually":     "1y",
		"biannually":   "2y",
		"quarterly":    "90d",
		"yearly":       "1y",
		"biyearly":     "2y",
	}
}

func (r *Replace) ReplaceWith(in string) string {
	if v, ok := r.has[in]; ok {
		return v
	}
	return in
}

// Add the provided Replace key values to the maintained map of replacements.
func (r *Replace) MergeReplace(p ...*Replace) {
	for _, pp := range p {
		for k, v := range pp.has {
			r.has[k] = v
		}
	}
}
