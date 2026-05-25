package nullable

import (
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test fixture types for the various Nullable / Zeroable shapes.

// nullableValueT implements Nullable on a value receiver.
type nullableValueT struct{ null bool }

func (n nullableValueT) IsNull() bool { return n.null }

// nullablePtrT implements Nullable on a pointer receiver only.
// IsNull is deliberately written to panic on a nil receiver so that
// ReflectIsNull is forced to detect the typed-nil case before dispatch.
type nullablePtrT struct{ null bool }

func (n *nullablePtrT) IsNull() bool { return n.null }

// zeroableValueT implements only Zeroable.
type zeroableValueT struct{ zero bool }

func (z zeroableValueT) IsZero() bool { return z.zero }

// zeroablePtrT implements Zeroable on a pointer receiver only.
type zeroablePtrT struct{ zero bool }

func (z *zeroablePtrT) IsZero() bool { return z.zero }

// bothT implements both Nullable and Zeroable; Nullable must win.
type bothT struct {
	nullable bool
	zero     bool
}

func (b bothT) IsNull() bool { return b.nullable }
func (b bothT) IsZero() bool { return b.zero }

// plainT implements neither interface.
type plainT struct {
	X int
	S string
}

func TestReflectIsNull(t *testing.T) {
	t.Run("Original table (regression)", func(t *testing.T) {
		// given
		tests := []struct {
			name string
			v    reflect.Value
			want bool
		}{
			{"zero reflect.Value", reflect.Value{}, true},
			{"nil interface", reflect.ValueOf(nil), true},
			{"nil int ptr", reflect.ValueOf((*int)(nil)), true},
			{"nil Nullable", reflect.ValueOf(Nullable(nil)), true},
			{"nil Zeroable", reflect.ValueOf(Zeroable(nil)), true},
			{"nil IntArray", reflect.ValueOf(IntArray(nil)), true},
			{"nil func", reflect.ValueOf((func())(nil)), true},
			{"zero time.Time", reflect.ValueOf(time.Time{}), true},
			{"zero time.Time ptr", reflect.ValueOf(new(time.Time)), true},
			{"nil time.Time ptr", reflect.ValueOf((*time.Time)(nil)), true},
			{"TimeNull", reflect.ValueOf(TimeNull), true},

			{"0", reflect.ValueOf(0), false},
			{"0 ptr", reflect.ValueOf(new(int)), false},
			{"empty IntArray", reflect.ValueOf(IntArray{}), false},
			{"empty string", reflect.ValueOf(""), false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				require.NotPanics(t, func() {
					assert.Equal(t, tt.want, ReflectIsNull(tt.v))
				})
			})
		}
	})

	t.Run("Nested pointers — any nil along the chain is null", func(t *testing.T) {
		// given
		var innerNil *int
		ppNil := &innerNil // **int, outer non-nil, inner nil

		var deepNil **int
		pppNil := &deepNil // ***int, outermost non-nil, middle nil

		i := 7
		p := &i
		pp := &p // **int, fully non-nil

		// then
		assert.True(t, ReflectIsNull(reflect.ValueOf(ppNil)), "**int with nil inner")
		assert.True(t, ReflectIsNull(reflect.ValueOf(pppNil)), "***int with nil middle")
		assert.False(t, ReflectIsNull(reflect.ValueOf(pp)), "**int fully non-nil")
	})

	t.Run("Interface-wrapped typed-nil pointer", func(t *testing.T) {
		// Putting a typed-nil pointer into an interface creates a non-nil
		// interface whose dynamic value is nil. Only reachable via reflect
		// when the Value is itself of Kind Interface (e.g. via a struct field).
		type holder struct{ Any any }

		t.Run("nil *int in interface field", func(t *testing.T) {
			// given
			var p *int
			h := holder{Any: p}
			fv := reflect.ValueOf(h).Field(0)

			// when / then
			require.Equal(t, reflect.Interface, fv.Kind())
			require.False(t, fv.IsNil(), "interface holding typed-nil is not itself nil")
			require.NotPanics(t, func() {
				assert.True(t, ReflectIsNull(fv))
			})
		})

		t.Run("nil pointer-receiver Nullable in interface field", func(t *testing.T) {
			// given — nullablePtrT.IsNull would dereference n.null and panic
			// on a nil receiver; ReflectIsNull must short-circuit before dispatch.
			var n *nullablePtrT
			h := holder{Any: n}
			fv := reflect.ValueOf(h).Field(0)

			// then
			require.NotPanics(t, func() {
				assert.True(t, ReflectIsNull(fv))
			})
		})

		t.Run("non-nil value in interface field", func(t *testing.T) {
			// given
			h := holder{Any: 42}
			fv := reflect.ValueOf(h).Field(0)

			// then
			assert.False(t, ReflectIsNull(fv))
		})

		t.Run("Nullable=true value in interface field", func(t *testing.T) {
			// given
			h := holder{Any: nullableValueT{null: true}}
			fv := reflect.ValueOf(h).Field(0)

			// then
			assert.True(t, ReflectIsNull(fv))
		})
	})

	t.Run("Nilable reference types", func(t *testing.T) {
		// given
		var (
			nilMap   map[string]int
			nilSlice []int
			nilChan  chan int
			nilFunc  func()
			nilUnsP  unsafe.Pointer
		)

		// then
		assert.True(t, ReflectIsNull(reflect.ValueOf(nilMap)), "nil map")
		assert.True(t, ReflectIsNull(reflect.ValueOf(nilSlice)), "nil slice")
		assert.True(t, ReflectIsNull(reflect.ValueOf(nilChan)), "nil chan")
		assert.True(t, ReflectIsNull(reflect.ValueOf(nilFunc)), "nil func")
		assert.True(t, ReflectIsNull(reflect.ValueOf(nilUnsP)), "nil unsafe.Pointer")

		// non-nil but empty / allocated
		assert.False(t, ReflectIsNull(reflect.ValueOf(map[string]int{})), "empty map")
		assert.False(t, ReflectIsNull(reflect.ValueOf([]int{})), "empty slice")
		assert.False(t, ReflectIsNull(reflect.ValueOf(make(chan int))), "non-nil chan")
		assert.False(t, ReflectIsNull(reflect.ValueOf(func() {})), "non-nil func")
		x := 1
		assert.False(t, ReflectIsNull(reflect.ValueOf(unsafe.Pointer(&x))), "non-nil unsafe.Pointer")
	})

	t.Run("Nullable dispatch", func(t *testing.T) {
		t.Run("value receiver, IsNull=true", func(t *testing.T) {
			assert.True(t, ReflectIsNull(reflect.ValueOf(nullableValueT{null: true})))
		})
		t.Run("value receiver, IsNull=false", func(t *testing.T) {
			assert.False(t, ReflectIsNull(reflect.ValueOf(nullableValueT{null: false})))
		})
		t.Run("pointer receiver via &T, IsNull=true", func(t *testing.T) {
			n := nullablePtrT{null: true}
			assert.True(t, ReflectIsNull(reflect.ValueOf(&n)))
		})
		t.Run("pointer receiver via &T, IsNull=false", func(t *testing.T) {
			n := nullablePtrT{null: false}
			assert.False(t, ReflectIsNull(reflect.ValueOf(&n)))
		})
		t.Run("pointer receiver via T (not addressable) — method invisible", func(t *testing.T) {
			// reflect.ValueOf(T) yields a non-addressable Value, so the
			// *T receiver is unreachable. This matches the original behavior.
			assert.False(t, ReflectIsNull(reflect.ValueOf(nullablePtrT{null: true})))
		})
	})

	t.Run("Zeroable dispatch", func(t *testing.T) {
		t.Run("value receiver, IsZero=true", func(t *testing.T) {
			assert.True(t, ReflectIsNull(reflect.ValueOf(zeroableValueT{zero: true})))
		})
		t.Run("value receiver, IsZero=false", func(t *testing.T) {
			assert.False(t, ReflectIsNull(reflect.ValueOf(zeroableValueT{zero: false})))
		})
		t.Run("pointer receiver via &T", func(t *testing.T) {
			z := zeroablePtrT{zero: true}
			assert.True(t, ReflectIsNull(reflect.ValueOf(&z)))
		})
	})

	t.Run("Nullable takes precedence over Zeroable", func(t *testing.T) {
		t.Run("Nullable=false, Zeroable=true → false", func(t *testing.T) {
			assert.False(t, ReflectIsNull(reflect.ValueOf(bothT{nullable: false, zero: true})))
		})
		t.Run("Nullable=true, Zeroable=false → true", func(t *testing.T) {
			assert.True(t, ReflectIsNull(reflect.ValueOf(bothT{nullable: true, zero: false})))
		})
	})

	t.Run("Unexported struct fields do not panic", func(t *testing.T) {
		// given — Interface() panics on unexported fields, so ReflectIsNull
		// must guard with CanInterface() and treat such values as not-null.
		type s struct {
			plainHidden     int
			ptrHidden       *int
			nullableHidden  nullableValueT
			zeroableHidden  zeroableValueT
			sliceHidden     []int
			interfaceHidden any
		}
		v := reflect.ValueOf(s{
			plainHidden:    42,
			ptrHidden:      nil,
			nullableHidden: nullableValueT{null: true},
			zeroableHidden: zeroableValueT{zero: true},
			sliceHidden:    nil,
		})

		// then — pointer/slice/interface nil checks run on Kind before
		// touching Interface(), so they still report true. The struct
		// fields fall through to the CanInterface gate and are reported
		// as not-null without panicking.
		require.NotPanics(t, func() {
			assert.False(t, ReflectIsNull(v.FieldByName("plainHidden")), "unexported int field, value 42")
			assert.True(t, ReflectIsNull(v.FieldByName("ptrHidden")), "unexported nil pointer")
			assert.False(t, ReflectIsNull(v.FieldByName("nullableHidden")), "unexported Nullable — can't dispatch, treated not-null")
			assert.False(t, ReflectIsNull(v.FieldByName("zeroableHidden")), "unexported Zeroable — can't dispatch, treated not-null")
			assert.True(t, ReflectIsNull(v.FieldByName("sliceHidden")), "unexported nil slice")
			assert.True(t, ReflectIsNull(v.FieldByName("interfaceHidden")), "unexported nil interface")
		})
	})

	t.Run("Plain types without Nullable/Zeroable", func(t *testing.T) {
		// Zero values of types implementing neither interface are not null.
		assert.False(t, ReflectIsNull(reflect.ValueOf(plainT{})), "zero plainT")
		assert.False(t, ReflectIsNull(reflect.ValueOf(plainT{X: 1, S: "x"})), "non-zero plainT")
	})

	t.Run("Pointer to value-receiver Nullable", func(t *testing.T) {
		// *T where T implements Nullable on a value receiver — unwrap finds
		// the inner T, type-asserts Nullable, dispatches IsNull().
		n := nullableValueT{null: true}
		assert.True(t, ReflectIsNull(reflect.ValueOf(&n)))

		n2 := nullableValueT{null: false}
		assert.False(t, ReflectIsNull(reflect.ValueOf(&n2)))
	})
}

func BenchmarkReflectIsNull(b *testing.B) {
	// Hot-path benchmark mirroring the call sites in strfmt and go-structtable:
	// the value usually implements neither Nullable nor Zeroable, so the
	// Implements() gate should skip the boxing v.Interface() does.
	v := reflect.ValueOf(plainT{X: 1, S: "hello"})
	b.ResetTimer()
	for range b.N {
		if ReflectIsNull(v) {
			b.Fatal("plainT must not be null")
		}
	}
}
