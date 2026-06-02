package parser

import (
	"fmt"
	"github.com/snowmerak/antiserial/compiler/ast"
)

// Parser parses tokens into an AST.
type Parser struct {
	lexer *Lexer
	curr  Token
}

// NewParser creates a new Parser instance.
func NewParser(src string) *Parser {
	p := &Parser{
		lexer: NewLexer(src),
	}
	p.advance()
	return p
}

// advance fetches the next token from the lexer.
func (p *Parser) advance() {
	p.curr = p.lexer.NextToken()
}

// Parse runs the parser and returns the completed AST.
func (p *Parser) Parse() (ast.AST, error) {
	var structs []ast.Struct
	structNames := make(map[string]bool)

	for p.curr.Type != TokenEOF {
		if p.curr.Type == TokenError {
			return ast.AST{}, fmt.Errorf("lexical error at line %d, col %d: %s", p.curr.Line, p.curr.Col, p.curr.Value)
		}

		if p.curr.Type != TokenKeywordStruct {
			return ast.AST{}, fmt.Errorf("line %d, col %d: expected 'struct' keyword, got %q", p.curr.Line, p.curr.Col, p.curr.Value)
		}
		p.advance()

		if p.curr.Type != TokenIdentifier {
			return ast.AST{}, fmt.Errorf("line %d, col %d: expected struct name, got %q", p.curr.Line, p.curr.Col, p.curr.Value)
		}
		structName := p.curr.Value
		p.advance()

		if structNames[structName] {
			return ast.AST{}, fmt.Errorf("duplicate struct definition: %s", structName)
		}
		structNames[structName] = true

		if p.curr.Type != TokenLeftBrace {
			return ast.AST{}, fmt.Errorf("line %d, col %d: expected '{' after struct name, got %q", p.curr.Line, p.curr.Col, p.curr.Value)
		}
		p.advance()

		var fields []ast.Field
		fieldNames := make(map[string]bool)

		for p.curr.Type != TokenRightBrace && p.curr.Type != TokenEOF {
			if p.curr.Type == TokenError {
				return ast.AST{}, fmt.Errorf("lexical error at line %d, col %d: %s", p.curr.Line, p.curr.Col, p.curr.Value)
			}

			if p.curr.Type != TokenIdentifier {
				return ast.AST{}, fmt.Errorf("line %d, col %d: expected field name, got %q", p.curr.Line, p.curr.Col, p.curr.Value)
			}
			fieldName := p.curr.Value
			p.advance()

			if fieldNames[fieldName] {
				return ast.AST{}, fmt.Errorf("duplicate field %q in struct %q", fieldName, structName)
			}
			fieldNames[fieldName] = true

			if p.curr.Type != TokenColon {
				return ast.AST{}, fmt.Errorf("line %d, col %d: expected ':' after field name, got %q", p.curr.Line, p.curr.Col, p.curr.Value)
			}
			p.advance()

			fType, err := p.parseType()
			if err != nil {
				return ast.AST{}, err
			}

			if p.curr.Type != TokenSemicolon {
				return ast.AST{}, fmt.Errorf("line %d, col %d: expected ';' after field type, got %q", p.curr.Line, p.curr.Col, p.curr.Value)
			}
			p.advance()

			fields = append(fields, ast.Field{
				Name: fieldName,
				Type: fType,
			})
		}

		if p.curr.Type != TokenRightBrace {
			return ast.AST{}, fmt.Errorf("line %d, col %d: expected '}' at the end of struct %q", p.curr.Line, p.curr.Col, structName)
		}
		p.advance()

		structs = append(structs, ast.Struct{
			Name:   structName,
			Fields: fields,
		})
	}

	resultAST := ast.AST{Structs: structs}

	// Semantic Analysis: Resolve and validate type references
	for i := range resultAST.Structs {
		for j := range resultAST.Structs[i].Fields {
			err := resolveTypes(&resultAST.Structs[i].Fields[j].Type, structNames)
			if err != nil {
				return ast.AST{}, fmt.Errorf("semantic error in struct %s field %s: %w", resultAST.Structs[i].Name, resultAST.Structs[i].Fields[j].Name, err)
			}
		}
	}

	return resultAST, nil
}

