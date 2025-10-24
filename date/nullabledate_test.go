package date

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NullableDate(t *testing.T) {
	var n NullableDate
	assert.True(t, n.Valid(), "empty NullableDate is valid")
	assert.NoError(t, n.Validate(), "empty NullableDate is valid")

	n = "0001-01-01"
	assert.True(t, n.Valid(), "empty NullableDate is valid")
	assert.NoError(t, n.Validate(), "empty NullableDate is valid")

	assert.Empty(t, n.NormalizedOrNull())
}

func Test_NullableDate_UnmarshalJSON(t *testing.T) {
	sourceJSON := `{
		"empty": "",
		"null": null,
		"notNull": "2012-12-12",
		"invalid": "Not a date!"
	}`
	s := struct {
		Empty   NullableDate `json:"empty"`
		Null    NullableDate `json:"null"`
		NotNull NullableDate `json:"notNull"`
		Invalid NullableDate `json:"invalid"`
	}{}
	err := json.Unmarshal([]byte(sourceJSON), &s)
	assert.NoError(t, err, "json.Unmarshal")
	assert.Equal(t, Null, s.Empty, "empty JSON string is Null")
	assert.Equal(t, Null, s.Null, "JSON null value as Null")
	assert.Equal(t, NullableDate("2012-12-12"), s.NotNull, "valid NullableDate")
	assert.Equal(t, NullableDate("Not a date!"), s.Invalid, "invalid NullableDate parsed as is, without error")
	assert.False(t, s.Invalid.Valid(), "invalid NullableDate parsed as is, not valid")
}
func TestNullableDate_IsNull(t *testing.T) {
	assert.True(t, NullableDate("").IsNull())
	assert.True(t, Null.IsNull())
	assert.False(t, NullableDate("2023-12-25").IsNull())
}

func TestNullableDate_IsNotNull(t *testing.T) {
	assert.False(t, NullableDate("").IsNotNull())
	assert.False(t, Null.IsNotNull())
	assert.True(t, NullableDate("2023-12-25").IsNotNull())
}

func TestNullableDate_IsZero(t *testing.T) {
	assert.True(t, NullableDate("").IsZero())
	assert.True(t, NullableDate("0000-00-00").IsZero())
	assert.True(t, NullableDate("0001-01-01").IsZero())
	assert.False(t, NullableDate("2023-12-25").IsZero())
}

func TestNullableDate_Valid(t *testing.T) {
	assert.True(t, Null.Valid())
	assert.True(t, NullableDate("2023-12-25").Valid())
	assert.False(t, NullableDate("invalid").Valid())
}

func TestNullableDate_ValidAndNotNull(t *testing.T) {
	assert.False(t, Null.ValidAndNotNull())
	assert.True(t, NullableDate("2023-12-25").ValidAndNotNull())
	assert.False(t, NullableDate("invalid").ValidAndNotNull())
}

func TestNullableDate_ValidAndNormalized(t *testing.T) {
	assert.True(t, Null.ValidAndNormalized()) // null normalizes to null (no change)
	assert.True(t, NullableDate("2023-12-25").ValidAndNormalized())
	assert.False(t, NullableDate("25.12.2023").ValidAndNormalized())
}

func TestNullableDate_Get(t *testing.T) {
	d := NullableDate("2023-12-25")
	assert.Equal(t, Date("2023-12-25"), d.Get())
	
	assert.Panics(t, func() {
		Null.Get()
	})
}

func TestNullableDate_GetOr(t *testing.T) {
	d := NullableDate("2023-12-25")
	assert.Equal(t, Date("2023-12-25"), d.GetOr("2020-01-01"))
	assert.Equal(t, Date("2020-01-01"), Null.GetOr("2020-01-01"))
}

func TestNullableDate_Date(t *testing.T) {
	d := NullableDate("2023-12-25")
	assert.Equal(t, Date("2023-12-25"), d.Date())
	assert.Equal(t, Date(""), Null.Date())
}

