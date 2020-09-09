package strfmt

import (
	"testing"
	"time"
)

func TestScanConfig_ParseTime(t *testing.T) {
	config := NewScanConfig()
	tests := []struct {
		name   string
		str    string
		wantT  time.Time
		wantOK bool
	}{
		{name: "2019-02-28 23:13:52", str: "2019-02-28 23:13:52", wantT: time.Date(2019, 2, 28, 23, 13, 52, 0, time.UTC), wantOK: true},
		{name: "2019-02-28 23:13", str: "2019-02-28 23:13", wantT: time.Date(2019, 2, 28, 23, 13, 0, 0, time.UTC), wantOK: true},
		{name: "2019-02-28", str: "2019-02-28", wantT: time.Date(2019, 2, 28, 0, 0, 0, 0, time.UTC), wantOK: true},
		{name: "empty string", str: "", wantT: time.Time{}, wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotT, gotOk := config.ParseTime(tt.str)
			if !gotT.Equal(tt.wantT) {
				t.Errorf("ScanConfig.ParseTime() gotT = %v, want %v", gotT, tt.wantT)
			}
			if gotOk != tt.wantOK {
				t.Errorf("ScanConfig.ParseTime() gotOk = %v, want %v", gotOk, tt.wantOK)
			}
		})
	}
}
