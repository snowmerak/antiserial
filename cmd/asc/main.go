package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"


	"github.com/snowmerak/antiserial/compiler/codegen/cpp"
	"github.com/snowmerak/antiserial/compiler/codegen/golang"
	"github.com/snowmerak/antiserial/compiler/codegen/python"
	"github.com/snowmerak/antiserial/compiler/codegen/rust"
	"github.com/snowmerak/antiserial/compiler/codegen/typescript"
	"github.com/snowmerak/antiserial/compiler/guardian"
	"github.com/snowmerak/antiserial/compiler/parser"
)

func main() {
	// CLI Flags
	goOut := flag.String("go_out", "", "Directory path for generated Go source code")
	rustOut := flag.String("rust_out", "", "Directory path for generated Rust source code")
	cppOut := flag.String("cpp_out", "", "Directory path for generated C++ header source code")
	tsOut := flag.String("ts_out", "", "Directory path for generated TypeScript source code")
	pyOut := flag.String("py_out", "", "Directory path for generated Python source code")
	baseSchema := flag.String("base_schema", "", "Path to the base schema file for backward compatibility validation")
	validateOnly := flag.Bool("validate_only", false, "Perform validation only, do not generate source code")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: asc [options] <schema_file.as>\n\nOptions:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Error: missing schema file argument\n\n")
		flag.Usage()
		os.Exit(1)
	}

	schemaPath := args[0]
	schemaSrc, err := os.ReadFile(schemaPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading schema file %s: %v\n", schemaPath, err)
		os.Exit(1)
	}

	// 1. Parse current schema
	p := parser.NewParser(string(schemaSrc))
	currentAST, err := p.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Syntax/Semantic Error parsing %s:\n%v\n", schemaPath, err)
		os.Exit(1)
	}

	// 2. Perform Schema Guardian backward compatibility check if base schema is provided
	if *baseSchema != "" {
		baseSrc, err := os.ReadFile(*baseSchema)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading base schema file %s: %v\n", *baseSchema, err)
			os.Exit(1)
		}

		baseP := parser.NewParser(string(baseSrc))
		baseAST, err := baseP.Parse()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Syntax/Semantic Error parsing base schema %s:\n%v\n", *baseSchema, err)
			os.Exit(1)
		}

		err = guardian.ValidateSchemaEvolution(&baseAST, &currentAST)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Schema Evolution Check Failed:\n%v\n", err)
			os.Exit(1)
		}
		fmt.Println("Backward compatibility check: PASSED")
	}

	if *validateOnly {
		fmt.Println("Schema validation: SUCCESSFUL")
		os.Exit(0)
	}

	// Determine output base file name
	baseName := filepath.Base(schemaPath)
	ext := filepath.Ext(baseName)
	rawName := strings.TrimSuffix(baseName, ext)

	// 3. Code generation
	if *goOut != "" {
		// Create output dir if it does not exist
		if err := os.MkdirAll(*goOut, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating Go output directory: %v\n", err)
			os.Exit(1)
		}

		// Extract package name from goOut path or default to "antiserial"
		pkgName := filepath.Base(*goOut)
		if pkgName == "." || pkgName == "/" || pkgName == "\\" {
			pkgName = "antiserial"
		}

		generatedGo, err := golang.Generate(pkgName, currentAST)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Go Code Generation failed: %v\n", err)
			os.Exit(1)
		}

		outFilePath := filepath.Join(*goOut, rawName+".go")
		if err := os.WriteFile(outFilePath, []byte(generatedGo), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing Go output file %s: %v\n", outFilePath, err)
			os.Exit(1)
		}
		fmt.Printf("Go code generated successfully: %s\n", outFilePath)
	}

	if *rustOut != "" {
		if err := os.MkdirAll(*rustOut, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating Rust output directory: %v\n", err)
			os.Exit(1)
		}

		generatedRust, err := rust.Generate(currentAST)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Rust Code Generation failed: %v\n", err)
			os.Exit(1)
		}

		outFilePath := filepath.Join(*rustOut, rawName+".rs")
		if err := os.WriteFile(outFilePath, []byte(generatedRust), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing Rust output file %s: %v\n", outFilePath, err)
			os.Exit(1)
		}
		fmt.Printf("Rust code generated successfully: %s\n", outFilePath)
	}

	if *cppOut != "" {
		if err := os.MkdirAll(*cppOut, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating C++ output directory: %v\n", err)
			os.Exit(1)
		}

		generatedCpp, err := cpp.Generate(currentAST)
		if err != nil {
			fmt.Fprintf(os.Stderr, "C++ Code Generation failed: %v\n", err)
			os.Exit(1)
		}

		outFilePath := filepath.Join(*cppOut, rawName+".hpp")
		if err := os.WriteFile(outFilePath, []byte(generatedCpp), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing C++ output file %s: %v\n", outFilePath, err)
			os.Exit(1)
		}
		fmt.Printf("C++ code generated successfully: %s\n", outFilePath)
	}

	if *tsOut != "" {
		if err := os.MkdirAll(*tsOut, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating TypeScript output directory: %v\n", err)
			os.Exit(1)
		}

		generatedTs, err := typescript.Generate(currentAST)
		if err != nil {
			fmt.Fprintf(os.Stderr, "TypeScript Code Generation failed: %v\n", err)
			os.Exit(1)
		}

		outFilePath := filepath.Join(*tsOut, rawName+".ts")
		if err := os.WriteFile(outFilePath, []byte(generatedTs), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing TypeScript output file %s: %v\n", outFilePath, err)
			os.Exit(1)
		}
		fmt.Printf("TypeScript code generated successfully: %s\n", outFilePath)
	}

	if *pyOut != "" {
		if err := os.MkdirAll(*pyOut, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating Python output directory: %v\n", err)
			os.Exit(1)
		}

		generatedPy, err := python.Generate(currentAST)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Python Code Generation failed: %v\n", err)
			os.Exit(1)
		}

		outFilePath := filepath.Join(*pyOut, rawName+".py")
		if err := os.WriteFile(outFilePath, []byte(generatedPy), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing Python output file %s: %v\n", outFilePath, err)
			os.Exit(1)
		}
		fmt.Printf("Python code generated successfully: %s\n", outFilePath)
	}
}
