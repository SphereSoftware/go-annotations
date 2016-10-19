# Go Annotations

Annotations support for Go


This project provides ability to add Java-style annotations in the comments to struct, interface, method and/or func.
Annotations from comments are transformed into objects and mapped to provided struct for use in the code.
Code generation (go generate) is used to create the registry of all annotations in the project.

## Quick start:

1. Make sure you're using Go 1.6+
2. Install or update it:  `go get -u github.com/SphereSoftware/go-annotations`
3. Define you first annotation in any go source file, for example `annotation.go`:

   ```go
   type Entity struct {
       Name string `default:"unknown"`
   }
   ```

4. Define your annotated entity in any other (or the same) go source file `example.go`:

   ```go
   package example

   //go:generate go-annotations

   type (
      // @Entity(name="test")
      Person struct {
         Id       int
         FullName string
      }
   )
   ```

5. Use that annotation in the code somewhere else:

   ```go
   import (
     "github.com/SphereSoftware/go-annotations/registry"
   )

   func Test() {
       a := registry.GetStructAnnotatoin(Person{}).(Entity)
       ...
   }
   ```

6. Run `go generate [package or file]`. This will create `example_annotations.go`
   in the same package with type `Person`

## Features

* Annotation class inside the comments is started with the '@' character
* Top-level annotation's package should be included with _ alias
* Structures, interfaces, methods and functions can be annotated
* Annotation can contain parameters: comma-separated list of property-value pairs
* If annotation has only one attribute then only its value can be specified as the parameter
* Array property value is enclosed by {} and elements are comma-separated
* Property value can be string, number or another annotation
* For each annotation the corresponding struct should be defined
* Optional annotation attributes are defined as pointers types
* Defauld attribute value could be specified using field tag "default:"
* Annotation registry file is generated for the whole package
* At least one source with annotations should contain `go:generate` tag
* The optional parameter of `go:generate` tag can specify the name of generated source file for regstry
* Default registry source is named `<package_name>`+"_annotations.go"
* The set of methods is provided to get annotations list for specified struct, func, interface or field name

## API

The following functions from `registry` package are provided to work with annotations:

* `func Map(s string, a Annotations)` - associates the set of annotations with given tag. Tag contains the
information about full package name and object name in form of `<full_package_name>`.`<object_name>`
* `func MapType(i interface{}, a Annotations)` - maps annotation bundle to the type and name of provided object
* `func GetStructAnnotations(s interface{}) []interface{}` - returns struct/interface/func level annotations.
Parameter can be either object instance or object's `refect.Type` instance or string with object's package and
name information, as described for `Map` function
* `func GetFieldAnnotations(s interface{}, fieldName string) []interface{}` - returns annotations bundle for 
specified field of provided object type
* `func GetMethodAnnotations(s interface{}, methodName string) []interface{}` - returns annotations bundle for 
specified method of provided object type
* `func GetFuncAnnotation(s interface{}) []interface{}` - returns annotations bundle for provided func type

## More examples of annotations

   ```go
   package example

   import (
       "github.com/SphereSoftware/go-annotations/example/test"
       "github.com/SphereSoftware/go-annotations/registry"
   )

   //go:generate go-annotations my_annotations

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
       ...
   }

   // @Book
   func JustAFunc() {
       ...
   }
   ```

   ```go
   package test

   import (
       r "github.com/SphereSoftware/go-annotations/example/test2"
   )

   type (
       Entity struct {
           Name  string
           Books []Book
       }

       Book struct {
           Name   string
           Price  float32 `default:"1.0"`
           Author *r.Person
       }
   )
   ```

   ```go
   package test2

   type (
       Person struct {
           Name string
       }
   )
   ```
