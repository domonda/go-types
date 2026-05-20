package nullable

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	_ sql.Scanner    = (*Time)(nil)
	_ driver.Valuer  = Time{}
	_ json.Marshaler = Time{}
)

var refTime = time.Date(2023, 6, 15, 12, 30, 45, 0, time.UTC)

func Test_TimeNow(t *testing.T) {
	now := TimeNow()
	assert.True(t, now.IsNotNull())
	assert.WithinDuration(t, time.Now(), now.Time, time.Second)
}

func Test_TimeParse(t *testing.T) {
	for _, nullStr := range []string{"", "null", "NULL"} {
		tm, err := TimeParse(time.RFC3339, nullStr)
		require.NoError(t, err, nullStr)
		assert.True(t, tm.IsNull(), nullStr)
	}

	tm, err := TimeParse(time.RFC3339, "2023-06-15T12:30:45Z")
	require.NoError(t, err)
	assert.True(t, tm.Equal(TimeFrom(refTime)))

	_, err = TimeParse(time.RFC3339, "not a time")
	assert.Error(t, err)
}

func Test_TimeParseInLocation(t *testing.T) {
	for _, nullStr := range []string{"", "null", "NULL"} {
		tm, err := TimeParseInLocation(time.RFC3339, nullStr, time.UTC)
		require.NoError(t, err, nullStr)
		assert.True(t, tm.IsNull(), nullStr)
	}

	tm, err := TimeParseInLocation("2006-01-02 15:04:05", "2023-06-15 12:30:45", time.UTC)
	require.NoError(t, err)
	assert.True(t, tm.Equal(TimeFrom(refTime)))

	_, err = TimeParseInLocation(time.RFC3339, "not a time", time.UTC)
	assert.Error(t, err)
}

func Test_TimeFrom(t *testing.T) {
	assert.True(t, TimeFrom(refTime).Equal(TimeFrom(refTime)))
	assert.True(t, TimeFrom(time.Time{}).IsNull())
}

func Test_TimeFromPtr(t *testing.T) {
	assert.True(t, TimeFromPtr(nil).IsNull())
	assert.True(t, TimeFromPtr(&refTime).Equal(TimeFrom(refTime)))
}

func Test_Time_Ptr(t *testing.T) {
	assert.Nil(t, TimeNull.Ptr())
	p := TimeFrom(refTime).Ptr()
	require.NotNil(t, p)
	assert.True(t, p.Equal(refTime))
}

func Test_Time_UTC(t *testing.T) {
	assert.True(t, TimeNull.UTC().IsNull())
	loc, _ := time.LoadLocation("Europe/Vienna")
	tm := TimeFrom(refTime.In(loc)).UTC()
	assert.Equal(t, time.UTC, tm.Time.Location())
	assert.True(t, tm.Equal(TimeFrom(refTime)))
}

func Test_Time_Add(t *testing.T) {
	assert.True(t, TimeNull.Add(time.Hour).IsNull())
	tm := TimeFrom(refTime).Add(time.Hour)
	assert.True(t, tm.Equal(TimeFrom(refTime.Add(time.Hour))))
}

func Test_Time_AddDate(t *testing.T) {
	assert.True(t, TimeNull.AddDate(1, 0, 0).IsNull())
	tm := TimeFrom(refTime).AddDate(1, 2, 3)
	assert.True(t, tm.Equal(TimeFrom(refTime.AddDate(1, 2, 3))))
}

func Test_Time_Equal(t *testing.T) {
	assert.True(t, TimeNull.Equal(TimeNull))
	assert.False(t, TimeNull.Equal(TimeFrom(refTime)))
	assert.False(t, TimeFrom(refTime).Equal(TimeNull))
	assert.True(t, TimeFrom(refTime).Equal(TimeFrom(refTime)))
	assert.False(t, TimeFrom(refTime).Equal(TimeFrom(refTime.Add(time.Second))))
}

func Test_Time_IsNull(t *testing.T) {
	assert.True(t, TimeNull.IsNull())
	assert.True(t, Time{}.IsNull())
	assert.False(t, TimeFrom(refTime).IsNull())

	assert.False(t, TimeNull.IsNotNull())
	assert.True(t, TimeFrom(refTime).IsNotNull())
}

func Test_Time_String(t *testing.T) {
	assert.Equal(t, "NULL", TimeNull.String())
	assert.Equal(t, refTime.String(), TimeFrom(refTime).String())

	assert.Equal(t, "n/a", TimeNull.StringOr("n/a"))
	assert.Equal(t, refTime.String(), TimeFrom(refTime).StringOr("n/a"))
}

func Test_Time_Format(t *testing.T) {
	assert.Equal(t, "", TimeNull.Format(time.RFC3339))
	assert.Equal(t, "2023-06-15T12:30:45Z", TimeFrom(refTime).Format(time.RFC3339))
}

