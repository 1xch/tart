package tart

func HolidaysUS(t *Tart) {
	t.AddRelative(
		holidaysUSBase(t),
		holidaysUSX(t),
	)
}

func holidaysUSBase(t *Tart) map[string]RelativeFunc {
	return map[string]RelativeFunc{
		//4th of july
		//halloween
		//thanksgiving
		//christmas
	}
}

func HolidaysUSX(t *Tart) {
	t.AddRelative(holidaysUSX(t))
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

//black
////-mlk
////-junteenth
//christian
////-good friday
////-easter
//drinkin'
//hindu
//islamic
//jewish
//other
