package parser

import (
	"testing"
)

func TestParseSuccess(t *testing.T) {
	src := `
	// A basic geometric coordinate struct
	struct Geo {
		lat: float64;
		lng: float64;
	}

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
	`
	p := NewParser(src)
	astVal, err := p.Parse()
	if err != nil {
		t.Fatalf("expected no parse error, got: %v", err)
	}

	if len(astVal.Structs) != 2 {
		t.Errorf("expected 2 structs, got %d", len(astVal.Structs))
	}

	geo, ok := astVal.FindStruct("Geo")
	if !ok {
		t.Fatal("Geo struct not found")
	}
	if len(geo.Fields) != 2 {
		t.Errorf("expected Geo to have 2 fields, got %d", len(geo.Fields))
	}

	payload, ok := astVal.FindStruct("ComplexPayload")
	if !ok {
		t.Fatal("ComplexPayload struct not found")
	}
	if len(payload.Fields) != 10 {
		t.Errorf("expected ComplexPayload to have 10 fields, got %d", len(payload.Fields))
	}
}

func TestParseUndefinedType(t *testing.T) {
	src := `
	struct Point {
		x: int32;
		y: InvalidType;
	}
	`
	p := NewParser(src)
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected undefined type error, got nil")
	}
}

func TestParseOptionalField(t *testing.T) {
	src := `
	struct Payload {
		id: int64;
		score: optional int32;
	}
	`
	astVal, err := NewParser(src).Parse()
	if err != nil {
		t.Fatalf("parse optional: %v", err)
	}
	payload, ok := astVal.FindStruct("Payload")
	if !ok || len(payload.Fields) != 2 {
		t.Fatalf("unexpected struct: %+v", payload)
	}
	if !payload.Fields[1].Type.Optional || payload.Fields[1].Type.Name != "int32" {
		t.Fatalf("score field type: %+v", payload.Fields[1].Type)
	}
}

func TestParseDuplicateField(t *testing.T) {
	src := `
	struct Point {
		x: int32;
		x: float64;
	}
	`
	p := NewParser(src)
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected duplicate field error, got nil")
	}
}
