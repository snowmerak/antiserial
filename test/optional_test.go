package test

import (
	"bytes"
	"testing"

	"github.com/snowmerak/antiserial/test/testgen_optional"
)

func TestOptionalInt32ZeroRoundTrip(t *testing.T) {
	p := testgen_optional.Payload{
		Id:    1,
		Score: new(int32(0)),
	}
	serialized, err := p.Marshal(nil)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// bitmap: id (0) + score (1) -> 0x03, id 8 bytes, score 4 bytes (zero)
	if len(serialized) != 1+8+4 {
		t.Fatalf("length %d, bytes %x", len(serialized), serialized)
	}
	if serialized[0] != 0x03 {
		t.Fatalf("bitmap 0x%02x want 0x03", serialized[0])
	}

	var decoded testgen_optional.Payload
	if _, err := decoded.Unmarshal(serialized); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.Score == nil || *decoded.Score != 0 {
		t.Fatalf("score: got %v want 0", decoded.Score)
	}
}

func TestOptionalAbsentOmitsField(t *testing.T) {
	p := testgen_optional.Payload{Id: 1}
	serialized, err := p.Marshal(nil)
	if err != nil {
		t.Fatal(err)
	}
	// only id present: bitmap 0x01 + 8 bytes
	want := []byte{0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if !bytes.Equal(serialized, want) {
		t.Fatalf("got %x want %x", serialized, want)
	}
}