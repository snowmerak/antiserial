package cpp

import (
	"bytes"
	"fmt"
	"github.com/snowmerak/antiserial/compiler/ast"
	"github.com/snowmerak/antiserial/compiler/codegen"
)

// Generate translates the AST into a high-performance, header-only C++ library.
func Generate(schemaAST ast.AST) (string, error) {
	var buf bytes.Buffer

	// Header guards and includes
	buf.WriteString(`#pragma once
#include <vector>
#include <string_view>
#include <unordered_map>
#include <span>
#include <cstdint>
#include <cstring>
#include <stdexcept>

`)

	// Forward declare all structs
	buf.WriteString("// Forward Declarations\n")
	for _, s := range schemaAST.Structs {
		buf.WriteString(fmt.Sprintf("struct %s;\n", s.Name))
	}
	buf.WriteString("\n")

	// Generate Struct definitions and inline methods
	for _, s := range schemaAST.Structs {
		// Struct definition
		buf.WriteString(fmt.Sprintf("struct %s {\n", s.Name))
		for _, field := range s.Fields {
			buf.WriteString(fmt.Sprintf("    %s %s{};\n", toCppType(field.Type), field.Name))
		}

		// Serialize Method
		buf.WriteString("\n    void serialize(std::vector<uint8_t>& buf) const {\n")
		// Check presence variables
		for i, field := range s.Fields {
			presenceExpr := genPresenceCheck("this->"+field.Name, field.Type)
			buf.WriteString(fmt.Sprintf("        bool f%d_present = %s;\n", i, presenceExpr))
		}

		numFields := len(s.Fields)
		numBitmapBytes := (numFields + 6) / 7
		if numBitmapBytes == 0 {
			numBitmapBytes = 1
		}

		// Write bitmap bytes
		for b := 0; b < numBitmapBytes; b++ {
			buf.WriteString(fmt.Sprintf("        uint8_t b%d = 0;\n", b))
			for bit := 0; bit < 7; bit++ {
				fieldIdx := b*7 + bit
				if fieldIdx < numFields {
					buf.WriteString(fmt.Sprintf("        if (f%d_present) {\n            b%d |= 1 << %d;\n        }\n", fieldIdx, b, bit))
				}
			}
			if b < numBitmapBytes-1 {
				buf.WriteString(fmt.Sprintf("        b%d |= 1 << 7;\n", b))
			}
			buf.WriteString(fmt.Sprintf("        buf.push_back(b%d);\n", b))
		}

		// Serialize present fields
		for i, field := range s.Fields {
			buf.WriteString(fmt.Sprintf("        if (f%d_present) {\n", i))
			serializeCode := genSerializeType(field.Type, "this->"+field.Name, 0)
			buf.WriteString(indent(serializeCode, "            "))
			buf.WriteString("\n        }\n")
		}
		buf.WriteString("    }\n")

		// Deserialize Method
		buf.WriteString("\n    void deserialize(const uint8_t* buf, size_t size, size_t& offset) {\n")
		buf.WriteString(`        size_t bitmap_start = offset;
        while (true) {
            if (offset >= size) {
                throw std::runtime_error("Unexpected EOF reading bitmap");
            }
            uint8_t b = buf[offset];
            offset++;
            if ((b & 0x80) == 0) {
                break;
            }
        }
        const uint8_t* bitmap_bytes = &buf[bitmap_start];
        size_t bitmap_len = offset - bitmap_start;

        auto is_present = [&](size_t field_idx) -> bool {
            size_t byte_idx = field_idx / 7;
            size_t bit_idx = field_idx % 7;
            if (byte_idx >= bitmap_len) {
                return false;
            }
            return (bitmap_bytes[byte_idx] & (1 << bit_idx)) != 0;
        };
`)

		for i, field := range s.Fields {
			buf.WriteString(fmt.Sprintf("\n        if (is_present(%d)) {\n", i))
			deserializeCode := genDeserializeType(field.Type, "this->"+field.Name, 0)
			buf.WriteString(indent(deserializeCode, "            "))
			buf.WriteString("\n        }\n")
		}
		buf.WriteString("    }\n};\n\n")
	}

	return buf.String(), nil
}

