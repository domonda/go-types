package money

import "testing"

func TestParseRate(t *testing.T) {
	type args struct {
		str              string
		acceptedDecimals []int
	}
	tests := []struct {
		name    string
		args    args
		want    Rate
		wantErr bool
	}{
		{name: "15", args: args{str: "15"}, want: 15},
		{name: "15%", args: args{str: "15%"}, want: 0.15},
		{name: "15 %", args: args{str: "15 %"}, want: 0.15},
		{name: " 15 % ", args: args{str: " 15 % "}, want: 0.15},
		{name: "8,382.00", args: args{str: "8,382.00"}, want: 8382},
		// TODO more cases
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRate(tt.args.str, tt.args.acceptedDecimals...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseRate() = %v, want %v", got, tt.want)
			}
		})
	}
}
