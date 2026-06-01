package nullable

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestType_JSONSchema(t *testing.T) {
	schema := Type[int]{}.JSONSchema()
	jsonSchemaBytes, err := json.Marshal(schema)
	require.NoError(t, err)
	require.Equal(t, `{"$schema":"https://json-schema.org/draft/2020-12/schema","oneOf":[{"type":"integer"},{"type":"null"}],"default":null}`, string(jsonSchemaBytes))
}

func Test_TypeFrom(t *testing.T) {
	v := TypeFrom(42)
	assert.True(t, v.IsNotNull())
	assert.Equal(t, 42, v.Get())
}

func Test_TypeFromPtr(t *testing.T) {
	assert.True(t, TypeFromPtr[int](nil).IsNull())

	i := 7
	v := TypeFromPtr(&i)
	assert.True(t, v.IsNotNull())
	assert.Equal(t, 7, v.Get())

	// Must be a copy, not an alias.
	i = 100
	assert.Equal(t, 7, v.Get())
}

func Test_Type_Ptr(t *testing.T) {
	var null Type[int]
	assert.Nil(t, null.Ptr())

	p := TypeFrom(5).Ptr()
	require.NotNil(t, p)
	assert.Equal(t, 5, *p)
}

func Test_Type_IsNull(t *testing.T) {
	var null Type[string]
	assert.True(t, null.IsNull())
	assert.False(t, null.IsNotNull())

	v := TypeFrom("x")
	assert.False(t, v.IsNull())
	assert.True(t, v.IsNotNull())
}

func Test_Type_Get(t *testing.T) {
	var null Type[int]
	assert.Panics(t, func() { null.Get() })
	assert.Equal(t, 9, TypeFrom(9).Get())
}

func Test_Type_GetOr(t *testing.T) {
	var null Type[int]
	assert.Equal(t, 99, null.GetOr(99))
	assert.Equal(t, 9, TypeFrom(9).GetOr(99))
}

func Test_Type_SetAndSetNull(t *testing.T) {
	var v Type[int]
	v.Set(3)
	assert.True(t, v.IsNotNull())
	assert.Equal(t, 3, v.Get())

	v.SetNull()
	assert.True(t, v.IsNull())
	assert.Nil(t, v.Ptr())
}

func Test_Type_Value(t *testing.T) {
	var null Type[int]
	val, err := null.Value()
	require.NoError(t, err)
	assert.Nil(t, val, "null Type returns SQL NULL")

	val, err = TypeFrom(42).Value()
	require.NoError(t, err)
	assert.Equal(t, int64(42), val, "int is converted to int64 for driver compatibility")

	type namedFloat float64
	val, err = TypeFrom(namedFloat(3.14)).Value()
	require.NoError(t, err)
	assert.Equal(t, float64(3.14), val, "named float64 is converted to primitive float64")
}

func Test_Type_Scan(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		v := TypeFrom(1)
		require.NoError(t, v.Scan(nil))
		assert.True(t, v.IsNull())
	})

	t.Run("int", func(t *testing.T) {
		var v Type[int64]
		require.NoError(t, v.Scan(int64(42)))
		assert.Equal(t, int64(42), v.Get())
	})

	t.Run("string", func(t *testing.T) {
		var v Type[string]
		require.NoError(t, v.Scan("hello"))
		assert.Equal(t, "hello", v.Get())
	})

	t.Run("bool", func(t *testing.T) {
		var v Type[bool]
		require.NoError(t, v.Scan(true))
		assert.Equal(t, true, v.Get())
	})

	t.Run("convert string to int", func(t *testing.T) {
		var v Type[int]
		require.NoError(t, v.Scan([]byte("123")))
		assert.Equal(t, 123, v.Get())
	})

	t.Run("error", func(t *testing.T) {
		var v Type[int]
		assert.Error(t, v.Scan("not a number"))
	})
}

func Test_Type_MarshalJSON(t *testing.T) {
	var null Type[int]
	b, err := null.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, "null", string(b))

	b, err = TypeFrom(42).MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, "42", string(b))
}

func Test_Type_UnmarshalJSON(t *testing.T) {
	var v Type[int]
	for _, nullJSON := range [][]byte{nil, {}, []byte("null")} {
		v = TypeFrom(1)
		require.NoError(t, v.UnmarshalJSON(nullJSON))
		assert.True(t, v.IsNull())
	}

	require.NoError(t, v.UnmarshalJSON([]byte("42")))
	assert.Equal(t, 42, v.Get())

	assert.Error(t, v.UnmarshalJSON([]byte("not json")))
}

func Test_Type_RoundTrip_JSON(t *testing.T) {
	type wrapper struct {
		V Type[string]
	}
	for _, original := range []Type[string]{{}, TypeFrom("hello")} {
		data, err := json.Marshal(wrapper{V: original})
		require.NoError(t, err)

		var scanned wrapper
		require.NoError(t, json.Unmarshal(data, &scanned))
		assert.Equal(t, original, scanned.V)
	}
}

func Test_Type_RoundTrip_SQL(t *testing.T) {
	for _, original := range []Type[int64]{{}, TypeFrom(int64(42))} {
		val, err := original.Value()
		require.NoError(t, err)

		var scanned Type[int64]
		require.NoError(t, scanned.Scan(val))
		assert.Equal(t, original, scanned)
	}
}

func Test_Type_JSONSchema_String(t *testing.T) {
	schema := Type[string]{}.JSONSchema()
	require.NotNil(t, schema)
	require.Len(t, schema.OneOf, 2)
	assert.Equal(t, "string", schema.OneOf[0].Type)
	assert.Equal(t, "null", schema.OneOf[1].Type)
}

