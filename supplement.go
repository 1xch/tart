package tart

import "time"

// Supplement is a struct holding supplementary data (string to units & string
// to replacement data).
type Supplement struct {
	*units
	*replace
}

func defaultSupplement() *Supplement {
	return &Supplement{
		defaultUnits(),
		defaultReplace(),
	}
}

type units struct {
	has map[string]float64
}

func defaultUnits() *units {
	return &units{defaultUnitMap()}
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
		//"year":    float64(time.Hour * 24 * 365),
		//"yr": float64(time.Hour * 24 * 365),
		//"y":       float64(time.Hour * 24 * 365),
	}
}

// MergeUnits ...
func (u *units) MergeUnits(in ...*units) {
	for _, nu := range in {
		for k, v := range nu.has {
			u.has[k] = v
		}
	}
}

// GetUnit ...
func (u *units) GetUnit(k string) (float64, bool) {
	if r, ok := u.has[k]; ok {
		return r, true
	}
	return 0, false
}

type replace struct {
	has map[string]string
}

func defaultReplace() *replace {
	r := &replace{make(map[string]string)}
	r.MergeMap(defaultReplacement(), ordinalReplacement())
	return r
}

// Add the provided Replace key values to the maintained map of replacements.
func (r *replace) MergeReplace(p ...*replace) {
	for _, pp := range p {
		r.MergeMap(pp.has)
	}
}

// MergeMap ...
func (r *replace) MergeMap(p ...map[string]string) {
	for _, pp := range p {
		for k, v := range pp {
			r.has[k] = v
		}
	}
}

// ReplaceWith ...
func (r *replace) ReplaceWith(in string) string {
	if v, ok := r.has[in]; ok {
		return v
	}
	return in
}

func defaultReplacement() map[string]string {
	return map[string]string{
		"hourly":       "1h",
		"daily":        "1d",
		"week":         "1w",
		"weekly":       "1w",
		"biweekly":     "2w",
		"fortnight":    "2w",
		"monthly":      "1m",
		"bimonthly":    "2m",
		"semiannually": "183d",
		"annually":     "1y",
		"biannually":   "2y",
		"quarterly":    "90d",
		"yearly":       "1y",
		"biyearly":     "2y",
	}
}

func ordinalReplacement() map[string]string {
	return map[string]string{
		"first":          "1",
		"1st":            "1",
		"second":         "2",
		"2nd":            "2",
		"third":          "3",
		"3rd":            "3",
		"fourth":         "4",
		"4th":            "4",
		"fifth":          "5",
		"5th":            "5",
		"sixth":          "6",
		"6th":            "6",
		"seventh":        "7",
		"7th":            "7",
		"eighth":         "8",
		"8th":            "8",
		"ninth":          "9",
		"9th":            "9",
		"tenth":          "10",
		"10th":           "10",
		"eleventh":       "11",
		"11th":           "11",
		"twelfth":        "12",
		"12th":           "12",
		"thirteenth":     "13",
		"13th":           "13",
		"fourteenth":     "14",
		"14th":           "14",
		"fifteenth":      "15",
		"15th":           "15",
		"sixteenth":      "16",
		"16th":           "16",
		"seventeenth":    "17",
		"17th":           "17",
		"eighteenth":     "18",
		"18th":           "18",
		"nineteenth":     "19",
		"19th":           "19",
		"twentieth":      "20",
		"20th":           "20",
		"twenty first":   "21",
		"twenty-first":   "21",
		"21st":           "21",
		"twenty second":  "22",
		"twenty-second":  "22",
		"22nd":           "22",
		"twenty third":   "23",
		"twenty-third":   "23",
		"23rd":           "23",
		"twenty fourth":  "24",
		"twenty-fourth":  "24",
		"24th":           "24",
		"twenty fifth":   "25",
		"twenty-fifth":   "25",
		"25th":           "25",
		"twenty sixth":   "26",
		"twenty-sixth":   "26",
		"26th":           "26",
		"twenty seventh": "27",
		"twenty-seventh": "27",
		"27th":           "27",
		"twenty eighth":  "28",
		"twenty-eighth":  "28",
		"28th":           "28",
		"twenty ninth":   "29",
		"twenty-ninth":   "29",
		"29th":           "29",
		"thirtieth":      "30",
		"30th":           "30",
		"thirty first":   "31",
		"thirty-first":   "31",
		"31st":           "31",
	}
}
