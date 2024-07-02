package strutil

import (
	"reflect"
	"testing"
)

func TestTrimSpace(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		s    string
		want string
	}{
		{s: "\u200b\tZERO WIDTH SPACE\r\n", want: "ZERO WIDTH SPACE"},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			if got := TrimSpace(tt.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TrimSpace(%#v) = %#v, want %#v", tt.s, got, tt.want)
			}
		})
	}
}
