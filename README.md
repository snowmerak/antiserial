# AntiSerial (As)

AntiSerial (As) is a compile-time static, tagless binary serialization protocol designed for zero-allocation memory operations, optimized binary size, and high-performance serialization/deserialization. 

This repository implements the AntiSerial compiler (`asc`), a schema compatibility validator (Schema Guardian), and code generation engines for Go, Rust, C++, TypeScript, and Python.

---

## Key Features

1. **Tagless Serialization**: No individual field numbers or metadata tags are recorded in the stream. Fields are sequentially serialized based on a statically agreed compile-time order.
2. **Centralized Varint-like Bitmap Header**: A compact header at the beginning of the payload uses continuation bits (MSB) to record the presence of fields. Primitives and collections that are omitted (null or default-valued) consume 0 bytes of payload.
3. **No Memory Padding**: All fields are packed tightly without padding bytes, ensuring the smallest possible wire size.
4. **Zero-Copy Projection**: Deserialization maps variable-length fields (strings and byte slices) directly to input buffer pointers (e.g., `unsafe.String` in Go, references with lifetime `'a` in Rust, `std::string_view` / `std::span` in C++, `Uint8Array` subarrays in TypeScript, and `memoryview` projections in Python) avoiding intermediate heap allocations.
5. **Strict Backward Compatibility**: New fields can only be appended to the end of a struct (Append-Only rule). Older clients can safely decode upgraded payloads by parsing known fields and exiting early (End-of-Stream Cut-off), ignoring any trailing unknown fields.

---

## IDL Syntax Spec (`.as`)

AntiSerial uses a clean C-style structural interface definition language:

```protobuf
// Basic geometric structure
struct Geo {
    lat: float64;
    lng: float64;
}

// Complex payload demonstrating nested structs, lists, and maps
struct ComplexPayload {
    id: int64;
    uuid: string;
    age: int32;
    score: float32;
    ratio: float64;
    active: bool;
    data: bytes;
    tags: list<string>;
    geo_points: list<Geo>;
    attributes: map<string, string>;
}
```

### Type Mapping

| AntiSerial Type | Go Type | Rust Type | C++ Type | TS Type | Python Type | Size (Bytes) |
| :--- | :--- | :--- | :--- | :--- | :--- | :---: |
| **bool** | `bool` | `bool` | `bool` | `boolean` | `bool` | 1 |
| **int32** | `int32` | `i32` | `int32_t` | `number` | `int` | 4 |
| **uint32** | `uint32` | `u32` | `uint32_t` | `number` | `int` | 4 |
| **int64** | `int64` | `i64` | `int64_t` | `bigint` | `int` | 8 |
| **uint64** | `uint64` | `u64` | `uint64_t` | `bigint` | `int` | 8 |
| **float32 / float** | `float32` | `f32` | `float` | `number` | `float` | 4 |
| **float64 / double**| `float64` | `f64` | `double` | `number` | `float` | 8 |
| **string** | `string` | `&'a str` | `std::string_view` | `string` | `str` | 2 (len) + variable |
| **bytes** | `[]byte` | `&'a [u8]` | `std::span<const uint8_t>` | `Uint8Array` | `bytes` | 4 (len) + variable |
| **list\<T\>** | `[]T` | `Vec<T>` | `std::vector<T>` | `T[]` | `list[T]` | 2 (len) + variable |
| **map\<K,V\>** | `map[K]V` | `HashMap<K, V>` | `std::unordered_map<K, V>` | `Map<K, V>` | `dict[K, V]` | 2 (len) + variable |

### Wire Size Limits

Generated serializers enforce these limits at **marshal** time (overflow returns an error instead of truncating):

| Field kind | Length prefix | Maximum |
| :--- | :---: | :---: |
| `string`, `list`, `map` | `uint16` (2 bytes, LE) | **65,535** |
| `bytes` | `uint32` (4 bytes, LE) | **4,294,967,295** |

### Presence Semantics

**Implicit (default) fields** use zero / empty as “not on the wire”: `0`, `0.0`, `""`, empty `bytes`/`list`/`map`, and `false` for `bool`. You cannot transmit an explicit zero or false with these fields alone.

**`optional` fields** use the bitmap only: when the bit is set, the value is written **including** zeros, `false`, and empty collections. In Go this maps to pointers (`*int32`, `*bool`, …); `nil` means absent.

```protobuf
struct Payload {
    id: int64;              // implicit: 0 is omitted
    score: optional int32;  // Some(0) is written as four zero bytes
}
```

Non-optional nested structs are **always** serialized (empty inner bitmap included) so all languages agree on the wire layout.

Go example (requires Go 1.26+ value `new`):

```go
p := Payload{
    Id:    1,
    Score: new(int32(0)), // present on the wire with value 0
}
buf, err := p.Marshal(nil)
```

### Schema Guardian Type Aliases