func TestNullableDate_DateOr(t *testing.T) {
	d := NullableDate("2023-12-25")
	assert.Equal(t, Date("2023-12-25"), d.DateOr("2020-01-01"))
	assert.Equal(t, Date("2020-01-01"), Null.DateOr("2020-01-01"))
}

func TestNullableDate_String(t *testing.T) {
	assert.Equal(t, "2023-12-25", NullableDate("2023-12-25").String())
	assert.Equal(t, "", Null.String())
}

func TestNullableDate_StringOr(t *testing.T) {
	d := NullableDate("2023-12-25")
	assert.Equal(t, "2023-12-25", d.StringOr("N/A"))
	assert.Equal(t, "N/A", Null.StringOr("N/A"))
}

func TestNullableDate_Year(t *testing.T) {
	assert.Equal(t, 2023, NullableDate("2023-12-25").Year())
	assert.Equal(t, 0, Null.Year())
}

func TestNullableDate_Month(t *testing.T) {
	assert.Equal(t, 12, int(NullableDate("2023-12-25").Month()))
	assert.Equal(t, 0, int(Null.Month()))
}

func TestNullableDate_Day(t *testing.T) {
	assert.Equal(t, 25, NullableDate("2023-12-25").Day())
	assert.Equal(t, 0, Null.Day())
}

func TestNullableDate_YearMonthDay(t *testing.T) {
	y, m, d := NullableDate("2023-12-25").YearMonthDay()
	assert.Equal(t, 2023, y)
	assert.Equal(t, 12, int(m))
	assert.Equal(t, 25, d)
	
	y, m, d = Null.YearMonthDay()
	assert.Equal(t, 0, y)
	assert.Equal(t, 0, int(m))
	assert.Equal(t, 0, d)
}

func TestNullableDate_Weekday(t *testing.T) {
	// 2023-12-25 is Monday
	assert.Equal(t, 1, int(NullableDate("2023-12-25").Weekday()))
	assert.Equal(t, 0, int(Null.Weekday()))
}

func TestNullableDate_ISOWeek(t *testing.T) {
	y, w := NullableDate("2023-12-25").ISOWeek()
	assert.Equal(t, 2023, y)
	assert.Equal(t, 52, w)
	
	y, w = Null.ISOWeek()
	assert.Equal(t, 0, y)
	assert.Equal(t, 0, w)
}

func TestNullableDate_Format(t *testing.T) {
	d := NullableDate("2023-12-25")
	assert.Equal(t, "2023-12-25", d.Format(Layout))
	assert.Equal(t, "", Null.Format(Layout))
}

func TestNullableDate_Compare(t *testing.T) {
	d1 := NullableDate("2023-12-25")
	d2 := NullableDate("2023-12-26")
	
	assert.Equal(t, -1, d1.Compare(d2))
	assert.Equal(t, 1, d2.Compare(d1))
	assert.Equal(t, 0, d1.Compare(d1))
	assert.Equal(t, 0, Null.Compare(Null))
}

func TestNullableDate_After(t *testing.T) {
	d1 := NullableDate("2023-12-25")
	d2 := NullableDate("2023-12-26")
	
	assert.False(t, d1.After(d2))
	assert.True(t, d2.After(d1))
	assert.False(t, Null.After(d1))
}

func TestNullableDate_Before(t *testing.T) {
	d1 := NullableDate("2023-12-25")
	d2 := NullableDate("2023-12-26")

	assert.True(t, d1.Before(d2))
	assert.False(t, d2.Before(d1))
	assert.True(t, Null.Before(d1)) // null is always before any non-null date
	assert.False(t, Null.Before(Null)) // null is not before null
}

func TestNullableDate_EqualOrAfter(t *testing.T) {
	d1 := NullableDate("2023-12-25")
	d2 := NullableDate("2023-12-26")
	
	assert.False(t, d1.EqualOrAfter(d2))
	assert.True(t, d2.EqualOrAfter(d1))
	assert.True(t, d1.EqualOrAfter(d1))
}

