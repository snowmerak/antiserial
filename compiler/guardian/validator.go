package guardian

import (
	"fmt"
	"github.com/snowmerak/antiserial/compiler/ast"
)

// normalizePrimitiveName maps wire-compatible primitive aliases to a canonical name.
func normalizePrimitiveName(name string) string {
	switch name {
	case "float":
		return "float32"
	case "double":
		return "float64"
	default:
		return name
	}
}

// TypesEqual returns true if two AST field types are structurally identical.
func TypesEqual(t1, t2 ast.FieldType) bool {
	if t1.Optional != t2.Optional {
		return false
	}
	if t1.Kind != t2.Kind {
		return false
	}
	switch t1.Kind {
	case ast.TypePrimitive:
		return normalizePrimitiveName(t1.Name) == normalizePrimitiveName(t2.Name)
	case ast.TypeStruct:
		return t1.Name == t2.Name
	case ast.TypeList:
		if t1.ElemType == nil || t2.ElemType == nil {
			return t1.ElemType == t2.ElemType
		}
		return TypesEqual(*t1.ElemType, *t2.ElemType)
	case ast.TypeMap:
		if t1.KeyType == nil || t2.KeyType == nil || t1.ValType == nil || t2.ValType == nil {
			return false
		}
		return TypesEqual(*t1.KeyType, *t2.KeyType) && TypesEqual(*t1.ValType, *t2.ValType)
	}
	return false
}

// ValidateSchemaEvolution validates that the current schema is backward compatible with the base schema.
// It enforces the Append-Only rule for all structs defined in the base schema.
func ValidateSchemaEvolution(baseAST, currentAST *ast.AST) error {
	for _, baseStruct := range baseAST.Structs {
		currentStruct, exists := currentAST.FindStruct(baseStruct.Name)
		if !exists {
			return fmt.Errorf("backward compatibility error: struct %s was deleted", baseStruct.Name)
		}

		for i, baseField := range baseStruct.Fields {
			if i >= len(currentStruct.Fields) {
				return fmt.Errorf("backward compatibility error: struct %s field %q was deleted", baseStruct.Name, baseField.Name)
			}
			currentField := currentStruct.Fields[i]
			if baseField.Name != currentField.Name {
				return fmt.Errorf("backward compatibility error: struct %s field %d name changed from %q to %q (reordering or inserting in the middle is prohibited)", baseStruct.Name, i, baseField.Name, currentField.Name)
			}
			if !TypesEqual(baseField.Type, currentField.Type) {
				return fmt.Errorf("backward compatibility error: struct %s field %q type changed from %q to %q", baseStruct.Name, baseField.Name, baseField.Type.String(), currentField.Type.String())
			}
		}
	}
	return nil
}
