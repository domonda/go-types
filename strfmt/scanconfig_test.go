package strfmt

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/nullable"
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

// TestScanConfig_SetTypeScannerWithoutNewScanConfig protects that a
// ScanConfig that was not created by NewScanConfig, like a struct literal
// or one decoded from JSON, which excludes TypeScanners, can register a
// Scanner instead of panicking on the nil TypeScanners map.
func TestScanConfig_SetTypeScannerWithoutNewScanConfig(t *testing.T) {
	type custom int

	var config ScanConfig
	assert.Nil(t, config.TypeScanners, "test needs a nil TypeScanners map")
	assert.NotPanics(t, func() {
		config.SetTypeScanner(reflect.TypeFor[custom](), ScannerFunc(func(dest reflect.Value, str string, config *ScanConfig) error {
			dest.SetInt(42)
			return nil
		}))
	})

	var dest custom
	assert.NoError(t, Scan(reflect.ValueOf(&dest).Elem(), "anything", &config))
	assert.Equal(t, custom(42), dest, "the registered Scanner is used")
}

// TestScanConfig_roundTripsFormatPresets protects that the default ScanConfig
// reads back what the FormatConfig presets write, so a CSV file or spreadsheet
// exported with NewFormatConfig, NewEnglishFormatConfig or NewGermanFormatConfig
// can be imported again with NewScanConfig. The German preset writes bools as
// "ja" and "nein" and both locale presets write times in layouts that the
// default TimeFormats did not contain, so those values failed to scan.
func TestScanConfig_roundTripsFormatPresets(t *testing.T) {
	someTime := time.Date(2024, 3, 5, 6, 7, 8, 0, time.UTC)
	values := []any{
		true,
		false,
		someTime,
		nullable.TimeFrom(someTime),
		1234.5,
		-0.125,
		1e-7,
		123456789.25,
		money.Amount(1234.5),
		money.Amount(-1234567.89),
		money.Amount(0.05),
		money.Amount(1000),
	}
	presets := []struct {
		name   string
		config *FormatConfig
	}{
		{name: "NewFormatConfig", config: NewFormatConfig()},
		{name: "NewEnglishFormatConfig", config: NewEnglishFormatConfig()},
		{name: "NewGermanFormatConfig", config: NewGermanFormatConfig()},
	}
	scanConfig := NewScanConfig()
	for _, preset := range presets {
		for _, value := range values {
			str := Format(value, preset.config)
			dest := reflect.New(reflect.TypeOf(value)).Elem()
			if !assert.NoError(t, Scan(dest, str, scanConfig), "%s: Scan(%T, %q)", preset.name, value, str) {
				continue
			}
			switch want := value.(type) {
			case time.Time:
				assert.True(t, want.Equal(dest.Interface().(time.Time)), "%s: Scan(%T, %q) = %v", preset.name, value, str, dest.Interface())
			case nullable.Time:
				assert.True(t, want.Get().Equal(dest.Interface().(nullable.Time).Get()), "%s: Scan(%T, %q) = %v", preset.name, value, str, dest.Interface())
			default:
				assert.Equal(t, value, dest.Interface(), "%s: Scan(%T, %q)", preset.name, value, str)
			}
		}
	}
}
