package guardian

import (
	"testing"
	"github.com/snowmerak/antiserial/compiler/parser"
)

func TestValidationSuccess(t *testing.T) {
	baseSrc := `
	struct User {
		id: int64;
		name: string;
	}
	`
	currentSrc := `
	struct User {
		id: int64;
		name: string;
		email: string;
	}
	`

	baseAST, err := parser.NewParser(baseSrc).Parse()
	if err != nil {
		t.Fatal(err)
	}

	currentAST, err := parser.NewParser(currentSrc).Parse()
	if err != nil {
		t.Fatal(err)
	}

	err = ValidateSchemaEvolution(&baseAST, &currentAST)
	if err != nil {
		t.Fatalf("expected compatibility validation to pass, got: %v", err)
	}
}

func TestValidationFailDeletedField(t *testing.T) {
	baseSrc := `
	struct User {
		id: int64;
		name: string;
	}
	`
	currentSrc := `
	struct User {
		id: int64;
	}
	`

	baseAST, err := parser.NewParser(baseSrc).Parse()
	if err != nil {
		t.Fatal(err)
	}

	currentAST, err := parser.NewParser(currentSrc).Parse()
	if err != nil {
		t.Fatal(err)
	}

	err = ValidateSchemaEvolution(&baseAST, &currentAST)
	if err == nil {
		t.Fatal("expected compatibility check to fail due to deleted field")
	}
}

func TestValidationFailTypeChanged(t *testing.T) {
	baseSrc := `
	struct User {
		id: int64;
		name: string;
	}
	`
	currentSrc := `
	struct User {
		id: int32;
		name: string;
	}
	`

	baseAST, err := parser.NewParser(baseSrc).Parse()
	if err != nil {
		t.Fatal(err)
	}

	currentAST, err := parser.NewParser(currentSrc).Parse()
	if err != nil {
		t.Fatal(err)
	}

	err = ValidateSchemaEvolution(&baseAST, &currentAST)
	if err == nil {
		t.Fatal("expected compatibility check to fail due to type change")
	}
}

func TestValidationFailInsertionOrReorder(t *testing.T) {
	baseSrc := `
	struct User {
		id: int64;
		name: string;
	}
	`
	currentSrc := `
	struct User {
		id: int64;
		email: string;
		name: string;
	}
	`

	baseAST, err := parser.NewParser(baseSrc).Parse()
	if err != nil {
		t.Fatal(err)
	}

	currentAST, err := parser.NewParser(currentSrc).Parse()
	if err != nil {
		t.Fatal(err)
	}

	err = ValidateSchemaEvolution(&baseAST, &currentAST)
	if err == nil {
		t.Fatal("expected compatibility check to fail due to field insertion/reorder")
	}
}
