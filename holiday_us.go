package tart

// HolidaysUS ...
func HolidaysUS(t *Tart) {
	_ = t.SetBatch(
		holidaysUSBase(t),
		holidaysUSX(t),
	)
}

// HolidaysUSBase ...
func HolidaysUSBase(t *Tart) {
	_ = t.SetBatch(holidaysUSBase(t))
}

func holidaysUSBase(t *Tart) map[string]RelativeFunc {
	return map[string]RelativeFunc{
		//4th of july
		//halloween
		//thanksgiving
		//christmas
	}
}

// HolidaysUSX ...
func HolidaysUSX(t *Tart) {
	_ = t.SetBatch(holidaysUSX(t))
}

func holidaysUSX(t *Tart) map[string]RelativeFunc {
	return map[string]RelativeFunc{
		//mlk
		//presidents
		//valentines
		//st patricks
		//memorial
		//mothers
		//fathers
		//juneteenth
		//labor
	}
}
