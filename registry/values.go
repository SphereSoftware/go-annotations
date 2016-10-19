package registry

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

// Returns annotated entry annotations descriptor and list of all annotations packages in it
// Parameters:
// - annotated entry
// - full package name
// - list of imports in the entry source
func GenerateAnnotationValue(a *AnnotatedEntry, packageName string, foundImports []string) (string, []string) {
	var b bytes.Buffer
	var allPackages []string
	log.Printf("Generating annotations values in package %s for %s %s\n", packageName, a.Type, a.Name)
	// write self
	if len(a.AnnotationsData.Self) > 0 {
		log.Printf("Self : %d\n", len(a.AnnotationsData.Self))
	}
	b.WriteString("        _base.Annotations {\n            Self: []interface{} {\n")
	for _, self := range a.AnnotationsData.Self {
		s, packages := generateStruct(&self, packageName, foundImports, "                ")
		allPackages = combinePackages(allPackages, packages)
		b.WriteString(s)
		b.WriteString(",\n")
	}
	b.WriteString("            },\n            Fields: map[string][]interface{} {\n")
	for field, fieldAnnotations := range a.AnnotationsData.Fields {
		if len(fieldAnnotations) > 0 {
			log.Printf("Field .%s: %d\n", field, len(fieldAnnotations))
		}
		b.WriteString("                " + strconv.Quote(field) + ": []interface{} {\n")
		for _, an := range fieldAnnotations {
			s, packages := generateStruct(&an, packageName, foundImports, "                    ")
			allPackages = combinePackages(allPackages, packages)
			b.WriteString(s)
			b.WriteString(",\n")
		}
		b.WriteString("                },\n")
	}
	b.WriteString("            },\n            Methods: map[string][]interface{} {\n")
	for method, methodAnnotations := range a.AnnotationsData.Methods {
		if len(methodAnnotations) > 0 {
			log.Printf("Method %s(): %d\n", method, len(methodAnnotations))
		}
		b.WriteString("                " + strconv.Quote(method) + ": []interface{} {\n")
		for _, an := range methodAnnotations {
			s, packages := generateStruct(&an, packageName, foundImports, "                    ")
			allPackages = combinePackages(allPackages, packages)
			b.WriteString(s)
			b.WriteString(",\n")
		}
		b.WriteString("},\n")
	}
	b.WriteString("}}")
	return b.String(), allPackages
}