Backward-compatibility checks treat wire-equivalent primitive aliases as the same type: `float` ↔ `float32`, `double` ↔ `float64`.

---

## Compiler Usage (`asc`)

Compile schema files into targeted source code and validate backward compatibility.

```bash
# Build compiler (optional, or run directly via go run)
go build -o asc ./cmd/asc

# Options and usage
asc [options] <schema_file.as>

Options:
  --go_out=<dir>         Directory for generated Go source code
  --rust_out=<dir>       Directory for generated Rust source code
  --cpp_out=<dir>        Directory for generated C++ header source code
  --ts_out=<dir>         Directory for generated TypeScript source code
  --py_out=<dir>         Directory for generated Python source code
  --base_schema=<file>   Path to the base schema file to validate backward compatibility
  --validate_only        Perform validation only without generating code
```

### Example Evolution Validation

To guarantee schema safety before deploying in a CI/CD pipeline:

```bash
asc --base_schema=schema.base.as --validate_only schema.as
```

### Generated Go API

Go structs expose:

```go
func (s *Payload) Marshal(buf []byte) ([]byte, error)
func (s *Payload) Unmarshal(buf []byte) (int, error)
```

`Marshal` appends to `buf` (pass `nil` to allocate) and returns an error if any `string`/`list`/`map` length exceeds 65,535 or `bytes` exceeds the `uint32` limit.

---

## Verification and Testing

This repo uses [Task](https://taskfile.dev). Install Task, then from the repository root:

```bash
# Build asc and run all unit + multi-language E2E tests
task test:all

# Individual targets
task build          # asc.exe
task codegen        # refresh test/testgen_* and language fixtures from .as files
task test:unit      # parser + guardian
task test:go        # Go integration, limits, nested struct, optional, zero-copy
task test:golden    # same bytes decoded by Go, Python, Rust, TypeScript, C++
task test:rust      # Rust wire-format E2E (cargo run)
task test:ts        # Deno: test.ts + test_limits.ts
task test:py        # Python: test.py + test_limits.py + test_nested.py
task test:cpp       # C++ wire-format E2E (CMake + ctest; requires a C++17 toolchain)
task bench          # comparative benchmarks (in test/)
```

Equivalent raw commands:

```bash
go test ./compiler/...
go test -v github.com/snowmerak/antiserial/test
go test -bench=Benchmark ./test/
```

After changing `.as` schemas or codegen, run `task codegen` before committing generated sources under `test/`.

Cross-language wire compatibility is checked against `test/golden/payload_v2.bin`: Go marshaling must match the file, and Python/Rust/TypeScript/C++ decoders must read the same bytes to identical field values.

### Benchmark Comparison (AntiSerial vs JSON vs Protobuf vs FlatBuffers vs Apache Fory vs MessagePack vs Sonic)

Benchmarks executed on Go 1.26+, arm64 system showing direct comparative performance:

* **Marshal Performance**:
  * **AntiSerial Marshal**: **11.92 ns/op** | **0 B/op** | **0 allocs/op** (9.5x faster than Fory, 9.9x faster than Protobuf, 10.8x faster than FlatBuffers, 16.1x faster than JSON, 20.6x faster than MessagePack, 28.0x faster than Sonic)
  * **Apache Fory Marshal**: **113.00 ns/op** | **0 B/op** | **0 allocs/op**
  * **Protobuf Marshal**: **118.30 ns/op** | **24 B/op** | **1 allocs/op**
  * **FlatBuffers Marshal**: **128.20 ns/op** | **8 B/op** | **1 allocs/op**
  * **JSON Marshal**: **191.50 ns/op** | **80 B/op** | **1 allocs/op**
  * **MessagePack Marshal**: **245.60 ns/op** | **136 B/op** | **3 allocs/op**
  * **Bytedance Sonic Marshal**: **333.50 ns/op** | **98 B/op** | **2 allocs/op**

* **Unmarshal Performance**:
  * **AntiSerial Unmarshal**: **32.00 ns/op** | **32 B/op** | **1 allocs/op** (1.03x faster than FlatBuffers, 5.1x faster than Fory, 5.6x faster than Protobuf, 10.9x faster than MessagePack, 10.9x faster than Sonic, 28.1x faster than JSON)
  * **FlatBuffers Unmarshal**: **32.87 ns/op** | **0 B/op** | **0 allocs/op**
  * **Apache Fory Unmarshal**: **162.40 ns/op** | **44 B/op** | **4 allocs/op**
  * **Protobuf Unmarshal**: **179.70 ns/op** | **60 B/op** | **5 allocs/op**
  * **Bytedance Sonic Unmarshal**: **347.90 ns/op** | **232 B/op** | **3 allocs/op**
  * **MessagePack Unmarshal**: **348.40 ns/op** | **60 B/op** | **4 allocs/op**
  * **JSON Unmarshal**: **899.10 ns/op** | **248 B/op** | **8 allocs/op**


