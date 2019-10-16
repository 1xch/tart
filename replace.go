package tart

// Add the provided Replace key values to the maintained map of replacements.
func (t *Tart) AddReplace(p ...Replace) {
	for _, pp := range p {
		for k, v := range pp {
			t.p[k] = v
		}
	}
}

type Replace map[string]string

func (r Replace) Replace(in string) string {
	if v, ok := r[in]; ok {
		return v
	}
	return in
}

func defaultReplace() Replace {
	return Replace{
		"hourly":       "1h",
		"daily":        "1d",
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