// Generates structure initialization for provided annotation.
// Returns the generated code and list of packages used for
// all enclosed annotations types.
// Annotations types inside the code will be prefixed by full
// package name.
// Parameters:
// - annotation instance
// - full package name where given instance is found
// - list of imports found in the file containing the annotated entry
// - string of spaces for idents
func generateStruct(a *AnnotationDoc, packageName string, imports []string, indent string) (string, []string) {
	var allAnnotationsPackages []string
	possiblePackagesForA := combinePackages(imports, []string{packageName})
	ts, foundPackageOfA, foundImportsOfA := getAnnotationStruct(a.Name, possiblePackagesForA)
	allAnnotationsPackages = combinePackages(allAnnotationsPackages, []string{foundPackageOfA})
	str, _ := ts.Type.(*ast.StructType)
	var b bytes.Buffer
	b.WriteString(indent)
	b.WriteString(foundPackageOfA)
	b.WriteString(".")
	b.WriteString(a.Name)
	b.WriteString("{\n")
	childIndent := indent + "    "
	for _, f := range str.Fields.List {
		fieldName := getFieldName(f)
		defValue := getDefaultValue(f)
		fieldKey := fieldName
		// consider special case when only default parameter is specified
		if len(str.Fields.List) == 1 && len(a.Content) == 1 {
			for key := range a.Content {
				if key == DEFAULT_PARAM {
					fieldKey = DEFAULT_PARAM
				}
			}
		}
		value, found := a.Content[fieldKey]
		if found {
			switch t := value.(type) {
			case string:
				b.WriteString(childIndent)
				b.WriteString(getLiteral(f.Type, t, false))
				b.WriteString(",\n")
			case []string:
				b.WriteString(childIndent)
				b.WriteString(getFieldConstructor(f.Type))
				b.WriteString("\n")
				for _, elem := range t {
					b.WriteString(childIndent + "    ")
					b.WriteString(elem)
					b.WriteString(",\n")
				}
				b.WriteString(childIndent)
				b.WriteString("}")
			case []AnnotationDoc:
				// calculate array's elements
				var bb bytes.Buffer
				for _, sa := range t {
					childCode, foundImportsOfChild := generateStruct(&sa, foundPackageOfA, foundImportsOfA, childIndent+"    ")
					allAnnotationsPackages = combinePackages(allAnnotationsPackages, foundImportsOfChild)
					bb.WriteString(childCode)
					bb.WriteString(",\n")
				}
				b.WriteString(childIndent)
				// insert array initialzer of child annotation type
				s := writeArrayInitializer(&b, bb.String())
				// append array of child annotations
				b.WriteString("{\n")
				b.WriteString(childIndent + "    ")
				b.WriteString(s)
				b.WriteString(childIndent)
				b.WriteString("},\n")
			case AnnotationDoc:
				childCode, foundImportsOfChild := generateStruct(&t, foundPackageOfA, foundImportsOfA, childIndent)
				allAnnotationsPackages = combinePackages(allAnnotationsPackages, foundImportsOfChild)
				b.WriteString(childIndent)
				if isOptional(f.Type) {
					b.WriteString("&")
				}
				b.WriteString(strings.TrimLeft(childCode, " "))
				b.WriteString(",\n")
			default:
				panic("Unexpected annotation value type")
			}
		} else {
			b.WriteString(childIndent)
			b.WriteString(defValue)
			b.WriteString(",\n")
		}
	}
	b.WriteString(indent)
	b.WriteString("}")
	return b.String(), allAnnotationsPackages
}

// Writes array initializer for type X as "[]X {" into provided text buffer
// Parameters:
// - address of text buffer to write
// - structural constructor text starting from spaces
func writeArrayInitializer(b *bytes.Buffer, s string) string {
	start := 0
	n := len(s)
	for start < n && s[start] == ' ' {
		start++
	}
	if start < n {
		s = s[start:]
		pos := strings.Index(s, "{")
		b.WriteString("[]")
		b.WriteString(s[:pos])
		return s
	} else {
		panic("Empty annotation is returned")
	}
}

// Returns true if given type is a pointer
func isOptional(e ast.Expr) bool {
	switch e.(type) {
	case *ast.StarExpr:
		return true
	}
	return false
}

// Returns TypeSpec for the annotation struct, its package and list of imports
// from the file where that struct is defined.
func getAnnotationStruct(name string, possiblePackages []string) (*ast.TypeSpec, string, []string) {
	var result *ast.TypeSpec
	var foundPackage string
	var foundImports []string
	var foundDir string
	for _, pck := range possiblePackages {
		for _, dir := range findDirs(pck) {
			// get all go sources from package folder
			files, err := ioutil.ReadDir(dir)
			if err != nil {
				panic(err)
			}
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".go") {
					source := filepath.Join(dir, file.Name())
					// parse source file
					fset := token.NewFileSet()
					fileNode, err := parser.ParseFile(fset, source, nil, parser.ParseComments)
					if err != nil {
						panic("Error while parse source file " + source + ":\n" + err.Error())
					}
					for _, decl := range fileNode.Decls {
						gd, ok := decl.(*ast.GenDecl)
						if !ok {
							continue
						}
						for _, spec := range gd.Specs {
							ts, ok := spec.(*ast.TypeSpec)
							if !ok {
								is, ok := spec.(*ast.ImportSpec)
								if !ok {
									continue
								}
								v := is.Path.Value
								if strings.HasPrefix(v, "\"") && strings.HasSuffix(v, "\"") {
									v = v[1 : len(v)-1]
								}
								foundImports = combinePackages(foundImports, []string{v})
							} else {
								str, ok := ts.Type.(*ast.StructType)
								if !ok || str.Incomplete {
									continue
								}
								if ts.Name.Name == name {
									if result == nil {
										result = ts
										foundPackage = pck
										foundDir = dir
										break
									} else {
										panicReason(name, dir, foundDir, pck, foundPackage)
									}
								}
							}
						}
					}
				}
			}
		}
	}
	if result != nil {
		return result, foundPackage, foundImports
	}
	panic("Annotation source for '" + name + "' is not found")
}