func TestNullableDate_EqualOrBefore(t *testing.T) {
	d1 := NullableDate("2023-12-25")
	d2 := NullableDate("2023-12-26")
	
	assert.True(t, d1.EqualOrBefore(d2))
	assert.False(t, d2.EqualOrBefore(d1))
	assert.True(t, d1.EqualOrBefore(d1))
}

func TestNullableDate_AddYears(t *testing.T) {
	d := NullableDate("2023-12-25")
	assert.Equal(t, NullableDate("2024-12-25"), d.AddYears(1))
	assert.Equal(t, NullableDate("2022-12-25"), d.AddYears(-1))
	assert.Equal(t, Null, Null.AddYears(1))
}

func TestNullableDate_AddMonths(t *testing.T) {
	d := NullableDate("2023-12-25")
	assert.Equal(t, NullableDate("2024-01-25"), d.AddMonths(1))
	assert.Equal(t, NullableDate("2023-11-25"), d.AddMonths(-1))
	assert.Equal(t, Null, Null.AddMonths(1))
}

func TestNullableDate_AddDays(t *testing.T) {
	d := NullableDate("2023-12-25")
	assert.Equal(t, NullableDate("2023-12-26"), d.AddDays(1))
	assert.Equal(t, NullableDate("2023-12-24"), d.AddDays(-1))
	assert.Equal(t, Null, Null.AddDays(1))
}

func TestNullableDate_MidnightUTC(t *testing.T) {
	d := NullableDate("2023-12-25")
	tm := d.MidnightUTC()
	assert.False(t, tm.IsNull())
	assert.Equal(t, 2023, tm.Get().Year())
	
	assert.True(t, Null.MidnightUTC().IsNull())
}

func TestNullableDate_Midnight(t *testing.T) {
	d := NullableDate("2023-12-25")
	tm := d.Midnight()
	assert.False(t, tm.IsNull())
	
	assert.True(t, Null.Midnight().IsNull())
}

func TestNullableDate_BeginningOfWeek(t *testing.T) {
	d := NullableDate("2023-12-27") // Wednesday
	beginning := d.BeginningOfWeek(1) // Monday
	assert.Equal(t, NullableDate("2023-12-25"), beginning)
	
	assert.Equal(t, Null, Null.BeginningOfWeek(1))
}

func TestNullableDate_BeginningOfMonth(t *testing.T) {
	d := NullableDate("2023-12-25")
	assert.Equal(t, NullableDate("2023-12-01"), d.BeginningOfMonth())
	assert.Equal(t, Null, Null.BeginningOfMonth())
}

func TestNullableDate_BeginningOfQuarter(t *testing.T) {
	d := NullableDate("2023-11-15")
	assert.Equal(t, NullableDate("2023-10-01"), d.BeginningOfQuarter())
	assert.Equal(t, Null, Null.BeginningOfQuarter())
}

func TestNullableDate_BeginningOfYear(t *testing.T) {
	d := NullableDate("2023-12-25")
	assert.Equal(t, NullableDate("2023-01-01"), d.BeginningOfYear())
	assert.Equal(t, Null, Null.BeginningOfYear())
}

func TestNullableDate_EndOfWeek(t *testing.T) {
	d := NullableDate("2023-12-25") // Monday
	end := d.EndOfWeek(1) // Monday as week start
	assert.Equal(t, NullableDate("2023-12-31"), end)
	
	assert.Equal(t, Null, Null.EndOfWeek(1))
}

func TestNullableDate_EndOfMonth(t *testing.T) {
	d := NullableDate("2023-12-25")
	assert.Equal(t, NullableDate("2023-12-31"), d.EndOfMonth())
	assert.Equal(t, Null, Null.EndOfMonth())
}

