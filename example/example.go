package example

import (
	"github.com/SphereSoftware/go-annotations/example/test"
	"github.com/SphereSoftware/go-annotations/registry"
	"log"
)

//go:generate go-annotations example_annotations

type (
	// @Entity(Name="test", Books={@Book(Name="book",Author=@Person("Mr.X"))})
	Test struct{}

	//@Entity
	Test2 struct {
		//@Book
		Name string
	}

	// @Entity
	Sample interface {
		// @Book
		doSomething(s string) int
	}
)

// @Entity
func (*Test) methodOfTest() {
	// do nothing
}

// @Book
func JustAFunc() {
	// do nothing
}

func TestAnnotations() {
	a := registry.GetStructAnnotations("github.com/SphereSoftware/go-annotations/example.Test")
	if a != nil && len(a) > 0 {
		e := a[0].(test.Entity)
		log.Printf("Test example annotation: %#v\n", e)
	} else {
		log.Printf("Test example is not found\n")
	}
}