// Makes the panic call with the reason corresponding to situation
func panicReason(name, dir, foundDir, pck, foundPackage string) {
	if dir == foundDir {
		panic("Ambiguous reference to annotation '" + name +
			"'\nIt exists in packages '" + foundPackage +
			"' and '" + pck + "'")
	} else if pck == foundPackage {
		panic("Ambiguous reference to annotation '" + name +
			":\nthe same package '" + pck + "' exists in folders:\n'" + dir +
			"'\n and in folder \n'" + foundDir + "'")
	} else {
		panic("Ambiguous reference to annotation '" + name + "':\n" +
			"- " + dir + " / " + pck + "\n - " + foundDir + " / " + foundPackage)
	}
}

// Extracts default value for the field from its tag
// default value is stored in form `deafult:"XXX"`
func getDefaultValue(f *ast.Field) string {
	if f.Tag != nil {
		tag := f.Tag.Value
		n := len(tag) - 1
		if n > 1 {
			tag := reflect.StructTag(tag[1:n]).Get("default")
			if len(tag) > 0 {
				return getLiteral(f.Type, tag, false)
			}
		}
	}
	return getZeroLiteral(f.Type)
}

// Returns the literal value representation beased on its type
func getLiteral(e ast.Expr, value string, wasPointer bool) string {
	switch t := e.(type) {
	case *ast.StarExpr:
		if wasPointer {
			panic("Poniter to pointers should not be used as annotation field")
		}
		return getLiteral(t.X, value, true)
	case *ast.Ident:
		switch t.Name {
		case "string":
			return strconv.Quote(value)
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64",
			"float32", "float64", "byte", "rune":
			return value
		default:
			panic("Type '" + t.Name + "' doesn't support default value for annotation field")
		}
	default:
		panic("Unsupported field type in annotation definition")
	}
}

// Returns literal representatin of zero value for provided type
func getZeroLiteral(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.StarExpr:
		return "nil"
	case *ast.Ident:
		switch t.Name {
		case "string":
			return "\"\""
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64",
			"float32", "float64", "byte", "rune":
			return "0"
		default:
			panic("Type '" + t.Name + "' doesn't support default value for annotation field")
		}
	case *ast.ArrayType:
		return "nil"
	default:
		panic("Unsupported fied type in annotation definition")
	}
}

// Returns the constructor expression which can create the value of given type
func getFieldConstructor(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.StarExpr:
		switch t.X.(type) {
		case *ast.StarExpr:
			panic("Ponter on pointers is not supported in annotation struct")
		case *ast.ArrayType:
			panic("Pointer on arrays is not supported in annotation struct")
		default:
			return "&" + getFieldConstructor(t.X)
		}
	case *ast.ArrayType:
		switch elemType := t.Elt.(type) {
		case *ast.StarExpr:
			panic("Array of pointers is not supported in annotation struct")
		case *ast.ArrayType:
			panic("Array of arrays is not supported in annotation struct")
		default:
			return "[]" + getFieldConstructor(elemType)
		}
	case *ast.Ident:
		switch t.Name {
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64",
			"float32", "float64", "byte", "rune", "string":
			return t.Name + "{"
		case "complex64", "complex128", "uintptr":
			panic("Type '" + t.Name + "' is not supported in annotation struct")
		default:
			return t.Name + "{"
		}
	default:
		panic("Unsupported field type in annotation")
	}
}
