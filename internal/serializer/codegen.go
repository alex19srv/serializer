// SPDX-License-Identifier: MIT
// Copyright 2026 Alex Syrnikov <alex19srv@gmail.com>

package serializer

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
	typeInt8
	typeUint8
	typeInt16
	typeUint16
	typeInt32
	typeUint32
	typeInt64
	typeUint64
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

	if len(structsData) == 0 {
		return fmt.Errorf("no structs marked with // saga:serialize found in %s", inputPath)
	}

	code, err := generateSrcFileText()
	if err != nil {
		return fmt.Errorf("generate code: %w", err)
	}

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
		case "int8":
			fieldTypeID = typeInt8
		case "uint8":
			fieldTypeID = typeUint8
		case "int16":
			fieldTypeID = typeInt16
		case "uint16":
			fieldTypeID = typeUint16
		case "int32":
			fieldTypeID = typeInt32
		case "uint32":
			fieldTypeID = typeUint32
		case "int64":
			fieldTypeID = typeInt64
		case "uint64":
			fieldTypeID = typeUint64
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
		methodText, err := generateStructMethods(data)
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

func generateStructMethods(data StructData) (string, error) {
	var serializeOutBuffer bytes.Buffer
	var serializeSizeOutBuffer bytes.Buffer
	var deserializeOutBuffer bytes.Buffer
	err := serializeHeaderTmpl.Execute(&serializeOutBuffer,
		&struct {
			StructName string
		}{
			data.Name,
		})
	if err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	err = serializedSizeHeaderTmpl.Execute(&serializeSizeOutBuffer,
		&struct {
			StructName string
		}{
			data.Name,
		})
	if err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	err = deserializeHeaderTmpl.Execute(&deserializeOutBuffer,
		&struct {
			StructName string
		}{
			data.Name,
		})
	if err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	for _, field := range data.Fields {
		switch field.Type {
		case typeInt:
			err = serializeIntField("int", &serializeOutBuffer, &serializeSizeOutBuffer, &deserializeOutBuffer, field)
		case typeUint:
			err = serializeIntField("uint", &serializeOutBuffer, &serializeSizeOutBuffer, &deserializeOutBuffer, field)
		case typeInt8:
			err = serializeIntField("int8", &serializeOutBuffer, &serializeSizeOutBuffer, &deserializeOutBuffer, field)
		case typeUint8:
			err = serializeIntField("uint8", &serializeOutBuffer, &serializeSizeOutBuffer, &deserializeOutBuffer, field)
		case typeInt16:
			err = serializeIntField("int16", &serializeOutBuffer, &serializeSizeOutBuffer, &deserializeOutBuffer, field)
		case typeUint16:
			err = serializeIntField("uint16", &serializeOutBuffer, &serializeSizeOutBuffer, &deserializeOutBuffer, field)
		case typeInt32:
			err = serializeIntField("int32", &serializeOutBuffer, &serializeSizeOutBuffer, &deserializeOutBuffer, field)
		case typeUint32:
			err = serializeIntField("uint32", &serializeOutBuffer, &serializeSizeOutBuffer, &deserializeOutBuffer, field)
		case typeInt64:
			err = serializeIntField("int64", &serializeOutBuffer, &serializeSizeOutBuffer, &deserializeOutBuffer, field)
		case typeUint64:
			err = serializeIntField("uint64", &serializeOutBuffer, &serializeSizeOutBuffer, &deserializeOutBuffer, field)
		case typeString:
			err = serializeStringTmpl.Execute(&serializeOutBuffer,
				&struct {
					FieldName string
					CopyData  bool
				}{
					field.Name,
					true,
				})
			if err != nil {
				break
			}
			err = serializeStringTmpl.Execute(&serializeSizeOutBuffer,
				&struct {
					FieldName string
					CopyData  bool
				}{
					field.Name,
					false,
				})
			if err != nil {
				break
			}
			err = deserializeStringTmpl.Execute(&deserializeOutBuffer,
				&struct {
					FieldName string
				}{
					field.Name,
				})
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
	err = serializeFooterTmpl.Execute(&serializeSizeOutBuffer, &data)
	if err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	err = deserializeFooterTmpl.Execute(&deserializeOutBuffer, &data)
	if err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	serializeOutBuffer.Write(serializeSizeOutBuffer.Bytes())
	serializeOutBuffer.Write(deserializeOutBuffer.Bytes())

	return serializeOutBuffer.String(), nil
}

func serializeIntField(
	intTypeName string,
	serializeOutBuffer *bytes.Buffer,
	serializeSizeOutBuffer *bytes.Buffer,
	deserializeOutBuffer *bytes.Buffer,
	field Field,
) error {
	err := serializeIntTmpl.Execute(serializeOutBuffer,
		&struct {
			FieldName string
			CopyData  bool
			TypeName  string
		}{
			field.Name,
			true,
			intTypeName,
		})
	if err != nil {
		return err
	}
	err = serializeIntTmpl.Execute(serializeSizeOutBuffer,
		&struct {
			FieldName string
			CopyData  bool
			TypeName  string
		}{
			field.Name,
			false,
			intTypeName,
		})
	if err != nil {
		return err
	}
	err = deserializeIntTmpl.Execute(deserializeOutBuffer,
		&struct {
			FieldName string
			TypeName  string
		}{
			field.Name,
			intTypeName,
		})
	return err
}
