package date

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/domonda/go-types/language"
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

func TestNullableDate_UnmarshalJSONResetsAndRejects(t *testing.T) {
	type doc struct {
		Paid NullableDate `json:"paid"`
	}
	// JSON null must reset a field that holds a value, like for
	// uu.NullableID and email.NullableAddress, else decoding a partial
	// update with null into an existing struct keeps the old date.
	d := doc{Paid: "2024-01-02"}
	require.NoError(t, json.Unmarshal([]byte(`{"paid":null}`), &d))
	assert.Equal(t, Null, d.Paid)

	// A JSON string is assigned unchanged, without normalization or
	// validation, like encoding/json did before NullableDate had
	// UnmarshalJSON (see Test_NullableDate_UnmarshalJSON).
	for _, str := range []string{"2023-12-25", "25.12.2023", "bad", ""} {
		d := doc{Paid: "2024-01-02"}
		data, err := json.Marshal(map[string]string{"paid": str})
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(data, &d))
		assert.Equal(t, NullableDate(str), d.Paid, "json.Unmarshal(%s)", data)
	}

	// Any other JSON value is an error and leaves the value unchanged.
	for _, data := range []string{`20231225`, `true`, `{}`, `["2023-12-25"]`} {
		n := NullableDate("2024-01-02")
		assert.Error(t, json.Unmarshal([]byte(data), &n), "json.Unmarshal(%s)", data)
		assert.Equal(t, NullableDate("2024-01-02"), n, "json.Unmarshal(%s)", data)
	}
}
func TestNullableDate_IsNull(t *testing.T) {
	assert.True(t, NullableDate("").IsNull())
	assert.True(t, Null.IsNull())
	assert.False(t, NullableDate("2023-12-25").IsNull())
}

func TestNullableDate_WhitespaceIsNull(t *testing.T) {
	// Like country.NullableCode and vat.NullableID, a whitespace only
	// value is null, else a value that carries no date would be handled
	// as an invalid date: written to JSON as whitespace and failing
	// Value, Scan and validation.
	for _, n := range []NullableDate{" ", "  ", "\t\n", " "} {
		t.Run(fmt.Sprintf("%q", string(n)), func(t *testing.T) {
			assert.True(t, n.IsNull())
			assert.False(t, n.IsNotNull())
			assert.True(t, n.Valid())
			assert.NoError(t, n.Validate())
			assert.False(t, n.ValidAndNotNull())
			assert.False(t, n.ValidAndNormalized(), "whitespace is null, but not the normalized Null")

			norm, err := n.Normalized()
			require.NoError(t, err)
			assert.Equal(t, Null, norm)
			assert.Equal(t, Null, n.NormalizedOrNull())
			assert.True(t, n.NormalizedEqual(Null))

			assert.Equal(t, "", n.String())
			assert.Equal(t, "N/A", n.StringOr("N/A"))
			assert.Equal(t, Date("2020-01-01"), n.GetOr("2020-01-01"))
			assert.Equal(t, Date("2020-01-01"), n.DateOr("2020-01-01"))
			assert.Panics(t, func() { n.Get() })

			v, err := n.Value()
			require.NoError(t, err)
			assert.Nil(t, v, "must be written as SQL NULL instead of an error")
			j, err := json.Marshal(n)
			require.NoError(t, err)
			assert.Equal(t, `null`, string(j))

			// Reading whitespace gives Null like reading "" does
			scanned := NullableDate("2000-01-01")
			require.NoError(t, scanned.Scan(string(n)))
			assert.Equal(t, Null, scanned, "Scan")
			scanned = NullableDate("2000-01-01")
			require.NoError(t, scanned.ScanString(string(n), true))
			assert.Equal(t, Null, scanned, "ScanString")
			scanned = NullableDate("2000-01-01")
			_, _, err = scanned.ScanStringWithLang(string(n), language.EN)
			require.NoError(t, err)
			assert.Equal(t, Null, scanned, "ScanStringWithLang")
		})
	}
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
	// The zero dates are null (IsNull, Value writes SQL NULL), so a caller
	// that uses ValidAndNotNull to require a date must not accept them.
	assert.True(t, NullableDate("0001-01-01").IsNull())
	assert.False(t, NullableDate("0001-01-01").ValidAndNotNull())
	assert.False(t, NullableDate("0000-00-00").ValidAndNotNull())
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
	// Compare is a sort order, so an invalid date is ordered like Null
	// (see TestDate_Compare) and not by its raw string.
	assert.Equal(t, 0, NullableDate("zzz").Compare(Null))
	assert.Equal(t, 0, NullableDate("bad").Compare("worse"))
	assert.Equal(t, -1, NullableDate("zzz").Compare(d1))
	assert.Equal(t, 1, d1.Compare("zzz"))
}

