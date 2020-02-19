package tart

import (
	"time"
)

// HolidaysBase ...
func HolidaysBase(t *Tart) error {
	return t.SetBatch(
		holidaysBase(t),
	)
}

func holidaysBase(*Tart) map[string]Relation {
	return map[string]Relation{
		"christmas": newRelation(christmas()),
	}
}

func christmas() RelativeFunc {
	return func(t *Tart) TimeFunc {
		yr := t.Year()
		xmas := time.Date(yr, time.December, 25, 12, 0, 0, 0, time.Local)
		if t.After(xmas) {
			time.Date(yr+1, time.December, 25, 12, 0, 0, 0, time.Local)
		}
		xmas = pumpShift(xmas, t.last)
		return func() time.Time {
			return xmas
		}
	}
}
