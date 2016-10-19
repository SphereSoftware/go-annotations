package registry

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

// Parses provided source file and extract annotations for all objects
// (structures, interfaces, methods, functions) as annotated entries.
// Also it returns all found imports and full package name of the parsed file
func ParseFile(path, file string) ([]AnnotatedEntry, []string, string) {
	var foundImports []string
	var foundAnnotations []AnnotatedEntry
	var foundPackage string
	source := filepath.Join(path, file)
	fset := token.NewFileSet()
	fileNode, err := parser.ParseFile(fset, source, nil, parser.ParseComments)
	if err != nil {
		panic("Error while parse source file " + source + ":\n" + err.Error())
	}
	foundPackage = fileNode.Name.Name
	fullPackage := resolveFullPackage(path, foundPackage)
	for _, decl := range fileNode.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if fd.Recv == nil {
				processFunc(fd, &foundAnnotations, fullPackage)
			} else {
				processMethod(fd, &foundAnnotations, fullPackage)
			}
		} else {
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					is, ok := spec.(*ast.ImportSpec)
					if !ok {
						continue
					} else {
						processImports(is, &foundImports)
					}
				} else {
					str, ok := ts.Type.(*ast.StructType)
					if !ok {
						intf, ok := ts.Type.(*ast.InterfaceType)
						if !ok {
							continue
						} else {
							processInterface(ts, intf, &foundAnnotations, fullPackage)
						}
					} else if str.Incomplete {
						continue
					} else {
						processStruct(ts, str, &foundAnnotations, fullPackage)
					}
				}
			}
		}
	}
	return foundAnnotations, foundImports, fullPackage
}

func processImports(is *ast.ImportSpec, foundImports *[]string) {
	v := is.Path.Value
	if strings.HasPrefix(v, "\"") && strings.HasSuffix(v, "\"") {
		v = v[1 : len(v)-1]
	}
	*foundImports = append(*foundImports, v)
}

func processFunc(fd *ast.FuncDecl, foundAnnotations *[]AnnotatedEntry, fullPackage string) {
	name := fd.Name.Name
	doc := fd.Doc.Text()
	a := FindAnnotations(doc)
	if len(a) > 0 {
		*foundAnnotations = append(*foundAnnotations,
			AnnotatedEntry{"func", fullPackage, name, AnnotationsData{a, nil, nil}})
	}
}

func processMethod(fd *ast.FuncDecl, foundAnnotations *[]AnnotatedEntry, fullPackage string) {
	name := fd.Name.Name
	if len(fd.Recv.List) == 1 {
		tp := getReceiverType(fd.Recv.List[0].Type)
		doc := fd.Doc.Text()
		a := FindAnnotations(doc)
		if len(a) > 0 {
			fieldsMap := map[string][]AnnotationDoc{name: a}
			*foundAnnotations = append(*foundAnnotations,
				AnnotatedEntry{"struct", fullPackage, tp, AnnotationsData{nil, fieldsMap, nil}})
		}
	}
}

// Returns method receiver's type name as a string
// if receiver is a pointer than star is not added to the name
func getReceiverType(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.StarExpr:
		return getReceiverType(t.X)
	case *ast.Ident:
		return t.Name
	}
	panic("Unsupported receiver type")
}

func processStruct(ts *ast.TypeSpec, str *ast.StructType, foundAnnotations *[]AnnotatedEntry, fullPackage string) {
	name := ts.Name.Name
	doc := ts.Doc.Text()
	selfAnnotations := FindAnnotations(doc)
	fieldsAnnotations := make(map[string][]AnnotationDoc)
	for _, field := range str.Fields.List {
		doc := field.Doc.Text()
		fieldAnnotations := FindAnnotations(doc)
		if len(fieldAnnotations) > 0 {
			fieldName := getFieldName(field)
			fieldsAnnotations[fieldName] = fieldAnnotations
		}
	}
	if len(selfAnnotations) > 0 || len(fieldsAnnotations) > 0 {
		*foundAnnotations = append(*foundAnnotations,
			AnnotatedEntry{"struct", fullPackage, name,
				AnnotationsData{selfAnnotations, fieldsAnnotations, nil}})
	}
}

func processInterface(ts *ast.TypeSpec, intf *ast.InterfaceType, foundAnnotations *[]AnnotatedEntry, fullPackage string) {
	name := ts.Name.Name
	doc := ts.Doc.Text()
	selfAnnotations := FindAnnotations(doc)
	methodsAnnotations := make(map[string][]AnnotationDoc)
	for _, method := range intf.Methods.List {
		doc := method.Doc.Text()
		methodAnnotations := FindAnnotations(doc)
		if len(methodAnnotations) > 0 {
			methodName := method.Names[0].Name
			methodsAnnotations[methodName] = methodAnnotations
		}
	}
	if len(selfAnnotations) > 0 || len(methodsAnnotations) > 0 {
		*foundAnnotations = append(*foundAnnotations,
			AnnotatedEntry{"interface", fullPackage, name,
				AnnotationsData{selfAnnotations, nil, methodsAnnotations}})
	}
}
