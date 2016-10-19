package registry

import (
	"reflect"
)

type (
	// Annotation properties extracted from the comments
	AnnotationDoc struct {
		Name    string
		Content map[string]interface{}
	}

	// Bundle of annotations related to object in source code
	AnnotationsData struct {
		Self    []AnnotationDoc            // annotations related to struct/func/interface name
		Fields  map[string][]AnnotationDoc // annotations related to struct field
		Methods map[string][]AnnotationDoc // annotations related to struct methods
	}

	// Full description of annotated entry
	AnnotatedEntry struct {
		Type            string // type name of annotated entry of func/method name
		FullPackage     string // full package name of annotated entry
		Name            string // the name of annotated struct/func/interface/method
		AnnotationsData        // related annotation data
	}

	// The annotations bundle stored in Registry for each entry.
	// It is automatically generated and consists of structs representing custom annotations
	Annotations struct {
		Self    []interface{}
		Fields  map[string][]interface{}
		Methods map[string][]interface{}
	}
)

var (
	typeRegistry = make(map[string]Annotations)
)

// Maps annotations bundle to provided string.
// Usually string contains the type and the name of annotated entry
func Map(s string, a Annotations) {
	typeRegistry[s] = a
}

// Maps annotation bundle to the type and name of provided object
func MapType(i interface{}, a Annotations) {
	// check whether provided object is a type or an instance
	var typ reflect.Type
	switch t := i.(type) {
	case reflect.Type:
		typ = t
	default:
		typ = reflect.TypeOf(i)
	}
	// check for predeclared or unnamed type
	pck := typ.PkgPath()
	if pck == "" || typ.Name() == "" {
		panic("Unable to annotate predeclared or unnamed type " + typ.String())
	}
	switch typ.Kind() {
	case reflect.Struct, reflect.Interface, reflect.Func:
		typeRegistry[pck+"."+typ.Name()] = a
	default:
		panic("Unable to annotate object of type " + typ.String())
	}
}

// Returns annotations bundle for provided struct type.
// Object instance or its reflect.Type is passed as the parameter.
// If no annotation defined for given type then nil is returned
func GetStructAnnotations(s interface{}) []interface{} {
	a, found := findAnnotationsByType(s)
	if found {
		return a.Self
	}
	return nil
}

// Returns annotations bundle for specified field of provided object type.
// Object instance or its reflect.Type is passed as the first parameter.
// Field name is passed as the second parameter.
// If no annotation defined for given type or specified field is not annotated
// or no such field exist then nil is returned
func GetFieldAnnotations(s interface{}, fieldName string) []interface{} {
	a, found := findAnnotationsByType(s)
	if found {
		return a.Fields[fieldName]
	}
	return nil
}

// Returns annotations bundle for specified method of provided object type.
// Object instance or its reflect.Type is passed as the first parameter.
// Method name is passed as the second parameter.
// If no annotation defined for given type then nil is returned
func GetMethodAnnotations(s interface{}, methodName string) []interface{} {
	a, found := findAnnotationsByType(s)
	if found {
		return a.Methods[methodName]
	}
	return nil
}

// Returns annotations bundle for provided func type.
// Func instance or its reflect.Type is passed as the parameter.
// If no annotation defined for given type then nil is returned
func GetFuncAnnotation(s interface{}) []interface{} {
	a, found := findAnnotationsByType(s)
	if found {
		return a.Self
	}
	return nil
}

func findAnnotationsByType(s interface{}) (*Annotations, bool) {
	var path string
	switch t := s.(type) {
	case string:
		path = t
	case reflect.Type:
		path = t.PkgPath() + "." + t.Name()
	default:
		tp := reflect.TypeOf(t)
		path = tp.PkgPath() + "." + tp.Name()
	}
	a, found := typeRegistry[path]
	return &a, found
}
