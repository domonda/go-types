package email

import (
	"reflect"
	"testing"
)

func TestAddressSet_Scan(t *testing.T) {
	tests := []struct {
		name    string
		set     AddressSet
		value   interface{}
		want    AddressSet
		wantErr bool
	}{
		{
			name:  "SplitArray bug",
			value: "{some@example.com,if.need.of.a.''declaration.of.compliance''.please.contact.us@example.com}",
			want:  MakeAddressSet("some@example.com", "if.need.of.a.''declaration.of.compliance''.please.contact.us@example.com"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.set.Scan(tt.value); (err != nil) != tt.wantErr {
				t.Errorf("AddressSet.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.set, tt.want) {
				t.Errorf("AddressSet.Scan() = %v, want %v", tt.set, tt.want)
			}
		})
	}
}