func TestNullableDate_After(t *testing.T) {
	d1 := NullableDate("2023-12-25")
	d2 := NullableDate("2023-12-26")

	assert.False(t, d1.After(d2))
	assert.True(t, d2.After(d1))
	assert.False(t, Null.After(d1))
	assert.True(t, d1.After(Null)) // any date is after null
}

func TestNullableDate_Before(t *testing.T) {
	d1 := NullableDate("2023-12-25")
	d2 := NullableDate("2023-12-26")

	assert.True(t, d1.Before(d2))
	assert.False(t, d2.Before(d1))
	assert.True(t, Null.Before(d1))    // null is always before any non-null date
	assert.False(t, Null.Before(Null)) // null is not before null
}

func TestNullableDate_NullOrdering(t *testing.T) {
	// Null is always before any date (see Compare and Before),
	// so After must be the mirror of Before: sorting or range checks
	// must not depend on which operand a null value is passed as.
	d := NullableDate("2023-12-25")

	assert.True(t, Null.Before(d), "Null.Before(d)")
	assert.True(t, d.After(Null), "d.After(Null)")
	assert.False(t, d.Before(Null), "d.Before(Null)")
	assert.False(t, Null.After(d), "Null.After(d)")

	assert.True(t, d.EqualOrAfter(Null), "d.EqualOrAfter(Null)")
	assert.True(t, Null.EqualOrBefore(d), "Null.EqualOrBefore(d)")
	assert.False(t, Null.EqualOrAfter(d), "Null.EqualOrAfter(d)")
	assert.False(t, d.EqualOrBefore(Null), "d.EqualOrBefore(Null)")

	// Null is neither before nor after itself
	assert.False(t, Null.Before(Null), "Null.Before(Null)")
	assert.False(t, Null.After(Null), "Null.After(Null)")
	assert.True(t, Null.EqualOrAfter(Null), "Null.EqualOrAfter(Null)")
	assert.True(t, Null.EqualOrBefore(Null), "Null.EqualOrBefore(Null)")

	// Consistent with Compare
	assert.Equal(t, -1, Null.Compare(d))
	assert.Equal(t, 1, d.Compare(Null))
}

func TestNullableDate_NonNormalizedValidInput(t *testing.T) {
	// A valid NullableDate that is not in the normalized YYYY-MM-DD form
	// must behave like its normalized form, like Date does,
	// instead of being treated like an invalid date (or panicking in After).
	n := NullableDate("25.12.2023")
	norm := NullableDate("2023-12-25")

	assert.Equal(t, norm.MidnightUTC(), n.MidnightUTC(), "MidnightUTC")
	assert.Equal(t, norm.Midnight(), n.Midnight(), "Midnight")
	assert.Equal(t, "25.12.2023", n.Format("02.01.2006"), "Format")
	assert.Equal(t, "2023-12-25", n.Format(Layout), "Format(Layout)")
	assert.True(t, n.After("2023-12-24"), "After")
	assert.True(t, NullableDate("2023-12-26").After(n), "After non-normalized other")
	assert.True(t, n.Before("2023-12-26"), "Before")
	assert.True(t, n.AfterTime(time.Date(2023, 12, 24, 12, 0, 0, 0, time.UTC)), "AfterTime")
	assert.True(t, n.BeforeTime(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC)), "BeforeTime")
	assert.Equal(t, NullableDate("2023-12-26"), n.AddDays(1), "AddDays")
	assert.Equal(t, time.Monday, n.Weekday(), "Weekday")

	// Invalid dates still have no time and are not after anything
	assert.True(t, NullableDate("invalid").MidnightUTC().IsNull(), "invalid MidnightUTC")
	assert.False(t, NullableDate("invalid").After("2023-12-24"), "invalid After")
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
	d := NullableDate("2023-12-27")   // Wednesday
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
	end := d.EndOfWeek(1)           // Monday as week start
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
	// A valid but non-normalized today is still today
	assert.True(t, NullableDate(today.Format("02.01.2006")).IsToday())
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

	// The normalization contract of the module: on error the receiver
	// comes back unchanged. An invalid NullableDate must in particular
	// not come back as Null, which would turn invalid input into a valid
	// "no date" for a caller that keeps the result.
	for _, n := range []NullableDate{"foo", "2020-13-01"} {
		normalized, err = n.Normalized()
		assert.Error(t, err, "NullableDate(%q).Normalized()", string(n))
		assert.Equal(t, n, normalized, "NullableDate(%q).Normalized()", string(n))
	}
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
	// Zero dates are null, so they normalize to and equal Null
	assert.True(t, NullableDate("0001-01-01").NormalizedEqual(Null))
	// An invalid date has no normalized value, so it must not be equal
	// to anything, in particular not to Null or to another invalid date
	// (like NaN != NaN), see TestDate_NormalizedEqual.
	assert.False(t, NullableDate("bad").NormalizedEqual("worse"))
	assert.False(t, NullableDate("bad").NormalizedEqual("bad"))
	assert.False(t, NullableDate("bad").NormalizedEqual(Null))
	assert.False(t, Null.NormalizedEqual("bad"))
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

