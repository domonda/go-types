package strfmt

import "github.com/domonda/go-types/language"

type Detector struct {
	normalizers map[string]Normalizer
}

func NewDetector() *Detector {
	return &Detector{
		normalizers: make(map[string]Normalizer),
	}
}

func (td *Detector) Register(name string, normalizer Normalizer) {
	td.normalizers[name] = normalizer
}

func (td *Detector) Detect(str string, langHints ...language.Code) map[string]string {
	detected := make(map[string]string)
	for name, normalizer := range td.normalizers {
		normalized, err := normalizer.Normalize(str, langHints...)
		if err == nil {
			detected[name] = normalized
		}
	}
	return detected
}
