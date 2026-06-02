package ast

import "fmt"

// FieldTypeKind defines the category of a field type.
type FieldTypeKind int

const (
	TypePrimitive FieldTypeKind = iota
	TypeList
	TypeMap
	TypeStruct
)

// String returns the string representation of the FieldTypeKind.
func (k FieldTypeKind) String() string {
	switch k {
	case TypePrimitive:
		return "Primitive"
	case TypeList:
		return "List"
	case TypeMap:
		return "Map"
	case TypeStruct:
		return "Struct"
	default:
		return "Unknown"
	}
}

// FieldType represents a type in the schema.
type FieldType struct {
	Kind     FieldTypeKind
	Optional bool       // When true, bitmap controls presence; zero values are written when present
	Name     string     // "bool", "int32", "string", or struct name
	ElemType *FieldType // For list
	KeyType  *FieldType // For map key
	ValType  *FieldType // For map value
}

// String returns a human-readable representation of the type.
func (t FieldType) String() string {
	prefix := ""
	if t.Optional {
		prefix = "optional "
	}
	switch t.Kind {
	case TypePrimitive, TypeStruct:
		return prefix + t.Name
	case TypeList:
		return prefix + fmt.Sprintf("list<%s>", t.ElemType.String())
	case TypeMap:
		return prefix + fmt.Sprintf("map<%s, %s>", t.KeyType.String(), t.ValType.String())
	default:
		return "unknown"
	}
}

// IsPrimitiveFixed returns true if the type is a primitive fixed-size type.
func (t FieldType) IsPrimitiveFixed() bool {
	if t.Kind != TypePrimitive {
		return false
	}
	switch t.Name {
	case "bool", "int32", "uint32", "int64", "uint64", "float32", "float64", "float", "double":
		return true
	}
	return false
}

// PrimitiveFixedSize returns the byte size of a fixed-size primitive.
func (t FieldType) PrimitiveFixedSize() int {
	switch t.Name {
	case "bool":
		return 1
	case "int32", "uint32", "float32", "float":
		return 4
	case "int64", "uint64", "float64", "double":
		return 8
	default:
		return 0
	}
}

// Field represents a struct field definition.
type Field struct {
	Name string
	Type FieldType
}

// Struct represents a parsed struct.
type Struct struct {
	Name   string
	Fields []Field
}

// IsFixedSize returns true if the struct only contains fixed-size types (primitives or other fixed structs).
func (s Struct) IsFixedSize(ast *AST) bool {
	for _, field := range s.Fields {
		if !field.Type.IsFixedSize(ast) {
			return false
		}
	}
	return true
}

// FixedSize returns the cumulative byte size of the struct's fields if it's fixed size.
func (s Struct) FixedSize(ast *AST) (int, bool) {
	size := 0
	for _, field := range s.Fields {
		fieldSize, isFixed := field.Type.FixedSize(ast)
		if !isFixed {
			return 0, false
		}
		size += fieldSize
	}
	return size, true
}

// IsFixedSize returns true if the type is fixed-size (primitive fixed or a fixed struct).
func (t FieldType) IsFixedSize(ast *AST) bool {
	if t.IsPrimitiveFixed() {
		return true
	}
	if t.Kind == TypeStruct {
		if s, ok := ast.FindStruct(t.Name); ok {
			return s.IsFixedSize(ast)
		}
	}
	return false
}

// FixedSize returns the size and true if the type is fixed size, otherwise 0 and false.
func (t FieldType) FixedSize(ast *AST) (int, bool) {
	if t.IsPrimitiveFixed() {
		return t.PrimitiveFixedSize(), true
	}
	if t.Kind == TypeStruct {
		if s, ok := ast.FindStruct(t.Name); ok {
			return s.FixedSize(ast)
		}
	}
	return 0, false
}

// AST is the top-level schema AST.
type AST struct {
	Structs []Struct
}

// FindStruct searches for a struct by name in the AST.
func (a *AST) FindStruct(name string) (Struct, bool) {
	for _, s := range a.Structs {
		if s.Name == name {
			return s, true
		}
	}
	return Struct{}, false
}
