package registry

import (
	"testing"
)

func TestFindAnnotations(t *testing.T) {
	doc := "  @Entity(param=\"value\")"
	r := FindAnnotations(doc)
	if len(r) != 1 {
		t.Fatalf("Incorrect amount of entities: %s", len(r))
	}
	a := r[0]
	if a.Name != "Entity" {
		t.Errorf("Annotation name is wrong. Expected %s but got %s", "Entity", a.Name)
	}
	if len(a.Content) != 1 {
		t.Fatalf("Content should have 1 parameter but it has %s", len(a.Content))
	}
	v, ok := a.Content["param"]
	if !ok {
		t.Fatal("Parameter 'param' is not found in content")
	}
	if v != "value" {
		t.Errorf("Expected parameter value is 'value' but it is %#v", v)
	}
}

func TestFindAnnotation2(t *testing.T) {
	doc := "  @Entity(\"value\")"
	r := FindAnnotations(doc)
	if len(r) != 1 {
		t.Fatalf("Incorrect amount of entities: %s", len(r))
	}
	a := r[0]
	if a.Name != "Entity" {
		t.Errorf("Annotation name is wrong. Expected %s but got %s", "Entity", a.Name)
	}
	if len(a.Content) != 1 {
		t.Fatalf("Content should have 1 parameter but it has %s", len(a.Content))
	}
	v, ok := a.Content[DEFAULT_PARAM]
	if !ok {
		t.Fatalf("Parameter %s is not found in content", DEFAULT_PARAM)
	}
	if v != "value" {
		t.Errorf("Expected parameter value is 'value' but it is %#v", v)
	}
}

func TestFindAnnotation3(t *testing.T) {
	doc := "  @Entity({\"value1\", \"value2\"})"
	r := FindAnnotations(doc)
	if len(r) != 1 {
		t.Fatalf("Incorrect amount of entities: %s", len(r))
	}
	a := r[0]
	if a.Name != "Entity" {
		t.Errorf("Annotation name is wrong. Expected %s but got %s", "Entity", a.Name)
	}
	if len(a.Content) != 1 {
		t.Fatalf("Content should have 1 parameter but it has %s", len(a.Content))
	}
	v, ok := a.Content[DEFAULT_PARAM]
	if !ok {
		t.Fatalf("Parameter %s is not found in content", DEFAULT_PARAM)
	}
	s, ok := v.([]string)
	if !ok {
		t.Fatalf("Incorrect value type. Expected []string but it is %#v", v)
	}
	if len(s) != 2 {
		t.Fatalf("Incorrect value array size. Expected 2 but it is %d", len(s))
	}
	if s[0] != "value1" || s[1] != "value2" {
		t.Errorf("Expected parameter value is {'value1','value2'} but it is %#v", s)
	}
}

func TestFindAnnotations4(t *testing.T) {
	doc := "  @Entity(@SubEntity)"
	r := FindAnnotations(doc)
	if len(r) != 1 {
		t.Fatalf("Incorrect amount of entities: %s", len(r))
	}
	a := r[0]
	if a.Name != "Entity" {
		t.Errorf("Annotation name is wrong. Expected %s but got %s", "Entity", a.Name)
	}
	if len(a.Content) != 1 {
		t.Fatalf("Content should have 1 parameter but it has %s", len(a.Content))
	}
	v, found := a.Content[DEFAULT_PARAM]
	if !found {
		t.Fatalf("Parameter %s is not found in content", DEFAULT_PARAM)
	}
	s, ok := v.(AnnotationDoc)
	if !ok {
		t.Fatalf("Wrong parameter type. Expected AnnotationDoc but found %#v", v)
	}
	if s.Name != "SubEntity" {
		t.Errorf("Expected parameter value name is 'SubEntity' but it is %s", s.Name)
	}
	if len(s.Content) != 0 {
		t.Errorf("Value content should be empty but it is %#v", s.Content)
	}
}

func TestFindAnnotations5(t *testing.T) {
	doc := "@Entity(param1=\"value1\",param2=@SubEntity(col1={0, 1},col2=2))"
	r := FindAnnotations(doc)
	if len(r) != 1 {
		t.Fatalf("Incorrect amount of entities: %s", len(r))
	}
	a := r[0]
	if a.Name != "Entity" {
		t.Errorf("Annotation name is wrong. Expected %s but got %s", "Entity", a.Name)
	}
	if len(a.Content) != 2 {
		t.Fatalf("Content should have 2 parameters but it has %s", len(a.Content))
	}
	v, found := a.Content["param1"]
	if !found {
		t.Fatal("Parameter 'param1' is not found in content")
	}
	if v != "value1" {
		t.Errorf("Expected parameter value is 'value1' but it is %#v", v)
	}
	s, found := a.Content["param2"]
	if !found {
		t.Fatal("Parameter 'param2' is not found in content")
	}
	sa, ok := s.(AnnotationDoc)
	if !ok {
		t.Fatal("Parameter 2 has incorrect type. Expected AnnotationDoc but it is %#v", s)
	}
	if sa.Name != "SubEntity" {
		t.Errorf("Parameter 2 has incorrect name. Expected 'SubEntity' but found %s", sa.Name)
	}
	if len(sa.Content) != 2 {
		t.Fatalf("Incorrect parameter 2 content length. Expected 2 but found %d", len(sa.Content))
	}
	sv1, found := sa.Content["col1"]
	if !found {
		t.Fatal("Parameter 2 parameter value is not found. Expected name 'col1'")
	}
	sv11, ok := sv1.([]string)
	if !ok {
		t.Fatalf("Parameter 2 value is not []string. It is %#v", sv1)
	}
	if len(sv11) != 2 {
		t.Fatalf("Parameter 2 value array has incorrect size %d. Expected is 2", len(sv11))
	}
	if sv11[0] != "0" || sv11[1] != "1" {
		t.Fatalf("Incorrect value of parameters of parameter 2: %#v", sv1)
	}
	sv2, found := sa.Content["col2"]
	if !found {
		t.Fatal("Parameter 2 parameter value is not found. Expected name 'col2'")
	}
	sv21, ok := sv2.(string)
	if !ok {
		t.Fatalf("Parameter 2 value is not string. It is %#v", sv2)
	}
	if sv21 != "2" {
		t.Fatalf("Incorrect value of parameters of parameter 2: %#v", sv21)
	}
}
