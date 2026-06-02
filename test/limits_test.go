package test

import (
	"strings"
	"testing"

	"github.com/snowmerak/antiserial/compiler/codegen"
	"github.com/snowmerak/antiserial/compiler/guardian"
	"github.com/snowmerak/antiserial/compiler/parser"
	"github.com/snowmerak/antiserial/test/testgen_v2"
)

func TestMarshalStringExceedsUint16Max(t *testing.T) {
	p := testgen_v2.Payload{
		Uuid: strings.Repeat("a", codegen.MaxUint16+1),
	}
	_, err := p.Marshal(nil)
	if err == nil {
		t.Fatal("expected Marshal error for string length > 65535")
	}
	if !strings.Contains(err.Error(), "uint16") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMarshalListExceedsUint16Max(t *testing.T) {
	tags := make([]string, codegen.MaxUint16+1)
	for i := range tags {
		tags[i] = "x"
	}
	p := testgen_v2.Payload{Tags: tags}
	_, err := p.Marshal(nil)
	if err == nil {
		t.Fatal("expected Marshal error for list length > 65535")
	}
	if !strings.Contains(err.Error(), "uint16") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMarshalStringAtUint16Max(t *testing.T) {
	p := testgen_v2.Payload{
		Uuid: strings.Repeat("a", codegen.MaxUint16),
	}
	serialized, err := p.Marshal(nil)
	if err != nil {
		t.Fatalf("Marshal failed at max length: %v", err)
	}

	var decoded testgen_v2.Payload
	if _, err := decoded.Unmarshal(serialized); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if len(decoded.Uuid) != codegen.MaxUint16 {
		t.Fatalf("round-trip length: got %d want %d", len(decoded.Uuid), codegen.MaxUint16)
	}
}

func TestGuardianPrimitiveAliasInProcess(t *testing.T) {
	baseSrc := `
	struct User {
		score: float;
		ratio: double;
	}
	`
	currentSrc := `
	struct User {
		score: float32;
		ratio: float64;
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
	if err := guardian.ValidateSchemaEvolution(&baseAST, &currentAST); err != nil {
		t.Fatalf("expected alias-compatible evolution, got: %v", err)
	}
}