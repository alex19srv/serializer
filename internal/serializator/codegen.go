// SPDX-License-Identifier: MIT
// Copyright 2026 Alex Syrnikov <alex19srv@gmail.com>

package serializator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

const (
	typeInt = iota
	typeUint
	typeString
)

type PackageData struct {
	Name string
}

type StructData struct {
	Name   string  // struct name
	Fields []Field // struct fields
}
type Field struct {
	Name string
	Type uint32
}

var structsData []StructData
var packageData PackageData

func GenerateSerializationsForFile(inputPath, outputPath string) error {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, inputPath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse %s: %w", inputPath, err)
	}
	// ast.Fprint(os.Stdout, fileSet, file, nil) // debugging method. Do not remove!!! just comment it in production

	packageData.Name = file.Name.Name

	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.TYPE {
			continue
		}
		for _, specification := range general.Specs {
			typeSpec, ok := specification.(*ast.TypeSpec)
			if !ok {
				continue
			}
			astStructType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			if !hasSerializeMarker(general.Doc) {
				continue
			}
			structName := typeSpec.Name.Name
			fields, err := readStructFields(astStructType)
			if err != nil {
				return fmt.Errorf("generate %s: %w", typeSpec.Name.Name, err)
			}
			structsData = append(structsData, StructData{Name: structName, Fields: fields})
		}
	}
	// fmt.Printf("Structs %v\n", structsData)

	if len(structsData) == 0 {
		return fmt.Errorf("no structs marked with // saga:serialize found in %s", inputPath)
	}

	code, err := generateSrcFileText()
	if err != nil {
		return fmt.Errorf("generate code: %w", err)
	}
	// fmt.Println(string(code))

	if err := os.WriteFile(outputPath, code, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", outputPath, err)
	}

	return nil
}

func hasSerializeMarker(group *ast.CommentGroup) bool {
	if group == nil {
		return false
	}
	for _, comment := range group.List {
		text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
		if text == "saga:serialize" || strings.HasPrefix(text, "saga:serialize ") {
			return true
		}
	}

	return false
}

func readStructFields(astType *ast.StructType) ([]Field, error) {
	var fields []Field

	if len(astType.Fields.List) == 0 {
		return fields, fmt.Errorf("empty structs are not supported")
	}

	for _, field := range astType.Fields.List {
		fieldType, ok := field.Type.(*ast.Ident)
		if !ok {
			return fields, fmt.Errorf("unsupported field type %s", expressionString(field.Type))
		}
		fieldTypeName := fieldType.Name
		var fieldTypeID uint32
		switch fieldTypeName {
		case "int":
			fieldTypeID = typeInt
		case "uint":
			fieldTypeID = typeUint
		case "string":
			fieldTypeID = typeString
		default:
			return fields, fmt.Errorf("unsupported type %s", fieldTypeName)
		}
		if len(field.Names) == 0 {
			return fields, fmt.Errorf("embedded fields are not supported")
		}
		for _, name := range field.Names {
			fields = append(fields, Field{Name: name.Name, Type: fieldTypeID})
		}
	}

	return fields, nil
}

func expressionString(expression ast.Expr) string {
	var output bytes.Buffer
	_ = format.Node(&output, token.NewFileSet(), expression)
	return output.String()
}

func generateSrcFileText() ([]byte, error) {
	var output bytes.Buffer
	headerTmpl.Execute(&output, &packageData)

	// generate methods
	for _, data := range structsData {
		methodText, err := generateMethod(data)
		if err != nil {
			return nil, fmt.Errorf("generate method for %s: %w", data.Name, err)
		}
		output.WriteString(methodText)
	}

	formatted, err := format.Source(output.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated source: %w", err)
	}

	return formatted, nil
}

func generateMethod(data StructData) (string, error) {
	var serializeOutBuffer bytes.Buffer
	var deserializeOutBuffer bytes.Buffer
	err := serializeHeaderTmpl.Execute(&serializeOutBuffer, &data)
	if err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	err = deserializeHeaderTmpl.Execute(&deserializeOutBuffer, &data)
	if err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	for _, field := range data.Fields {
		switch field.Type {
		case typeInt:
			err = serializeIntTmpl.Execute(&serializeOutBuffer, &field)
			if err != nil {
				break
			}
			err = deserializeIntTmpl.Execute(&deserializeOutBuffer, &field)
		// case typeUint:
		// 	err = serializeUintTmpl.Execute(&output, &field)
		case typeString:
			err = serializeStringTmpl.Execute(&serializeOutBuffer, &field)
			if err != nil {
				break
			}
			err = deserializeStringTmpl.Execute(&deserializeOutBuffer, &field)
		default:
			return "", fmt.Errorf("unsupported type %d", field.Type)
		}
		if err != nil {
			return "", fmt.Errorf("execute template: %w", err)
		}
	}
	err = serializeFooterTmpl.Execute(&serializeOutBuffer, &data)
	if err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	err = deserializeFooterTmpl.Execute(&deserializeOutBuffer, &data)
	if err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	serializeOutBuffer.Write(deserializeOutBuffer.Bytes())

	return serializeOutBuffer.String(), nil
}
