package date

// weekdayNameSet holds lowercase weekday names that may appear as a
// leading word in human-written dates ("Mon Jan 02 2006",
// "Wednesday, 28 Feb 2018"). The parser drops them before format
// dispatch.
var weekdayNameSet = map[string]struct{}{
	"mon": {}, "monday": {},
	"tue": {}, "tues": {}, "tuesday": {},
	"wed": {}, "wednesday": {},
	"thu": {}, "thur": {}, "thurs": {}, "thursday": {},
	"fri": {}, "friday": {},
	"sat": {}, "saturday": {},
	"sun": {}, "sunday": {},
}

var monthNameMap = map[string]int{
	"jan":     1,
	"jän":     1,
	"jaen":    1,
	"januar":  1,
	"january": 1,
	"jänner":  1,
	"jaenner": 1,
	"janv":    1, // French (janvier)
	"janvier": 1,
	"gen":     1, // Italian (gennaio)
	"gennaio": 1,
	"ene":     1, // Spanish (enero)
	"enero":   1,

	"ian":     1,
	"iän":     1,
	"iaen":    1,
	"ianuar":  1,
	"ianuary": 1,
	"iänner":  1,
	"iaenner": 1,

	"1an":     1,
	"1än":     1,
	"1aen":    1,
	"1anuar":  1,
	"1anuary": 1,
	"1änner":  1,
	"1aenner": 1,

	"feb":      2,
	"febr":     2,
	"februar":  2,
	"february": 2,
	"févr":     2, // French (février)
	"fevr":     2,
	"fév":      2,
	"fev":      2,
	"février":  2,
	"fevrier":  2,
	"febbraio": 2, // Italian
	"febrero":  2, // Spanish

	"mar":   3,
	"mär":   3,
	"maer":  3,
	"märz":  3,
	"maerz": 3,
	"march": 3,
	"mars":  3, // French
	"marzo": 3, // Italian, Spanish

	"apr":    4,
	"april":  4,
	"avr":    4, // French (avril)
	"avril":  4,
	"aprile": 4, // Italian
	"abr":    4, // Spanish (abril)
	"abril":  4,

	"aprij": 4,
	"apri1": 4,
	"aprjj": 4,
	"aprj1": 4,
	"apr1j": 4,
	"apr11": 4,

	"may":    5,
	"mai":    5,
	"ma1":    5,
	"maj":    5,
	"mag":    5, // Italian (maggio)
	"maggio": 5,
	"mayo":   5, // Spanish

	"jun":    6,
	"june":   6,
	"juni":   6,
	"junj":   6,
	"jun1":   6,
	"juin":   6, // French
	"giu":    6, // Italian (giugno)
	"giugno": 6,
	"junio":  6, // Spanish

	"iun":  6,
	"iune": 6,
	"iuni": 6,

	"1un":  6,
	"1une": 6,
	"1uni": 6,

	"jul":     7,
	"july":    7,
	"juli":    7,
	"julj":    7,
	"jul1":    7,
	"juil":    7, // French (juillet)
	"juill":   7,
	"juillet": 7,
	"lug":     7, // Italian (luglio)
	"luglio":  7,
	"julio":   7, // Spanish

	"iul":  7,
	"iuly": 7,
	"iuli": 7,
	"iulj": 7,
	"iul1": 7,

	"1ul":  7,
	"1uly": 7,
	"1uli": 7,
	"1ulj": 7,
	"1ul1": 7,

	"aug":    8,
	"august": 8,
	"août":   8, // French
	"aout":   8,
	"ago":    8, // Italian, Spanish (agosto)
	"agosto": 8,

	"sep":        9,
	"sept":       9,
	"september":  9,
	"septembre":  9, // French
	"set":        9, // Italian (settembre)
	"settembre":  9,
	"septiembre": 9, // Spanish
	"setiembre":  9,

	"sepl":      9,
	"seplember": 9,
	"sepj":      9,
	"sepjember": 9,
	"sep1":      9,
	"sep1ember": 9,

	"okt":     10,
	"oct":     10,
	"oktober": 10,
	"october": 10,
	"octobre": 10, // French
	"ott":     10, // Italian (ottobre)
	"ottobre": 10,
	"octubre": 10, // Spanish

	"nov":       11,
	"november":  11,
	"novembre":  11, // French, Italian
	"noviembre": 11, // Spanish

	"dec":       12,
	"dez":       12,
	"december":  12,
	"dezember":  12,
	"déc":       12, // French (décembre)
	"décembre":  12,
	"decembre":  12,
	"dic":       12, // Italian, Spanish (dicembre, diciembre)
	"dicembre":  12,
	"diciembre": 12,
}
