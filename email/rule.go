package email

type Rule interface {
	AppliesToMessage(msg *Message) bool
}

type BoolRule bool

func (r BoolRule) AppliesToMessage(*Message) bool {
	return bool(r)
}

type AllRule []Rule

func (r AllRule) AppliesToMessage(msg *Message) bool {
	for _, rule := range r {
		if !rule.AppliesToMessage(msg) {
			return false
		}
	}
	return true
}

type AnyRule []Rule

func (r AnyRule) AppliesToMessage(msg *Message) bool {
	for _, rule := range r {
		if rule.AppliesToMessage(msg) {
			return true
		}
	}
	return false
}
