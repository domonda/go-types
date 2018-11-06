package date

var monthNameMap = map[string]int{
	"jan":     1,
	"jän":     1,
	"jaen":    1,
	"januar":  1,
	"january": 1,
	"jänner":  1,
	"jaenner": 1,

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

	"mar":   3,
	"mär":   3,
	"maer":  3,
	"märz":  3,
	"maerz": 3,
	"march": 3,

	"apr":   4,
	"april": 4,

	"aprij": 4,
	"apri1": 4,
	"aprjj": 4,
	"aprj1": 4,
	"apr1j": 4,
	"apr11": 4,

	"may": 5,
	"mai": 5,
	"ma1": 5,
	"maj": 5,

	"jun":  6,
	"june": 6,
	"juni": 6,
	"junj": 6,
	"jun1": 6,

	"iun":  6,
	"iune": 6,
	"iuni": 6,

	"1un":  6,
	"1une": 6,
	"1uni": 6,

	"jul":  7,
	"july": 7,
	"juli": 7,
	"julj": 7,
	"jul1": 7,

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

	"sep":       9,
	"sept":      9,
	"september": 9,

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

	"nov":      11,
	"november": 11,

	"dec":      12,
	"dez":      12,
	"december": 12,
	"dezember": 12,
}