func TestNormalizeNullable(t *testing.T) {
	tests := []struct {
		input   string
		lang    []language.Code
		want    NullableDate
		wantErr bool
	}{
		{input: "25.12.2023", want: "2023-12-25"},
		{input: "05/06/2024", lang: []language.Code{language.EN}, want: "2024-05-06"},
		// Empty and zero dates are valid NULL values for NullableDate
		{input: "", want: Null},
		{input: "0001-01-01", want: Null},
		{input: "0000-00-00", want: Null},
		{input: "not a date", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := NormalizeNullable(tt.input, tt.lang...)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMustNullable(t *testing.T) {
	assert.Equal(t, NullableDate("2023-12-25"), MustNullable("25.12.2023"))
	// Unlike Must for Date, an empty string is a valid NULL and must not panic
	assert.Equal(t, Null, MustNullable(""))
	assert.Panics(t, func() { MustNullable("not a date") })
}

func TestNullableDate_SetAndSetNull(t *testing.T) {
	var n NullableDate
	n.Set("2023-12-25")
	assert.Equal(t, NullableDate("2023-12-25"), n)
	assert.True(t, n.IsNotNull())

	n.SetNull()
	assert.Equal(t, Null, n)
	assert.True(t, n.IsNull())
}

func TestNullableDate_ScanString(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		validate bool
		want     NullableDate
		wantErr  bool
	}{
		// Textual NULL representations from CSV/form input must become NULL
		// and must not be treated as invalid dates.
		{name: "empty", source: "", validate: true, want: Null},
		{name: "NULL", source: "NULL", validate: true, want: Null},
		{name: "null", source: "null", validate: true, want: Null},
		{name: "nil", source: "nil", validate: true, want: Null},
		{name: "zero date", source: "0001-01-01", validate: true, want: Null},
		{name: "valid is normalized", source: "25.12.2023", validate: true, want: "2023-12-25"},
		{name: "invalid with validation", source: "invalid", validate: true, wantErr: true},
		{name: "invalid without validation is kept", source: "invalid", validate: false, want: "invalid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NullableDate("2000-01-01")
			err := n.ScanString(tt.source, tt.validate)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, NullableDate("2000-01-01"), n, "unchanged on error")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, n)
		})
	}
}

func TestNullableDate_ScanStringWithLang(t *testing.T) {
	tests := []struct {
		name                 string
		source               string
		lang                 language.Code
		want                 NullableDate
		wantNormalized       bool
		wantMonthMustBeFirst bool
		wantErr              bool
	}{
		{name: "empty is Null", source: "", want: Null},
		{name: "zero date is Null", source: "0000-00-00", want: Null},
		{name: "already normalized", source: "2023-12-25", want: "2023-12-25", wantNormalized: true},
		{name: "German day first", source: "25.12.2023", lang: language.DE, want: "2023-12-25"},
		// 25 can't be a month, so the input proves month-first ordering
		// which callers use to detect the date format of a whole document.
		{name: "unambiguous US order", source: "12/25/2023", lang: language.DE, want: "2023-12-25", wantMonthMustBeFirst: true},
		{name: "ambiguous with EN hint", source: "05/06/2024", lang: language.EN, want: "2024-05-06"},
		{name: "ambiguous with DE hint", source: "05/06/2024", lang: language.DE, want: "2024-06-05"},
		{name: "invalid", source: "invalid", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NullableDate("2000-01-01")
			wasNormalized, monthMustBeFirst, err := n.ScanStringWithLang(tt.source, tt.lang)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, NullableDate("2000-01-01"), n, "unchanged on error")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, n)
			assert.Equal(t, tt.wantNormalized, wasNormalized, "wasNormalized")
			assert.Equal(t, tt.wantMonthMustBeFirst, monthMustBeFirst, "monthMustBeFirst")
		})
	}
}

