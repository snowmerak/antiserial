package test

import (
	"testing"

	"github.com/snowmerak/antiserial/test/testgen_nested"
)

// TestNestedStructAlwaysSerialized verifies that an empty nested struct still
// occupies wire space (bitmap byte), matching Go/Rust/C++/TS/Python codegen behavior.
func TestNestedStructAlwaysSerialized(t *testing.T) {
	w := testgen_nested.WithGeo{Id: 42}
	serialized, err := w.Marshal(nil)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// WithGeo bitmap: field 0 (id) + field 1 (geo) -> 0x03
	// id: 8 bytes LE
	// geo: empty struct bitmap 0x00
	wantLen := 1 + 8 + 1
	if len(serialized) != wantLen {
		t.Fatalf("wire length: got %d (%x) want %d", len(serialized), serialized, wantLen)
	}
	if serialized[0] != 0x03 {
		t.Fatalf("outer bitmap: got 0x%02x want 0x03", serialized[0])
	}
	if serialized[len(serialized)-1] != 0x00 {
		t.Fatalf("nested geo bitmap: got 0x%02x want 0x00", serialized[len(serialized)-1])
	}

	var decoded testgen_nested.WithGeo
	if _, err := decoded.Unmarshal(serialized); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if decoded.Id != 42 {
		t.Errorf("id: got %d want 42", decoded.Id)
	}
}