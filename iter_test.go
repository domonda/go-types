package types

import (
	"iter"
	"slices"
	"testing"

	"github.com/domonda/go-errs"
)

// TestYield checks that Yield produces exactly the one passed value,
// and that breaking out of the loop is safe.
func TestYield(t *testing.T) {
	if got := slices.Collect(Yield(42)); !slices.Equal(got, []int{42}) {
		t.Errorf("Collect(Yield(42)) = %v, want [42]", got)
	}
	for range Yield("x") {
		break // must not panic with "continued iteration after loop exit"
	}
}

// TestYield2 checks that Yield2 produces exactly the one passed pair.
func TestYield2(t *testing.T) {
	n := 0
	for k, v := range Yield2("key", 7) {
		n++
		if k != "key" || v != 7 {
			t.Errorf("Yield2 yielded (%q, %d), want (%q, %d)", k, v, "key", 7)
		}
	}
	if n != 1 {
		t.Errorf("Yield2 yielded %d pairs, want 1", n)
	}
}

// TestYieldErr checks that YieldErr yields the zero key together with
// the unchanged error, so callers ranging over a Seq2[K, error] get the
// error as the only element and can inspect it with errors.Is.
func TestYieldErr(t *testing.T) {
	wantErr := errs.New("test error")
	n := 0
	for k, err := range YieldErr[string](wantErr) {
		n++
		if k != "" {
			t.Errorf("YieldErr key = %q, want zero value", k)
		}
		if err != wantErr {
			t.Errorf("YieldErr error = %v, want %v unchanged", err, wantErr)
		}
	}
	if n != 1 {
		t.Errorf("YieldErr yielded %d pairs, want 1", n)
	}
}

// TestSeq2NilError checks that every value of the source sequence is
// yielded in order with a nil error, meaning success for every value.
func TestSeq2NilError(t *testing.T) {
	var got []int
	for v, err := range Seq2NilError(slices.Values([]int{1, 2, 3})) {
		if err != nil {
			t.Errorf("Seq2NilError yielded error %v for %d, want nil", err, v)
		}
		got = append(got, v)
	}
	if !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("Seq2NilError values = %v, want [1 2 3]", got)
	}

	for range Seq2NilError(slices.Values([]int(nil))) {
		t.Error("Seq2NilError of empty sequence yielded a pair")
	}
}

// TestSeq2NilError_StopsSource checks that breaking out of a loop over
// Seq2NilError stops the source sequence too, so expensive or infinite
// sources are not consumed beyond what the caller asked for.
func TestSeq2NilError_StopsSource(t *testing.T) {
	produced := 0
	var infinite iter.Seq[int] = func(yield func(int) bool) {
		for i := 0; ; i++ {
			produced++
			if !yield(i) {
				return
			}
		}
	}
	var got []int
	for v := range Seq2NilError(infinite) {
		got = append(got, v)
		if len(got) == 2 {
			break
		}
	}
	if !slices.Equal(got, []int{0, 1}) {
		t.Errorf("got %v, want [0 1]", got)
	}
	if produced != 2 {
		t.Errorf("source produced %d values, want 2", produced)
	}
}