func TestNullableDate_String_Invalid(t *testing.T) {
	// Invalid dates are returned unchanged so that the
	// original input is visible in logs and error messages.
	assert.Equal(t, "invalid", NullableDate("invalid").String())
	assert.Equal(t, "2023-12-25", NullableDate("25.12.2023").String())
}

func TestNullableDate_AfterTime(t *testing.T) {
	loc := time.FixedZone("UTC+2", 2*60*60)
	d := NullableDate("2023-12-25")

	assert.True(t, d.AfterTime(time.Date(2023, 12, 24, 23, 59, 0, 0, loc)))
	assert.False(t, d.AfterTime(time.Date(2023, 12, 25, 0, 0, 1, 0, loc)))
	// Midnight is evaluated in the location of the compared time:
	// 00:30 UTC+2 is after midnight in UTC+2, although it is
	// before midnight UTC (it's 22:30 UTC of the previous day).
	assert.False(t, d.AfterTime(time.Date(2023, 12, 25, 0, 30, 0, 0, loc)))
	// Any date is after the zero time (no value)
	assert.True(t, d.AfterTime(time.Time{}))
	// NULL is never after anything
	assert.False(t, Null.AfterTime(time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)))
	assert.False(t, Null.AfterTime(time.Time{}))
}

func TestNullableDate_BeforeTime(t *testing.T) {
	d := NullableDate("2023-12-25")

	assert.True(t, d.BeforeTime(time.Date(2023, 12, 25, 0, 0, 1, 0, time.UTC)))
	assert.False(t, d.BeforeTime(time.Date(2023, 12, 24, 23, 59, 0, 0, time.UTC)))
	// NULL is before every non-zero time, consistent with Before(),
	// but not before the zero time which also represents "no value".
	assert.True(t, Null.BeforeTime(time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)))
	assert.False(t, Null.BeforeTime(time.Time{}))
}

func TestNullableDate_AddDate(t *testing.T) {
	d := NullableDate("2023-01-31")
	assert.Equal(t, NullableDate("2024-03-03"), d.AddDate(1, 1, 1), "2024-02-32 overflows into March like time.Time.AddDate")
	assert.Equal(t, NullableDate("2022-12-31"), d.AddDate(0, 0, -31))
	assert.Equal(t, Null, Null.AddDate(1, 1, 1), "NULL stays NULL")
}

func TestNullableDate_Add(t *testing.T) {
	d := NullableDate("2023-12-31")
	assert.Equal(t, NullableDate("2024-01-01"), d.Add(24*time.Hour))
	assert.Equal(t, NullableDate("2023-12-31"), d.Add(23*time.Hour), "partial days are truncated")
	assert.Equal(t, Null, Null.Add(24*time.Hour), "NULL stays NULL")
}

func TestNullableDate_Sub(t *testing.T) {
	a := NullableDate("2023-12-25")
	b := NullableDate("2023-12-20")
	assert.Equal(t, 5*24*time.Hour, a.Sub(b))
	assert.Equal(t, -5*24*time.Hour, b.Sub(a))
	assert.Equal(t, time.Duration(0), a.Sub(a))
	// There is no meaningful duration to or from NULL
	assert.Equal(t, time.Duration(0), a.Sub(Null))
	assert.Equal(t, time.Duration(0), Null.Sub(a))
}

func TestNullableDate_AfterTodayInUTC(t *testing.T) {
	tomorrow := NullableDate(OfNowInUTC().AddDays(1))
	yesterday := NullableDate(OfNowInUTC().AddDays(-1))

	assert.True(t, tomorrow.AfterTodayInUTC())
	assert.False(t, yesterday.AfterTodayInUTC())
	assert.False(t, Null.AfterTodayInUTC(), "NULL is never after today")
}

func TestNullableDate_BeforeTodayInUTC(t *testing.T) {
	tomorrow := NullableDate(OfNowInUTC().AddDays(1))
	yesterday := NullableDate(OfNowInUTC().AddDays(-1))

	assert.False(t, tomorrow.BeforeTodayInUTC())
	assert.True(t, yesterday.BeforeTodayInUTC())
	assert.True(t, Null.BeforeTodayInUTC(), "NULL is always before today")
}

