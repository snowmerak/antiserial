package test

import (
	"encoding/json"
	"testing"

	"github.com/apache/fory/go/fory"
	"github.com/bytedance/sonic"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/snowmerak/antiserial/test/testgen_fbs/fbs"
	"github.com/snowmerak/antiserial/test/testgen_pb"
	"github.com/snowmerak/antiserial/test/testgen_v1"
	"github.com/snowmerak/antiserial/test/testgen_v2"
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/protobuf/proto"
)

// Global Fory instance initialized once for the benchmark
var foryInstance = func() *fory.Fory {
	f := fory.New()
	if err := f.RegisterStruct(testgen_v2.Payload{}, 1); err != nil {
		panic(err)
	}
	return f
}()

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

// === AntiSerial Benchmarks ===

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

// === Apache Fory Benchmarks ===

func BenchmarkForyMarshal(b *testing.B) {
	p := testgen_v2.Payload{
		Id:     1234567890,
		Uuid:   "abc",
		Active: true,
		Tags:   []string{"go", "rust"},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = foryInstance.Serialize(&p)
	}
}

func BenchmarkForyUnmarshal(b *testing.B) {
	p := testgen_v2.Payload{
		Id:     1234567890,
		Uuid:   "abc",
		Active: true,
		Tags:   []string{"go", "rust"},
	}
	serialized, _ := foryInstance.Serialize(&p)
	var decoded testgen_v2.Payload

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = foryInstance.Deserialize(serialized, &decoded)
	}
}

// === MessagePack Benchmarks ===

func BenchmarkMessagePackMarshal(b *testing.B) {
	p := testgen_v2.Payload{
		Id:     1234567890,
		Uuid:   "abc",
		Active: true,
		Tags:   []string{"go", "rust"},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = msgpack.Marshal(&p)
	}
}

func BenchmarkMessagePackUnmarshal(b *testing.B) {
	p := testgen_v2.Payload{
		Id:     1234567890,
		Uuid:   "abc",
		Active: true,
		Tags:   []string{"go", "rust"},
	}
	serialized, _ := msgpack.Marshal(&p)
	var decoded testgen_v2.Payload

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = msgpack.Unmarshal(serialized, &decoded)
	}
}

// === Bytedance Sonic Benchmarks ===

func BenchmarkSonicMarshal(b *testing.B) {
	p := testgen_v2.Payload{
		Id:     1234567890,
		Uuid:   "abc",
		Active: true,
		Tags:   []string{"go", "rust"},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = sonic.Marshal(&p)
	}
}

func BenchmarkSonicUnmarshal(b *testing.B) {
	p := testgen_v2.Payload{
		Id:     1234567890,
		Uuid:   "abc",
		Active: true,
		Tags:   []string{"go", "rust"},
	}
	serialized, _ := sonic.Marshal(&p)
	var decoded testgen_v2.Payload

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = sonic.Unmarshal(serialized, &decoded)
	}
}

// === Protobuf Benchmarks ===

func BenchmarkProtobufMarshal(b *testing.B) {
	p := &testgen_pb.Payload{
		Id:     1234567890,
		Uuid:   "abc",
		Active: true,
		Tags:   []string{"go", "rust"},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(p)
	}
}

func BenchmarkProtobufUnmarshal(b *testing.B) {
	p := &testgen_pb.Payload{
		Id:     1234567890,
		Uuid:   "abc",
		Active: true,
		Tags:   []string{"go", "rust"},
	}
	serialized, _ := proto.Marshal(p)
	var decoded testgen_pb.Payload

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = proto.Unmarshal(serialized, &decoded)
	}
}

// === FlatBuffers Benchmarks ===

func BenchmarkFlatBuffersMarshal(b *testing.B) {
	p := testgen_v2.Payload{
		Id:     1234567890,
		Uuid:   "abc",
		Active: true,
		Tags:   []string{"go", "rust"},
	}

	builder := flatbuffers.NewBuilder(1024)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		builder.Reset()

		// Build strings and vectors
		uuidOffset := builder.CreateString(p.Uuid)

		var tagOffsets []flatbuffers.UOffsetT
		for _, tag := range p.Tags {
			offset := builder.CreateString(tag)
			tagOffsets = append(tagOffsets, offset)
		}

		fbs.PayloadStartTagsVector(builder, len(tagOffsets))
		for j := len(tagOffsets) - 1; j >= 0; j-- {
			builder.PrependUOffsetT(tagOffsets[j])
		}
		tagsVecOffset := builder.EndVector(len(tagOffsets))

		// Build table
		fbs.PayloadStart(builder)
		fbs.PayloadAddId(builder, p.Id)
		fbs.PayloadAddUuid(builder, uuidOffset)
		fbs.PayloadAddActive(builder, p.Active)
		fbs.PayloadAddTags(builder, tagsVecOffset)
		payloadOffset := fbs.PayloadEnd(builder)

		builder.Finish(payloadOffset)
		_ = builder.FinishedBytes()
	}
}

func BenchmarkFlatBuffersUnmarshal(b *testing.B) {
	p := testgen_v2.Payload{
		Id:     1234567890,
		Uuid:   "abc",
		Active: true,
		Tags:   []string{"go", "rust"},
	}

	builder := flatbuffers.NewBuilder(1024)
	uuidOffset := builder.CreateString(p.Uuid)
	var tagOffsets []flatbuffers.UOffsetT
	for _, tag := range p.Tags {
		offset := builder.CreateString(tag)
		tagOffsets = append(tagOffsets, offset)
	}
	fbs.PayloadStartTagsVector(builder, len(tagOffsets))
	for j := len(tagOffsets) - 1; j >= 0; j-- {
		builder.PrependUOffsetT(tagOffsets[j])
	}
	tagsVecOffset := builder.EndVector(len(tagOffsets))

	fbs.PayloadStart(builder)
	fbs.PayloadAddId(builder, p.Id)
	fbs.PayloadAddUuid(builder, uuidOffset)
	fbs.PayloadAddActive(builder, p.Active)
	fbs.PayloadAddTags(builder, tagsVecOffset)
	payloadOffset := fbs.PayloadEnd(builder)
	builder.Finish(payloadOffset)
	serialized := builder.FinishedBytes()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Read FlatBuffer fields
		fbPayload := fbs.GetRootAsPayload(serialized, 0)
		_ = fbPayload.Id()
		_ = fbPayload.Uuid()
		_ = fbPayload.Active()
		tagsLen := fbPayload.TagsLength()
		for j := 0; j < tagsLen; j++ {
			_ = fbPayload.Tags(j)
		}
	}
}

// === JSON Benchmarks ===

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
