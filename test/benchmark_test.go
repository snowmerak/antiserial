package test

import (
	"encoding/json"
	"testing"

	"github.com/snowmerak/antiserial/test/testgen_v1"
	"github.com/snowmerak/antiserial/test/testgen_v2"
)

// TestUnmarshalZeroAllocations verifies that unmarshaling a payload with no dynamic collections
// causes exactly 0 heap allocations, demonstrating the effectiveness of unsafe string projections.
func TestUnmarshalZeroAllocations(t *testing.T) {
	p := testgen_v1.Payload{
		Id:     1234567890,
		Uuid:   "abc",
		Active: true,
	}
	serialized := p.Marshal(nil)

	var decoded testgen_v1.Payload
	allocs := testing.AllocsPerRun(1000, func() {
		_, err := decoded.Unmarshal(serialized)
		if err != nil {
			t.Fatal(err)
		}
	})

	if allocs > 0 {
		t.Errorf("expected 0 allocations, got %f", allocs)
	}
}

// BenchmarkAntiSerialMarshal measures serialization throughput and memory allocation.
func BenchmarkAntiSerialMarshal(b *testing.B) {
	p := testgen_v2.Payload{
		Id:     1234567890,
		Uuid:   "abc",
		Active: true,
		Tags:   []string{"go", "rust"},
	}
	buf := make([]byte, 0, 1024)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = p.Marshal(buf[:0])
	}
}

// BenchmarkAntiSerialUnmarshal measures deserialization throughput and memory allocation.
func BenchmarkAntiSerialUnmarshal(b *testing.B) {
	p := testgen_v2.Payload{
		Id:     1234567890,
		Uuid:   "abc",
		Active: true,
		Tags:   []string{"go", "rust"},
	}
	serialized := p.Marshal(nil)
	var decoded testgen_v2.Payload

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = decoded.Unmarshal(serialized)
	}
}

// BenchmarkJSONMarshal measures encoding/json serialization.
func BenchmarkJSONMarshal(b *testing.B) {
	p := testgen_v2.Payload{
		Id:     1234567890,
		Uuid:   "abc",
		Active: true,
		Tags:   []string{"go", "rust"},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(&p)
	}
}

// BenchmarkJSONUnmarshal measures encoding/json deserialization.
func BenchmarkJSONUnmarshal(b *testing.B) {
	p := testgen_v2.Payload{
		Id:     1234567890,
		Uuid:   "abc",
		Active: true,
		Tags:   []string{"go", "rust"},
	}
	serialized, _ := json.Marshal(&p)
	var decoded testgen_v2.Payload

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = json.Unmarshal(serialized, &decoded)
	}
}
