package test

import (
	"bytes"
	"testing"

	"github.com/snowmerak/antiserial/test/testgen_v1"
	"github.com/snowmerak/antiserial/test/testgen_v2"
)

func TestEndToEnd(t *testing.T) {
	// 1. Create a version 2 payload
	p2 := testgen_v2.Payload{
		Id:     1234567890,
		Uuid:   "abc",
		Active: true,
		Tags:   []string{"go", "rust"},
	}

	// 2. Marshal using version 2 code
	serialized, err := p2.Marshal(nil)
	if err != nil {
		t.Fatalf("v2 Marshal failed: %v", err)
	}

	// Verify expected binary layout:
	// - Bitmap (1 byte): 0x0F (fields 0, 1, 2, 3 present, MSB 0)
	// - Id (8 bytes): 1234567890 in Little-Endian -> 0xD2 0x02 0x96 0x49 0x00 0x00 0x00 0x00
	// - Uuid (2B length + 3B content): 0x03 0x00 'a' 'b' 'c' -> 0x03 0x00 0x61 0x62 0x63
	// - Active (1 byte): 0x01
	// - Tags (2B count + elements):
	//   - Count (2 bytes): 2 -> 0x02 0x00
	//   - Element 0 ("go"): 2B len + content -> 0x02 0x00 0x67 0x6f
	//   - Element 1 ("rust"): 2B len + content -> 0x04 0x00 0x72 0x75 0x73 0x74
	expectedBytes := []byte{
		0x0F,                                           // Bitmap
		0xD2, 0x02, 0x96, 0x49, 0x00, 0x00, 0x00, 0x00, // Id
		0x03, 0x00, 0x61, 0x62, 0x63,                   // Uuid
		0x01,                                           // Active
		0x02, 0x00,                                     // Tags Count
		0x02, 0x00, 0x67, 0x6F,                         // Tag 0 ("go")
		0x04, 0x00, 0x72, 0x75, 0x73, 0x74,             // Tag 1 ("rust")
	}

	if !bytes.Equal(serialized, expectedBytes) {
		t.Errorf("Wire format mismatch.\nGot:  %x\nWant: %x", serialized, expectedBytes)
	}

	// 3. Deserialize back using version 2 code
	var decoded2 testgen_v2.Payload
	n2, err := decoded2.Unmarshal(serialized)
	if err != nil {
		t.Fatalf("v2 Unmarshal failed: %v", err)
	}

	if n2 != len(serialized) {
		t.Errorf("v2 Unmarshal did not consume all bytes. Consumed: %d, Total: %d", n2, len(serialized))
	}

	if decoded2.Id != p2.Id || decoded2.Uuid != p2.Uuid || decoded2.Active != p2.Active {
		t.Errorf("v2 fields mismatch. Got %+v, Want %+v", decoded2, p2)
	}

	if len(decoded2.Tags) != 2 || decoded2.Tags[0] != "go" || decoded2.Tags[1] != "rust" {
		t.Errorf("v2 tags mismatch. Got %v", decoded2.Tags)
	}

	// 4. Backward Compatibility Test:
	// Deserialize the v2 serialized payload using v1 code.
	// v1 code should read the bitmap, parse the fields it recognizes (id, uuid, active),
	// and then safely stop/exit without throwing any errors or being affected by the appended 'tags' field.
	var decoded1 testgen_v1.Payload
	n1, err := decoded1.Unmarshal(serialized)
	if err != nil {
		t.Fatalf("v1 Unmarshal failed: %v", err)
	}

	// v1 should consume bytes exactly up to the end of 'active' field.
	// Total bytes up to 'active' field:
	// - 1B bitmap + 8B Id + 5B Uuid + 1B Active = 15 bytes.
	expectedConsumedV1 := 1 + 8 + 5 + 1
	if n1 != expectedConsumedV1 {
		t.Errorf("v1 consumed byte mismatch. Got %d, Want %d", n1, expectedConsumedV1)
	}

	if decoded1.Id != p2.Id || decoded1.Uuid != p2.Uuid || decoded1.Active != p2.Active {
		t.Errorf("v1 backward compatible fields mismatch. Got %+v, Want %+v", decoded1, p2)
	}
}

func TestGeoStruct(t *testing.T) {
	g := testgen_v2.Geo{
		Lat: 37.7749,
		Lng: -122.4194,
	}
	serialized, err := g.Marshal(nil)
	if err != nil {
		t.Fatalf("Geo Marshal failed: %v", err)
	}

	// Expected size of Geo struct:
	// - Bitmap (1 byte): 0x03 (lat, lng present)
	// - Lat (8 bytes): float64
	// - Lng (8 bytes): float64
	// Total: 17 bytes
	if len(serialized) != 17 {
		t.Errorf("expected Geo serialized length to be 17, got %d", len(serialized))
	}

	var decoded testgen_v2.Geo
	_, err = decoded.Unmarshal(serialized)
	if err != nil {
		t.Fatalf("Geo unmarshal failed: %v", err)
	}

	if decoded.Lat != g.Lat || decoded.Lng != g.Lng {
		t.Errorf("Geo mismatch. Got %+v, Want %+v", decoded, g)
	}
}