func TestNullableDate_EndOfQuarter(t *testing.T) {
	d := NullableDate("2023-11-15")
	assert.Equal(t, NullableDate("2023-12-31"), d.EndOfQuarter())
	assert.Equal(t, Null, Null.EndOfQuarter())
}

func TestNullableDate_EndOfYear(t *testing.T) {
	d := NullableDate("2023-06-15")
	assert.Equal(t, NullableDate("2023-12-31"), d.EndOfYear())
	assert.Equal(t, Null, Null.EndOfYear())
}

func TestNullableDate_LastMonday(t *testing.T) {
	d := NullableDate("2023-12-27") // Wednesday
	monday := d.LastMonday()
	assert.Equal(t, NullableDate("2023-12-25"), monday)
	assert.Equal(t, Null, Null.LastMonday())
}

func TestNullableDate_NextSunday(t *testing.T) {
	d := NullableDate("2023-12-25") // Monday
	sunday := d.NextSunday()
	assert.Equal(t, NullableDate("2023-12-31"), sunday)
	assert.Equal(t, Null, Null.NextSunday())
}

func TestNullableDate_IsToday(t *testing.T) {
	today := NullableDate(OfToday())
	assert.True(t, today.IsToday())
	assert.False(t, NullableDate("2020-01-01").IsToday())
	assert.False(t, Null.IsToday())
}

func TestNullableDate_IsTodayInUTC(t *testing.T) {
	today := NullableDate(OfNowInUTC())
	assert.True(t, today.IsTodayInUTC())
	assert.False(t, Null.IsTodayInUTC())
}

func TestNullableDate_AfterToday(t *testing.T) {
	tomorrow := NullableDate(OfToday().AddDays(1))
	yesterday := NullableDate(OfToday().AddDays(-1))
	
	assert.True(t, tomorrow.AfterToday())
	assert.False(t, yesterday.AfterToday())
	assert.False(t, Null.AfterToday())
}

func TestNullableDate_BeforeToday(t *testing.T) {
	tomorrow := NullableDate(OfToday().AddDays(1))
	yesterday := NullableDate(OfToday().AddDays(-1))

	assert.False(t, tomorrow.BeforeToday())
	assert.True(t, yesterday.BeforeToday())
	assert.True(t, Null.BeforeToday()) // null date is always before today
}

func TestNullableDate_Normalized(t *testing.T) {
	d := NullableDate("25.12.2023")
	normalized, err := d.Normalized()
	assert.NoError(t, err)
	assert.Equal(t, NullableDate("2023-12-25"), normalized)
	
	normalized, err = Null.Normalized()
	assert.NoError(t, err)
	assert.Equal(t, Null, normalized)
}

func TestNullableDate_NormalizedOrUnchanged(t *testing.T) {
	d := NullableDate("25.12.2023")
	assert.Equal(t, NullableDate("2023-12-25"), d.NormalizedOrUnchanged())
	assert.Equal(t, Null, Null.NormalizedOrUnchanged())
}

func TestNullableDate_NormalizedOrNull(t *testing.T) {
	d := NullableDate("25.12.2023")
	assert.Equal(t, NullableDate("2023-12-25"), d.NormalizedOrNull())
	assert.Equal(t, Null, Null.NormalizedOrNull())
	assert.Equal(t, Null, NullableDate("invalid").NormalizedOrNull())
}

func TestNullableDate_NormalizedEqual(t *testing.T) {
	d1 := NullableDate("2023-12-25")
	d2 := NullableDate("25.12.2023")
	
	assert.True(t, d1.NormalizedEqual(d2))
	assert.True(t, Null.NormalizedEqual(Null))
	assert.False(t, d1.NormalizedEqual(Null))
}

func TestNullableDate_Value(t *testing.T) {
	d := NullableDate("2023-12-25")
	val, err := d.Value()
	assert.NoError(t, err)
	assert.Equal(t, "2023-12-25", val)
	
	val, err = Null.Value()
	assert.NoError(t, err)
	assert.Nil(t, val)
}
