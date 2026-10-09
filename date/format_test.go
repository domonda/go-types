package date

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/domonda/go-types/language"
	"github.com/domonda/go-types/nullable"
)

func TestFormat_Format(t *testing.T) {
	f := &Format{Layout: "02.01.2006"}
	assert.Equal(t, "25.12.2023", f.Format("2023-12-25"))
	// An empty Date has nothing to format and must not produce a
	// fake "01.01.0001" from the zero time.Time.
	assert.Equal(t, "", f.Format(""))
}

func TestFormat_Parse(t *testing.T) {
	// Parse must accept the many real-world input shapes understood
	// by Normalize and re-emit them in the configured layout,
	// so that e.g. a German invoice date ends up in a uniform format.
	f := &Format{Layout: "02.01.2006"}

	tests := []struct {
		name  string
		input string
		lang  []language.Code
		want  string
	}{
		{name: "ISO", input: "2023-12-25", want: "25.12.2023"},
		{name: "German dotted", input: "25.12.2023", want: "25.12.2023"},
		{name: "German month name", input: "1. Dezember 2016", want: "01.12.2016"},
		{name: "US slashes", input: "12/25/2023", want: "25.12.2023"},
		{name: "ambiguous dd/mm with EN hint is month first", input: "05/06/2024", lang: []language.Code{language.EN}, want: "06.05.2024"},
		{name: "ambiguous dd.mm with DE hint is day first", input: "05.06.2024", lang: []language.Code{language.DE}, want: "05.06.2024"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := f.Parse(tt.input, tt.lang...)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("invalid returns empty string and error", func(t *testing.T) {
		got, err := f.Parse("not a date")
		require.Error(t, err)
		assert.Equal(t, "", got)
	})
}

func TestFormat_AssignString(t *testing.T) {
	// Layout "" means Normalize is used to parse the many supported
	// date formats instead of a fixed time layout.
	normalizing := &Format{}
	fixedLayout := &Format{Layout: "02.01.2006"}

	t.Run("Date", func(t *testing.T) {
		var d Date
		require.NoError(t, normalizing.AssignString(reflect.ValueOf(&d).Elem(), " 25.12.2023 "))
		assert.Equal(t, Date("2023-12-25"), d, "surrounding whitespace is trimmed")

		require.NoError(t, fixedLayout.AssignString(reflect.ValueOf(&d).Elem(), "24.12.2023"))
		assert.Equal(t, Date("2023-12-24"), d)

		require.NoError(t, normalizing.AssignString(reflect.ValueOf(&d).Elem(), ""))
		assert.Equal(t, Date(""), d, "empty source resets the Date")
	})

	t.Run("NullableDate", func(t *testing.T) {
		var n NullableDate
		require.NoError(t, normalizing.AssignString(reflect.ValueOf(&n).Elem(), "Dec 25, 2023"))
		assert.Equal(t, NullableDate("2023-12-25"), n)

		require.NoError(t, normalizing.AssignString(reflect.ValueOf(&n).Elem(), "   "))
		assert.Equal(t, Null, n, "blank source must set NULL")
	})

	t.Run("time.Time", func(t *testing.T) {
		var tm time.Time
		require.NoError(t, fixedLayout.AssignString(reflect.ValueOf(&tm).Elem(), "25.12.2023"))
		assert.Equal(t, time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC), tm)

		// Without a layout the date is interpreted as local midnight
		require.NoError(t, normalizing.AssignString(reflect.ValueOf(&tm).Elem(), "2023-12-25"))
		assert.Equal(t, time.Date(2023, 12, 25, 0, 0, 0, 0, time.Local), tm)

		require.NoError(t, normalizing.AssignString(reflect.ValueOf(&tm).Elem(), ""))
		assert.True(t, tm.IsZero(), "empty source resets the time")
	})

	t.Run("*time.Time", func(t *testing.T) {
		var tm *time.Time
		require.NoError(t, fixedLayout.AssignString(reflect.ValueOf(&tm).Elem(), "25.12.2023"))
		require.NotNil(t, tm)
		assert.Equal(t, time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC), *tm)
	})

	t.Run("nullable.Time", func(t *testing.T) {
		var nt nullable.Time
		require.NoError(t, fixedLayout.AssignString(reflect.ValueOf(&nt).Elem(), "25.12.2023"))
		assert.Equal(t, time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC), nt.Get())

		require.NoError(t, fixedLayout.AssignString(reflect.ValueOf(&nt).Elem(), ""))
		assert.True(t, nt.IsNull(), "empty source must set NULL")
	})

	t.Run("errors", func(t *testing.T) {
		var d Date
		dest := reflect.ValueOf(&d).Elem()
		assert.Error(t, normalizing.AssignString(dest, "not a date"), "unparseable with Normalize")
		assert.Error(t, fixedLayout.AssignString(dest, "2023-12-25"), "does not match the fixed layout")
		// The zero time.Time is used as "no value", so a parsed
		// zero date would be indistinguishable from an empty one.
		assert.Error(t, fixedLayout.AssignString(dest, "01.01.0001"), "zero time can't be assigned")

		var s string
		assert.Error(t, normalizing.AssignString(reflect.ValueOf(&s).Elem(), "2023-12-25"), "unsupported destination type")
	})
}

func TestFormat_FormatString(t *testing.T) {
	f := &Format{Layout: "02.01.2006", NilString: "<nil>", ZeroString: "<zero>"}

	date := Date("2023-12-25")
	var nilDatePtr *Date

	tests := []struct {
		name string
		val  reflect.Value
		want string
	}{
		{name: "Date", val: reflect.ValueOf(date), want: "25.12.2023"},
		{name: "*Date", val: reflect.ValueOf(&date), want: "25.12.2023"},
		{name: "NullableDate", val: reflect.ValueOf(NullableDate("2023-12-25")), want: "25.12.2023"},
		{name: "time.Time", val: reflect.ValueOf(time.Date(2023, 12, 25, 13, 14, 15, 0, time.UTC)), want: "25.12.2023"},
		{name: "zero Date", val: reflect.ValueOf(Date("")), want: "<zero>"},
		{name: "zero time.Time", val: reflect.ValueOf(time.Time{}), want: "<zero>"},
		{name: "nil *Date", val: reflect.ValueOf(nilDatePtr), want: "<nil>"},
		{name: "invalid reflect.Value", val: reflect.Value{}, want: "<nil>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := f.FormatString(tt.val)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("unsupported type", func(t *testing.T) {
		_, err := f.FormatString(reflect.ValueOf(42))
		assert.Error(t, err)
	})
}
