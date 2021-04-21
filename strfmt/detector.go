package strfmt

// import "github.com/domonda/go-types/language"

// type Detector struct {
// 	parsers map[string]Parser
// }

// func NewDetector() *Detector {
// 	return &Detector{
// 		parsers: make(map[string]Parser),
// 	}
// }

// func (td *Detector) Register(name string, parser Parser) {
// 	td.parsers[name] = parser
// }

// func (td *Detector) Detect(str string, langHints ...language.Code) map[string]string {
// 	detected := make(map[string]string)
// 	for name, parser := range td.parsers {
// 		normalized, err := parser.Parse(str, langHints...)
// 		if err == nil {
// 			detected[name] = normalized
// 		}
// 	}
// 	return detected
// }

// type Parser interface {
// 	// Parse str using optional language hints and
// 	// returns a normalized version of str or an parsing error.
// 	Parse(str string, langHints ...language.Code) (normalized string, err error)
// }
