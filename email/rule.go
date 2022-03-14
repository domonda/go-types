package email

type Rule interface {
	AppliesToMessage(msg *Message) bool
}

type AllRule []Rule

func (all AllRule) AppliesToMessage(msg *Message) bool {
	for _, rule := range all {
		if !rule.AppliesToMessage(msg) {
			return false
		}
	}
	return true
}

type AnyRule []Rule

func (any AnyRule) AppliesToMessage(msg *Message) bool {
	for _, rule := range any {
		if rule.AppliesToMessage(msg) {
			return true
		}
	}
	return false
}