func Test_Time_AppendFormat(t *testing.T) {
	assert.Equal(t, []byte("x"), TimeNull.AppendFormat([]byte("x"), time.RFC3339))
	assert.Equal(t, []byte("x2023-06-15T12:30:45Z"), TimeFrom(refTime).AppendFormat([]byte("x"), time.RFC3339))
}

func Test_Time_Get(t *testing.T) {
	assert.Panics(t, func() { TimeNull.Get() })
	assert.True(t, TimeFrom(refTime).Get().Equal(refTime))

	assert.True(t, TimeNull.GetOr(refTime).Equal(refTime))
	assert.True(t, TimeFrom(refTime).GetOr(time.Time{}).Equal(refTime))
}

func Test_Time_SetAndSetNull(t *testing.T) {
	var n Time
	n.Set(refTime)
	assert.True(t, n.Equal(TimeFrom(refTime)))
	n.SetNull()
	assert.True(t, n.IsNull())

	// Setting zero time results in null.
	n.Set(time.Time{})
	assert.True(t, n.IsNull())
}

func Test_Time_Scan(t *testing.T) {
	var n Time
	require.NoError(t, n.Scan(nil))
	assert.True(t, n.IsNull())

	require.NoError(t, n.Scan(refTime))
	assert.True(t, n.Equal(TimeFrom(refTime)))

	assert.Error(t, n.Scan("2023-06-15"))
	assert.Error(t, n.Scan(123))
}

func Test_Time_Value(t *testing.T) {
	val, err := TimeNull.Value()
	require.NoError(t, err)
	assert.Nil(t, val, "null Time returns SQL NULL")

	val, err = TimeFrom(refTime).Value()
	require.NoError(t, err)
	assert.Equal(t, refTime, val)
}

func Test_Time_RoundTrip_SQL(t *testing.T) {
	for _, original := range []Time{TimeNull, TimeFrom(refTime)} {
		val, err := original.Value()
		require.NoError(t, err)

		var scanned Time
		require.NoError(t, scanned.Scan(val))
		assert.True(t, original.Equal(scanned))
	}
}

func Test_Time_MarshalJSON(t *testing.T) {
	b, err := TimeNull.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, "null", string(b))

	b, err = TimeFrom(refTime).MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, `"2023-06-15T12:30:45Z"`, string(b))
}

func Test_Time_UnmarshalJSON(t *testing.T) {
	var n Time
	for _, nullJSON := range [][]byte{nil, {}, []byte("null")} {
		require.NoError(t, n.UnmarshalJSON(nullJSON))
		assert.True(t, n.IsNull())
	}

	require.NoError(t, n.UnmarshalJSON([]byte(`"2023-06-15T12:30:45Z"`)))
	assert.True(t, n.Equal(TimeFrom(refTime)))

	assert.Error(t, n.UnmarshalJSON([]byte(`"not a time"`)))
}

func Test_Time_RoundTrip_JSON(t *testing.T) {
	type wrapper struct {
		T Time
	}
	for _, original := range []Time{TimeNull, TimeFrom(refTime)} {
		data, err := json.Marshal(wrapper{T: original})
		require.NoError(t, err)

		var scanned wrapper
		require.NoError(t, json.Unmarshal(data, &scanned))
		assert.True(t, original.Equal(scanned.T))
	}

	// Verify null marshalling.
	data, err := json.Marshal(wrapper{})
	require.NoError(t, err)
	assert.Equal(t, `{"T":null}`, string(data))
}

func Test_Time_JSONSchema(t *testing.T) {
	schema := Time{}.JSONSchema()
	require.NotNil(t, schema)
	assert.Equal(t, "Nullable Time", schema.Title)
	require.Len(t, schema.OneOf, 2)
	assert.Equal(t, "string", schema.OneOf[0].Type)
	assert.Equal(t, "date-time", schema.OneOf[0].Format)
	assert.Equal(t, "null", schema.OneOf[1].Type)
}

func Test_Time_MarshalText(t *testing.T) {
	b, err := TimeNull.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "NULL", string(b))

	b, err = TimeFrom(refTime).MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "2023-06-15T12:30:45Z", string(b))
}

func Test_Time_UnmarshalText(t *testing.T) {
	var n Time
	for _, nullText := range [][]byte{nil, {}, []byte("null"), []byte("NULL")} {
		require.NoError(t, n.UnmarshalText(nullText))
		assert.True(t, n.IsNull())
	}

	require.NoError(t, n.UnmarshalText([]byte("2023-06-15T12:30:45Z")))
	assert.True(t, n.Equal(TimeFrom(refTime)))

	assert.Error(t, n.UnmarshalText([]byte("not a time")))
}

func Test_Time_PrettyPrint(t *testing.T) {
	var buf bytes.Buffer
	n, err := TimeNull.PrettyPrint(&buf)
	require.NoError(t, err)
	assert.Equal(t, "NULL", buf.String())
	assert.Equal(t, buf.Len(), n)

	buf.Reset()
	_, err = TimeFrom(refTime).PrettyPrint(&buf)
	require.NoError(t, err)
	assert.NotEmpty(t, buf.String())
}