// toCppType maps an AST type to a C++ type.
func toCppType(t ast.FieldType) string {
	switch t.Kind {
	case ast.TypePrimitive:
		switch t.Name {
		case "bool":
			return "bool"
		case "int32":
			return "int32_t"
		case "uint32":
			return "uint32_t"
		case "int64":
			return "int64_t"
		case "uint64":
			return "uint64_t"
		case "float32", "float":
			return "float"
		case "float64", "double":
			return "double"
		case "string":
			return "std::string_view"
		case "bytes":
			return "std::span<const uint8_t>"
		default:
			return t.Name
		}
	case ast.TypeStruct:
		return t.Name
	case ast.TypeList:
		return "std::vector<" + toCppType(*t.ElemType) + ">"
	case ast.TypeMap:
		return "std::unordered_map<" + toCppType(*t.KeyType) + ", " + toCppType(*t.ValType) + ">"
	default:
		return ""
	}
}

// genPresenceCheck checks if a field is present in C++.
func genPresenceCheck(expr string, t ast.FieldType) string {
	switch t.Kind {
	case ast.TypePrimitive:
		switch t.Name {
		case "bool":
			return expr
		case "int32", "uint32", "int64", "uint64":
			return expr + " != 0"
		case "float32", "float", "float64", "double":
			return expr + " != 0.0"
		case "string", "bytes":
			return "!" + expr + ".empty()"
		}
	case ast.TypeStruct:
		return "true"
	case ast.TypeList, ast.TypeMap:
		return "!" + expr + ".empty()"
	}
	return "false"
}

// genSerializeType recursively generates C++ serialization code.
func genSerializeType(t ast.FieldType, expr string, depth int) string {
	switch t.Kind {
	case ast.TypePrimitive:
		switch t.Name {
		case "bool":
			return fmt.Sprintf(`{
    uint8_t val = %s ? 1 : 0;
    buf.push_back(val);
}`, expr)
		case "int32", "uint32", "int64", "uint64", "float32", "float", "float64", "double":
			cppType := toCppType(t)
			return fmt.Sprintf(`{
    %s val = %s;
    size_t start = buf.size();
    buf.resize(start + sizeof(%s));
    std::memcpy(&buf[start], &val, sizeof(%s));
}`, cppType, expr, cppType, cppType)
		case "string":
			return fmt.Sprintf(`{
    if (%s.size() > %d) {
        throw std::runtime_error("string length exceeds uint16 maximum");
    }
    uint16_t length = static_cast<uint16_t>(%s.size());
    size_t start = buf.size();
    buf.resize(start + 2 + length);
    std::memcpy(&buf[start], &length, 2);
    if (length > 0) {
        std::memcpy(&buf[start + 2], %s.data(), length);
    }
}`, expr, codegen.MaxUint16, expr, expr)
		case "bytes":
			return fmt.Sprintf(`{
    if (%s.size() > UINT32_MAX) {
        throw std::runtime_error("bytes length exceeds uint32 maximum");
    }
    uint32_t length = static_cast<uint32_t>(%s.size());
    size_t start = buf.size();
    buf.resize(start + 4 + length);
    std::memcpy(&buf[start], &length, 4);
    if (length > 0) {
        std::memcpy(&buf[start + 4], %s.data(), length);
    }
}`, expr, expr, expr)
		}
	case ast.TypeStruct:
		return fmt.Sprintf("%s.serialize(buf);", expr)
	case ast.TypeList:
		elemName := fmt.Sprintf("elem%d", depth)
		elemCode := genSerializeType(*t.ElemType, elemName, depth+1)
		return fmt.Sprintf(`{
    if (%s.size() > %d) {
        throw std::runtime_error("list length exceeds uint16 maximum");
    }
    uint16_t count = static_cast<uint16_t>(%s.size());
    size_t start = buf.size();
    buf.resize(start + 2);
    std::memcpy(&buf[start], &count, 2);
    for (const auto& %s : %s) {
%s
    }
}`, expr, codegen.MaxUint16, expr, elemName, expr, indent(elemCode, "        "))
	case ast.TypeMap:
		keyName := fmt.Sprintf("k%d", depth)
		valName := fmt.Sprintf("v%d", depth)
		keyCode := genSerializeType(*t.KeyType, keyName, depth+1)
		valCode := genSerializeType(*t.ValType, valName, depth+1)
		return fmt.Sprintf(`{
    if (%s.size() > %d) {
        throw std::runtime_error("map length exceeds uint16 maximum");
    }
    uint16_t count = static_cast<uint16_t>(%s.size());
    size_t start = buf.size();
    buf.resize(start + 2);
    std::memcpy(&buf[start], &count, 2);
    for (const auto& [%s, %s] : %s) {
%s
%s
    }
}`, expr, codegen.MaxUint16, expr, keyName, valName, expr, indent(keyCode, "        "), indent(valCode, "        "))
	}
	return ""
}