func TestNullableDate_Format_Layouts(t *testing.T) {
	d := NullableDate("2023-12-25")
	assert.Equal(t, "25.12.2023", d.Format("02.01.2006"))
	assert.Equal(t, "", d.Format(""), "empty layout")
	// NULL must format as empty string, not as the zero time "01.01.0001"
	assert.Equal(t, "", Null.Format("02.01.2006"))
	assert.Equal(t, "", NullableDate("0001-01-01").Format("02.01.2006"))
}

func TestNullableDate_Scan(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    NullableDate
		wantErr bool
	}{
		{name: "SQL NULL", value: nil, want: Null},
		{name: "string", value: "2023-12-25", want: "2023-12-25"},
		{name: "string is normalized", value: "25.12.2023", want: "2023-12-25"},
		// Zero dates written by other systems must be read back as NULL
		{name: "empty string", value: "", want: Null},
		{name: "zero date 0000-00-00", value: "0000-00-00", want: Null},
		{name: "zero date 0001-01-01", value: "0001-01-01", want: Null},
		// Several SQL drivers return DATE and text columns as []byte,
		// which must scan exactly like the same string.
		{name: "[]byte", value: []byte("2023-12-25"), want: "2023-12-25"},
		{name: "[]byte is normalized", value: []byte("25.12.2023"), want: "2023-12-25"},
		{name: "empty []byte", value: []byte{}, want: Null},
		{name: "zero date []byte", value: []byte("0001-01-01"), want: Null},
		{name: "invalid []byte", value: []byte("invalid"), wantErr: true},
		{name: "time.Time", value: time.Date(2023, 12, 25, 23, 59, 0, 0, time.UTC), want: "2023-12-25"},
		{name: "zero time.Time", value: time.Time{}, want: Null},
		{name: "invalid string", value: "invalid", wantErr: true},
		{name: "unsupported type", value: 20231225, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NullableDate("2000-01-01")
			err := n.Scan(tt.value)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, n)
		})
	}
}

func TestNullableDate_Value_EdgeCases(t *testing.T) {
	// Zero dates must be written as SQL NULL, not as a bogus year 1 date
	val, err := NullableDate("0001-01-01").Value()
	require.NoError(t, err)
	assert.Nil(t, val)

	// Non-normalized dates are normalized before being written
	val, err = NullableDate("25.12.2023").Value()
	require.NoError(t, err)
	assert.Equal(t, "2023-12-25", val)

	// Invalid dates must not be silently written to the database
	_, err = NullableDate("invalid").Value()
	assert.Error(t, err)
}

func TestNullableDate_SQLRoundTrip(t *testing.T) {
	for _, original := range []NullableDate{Null, "2023-12-25", "1999-01-01"} {
		val, err := original.Value()
		require.NoError(t, err)

		var scanned NullableDate
		require.NoError(t, scanned.Scan(val))
		assert.Equal(t, original, scanned)
	}
}

func TestNullableDate_MarshalJSON(t *testing.T) {
	type wrapper struct {
		D NullableDate `json:"d"`
	}
	tests := []struct {
		date NullableDate
		want string
	}{
		{date: Null, want: `{"d":null}`},
		// A zero date is NULL and must be marshalled as JSON null
		{date: "0001-01-01", want: `{"d":null}`},
		{date: "2023-12-25", want: `{"d":"2023-12-25"}`},
	}
	for _, tt := range tests {
		t.Run(string(tt.date), func(t *testing.T) {
			data, err := json.Marshal(wrapper{D: tt.date})
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(data))
		})
	}

	t.Run("round trip", func(t *testing.T) {
		for _, original := range []NullableDate{Null, "2023-12-25"} {
			data, err := json.Marshal(wrapper{D: original})
			require.NoError(t, err)
			var decoded wrapper
			require.NoError(t, json.Unmarshal(data, &decoded))
			assert.Equal(t, original, decoded.D)
		}
	})
}

func TestNullableDate_JSONSchema(t *testing.T) {
	schema := NullableDate("").JSONSchema()
	require.Len(t, schema.OneOf, 2)
	assert.Equal(t, "string", schema.OneOf[0].Type)
	assert.Equal(t, "date", schema.OneOf[0].Format)
	assert.Equal(t, "null", schema.OneOf[1].Type, "NULL must be allowed by the schema")

	data, err := json.Marshal(schema.Default)
	require.NoError(t, err)
	assert.Equal(t, "null", string(data), "default value is JSON null")
}
