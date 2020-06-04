package strfmt

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScan(t *testing.T) {
	type test struct {
		Int    int
		IntPtr *int
	}

	var (
		config     = NewScanConfig()
		int666 int = 666
	)

	type args struct {
		destField string
		source    string
		config    *ScanConfig
	}
	tests := []struct {
		name     string
		dest     test
		args     args
		wantErr  bool
		wantDest test
	}{
		{name: `666 Int`, dest: test{}, args: args{destField: "Int", source: "666", config: config}, wantErr: false, wantDest: test{Int: 666}},
		{name: `"" IntPtr`, dest: test{}, args: args{destField: "IntPtr", source: "", config: config}, wantErr: false, wantDest: test{}},
		{name: `null IntPtr`, dest: test{}, args: args{destField: "IntPtr", source: "null", config: config}, wantErr: false, wantDest: test{}},
		{name: `666 IntPtr`, dest: test{}, args: args{destField: "IntPtr", source: "666", config: config}, wantErr: false, wantDest: test{IntPtr: &int666}},
		// TODO
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dest := reflect.ValueOf(&tt.dest).Elem().FieldByName(tt.args.destField)
			err := Scan(dest, tt.args.source, tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.wantDest, tt.dest)
		})
	}
}