func Test_Type_JSONSchema_NestedNullable(t *testing.T) {
	// A Type wrapping a type that already has a null oneOf option
	// (TrimmedString) must not add a second null option.
	schema := Type[TrimmedString]{}.JSONSchema()
	require.NotNil(t, schema)
	nullCount := 0
	for _, s := range schema.OneOf {
		if s.Type == "null" {
			nullCount++
		}
	}
	assert.Equal(t, 1, nullCount, "must not duplicate null oneOf option")
}

func Test_Type_JSONSchema_OneOfWithNull(t *testing.T) {
	// Time reflects to a OneOf schema that already contains a
	// "null" option, exercising the ContainsFunc/Default branch.
	schema := Type[Time]{}.JSONSchema()
	require.NotNil(t, schema)
	require.NotEmpty(t, schema.OneOf)
	nullCount := 0
	for _, s := range schema.OneOf {
		if s.Type == "null" {
			nullCount++
		}
	}
	assert.Equal(t, 1, nullCount, "must not duplicate the existing null option")
	require.NotNil(t, schema.Default)
}

// Test_Type_Scan_convertAssign exercises the copied database/sql
// convertAssign / asString / asBytes / strconvErr helpers through
// Type[T].Scan for a range of source and destination types.
func Test_Type_Scan_convertAssign(t *testing.T) {
	t.Run("string into string", func(t *testing.T) {
		var v Type[string]
		require.NoError(t, v.Scan("abc"))
		assert.Equal(t, "abc", v.Get())
	})

	t.Run("bytes into string", func(t *testing.T) {
		var v Type[string]
		require.NoError(t, v.Scan([]byte("abc")))
		assert.Equal(t, "abc", v.Get())
	})

	t.Run("string into bytes", func(t *testing.T) {
		var v Type[[]byte]
		require.NoError(t, v.Scan("abc"))
		assert.Equal(t, []byte("abc"), v.Get())
	})

	t.Run("bytes into bytes", func(t *testing.T) {
		var v Type[[]byte]
		require.NoError(t, v.Scan([]byte("abc")))
		assert.Equal(t, []byte("abc"), v.Get())
	})

	t.Run("int into bytes via asBytes", func(t *testing.T) {
		var v Type[[]byte]
		require.NoError(t, v.Scan(int64(123)))
		assert.Equal(t, []byte("123"), v.Get())
	})

	t.Run("uint into bytes via asBytes", func(t *testing.T) {
		var v Type[[]byte]
		require.NoError(t, v.Scan(uint64(7)))
		assert.Equal(t, []byte("7"), v.Get())
	})

	t.Run("float into bytes via asBytes", func(t *testing.T) {
		var v Type[[]byte]
		require.NoError(t, v.Scan(2.5))
		assert.Equal(t, []byte("2.5"), v.Get())
	})

	t.Run("bool into bytes via asBytes", func(t *testing.T) {
		var v Type[[]byte]
		require.NoError(t, v.Scan(false))
		assert.Equal(t, []byte("false"), v.Get())
	})

	t.Run("int into string via asString", func(t *testing.T) {
		var v Type[string]
		require.NoError(t, v.Scan(int64(123)))
		assert.Equal(t, "123", v.Get())
	})

	t.Run("bool into string via asString", func(t *testing.T) {
		var v Type[string]
		require.NoError(t, v.Scan(true))
		assert.Equal(t, "true", v.Get())
	})

	t.Run("float into string via asString", func(t *testing.T) {
		var v Type[string]
		require.NoError(t, v.Scan(1.5))
		assert.Equal(t, "1.5", v.Get())
	})

	t.Run("uint into string via asString", func(t *testing.T) {
		var v Type[string]
		require.NoError(t, v.Scan(uint64(5)))
		assert.Equal(t, "5", v.Get())
	})

	t.Run("float32 into string via asString", func(t *testing.T) {
		var v Type[string]
		require.NoError(t, v.Scan(float32(2.5)))
		assert.Equal(t, "2.5", v.Get())
	})

	t.Run("int into int8 (narrow)", func(t *testing.T) {
		var v Type[int8]
		require.NoError(t, v.Scan("100"))
		assert.Equal(t, int8(100), v.Get())
	})

	t.Run("int into bool via driver.Bool", func(t *testing.T) {
		var v Type[bool]
		require.NoError(t, v.Scan(int64(1)))
		assert.Equal(t, true, v.Get())
	})

	t.Run("int64 into int (numeric conversion)", func(t *testing.T) {
		var v Type[int]
		require.NoError(t, v.Scan(int64(7)))
		assert.Equal(t, 7, v.Get())
	})

	t.Run("string into uint", func(t *testing.T) {
		var v Type[uint]
		require.NoError(t, v.Scan("42"))
		assert.Equal(t, uint(42), v.Get())
	})

	t.Run("string into float", func(t *testing.T) {
		var v Type[float64]
		require.NoError(t, v.Scan("3.5"))
		assert.Equal(t, 3.5, v.Get())
	})

	t.Run("any destination", func(t *testing.T) {
		var v Type[any]
		require.NoError(t, v.Scan(int64(9)))
		assert.Equal(t, int64(9), v.Get())
	})

	t.Run("overflow error via strconvErr", func(t *testing.T) {
		var v Type[int8]
		assert.Error(t, v.Scan("99999"))
	})

	t.Run("invalid uint", func(t *testing.T) {
		var v Type[uint]
		assert.Error(t, v.Scan("not a uint"))
	})

	t.Run("invalid float", func(t *testing.T) {
		var v Type[float64]
		assert.Error(t, v.Scan("not a float"))
	})

	t.Run("unsupported conversion", func(t *testing.T) {
		var v Type[int]
		assert.Error(t, v.Scan(struct{}{}))
	})
}