// genDeserializeType recursively generates C++ deserialization code.
func genDeserializeType(t ast.FieldType, expr string, depth int) string {
	switch t.Kind {
	case ast.TypePrimitive:
		switch t.Name {
		case "bool":
			return fmt.Sprintf(`if (offset + 1 > size) {
    throw std::runtime_error("Unexpected EOF");
}
%s = buf[offset] != 0;
offset += 1;`, expr)
		case "int32", "uint32", "int64", "uint64", "float32", "float", "float64", "double":
			cppType := toCppType(t)
			return fmt.Sprintf(`if (offset + sizeof(%s) > size) {
    throw std::runtime_error("Unexpected EOF");
}
std::memcpy(&%s, &buf[offset], sizeof(%s));
offset += sizeof(%s);`, cppType, expr, cppType, cppType)
		case "string":
			return fmt.Sprintf(`if (offset + 2 > size) {
    throw std::runtime_error("Unexpected EOF");
}
{
    uint16_t length;
    std::memcpy(&length, &buf[offset], 2);
    offset += 2;
    if (offset + length > size) {
        throw std::runtime_error("Unexpected EOF");
    }
    if (length > 0) {
        %s = std::string_view(reinterpret_cast<const char*>(&buf[offset]), length);
        offset += length;
    } else {
        %s = std::string_view();
    }
}`, expr, expr)
		case "bytes":
			return fmt.Sprintf(`if (offset + 4 > size) {
    throw std::runtime_error("Unexpected EOF");
}
{
    uint32_t length;
    std::memcpy(&length, &buf[offset], 4);
    offset += 4;
    if (offset + length > size) {
        throw std::runtime_error("Unexpected EOF");
    }
    if (length > 0) {
        %s = std::span<const uint8_t>(&buf[offset], length);
        offset += length;
    } else {
        %s = std::span<const uint8_t>();
    }
}`, expr, expr)
		}
	case ast.TypeStruct:
		return fmt.Sprintf("%s.deserialize(buf, size, offset);", expr)
	case ast.TypeList:
		elemName := fmt.Sprintf("elem%d", depth)
		elemTypeStr := toCppType(*t.ElemType)
		elemCode := genDeserializeType(*t.ElemType, elemName, depth+1)
		return fmt.Sprintf(`if (offset + 2 > size) {
    throw std::runtime_error("Unexpected EOF");
}
{
    uint16_t count;
    std::memcpy(&count, &buf[offset], 2);
    offset += 2;
    %s.resize(count);
    for (size_t i = 0; i < count; i++) {
        %s %s{};
%s
        %s[i] = %s;
    }
}`, expr, elemTypeStr, elemName, indent(elemCode, "        "), expr, elemName)
	case ast.TypeMap:
		keyName := fmt.Sprintf("k%d", depth)
		valName := fmt.Sprintf("v%d", depth)
		keyTypeStr := toCppType(*t.KeyType)
		valTypeStr := toCppType(*t.ValType)
		keyCode := genDeserializeType(*t.KeyType, keyName, depth+1)
		valCode := genDeserializeType(*t.ValType, valName, depth+1)
		return fmt.Sprintf(`if (offset + 2 > size) {
    throw std::runtime_error("Unexpected EOF");
}
{
    uint16_t count;
    std::memcpy(&count, &buf[offset], 2);
    offset += 2;
    %s.clear();
    for (size_t i = 0; i < count; i++) {
        %s %s{};
%s
        %s %s{};
%s
        %s[%s] = %s;
    }
}`, expr, keyTypeStr, keyName, indent(keyCode, "        "), valTypeStr, valName, indent(valCode, "        "), expr, keyName, valName)
	}
	return ""
}

// indent prefixes each line of s with prefix.
func indent(s, prefix string) string {
	var buf bytes.Buffer
	lines := bytes.Split([]byte(s), []byte("\n"))
	for i, line := range lines {
		if i > 0 {
			buf.WriteByte('\n')
		}
		if len(line) > 0 {
			buf.WriteString(prefix)
			buf.Write(line)
		}
	}
	return buf.String()
}
