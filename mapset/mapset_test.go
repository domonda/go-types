package mapset

import (
	"maps"
	"slices"
	"testing"
)

// set is the canonical map[K]struct{} representation used by most tests.
type set = map[string]struct{}

// namedSet exercises the ~map[K]V constraint: the functions must work on
// defined types like uu.IDSet or email.AddressSet, not just on the
// unnamed map type, and the ones returning MX must return the named type.
type namedSet map[string]struct{}

func sorted[M ~map[string]V, V bool | struct{}](m M) []string {
	return slices.Sorted(maps.Keys(m))
}

func eq(t *testing.T, got []string, want ...string) {
	t.Helper()
	if want == nil {
		want = []string{}
	}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestOfAndCollect(t *testing.T) {
	eq(t, sorted(Of("b", "a", "b")), "a", "b")
	eq(t, sorted(Of[string]()))
	eq(t, sorted(Collect(slices.Values([]string{"b", "a", "b"}))), "a", "b")
	eq(t, sorted(Collect(slices.Values([]string(nil)))))

	// The bool variants must store true, not the false zero value:
	// legacy map[K]bool sets are read with `if m[k]`, which would
	// report a false-valued key as absent.
	for _, m := range []map[string]bool{OfBool("a"), CollectBool(slices.Values([]string{"a"}))} {
		if !m["a"] {
			t.Errorf("map[K]bool set must store true, got %v", m)
		}
	}
}

func TestString(t *testing.T) {
	// Sorted output so the representation is deterministic.
	if got, want := String(Of("c", "a", "b")), "{a b c}"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
	if got, want := String(set(nil)), "{}"; got != want {
		t.Errorf("String(nil) = %q, want %q", got, want)
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		name string
		x, y set
		want bool
	}{
		{"both empty", Of[string](), Of[string](), true},
		{"nil equals empty", nil, Of[string](), true},
		{"nil equals nil", nil, nil, true},
		{"same elements", Of("a", "b"), Of("b", "a"), true},
		{"different length", Of("a"), Of("a", "b"), false},
		{"same length different elements", Of("a"), Of("b"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Equal(tt.x, tt.y); got != tt.want {
				t.Errorf("Equal(%v, %v) = %v, want %v", tt.x, tt.y, got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	x := Of("a", "b")
	if !Contains(x, "a") {
		t.Error("Contains(x, a) = false, want true")
	}
	if Contains(x, "z") {
		t.Error("Contains(x, z) = true, want false")
	}
	// A nil set is a valid empty set, not a panic.
	if Contains(set(nil), "a") {
		t.Error("Contains(nil, a) = true, want false")
	}
}

func TestContainsAll(t *testing.T) {
	x := Of("a", "b", "c")
	if !ContainsAll(x, slices.Values([]string{"a", "c"})) {
		t.Error("ContainsAll(x, [a c]) = false, want true")
	}
	if ContainsAll(x, slices.Values([]string{"a", "z"})) {
		t.Error("ContainsAll(x, [a z]) = true, want false")
	}
	// Vacuously true, matching the "for all" reading of the name.
	if !ContainsAll(x, slices.Values([]string(nil))) {
		t.Error("ContainsAll(x, []) = false, want true")
	}
	if !ContainsAll(set(nil), slices.Values([]string(nil))) {
		t.Error("ContainsAll(nil, []) = false, want true")
	}
	if ContainsAll(set(nil), slices.Values([]string{"a"})) {
		t.Error("ContainsAll(nil, [a]) = true, want false")
	}
}

func TestAll(t *testing.T) {
	eq(t, slices.Sorted(All(Of("b", "a"))), "a", "b")
	eq(t, slices.Sorted(All(set(nil))))

	// All must honour an early return from the loop body,
	// otherwise `break` in a range-over-func loop would not work.
	count := 0
	for range All(Of("a", "b", "c")) {
		count++
		break
	}
	if count != 1 {
		t.Errorf("All() yielded %d elements after break, want 1", count)
	}
}

func TestIntersects(t *testing.T) {
	tests := []struct {
		name string
		x, y set
		want bool
	}{
		{"common element", Of("a", "b"), Of("b", "c"), true},
		{"disjoint", Of("a"), Of("b"), false},
		{"nil left", nil, Of("a"), false},
		{"nil right", Of("a"), nil, false},
		// The implementation iterates the smaller set, so both
		// size relations must be covered.
		{"larger left", Of("a", "b", "c"), Of("c"), true},
		{"larger right", Of("c"), Of("a", "b", "c"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Intersects(tt.x, tt.y); got != tt.want {
				t.Errorf("Intersects(%v, %v) = %v, want %v", tt.x, tt.y, got, tt.want)
			}
		})
	}
}

func TestUnion(t *testing.T) {
	x, y := Of("a", "b"), Of("b", "c")
	eq(t, sorted(Union(x, y)), "a", "b", "c")
	// Operands must not be modified.
	eq(t, sorted(x), "a", "b")
	eq(t, sorted(y), "b", "c")

	eq(t, sorted(Union(set(nil), Of("a"))), "a")
	eq(t, sorted(Union(Of("a"), set(nil))), "a")
	eq(t, sorted(Union(set(nil), set(nil))))
}

func TestIntersection(t *testing.T) {
	x, y := Of("a", "b"), Of("b", "c")
	eq(t, sorted(Intersection(x, y)), "b")
	eq(t, sorted(x), "a", "b")
	eq(t, sorted(y), "b", "c")

	eq(t, sorted(Intersection(set(nil), Of("a"))))
	eq(t, sorted(Intersection(Of("a"), set(nil))))
}

func TestDifference(t *testing.T) {
	x, y := Of("a", "b"), Of("b", "c")
	// Asymmetric: only the elements of x that are not in y.
	// This is the semantic that differs from the pre-2026-08 Diff/Difference
	// methods of this module, which returned the symmetric difference.
	eq(t, sorted(Difference(x, y)), "a")
	eq(t, sorted(Difference(y, x)), "c")
	eq(t, sorted(x), "a", "b")
	eq(t, sorted(y), "b", "c")

	eq(t, sorted(Difference(set(nil), Of("a"))))
	eq(t, sorted(Difference(Of("a"), set(nil))), "a")
}

func TestSymmetricDifference(t *testing.T) {
	x, y := Of("a", "b"), Of("b", "c")
	eq(t, sorted(SymmetricDifference(x, y)), "a", "c")
	// Symmetric, unlike Difference.
	eq(t, sorted(SymmetricDifference(y, x)), "a", "c")
	eq(t, sorted(x), "a", "b")
	eq(t, sorted(y), "b", "c")

	eq(t, sorted(SymmetricDifference(set(nil), Of("a"))), "a")
	eq(t, sorted(SymmetricDifference(Of("a"), set(nil))), "a")
}

func TestInsert(t *testing.T) {
	x := Of[string]()
	if !Insert(x, "a") {
		t.Error("Insert of new element = false, want true")
	}
	// The bool result exists so callers can tell "added" from "already there";
	// re-inserting must report no change.
	if Insert(x, "a") {
		t.Error("Insert of existing element = true, want false")
	}
	eq(t, sorted(x), "a")
}

func TestInsertAll(t *testing.T) {
	x := Of("a")
	if !InsertAll(x, slices.Values([]string{"a", "b"})) {
		t.Error("InsertAll adding b = false, want true")
	}
	eq(t, sorted(x), "a", "b")
	if InsertAll(x, slices.Values([]string{"a", "b"})) {
		t.Error("InsertAll of existing elements = true, want false")
	}
	if InsertAll(x, slices.Values([]string(nil))) {
		t.Error("InsertAll of empty sequence = true, want false")
	}
	// Nothing to store, so a nil set must not panic.
	if InsertAll(set(nil), slices.Values([]string(nil))) {
		t.Error("InsertAll(nil, []) = true, want false")
	}
}

func TestDelete(t *testing.T) {
	x := Of("a")
	if !Delete(x, "a") {
		t.Error("Delete of present element = false, want true")
	}
	if Delete(x, "a") {
		t.Error("Delete of absent element = true, want false")
	}
	// Deleting from a nil map is legal in Go and must stay legal here.
	if Delete(set(nil), "a") {
		t.Error("Delete(nil, a) = true, want false")
	}
}

func TestDeleteAll(t *testing.T) {
	x := Of("a", "b", "c")
	if !DeleteAll(x, slices.Values([]string{"a", "z"})) {
		t.Error("DeleteAll removing a = false, want true")
	}
	eq(t, sorted(x), "b", "c")
	if DeleteAll(x, slices.Values([]string{"z"})) {
		t.Error("DeleteAll of absent elements = true, want false")
	}
	if DeleteAll(set(nil), slices.Values([]string{"a"})) {
		t.Error("DeleteAll(nil, [a]) = true, want false")
	}
}

func TestDeleteFunc(t *testing.T) {
	x := Of("apple", "banana", "avocado")
	if !DeleteFunc(x, func(s string) bool { return s[0] == 'a' }) {
		t.Error("DeleteFunc removing elements = false, want true")
	}
	eq(t, sorted(x), "banana")
	if DeleteFunc(x, func(string) bool { return false }) {
		t.Error("DeleteFunc removing nothing = true, want false")
	}
	if DeleteFunc(set(nil), func(string) bool { return true }) {
		t.Error("DeleteFunc(nil, _) = true, want false")
	}
}

func TestUnionWith(t *testing.T) {
	x, y := Of("a", "b"), Of("b", "c")
	UnionWith(x, y)
	eq(t, sorted(x), "a", "b", "c")
	// The right operand must not be modified.
	eq(t, sorted(y), "b", "c")

	// Nothing to store, so a nil left operand must not panic.
	UnionWith(set(nil), set(nil))
	UnionWith(set(nil), Of[string]())
}

func TestIntersectionWith(t *testing.T) {
	x, y := Of("a", "b"), Of("b", "c")
	IntersectionWith(x, y)
	eq(t, sorted(x), "b")
	eq(t, sorted(y), "b", "c")

	// Only removes, so a nil left operand is valid.
	IntersectionWith(set(nil), Of("a"))
}

func TestDifferenceWith(t *testing.T) {
	x, y := Of("a", "b"), Of("b", "c")
	DifferenceWith(x, y)
	eq(t, sorted(x), "a")
	eq(t, sorted(y), "b", "c")

	// Only removes, so a nil left operand is valid.
	DifferenceWith(set(nil), Of("a"))
}

func TestSymmetricDifferenceWith(t *testing.T) {
	x, y := Of("a", "b"), Of("b", "c")
	SymmetricDifferenceWith(x, y)
	eq(t, sorted(x), "a", "c")
	eq(t, sorted(y), "b", "c")

	// Nothing to store, so a nil left operand must not panic.
	SymmetricDifferenceWith(set(nil), set(nil))
}

// TestNilPanics pins down the documented boundary between the operations
// that accept a nil set and the ones that panic like a nil map assignment.
// The set types of this module rely on this split: it is why
// email.AddressSet keeps pointer-receiver Add and AddSet methods.
func TestNilPanics(t *testing.T) {
	tests := []struct {
		name string
		call func()
	}{
		{"Insert", func() { Insert(set(nil), "a") }},
		{"InsertAll", func() { InsertAll(set(nil), slices.Values([]string{"a"})) }},
		{"UnionWith", func() { UnionWith(set(nil), Of("a")) }},
		{"SymmetricDifferenceWith", func() { SymmetricDifferenceWith(set(nil), Of("a")) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("%s on a nil set did not panic", tt.name)
				}
			}()
			tt.call()
		})
	}
}

// TestNamedSetTypes checks that the ~map[K]V constraints accept defined
// map types and that the functions returning MX return that same defined
// type, which is what lets the set types of this module delegate here.
func TestNamedSetTypes(t *testing.T) {
	x := namedSet{"a": {}, "b": {}}
	y := namedSet{"b": {}, "c": {}}

	var union namedSet = Union(x, y)
	eq(t, sorted(union), "a", "b", "c")
	var intersection namedSet = Intersection(x, y)
	eq(t, sorted(intersection), "b")
	var difference namedSet = Difference(x, y)
	eq(t, sorted(difference), "a")
	var symmetric namedSet = SymmetricDifference(x, y)
	eq(t, sorted(symmetric), "a", "c")

	if !Insert(x, "d") || !Delete(x, "d") {
		t.Error("Insert/Delete on a named set type did not report a change")
	}
	if !Contains(x, "a") || !Intersects(x, y) || !Equal(x, namedSet{"a": {}, "b": {}}) {
		t.Error("query functions on a named set type returned the wrong result")
	}
}

// TestBoolSets covers the map[K]bool representation the package supports
// for legacy sets. Membership is the presence of the key, but a stored
// value of true keeps `if m[k]` working for code that predates this package.
func TestBoolSets(t *testing.T) {
	x := OfBool("a", "b")
	y := OfBool("b", "c")

	if !Contains(x, "a") || !Equal(x, OfBool("b", "a")) || !Intersects(x, y) {
		t.Error("query functions on a map[K]bool set returned the wrong result")
	}
	eq(t, sorted(Union(x, y)), "a", "b", "c")
	eq(t, sorted(Difference(x, y)), "a")

	for _, k := range []string{"a", "b", "c"} {
		if !Union(x, y)[k] {
			t.Errorf("Union of map[K]bool sets stored false for %q", k)
		}
	}
	if !Insert(x, "d") {
		t.Error("Insert into a map[K]bool set = false, want true")
	}
	if !x["d"] {
		t.Error("Insert into a map[K]bool set stored false, want true")
	}
	UnionWith(x, OfBool("e"))
	if !x["e"] {
		t.Error("UnionWith on a map[K]bool set stored false, want true")
	}
	SymmetricDifferenceWith(x, OfBool("f"))
	if !x["f"] {
		t.Error("SymmetricDifferenceWith on a map[K]bool set stored false, want true")
	}
}

// TestMixedRepresentations covers the independent VX/VY type parameters of
// Union, Intersection, Difference, SymmetricDifference and UnionWith. Their
// signatures allow mixing the map[K]struct{} and map[K]bool representations,
// but every other test instantiates them homogeneously, and per-block coverage
// cannot see an instantiation that is never compiled. Membership depends only
// on the key, so mixing must work and the result must take the representation
// of the left operand.
//
// Note the asymmetry this test pins by omission: IntersectionWith,
// DifferenceWith and SymmetricDifferenceWith declare both operands as the same
// M, and Intersects shares one V, so none of them compile with mixed
// representations. Only UnionWith does.
func TestMixedRepresentations(t *testing.T) {
	structs := Of("a", "b")
	bools := OfBool("b", "c")

	var union map[string]struct{} = Union(structs, bools)
	eq(t, sorted(union), "a", "b", "c")
	var intersection map[string]struct{} = Intersection(structs, bools)
	eq(t, sorted(intersection), "b")
	var difference map[string]struct{} = Difference(structs, bools)
	eq(t, sorted(difference), "a")
	var symmetric map[string]struct{} = SymmetricDifference(structs, bools)
	eq(t, sorted(symmetric), "a", "c")

	// Reversed, so the bool set is the left operand whose representation wins.
	var boolUnion map[string]bool = Union(bools, structs)
	eq(t, sorted(boolUnion), "a", "b", "c")
	for k, v := range boolUnion {
		if !v {
			t.Errorf("Union into a bool set stored false for %q, want true", k)
		}
	}

	UnionWith(structs, bools)
	eq(t, sorted(structs), "a", "b", "c")
	// The argument must survive the in-place operation unchanged.
	eq(t, sorted(bools), "b", "c")
}

// TestMixedNamedTypes covers the independent MX/MY of Intersects, which share
// one representation V but allow two different named map types.
func TestMixedNamedTypes(t *testing.T) {
	named := namedSet{"a": {}, "b": {}}
	plain := Of("b", "c")

	if !Intersects(named, plain) {
		t.Error("Intersects across named and plain map types = false, want true")
	}
	if Intersects(named, Of("z")) {
		t.Error("Intersects of disjoint sets = true, want false")
	}
}

// TestNilOperandsAllocate pins the package doc's promise that Union,
// Intersection, Difference and SymmetricDifference "accept nil operands and
// always return a newly allocated set". Every other nil-operand assertion goes
// through sorted(), which renders a nil and an empty map identically, so a
// short-circuit `return nil` for two empty operands would keep the suite green
// and then panic on the caller's next Insert.
func TestNilOperandsAllocate(t *testing.T) {
	var nilSet map[string]struct{}

	ops := map[string]func() map[string]struct{}{
		"Union":               func() map[string]struct{} { return Union(nilSet, nilSet) },
		"Intersection":        func() map[string]struct{} { return Intersection(nilSet, nilSet) },
		"Difference":          func() map[string]struct{} { return Difference(nilSet, nilSet) },
		"SymmetricDifference": func() map[string]struct{} { return SymmetricDifference(nilSet, nilSet) },
	}
	for name, op := range ops {
		t.Run(name, func(t *testing.T) {
			got := op()
			if got == nil {
				t.Fatalf("%s of two nil sets returned nil, want an allocated empty set", name)
			}
			// Writable is the property that actually matters to callers.
			if !Insert(got, "a") {
				t.Errorf("%s result did not accept an Insert", name)
			}
		})
	}
}
