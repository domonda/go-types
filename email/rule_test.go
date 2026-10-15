package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRules protects the boolean semantics used to compose message
// filters: an empty AllRule applies (nothing to violate) while an empty
// AnyRule does not apply (nothing matched), like the vacuous truth of
// "for all" and "exists".
func TestRules(t *testing.T) {
	fromDomonda := RuleFunc(func(msg *Message) bool {
		return msg.From.DomainPart() == "domonda.com"
	})
	hasSubject := RuleFunc(func(msg *Message) bool {
		return msg.Subject != ""
	})
	msg := &Message{From: "erik@domonda.com", Subject: ""}

	tests := []struct {
		name string
		rule Rule
		want bool
	}{
		{name: "RuleFunc receives the message", rule: fromDomonda, want: true},
		{name: "RuleFunc false", rule: hasSubject, want: false},
		{name: "BoolRule true", rule: BoolRule(true), want: true},
		{name: "BoolRule false", rule: BoolRule(false), want: false},
		{name: "empty AllRule", rule: AllRule{}, want: true},
		{name: "AllRule all apply", rule: AllRule{fromDomonda, BoolRule(true)}, want: true},
		{name: "AllRule one does not apply", rule: AllRule{fromDomonda, hasSubject}, want: false},
		{name: "empty AnyRule", rule: AnyRule{}, want: false},
		{name: "AnyRule one applies", rule: AnyRule{hasSubject, fromDomonda}, want: true},
		{name: "AnyRule none applies", rule: AnyRule{hasSubject, BoolRule(false)}, want: false},
		{name: "nested", rule: AnyRule{AllRule{fromDomonda, hasSubject}, AllRule{fromDomonda, BoolRule(true)}}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.rule.AppliesToMessage(msg))
		})
	}
}
