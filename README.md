# AntiSerial (As)

AntiSerial (As) is a compile-time static, tagless binary serialization protocol designed for zero-allocation memory operations, optimized binary size, and high-performance serialization/deserialization. 

This repository implements the AntiSerial compiler (`asc`), a schema compatibility validator (Schema Guardian), and code generation engines for Go, Rust, and C++.

---

## Key Features

1. **Tagless Serialization**: No individual field numbers or metadata tags are recorded in the stream. Fields are sequentially serialized based on a statically agreed compile-time order.
2. **Centralized Varint-like Bitmap Header**: A compact header at the beginning of the payload uses continuation bits (MSB) to record the presence of fields. Primitives and collections that are omitted (null or default-valued) consume 0 bytes of payload.
3. **No Memory Padding**: All fields are packed tightly without padding bytes, ensuring the smallest possible wire size.
4. **Zero-Copy Projection**: Deserialization maps variable-length fields (strings and byte slices) directly to input buffer pointers (e.g., `unsafe.String` in Go, references with lifetime `'a` in Rust, `std::string_view` / `std::span` in C++) avoiding heap allocations.
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

| AntiSerial Type | Go Type | Rust Type | C++ Type | Size (Bytes) |
| :--- | :--- | :--- | :--- | :---: |
| **bool** | `bool` | `bool` | `bool` | 1 |
| **int32** | `int32` | `i32` | `int32_t` | 4 |
| **uint32** | `uint32` | `u32` | `uint32_t` | 4 |
| **int64** | `int64` | `i64` | `int64_t` | 8 |
| **uint64** | `uint64` | `u64` | `uint64_t` | 8 |
| **float32 / float** | `float32` | `f32` | `float` | 4 |
| **float64 / double**| `float64` | `f64` | `double` | 8 |
| **string** | `string` | `&'a str` | `std::string_view` | 2 (len) + variable |
| **bytes** | `[]byte` | `&'a [u8]` | `std::span<const uint8_t>` | 4 (len) + variable |
| **list\<T\>** | `[]T` | `Vec<T>` | `std::vector<T>` | 2 (len) + variable |
| **map\<K,V\>** | `map[K]V` | `HashMap<K, V>` | `std::unordered_map<K, V>`| 2 (len) + variable |

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
  --base_schema=<file>   Path to the base schema file to validate backward compatibility
  --validate_only        Perform validation only without generating code
```

### Example Evolution Validation

To guarantee schema safety before deploying in a CI/CD pipeline:

```bash
asc --base_schema=schema.base.as --validate_only schema.as
```

---

## Verification and Testing

To run the parser, validator, and Go integration/zero-copy tests:

```bash
# Run unit tests
go test ./compiler/...

# Run end-to-end integration tests
go test ./test/...

# Run benchmark suite
go test -bench=Benchmark ./test/
```

### Benchmark Comparison (AntiSerial vs JSON)

Benchmarks executed on Go 1.20+, arm64 system showing direct comparative performance:

* **Marshal Performance**:
  * **AntiSerial Marshal**: **17.96 ns/op** | **0 B/op** | **0 allocs/op** (15.5x faster than JSON)
  * **JSON Marshal**: **279.30 ns/op** | **80 B/op** | **1 allocs/op**
* **Unmarshal Performance**:
  * **AntiSerial Unmarshal**: **44.23 ns/op** | **32 B/op** | **1 allocs/op** (30.2x faster than JSON)
  * **JSON Unmarshal**: **1337.00 ns/op** | **248 B/op** | **8 allocs/op**

