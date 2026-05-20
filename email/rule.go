package email

// Rule reports whether it applies to a given email Message.
type Rule interface {
	AppliesToMessage(msg *Message) bool
}

// RuleFunc is a function implementing the Rule interface.
type RuleFunc func(*Message) bool

// AppliesToMessage implements the Rule interface by calling f.
func (f RuleFunc) AppliesToMessage(msg *Message) bool {
	return f(msg)
}

// BoolRule is a constant Rule that always returns its own bool value.
type BoolRule bool

// AppliesToMessage implements the Rule interface
// by returning the bool value of r.
func (r BoolRule) AppliesToMessage(*Message) bool {
	return bool(r)
}

// AllRule is a Rule that applies only if all of its rules apply.
type AllRule []Rule

// AppliesToMessage implements the Rule interface by returning true
// only if every rule in r applies to the message.
func (r AllRule) AppliesToMessage(msg *Message) bool {
	for _, rule := range r {
		if !rule.AppliesToMessage(msg) {
			return false
		}
	}
	return true
}

// AnyRule is a Rule that applies if any of its rules apply.
type AnyRule []Rule

// AppliesToMessage implements the Rule interface by returning true
// if any rule in r applies to the message.
func (r AnyRule) AppliesToMessage(msg *Message) bool {
	for _, rule := range r {
		if rule.AppliesToMessage(msg) {
			return true
		}
	}
	return false
}
