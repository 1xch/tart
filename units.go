package tart

import "time"

// Add the provided Units key values to the instance maintained map of units.
func (t *Tart) AddUnits(u ...Units) {
	for _, uu := range u {
		for k, v := range uu {
			t.u[k] = v
		}
	}
}

type Units map[string]float64

func defaultUnits() Units {
	return Units{
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