// parseType parses field types including nested generics.
func (p *Parser) parseType() (ast.FieldType, error) {
	optional := false
	if p.curr.Type == TokenKeywordOptional {
		optional = true
		p.advance()
	}

	ft, err := p.parseTypeCore()
	if err != nil {
		return ast.FieldType{}, err
	}
	ft.Optional = optional
	return ft, nil
}

func (p *Parser) parseTypeCore() (ast.FieldType, error) {
	switch p.curr.Type {
	case TokenIdentifier:
		name := p.curr.Value
		p.advance()
		return ast.FieldType{
			Kind: ast.TypePrimitive, // Resolved post-parsing
			Name: name,
		}, nil

	case TokenKeywordList:
		p.advance()
		if p.curr.Type != TokenLeftAngle {
			return ast.FieldType{}, fmt.Errorf("line %d, col %d: expected '<' after list, got %q", p.curr.Line, p.curr.Col, p.curr.Value)
		}
		p.advance()

		elemType, err := p.parseTypeCore()
		if err != nil {
			return ast.FieldType{}, err
		}

		if p.curr.Type != TokenRightAngle {
			return ast.FieldType{}, fmt.Errorf("line %d, col %d: expected '>' after list element type, got %q", p.curr.Line, p.curr.Col, p.curr.Value)
		}
		p.advance()

		return ast.FieldType{
			Kind:     ast.TypeList,
			ElemType: &elemType,
		}, nil

	case TokenKeywordMap:
		p.advance()
		if p.curr.Type != TokenLeftAngle {
			return ast.FieldType{}, fmt.Errorf("line %d, col %d: expected '<' after map, got %q", p.curr.Line, p.curr.Col, p.curr.Value)
		}
		p.advance()

		keyType, err := p.parseTypeCore()
		if err != nil {
			return ast.FieldType{}, err
		}

		if p.curr.Type != TokenComma {
			return ast.FieldType{}, fmt.Errorf("line %d, col %d: expected ',' between map key and value types, got %q", p.curr.Line, p.curr.Col, p.curr.Value)
		}
		p.advance()

		valType, err := p.parseTypeCore()
		if err != nil {
			return ast.FieldType{}, err
		}

		if p.curr.Type != TokenRightAngle {
			return ast.FieldType{}, fmt.Errorf("line %d, col %d: expected '>' after map value type, got %q", p.curr.Line, p.curr.Col, p.curr.Value)
		}
		p.advance()

		return ast.FieldType{
			Kind:    ast.TypeMap,
			KeyType: &keyType,
			ValType: &valType,
		}, nil

	default:
		return ast.FieldType{}, fmt.Errorf("line %d, col %d: expected type identifier, 'optional', 'list', or 'map', got %q", p.curr.Line, p.curr.Col, p.curr.Value)
	}
}

// resolveTypes classifies simple identifiers as primitives or user-defined structs.
func resolveTypes(ft *ast.FieldType, structNames map[string]bool) error {
	if ft.Kind == ast.TypePrimitive {
		switch ft.Name {
		case "bool", "int32", "uint32", "int64", "uint64", "float32", "float64", "float", "double", "string", "bytes":
			return nil
		default:
			if structNames[ft.Name] {
				ft.Kind = ast.TypeStruct
				return nil
			}
			return fmt.Errorf("undefined type %q", ft.Name)
		}
	} else if ft.Kind == ast.TypeList {
		return resolveTypes(ft.ElemType, structNames)
	} else if ft.Kind == ast.TypeMap {
		if err := resolveTypes(ft.KeyType, structNames); err != nil {
			return err
		}
		if ft.KeyType.Kind != ast.TypePrimitive {
			return fmt.Errorf("map key type must be a primitive type, got %s", ft.KeyType.String())
		}
		return resolveTypes(ft.ValType, structNames)
	}
	return nil
}
